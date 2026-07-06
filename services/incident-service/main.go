package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
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
	media     map[string]mediaRecord
	audit     []auditEvent
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
	ID                  string               `json:"id"`
	Reference           string               `json:"reference"`
	Type                string               `json:"type"`
	Severity            string               `json:"severity"`
	Status              string               `json:"status"`
	Description         string               `json:"description"`
	Location            coordinates          `json:"location"`
	PeopleAffected      int                  `json:"peopleAffected"`
	InjuriesReported    bool                 `json:"injuriesReported"`
	Urgency             string               `json:"urgency"`
	Anonymous           bool                 `json:"anonymous"`
	ContactPermission   bool                 `json:"contactPermission"`
	AccessibilityNeeds  string               `json:"accessibilityNeeds,omitempty"`
	Media               []string             `json:"media"`
	PriorityReview      bool                 `json:"priorityReview"`
	DuplicateCandidates []duplicateCandidate `json:"duplicateCandidates"`
	MergedIncidentIDs   []string             `json:"mergedIncidentIds"`
	ReportedBy          *reporterRef         `json:"reportedBy,omitempty"`
	Assignments         []incidentAssignment `json:"assignments"`
	Timeline            []timelineEvent      `json:"timeline"`
	MergedIntoID        string               `json:"mergedIntoId,omitempty"`
	MergedBy            string               `json:"mergedBy,omitempty"`
	MergedAt            *time.Time           `json:"mergedAt,omitempty"`
	MergeReason         string               `json:"mergeReason,omitempty"`
	VerifiedBy          string               `json:"verifiedBy,omitempty"`
	VerifiedAt          *time.Time           `json:"verifiedAt,omitempty"`
	StatusUpdatedBy     string               `json:"statusUpdatedBy,omitempty"`
	StatusReason        string               `json:"statusReason,omitempty"`
	ResolutionNotes     string               `json:"resolutionNotes,omitempty"`
	ClosedAt            *time.Time           `json:"closedAt,omitempty"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

type duplicateCandidate struct {
	IncidentID     string   `json:"incidentId"`
	Reference      string   `json:"reference"`
	Score          float64  `json:"score"`
	DistanceMeters float64  `json:"distanceMeters"`
	MinutesApart   int      `json:"minutesApart"`
	Reasons        []string `json:"reasons"`
}

type createIncidentResponse struct {
	ID                  string               `json:"id"`
	Reference           string               `json:"reference"`
	Status              string               `json:"status"`
	Severity            string               `json:"severity"`
	PriorityReview      bool                 `json:"priorityReview"`
	DuplicateCandidates []duplicateCandidate `json:"duplicateCandidates"`
}

type incidentListResponse struct {
	Incidents []incidentRecord `json:"incidents"`
}

type duplicateReviewResponse struct {
	Incident   incidentRecord             `json:"incident"`
	Candidates []duplicateReviewCandidate `json:"candidates"`
}

type duplicateReviewCandidate struct {
	Candidate duplicateCandidate `json:"candidate"`
	Incident  incidentRecord     `json:"incident"`
}

type mergeIncidentsRequest struct {
	DuplicateIncidentIDs []string `json:"duplicateIncidentIds"`
	Note                 string   `json:"note"`
}

type mergeIncidentsResponse struct {
	Incident        incidentRecord   `json:"incident"`
	MergedIncidents []incidentRecord `json:"mergedIncidents"`
}

type assignmentRequest struct {
	AgencyID      string `json:"agencyId"`
	AgencyName    string `json:"agencyName"`
	AgencyType    string `json:"agencyType"`
	Priority      string `json:"priority"`
	Instructions  string `json:"instructions"`
	ResponderLead string `json:"responderLead"`
}

type incidentAssignment struct {
	ID            string    `json:"id"`
	AgencyID      string    `json:"agencyId"`
	AgencyName    string    `json:"agencyName"`
	AgencyType    string    `json:"agencyType"`
	Priority      string    `json:"priority"`
	Instructions  string    `json:"instructions"`
	ResponderLead string    `json:"responderLead,omitempty"`
	Status        string    `json:"status"`
	AssignedBy    string    `json:"assignedBy"`
	AssignedAt    time.Time `json:"assignedAt"`
}

type timelineEvent struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Message       string            `json:"message"`
	ActorUserID   string            `json:"actorUserId,omitempty"`
	ActorAgencyID string            `json:"actorAgencyId,omitempty"`
	ActorRole     string            `json:"actorRole,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
}

type incidentWorkflowRequest struct {
	Note            string `json:"note"`
	ResolutionNotes string `json:"resolutionNotes"`
}

type incidentStatusRequest struct {
	Status          string `json:"status"`
	Note            string `json:"note"`
	ResolutionNotes string `json:"resolutionNotes"`
}

type incidentAuditListResponse struct {
	Logs []auditEvent `json:"logs"`
}

type auditEvent struct {
	ID            string         `json:"id"`
	ActorUserID   string         `json:"actorUserId"`
	ActorAgencyID string         `json:"actorAgencyId"`
	ActorRole     string         `json:"actorRole"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId"`
	RequestID     string         `json:"requestId,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type authorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

type initiateMediaUploadRequest struct {
	Purpose     string `json:"purpose"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	UploadedBy  string `json:"uploadedBy,omitempty"`
}

