package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *memoryStore
}

type memoryStore struct {
	mu           sync.RWMutex
	contracts    []integrationContract
	observations []weatherHydrologyObservation
	syncEvents   []syncEvent
}

type integrationContract struct {
	ID                     string            `json:"id"`
	Partner                string            `json:"partner"`
	PartnerType            string            `json:"partnerType"`
	Domain                 string            `json:"domain"`
	Direction              string            `json:"direction"`
	DataOwner              string            `json:"dataOwner"`
	Cadence                string            `json:"cadence"`
	Authentication         authentication    `json:"authentication"`
	Payloads               []payloadContract `json:"payloads"`
	FailureBehavior        failureBehavior   `json:"failureBehavior"`
	SourceOfTruth          string            `json:"sourceOfTruth"`
	FreshnessWindowMinutes int               `json:"freshnessWindowMinutes"`
	ContactPoint           string            `json:"contactPoint"`
	Status                 string            `json:"status"`
	Notes                  string            `json:"notes"`
	UpdatedAt              time.Time         `json:"updatedAt"`
}

type authentication struct {
	Mode            string   `json:"mode"`
	RequiredHeaders []string `json:"requiredHeaders,omitempty"`
	SecretScope     string   `json:"secretScope,omitempty"`
}

type payloadContract struct {
	Name           string   `json:"name"`
	ContentType    string   `json:"contentType"`
	RequiredFields []string `json:"requiredFields"`
	OptionalFields []string `json:"optionalFields,omitempty"`
	PII            string   `json:"pii"`
	Geometry       string   `json:"geometry,omitempty"`
	ExampleRef     string   `json:"exampleRef"`
}

type failureBehavior struct {
	Retryable       bool   `json:"retryable"`
	MaxAttempts     int    `json:"maxAttempts"`
	BackoffSeconds  []int  `json:"backoffSeconds"`
	DeadLetterQueue string `json:"deadLetterQueue"`
	ManualFallback  string `json:"manualFallback"`
}

