package main

import (
	"encoding/json"
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
	mu       sync.RWMutex
	shelters []shelterRecord
	recovery []recoverySupportLocation
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type shelterRecord struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Type             string      `json:"type"`
	Region           string      `json:"region"`
	District         string      `json:"district"`
	Address          string      `json:"address"`
	Location         coordinates `json:"location"`
	Capacity         int         `json:"capacity"`
	CurrentOccupancy int         `json:"currentOccupancy"`
	Status           string      `json:"status"`
	Contact          string      `json:"contact"`
	Facilities       []string    `json:"facilities"`
	Notes            string      `json:"notes,omitempty"`
	DistanceMeters   int         `json:"distanceMeters,omitempty"`
	UpdatedBy        string      `json:"updatedBy,omitempty"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

type recoverySupportLocation struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Region         string      `json:"region"`
	District       string      `json:"district"`
	Address        string      `json:"address"`
	Location       coordinates `json:"location"`
	Contact        string      `json:"contact"`
	Services       []string    `json:"services"`
	Hours          string      `json:"hours"`
	Status         string      `json:"status"`
	DistanceMeters int         `json:"distanceMeters,omitempty"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type shelterListResponse struct {
	Shelters    []shelterRecord `json:"shelters"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

type nearbyShelterResponse struct {
	Shelters        []shelterRecord           `json:"shelters"`
	RecoverySupport []recoverySupportLocation `json:"recoverySupport"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
}

type recoverySupportResponse struct {
	RecoverySupport []recoverySupportLocation `json:"recoverySupport"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
}

type occupancyUpdateRequest struct {
	Capacity         *int   `json:"capacity,omitempty"`
	CurrentOccupancy *int   `json:"currentOccupancy,omitempty"`
	Status           string `json:"status,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

type shelterUpdateResponse struct {
	Shelter shelterRecord `json:"shelter"`
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

var shelterUpdateRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

var allowedShelterStatuses = map[string]bool{
	"open":    true,
	"full":    true,
	"closed":  true,
	"unknown": true,
}

const (
	earthRadiusMeters  = 6371000.0
	nearbySearchMeters = 30000.0
	defaultNearbyLimit = 6
)

func main() {
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/shelters", srv.listSheltersHandler)
	mux.HandleFunc("GET /api/v1/shelters/nearby", srv.nearbySheltersHandler)
	mux.HandleFunc("GET /api/v1/recovery-support/nearby", srv.nearbyRecoverySupportHandler)
	mux.HandleFunc("PATCH /api/v1/shelters/{id}/occupancy", srv.updateShelterOccupancyHandler)

	addr := envOrDefault("NADAA_SHELTER_ADDR", ":8093")
	log.Printf("shelter-service listening on %s", addr)
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
		shelters: []shelterRecord{
			{
				ID:               "00000000-0000-0000-0000-000000000301",
				Name:             "Accra Metro Assembly Shelter",
				Type:             "evacuation_shelter",
				Region:           "Greater Accra",
				District:         "Accra Metropolitan",
				Address:          "Accra Metropolitan Assembly Hall",
				Location:         coordinates{Lat: 5.560, Lng: -0.200},
				Capacity:         450,
				CurrentOccupancy: 116,
				Status:           "open",
				Contact:          "112",
				Facilities:       []string{"water", "first_aid", "accessible_entry", "family_area"},
				Notes:            "Primary flood evacuation shelter for central Accra.",
				UpdatedAt:        now,
			},
			{
				ID:               "00000000-0000-0000-0000-000000000302",
				Name:             "Osu Community Hall",
				Type:             "temporary_shelter",
				Region:           "Greater Accra",
				District:         "Korle Klottey",
				Address:          "Osu Community Hall",
				Location:         coordinates{Lat: 5.550, Lng: -0.180},
				Capacity:         220,
				CurrentOccupancy: 34,
				Status:           "open",
				Contact:          "112",
				Facilities:       []string{"water", "first_aid", "family_area"},
				Notes:            "Suitable for short-term shelter and reunification.",
				UpdatedAt:        now,
			},
			{
				ID:               "00000000-0000-0000-0000-000000000303",
				Name:             "Kaneshie Social Centre",
				Type:             "relief_shelter",
				Region:           "Greater Accra",
				District:         "Okaikwei South",
				Address:          "Kaneshie Market Road",
				Location:         coordinates{Lat: 5.566, Lng: -0.242},
				Capacity:         180,
				CurrentOccupancy: 180,
				Status:           "full",
				Contact:          "112",
				Facilities:       []string{"water", "food_distribution"},
				Notes:            "At capacity; redirect new arrivals unless occupancy changes.",
				UpdatedAt:        now,
			},
		},
		recovery: []recoverySupportLocation{
			{
				ID:        "recovery_ama_relief_001",
				Name:      "AMA Relief Distribution Point",
				Type:      "relief_point",
				Region:    "Greater Accra",
				District:  "Accra Metropolitan",
				Address:   "Independence Avenue recovery desk",
				Location:  coordinates{Lat: 5.558, Lng: -0.197},
				Contact:   "112",
				Services:  []string{"food", "water", "blankets", "family_reunification"},
				Hours:     "08:00-20:00",
				Status:    "open",
				UpdatedAt: now,
			},
			{
				ID:        "recovery_korle_bu_medical_001",
				Name:      "Korle Bu Emergency Stabilization Desk",
				Type:      "medical_support",
				Region:    "Greater Accra",
				District:  "Accra Metropolitan",
				Address:   "Korle Bu emergency entrance",
				Location:  coordinates{Lat: 5.536, Lng: -0.227},
				Contact:   "112",
				Services:  []string{"first_aid", "triage", "medical_referral"},
				Hours:     "24 hours",
				Status:    "open",
				UpdatedAt: now,
			},
			{
				ID:        "recovery_osu_registration_001",
				Name:      "Osu Recovery Registration Desk",
				Type:      "recovery_registration",
				Region:    "Greater Accra",
				District:  "Korle Klottey",
				Address:   "Osu Community Hall annex",
				Location:  coordinates{Lat: 5.551, Lng: -0.181},
				Contact:   "112",
				Services:  []string{"needs_registration", "damage_reporting", "case_follow_up"},
				Hours:     "08:00-18:00",
				Status:    "open",
				UpdatedAt: now,
			},
		},
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "shelter-service"})
}

func (s *server) listSheltersHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, shelterListResponse{Shelters: s.store.listShelters(), GeneratedAt: s.now().UTC()})
}