type mediaUploadResponse struct {
	MediaID      string            `json:"mediaId"`
	UploadURL    string            `json:"uploadUrl"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	ExpiresAt    time.Time         `json:"expiresAt"`
	MaxSizeBytes int64             `json:"maxSizeBytes"`
	Access       string            `json:"access"`
}

type mediaRecord struct {
	ID          string     `json:"id"`
	Purpose     string     `json:"purpose"`
	FileName    string     `json:"fileName"`
	ContentType string     `json:"contentType"`
	SizeBytes   int64      `json:"sizeBytes"`
	UploadedBy  string     `json:"uploadedBy,omitempty"`
	IncidentID  string     `json:"incidentId,omitempty"`
	Access      string     `json:"access"`
	Status      string     `json:"status"`
	UploadURL   string     `json:"uploadUrl"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	LinkedAt    *time.Time `json:"linkedAt,omitempty"`
}

type mediaListResponse struct {
	Media []mediaRecord `json:"media"`
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
	allowedIncidentStatuses = map[string]bool{
		"reported":          true,
		"under_review":      true,
		"verified":          true,
		"assigned":          true,
		"response_en_route": true,
		"on_scene":          true,
		"contained":         true,
		"recovery_ongoing":  true,
		"closed":            true,
		"false_report":      true,
	}
	allowedIncidentTransitions = map[string]map[string]bool{
		"reported": {
			"under_review": true,
			"verified":     true,
			"false_report": true,
		},
		"under_review": {
			"verified":     true,
			"false_report": true,
		},
		"verified": {
			"assigned":          true,
			"response_en_route": true,
			"false_report":      true,
		},
		"assigned": {
			"response_en_route": true,
			"on_scene":          true,
			"contained":         true,
			"recovery_ongoing":  true,
			"closed":            true,
			"false_report":      true,
		},
		"response_en_route": {
			"on_scene":         true,
			"contained":        true,
			"recovery_ongoing": true,
			"closed":           true,
			"false_report":     true,
		},
		"on_scene": {
			"contained":        true,
			"recovery_ongoing": true,
			"closed":           true,
			"false_report":     true,
		},
		"contained": {
			"recovery_ongoing": true,
			"closed":           true,
			"false_report":     true,
		},
		"recovery_ongoing": {
			"closed":       true,
			"false_report": true,
		},
	}
	statusWorkflowRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"responder":        true,
	}
	verificationRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	incidentAuditRoles = map[string]bool{
		"system_admin":  true,
		"agency_admin":  true,
		"nadmo_officer": true,
	}
	assignmentRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	mergeRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	incidentReadRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"responder":        true,
		"agency_viewer":    true,
	}
	allowedAgencyTypes = map[string]bool{
		"nadmo":             true,
		"district_assembly": true,
		"police":            true,
		"fire":              true,
		"ambulance":         true,
		"meteorological":    true,
		"hydrological":      true,
		"hospital":          true,
		"utility":           true,
		"ngo":               true,
		"other":             true,
	}
	allowedAssignmentPriorities = map[string]bool{
		"low":    true,
		"normal": true,
		"high":   true,
		"urgent": true,
	}
	mediaRefPattern   = regexp.MustCompile(`^[A-Za-z0-9_-]{3,128}$`)
	wordPattern       = regexp.MustCompile(`[a-z0-9]+`)
	allowedMediaTypes = map[string]int64{
		"image/jpeg":      10 * 1024 * 1024,
		"image/png":       10 * 1024 * 1024,
		"image/webp":      10 * 1024 * 1024,
		"video/mp4":       100 * 1024 * 1024,
		"video/quicktime": 100 * 1024 * 1024,
		"audio/mpeg":      25 * 1024 * 1024,
		"audio/mp4":       25 * 1024 * 1024,
		"audio/wav":       25 * 1024 * 1024,
	}
)

const (
	duplicateCandidateLimit  = 5
	duplicateDistanceMeters  = 750.0
	duplicateReviewWindow    = 3 * time.Hour
	duplicateMinimumScore    = 0.45
	similarDescriptionCutoff = 0.25
	earthRadiusMeters        = 6371000.0
)

func main() {
	srv := newServerFromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/incidents", srv.createIncidentHandler)
	mux.HandleFunc("GET /api/v1/incidents", srv.listIncidentsHandler)
	mux.HandleFunc("GET /api/v1/incidents/{id}/duplicates", srv.duplicateReviewHandler)
	mux.HandleFunc("GET /api/v1/incidents/audit", srv.listIncidentAuditHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/merge", srv.mergeIncidentHandler)
	mux.HandleFunc("PATCH /api/v1/incidents/{id}/status", srv.updateIncidentStatusHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/verify", srv.verifyIncidentHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/assignments", srv.assignIncidentHandler)
	mux.HandleFunc("POST /api/v1/media/uploads", srv.initiateMediaUploadHandler)
	mux.HandleFunc("GET /api/v1/media", srv.listMediaHandler)

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
	return &memoryStore{
		incidents: map[string]incidentRecord{},
		media:     map[string]mediaRecord{},
		audit:     []auditEvent{},
	}
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

	if err := s.store.validateMediaReferences(normalized.Media); errors.Is(err, errUnknownMedia) {
		writeError(w, http.StatusBadRequest, "unknown_media", "media references must be created through the upload initiation endpoint before reporting")
		return
	} else if errors.Is(err, errMediaAlreadyLinked) {
		writeError(w, http.StatusBadRequest, "media_already_linked", "one or more media references are already linked to another incident")
		return
	} else if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_media", err.Error())
		return
	}

	record := s.store.createIncident(normalized, s.now())
	s.store.linkMediaToIncident(record.ID, record.Media, s.now())
	writeJSON(w, http.StatusCreated, createIncidentResponse{
		ID:                  record.ID,
		Reference:           record.Reference,
		Status:              record.Status,
		Severity:            record.Severity,
		PriorityReview:      record.PriorityReview,
		DuplicateCandidates: record.DuplicateCandidates,
	})
}

