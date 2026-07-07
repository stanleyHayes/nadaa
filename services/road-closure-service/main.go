package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *memoryStore
	now   func() time.Time
}

type memoryStore struct {
	mu            sync.RWMutex
	closures      []roadClosureRecord
	closureCounter int
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type lineStringGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type roadClosureRecord struct {
	ID            string             `json:"id"`
	RoadName      string             `json:"roadName"`
	Reason        string             `json:"reason,omitempty"`
	Status        string             `json:"status"`
	Severity      string             `json:"severity"`
	Source        string             `json:"source"`
	SourceRef     string             `json:"sourceRef,omitempty"`
	Geometry      lineStringGeometry `json:"geometry"`
	ValidFrom     time.Time          `json:"validFrom"`
	ValidTo       *time.Time         `json:"validTo,omitempty"`
	DetourNote    string             `json:"detourNote,omitempty"`
	DistanceMeters int               `json:"distanceMeters,omitempty"`
	CreatedBy     string             `json:"createdBy,omitempty"`
	UpdatedBy     string             `json:"updatedBy,omitempty"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
}

type roadClosureListResponse struct {
	Closures    []roadClosureRecord `json:"closures"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

type roadClosureResponse struct {
	Closure roadClosureRecord `json:"closure"`
}

type createRoadClosureRequest struct {
	RoadName   string             `json:"roadName"`
	Reason     string             `json:"reason,omitempty"`
	Status     string             `json:"status"`
	Severity   string             `json:"severity"`
	Source     string             `json:"source,omitempty"`
	SourceRef  string             `json:"sourceRef,omitempty"`
	Geometry   lineStringGeometry `json:"geometry"`
	ValidFrom  *time.Time         `json:"validFrom,omitempty"`
	ValidTo    *time.Time         `json:"validTo,omitempty"`
	DetourNote string             `json:"detourNote,omitempty"`
}

type updateRoadClosureRequest struct {
	RoadName   string              `json:"roadName,omitempty"`
	Reason     string              `json:"reason,omitempty"`
	Status     string              `json:"status,omitempty"`
	Severity   string              `json:"severity,omitempty"`
	Source     string              `json:"source,omitempty"`
	SourceRef  string              `json:"sourceRef,omitempty"`
	Geometry   *lineStringGeometry `json:"geometry,omitempty"`
	ValidFrom  *time.Time          `json:"validFrom,omitempty"`
	ValidTo    *time.Time          `json:"validTo,omitempty"`
	DetourNote string              `json:"detourNote,omitempty"`
}

type adapterImportRequest struct {
	Source    string     `json:"source"`
	SourceRef string     `json:"sourceRef,omitempty"`
	RoadName  string     `json:"roadName"`
	Status    string     `json:"status"`
	Reason    string     `json:"reason,omitempty"`
	Geometry  string     `json:"geometry"`
	ValidFrom time.Time  `json:"validFrom"`
	ValidTo   *time.Time `json:"validTo,omitempty"`
	Detour    string     `json:"detour,omitempty"`
}

type adapterImportResponse struct {
	Imported    int                 `json:"imported"`
	Closures    []roadClosureRecord `json:"closures"`
	GeneratedAt time.Time           `json:"generatedAt"`
	Source      string              `json:"source"`
}

type authorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var closureUpdateRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

var allowedClosureStatuses = map[string]bool{
	"active":    true,
	"scheduled": true,
	"lifted":    true,
	"cancelled": true,
}

var allowedClosureSeverities = map[string]bool{
	"low":       true,
	"moderate":  true,
	"high":      true,
	"severe":    true,
	"emergency": true,
}

const (
	earthRadiusMeters  = 6371000.0
	nearbySearchMeters = 30000.0
	defaultLimit       = 50
)

func main() {
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/road-closures", srv.listRoadClosuresHandler)
	mux.HandleFunc("POST /api/v1/road-closures", srv.createRoadClosureHandler)
	mux.HandleFunc("PATCH /api/v1/road-closures/{id}", srv.updateRoadClosureHandler)
	mux.HandleFunc("POST /api/v1/road-closures/imports/adapter", srv.importAdapterHandler)

	addr := envOrDefault("NADAA_ROAD_CLOSURE_ADDR", ":8095")
	log.Printf("road-closure-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServer() *server {
	now := time.Now
	return &server{store: newMemoryStore(now().UTC()), now: now}
}

func newMemoryStore(now time.Time) *memoryStore {
	return &memoryStore{
		closureCounter: 2,
		closures: []roadClosureRecord{
			{
				ID:        "road_closure_001",
				RoadName:  "Accra New Town Road",
				Reason:    "Flooding",
				Status:    "active",
				Severity:  "high",
				Source:    "manual",
				SourceRef: "nadmo-accra-ops",
				Geometry: lineStringGeometry{
					Type: "LineString",
					Coordinates: [][]float64{
						{-0.205, 5.570},
						{-0.190, 5.580},
					},
				},
				ValidFrom:  now.Add(-2 * time.Hour),
				ValidTo:    timePtr(now.Add(6 * time.Hour)),
				DetourNote: "Use Kanda Highway",
				CreatedBy:  "usr_dispatcher_001",
				UpdatedBy:  "usr_dispatcher_001",
				CreatedAt:  now.Add(-2 * time.Hour),
				UpdatedAt:  now.Add(-2 * time.Hour),
			},
			{
				ID:        "road_closure_002",
				RoadName:  "Kaneshie Market Road",
				Reason:    "Debris and flood water",
				Status:    "active",
				Severity:  "severe",
				Source:    "manual",
				SourceRef: "ghana-police",
				Geometry: lineStringGeometry{
					Type: "LineString",
					Coordinates: [][]float64{
						{-0.248, 5.566},
						{-0.240, 5.568},
						{-0.235, 5.564},
					},
				},
				ValidFrom:  now.Add(-1 * time.Hour),
				ValidTo:    timePtr(now.Add(12 * time.Hour)),
				DetourNote: "Use Mallam Road",
				CreatedBy:  "usr_dispatcher_002",
				UpdatedBy:  "usr_dispatcher_002",
				CreatedAt:  now.Add(-1 * time.Hour),
				UpdatedAt:  now.Add(-1 * time.Hour),
			},
			{
				ID:        "road_closure_003",
				RoadName:  "Labone Street",
				Reason:    "Scheduled drainage maintenance",
				Status:    "scheduled",
				Severity:  "low",
				Source:    "manual",
				SourceRef: "district-accra",
				Geometry: lineStringGeometry{
					Type: "LineString",
					Coordinates: [][]float64{
						{-0.183, 5.553},
						{-0.178, 5.555},
					},
				},
				ValidFrom:  now.Add(24 * time.Hour),
				ValidTo:    timePtr(now.Add(48 * time.Hour)),
				DetourNote: "Use Cantonments Road",
				CreatedBy:  "usr_district_officer_001",
				UpdatedBy:  "usr_district_officer_001",
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "road-closure-service"})
}

func (s *server) listRoadClosuresHandler(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseListFilter(w, r)
	if !ok {
		return
	}
	closures := s.store.listClosures(filter, s.now().UTC())
	log.Printf("INFO road-closure-service closure_list count=%d status=%s hasLocation=%t bbox=%t", len(closures), filter.Status, filter.Location != nil, filter.BBox != nil)
	writeJSON(w, http.StatusOK, roadClosureListResponse{Closures: closures, GeneratedAt: s.now().UTC()})
}

func (s *server) createRoadClosureHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, closureUpdateRoles)
	if !ok {
		return
	}

	var request createRoadClosureRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN road-closure-service create_closure invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreate(request, s.now().UTC())
	if code != "" {
		log.Printf("WARN road-closure-service create_closure validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	closure := s.store.createClosure(normalized, ctx, s.now().UTC())
	log.Printf("INFO road-closure-service create_closure completed id=%s actor=%s source=%s", closure.ID, ctx.ActorUserID, closure.Source)
	writeJSON(w, http.StatusCreated, roadClosureResponse{Closure: closure})
}

func (s *server) updateRoadClosureHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, closureUpdateRoles)
	if !ok {
		return
	}

	var request updateRoadClosureRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN road-closure-service update_closure invalid_json id=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdate(request)
	if code != "" {
		log.Printf("WARN road-closure-service update_closure validation_failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	closure, code, message := s.store.updateClosure(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN road-closure-service update_closure failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO road-closure-service update_closure completed id=%s actor=%s status=%s", closure.ID, ctx.ActorUserID, closure.Status)
	writeJSON(w, http.StatusOK, roadClosureResponse{Closure: closure})
}

func (s *server) importAdapterHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, closureUpdateRoles)
	if !ok {
		return
	}

	var request adapterImportRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN road-closure-service adapter_import invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAdapterImport(request)
	if code != "" {
		log.Printf("WARN road-closure-service adapter_import validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	closures := s.store.importAdapter(normalized, ctx, s.now().UTC())
	log.Printf("INFO road-closure-service adapter_import completed actor=%s source=%s imported=%d", ctx.ActorUserID, normalized.Source, len(closures))
	writeJSON(w, http.StatusOK, adapterImportResponse{
		Imported:    len(closures),
		Closures:    closures,
		GeneratedAt: s.now().UTC(),
		Source:      normalized.Source,
	})
}

type listFilter struct {
	Status        string
	Location      *coordinates
	RadiusMeters  float64
	BBox          *bbox
	Limit         int
	IncludeExpired bool
}

type bbox struct {
	MinLat float64
	MinLng float64
	MaxLat float64
	MaxLng float64
}

func (m *memoryStore) listClosures(filter listFilter, now time.Time) []roadClosureRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]roadClosureRecord, 0, len(m.closures))
	for _, closure := range m.closures {
		if filter.Status != "" && closure.Status != filter.Status {
			continue
		}
		if (filter.Status == "" || filter.Status == "active") && !filter.IncludeExpired && !isClosureEffective(closure, now) {
			continue
		}
		if filter.BBox != nil && !closureIntersectsBBox(closure.Geometry, *filter.BBox) {
			continue
		}
		if filter.Location != nil {
			closure.DistanceMeters = int(math.Round(minDistanceToLineString(*filter.Location, closure.Geometry)))
			if float64(closure.DistanceMeters) > filter.RadiusMeters {
				continue
			}
		}
		results = append(results, closure)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Status != results[j].Status {
			return statusRank(results[i].Status) < statusRank(results[j].Status)
		}
		if results[i].Severity != results[j].Severity {
			return severityRank(results[i].Severity) < severityRank(results[j].Severity)
		}
		if filter.Location != nil && results[i].DistanceMeters != results[j].DistanceMeters {
			return results[i].DistanceMeters < results[j].DistanceMeters
		}
		return results[i].RoadName < results[j].RoadName
	})

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}
	return copyClosures(results)
}