func (s *server) nearbySheltersHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, nearbyShelterResponse{
		Shelters:        s.store.nearbyShelters(location, defaultNearbyLimit),
		RecoverySupport: s.store.nearbyRecoverySupport(location, defaultNearbyLimit),
		GeneratedAt:     s.now().UTC(),
	})
}

func (s *server) nearbyRecoverySupportHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, recoverySupportResponse{
		RecoverySupport: s.store.nearbyRecoverySupport(location, defaultNearbyLimit),
		GeneratedAt:     s.now().UTC(),
	})
}

func (s *server) updateShelterOccupancyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, shelterUpdateRoles)
	if !ok {
		return
	}

	var request occupancyUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeOccupancyUpdate(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	shelter, code, message := s.store.updateShelter(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, shelterUpdateResponse{Shelter: shelter})
}

func (m *memoryStore) listShelters() []shelterRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shelters := copyShelters(m.shelters)
	sort.Slice(shelters, func(i, j int) bool {
		if shelters[i].Status == shelters[j].Status {
			return shelters[i].Name < shelters[j].Name
		}
		return shelterStatusRank(shelters[i].Status) < shelterStatusRank(shelters[j].Status)
	})
	return shelters
}

func (m *memoryStore) nearbyShelters(location coordinates, limit int) []shelterRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shelters := make([]shelterRecord, 0, len(m.shelters))
	for _, shelter := range m.shelters {
		shelter.DistanceMeters = int(math.Round(distanceMeters(location, shelter.Location)))
		if float64(shelter.DistanceMeters) <= nearbySearchMeters {
			shelters = append(shelters, shelter)
		}
	}

	sort.Slice(shelters, func(i, j int) bool {
		if shelters[i].DistanceMeters == shelters[j].DistanceMeters {
			return shelterStatusRank(shelters[i].Status) < shelterStatusRank(shelters[j].Status)
		}
		return shelters[i].DistanceMeters < shelters[j].DistanceMeters
	})
	if limit > 0 && len(shelters) > limit {
		shelters = shelters[:limit]
	}
	return copyShelters(shelters)
}