func (s *server) listIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	assignedToMe := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("assignedToMe"))) == "true"
	assignedAgencyID := strings.TrimSpace(r.URL.Query().Get("assignedAgencyId"))

	if assignedToMe {
		ctx, ok := requireAuthority(w, r, incidentReadRoles)
		if !ok {
			return
		}
		assignedAgencyID = ctx.ActorAgencyID
	} else if assignedAgencyID != "" {
		if _, ok := requireAuthority(w, r, incidentReadRoles); !ok {
			return
		}
	}

	writeJSON(w, http.StatusOK, incidentListResponse{Incidents: s.store.listIncidents(assignedAgencyID)})
}

func (s *server) duplicateReviewHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r, incidentReadRoles); !ok {
		return
	}

	payload, code, message := s.store.duplicateReview(r.PathValue("id"))
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *server) listIncidentAuditHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r, incidentAuditRoles); !ok {
		return
	}

	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}

	writeJSON(w, http.StatusOK, incidentAuditListResponse{Logs: s.store.listAudit(limit)})
}

func (s *server) verifyIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, verificationRoles)
	if !ok {
		return
	}

	var request incidentWorkflowRequest
	if err := optionalDecodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.Note = strings.TrimSpace(request.Note)
	if len(request.Note) > 1000 || unsafeText(request.Note) {
		writeError(w, http.StatusBadRequest, "invalid_note", "note must be 1000 safe characters or fewer")
		return
	}

	incident, code, message := s.store.transitionIncident(r.PathValue("id"), "verified", ctx, request, s.now())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, incident)
}

func (s *server) updateIncidentStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, statusWorkflowRoles)
	if !ok {
		return
	}

	var request incidentStatusRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeIncidentStatusRequest(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	incident, code, message := s.store.transitionIncident(
		r.PathValue("id"),
		normalized.Status,
		ctx,
		incidentWorkflowRequest{Note: normalized.Note, ResolutionNotes: normalized.ResolutionNotes},
		s.now(),
	)
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, incident)
}

func (s *server) mergeIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, mergeRoles)
	if !ok {
		return
	}

	var request mergeIncidentsRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeMergeRequest(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	payload, code, message := s.store.mergeIncidents(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *server) assignIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}

	var request assignmentRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAssignmentRequest(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	incident, code, message := s.store.assignIncident(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusCreated, incident)
}

func (s *server) initiateMediaUploadHandler(w http.ResponseWriter, r *http.Request) {
	var request initiateMediaUploadRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeMediaUploadRequest(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	record := s.store.createMediaUpload(normalized, s.now())
	writeJSON(w, http.StatusCreated, mediaUploadResponse{
		MediaID:   record.ID,
		UploadURL: record.UploadURL,
		Method:    "PUT",
		Headers: map[string]string{
			"Content-Type": record.ContentType,
		},
		ExpiresAt:    record.ExpiresAt,
		MaxSizeBytes: allowedMediaTypes[record.ContentType],
		Access:       record.Access,
	})
}

func (s *server) listMediaHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, mediaListResponse{Media: s.store.listMedia()})
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
		MergedIncidentIDs:  []string{},
		Assignments:        []incidentAssignment{},
		Timeline: []timelineEvent{
			newTimelineEvent("incident.reported", "Citizen report received", authorityContext{}, map[string]string{
				"reference": reference,
				"hazard":    request.Type,
				"urgency":   request.Urgency,
			}, timestamp),
		},
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
	record.DuplicateCandidates = m.duplicateCandidatesLocked(record)
	m.incidents[record.ID] = record
	m.linkReverseDuplicateCandidatesLocked(record)
	return record
}

func (m *memoryStore) listIncidents(assignedAgencyID string) []incidentRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incidents := make([]incidentRecord, 0, len(m.incidents))
	for _, incident := range m.incidents {
		if assignedAgencyID != "" && !incidentAssignedToAgency(incident, assignedAgencyID) {
			continue
		}
		incidents = append(incidents, incident)
	}
	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt.After(incidents[j].CreatedAt)
	})
	return incidents
}

func (m *memoryStore) duplicateReview(id string) (duplicateReviewResponse, string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incident, ok := m.incidents[id]
	if !ok {
		return duplicateReviewResponse{}, "not_found", "incident was not found"
	}

	candidates := make([]duplicateReviewCandidate, 0, len(incident.DuplicateCandidates))
	for _, candidate := range incident.DuplicateCandidates {
		candidateIncident, exists := m.incidents[candidate.IncidentID]
		if !exists || candidateIncident.MergedIntoID != "" || candidateIncident.Status == "false_report" {
			continue
		}
		candidates = append(candidates, duplicateReviewCandidate{
			Candidate: candidate,
			Incident:  candidateIncident,
		})
	}

	return duplicateReviewResponse{Incident: incident, Candidates: candidates}, "", ""
}