func (m *memoryStore) createClosure(request createRoadClosureRequest, ctx authorityContext, now time.Time) roadClosureRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closureCounter++
	closure := roadClosureRecord{
		ID:         fmt.Sprintf("road_closure_%03d", m.closureCounter),
		RoadName:   request.RoadName,
		Reason:     request.Reason,
		Status:     request.Status,
		Severity:   request.Severity,
		Source:     request.Source,
		SourceRef:  request.SourceRef,
		Geometry:   request.Geometry,
		ValidFrom:  now,
		ValidTo:    request.ValidTo,
		DetourNote: request.DetourNote,
		CreatedBy:  ctx.ActorUserID,
		UpdatedBy:  ctx.ActorUserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if request.ValidFrom != nil {
		closure.ValidFrom = *request.ValidFrom
	}
	m.closures = append(m.closures, closure)
	return closure
}

func (m *memoryStore) updateClosure(id string, request updateRoadClosureRequest, ctx authorityContext, now time.Time) (roadClosureRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.closures {
		if m.closures[index].ID != id {
			continue
		}
		next := m.closures[index]
		if request.RoadName != "" {
			next.RoadName = request.RoadName
		}
		if request.Reason != "" {
			next.Reason = request.Reason
		}
		if request.Status != "" {
			next.Status = request.Status
		}
		if request.Severity != "" {
			next.Severity = request.Severity
		}
		if request.Source != "" {
			next.Source = request.Source
		}
		if request.SourceRef != "" {
			next.SourceRef = request.SourceRef
		}
		if request.Geometry != nil {
			next.Geometry = *request.Geometry
		}
		if request.ValidFrom != nil {
			next.ValidFrom = *request.ValidFrom
		}
		if request.ValidTo != nil {
			next.ValidTo = request.ValidTo
		}
		if request.DetourNote != "" {
			next.DetourNote = request.DetourNote
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		m.closures[index] = next
		return next, "", ""
	}
	return roadClosureRecord{}, "not_found", "road closure was not found"
}

func (m *memoryStore) importAdapter(request adapterImportRequest, ctx authorityContext, now time.Time) []roadClosureRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	geometry, err := parseWKTLineString(request.Geometry)
	if err != nil {
		return nil
	}

	m.closureCounter++
	closure := roadClosureRecord{
		ID:         fmt.Sprintf("road_closure_%03d", m.closureCounter),
		RoadName:   request.RoadName,
		Reason:     request.Reason,
		Status:     request.Status,
		Severity:   severityFromStatus(request.Status),
		Source:     request.Source,
		SourceRef:  request.SourceRef,
		Geometry:   geometry,
		ValidFrom:  request.ValidFrom,
		ValidTo:    request.ValidTo,
		DetourNote: request.Detour,
		CreatedBy:  ctx.ActorUserID,
		UpdatedBy:  ctx.ActorUserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	m.closures = append(m.closures, closure)
	return []roadClosureRecord{closure}
}

func parseListFilter(w http.ResponseWriter, r *http.Request) (listFilter, bool) {
	filter := listFilter{
		Status:         strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status"))),
		RadiusMeters:   nearbySearchMeters,
		Limit:          defaultLimit,
		IncludeExpired: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("includeExpired"))) == "true",
	}

	if filter.Status != "" && !allowedClosureStatuses[filter.Status] {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be active, scheduled, lifted, or cancelled")
		return filter, false
	}

	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 100")
			return filter, false
		}
		filter.Limit = parsed
	}

	if value := strings.TrimSpace(r.URL.Query().Get("radius")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || parsed <= 0 || parsed > 100000 {
			writeError(w, http.StatusBadRequest, "invalid_radius", "radius must be between 1 and 100000 meters")
			return filter, false
		}
		filter.RadiusMeters = parsed
	}

	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText != "" || lngText != "" {
		if latText == "" || lngText == "" {
			writeError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng must be supplied together")
			return filter, false
		}
		lat, latErr := strconv.ParseFloat(latText, 64)
		lng, lngErr := strconv.ParseFloat(lngText, 64)
		if latErr != nil || lngErr != nil {
			writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
			return filter, false
		}
		loc := coordinates{Lat: lat, Lng: lng}
		if !validCoordinates(loc) {
			writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
			return filter, false
		}
		filter.Location = &loc
	}

	if value := strings.TrimSpace(r.URL.Query().Get("bbox")); value != "" {
		parts := strings.Split(value, ",")
		if len(parts) != 4 {
			writeError(w, http.StatusBadRequest, "invalid_bbox", "bbox must be minLng,minLat,maxLng,maxLat")
			return filter, false
		}
		var floats [4]float64
		for i := 0; i < 4; i++ {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_bbox", "bbox values must be valid decimal coordinates")
				return filter, false
			}
			floats[i] = parsed
		}
		filter.BBox = &bbox{MinLng: floats[0], MinLat: floats[1], MaxLng: floats[2], MaxLat: floats[3]}
	}

	return filter, true
}

