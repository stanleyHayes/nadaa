package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/route-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/utils"
)

const (
	planRequestTimeout     = 5 * time.Second
	externalRequestTimeout = 2 * time.Second
	fallbackOffsetMeters   = 3000.0
	// maxRiskSamples caps how many points along a route are probed for risk so
	// a long route cannot fan out into unbounded upstream calls.
	maxRiskSamples = 12
	// maxPlanBodyBytes caps the unauthenticated plan request body at 1 MiB.
	maxPlanBodyBytes = 1 << 20
	// maxDetourPasses caps how many times the polyline is re-sampled and
	// re-routed so pathological hazard layouts cannot loop forever.
	maxDetourPasses = 6
	// maxDetourOffsetMeters bounds how far a detour waypoint may push away
	// from a blocked sample while searching for a clearing side.
	maxDetourOffsetMeters = 8 * utils.DefaultDetourOffsetMeters
)

var allowedWaypointTypes = map[string]bool{
	"shelter":       true,
	"higher_ground": true,
	"manual":        true,
}

func (s *Server) optionsHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, models.OptionsResponse{
		WaypointTypes: []string{"shelter", "higher_ground", "manual"},
		GeneratedAt:   s.now().UTC(),
	})
}

func (s *Server) planRouteHandler(w http.ResponseWriter, r *http.Request) {
	var request models.RoutePlanRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxPlanBodyBytes)
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN route-service plan_route invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.WaypointType = utils.NormalizeString(request.WaypointType)
	request.AvoidRiskLevels = utils.NormalizeRiskLevels(request.AvoidRiskLevels)
	if len(request.AvoidRiskLevels) == 0 {
		request.AvoidRiskLevels = []string{"severe", "emergency"}
	}
	if request.ClosureBufferMeters <= 0 {
		request.ClosureBufferMeters = utils.DefaultClosureBufferMeters
	}

	if code, message := validateRequest(request); code != "" {
		log.Printf("WARN route-service plan_route validation_failed code=%s", code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), planRequestTimeout)
	defer cancel()

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	destination, targetShelter, err := s.resolveDestination(ctx, request, authHeader)
	if err != nil {
		switch {
		case errors.Is(err, errShelterLookupFailed):
			log.Printf("WARN route-service plan_route shelter_lookup_failed error=%v", err)
			utils.WriteError(w, http.StatusBadGateway, "shelter_lookup_failed", "shelter lookup is currently degraded; cannot plan a shelter route")
		case errors.Is(err, errNoShelterAvailable):
			log.Printf("WARN route-service plan_route no_shelter_available")
			utils.WriteError(w, http.StatusNotFound, "no_shelter_available", "no open shelter could be found near the origin")
		default:
			log.Printf("WARN route-service plan_route destination_resolution_failed error=%v", err)
			utils.WriteError(w, http.StatusBadRequest, "destination_unresolvable", "could not determine a destination for the route")
		}
		return
	}

	closures, closuresOK := s.fetchClosures(ctx, request.Origin, *destination, request.ClosureBufferMeters, authHeader)
	riskAreas, riskOK := s.fetchRiskZones(ctx, request.Origin, *destination, request.AvoidRiskLevels, authHeader)

	response := buildRoute(request, *destination, targetShelter, closures, riskAreas, s.now().UTC())
	// A failed hazard lookup must not be indistinguishable from a verified
	// hazard-free corridor: flag the response degraded and say what failed.
	if !closuresOK || !riskOK {
		response.Degraded = true
		response.EnrichmentStatus = enrichmentStatus(closuresOK, riskOK)
	}
	log.Printf("INFO route-service plan_route completed origin=%.5f,%.5f destination=%.5f,%.5f distance=%d duration=%d closures=%d risk=%d degraded=%t",
		request.Origin.Lat, request.Origin.Lng,
		destination.Lat, destination.Lng,
		response.DistanceMeters, response.EstimatedDurationMinutes,
		len(response.AvoidedClosures), len(response.AvoidedRiskZones), response.Degraded)
	utils.WriteJSON(w, http.StatusOK, response)
}

// enrichmentStatus names the failed hazard lookups for degraded responses.
func enrichmentStatus(closuresOK, riskOK bool) string {
	switch {
	case !closuresOK && !riskOK:
		return "closure_lookup_failed,risk_lookup_failed"
	case !closuresOK:
		return "closure_lookup_failed"
	default:
		return "risk_lookup_failed"
	}
}