func (m *memoryStore) transitionIncident(id string, nextStatus string, ctx authorityContext, request incidentWorkflowRequest, now time.Time) (incidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return incidentRecord{}, "not_found", "incident was not found"
	}

	nextStatus = incidentStatusSlug(nextStatus)
	if !allowedIncidentStatuses[nextStatus] {
		return incidentRecord{}, "invalid_status", "status must be a supported incident status"
	}
	if incident.Status == nextStatus {
		return incidentRecord{}, "invalid_transition", "incident is already in that status"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return incidentRecord{}, "invalid_transition", "closed and false-report incidents are terminal"
	}
	if !allowedIncidentTransitions[incident.Status][nextStatus] {
		return incidentRecord{}, "invalid_transition", fmt.Sprintf("cannot move incident from %s to %s", incident.Status, nextStatus)
	}

	note := strings.TrimSpace(request.Note)
	resolutionNotes := strings.TrimSpace(request.ResolutionNotes)
	if requiresResolutionNotes(nextStatus) && resolutionNotes == "" {
		return incidentRecord{}, "missing_resolution_notes", "resolutionNotes are required for closed and false report statuses"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	incident.Status = nextStatus
	incident.StatusUpdatedBy = ctx.ActorUserID
	incident.StatusReason = note
	incident.UpdatedAt = timestamp

	action := "incident.status_changed"
	if nextStatus == "verified" {
		action = "incident.verified"
		incident.VerifiedBy = ctx.ActorUserID
		if incident.VerifiedAt == nil {
			incident.VerifiedAt = &timestamp
		}
	}
	if requiresResolutionNotes(nextStatus) {
		incident.ResolutionNotes = resolutionNotes
		incident.ClosedAt = &timestamp
		if nextStatus == "closed" {
			action = "incident.closed"
		} else {
			action = "incident.false_reported"
		}
	}

	incident.Timeline = append(incident.Timeline, newTimelineEvent(action, timelineMessageForStatus(nextStatus, note, resolutionNotes), ctx, map[string]string{
		"fromStatus": before["status"].(string),
		"toStatus":   nextStatus,
	}, timestamp))
	m.incidents[incident.ID] = incident
	m.appendAuditLocked(action, ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	return incident, "", ""
}

func (m *memoryStore) mergeIncidents(primaryID string, request mergeIncidentsRequest, ctx authorityContext, now time.Time) (mergeIncidentsResponse, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	primary, ok := m.incidents[primaryID]
	if !ok {
		return mergeIncidentsResponse{}, "not_found", "incident was not found"
	}
	if primary.MergedIntoID != "" {
		return mergeIncidentsResponse{}, "invalid_merge", "primary incident is already merged into another incident"
	}
	if primary.Status == "closed" || primary.Status == "false_report" {
		return mergeIncidentsResponse{}, "invalid_merge", "closed and false-report incidents cannot receive duplicate merges"
	}

	duplicates := make([]incidentRecord, 0, len(request.DuplicateIncidentIDs))
	for _, duplicateID := range request.DuplicateIncidentIDs {
		duplicate, exists := m.incidents[duplicateID]
		if !exists {
			return mergeIncidentsResponse{}, "not_found", fmt.Sprintf("duplicate incident %s was not found", duplicateID)
		}
		if duplicate.ID == primary.ID {
			return mergeIncidentsResponse{}, "invalid_merge", "primary incident cannot be merged into itself"
		}
		if duplicate.MergedIntoID != "" {
			return mergeIncidentsResponse{}, "invalid_merge", fmt.Sprintf("duplicate incident %s is already merged", duplicate.Reference)
		}
		if duplicate.Status == "closed" || duplicate.Status == "false_report" {
			return mergeIncidentsResponse{}, "invalid_merge", fmt.Sprintf("duplicate incident %s is terminal", duplicate.Reference)
		}
		if _, ok := duplicateCandidateBetween(primary, duplicate); !ok {
			return mergeIncidentsResponse{}, "invalid_duplicate", fmt.Sprintf("incident %s is not a duplicate candidate for %s", duplicate.Reference, primary.Reference)
		}
		duplicates = append(duplicates, duplicate)
	}

	beforePrimary := snapshotIncident(primary)
	timestamp := now.UTC()
	mergedIDs := make([]string, 0, len(duplicates))
	mergedIncidents := make([]incidentRecord, 0, len(duplicates))
	removeIDs := map[string]bool{}
	for _, duplicate := range duplicates {
		removeIDs[duplicate.ID] = true
		mergedIDs = append(mergedIDs, duplicate.ID)
	}

	for _, duplicate := range duplicates {
		beforeDuplicate := snapshotIncident(duplicate)
		duplicate.MergedIntoID = primary.ID
		duplicate.MergedBy = ctx.ActorUserID
		duplicate.MergedAt = &timestamp
		duplicate.MergeReason = request.Note
		duplicate.Status = "closed"
		duplicate.StatusUpdatedBy = ctx.ActorUserID
		duplicate.StatusReason = fmt.Sprintf("Merged into %s", primary.Reference)
		duplicate.ResolutionNotes = request.Note
		duplicate.ClosedAt = &timestamp
		duplicate.UpdatedAt = timestamp
		duplicate.DuplicateCandidates = filterDuplicateCandidates(duplicate.DuplicateCandidates, map[string]bool{primary.ID: true})
		duplicate.Timeline = append(duplicate.Timeline, newTimelineEvent("incident.merged_into", fmt.Sprintf("Merged into %s", primary.Reference), ctx, map[string]string{
			"primaryIncidentId": primary.ID,
			"primaryReference":  primary.Reference,
			"note":              request.Note,
		}, timestamp))

		m.incidents[duplicate.ID] = duplicate
		m.appendAuditLocked("incident.merged_into", ctx, duplicate.ID, beforeDuplicate, snapshotIncident(duplicate), timestamp)
		mergedIncidents = append(mergedIncidents, duplicate)
	}

	primary.MergedIncidentIDs = appendUniqueStrings(primary.MergedIncidentIDs, mergedIDs...)
	sort.Strings(primary.MergedIncidentIDs)
	primary.DuplicateCandidates = filterDuplicateCandidates(primary.DuplicateCandidates, removeIDs)
	primary.StatusUpdatedBy = ctx.ActorUserID
	primary.StatusReason = fmt.Sprintf("Merged %d duplicate report(s)", len(mergedIDs))
	primary.UpdatedAt = timestamp
	primary.Timeline = append(primary.Timeline, newTimelineEvent("incident.merged", fmt.Sprintf("Merged %d duplicate report(s)", len(mergedIDs)), ctx, map[string]string{
		"duplicateIncidentIds": strings.Join(mergedIDs, ","),
		"note":                 request.Note,
	}, timestamp))

	m.incidents[primary.ID] = primary
	m.appendAuditLocked("incident.merged", ctx, primary.ID, beforePrimary, snapshotIncident(primary), timestamp)
	return mergeIncidentsResponse{Incident: primary, MergedIncidents: mergedIncidents}, "", ""
}

func (m *memoryStore) assignIncident(id string, request assignmentRequest, ctx authorityContext, now time.Time) (incidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return incidentRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "reported" || incident.Status == "under_review" {
		return incidentRecord{}, "invalid_transition", "incident must be verified before assignment"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return incidentRecord{}, "invalid_transition", "closed and false-report incidents cannot be assigned"
	}
	if ctx.ActorRole == "agency_admin" && ctx.ActorAgencyID != request.AgencyID {
		return incidentRecord{}, "forbidden", "agency admins can assign only to their own agency"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	assignment := incidentAssignment{
		ID:            fmt.Sprintf("asg_%06d", len(incident.Assignments)+1),
		AgencyID:      request.AgencyID,
		AgencyName:    request.AgencyName,
		AgencyType:    request.AgencyType,
		Priority:      request.Priority,
		Instructions:  request.Instructions,
		ResponderLead: request.ResponderLead,
		Status:        "active",
		AssignedBy:    ctx.ActorUserID,
		AssignedAt:    timestamp,
	}

	incident.Assignments = append(incident.Assignments, assignment)
	if incident.Status == "verified" {
		incident.Status = "assigned"
	}
	incident.StatusUpdatedBy = ctx.ActorUserID
	incident.StatusReason = fmt.Sprintf("Assigned to %s", assignment.AgencyName)
	incident.UpdatedAt = timestamp
	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.assigned", fmt.Sprintf("Assigned to %s", assignment.AgencyName), ctx, map[string]string{
		"assignmentId": assignment.ID,
		"agencyId":     assignment.AgencyID,
		"agencyName":   assignment.AgencyName,
		"agencyType":   assignment.AgencyType,
		"priority":     assignment.Priority,
	}, timestamp))

	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.assigned", ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	return incident, "", ""
}