func normalizeCreate(request createRoadClosureRequest, now time.Time) (createRoadClosureRequest, string, string) {
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Status = strings.TrimSpace(strings.ToLower(request.Status))
	request.Severity = strings.TrimSpace(strings.ToLower(request.Severity))
	request.Source = strings.TrimSpace(strings.ToLower(request.Source))
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.DetourNote = strings.TrimSpace(request.DetourNote)

	if request.RoadName == "" || len(request.RoadName) > 200 || unsafeText(request.RoadName) {
		return request, "invalid_road_name", "roadName is required and must be 200 safe characters or fewer"
	}
	if request.Status == "" {
		request.Status = "active"
	}
	if !allowedClosureStatuses[request.Status] {
		return request, "invalid_status", "status must be active, scheduled, lifted, or cancelled"
	}
	if request.Severity == "" {
		request.Severity = severityFromStatus(request.Status)
	}
	if !allowedClosureSeverities[request.Severity] {
		return request, "invalid_severity", "severity must be low, moderate, high, severe, or emergency"
	}
	if errCode, errMsg := validateGeometry(request.Geometry); errCode != "" {
		return request, errCode, errMsg
	}
	if request.Source == "" {
		request.Source = "manual"
	}
	if len(request.Source) > 80 || unsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || unsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	if len(request.Reason) > 200 || unsafeText(request.Reason) {
		return request, "invalid_reason", "reason must be 200 safe characters or fewer"
	}
	if len(request.DetourNote) > 500 || unsafeText(request.DetourNote) {
		return request, "invalid_detour_note", "detourNote must be 500 safe characters or fewer"
	}
	if request.ValidFrom != nil && request.ValidTo != nil && request.ValidTo.Before(*request.ValidFrom) {
		return request, "invalid_valid_to", "validTo must be after validFrom"
	}
	return request, "", ""
}

