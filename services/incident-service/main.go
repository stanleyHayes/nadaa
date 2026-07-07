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
	mu                    sync.RWMutex
	sequence              int
	volunteerSequence     int
	volunteerTaskSequence int
	incidents             map[string]incidentRecord
	media                 map[string]mediaRecord
	volunteers            map[string]volunteerProfile
	volunteerTasks        map[string]volunteerTaskRecord
	audit                 []auditEvent
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
	Privacy             incidentPrivacy      `json:"privacy"`
	AccessibilityNeeds  string               `json:"accessibilityNeeds,omitempty"`
	Media               []string             `json:"media"`
	PriorityReview      bool                 `json:"priorityReview"`
	AbuseSignals        []abuseSignal        `json:"abuseSignals"`
	AbuseScore          float64              `json:"abuseScore"`
	AbuseReviewRequired bool                 `json:"abuseReviewRequired"`
	AbuseReviewReason   string               `json:"abuseReviewReason,omitempty"`
	AbuseReviewDecision string               `json:"abuseReviewDecision,omitempty"`
	AbuseReviewedBy     string               `json:"abuseReviewedBy,omitempty"`
	AbuseReviewedAt     *time.Time           `json:"abuseReviewedAt,omitempty"`
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

type incidentPrivacy struct {
	ReporterIdentityVisible bool     `json:"reporterIdentityVisible"`
	ReporterContactVisible  bool     `json:"reporterContactVisible"`
	LocationPrecision       string   `json:"locationPrecision"`
	LocationUse             string   `json:"locationUse"`
	Disclosure              string   `json:"disclosure"`
	Notes                   []string `json:"notes"`
}

type duplicateCandidate struct {
	IncidentID     string   `json:"incidentId"`
	Reference      string   `json:"reference"`
	Score          float64  `json:"score"`
	DistanceMeters float64  `json:"distanceMeters"`
	MinutesApart   int      `json:"minutesApart"`
	Reasons        []string `json:"reasons"`
}

type abuseSignal struct {
	Code   string  `json:"code"`
	Label  string  `json:"label"`
	Detail string  `json:"detail"`
	Weight float64 `json:"weight"`
}