func (m *memoryStore) nearbyRecoverySupport(location coordinates, limit int) []recoverySupportLocation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locations := make([]recoverySupportLocation, 0, len(m.recovery))
	for _, item := range m.recovery {
		item.DistanceMeters = int(math.Round(distanceMeters(location, item.Location)))
		if float64(item.DistanceMeters) <= nearbySearchMeters {
			locations = append(locations, item)
		}
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].DistanceMeters < locations[j].DistanceMeters
	})
	if limit > 0 && len(locations) > limit {
		locations = locations[:limit]
	}
	return copyRecovery(locations)
}

func (m *memoryStore) updateShelter(id string, request occupancyUpdateRequest, ctx authorityContext, now time.Time) (shelterRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.shelters {
		if m.shelters[index].ID != id {
			continue
		}

		next := m.shelters[index]
		if request.Capacity != nil {
			next.Capacity = *request.Capacity
		}
		if request.CurrentOccupancy != nil {
			next.CurrentOccupancy = *request.CurrentOccupancy
		}
		if request.Status != "" {
			next.Status = request.Status
		} else {
			next.Status = statusForOccupancy(next.Capacity, next.CurrentOccupancy, next.Status)
		}
		if request.Notes != "" {
			next.Notes = request.Notes
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		m.shelters[index] = next
		return next, "", ""
	}

	return shelterRecord{}, "not_found", "shelter was not found"
}

func normalizeOccupancyUpdate(request occupancyUpdateRequest) (occupancyUpdateRequest, string, string) {
	request.Status = strings.TrimSpace(strings.ToLower(request.Status))
	request.Notes = strings.TrimSpace(request.Notes)

	if request.Capacity == nil && request.CurrentOccupancy == nil && request.Status == "" && request.Notes == "" {
		return request, "no_changes", "at least one occupancy field must be supplied"
	}
	if request.Capacity != nil && *request.Capacity < 0 {
		return request, "invalid_capacity", "capacity must be zero or greater"
	}
	if request.CurrentOccupancy != nil && *request.CurrentOccupancy < 0 {
		return request, "invalid_occupancy", "currentOccupancy must be zero or greater"
	}
	if request.Capacity != nil && request.CurrentOccupancy != nil && *request.CurrentOccupancy > *request.Capacity {
		return request, "invalid_occupancy", "currentOccupancy cannot exceed capacity"
	}
	if request.Status != "" && !allowedShelterStatuses[request.Status] {
		return request, "invalid_status", "status must be open, full, closed, or unknown"
	}
	if len(request.Notes) > 500 || unsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 500 safe characters or fewer"
	}
	return request, "", ""
}

func parseLocation(w http.ResponseWriter, r *http.Request) (coordinates, bool) {
	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" || lngText == "" {
		writeError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng query parameters are required")
		return coordinates{}, false
	}

	lat, latErr := strconv.ParseFloat(latText, 64)
	lng, lngErr := strconv.ParseFloat(lngText, 64)
	if latErr != nil || lngErr != nil {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
		return coordinates{}, false
	}

	location := coordinates{Lat: lat, Lng: lng}
	if !validCoordinates(location) {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
		return coordinates{}, false
	}
	return location, true
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
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for shelter updates")
		return authorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to update shelter capacity")
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
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NADAA-Actor-ID, X-NADAA-Actor-Role, X-NADAA-Agency-ID, X-NADAA-MFA-Completed, X-NADAA-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func statusForCode(code string) int {
	if code == "not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func copyShelters(source []shelterRecord) []shelterRecord {
	shelters := make([]shelterRecord, 0, len(source))
	for _, shelter := range source {
		shelter.Facilities = append([]string{}, shelter.Facilities...)
		shelters = append(shelters, shelter)
	}
	return shelters
}

func copyRecovery(source []recoverySupportLocation) []recoverySupportLocation {
	locations := make([]recoverySupportLocation, 0, len(source))
	for _, item := range source {
		item.Services = append([]string{}, item.Services...)
		locations = append(locations, item)
	}
	return locations
}

func shelterStatusRank(status string) int {
	switch status {
	case "open":
		return 0
	case "unknown":
		return 1
	case "full":
		return 2
	case "closed":
		return 3
	default:
		return 4
	}
}

func statusForOccupancy(capacity int, occupancy int, fallback string) string {
	if capacity > 0 && occupancy >= capacity {
		return "full"
	}
	if fallback == "full" && occupancy < capacity {
		return "open"
	}
	if fallback == "" {
		return "open"
	}
	return fallback
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