func normalizeUpdate(request updateRoadClosureRequest) (updateRoadClosureRequest, string, string) {
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Status = strings.TrimSpace(strings.ToLower(request.Status))
	request.Severity = strings.TrimSpace(strings.ToLower(request.Severity))
	request.Source = strings.TrimSpace(strings.ToLower(request.Source))
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.DetourNote = strings.TrimSpace(request.DetourNote)

	if request.RoadName != "" && (len(request.RoadName) > 200 || unsafeText(request.RoadName)) {
		return request, "invalid_road_name", "roadName must be 200 safe characters or fewer"
	}
	if request.Status != "" && !allowedClosureStatuses[request.Status] {
		return request, "invalid_status", "status must be active, scheduled, lifted, or cancelled"
	}
	if request.Severity != "" && !allowedClosureSeverities[request.Severity] {
		return request, "invalid_severity", "severity must be low, moderate, high, severe, or emergency"
	}
	if request.Geometry != nil {
		if errCode, errMsg := validateGeometry(*request.Geometry); errCode != "" {
			return request, errCode, errMsg
		}
	}
	if request.Source != "" && (len(request.Source) > 80 || unsafeText(request.Source)) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if request.SourceRef != "" && (len(request.SourceRef) > 120 || unsafeText(request.SourceRef)) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	if len(request.Reason) > 200 || unsafeText(request.Reason) {
		return request, "invalid_reason", "reason must be 200 safe characters or fewer"
	}
	if len(request.DetourNote) > 500 || unsafeText(request.DetourNote) {
		return request, "invalid_detour_note", "detourNote must be 500 safe characters or fewer"
	}
	if request.ValidFrom != nil && request.ValidTo != nil && request.ValidTo.Before(*request.ValidFrom) {
		return request, "invalid_valid_to", "validTo must be after validFrom"
	}
	if request.RoadName == "" && request.Reason == "" && request.Status == "" && request.Severity == "" &&
		request.Source == "" && request.SourceRef == "" && request.Geometry == nil &&
		request.ValidFrom == nil && request.ValidTo == nil && request.DetourNote == "" {
		return request, "no_changes", "at least one closure field must be supplied"
	}
	return request, "", ""
}