func validateRequest(request models.RoutePlanRequest) (string, string) {
	if !utils.ValidCoordinates(request.Origin) {
		return "invalid_origin", "origin must be a valid WGS84 latitude and longitude"
	}
	if request.WaypointType == "" {
		request.WaypointType = "manual"
	}
	if !allowedWaypointTypes[request.WaypointType] {
		return "invalid_waypoint_type", "waypointType must be shelter, higher_ground, or manual"
	}
	if request.Destination != nil && !utils.ValidCoordinates(*request.Destination) {
		return "invalid_destination", "destination must be a valid WGS84 latitude and longitude"
	}
	if request.WaypointType == "manual" && request.Destination == nil {
		return "missing_destination", "manual waypointType requires a destination"
	}
	return "", ""
}

var (
	// errShelterLookupFailed marks a degraded shelter-service lookup.
	errShelterLookupFailed = errors.New("shelter lookup failed")
	// errNoShelterAvailable marks a successful lookup with no open shelter.
	errNoShelterAvailable = errors.New("no open shelter available")
)

func (s *Server) resolveDestination(ctx context.Context, request models.RoutePlanRequest, authHeader string) (*models.Coordinates, *models.Shelter, error) {
	if request.Destination != nil {
		dest := *request.Destination
		return &dest, nil, nil
	}

	if request.WaypointType == "shelter" {
		// Never substitute a fabricated destination for shelter routing: a
		// failed or empty lookup must surface as an error, not a made-up point.
		shelter, err := s.findNearestShelter(ctx, request.Origin, authHeader)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errShelterLookupFailed, err)
		}
		if shelter == nil {
			return nil, nil, errNoShelterAvailable
		}
		dest := shelter.Location
		return &dest, shelter, nil
	}

	// Fallback higher-ground waypoint: a fixed offset toward the north-east.
	fallback := utils.DestinationPoint(request.Origin, fallbackOffsetMeters, 45)
	return &fallback, nil, nil
}

func (s *Server) findNearestShelter(ctx context.Context, origin models.Coordinates, authHeader string) (*models.Shelter, error) {
	endpoint, err := url.JoinPath(s.config.ShelterServiceURL, "/api/v1/shelters/nearby")
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("lat", fmt.Sprintf("%f", origin.Lat))
	query.Set("lng", fmt.Sprintf("%f", origin.Lng))
	endpoint = endpoint + "?" + query.Encode()

	var response models.NearbyShelterResponse
	if err := s.fetchJSON(ctx, endpoint, authHeader, &response); err != nil {
		return nil, err
	}

	var nearest *models.Shelter
	nearestDistance := math.MaxFloat64
	for i := range response.Shelters {
		shelter := &response.Shelters[i]
		if !shelterOpen(shelter.Status) {
			continue
		}
		d := utils.DistanceMeters(origin, shelter.Location)
		if d < nearestDistance {
			nearestDistance = d
			nearest = shelter
		}
	}
	return nearest, nil
}

func shelterOpen(status string) bool {
	switch utils.NormalizeString(status) {
	case "", "open", "active", "operational":
		return true
	default:
		return false
	}
}

func (s *Server) fetchClosures(ctx context.Context, origin, destination models.Coordinates, bufferMeters float64, authHeader string) ([]models.RoadClosure, bool) {
	endpoint, err := url.JoinPath(s.config.RoadClosureServiceURL, "/api/v1/road-closures")
	if err != nil {
		log.Printf("WARN route-service invalid_closure_url error=%v", err)
		return nil, false
	}
	query := url.Values{}
	query.Set("status", "active")
	// Cover the whole origin→destination corridor, not just the origin area;
	// the padding keeps closures within detour reach of the route line.
	query.Set("bbox", utils.CorridorBBox(origin, destination, bufferMeters+utils.DefaultDetourOffsetMeters))
	endpoint = endpoint + "?" + query.Encode()

	var response models.RoadClosureListResponse
	if err := s.fetchJSON(ctx, endpoint, authHeader, &response); err != nil {
		log.Printf("WARN route-service closure_lookup_failed error=%v", err)
		return nil, false
	}
	return response.Closures, true
}

