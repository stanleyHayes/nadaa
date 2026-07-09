package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/stanleyHayes/nadaa/services/route-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/utils"
)

const (
	planRequestTimeout     = 5 * time.Second
	externalRequestTimeout = 2 * time.Second
	fallbackOffsetMeters   = 3000.0
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

	destination, targetShelter, err := s.resolveDestination(ctx, request)
	if err != nil {
		log.Printf("WARN route-service plan_route destination_resolution_failed error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "destination_unresolvable", "could not determine a destination for the route")
		return
	}

	closures := s.fetchClosures(ctx, request.Origin, *destination)
	riskAreas := s.fetchRiskAreas(ctx, request.Origin, *destination)

	response := buildRoute(request, *destination, targetShelter, closures, riskAreas, s.now().UTC())
	log.Printf("INFO route-service plan_route completed origin=%.5f,%.5f destination=%.5f,%.5f distance=%d duration=%d closures=%d risk=%d",
		request.Origin.Lat, request.Origin.Lng,
		destination.Lat, destination.Lng,
		response.DistanceMeters, response.EstimatedDurationMinutes,
		len(response.AvoidedClosures), len(response.AvoidedRiskZones))
	utils.WriteJSON(w, http.StatusOK, response)
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

func (s *Server) resolveDestination(ctx context.Context, request models.RoutePlanRequest) (*models.Coordinates, *models.Shelter, error) {
	if request.Destination != nil {
		dest := *request.Destination
		return &dest, nil, nil
	}

	if request.WaypointType == "shelter" {
		shelter := s.findNearestShelter(ctx, request.Origin)
		if shelter != nil {
			dest := shelter.Location
			return &dest, shelter, nil
		}
	}

	// Fallback higher-ground waypoint: a fixed offset toward the north-east.
	fallback := utils.DestinationPoint(request.Origin, fallbackOffsetMeters, 45)
	return &fallback, nil, nil
}

func (s *Server) findNearestShelter(ctx context.Context, origin models.Coordinates) *models.Shelter {
	endpoint, err := url.JoinPath(s.config.ShelterServiceURL, "/shelters/nearby")
	if err != nil {
		log.Printf("WARN route-service invalid_shelter_url error=%v", err)
		return nil
	}
	query := url.Values{}
	query.Set("lat", fmt.Sprintf("%f", origin.Lat))
	query.Set("lng", fmt.Sprintf("%f", origin.Lng))
	endpoint = endpoint + "?" + query.Encode()

	var response models.NearbyShelterResponse
	if err := s.fetchJSON(ctx, endpoint, &response); err != nil {
		log.Printf("WARN route-service shelter_lookup_failed error=%v", err)
		return nil
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
	return nearest
}

func shelterOpen(status string) bool {
	switch utils.NormalizeString(status) {
	case "", "open", "active", "operational":
		return true
	default:
		return false
	}
}

func (s *Server) fetchClosures(ctx context.Context, origin, destination models.Coordinates) []models.RoadClosure {
	endpoint, err := url.JoinPath(s.config.RoadClosureServiceURL, "/road-closures")
	if err != nil {
		log.Printf("WARN route-service invalid_closure_url error=%v", err)
		return nil
	}
	query := url.Values{}
	query.Set("lat", fmt.Sprintf("%f", origin.Lat))
	query.Set("lng", fmt.Sprintf("%f", origin.Lng))
	query.Set("status", "active")
	endpoint = endpoint + "?" + query.Encode()

	var response models.RoadClosureListResponse
	if err := s.fetchJSON(ctx, endpoint, &response); err != nil {
		log.Printf("WARN route-service closure_lookup_failed error=%v", err)
		return nil
	}
	return response.Closures
}

func (s *Server) fetchRiskAreas(ctx context.Context, origin, destination models.Coordinates) []models.RiskArea {
	endpoint, err := url.JoinPath(s.config.RiskServiceURL, "/risk/areas")
	if err != nil {
		log.Printf("WARN route-service invalid_risk_url error=%v", err)
		return nil
	}
	query := url.Values{}
	query.Set("lat", fmt.Sprintf("%f", origin.Lat))
	query.Set("lng", fmt.Sprintf("%f", origin.Lng))
	endpoint = endpoint + "?" + query.Encode()

	var response models.RiskAreaListResponse
	if err := s.fetchJSON(ctx, endpoint, &response); err != nil {
		log.Printf("WARN route-service risk_lookup_failed error=%v", err)
		return nil
	}
	return response.Areas
}

func (s *Server) fetchJSON(ctx context.Context, endpoint string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

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
	route := []models.Coordinates{request.Origin}

	totalDistance := utils.DistanceMeters(request.Origin, destination)
	sampleCount := 2
	if totalDistance > utils.DefaultSampleStepMeters {
		sampleCount = int(math.Ceil(totalDistance/utils.DefaultSampleStepMeters)) + 1
	}

	avoidedClosureIDs := map[string]bool{}
	avoidedRiskIDs := map[string]bool{}

	for i := 1; i < sampleCount; i++ {
		t := float64(i) / float64(sampleCount-1)
		if t >= 1 {
			continue
		}
		sample := utils.Interpolate(request.Origin, destination, t)

		blocked, byClosureID, byRiskID := sampleBlocked(sample, closures, riskAreas, request)
		if !blocked {
			continue
		}

		// Add a single detour for each contiguous blocked segment.
		if i > 1 {
			prev := utils.Interpolate(request.Origin, destination, float64(i-1)/float64(sampleCount-1))
			if _, prevClosureID, prevRiskID := sampleBlocked(prev, closures, riskAreas, request); prevClosureID != "" || prevRiskID != "" {
				// Still inside the same blocked region; do not add another detour.
				continue
			}
		}

		detour := perpendicularDetour(request.Origin, destination, sample, utils.DefaultDetourOffsetMeters)
		route = append(route, detour)
		if byClosureID != "" {
			avoidedClosureIDs[byClosureID] = true
		}
		if byRiskID != "" {
			avoidedRiskIDs[byRiskID] = true
		}
	}

	route = append(route, destination)

	distance := 0.0
	segments := make([]models.RouteSegment, 0, len(route)-1)
	for i := 0; i < len(route)-1; i++ {
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

func sampleBlocked(sample models.Coordinates, closures []models.RoadClosure, riskAreas []models.RiskArea, request models.RoutePlanRequest) (bool, string, string) {
	for _, closure := range closures {
		if utils.NormalizeString(closure.Status) != "active" {
			continue
		}
		d := utils.MinDistanceToLineString(sample, closure.Geometry.Coordinates)
		if d <= request.ClosureBufferMeters {
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
		if area.RadiusMeters > 0 && d <= area.RadiusMeters {
			return true, "", area.ID
		}
	}
	return false, "", ""
}

func perpendicularDetour(origin, destination, sample models.Coordinates, offsetMeters float64) models.Coordinates {
	bearing := utils.Bearing(origin, destination)
	// Push to the right of the forward direction.
	detourBearing := math.Mod(bearing+90, 360)
	return utils.DestinationPoint(sample, offsetMeters, detourBearing)
}

func keys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}
