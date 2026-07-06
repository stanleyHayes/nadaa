package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store       *memoryStore
	rateLimiter *rateLimiter
	now         func() time.Time
}

type memoryStore struct {
	mu        sync.RWMutex
	sequence  int
	incidents map[string]incidentRecord
}

type rateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	requests map[string][]time.Time
	now      func() time.Time
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type createIncidentRequest struct {
	Type               string       `json:"type"`
	Description        string       `json:"description"`
	Location           coordinates  `json:"location"`
	PeopleAffected     int          `json:"peopleAffected"`
	InjuriesReported   bool         `json:"injuriesReported"`
	Urgency            string       `json:"urgency"`
	Anonymous          bool         `json:"anonymous"`
	ContactPermission  bool         `json:"contactPermission"`
	AccessibilityNeeds string       `json:"accessibilityNeeds"`
	Media              []string     `json:"media"`
	Reporter           *reporterRef `json:"reporter,omitempty"`
}

type reporterRef struct {
	UserID string `json:"userId"`
	Phone  string `json:"phone,omitempty"`
}

type incidentRecord struct {
	ID                 string       `json:"id"`
	Reference          string       `json:"reference"`
	Type               string       `json:"type"`
	Severity           string       `json:"severity"`
	Status             string       `json:"status"`
	Description        string       `json:"description"`
	Location           coordinates  `json:"location"`
	PeopleAffected     int          `json:"peopleAffected"`
	InjuriesReported   bool         `json:"injuriesReported"`
	Urgency            string       `json:"urgency"`
	Anonymous          bool         `json:"anonymous"`
	ContactPermission  bool         `json:"contactPermission"`
	AccessibilityNeeds string       `json:"accessibilityNeeds,omitempty"`
	Media              []string     `json:"media"`
	PriorityReview     bool         `json:"priorityReview"`
	ReportedBy         *reporterRef `json:"reportedBy,omitempty"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
}

type createIncidentResponse struct {
	ID                  string   `json:"id"`
	Reference           string   `json:"reference"`
	Status              string   `json:"status"`
	Severity            string   `json:"severity"`
	PriorityReview      bool     `json:"priorityReview"`
	DuplicateCandidates []string `json:"duplicateCandidates"`
}

type incidentListResponse struct {
	Incidents []incidentRecord `json:"incidents"`
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	allowedHazards = map[string]bool{
		"flood":             true,
		"fire":              true,
		"road_crash":        true,
		"building_collapse": true,
		"medical_emergency": true,
		"security_incident": true,
		"disease_outbreak":  true,
		"electrical_hazard": true,
		"blocked_drain":     true,
		"landslide":         true,
		"marine_accident":   true,
		"storm":             true,
		"tidal_wave":        true,
		"other":             true,
	}
	allowedUrgencies = map[string]bool{
		"low":              true,
		"moderate":         true,
		"high":             true,
		"life_threatening": true,
	}
	mediaRefPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,128}$`)
)

func main() {
	srv := newServerFromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/incidents", srv.createIncidentHandler)
	mux.HandleFunc("GET /api/v1/incidents", srv.listIncidentsHandler)

	addr := envOrDefault("NADAA_INCIDENT_ADDR", ":8084")
	log.Printf("incident-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServerFromEnv() *server {
	limit := envIntOrDefault("NADAA_INCIDENT_RATE_LIMIT", 60)
	windowSeconds := envIntOrDefault("NADAA_INCIDENT_RATE_WINDOW_SECONDS", 60)
	now := time.Now
	return &server{
		store:       newMemoryStore(),
		rateLimiter: newRateLimiter(limit, time.Duration(windowSeconds)*time.Second, now),
		now:         now,
	}
}

func newMemoryStore() *memoryStore {
	return &memoryStore{incidents: map[string]incidentRecord{}}
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return &rateLimiter{limit: limit, window: window, requests: map[string][]time.Time{}, now: now}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "incident-service"})
}

func (s *server) createIncidentHandler(w http.ResponseWriter, r *http.Request) {
	clientID := clientIdentifier(r)
	if !s.rateLimiter.Allow(clientID) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many incident reports submitted; please wait before trying again")
		return
	}

	var request createIncidentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	normalized, err := normalizeIncidentRequest(request)
	if errors.Is(err, errValidation) {
		writeError(w, http.StatusBadRequest, validationCode(normalized), validationMessage(normalized))
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_incident", err.Error())
		return
	}

	record := s.store.createIncident(normalized, s.now())
	writeJSON(w, http.StatusCreated, createIncidentResponse{
		ID:                  record.ID,
		Reference:           record.Reference,
		Status:              record.Status,
		Severity:            record.Severity,
		PriorityReview:      record.PriorityReview,
		DuplicateCandidates: []string{},
	})
}

func (s *server) listIncidentsHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, incidentListResponse{Incidents: s.store.listIncidents()})
}

func (m *memoryStore) createIncident(request createIncidentRequest, now time.Time) incidentRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sequence++
	reference := fmt.Sprintf("INC-%06d", m.sequence)
	timestamp := now.UTC()
	record := incidentRecord{
		ID:                 newID("inc"),
		Reference:          reference,
		Type:               request.Type,
		Severity:           severityFromUrgency(request.Urgency),
		Status:             "reported",
		Description:        request.Description,
		Location:           request.Location,
		PeopleAffected:     request.PeopleAffected,
		InjuriesReported:   request.InjuriesReported,
		Urgency:            request.Urgency,
		Anonymous:          request.Anonymous,
		ContactPermission:  request.ContactPermission,
		AccessibilityNeeds: request.AccessibilityNeeds,
		Media:              append([]string{}, request.Media...),
		PriorityReview:     priorityReview(request),
		ReportedBy:         reportedByFor(request),
		CreatedAt:          timestamp,
		UpdatedAt:          timestamp,
	}
	m.incidents[record.ID] = record
	return record
}