func (m *memoryStore) listAudit(limit int) []auditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := append([]auditEvent(nil), m.audit...)
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return logs[i].ID > logs[j].ID
		}
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	if len(logs) > limit {
		return logs[:limit]
	}
	return logs
}

func (m *memoryStore) appendAuditLocked(action string, ctx authorityContext, targetID string, before map[string]any, after map[string]any, now time.Time) {
	m.audit = append(m.audit, auditEvent{
		ID:            fmt.Sprintf("aud_%06d", len(m.audit)+1),
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Action:        action,
		TargetType:    "incident",
		TargetID:      targetID,
		RequestID:     ctx.RequestID,
		Before:        before,
		After:         after,
		CreatedAt:     now,
	})
}

func (m *memoryStore) createMediaUpload(request initiateMediaUploadRequest, now time.Time) mediaRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	timestamp := now.UTC()
	id := newID("media")
	record := mediaRecord{
		ID:          id,
		Purpose:     request.Purpose,
		FileName:    request.FileName,
		ContentType: request.ContentType,
		SizeBytes:   request.SizeBytes,
		UploadedBy:  request.UploadedBy,
		Access:      "private",
		Status:      "pending_upload",
		UploadURL:   fmt.Sprintf("/dev/uploads/%s/%s", id, request.FileName),
		ExpiresAt:   timestamp.Add(15 * time.Minute),
		CreatedAt:   timestamp,
	}
	m.media[record.ID] = record
	return record
}