func normalizeAdapterImport(request adapterImportRequest) (adapterImportRequest, string, string) {
	request.Source = strings.TrimSpace(strings.ToLower(request.Source))
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Status = strings.TrimSpace(strings.ToLower(request.Status))
	request.Reason = strings.TrimSpace(request.Reason)
	request.Detour = strings.TrimSpace(request.Detour)

	if request.Source == "" {
		request.Source = "adapter"
	}
	if len(request.Source) > 80 || unsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || unsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	if request.RoadName == "" || len(request.RoadName) > 200 || unsafeText(request.RoadName) {
		return request, "invalid_road_name", "roadName is required and must be 200 safe characters or fewer"
	}
	if !allowedClosureStatuses[request.Status] {
		return request, "invalid_status", "status must be active, scheduled, lifted, or cancelled"
	}
	geometry, err := parseWKTLineString(request.Geometry)
	if err != nil {
		return request, "invalid_geometry", err.Error()
	}
	if errCode, errMsg := validateGeometry(geometry); errCode != "" {
		return request, errCode, errMsg
	}
	request.Geometry = formatWKTLineString(geometry)
	if len(request.Reason) > 200 || unsafeText(request.Reason) {
		return request, "invalid_reason", "reason must be 200 safe characters or fewer"
	}
	if len(request.Detour) > 500 || unsafeText(request.Detour) {
		return request, "invalid_detour", "detour must be 500 safe characters or fewer"
	}
	if request.ValidFrom.IsZero() {
		return request, "missing_valid_from", "validFrom is required"
	}
	if request.ValidTo != nil && request.ValidTo.Before(request.ValidFrom) {
		return request, "invalid_valid_to", "validTo must be after validFrom"
	}
	return request, "", ""
}

func validateGeometry(geometry lineStringGeometry) (string, string) {
	if geometry.Type != "LineString" {
		return "invalid_geometry_type", "geometry type must be LineString"
	}
	if len(geometry.Coordinates) < 2 {
		return "invalid_geometry", "LineString must contain at least two coordinates"
	}
	for _, point := range geometry.Coordinates {
		if len(point) != 2 {
			return "invalid_geometry", "each LineString coordinate must be [lng, lat]"
		}
		if point[0] < -180 || point[0] > 180 || point[1] < -90 || point[1] > 90 {
			return "invalid_geometry", "coordinates must be valid WGS84 longitude and latitude"
		}
	}
	return "", ""
}