// fetchRiskZones probes the risk-service at sample points along the route and
// turns every sample whose overallRisk should be avoided into a circular zone
// covering its stretch of the corridor. Risk-service has no areas endpoint, so
// sampling is the only supported contract. Probes run concurrently so up to
// maxRiskSamples slow upstream calls still fit the request budget; any failed
// probe marks the lookup degraded.
func (s *Server) fetchRiskZones(ctx context.Context, origin, destination models.Coordinates, avoidLevels []string, authHeader string) ([]models.RiskArea, bool) {
	distance := utils.DistanceMeters(origin, destination)
	sampleCount := int(math.Ceil(distance/utils.DefaultSampleStepMeters)) + 1
	sampleCount = min(max(sampleCount, 2), maxRiskSamples)
	// One spacing around each sample so zones tile the corridor, endpoints
	// included, and always cover the nearest route sample point.
	radius := distance / float64(sampleCount-1)

	type sampleResult struct {
		level string
		err   error
	}
	results := make([]sampleResult, sampleCount)
	var wg sync.WaitGroup
	for i := range sampleCount {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sample := utils.Interpolate(origin, destination, float64(i)/float64(sampleCount-1))
			level, err := s.fetchRiskLevel(ctx, sample, authHeader)
			if err != nil {
				log.Printf("WARN route-service risk_sample_failed error=%v", err)
			}
			results[i] = sampleResult{level: level, err: err}
		}(i)
	}
	wg.Wait()

	ok := true
	var zones []models.RiskArea
	for i, result := range results {
		if result.err != nil {
			ok = false
			continue
		}
		if !utils.ShouldAvoidRisk(result.level, avoidLevels) {
			continue
		}
		zones = append(zones, models.RiskArea{
			ID:           fmt.Sprintf("risk_sample_%d", i),
			RiskLevel:    result.level,
			Center:       utils.Interpolate(origin, destination, float64(i)/float64(sampleCount-1)),
			RadiusMeters: radius,
		})
	}
	return zones, ok
}

func (s *Server) fetchRiskLevel(ctx context.Context, location models.Coordinates, authHeader string) (string, error) {
	endpoint, err := url.JoinPath(s.config.RiskServiceURL, "/api/v1/risk")
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("lat", fmt.Sprintf("%f", location.Lat))
	query.Set("lng", fmt.Sprintf("%f", location.Lng))
	endpoint = endpoint + "?" + query.Encode()

	var response models.RiskSampleResponse
	if err := s.fetchJSON(ctx, endpoint, authHeader, &response); err != nil {
		return "", err
	}
	return response.OverallRisk, nil
}

func (s *Server) fetchJSON(ctx context.Context, endpoint, authHeader string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	// Forward the caller's credentials so upstream services can authorize the
	// request as the same actor; never fabricate actor headers.
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, externalRequestTimeout)
	defer cancel()
	req = req.WithContext(ctxTimeout)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func buildRoute(request models.RoutePlanRequest, destination models.Coordinates, targetShelter *models.Shelter, closures []models.RoadClosure, riskAreas []models.RiskArea, now time.Time) models.RoutePlanResponse {
	route := []models.Coordinates{request.Origin, destination}
	encounteredClosures := map[string]bool{}
	encounteredRiskZones := map[string]bool{}

	// Re-sample the actual polyline after every detour insertion: a detour can
	// re-cross the same or another hazard, so only the final line decides
	// which hazards were truly avoided.
	for range maxDetourPasses {
		insertions := detourInsertions(route, closures, riskAreas, request, encounteredClosures, encounteredRiskZones)
		if len(insertions) == 0 {
			break
		}
		route = applyDetours(route, insertions)
	}

	// Only hazards the final polyline fully clears may be reported as avoided.
	stillClosures := map[string]bool{}
	stillRiskZones := map[string]bool{}
	detourInsertions(route, closures, riskAreas, request, stillClosures, stillRiskZones)
	avoidedClosureIDs := map[string]bool{}
	for id := range encounteredClosures {
		if !stillClosures[id] {
			avoidedClosureIDs[id] = true
		}
	}
	avoidedRiskIDs := map[string]bool{}
	for id := range encounteredRiskZones {
		if !stillRiskZones[id] {
			avoidedRiskIDs[id] = true
		}
	}

	distance := 0.0
	segments := make([]models.RouteSegment, 0, len(route)-1)
	for i := range len(route) - 1 {
		segmentDistance := utils.DistanceMeters(route[i], route[i+1])
		distance += segmentDistance
		segments = append(segments, models.RouteSegment{
			Start:          route[i],
			End:            route[i+1],
			DistanceMeters: int(math.Round(segmentDistance)),
			Mode:           "walking",
		})
	}

	durationMinutes := 0
	if distance > 0 {
		durationMinutes = int(math.Round(distance / utils.WalkingSpeedMetersPerSecond / 60))
	}

	return models.RoutePlanResponse{
		Route:                    route,
		Segments:                 segments,
		DistanceMeters:           int(math.Round(distance)),
		EstimatedDurationMinutes: durationMinutes,
		TargetShelter:            targetShelter,
		AvoidedClosures:          keys(avoidedClosureIDs),
		AvoidedRiskZones:         keys(avoidedRiskIDs),
		Disclaimer:               "This route is decision support only; follow official emergency instructions.",
		GeneratedAt:              now,
		DecisionSupport:          true,
	}
}