type weatherHydrologyObservation struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	Metric      string      `json:"metric"`
	Value       float64     `json:"value"`
	Unit        string      `json:"unit"`
	StationID   string      `json:"stationId"`
	Location    coordinates `json:"location"`
	ObservedAt  time.Time   `json:"observedAt"`
	ValidFrom   time.Time   `json:"validFrom"`
	ValidTo     time.Time   `json:"validTo"`
	Quality     string      `json:"quality"`
	GeneratedBy string      `json:"generatedBy"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type syncRequest struct {
	Type            string      `json:"type"`
	SourceID        string      `json:"sourceId"`
	Reference       string      `json:"reference"`
	HazardType      string      `json:"hazardType"`
	Status          string      `json:"status"`
	Severity        string      `json:"severity"`
	Title           string      `json:"title,omitempty"`
	Summary         string      `json:"summary,omitempty"`
	Message         string      `json:"message,omitempty"`
	Location        coordinates `json:"location,omitempty"`
	TargetLabel     string      `json:"targetLabel,omitempty"`
	TargetAgencyIDs []string    `json:"targetAgencyIds"`
	CorrelationID   string      `json:"correlationId"`
	OccurredAt      time.Time   `json:"occurredAt"`
}

type syncEvent struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	SourceID        string    `json:"sourceId"`
	Reference       string    `json:"reference"`
	TargetAgencyIDs []string  `json:"targetAgencyIds"`
	CorrelationID   string    `json:"correlationId"`
	Status          string    `json:"status"`
	AdapterID       string    `json:"adapterId"`
	QueuedAt        time.Time `json:"queuedAt"`
	Retryable       bool      `json:"retryable"`
}

type contractListResponse struct {
	Contracts []integrationContract `json:"contracts"`
}

type observationListResponse struct {
	Observations []weatherHydrologyObservation `json:"observations"`
}

type syncEventListResponse struct {
	Events []syncEvent `json:"events"`
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var allowedDirections = map[string]bool{
	"inbound":       true,
	"outbound":      true,
	"bidirectional": true,
}

var allowedDomains = map[string]bool{
	"weather":           true,
	"hydrology":         true,
	"incident_sync":     true,
	"alert_sync":        true,
	"road_closure":      true,
	"hospital_capacity": true,
	"utility_outage":    true,
	"shelter_status":    true,
}

func main() {
	srv := &server{store: newMemoryStore()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/integrations/contracts", srv.listContractsHandler)
	mux.HandleFunc("GET /api/v1/integrations/mock/weather-hydrology/observations", srv.listObservationsHandler)
	mux.HandleFunc("POST /api/v1/integrations/mock/sync-events", srv.createSyncEventHandler)
	mux.HandleFunc("GET /api/v1/integrations/mock/sync-events", srv.listSyncEventsHandler)

	addr := envOrDefault("NADAA_INTEGRATION_ADDR", ":8088")
	log.Printf("integration-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newMemoryStore() *memoryStore {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &memoryStore{
		contracts:    seedContracts(now),
		observations: seedObservations(now),
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "integration-service"})
}

func (s *server) listContractsHandler(w http.ResponseWriter, r *http.Request) {
	domain := normalizeQueryValue(r.URL.Query().Get("domain"))
	direction := normalizeQueryValue(r.URL.Query().Get("direction"))
	partner := normalizeQueryValue(r.URL.Query().Get("partner"))

	if domain != "" && !allowedDomains[domain] {
		writeError(w, http.StatusBadRequest, "invalid_domain", "domain must be a supported integration domain")
		return
	}
	if direction != "" && !allowedDirections[direction] {
		writeError(w, http.StatusBadRequest, "invalid_direction", "direction must be inbound, outbound, or bidirectional")
		return
	}

	writeJSON(w, http.StatusOK, contractListResponse{Contracts: s.store.listContracts(domain, direction, partner)})
}

func (s *server) listObservationsHandler(w http.ResponseWriter, r *http.Request) {
	source := normalizeQueryValue(r.URL.Query().Get("source"))
	metric := normalizeQueryValue(r.URL.Query().Get("metric"))
	if metric != "" && metric != "rainfall_mm" && metric != "water_level_m" {
		writeError(w, http.StatusBadRequest, "invalid_metric", "metric must be rainfall_mm or water_level_m")
		return
	}

	writeJSON(w, http.StatusOK, observationListResponse{Observations: s.store.listObservations(source, metric)})
}

func (s *server) createSyncEventHandler(w http.ResponseWriter, r *http.Request) {
	var request syncRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.Type = normalizeQueryValue(request.Type)

	if code, message := validateSyncRequest(request); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	event := s.store.createSyncEvent(request, time.Now().UTC())
	writeJSON(w, http.StatusAccepted, event)
}

func (s *server) listSyncEventsHandler(w http.ResponseWriter, r *http.Request) {
	eventType := normalizeQueryValue(r.URL.Query().Get("type"))
	if eventType != "" && eventType != "incident" && eventType != "alert" {
		writeError(w, http.StatusBadRequest, "invalid_type", "type must be incident or alert")
		return
	}

	writeJSON(w, http.StatusOK, syncEventListResponse{Events: s.store.listSyncEvents(eventType)})
}

func (m *memoryStore) listContracts(domain string, direction string, partner string) []integrationContract {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contracts := make([]integrationContract, 0, len(m.contracts))
	for _, contract := range m.contracts {
		if domain != "" && contract.Domain != domain {
			continue
		}
		if direction != "" && contract.Direction != direction {
			continue
		}
		if partner != "" && !strings.Contains(normalizeQueryValue(contract.Partner), partner) {
			continue
		}
		contracts = append(contracts, contract)
	}

	sort.Slice(contracts, func(i, j int) bool {
		if contracts[i].Domain == contracts[j].Domain {
			return contracts[i].Partner < contracts[j].Partner
		}
		return contracts[i].Domain < contracts[j].Domain
	})
	return contracts
}

func (m *memoryStore) listObservations(source string, metric string) []weatherHydrologyObservation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	observations := make([]weatherHydrologyObservation, 0, len(m.observations))
	for _, observation := range m.observations {
		if source != "" && observation.Source != source {
			continue
		}
		if metric != "" && observation.Metric != metric {
			continue
		}
		observations = append(observations, observation)
	}

	sort.Slice(observations, func(i, j int) bool {
		return observations[i].ObservedAt.Before(observations[j].ObservedAt)
	})
	return observations
}

func (m *memoryStore) createSyncEvent(request syncRequest, now time.Time) syncEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	event := syncEvent{
		ID:              "sync_" + sanitizeID(request.Type) + "_" + sanitizeID(request.SourceID),
		Type:            request.Type,
		SourceID:        request.SourceID,
		Reference:       request.Reference,
		TargetAgencyIDs: request.TargetAgencyIDs,
		CorrelationID:   request.CorrelationID,
		Status:          "accepted",
		AdapterID:       adapterIDForSyncType(request.Type),
		QueuedAt:        now,
		Retryable:       true,
	}
	m.syncEvents = append(m.syncEvents, event)
	return event
}

func (m *memoryStore) listSyncEvents(eventType string) []syncEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]syncEvent, 0, len(m.syncEvents))
	for _, event := range m.syncEvents {
		if eventType != "" && event.Type != eventType {
			continue
		}
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].QueuedAt.After(events[j].QueuedAt)
	})
	return events
}

func validateSyncRequest(request syncRequest) (string, string) {
	if request.Type != "incident" && request.Type != "alert" {
		return "invalid_type", "type must be incident or alert"
	}
	if strings.TrimSpace(request.SourceID) == "" {
		return "missing_source_id", "sourceId is required"
	}
	if strings.TrimSpace(request.Reference) == "" {
		return "missing_reference", "reference is required"
	}
	if strings.TrimSpace(request.HazardType) == "" {
		return "missing_hazard_type", "hazardType is required"
	}
	if len(request.TargetAgencyIDs) == 0 {
		return "missing_target_agencies", "at least one targetAgencyId is required"
	}
	if strings.TrimSpace(request.CorrelationID) == "" {
		return "missing_correlation_id", "correlationId is required for idempotent sync"
	}
	return "", ""
}

func seedContracts(now time.Time) []integrationContract {
	return []integrationContract{
		newContract("gmet-rainfall-nowcast", "Ghana Meteorological Agency", "meteorological", "weather", "inbound", "GMet", "Every 15 minutes during watch/warning periods", "api_key", []string{"X-NADAA-Source", "X-NADAA-Signature"}, []payloadContract{weatherPayload()}, "Imported observations keep source, observedAt, validFrom, validTo, stationId, and location.", now),
		newContract("hydro-water-level-feed", "Ghana Hydrological Authority", "hydrological", "hydrology", "inbound", "Ghana Hydrological Authority", "Every 15 minutes during rainy season, hourly otherwise", "mtls", []string{"X-NADAA-Source"}, []payloadContract{hydrologyPayload()}, "Water-level records remain owned by the originating hydrological authority.", now),
		newContract("nadmo-incident-sync", "NADMO National Operations", "nadmo", "incident_sync", "outbound", "NADAA platform operator", "Near real time on verification, assignment, and closure", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []payloadContract{incidentSyncPayload()}, "Manual dispatcher call and dashboard export remain the fallback if sync fails.", now),
		newContract("nadmo-alert-sync", "NADMO National Operations", "nadmo", "alert_sync", "outbound", "NADAA platform operator", "Near real time after human approval", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []payloadContract{alertSyncPayload()}, "Public alert publication must continue through approved NADAA workflow if partner sync fails.", now),
		newContract("police-road-closure-feed", "Ghana Police Service", "police", "road_closure", "bidirectional", "Ghana Police Service", "On change, with hourly reconciliation", "signed_webhook", []string{"X-NADAA-Source", "X-NADAA-Signature"}, []payloadContract{roadClosurePayload(), incidentSyncPayload()}, "Road closures imported from police remain source-attributed and reviewable before route use.", now),
		newContract("fire-incident-dispatch", "Ghana National Fire Service", "fire", "incident_sync", "outbound", "NADAA platform operator", "Near real time for fire, flood rescue, and electrical hazard assignments", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []payloadContract{incidentSyncPayload()}, "If webhook delivery fails, dispatcher contacts fire service through 112 and records the manual handoff.", now),
		newContract("ambulance-medical-dispatch", "National Ambulance Service", "ambulance", "incident_sync", "outbound", "NADAA platform operator", "Near real time for injury and medical emergency assignments", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []payloadContract{incidentSyncPayload()}, "If partner endpoint is down, dispatcher uses 112 and keeps incident status manual.", now),
		newContract("district-shelter-status", "District Assemblies", "district_assembly", "shelter_status", "bidirectional", "District Assembly or shelter operator", "Every 30 minutes during response, daily otherwise", "api_key", []string{"X-NADAA-Source"}, []payloadContract{shelterStatusPayload()}, "Shelter updates are advisory until confirmed by authorized district or NADMO users.", now),
		newContract("hospital-capacity-feed", "Hospitals And Health Facilities", "hospital", "hospital_capacity", "inbound", "Participating health facility", "Every 30 minutes during active incidents", "api_key", []string{"X-NADAA-Source"}, []payloadContract{hospitalCapacityPayload()}, "Capacity data is operationally sensitive and should be visible only to authorized responders.", now),
		newContract("utility-outage-feed", "Utilities And Power Providers", "utility", "utility_outage", "inbound", "Originating utility", "On change, with hourly reconciliation", "signed_webhook", []string{"X-NADAA-Source", "X-NADAA-Signature"}, []payloadContract{utilityOutagePayload()}, "Electrical hazard and outage imports never suppress citizen reports; they enrich dispatcher context.", now),
	}
}

func newContract(id string, partner string, partnerType string, domain string, direction string, dataOwner string, cadence string, authMode string, headers []string, payloads []payloadContract, notes string, now time.Time) integrationContract {
	return integrationContract{
		ID:                     id,
		Partner:                partner,
		PartnerType:            partnerType,
		Domain:                 domain,
		Direction:              direction,
		DataOwner:              dataOwner,
		Cadence:                cadence,
		Authentication:         authentication{Mode: authMode, RequiredHeaders: headers, SecretScope: "environment_secret_manager"},
		Payloads:               payloads,
		FailureBehavior:        standardFailureBehavior(domain),
		SourceOfTruth:          sourceOfTruth(direction),
		FreshnessWindowMinutes: freshnessWindow(domain),
		ContactPoint:           "integration-owner@nadaa.local",
		Status:                 "mock_contract",
		Notes:                  notes,
		UpdatedAt:              now,
	}
}

func standardFailureBehavior(domain string) failureBehavior {
	queue := "integration.dead_letter." + domain
	return failureBehavior{
		Retryable:       true,
		MaxAttempts:     5,
		BackoffSeconds:  []int{30, 120, 300, 900, 1800},
		DeadLetterQueue: queue,
		ManualFallback:  "Record failed exchange in import job logs and continue manual incident response or alert approval.",
	}
}

func sourceOfTruth(direction string) string {
	if direction == "inbound" {
		return "originating_partner"
	}
	if direction == "outbound" {
		return "nadaa"
	}
	return "field_specific"
}

func freshnessWindow(domain string) int {
	switch domain {
	case "weather", "hydrology":
		return 30
	case "incident_sync", "alert_sync", "utility_outage", "road_closure":
		return 5
	default:
		return 60
	}
}

func seedObservations(now time.Time) []weatherHydrologyObservation {
	return []weatherHydrologyObservation{
		newObservation("obs_gmet_accra_001", "gmet-accra-nowcast", "rainfall_mm", 34.2, "mm", "GHA-ACC-RAIN-001", coordinates{Lat: 5.6037, Lng: -0.1870}, now.Add(-15*time.Minute), now),
		newObservation("obs_gmet_accra_002", "gmet-accra-nowcast", "rainfall_mm", 42.8, "mm", "GHA-ACC-RAIN-002", coordinates{Lat: 5.5600, Lng: -0.2000}, now.Add(-10*time.Minute), now),
		newObservation("obs_hydro_odaw_001", "hydro-odaw-level", "water_level_m", 1.76, "m", "GHA-ODAW-LVL-001", coordinates{Lat: 5.5750, Lng: -0.2050}, now.Add(-12*time.Minute), now),
		newObservation("obs_hydro_korle_001", "hydro-korle-level", "water_level_m", 1.34, "m", "GHA-KORLE-LVL-001", coordinates{Lat: 5.5400, Lng: -0.2150}, now.Add(-9*time.Minute), now),
	}
}

func newObservation(id string, source string, metric string, value float64, unit string, stationID string, location coordinates, observedAt time.Time, now time.Time) weatherHydrologyObservation {
	return weatherHydrologyObservation{
		ID:          id,
		Source:      source,
		Metric:      metric,
		Value:       value,
		Unit:        unit,
		StationID:   stationID,
		Location:    location,
		ObservedAt:  observedAt,
		ValidFrom:   observedAt,
		ValidTo:     observedAt.Add(30 * time.Minute),
		Quality:     "fixture",
		GeneratedBy: "mock_adapter",
	}
}

func weatherPayload() payloadContract {
	return payloadContract{
		Name:           "WeatherObservation",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "observedAt", "validFrom", "validTo", "location.lat", "location.lng", "rainfallMm"},
		OptionalFields: []string{"stationId", "forecastWindowMinutes", "confidence"},
		PII:            "none",
		Geometry:       "Point WGS84",
		ExampleRef:     "docs/integrations.md#weather-observation",
	}
}

func hydrologyPayload() payloadContract {
	return payloadContract{
		Name:           "HydrologyObservation",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "observedAt", "validFrom", "validTo", "location.lat", "location.lng", "waterLevelM"},
		OptionalFields: []string{"stationId", "riverBasin", "thresholdLevelM"},
		PII:            "none",
		Geometry:       "Point WGS84",
		ExampleRef:     "docs/integrations.md#hydrology-observation",
	}
}

func incidentSyncPayload() payloadContract {
	return payloadContract{
		Name:           "IncidentSync",
		ContentType:    "application/json",
		RequiredFields: []string{"type", "sourceId", "reference", "hazardType", "status", "severity", "location", "targetAgencyIds", "correlationId"},
		OptionalFields: []string{"summary", "occurredAt", "mediaCount", "accessibilityNeeds"},
		PII:            "minimal_operational",
		Geometry:       "Point WGS84",
		ExampleRef:     "docs/integrations.md#incident-sync",
	}
}

func alertSyncPayload() payloadContract {
	return payloadContract{
		Name:           "AlertSync",
		ContentType:    "application/json",
		RequiredFields: []string{"type", "sourceId", "reference", "hazardType", "severity", "title", "message", "targetLabel", "targetAgencyIds", "correlationId"},
		OptionalFields: []string{"startsAt", "expiresAt", "recommendedAction", "evacuationRequired"},
		PII:            "none",
		Geometry:       "Target geometry summary or reference",
		ExampleRef:     "docs/integrations.md#alert-sync",
	}
}

func roadClosurePayload() payloadContract {
	return payloadContract{
		Name:           "RoadClosure",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "roadName", "status", "geometry", "validFrom"},
		OptionalFields: []string{"validTo", "reason", "detour"},
		PII:            "none",
		Geometry:       "LineString WGS84",
		ExampleRef:     "docs/integrations.md#road-closure",
	}
}

func shelterStatusPayload() payloadContract {
	return payloadContract{
		Name:           "ShelterStatus",
		ContentType:    "application/json",
		RequiredFields: []string{"shelterId", "status", "capacity", "currentOccupancy", "updatedAt"},
		OptionalFields: []string{"facilities", "contact", "needs"},
		PII:            "aggregate_only",
		Geometry:       "Point WGS84 or shelter reference",
		ExampleRef:     "docs/integrations.md#shelter-status",
	}
}

func hospitalCapacityPayload() payloadContract {
	return payloadContract{
		Name:           "HospitalCapacity",
		ContentType:    "application/json",
		RequiredFields: []string{"facilityId", "availableBeds", "emergencyCapacity", "updatedAt"},
		OptionalFields: []string{"traumaCapacity", "ambulanceBayStatus", "contact"},
		PII:            "aggregate_only",
		Geometry:       "Point WGS84 or facility reference",
		ExampleRef:     "docs/integrations.md#hospital-capacity",
	}
}

func utilityOutagePayload() payloadContract {
	return payloadContract{
		Name:           "UtilityOutage",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "utilityType", "status", "area", "validFrom"},
		OptionalFields: []string{"validTo", "hazardType", "customerImpactEstimate"},
		PII:            "none",
		Geometry:       "Polygon or MultiPolygon WGS84",
		ExampleRef:     "docs/integrations.md#utility-outage",
	}
}

func adapterIDForSyncType(eventType string) string {
	if eventType == "alert" {
		return "mock-alert-sync-adapter"
	}
	return "mock-incident-sync-adapter"
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
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func sanitizeID(value string) string {
	value = normalizeQueryValue(value)
	replacer := strings.NewReplacer(" ", "_", "/", "_", ":", "_")
	return replacer.Replace(value)
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