func (m *memoryStore) listIncidents() []incidentRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incidents := make([]incidentRecord, 0, len(m.incidents))
	for _, incident := range m.incidents {
		incidents = append(incidents, incident)
	}
	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt.After(incidents[j].CreatedAt)
	})
	return incidents
}

func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	cutoff := now.Add(-r.window)
	events := r.requests[key]
	kept := events[:0]
	for _, event := range events {
		if event.After(cutoff) {
			kept = append(kept, event)
		}
	}

	if len(kept) >= r.limit {
		r.requests[key] = kept
		return false
	}

	kept = append(kept, now)
	r.requests[key] = kept
	return true
}

var errValidation = errors.New("validation failed")

func normalizeIncidentRequest(request createIncidentRequest) (createIncidentRequest, error) {
	request.Type = strings.TrimSpace(strings.ToLower(request.Type))
	request.Description = strings.TrimSpace(request.Description)
	request.Urgency = strings.TrimSpace(strings.ToLower(request.Urgency))
	request.AccessibilityNeeds = strings.TrimSpace(request.AccessibilityNeeds)

	if request.Urgency == "" {
		request.Urgency = "moderate"
	}

	if !allowedHazards[request.Type] {
		request.Type = "invalid_type"
		return request, errValidation
	}

	if len(request.Description) < 5 || len(request.Description) > 2000 || unsafeText(request.Description) {
		request.Description = "invalid_description"
		return request, errValidation
	}

	if !validCoordinates(request.Location) {
		request.Type = "invalid_location"
		return request, errValidation
	}

	if request.PeopleAffected < 0 || request.PeopleAffected > 1000000 {
		request.Type = "invalid_people_affected"
		return request, errValidation
	}

	if !allowedUrgencies[request.Urgency] {
		request.Type = "invalid_urgency"
		return request, errValidation
	}

	if request.Anonymous {
		request.Reporter = nil
	}

	if request.Reporter != nil {
		request.Reporter.UserID = strings.TrimSpace(request.Reporter.UserID)
		request.Reporter.Phone = strings.TrimSpace(request.Reporter.Phone)
		if request.Reporter.UserID == "" {
			request.Type = "invalid_reporter"
			return request, errValidation
		}
	}

	if len(request.Media) > 10 {
		request.Type = "too_many_media"
		return request, errValidation
	}
	for index, mediaRef := range request.Media {
		mediaRef = strings.TrimSpace(mediaRef)
		if !mediaRefPattern.MatchString(mediaRef) {
			request.Type = "invalid_media"
			return request, errValidation
		}
		request.Media[index] = mediaRef
	}

	if len(request.AccessibilityNeeds) > 500 || unsafeText(request.AccessibilityNeeds) {
		request.Type = "invalid_accessibility_needs"
		return request, errValidation
	}

	return request, nil
}

func validationCode(request createIncidentRequest) string {
	switch request.Type {
	case "invalid_type":
		return "unsupported_hazard"
	case "invalid_description":
		return "invalid_description"
	case "invalid_location":
		return "invalid_location"
	case "invalid_people_affected":
		return "invalid_people_affected"
	case "invalid_urgency":
		return "invalid_urgency"
	case "invalid_reporter":
		return "invalid_reporter"
	case "too_many_media":
		return "too_many_media"
	case "invalid_media":
		return "invalid_media"
	case "invalid_accessibility_needs":
		return "invalid_accessibility_needs"
	default:
		return "invalid_incident"
	}
}

func validationMessage(request createIncidentRequest) string {
	switch validationCode(request) {
	case "unsupported_hazard":
		return "type must be a supported hazard"
	case "invalid_description":
		return "description must be 5 to 2000 safe characters"
	case "invalid_location":
		return "location must contain valid lat and lng values"
	case "invalid_people_affected":
		return "peopleAffected must be between 0 and 1000000"
	case "invalid_urgency":
		return "urgency must be low, moderate, high, or life_threatening"
	case "invalid_reporter":
		return "reporter.userId is required when reporter is supplied"
	case "too_many_media":
		return "a report can reference at most 10 media items"
	case "invalid_media":
		return "media references must be 3 to 128 characters using letters, numbers, underscores, or dashes"
	case "invalid_accessibility_needs":
		return "accessibilityNeeds must be 500 safe characters or fewer"
	default:
		return "incident request is invalid"
	}
}

func severityFromUrgency(urgency string) string {
	switch urgency {
	case "low":
		return "low"
	case "high":
		return "high"
	case "life_threatening":
		return "emergency"
	default:
		return "moderate"
	}
}

func priorityReview(request createIncidentRequest) bool {
	return request.Urgency == "life_threatening" || request.InjuriesReported
}

func reportedByFor(request createIncidentRequest) *reporterRef {
	if request.Anonymous || request.Reporter == nil {
		return nil
	}

	reporter := *request.Reporter
	if !request.ContactPermission {
		reporter.Phone = ""
	}
	return &reporter
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

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

func unsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

func clientIdentifier(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}

func newID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, bytes)
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
