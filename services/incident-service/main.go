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
	ReportedBy          *reporterRef         `json:"reportedBy,omitempty"`
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

func (s *server) listIncidentsHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, incidentListResponse{Incidents: s.store.listIncidents()})
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
		CreatedAt:          timestamp,
		UpdatedAt:          timestamp,
	}
	record.DuplicateCandidates = m.duplicateCandidatesLocked(record)
	m.incidents[record.ID] = record
	m.linkReverseDuplicateCandidatesLocked(record)
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