// detourInsertions samples the polyline segment by segment and returns the
// detour waypoints to insert, keyed by segment start index. Each contiguous
// run of blocked samples on a segment yields one detour at the run's first
// sample; blocking hazards are recorded in the encountered sets.
func detourInsertions(route []models.Coordinates, closures []models.RoadClosure, riskAreas []models.RiskArea, request models.RoutePlanRequest, encounteredClosures, encounteredRiskZones map[string]bool) map[int][]models.Coordinates {
	insertions := map[int][]models.Coordinates{}
	origin, destination := route[0], route[len(route)-1]
	// Sample finely and treat anything within the buffer plus a half-step
	// margin as blocked, so a graze within the closure buffer or a zone
	// radius cannot slip between two sample points undetected.
	step := min(utils.DefaultSampleStepMeters, max(request.ClosureBufferMeters, 25))
	margin := step / 2
	for i := 0; i+1 < len(route); i++ {
		a, b := route[i], route[i+1]
		sampleCount := int(math.Ceil(utils.DistanceMeters(a, b)/step)) + 1
		sampleCount = max(sampleCount, 2)

		runStart := -1
		var runClosureID, runRiskID string
		flush := func() {
			if runStart < 0 {
				return
			}
			blocked := utils.Interpolate(a, b, float64(runStart)/float64(sampleCount-1))
			insertions[i] = append(insertions[i], clearingDetour(a, b, blocked, closures, riskAreas, request, margin))
			if runClosureID != "" {
				encounteredClosures[runClosureID] = true
			}
			if runRiskID != "" {
				encounteredRiskZones[runRiskID] = true
			}
			runStart, runClosureID, runRiskID = -1, "", ""
		}

		for j := range sampleCount {
			sample := utils.Interpolate(a, b, float64(j)/float64(sampleCount-1))
			// The journey's origin and destination are fixed: the traveler
			// starts and ends there regardless of hazards, so those exact
			// points are never detoured around.
			if utils.DistanceMeters(sample, origin) < 1 || utils.DistanceMeters(sample, destination) < 1 {
				continue
			}
			blocked, closureID, riskID := sampleBlocked(sample, closures, riskAreas, request, margin)
			if !blocked {
				flush()
				continue
			}
			if runStart < 0 {
				runStart, runClosureID, runRiskID = j, closureID, riskID
			}
		}
		flush()
	}
	return insertions
}

// applyDetours returns a new polyline with each segment's detour waypoints
// inserted between its start and end.
func applyDetours(route []models.Coordinates, insertions map[int][]models.Coordinates) []models.Coordinates {
	updated := make([]models.Coordinates, 0, len(route)+len(insertions))
	for i, point := range route {
		updated = append(updated, point)
		if i+1 < len(route) {
			updated = append(updated, insertions[i]...)
		}
	}
	return updated
}

// clearingDetour searches both sides of the blocked sample at growing
// perpendicular offsets for a waypoint that itself clears every hazard,
// falling back to the default right-hand offset when nothing clears.
func clearingDetour(a, b, blocked models.Coordinates, closures []models.RoadClosure, riskAreas []models.RiskArea, request models.RoutePlanRequest, margin float64) models.Coordinates {
	bearing := utils.Bearing(a, b)
	for offset := utils.DefaultDetourOffsetMeters; offset <= maxDetourOffsetMeters; offset += utils.DefaultDetourOffsetMeters {
		for _, side := range []float64{90, -90} {
			candidate := utils.DestinationPoint(blocked, offset, math.Mod(bearing+side+360, 360))
			if ok, _, _ := sampleBlocked(candidate, closures, riskAreas, request, margin); !ok {
				return candidate
			}
		}
	}
	return utils.DestinationPoint(blocked, utils.DefaultDetourOffsetMeters, math.Mod(bearing+90, 360))
}

// sampleBlocked reports whether sample lies within the closure buffer (plus
// margin) of an active closure or inside an avoid-listed risk zone (plus
// margin), returning the blocking hazard IDs.
func sampleBlocked(sample models.Coordinates, closures []models.RoadClosure, riskAreas []models.RiskArea, request models.RoutePlanRequest, margin float64) (bool, string, string) {
	for _, closure := range closures {
		if utils.NormalizeString(closure.Status) != "active" {
			continue
		}
		d := utils.MinDistanceToLineString(sample, closure.Geometry.Coordinates)
		if d <= request.ClosureBufferMeters+margin {
			return true, closure.ID, ""
		}
	}
	for _, area := range riskAreas {
		if !utils.ShouldAvoidRisk(area.RiskLevel, request.AvoidRiskLevels) {
			continue
		}
		if len(area.Polygon) > 0 {
			if utils.PointInPolygon(sample, area.Polygon) {
				return true, "", area.ID
			}
			continue
		}
		d := utils.DistanceMeters(sample, area.Center)
		if area.RadiusMeters > 0 && d <= area.RadiusMeters+margin {
			return true, "", area.ID
		}
	}
	return false, "", ""
}

func keys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}