type createIncidentResponse struct {
	ID                  string               `json:"id"`
	Reference           string               `json:"reference"`
	Status              string               `json:"status"`
	Severity            string               `json:"severity"`
	PriorityReview      bool                 `json:"priorityReview"`
	AbuseSignals        []abuseSignal        `json:"abuseSignals"`
	AbuseScore          float64              `json:"abuseScore"`
	AbuseReviewRequired bool                 `json:"abuseReviewRequired"`
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

type volunteerProfile struct {
	ID                 string     `json:"id"`
	CitizenUserID      string     `json:"citizenUserId"`
	Name               string     `json:"name"`
	Phone              string     `json:"phone,omitempty"`
	Region             string     `json:"region"`
	District           string     `json:"district"`
	Community          string     `json:"community"`
	GroupID            string     `json:"groupId"`
	Skills             []string   `json:"skills"`
	Languages          []string   `json:"languages"`
	AvailabilityStatus string     `json:"availabilityStatus"`
	VerificationStatus string     `json:"verificationStatus"`
	SafetyNotes        []string   `json:"safetyNotes"`
	VerifiedBy         string     `json:"verifiedBy,omitempty"`
	VerifiedAt         *time.Time `json:"verifiedAt,omitempty"`
	RejectionReason    string     `json:"rejectionReason,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type registerVolunteerRequest struct {
	CitizenUserID      string   `json:"citizenUserId"`
	Name               string   `json:"name"`
	Phone              string   `json:"phone"`
	Region             string   `json:"region"`
	District           string   `json:"district"`
	Community          string   `json:"community"`
	Skills             []string `json:"skills"`
	Languages          []string `json:"languages"`
	AvailabilityStatus string   `json:"availabilityStatus"`
}

type volunteerProfileResponse struct {
	Volunteer volunteerProfile `json:"volunteer"`
}

type volunteerListResponse struct {
	Volunteers []volunteerProfile `json:"volunteers"`
}

type verifyVolunteerRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

type volunteerTaskRequest struct {
	VolunteerID   string `json:"volunteerId"`
	Type          string `json:"type"`
	Priority      string `json:"priority"`
	Instructions  string `json:"instructions"`
	LocationLabel string `json:"locationLabel"`
}

type volunteerTaskRecord struct {
	ID                 string                `json:"id"`
	IncidentID         string                `json:"incidentId"`
	IncidentReference  string                `json:"incidentReference"`
	VolunteerID        string                `json:"volunteerId"`
	VolunteerName      string                `json:"volunteerName"`
	GroupID            string                `json:"groupId"`
	Type               string                `json:"type"`
	Priority           string                `json:"priority"`
	Instructions       string                `json:"instructions"`
	LocationLabel      string                `json:"locationLabel"`
	Status             string                `json:"status"`
	SafetyRules        []string              `json:"safetyRules"`
	EscalationRequired bool                  `json:"escalationRequired"`
	AssignedBy         string                `json:"assignedBy"`
	AssignedAt         time.Time             `json:"assignedAt"`
	UpdatedAt          time.Time             `json:"updatedAt"`
	AcceptedAt         *time.Time            `json:"acceptedAt,omitempty"`
	CompletedAt        *time.Time            `json:"completedAt,omitempty"`
	Updates            []volunteerTaskUpdate `json:"updates"`
}

type volunteerTaskUpdate struct {
	ID                  string       `json:"id"`
	Type                string       `json:"type"`
	Status              string       `json:"status,omitempty"`
	Note                string       `json:"note"`
	SafetyStatus        string       `json:"safetyStatus"`
	Location            *coordinates `json:"location,omitempty"`
	EscalationRequested bool         `json:"escalationRequested"`
	CreatedBy           string       `json:"createdBy"`
	CreatedAt           time.Time    `json:"createdAt"`
}

type volunteerTaskListResponse struct {
	Tasks []volunteerTaskRecord `json:"tasks"`
}

type volunteerTaskStatusRequest struct {
	VolunteerID  string       `json:"volunteerId"`
	Status       string       `json:"status"`
	Note         string       `json:"note"`
	SafetyStatus string       `json:"safetyStatus"`
	Location     *coordinates `json:"location,omitempty"`
}

type volunteerObservationRequest struct {
	VolunteerID         string       `json:"volunteerId"`
	Observation         string       `json:"observation"`
	SafetyStatus        string       `json:"safetyStatus"`
	Location            *coordinates `json:"location,omitempty"`
	EscalationRequested bool         `json:"escalationRequested"`
	Media               []string     `json:"media"`
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

type abuseReviewRequest struct {
	Decision        string `json:"decision"`
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
	abuseReviewRoles = map[string]bool{
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
	reporterContactRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
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
	allowedVolunteerAvailability = map[string]bool{
		"available": true,
		"busy":      true,
		"off_duty":  true,
	}
	allowedVolunteerVerificationDecisions = map[string]bool{
		"verify":  true,
		"reject":  true,
		"suspend": true,
	}
	allowedVolunteerTaskTypes = map[string]bool{
		"welfare_check":       true,
		"shelter_support":     true,
		"supply_distribution": true,
		"damage_observation":  true,
		"route_observation":   true,
		"community_alerting":  true,
	}
	allowedVolunteerTaskStatuses = map[string]bool{
		"accepted":         true,
		"en_route":         true,
		"on_scene":         true,
		"completed":        true,
		"cancelled":        true,
		"needs_escalation": true,
	}
	allowedVolunteerSafetyStatuses = map[string]bool{
		"safe":            true,
		"caution":         true,
		"unsafe":          true,
		"needs_authority": true,
	}
	allowedAbuseReviewDecisions = map[string]bool{
		"clear":        true,
		"monitor":      true,
		"false_report": true,
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
	abuseReviewThreshold     = 0.55
	reporterBurstWindow      = 30 * time.Minute
	reporterBurstPreviousMin = 2
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
	mux.HandleFunc("POST /api/v1/incidents/{id}/abuse-review", srv.reviewAbuseHandler)
	mux.HandleFunc("PATCH /api/v1/incidents/{id}/status", srv.updateIncidentStatusHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/verify", srv.verifyIncidentHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/assignments", srv.assignIncidentHandler)
	mux.HandleFunc("POST /api/v1/incidents/{id}/volunteer-tasks", srv.assignVolunteerTaskHandler)
	mux.HandleFunc("POST /api/v1/media/uploads", srv.initiateMediaUploadHandler)
	mux.HandleFunc("GET /api/v1/media", srv.listMediaHandler)
	mux.HandleFunc("POST /api/v1/volunteers", srv.registerVolunteerHandler)
	mux.HandleFunc("GET /api/v1/volunteers", srv.listVolunteersHandler)
	mux.HandleFunc("POST /api/v1/volunteers/{id}/verify", srv.verifyVolunteerHandler)
	mux.HandleFunc("GET /api/v1/volunteers/{id}/tasks", srv.listVolunteerTasksHandler)
	mux.HandleFunc("PATCH /api/v1/volunteer-tasks/{id}/status", srv.updateVolunteerTaskStatusHandler)
	mux.HandleFunc("POST /api/v1/volunteer-tasks/{id}/observations", srv.submitVolunteerObservationHandler)

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
		incidents:      map[string]incidentRecord{},
		media:          map[string]mediaRecord{},
		volunteers:     map[string]volunteerProfile{},
		volunteerTasks: map[string]volunteerTaskRecord{},
		audit:          []auditEvent{},
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
		AbuseSignals:        record.AbuseSignals,
		AbuseScore:          record.AbuseScore,
		AbuseReviewRequired: record.AbuseReviewRequired,
		DuplicateCandidates: record.DuplicateCandidates,
	})
}

func (s *server) listIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, incidentReadRoles)
	if !ok {
		return
	}

	assignedToMe := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("assignedToMe"))) == "true"
	assignedAgencyID := strings.TrimSpace(r.URL.Query().Get("assignedAgencyId"))

	if assignedToMe {
		assignedAgencyID = ctx.ActorAgencyID
	}

	writeJSON(w, http.StatusOK, incidentListResponse{
		Incidents: sanitizeIncidentsForAuthority(s.store.listIncidents(assignedAgencyID), ctx),
	})
}

func (s *server) duplicateReviewHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, incidentReadRoles)
	if !ok {
		return
	}

	payload, code, message := s.store.duplicateReview(r.PathValue("id"))
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, sanitizeDuplicateReviewForAuthority(payload, ctx))
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
	writeJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
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
	writeJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
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
	writeJSON(w, http.StatusOK, sanitizeMergeResponseForAuthority(payload, ctx))
}

func (s *server) reviewAbuseHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, abuseReviewRoles)
	if !ok {
		return
	}

	var request abuseReviewRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAbuseReviewRequest(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	incident, code, message := s.store.reviewAbuse(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
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
	writeJSON(w, http.StatusCreated, sanitizeIncidentForAuthority(incident, ctx))
}

func (s *server) registerVolunteerHandler(w http.ResponseWriter, r *http.Request) {
	var request registerVolunteerRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN incident-service volunteer_register invalid_json remote=%s error=%v", clientIdentifier(r), err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	log.Printf("INFO incident-service volunteer_register received citizenUserId=%s district=%s community=%s", request.CitizenUserID, request.District, request.Community)

	normalized, code, message := normalizeVolunteerRegistrationRequest(request)
	if code != "" {
		log.Printf("WARN incident-service volunteer_register validation_failed citizenUserId=%s code=%s", request.CitizenUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	volunteer := s.store.registerVolunteer(normalized, s.now())
	log.Printf("INFO incident-service volunteer_register created volunteerId=%s groupId=%s verificationStatus=%s", volunteer.ID, volunteer.GroupID, volunteer.VerificationStatus)
	writeJSON(w, http.StatusCreated, volunteerProfileResponse{Volunteer: volunteer})
}

func (s *server) listVolunteersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))
	district := strings.TrimSpace(r.URL.Query().Get("district"))
	volunteers := s.store.listVolunteers(status, district)
	log.Printf("INFO incident-service volunteer_list actor=%s role=%s status=%s district=%s count=%d", ctx.ActorUserID, ctx.ActorRole, status, district, len(volunteers))
	writeJSON(w, http.StatusOK, volunteerListResponse{Volunteers: volunteers})
}

func (s *server) verifyVolunteerHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}

	var request verifyVolunteerRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN incident-service volunteer_verify invalid_json volunteerId=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeVolunteerVerifyRequest(request)
	if code != "" {
		log.Printf("WARN incident-service volunteer_verify validation_failed volunteerId=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	volunteer, code, message := s.store.verifyVolunteer(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		log.Printf("WARN incident-service volunteer_verify failed volunteerId=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO incident-service volunteer_verify completed volunteerId=%s decision=%s actor=%s", volunteer.ID, normalized.Decision, ctx.ActorUserID)
	writeJSON(w, http.StatusOK, volunteerProfileResponse{Volunteer: volunteer})
}

func (s *server) listVolunteerTasksHandler(w http.ResponseWriter, r *http.Request) {
	volunteerID := strings.TrimSpace(r.PathValue("id"))
	tasks, code, message := s.store.listVolunteerTasks(volunteerID)
	if code != "" {
		log.Printf("WARN incident-service volunteer_tasks failed volunteerId=%s code=%s", volunteerID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO incident-service volunteer_tasks listed volunteerId=%s count=%d", volunteerID, len(tasks))
	writeJSON(w, http.StatusOK, volunteerTaskListResponse{Tasks: tasks})
}

func (s *server) assignVolunteerTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}

	var request volunteerTaskRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN incident-service volunteer_task_assign invalid_json incidentId=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	log.Printf("INFO incident-service volunteer_task_assign received incidentId=%s volunteerId=%s type=%s actor=%s", r.PathValue("id"), request.VolunteerID, request.Type, ctx.ActorUserID)

	normalized, code, message := normalizeVolunteerTaskRequest(request)
	if code != "" {
		log.Printf("WARN incident-service volunteer_task_assign validation_failed incidentId=%s volunteerId=%s code=%s", r.PathValue("id"), request.VolunteerID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	task, code, message := s.store.assignVolunteerTask(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		level := "WARN"
		if code == "store_error" {
			level = "ERROR"
		}
		log.Printf("%s incident-service volunteer_task_assign failed incidentId=%s volunteerId=%s code=%s", level, r.PathValue("id"), normalized.VolunteerID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO incident-service volunteer_task_assign completed incidentId=%s taskId=%s volunteerId=%s status=%s", task.IncidentID, task.ID, task.VolunteerID, task.Status)
	writeJSON(w, http.StatusCreated, task)
}

func (s *server) updateVolunteerTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	var request volunteerTaskStatusRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN incident-service volunteer_task_status invalid_json taskId=%s error=%v", r.PathValue("id"), err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	log.Printf("INFO incident-service volunteer_task_status received taskId=%s volunteerId=%s status=%s", r.PathValue("id"), request.VolunteerID, request.Status)

	normalized, code, message := normalizeVolunteerTaskStatusRequest(request)
	if code != "" {
		log.Printf("WARN incident-service volunteer_task_status validation_failed taskId=%s volunteerId=%s code=%s", r.PathValue("id"), request.VolunteerID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	task, code, message := s.store.updateVolunteerTaskStatus(r.PathValue("id"), normalized, s.now())
	if code != "" {
		log.Printf("WARN incident-service volunteer_task_status failed taskId=%s volunteerId=%s code=%s", r.PathValue("id"), normalized.VolunteerID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO incident-service volunteer_task_status completed taskId=%s volunteerId=%s status=%s escalation=%t", task.ID, task.VolunteerID, task.Status, task.EscalationRequired)
	writeJSON(w, http.StatusOK, task)
}

func (s *server) submitVolunteerObservationHandler(w http.ResponseWriter, r *http.Request) {
	var request volunteerObservationRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN incident-service volunteer_observation invalid_json taskId=%s error=%v", r.PathValue("id"), err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	log.Printf("INFO incident-service volunteer_observation received taskId=%s volunteerId=%s safetyStatus=%s escalationRequested=%t", r.PathValue("id"), request.VolunteerID, request.SafetyStatus, request.EscalationRequested)

	normalized, code, message := normalizeVolunteerObservationRequest(request)
	if code != "" {
		log.Printf("WARN incident-service volunteer_observation validation_failed taskId=%s volunteerId=%s code=%s", r.PathValue("id"), request.VolunteerID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	task, code, message := s.store.addVolunteerObservation(r.PathValue("id"), normalized, s.now())
	if code != "" {
		log.Printf("WARN incident-service volunteer_observation failed taskId=%s volunteerId=%s code=%s", r.PathValue("id"), normalized.VolunteerID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO incident-service volunteer_observation completed taskId=%s volunteerId=%s escalation=%t updateCount=%d", task.ID, task.VolunteerID, task.EscalationRequired, len(task.Updates))
	writeJSON(w, http.StatusOK, task)
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
	record.AbuseSignals = m.abuseSignalsLocked(record)
	record.AbuseScore = abuseScore(record.AbuseSignals)
	record.AbuseReviewRequired = record.AbuseScore >= abuseReviewThreshold
	if record.AbuseReviewRequired {
		record.AbuseReviewReason = abuseReviewReason(record.AbuseSignals)
		record.Timeline = append(record.Timeline, newTimelineEvent("incident.abuse_flagged", "Suspicious report signals flagged for dispatcher review", authorityContext{}, map[string]string{
			"score":   fmt.Sprintf("%.2f", record.AbuseScore),
			"signals": strings.Join(abuseSignalCodes(record.AbuseSignals), ","),
		}, timestamp))
	}
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
		incident.AbuseReviewRequired = false
		if nextStatus == "closed" {
			action = "incident.closed"
		} else {
			action = "incident.false_reported"
			incident.AbuseReviewDecision = "false_report"
			incident.AbuseReviewReason = note
			incident.AbuseReviewedBy = ctx.ActorUserID
			incident.AbuseReviewedAt = &timestamp
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

func (m *memoryStore) reviewAbuse(id string, request abuseReviewRequest, ctx authorityContext, now time.Time) (incidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return incidentRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return incidentRecord{}, "invalid_transition", "closed and false-report incidents are terminal"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	incident.AbuseReviewDecision = request.Decision
	incident.AbuseReviewReason = request.Note
	incident.AbuseReviewedBy = ctx.ActorUserID
	incident.AbuseReviewedAt = &timestamp
	incident.StatusReason = request.Note
	incident.UpdatedAt = timestamp

	action := "incident.abuse_reviewed"
	message := "Suspicious report review updated"
	metadata := map[string]string{"decision": request.Decision}

	switch request.Decision {
	case "clear":
		incident.AbuseReviewRequired = false
		action = "incident.abuse_cleared"
		message = "Suspicious report signals cleared"
	case "monitor":
		incident.AbuseReviewRequired = true
		action = "incident.abuse_monitored"
		message = "Suspicious report kept under dispatcher monitoring"
	case "false_report":
		if !allowedIncidentTransitions[incident.Status]["false_report"] {
			return incidentRecord{}, "invalid_transition", fmt.Sprintf("cannot move incident from %s to false_report", incident.Status)
		}
		incident.Status = "false_report"
		incident.StatusUpdatedBy = ctx.ActorUserID
		incident.ResolutionNotes = request.ResolutionNotes
		incident.ClosedAt = &timestamp
		incident.AbuseReviewRequired = false
		action = "incident.false_reported"
		message = "Incident marked as false report after abuse review"
		metadata["resolutionNotes"] = request.ResolutionNotes
	}

	incident.Timeline = append(incident.Timeline, newTimelineEvent(action, message, ctx, metadata, timestamp))
	m.incidents[incident.ID] = incident
	m.appendAuditLocked(action, ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	return incident, "", ""
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

func (m *memoryStore) registerVolunteer(request registerVolunteerRequest, now time.Time) volunteerProfile {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.volunteerSequence++
	timestamp := now.UTC()
	volunteer := volunteerProfile{
		ID:                 fmt.Sprintf("vol_%06d", m.volunteerSequence),
		CitizenUserID:      request.CitizenUserID,
		Name:               request.Name,
		Phone:              request.Phone,
		Region:             request.Region,
		District:           request.District,
		Community:          request.Community,
		GroupID:            volunteerGroupID(request.Region, request.District, request.Community),
		Skills:             append([]string{}, request.Skills...),
		Languages:          append([]string{}, request.Languages...),
		AvailabilityStatus: request.AvailabilityStatus,
		VerificationStatus: "pending",
		SafetyNotes:        volunteerSafetyRules(),
		CreatedAt:          timestamp,
		UpdatedAt:          timestamp,
	}
	m.volunteers[volunteer.ID] = volunteer
	m.appendAuditForTargetLocked("volunteer.registered", volunteerActorContext(volunteer), "volunteer_profile", volunteer.ID, nil, snapshotVolunteer(volunteer), timestamp)
	return volunteer
}

func (m *memoryStore) listVolunteers(status string, district string) []volunteerProfile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volunteers := make([]volunteerProfile, 0, len(m.volunteers))
	for _, volunteer := range m.volunteers {
		if status != "" && volunteer.VerificationStatus != status {
			continue
		}
		if district != "" && !strings.EqualFold(volunteer.District, district) {
			continue
		}
		volunteers = append(volunteers, volunteer)
	}
	sort.Slice(volunteers, func(i, j int) bool {
		if volunteers[i].UpdatedAt.Equal(volunteers[j].UpdatedAt) {
			return volunteers[i].ID < volunteers[j].ID
		}
		return volunteers[i].UpdatedAt.After(volunteers[j].UpdatedAt)
	})
	return volunteers
}

func (m *memoryStore) verifyVolunteer(id string, request verifyVolunteerRequest, ctx authorityContext, now time.Time) (volunteerProfile, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	volunteer, ok := m.volunteers[id]
	if !ok {
		return volunteerProfile{}, "not_found", "volunteer profile was not found"
	}

	before := snapshotVolunteer(volunteer)
	timestamp := now.UTC()
	volunteer.VerifiedBy = ctx.ActorUserID
	volunteer.VerifiedAt = &timestamp
	volunteer.UpdatedAt = timestamp
	volunteer.RejectionReason = ""
	switch request.Decision {
	case "verify":
		volunteer.VerificationStatus = "verified"
	case "reject":
		volunteer.VerificationStatus = "rejected"
		volunteer.RejectionReason = request.Note
	case "suspend":
		volunteer.VerificationStatus = "suspended"
		volunteer.RejectionReason = request.Note
	}
	m.volunteers[volunteer.ID] = volunteer
	m.appendAuditForTargetLocked("volunteer."+volunteer.VerificationStatus, ctx, "volunteer_profile", volunteer.ID, before, snapshotVolunteer(volunteer), timestamp)
	return volunteer, "", ""
}

func (m *memoryStore) listVolunteerTasks(volunteerID string) ([]volunteerTaskRecord, string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.volunteers[volunteerID]; !ok {
		return nil, "not_found", "volunteer profile was not found"
	}

	tasks := make([]volunteerTaskRecord, 0, len(m.volunteerTasks))
	for _, task := range m.volunteerTasks {
		if task.VolunteerID == volunteerID {
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
	return tasks, "", ""
}

func (m *memoryStore) assignVolunteerTask(incidentID string, request volunteerTaskRequest, ctx authorityContext, now time.Time) (volunteerTaskRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[incidentID]
	if !ok {
		return volunteerTaskRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "reported" || incident.Status == "under_review" {
		return volunteerTaskRecord{}, "invalid_transition", "incident must be verified before volunteer tasks can be assigned"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return volunteerTaskRecord{}, "invalid_transition", "closed and false-report incidents cannot receive volunteer tasks"
	}
	volunteer, ok := m.volunteers[request.VolunteerID]
	if !ok {
		return volunteerTaskRecord{}, "not_found", "volunteer profile was not found"
	}
	if volunteer.VerificationStatus != "verified" {
		return volunteerTaskRecord{}, "volunteer_not_verified", "volunteer must be verified before task assignment"
	}
	if volunteer.AvailabilityStatus != "available" {
		return volunteerTaskRecord{}, "volunteer_unavailable", "volunteer must be available before task assignment"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	m.volunteerTaskSequence++
	task := volunteerTaskRecord{
		ID:                 fmt.Sprintf("vtask_%06d", m.volunteerTaskSequence),
		IncidentID:         incident.ID,
		IncidentReference:  incident.Reference,
		VolunteerID:        volunteer.ID,
		VolunteerName:      volunteer.Name,
		GroupID:            volunteer.GroupID,
		Type:               request.Type,
		Priority:           request.Priority,
		Instructions:       request.Instructions,
		LocationLabel:      request.LocationLabel,
		Status:             "assigned",
		SafetyRules:        volunteerSafetyRules(),
		EscalationRequired: false,
		AssignedBy:         ctx.ActorUserID,
		AssignedAt:         timestamp,
		UpdatedAt:          timestamp,
		Updates:            []volunteerTaskUpdate{},
	}
	m.volunteerTasks[task.ID] = task

	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_assigned", fmt.Sprintf("Volunteer task assigned to %s", volunteer.Name), ctx, map[string]string{
		"taskId":      task.ID,
		"volunteerId": volunteer.ID,
		"groupId":     volunteer.GroupID,
		"taskType":    task.Type,
		"priority":    task.Priority,
	}, timestamp))
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.volunteer_assigned", ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	m.appendAuditForTargetLocked("volunteer_task.assigned", ctx, "volunteer_task", task.ID, nil, snapshotVolunteerTask(task), timestamp)
	return task, "", ""
}

func (m *memoryStore) updateVolunteerTaskStatus(taskID string, request volunteerTaskStatusRequest, now time.Time) (volunteerTaskRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.volunteerTasks[taskID]
	if !ok {
		return volunteerTaskRecord{}, "not_found", "volunteer task was not found"
	}
	if task.VolunteerID != request.VolunteerID {
		return volunteerTaskRecord{}, "forbidden", "volunteer can update only their own tasks"
	}
	incident, ok := m.incidents[task.IncidentID]
	if !ok {
		return volunteerTaskRecord{}, "not_found", "linked incident was not found"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	task.Status = request.Status
	task.UpdatedAt = timestamp
	if request.Status == "accepted" && task.AcceptedAt == nil {
		task.AcceptedAt = &timestamp
	}
	if request.Status == "completed" {
		task.CompletedAt = &timestamp
	}
	if request.Status == "needs_escalation" || request.SafetyStatus == "unsafe" || request.SafetyStatus == "needs_authority" {
		task.EscalationRequired = true
	}
	update := volunteerTaskUpdate{
		ID:                  fmt.Sprintf("vtup_%06d", len(task.Updates)+1),
		Type:                "status",
		Status:              request.Status,
		Note:                request.Note,
		SafetyStatus:        request.SafetyStatus,
		Location:            request.Location,
		EscalationRequested: task.EscalationRequired,
		CreatedBy:           request.VolunteerID,
		CreatedAt:           timestamp,
	}
	task.Updates = append(task.Updates, update)
	m.volunteerTasks[task.ID] = task

	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_status_updated", fmt.Sprintf("Volunteer task %s", request.Status), volunteerActorContextByID(request.VolunteerID), map[string]string{
		"taskId":       task.ID,
		"volunteerId":  request.VolunteerID,
		"status":       request.Status,
		"safetyStatus": request.SafetyStatus,
	}, timestamp))
	if task.EscalationRequired {
		incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_escalation", "Volunteer requested authority escalation", volunteerActorContextByID(request.VolunteerID), map[string]string{
			"taskId":       task.ID,
			"volunteerId":  request.VolunteerID,
			"safetyStatus": request.SafetyStatus,
		}, timestamp))
	}
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.volunteer_status_updated", volunteerActorContextByID(request.VolunteerID), incident.ID, before, snapshotIncident(incident), timestamp)
	return task, "", ""
}

func (m *memoryStore) addVolunteerObservation(taskID string, request volunteerObservationRequest, now time.Time) (volunteerTaskRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.volunteerTasks[taskID]
	if !ok {
		return volunteerTaskRecord{}, "not_found", "volunteer task was not found"
	}
	if task.VolunteerID != request.VolunteerID {
		return volunteerTaskRecord{}, "forbidden", "volunteer can add observations only to their own tasks"
	}
	incident, ok := m.incidents[task.IncidentID]
	if !ok {
		return volunteerTaskRecord{}, "not_found", "linked incident was not found"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	if request.EscalationRequested || request.SafetyStatus == "unsafe" || request.SafetyStatus == "needs_authority" {
		task.EscalationRequired = true
		task.Status = "needs_escalation"
	}
	update := volunteerTaskUpdate{
		ID:                  fmt.Sprintf("vtup_%06d", len(task.Updates)+1),
		Type:                "observation",
		Status:              task.Status,
		Note:                request.Observation,
		SafetyStatus:        request.SafetyStatus,
		Location:            request.Location,
		EscalationRequested: request.EscalationRequested,
		CreatedBy:           request.VolunteerID,
		CreatedAt:           timestamp,
	}
	task.Updates = append(task.Updates, update)
	task.UpdatedAt = timestamp
	m.volunteerTasks[task.ID] = task

	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_observation", "Volunteer field observation received", volunteerActorContextByID(request.VolunteerID), map[string]string{
		"taskId":       task.ID,
		"volunteerId":  request.VolunteerID,
		"safetyStatus": request.SafetyStatus,
		"mediaCount":   fmt.Sprintf("%d", len(request.Media)),
	}, timestamp))
	if task.EscalationRequired {
		incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_escalation", "Volunteer observation requires authority review", volunteerActorContextByID(request.VolunteerID), map[string]string{
			"taskId":       task.ID,
			"volunteerId":  request.VolunteerID,
			"safetyStatus": request.SafetyStatus,
		}, timestamp))
	}
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.volunteer_observation", volunteerActorContextByID(request.VolunteerID), incident.ID, before, snapshotIncident(incident), timestamp)
	return task, "", ""
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
	m.appendAuditForTargetLocked(action, ctx, "incident", targetID, before, after, now)
}

func (m *memoryStore) appendAuditForTargetLocked(action string, ctx authorityContext, targetType string, targetID string, before map[string]any, after map[string]any, now time.Time) {
	m.audit = append(m.audit, auditEvent{
		ID:            fmt.Sprintf("aud_%06d", len(m.audit)+1),
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Action:        action,
		TargetType:    targetType,
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

func (m *memoryStore) abuseSignalsLocked(record incidentRecord) []abuseSignal {
	signals := abuseSignalsForDescription(record.Description)

	reporterKey := reporterAbuseKey(record.ReportedBy)
	if reporterKey != "" {
		recentReports := 0
		cutoff := record.CreatedAt.Add(-reporterBurstWindow)
		for _, existing := range m.incidents {
			if existing.ID == record.ID || existing.CreatedAt.Before(cutoff) {
				continue
			}
			if reporterAbuseKey(existing.ReportedBy) == reporterKey {
				recentReports++
			}
		}
		if recentReports >= reporterBurstPreviousMin {
			signals = append(signals, abuseSignal{
				Code:   "reporter_burst",
				Label:  "Reporter burst",
				Detail: fmt.Sprintf("Reporter submitted %d other report(s) in the last %d minutes.", recentReports, int(reporterBurstWindow.Minutes())),
				Weight: 0.55,
			})
		}
	}

	sort.Slice(signals, func(i, j int) bool {
		if signals[i].Weight == signals[j].Weight {
			return signals[i].Code < signals[j].Code
		}
		return signals[i].Weight > signals[j].Weight
	})
	return signals
}

func abuseSignalsForDescription(description string) []abuseSignal {
	lower := strings.ToLower(description)
	signals := []abuseSignal{}

	if strings.Contains(lower, "http://") ||
		strings.Contains(lower, "https://") ||
		strings.Contains(lower, "www.") ||
		strings.Contains(lower, "bit.ly") {
		signals = append(signals, abuseSignal{
			Code:   "external_link",
			Label:  "External link",
			Detail: "Description includes a public link, which can indicate spam in citizen reporting.",
			Weight: 0.45,
		})
	}

	if containsAny(lower, []string{"free money", "promo", "promotion", "discount", "loan offer", "click here", "whatsapp me"}) {
		signals = append(signals, abuseSignal{
			Code:   "promotional_language",
			Label:  "Promotional wording",
			Detail: "Description includes marketing or solicitation language uncommon in emergency reports.",
			Weight: 0.35,
		})
	}

	if repeatedTokenRatio(lower) >= 0.50 {
		signals = append(signals, abuseSignal{
			Code:   "repeated_language",
			Label:  "Repeated language",
			Detail: "Description repeats the same terms unusually often.",
			Weight: 0.25,
		})
	}

	if len([]rune(strings.TrimSpace(description))) < 24 {
		signals = append(signals, abuseSignal{
			Code:   "low_detail",
			Label:  "Low detail",
			Detail: "Description is very short and may need dispatcher confirmation.",
			Weight: 0.20,
		})
	}

	return signals
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

func repeatedTokenRatio(value string) float64 {
	tokens := wordPattern.FindAllString(strings.ToLower(value), -1)
	if len(tokens) < 6 {
		return 0
	}

	counts := map[string]int{}
	maxCount := 0
	for _, token := range tokens {
		if len(token) < 3 {
			continue
		}
		counts[token]++
		if counts[token] > maxCount {
			maxCount = counts[token]
		}
	}
	if len(counts) == 0 {
		return 0
	}
	return float64(maxCount) / float64(len(tokens))
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func reporterAbuseKey(reporter *reporterRef) string {
	if reporter == nil {
		return ""
	}
	if reporter.UserID != "" {
		return "user:" + strings.ToLower(reporter.UserID)
	}
	if reporter.Phone != "" {
		return "phone:" + strings.ToLower(reporter.Phone)
	}
	return ""
}

func abuseScore(signals []abuseSignal) float64 {
	score := 0.0
	for _, signal := range signals {
		score += signal.Weight
	}
	return roundScore(clamp01(score))
}

func abuseReviewReason(signals []abuseSignal) string {
	if len(signals) == 0 {
		return ""
	}
	labels := make([]string, 0, len(signals))
	for _, signal := range signals {
		labels = append(labels, signal.Label)
	}
	return "Review requested: " + strings.Join(labels, ", ")
}

func abuseSignalCodes(signals []abuseSignal) []string {
	codes := make([]string, 0, len(signals))
	for _, signal := range signals {
		codes = append(codes, signal.Code)
	}
	return codes
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

func normalizeAbuseReviewRequest(request abuseReviewRequest) (abuseReviewRequest, string, string) {
	request.Decision = incidentStatusSlug(request.Decision)
	request.Note = strings.TrimSpace(request.Note)
	request.ResolutionNotes = strings.TrimSpace(request.ResolutionNotes)

	if !allowedAbuseReviewDecisions[request.Decision] {
		return request, "invalid_decision", "decision must be clear, monitor, or false_report"
	}
	if len(request.Note) < 5 || len(request.Note) > 1000 || unsafeText(request.Note) {
		return request, "invalid_note", "note must be 5 to 1000 safe characters"
	}
	if len(request.ResolutionNotes) > 2000 || unsafeText(request.ResolutionNotes) {
		return request, "invalid_resolution_notes", "resolutionNotes must be 2000 safe characters or fewer"
	}
	if request.Decision == "false_report" && request.ResolutionNotes == "" {
		return request, "missing_resolution_notes", "resolutionNotes are required when an abuse review marks a false report"
	}

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

func normalizeVolunteerRegistrationRequest(request registerVolunteerRequest) (registerVolunteerRequest, string, string) {
	request.CitizenUserID = strings.TrimSpace(request.CitizenUserID)
	request.Name = strings.TrimSpace(request.Name)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Community = strings.TrimSpace(request.Community)
	request.AvailabilityStatus = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.AvailabilityStatus)), "-", "_"), " ", "_")
	request.Skills = normalizeSafeList(request.Skills, 8, 64)
	request.Languages = normalizeSafeList(request.Languages, 6, 32)

	if request.CitizenUserID == "" || !mediaRefPattern.MatchString(request.CitizenUserID) {
		return request, "invalid_citizen_user_id", "citizenUserId is required and must be a safe user reference"
	}
	if len(request.Name) < 2 || len(request.Name) > 120 || unsafeText(request.Name) {
		return request, "invalid_volunteer_name", "name must be 2 to 120 safe characters"
	}
	if request.Phone == "" || len(request.Phone) > 32 || unsafeText(request.Phone) {
		return request, "invalid_volunteer_phone", "phone is required and must be 32 safe characters or fewer"
	}
	if len(request.Region) < 2 || len(request.Region) > 80 || unsafeText(request.Region) {
		return request, "invalid_region", "region must be 2 to 80 safe characters"
	}
	if len(request.District) < 2 || len(request.District) > 100 || unsafeText(request.District) {
		return request, "invalid_district", "district must be 2 to 100 safe characters"
	}
	if len(request.Community) < 2 || len(request.Community) > 100 || unsafeText(request.Community) {
		return request, "invalid_community", "community must be 2 to 100 safe characters"
	}
	if request.AvailabilityStatus == "" {
		request.AvailabilityStatus = "available"
	}
	if !allowedVolunteerAvailability[request.AvailabilityStatus] {
		return request, "invalid_availability", "availabilityStatus must be available, busy, or off_duty"
	}
	if len(request.Skills) == 0 {
		return request, "missing_skills", "at least one volunteer skill is required"
	}
	if len(request.Languages) == 0 {
		request.Languages = []string{"en"}
	}
	return request, "", ""
}

func normalizeVolunteerVerifyRequest(request verifyVolunteerRequest) (verifyVolunteerRequest, string, string) {
	request.Decision = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.Decision)), "-", "_"), " ", "_")
	request.Note = strings.TrimSpace(request.Note)
	if !allowedVolunteerVerificationDecisions[request.Decision] {
		return request, "invalid_decision", "decision must be verify, reject, or suspend"
	}
	if len(request.Note) < 5 || len(request.Note) > 1000 || unsafeText(request.Note) {
		return request, "invalid_note", "note must be 5 to 1000 safe characters"
	}
	return request, "", ""
}

func normalizeVolunteerTaskRequest(request volunteerTaskRequest) (volunteerTaskRequest, string, string) {
	request.VolunteerID = strings.TrimSpace(request.VolunteerID)
	request.Type = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.Type)), "-", "_"), " ", "_")
	request.Priority = strings.TrimSpace(strings.ToLower(request.Priority))
	request.Instructions = strings.TrimSpace(request.Instructions)
	request.LocationLabel = strings.TrimSpace(request.LocationLabel)
	if request.VolunteerID == "" || !mediaRefPattern.MatchString(request.VolunteerID) {
		return request, "invalid_volunteer_id", "volunteerId is required and must be a safe volunteer reference"
	}
	if !allowedVolunteerTaskTypes[request.Type] {
		return request, "invalid_task_type", "type must be welfare_check, shelter_support, supply_distribution, damage_observation, route_observation, or community_alerting"
	}
	if request.Priority == "" {
		request.Priority = "normal"
	}
	if !allowedAssignmentPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, normal, high, or urgent"
	}
	if len(request.Instructions) < 10 || len(request.Instructions) > 1200 || unsafeText(request.Instructions) {
		return request, "invalid_instructions", "instructions must be 10 to 1200 safe characters"
	}
	if unsafeVolunteerInstructions(request.Instructions) {
		return request, "unsafe_volunteer_instructions", "volunteer tasks must not instruct civilians to enter floodwater, fight fires, conduct rescues, or approach violent/structural hazards"
	}
	if len(request.LocationLabel) < 2 || len(request.LocationLabel) > 180 || unsafeText(request.LocationLabel) {
		return request, "invalid_location_label", "locationLabel must be 2 to 180 safe characters"
	}
	return request, "", ""
}

func normalizeVolunteerTaskStatusRequest(request volunteerTaskStatusRequest) (volunteerTaskStatusRequest, string, string) {
	request.VolunteerID = strings.TrimSpace(request.VolunteerID)
	request.Status = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.Status)), "-", "_"), " ", "_")
	request.Note = strings.TrimSpace(request.Note)
	request.SafetyStatus = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.SafetyStatus)), "-", "_"), " ", "_")
	if request.VolunteerID == "" || !mediaRefPattern.MatchString(request.VolunteerID) {
		return request, "invalid_volunteer_id", "volunteerId is required and must be a safe volunteer reference"
	}
	if !allowedVolunteerTaskStatuses[request.Status] {
		return request, "invalid_task_status", "status must be accepted, en_route, on_scene, completed, cancelled, or needs_escalation"
	}
	if len(request.Note) > 1000 || unsafeText(request.Note) {
		return request, "invalid_note", "note must be 1000 safe characters or fewer"
	}
	if request.SafetyStatus == "" {
		request.SafetyStatus = "safe"
	}
	if !allowedVolunteerSafetyStatuses[request.SafetyStatus] {
		return request, "invalid_safety_status", "safetyStatus must be safe, caution, unsafe, or needs_authority"
	}
	if request.Location != nil && !validCoordinates(*request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	return request, "", ""
}

func normalizeVolunteerObservationRequest(request volunteerObservationRequest) (volunteerObservationRequest, string, string) {
	request.VolunteerID = strings.TrimSpace(request.VolunteerID)
	request.Observation = strings.TrimSpace(request.Observation)
	request.SafetyStatus = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.SafetyStatus)), "-", "_"), " ", "_")
	request.Media = normalizeSafeList(request.Media, 8, 128)
	if request.VolunteerID == "" || !mediaRefPattern.MatchString(request.VolunteerID) {
		return request, "invalid_volunteer_id", "volunteerId is required and must be a safe volunteer reference"
	}
	if len(request.Observation) < 5 || len(request.Observation) > 1500 || unsafeText(request.Observation) {
		return request, "invalid_observation", "observation must be 5 to 1500 safe characters"
	}
	if request.SafetyStatus == "" {
		request.SafetyStatus = "safe"
	}
	if !allowedVolunteerSafetyStatuses[request.SafetyStatus] {
		return request, "invalid_safety_status", "safetyStatus must be safe, caution, unsafe, or needs_authority"
	}
	if request.Location != nil && !validCoordinates(*request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
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

func normalizeSafeList(values []string, limit int, maxLen int) []string {
	seen := map[string]bool{}
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || len(value) > maxLen || unsafeText(value) || seen[strings.ToLower(value)] {
			continue
		}
		seen[strings.ToLower(value)] = true
		normalized = append(normalized, value)
		if len(normalized) >= limit {
			break
		}
	}
	return normalized
}

func unsafeVolunteerInstructions(value string) bool {
	lower := strings.ToLower(value)
	unsafePhrases := []string{
		"enter floodwater",
		"walk through flood",
		"wade through",
		"fight fire",
		"put out fire",
		"rescue trapped",
		"enter collapsed",
		"go inside collapsed",
		"approach armed",
		"handle violent",
		"direct traffic on highway",
	}
	for _, phrase := range unsafePhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func volunteerGroupID(region string, district string, community string) string {
	return fmt.Sprintf("grp_%s_%s_%s", slugRef(region), slugRef(district), slugRef(community))
}

func slugRef(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	previousDash := false
	for _, char := range lower {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			previousDash = false
			continue
		}
		if !previousDash && builder.Len() > 0 {
			builder.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func volunteerSafetyRules() []string {
	return []string{
		"Stay in public, safe areas and never enter floodwater, fire zones, collapsed structures, or violent scenes.",
		"Call 112 and request authority escalation for injuries, trapped people, unsafe crowds, or blocked emergency access.",
		"Share observations, photos, and status updates only when doing so does not delay evacuation or personal safety.",
	}
}

func volunteerActorContext(volunteer volunteerProfile) authorityContext {
	return volunteerActorContextByID(volunteer.ID)
}

func volunteerActorContextByID(volunteerID string) authorityContext {
	return authorityContext{
		ActorUserID: volunteerID,
		ActorRole:   "citizen",
	}
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

func sanitizeIncidentsForAuthority(incidents []incidentRecord, ctx authorityContext) []incidentRecord {
	sanitized := make([]incidentRecord, 0, len(incidents))
	for _, incident := range incidents {
		sanitized = append(sanitized, sanitizeIncidentForAuthority(incident, ctx))
	}
	return sanitized
}

func sanitizeDuplicateReviewForAuthority(payload duplicateReviewResponse, ctx authorityContext) duplicateReviewResponse {
	payload.Incident = sanitizeIncidentForAuthority(payload.Incident, ctx)
	for index := range payload.Candidates {
		payload.Candidates[index].Incident = sanitizeIncidentForAuthority(payload.Candidates[index].Incident, ctx)
	}
	return payload
}

func sanitizeMergeResponseForAuthority(payload mergeIncidentsResponse, ctx authorityContext) mergeIncidentsResponse {
	payload.Incident = sanitizeIncidentForAuthority(payload.Incident, ctx)
	payload.MergedIncidents = sanitizeIncidentsForAuthority(payload.MergedIncidents, ctx)
	return payload
}

func sanitizeIncidentForAuthority(incident incidentRecord, ctx authorityContext) incidentRecord {
	privacy := privacyForIncident(incident, ctx)
	incident.Privacy = privacy

	if !privacy.ReporterIdentityVisible {
		incident.ReportedBy = nil
		return incident
	}

	if !privacy.ReporterContactVisible && incident.ReportedBy != nil {
		reporter := *incident.ReportedBy
		reporter.Phone = ""
		incident.ReportedBy = &reporter
	}
	return incident
}

func privacyForIncident(incident incidentRecord, ctx authorityContext) incidentPrivacy {
	canViewContact := reporterContactRoles[ctx.ActorRole]
	hasReporter := incident.ReportedBy != nil
	reporterIdentityVisible := hasReporter && !incident.Anonymous && incident.ContactPermission && canViewContact
	reporterContactVisible := reporterIdentityVisible && incident.ReportedBy.Phone != ""

	notes := []string{
		"Exact incident location is available only to MFA-verified authority users for emergency response coordination.",
	}
	if incident.Anonymous {
		notes = append(notes, "Reporter chose anonymous reporting; citizen identity is hidden in authority views.")
	}
	if !incident.ContactPermission {
		notes = append(notes, "Reporter did not grant contact permission; contact details are hidden.")
	}
	if hasReporter && !canViewContact {
		notes = append(notes, "Current authority role receives a standard operational view without reporter contact details.")
	}

	return incidentPrivacy{
		ReporterIdentityVisible: reporterIdentityVisible,
		ReporterContactVisible:  reporterContactVisible,
		LocationPrecision:       "exact",
		LocationUse:             "emergency_response",
		Disclosure:              "Location is used to route emergency response, detect duplicates, and coordinate verified authority actions.",
		Notes:                   notes,
	}
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
		"id":                  incident.ID,
		"reference":           incident.Reference,
		"type":                incident.Type,
		"severity":            incident.Severity,
		"status":              incident.Status,
		"priorityReview":      incident.PriorityReview,
		"verifiedBy":          incident.VerifiedBy,
		"statusUpdatedBy":     incident.StatusUpdatedBy,
		"statusReason":        incident.StatusReason,
		"resolutionNotes":     incident.ResolutionNotes,
		"abuseScore":          incident.AbuseScore,
		"abuseReviewRequired": incident.AbuseReviewRequired,
		"abuseReviewDecision": incident.AbuseReviewDecision,
		"mergedIncidentIds":   append([]string{}, incident.MergedIncidentIDs...),
		"mergedIntoId":        incident.MergedIntoID,
		"mergeReason":         incident.MergeReason,
		"duplicateCount":      len(incident.DuplicateCandidates),
		"assignmentCount":     len(incident.Assignments),
		"assignedAgencyIds":   assignmentAgencyIDs(incident.Assignments),
	}
}

func snapshotVolunteer(volunteer volunteerProfile) map[string]any {
	return map[string]any{
		"id":                 volunteer.ID,
		"citizenUserId":      volunteer.CitizenUserID,
		"groupId":            volunteer.GroupID,
		"district":           volunteer.District,
		"community":          volunteer.Community,
		"skills":             append([]string{}, volunteer.Skills...),
		"availabilityStatus": volunteer.AvailabilityStatus,
		"verificationStatus": volunteer.VerificationStatus,
		"verifiedBy":         volunteer.VerifiedBy,
		"rejectionReason":    volunteer.RejectionReason,
	}
}

func snapshotVolunteerTask(task volunteerTaskRecord) map[string]any {
	return map[string]any{
		"id":                 task.ID,
		"incidentId":         task.IncidentID,
		"incidentReference":  task.IncidentReference,
		"volunteerId":        task.VolunteerID,
		"groupId":            task.GroupID,
		"type":               task.Type,
		"priority":           task.Priority,
		"status":             task.Status,
		"escalationRequired": task.EscalationRequired,
		"updateCount":        len(task.Updates),
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
		log.Printf("ERROR incident-service write_json_response_failed error=%v", err)
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