func parseWKTLineString(value string) (lineStringGeometry, error) {
	value = strings.TrimSpace(value)
	prefix := "LINESTRING("
	suffix := ")"
	if !strings.HasPrefix(strings.ToUpper(value), prefix) || !strings.HasSuffix(value, suffix) {
		return lineStringGeometry{}, fmt.Errorf("geometry must be a LINESTRING(...)")
	}
	inner := value[len(prefix) : len(value)-len(suffix)]
	parts := strings.Split(inner, ",")
	coords := make([][]float64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		nums := strings.Fields(part)
		if len(nums) != 2 {
			return lineStringGeometry{}, fmt.Errorf("each LINESTRING coordinate must have two numbers")
		}
		lng, err1 := strconv.ParseFloat(nums[0], 64)
		lat, err2 := strconv.ParseFloat(nums[1], 64)
		if err1 != nil || err2 != nil {
			return lineStringGeometry{}, fmt.Errorf("LINESTRING coordinates must be valid decimal numbers")
		}
		coords = append(coords, []float64{lng, lat})
	}
	return lineStringGeometry{Type: "LineString", Coordinates: coords}, nil
}

func formatWKTLineString(geometry lineStringGeometry) string {
	parts := make([]string, 0, len(geometry.Coordinates))
	for _, c := range geometry.Coordinates {
		parts = append(parts, fmt.Sprintf("%f %f", c[0], c[1]))
	}
	return "LINESTRING(" + strings.Join(parts, ", ") + ")"
}

func severityFromStatus(status string) string {
	switch status {
	case "active":
		return "high"
	case "scheduled":
		return "low"
	case "lifted", "cancelled":
		return "low"
	default:
		return "moderate"
	}
}

func isClosureEffective(closure roadClosureRecord, now time.Time) bool {
	if closure.Status == "lifted" || closure.Status == "cancelled" {
		return false
	}
	if closure.ValidFrom.After(now) {
		return false
	}
	if closure.ValidTo != nil && closure.ValidTo.Before(now) {
		return false
	}
	return true
}

func closureIntersectsBBox(geometry lineStringGeometry, box bbox) bool {
	for _, point := range geometry.Coordinates {
		lng, lat := point[0], point[1]
		if lng >= box.MinLng && lng <= box.MaxLng && lat >= box.MinLat && lat <= box.MaxLat {
			return true
		}
	}
	return false
}

func minDistanceToLineString(location coordinates, geometry lineStringGeometry) float64 {
	minDistance := math.MaxFloat64
	for _, point := range geometry.Coordinates {
		d := distanceMeters(location, coordinates{Lat: point[1], Lng: point[0]})
		if d < minDistance {
			minDistance = d
		}
	}
	return minDistance
}

func statusRank(status string) int {
	switch status {
	case "active":
		return 0
	case "scheduled":
		return 1
	case "lifted":
		return 2
	case "cancelled":
		return 3
	default:
		return 4
	}
}

func severityRank(severity string) int {
	switch severity {
	case "emergency":
		return 0
	case "severe":
		return 1
	case "high":
		return 2
	case "moderate":
		return 3
	case "low":
		return 4
	default:
		return 5
	}
}

func copyClosures(source []roadClosureRecord) []roadClosureRecord {
	closures := make([]roadClosureRecord, 0, len(source))
	for _, closure := range source {
		closure.Geometry.Coordinates = append([][]float64{}, closure.Geometry.Coordinates...)
		closures = append(closures, closure)
	}
	return closures
}

func requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (authorityContext, bool) {
	ctx := authorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		writeError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return authorityContext{}, false
	}
	if !ctx.MFACompleted {
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for road closure updates")
		return authorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to update road closures")
		return authorityContext{}, false
	}
	return ctx, true
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR road-closure-service write_json_response_failed error=%v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsFromEnv()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Cache-Control", "no-store")
}

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NADAA-Actor-ID, X-NADAA-Actor-Role, X-NADAA-Agency-ID, X-NADAA-MFA-Completed, X-NADAA-Request-ID")
}

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}
	allowed := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func statusForCode(code string) int {
	if code == "not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

func distanceMeters(a coordinates, b coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)
	h := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func unsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func timePtr(t time.Time) *time.Time {
	return &t
}