var (
	errUnknownMedia       = errors.New("unknown media")
	errMediaAlreadyLinked = errors.New("media already linked")
	errDuplicateMediaRef  = errors.New("duplicate media reference")
)

func (m *memoryStore) validateMediaReferences(mediaIDs []string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := map[string]bool{}
	for _, mediaID := range mediaIDs {
		if seen[mediaID] {
			return errDuplicateMediaRef
		}
		seen[mediaID] = true

		record, ok := m.media[mediaID]
		if !ok {
			return errUnknownMedia
		}
		if record.IncidentID != "" || record.Status == "linked" {
			return errMediaAlreadyLinked
		}
	}
	return nil
}

func (m *memoryStore) linkMediaToIncident(incidentID string, mediaIDs []string, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	linkedAt := now.UTC()
	for _, mediaID := range mediaIDs {
		record := m.media[mediaID]
		record.IncidentID = incidentID
		record.Status = "linked"
		record.LinkedAt = &linkedAt
		m.media[mediaID] = record
	}
}

func (m *memoryStore) listMedia() []mediaRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	media := make([]mediaRecord, 0, len(m.media))
	for _, record := range m.media {
		media = append(media, record)
	}
	sort.Slice(media, func(i, j int) bool {
		return media[i].CreatedAt.After(media[j].CreatedAt)
	})
	return media
}

func (m *memoryStore) duplicateCandidatesLocked(record incidentRecord) []duplicateCandidate {
	candidates := make([]duplicateCandidate, 0, duplicateCandidateLimit)
	for _, existing := range m.incidents {
		candidate, ok := scoreDuplicateCandidate(record, existing)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].DistanceMeters < candidates[j].DistanceMeters
		}
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > duplicateCandidateLimit {
		return candidates[:duplicateCandidateLimit]
	}
	return candidates
}

func (m *memoryStore) linkReverseDuplicateCandidatesLocked(record incidentRecord) {
	for _, candidate := range record.DuplicateCandidates {
		existing := m.incidents[candidate.IncidentID]
		if hasDuplicateCandidate(existing.DuplicateCandidates, record.ID) {
			continue
		}

		existing.DuplicateCandidates = append(existing.DuplicateCandidates, reverseDuplicateCandidate(record, candidate))
		sort.Slice(existing.DuplicateCandidates, func(i, j int) bool {
			if existing.DuplicateCandidates[i].Score == existing.DuplicateCandidates[j].Score {
				return existing.DuplicateCandidates[i].DistanceMeters < existing.DuplicateCandidates[j].DistanceMeters
			}
			return existing.DuplicateCandidates[i].Score > existing.DuplicateCandidates[j].Score
		})
		if len(existing.DuplicateCandidates) > duplicateCandidateLimit {
			existing.DuplicateCandidates = existing.DuplicateCandidates[:duplicateCandidateLimit]
		}
		existing.UpdatedAt = record.UpdatedAt
		m.incidents[existing.ID] = existing
	}
}

func scoreDuplicateCandidate(record incidentRecord, existing incidentRecord) (duplicateCandidate, bool) {
	if record.ID == existing.ID || record.Type != existing.Type || existing.Status == "false_report" {
		return duplicateCandidate{}, false
	}

	timeApart := absoluteDuration(record.CreatedAt.Sub(existing.CreatedAt))
	if timeApart > duplicateReviewWindow {
		return duplicateCandidate{}, false
	}

	distance := haversineMeters(record.Location, existing.Location)
	if distance > duplicateDistanceMeters {
		return duplicateCandidate{}, false
	}

	descriptionScore := descriptionSimilarity(record.Description, existing.Description)
	distanceScore := clamp01(1 - distance/duplicateDistanceMeters)
	timeScore := clamp01(1 - timeApart.Seconds()/duplicateReviewWindow.Seconds())
	score := roundScore(0.50*distanceScore + 0.30*timeScore + 0.20*descriptionScore)
	if score < duplicateMinimumScore {
		return duplicateCandidate{}, false
	}

	reasons := []string{"same_hazard", "nearby_location", "recent_report"}
	if descriptionScore >= similarDescriptionCutoff {
		reasons = append(reasons, "similar_description")
	}

	return duplicateCandidate{
		IncidentID:     existing.ID,
		Reference:      existing.Reference,
		Score:          score,
		DistanceMeters: math.Round(distance),
		MinutesApart:   int(math.Round(timeApart.Minutes())),
		Reasons:        reasons,
	}, true
}

func reverseDuplicateCandidate(record incidentRecord, candidate duplicateCandidate) duplicateCandidate {
	return duplicateCandidate{
		IncidentID:     record.ID,
		Reference:      record.Reference,
		Score:          candidate.Score,
		DistanceMeters: candidate.DistanceMeters,
		MinutesApart:   candidate.MinutesApart,
		Reasons:        append([]string{}, candidate.Reasons...),
	}
}

func hasDuplicateCandidate(candidates []duplicateCandidate, incidentID string) bool {
	for _, candidate := range candidates {
		if candidate.IncidentID == incidentID {
			return true
		}
	}
	return false
}

func haversineMeters(a coordinates, b coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)

	sinLat := math.Sin(deltaLat / 2)
	sinLng := math.Sin(deltaLng / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func descriptionSimilarity(a string, b string) float64 {
	aTokens := tokenSet(a)
	bTokens := tokenSet(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}

	intersection := 0
	union := map[string]bool{}
	for token := range aTokens {
		union[token] = true
		if bTokens[token] {
			intersection++
		}
	}
	for token := range bTokens {
		union[token] = true
	}

	return float64(intersection) / float64(len(union))
}

func tokenSet(value string) map[string]bool {
	tokens := wordPattern.FindAllString(strings.ToLower(value), -1)
	set := map[string]bool{}
	for _, token := range tokens {
		if len(token) < 3 {
			continue
		}
		set[token] = true
	}
	return set
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func roundScore(value float64) float64 {
	return math.Round(value*100) / 100
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

func normalizeMediaUploadRequest(request initiateMediaUploadRequest) (initiateMediaUploadRequest, string, string) {
	request.Purpose = strings.TrimSpace(strings.ToLower(request.Purpose))
	request.FileName = strings.TrimSpace(request.FileName)
	request.ContentType = strings.TrimSpace(strings.ToLower(request.ContentType))
	request.UploadedBy = strings.TrimSpace(request.UploadedBy)

	if request.Purpose == "" {
		request.Purpose = "incident_media"
	}
	if request.Purpose != "incident_media" {
		return request, "invalid_purpose", "purpose must be incident_media"
	}

	if !validFileName(request.FileName) {
		return request, "invalid_file_name", "fileName must be 1 to 180 safe characters without path separators"
	}

	maxSize, ok := allowedMediaTypes[request.ContentType]
	if !ok {
		return request, "unsupported_media_type", "contentType must be a supported image, video, or audio type"
	}

	if request.SizeBytes <= 0 || request.SizeBytes > maxSize {
		return request, "invalid_file_size", fmt.Sprintf("sizeBytes must be between 1 and %d for %s", maxSize, request.ContentType)
	}

	if request.UploadedBy != "" && !mediaRefPattern.MatchString(request.UploadedBy) {
		return request, "invalid_uploaded_by", "uploadedBy must be a safe user reference when supplied"
	}

	return request, "", ""
}

func normalizeIncidentStatusRequest(request incidentStatusRequest) (incidentStatusRequest, string, string) {
	request.Status = incidentStatusSlug(request.Status)
	request.Note = strings.TrimSpace(request.Note)
	request.ResolutionNotes = strings.TrimSpace(request.ResolutionNotes)

	if !allowedIncidentStatuses[request.Status] {
		return request, "invalid_status", "status must be reported, under_review, verified, assigned, response_en_route, on_scene, contained, recovery_ongoing, closed, or false_report"
	}
	if len(request.Note) > 1000 || unsafeText(request.Note) {
		return request, "invalid_note", "note must be 1000 safe characters or fewer"
	}
	if len(request.ResolutionNotes) > 2000 || unsafeText(request.ResolutionNotes) {
		return request, "invalid_resolution_notes", "resolutionNotes must be 2000 safe characters or fewer"
	}
	if requiresResolutionNotes(request.Status) && request.ResolutionNotes == "" {
		return request, "missing_resolution_notes", "resolutionNotes are required for closed and false report statuses"
	}
	return request, "", ""
}

func normalizeMergeRequest(request mergeIncidentsRequest) (mergeIncidentsRequest, string, string) {
	request.Note = strings.TrimSpace(request.Note)

	if len(request.DuplicateIncidentIDs) == 0 {
		return request, "missing_duplicates", "duplicateIncidentIds must include at least one incident"
	}
	if len(request.DuplicateIncidentIDs) > duplicateCandidateLimit {
		return request, "too_many_duplicates", fmt.Sprintf("duplicateIncidentIds can include at most %d incidents", duplicateCandidateLimit)
	}

	normalizedIDs := make([]string, 0, len(request.DuplicateIncidentIDs))
	seen := map[string]bool{}
	for _, incidentID := range request.DuplicateIncidentIDs {
		incidentID = strings.TrimSpace(incidentID)
		if incidentID == "" || !mediaRefPattern.MatchString(incidentID) {
			return request, "invalid_duplicate_id", "duplicateIncidentIds must contain safe incident references"
		}
		if seen[incidentID] {
			return request, "duplicate_duplicate_id", "duplicateIncidentIds must not contain the same incident more than once"
		}
		seen[incidentID] = true
		normalizedIDs = append(normalizedIDs, incidentID)
	}

	if len(request.Note) < 5 || len(request.Note) > 1000 || unsafeText(request.Note) {
		return request, "invalid_note", "note must be 5 to 1000 safe characters"
	}

	request.DuplicateIncidentIDs = normalizedIDs
	return request, "", ""
}

func normalizeAssignmentRequest(request assignmentRequest) (assignmentRequest, string, string) {
	request.AgencyID = strings.TrimSpace(request.AgencyID)
	request.AgencyName = strings.TrimSpace(request.AgencyName)
	request.AgencyType = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.AgencyType)), "-", "_"), " ", "_")
	request.Priority = strings.TrimSpace(strings.ToLower(request.Priority))
	request.Instructions = strings.TrimSpace(request.Instructions)
	request.ResponderLead = strings.TrimSpace(request.ResponderLead)

	if request.AgencyID == "" || !mediaRefPattern.MatchString(request.AgencyID) {
		return request, "invalid_agency_id", "agencyId is required and must be a safe agency reference"
	}
	if len(request.AgencyName) < 2 || len(request.AgencyName) > 140 || unsafeText(request.AgencyName) {
		return request, "invalid_agency_name", "agencyName must be 2 to 140 safe characters"
	}
	if !allowedAgencyTypes[request.AgencyType] {
		return request, "invalid_agency_type", "agencyType must be police, fire, ambulance, nadmo, district_assembly, or another supported agency type"
	}
	if request.Priority == "" {
		request.Priority = "normal"
	}
	if !allowedAssignmentPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, normal, high, or urgent"
	}
	if len(request.Instructions) < 5 || len(request.Instructions) > 1000 || unsafeText(request.Instructions) {
		return request, "invalid_instructions", "instructions must be 5 to 1000 safe characters"
	}
	if len(request.ResponderLead) > 140 || unsafeText(request.ResponderLead) {
		return request, "invalid_responder_lead", "responderLead must be 140 safe characters or fewer"
	}
	return request, "", ""
}

func incidentStatusSlug(status string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(status)), "-", "_"), " ", "_")
}

func requiresResolutionNotes(status string) bool {
	return status == "closed" || status == "false_report"
}

func incidentAssignedToAgency(incident incidentRecord, agencyID string) bool {
	for _, assignment := range incident.Assignments {
		if assignment.AgencyID == agencyID && assignment.Status == "active" {
			return true
		}
	}
	return false
}

func assignmentAgencyIDs(assignments []incidentAssignment) []string {
	ids := make([]string, 0, len(assignments))
	seen := map[string]bool{}
	for _, assignment := range assignments {
		if assignment.AgencyID == "" || seen[assignment.AgencyID] {
			continue
		}
		seen[assignment.AgencyID] = true
		ids = append(ids, assignment.AgencyID)
	}
	sort.Strings(ids)
	return ids
}

func duplicateCandidateBetween(primary incidentRecord, duplicate incidentRecord) (duplicateCandidate, bool) {
	for _, candidate := range primary.DuplicateCandidates {
		if candidate.IncidentID == duplicate.ID {
			return candidate, true
		}
	}
	for _, candidate := range duplicate.DuplicateCandidates {
		if candidate.IncidentID == primary.ID {
			return reverseDuplicateCandidate(duplicate, candidate), true
		}
	}
	return duplicateCandidate{}, false
}

func filterDuplicateCandidates(candidates []duplicateCandidate, removeIDs map[string]bool) []duplicateCandidate {
	if len(candidates) == 0 || len(removeIDs) == 0 {
		return candidates
	}
	filtered := make([]duplicateCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if removeIDs[candidate.IncidentID] {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered
}

func appendUniqueStrings(values []string, additions ...string) []string {
	seen := map[string]bool{}
	next := make([]string, 0, len(values)+len(additions))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		next = append(next, value)
	}
	for _, value := range additions {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		next = append(next, value)
	}
	return next
}

func newTimelineEvent(eventType string, message string, ctx authorityContext, metadata map[string]string, now time.Time) timelineEvent {
	return timelineEvent{
		ID:            newID("tle"),
		Type:          eventType,
		Message:       message,
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Metadata:      metadata,
		CreatedAt:     now,
	}
}

func timelineMessageForStatus(status string, note string, resolutionNotes string) string {
	switch status {
	case "verified":
		return "Incident verified"
	case "closed":
		if resolutionNotes != "" {
			return "Incident closed with resolution notes"
		}
		return "Incident closed"
	case "false_report":
		return "Incident marked as false report"
	default:
		if note != "" {
			return fmt.Sprintf("Status changed to %s: %s", status, note)
		}
		return fmt.Sprintf("Status changed to %s", status)
	}
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

func snapshotIncident(incident incidentRecord) map[string]any {
	return map[string]any{
		"id":                incident.ID,
		"reference":         incident.Reference,
		"type":              incident.Type,
		"severity":          incident.Severity,
		"status":            incident.Status,
		"priorityReview":    incident.PriorityReview,
		"verifiedBy":        incident.VerifiedBy,
		"statusUpdatedBy":   incident.StatusUpdatedBy,
		"statusReason":      incident.StatusReason,
		"resolutionNotes":   incident.ResolutionNotes,
		"mergedIncidentIds": append([]string{}, incident.MergedIncidentIDs...),
		"mergedIntoId":      incident.MergedIntoID,
		"mergeReason":       incident.MergeReason,
		"duplicateCount":    len(incident.DuplicateCandidates),
		"assignmentCount":   len(incident.Assignments),
		"assignedAgencyIds": assignmentAgencyIDs(incident.Assignments),
	}
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
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for incident workflow actions")
		return authorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this incident workflow action")
		return authorityContext{}, false
	}

	return ctx, true
}

func statusForCode(code string) int {
	switch code {
	case "not_found":
		return http.StatusNotFound
	case "forbidden":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func optionalDecodeJSON(r *http.Request, target any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	return decodeJSON(r, target)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NADAA-Actor-ID, X-NADAA-Actor-Role, X-NADAA-Agency-ID, X-NADAA-MFA-Completed, X-NADAA-Request-ID")
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

func validFileName(fileName string) bool {
	if fileName == "" || len(fileName) > 180 || unsafeText(fileName) {
		return false
	}
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || strings.Contains(fileName, "..") {
		return false
	}
	return true
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
