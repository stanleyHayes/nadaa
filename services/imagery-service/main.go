package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxUploadSize        = 20 << 20 // 20 MB
	defaultRetentionDays = 90
)

var (
	allowedSources = map[string]bool{
		"drone":     true,
		"satellite": true,
		"other":     true,
	}

	allowedStatuses = map[string]bool{
		"active":  true,
		"expired": true,
	}

	allowedAuthorityRoles = map[string]bool{
		"system_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"police":           true,
		"fire":             true,
		"ambulance":        true,
		"rescue":           true,
		"analyst":          true,
	}
)

type config struct {
	storagePath   string
	retentionDays int
}

type server struct {
	store  *memoryStore
	now    func() time.Time
	config config
}

type memoryStore struct {
	mu      sync.RWMutex
	records []imageryRecord
}

type imageryRecord struct {
	ID                string          `json:"id"`
	Reference         string          `json:"reference"`
	Source            string          `json:"source"`
	CaptureTime       time.Time       `json:"captureTime"`
	Geometry          json.RawMessage `json:"geometry"`
	CoverageAreaKm2   float64         `json:"coverageAreaKm2"`
	ResolutionMeters  float64         `json:"resolutionMeters"`
	License           string          `json:"license,omitempty"`
	RelatedIncidentID string          `json:"relatedIncidentId,omitempty"`
	RelatedRiskZoneID string          `json:"relatedRiskZoneId,omitempty"`
	MlWorkflowID      string          `json:"mlWorkflowId,omitempty"`
	FileName          string          `json:"fileName"`
	ContentType       string          `json:"contentType"`
	SizeBytes         int64           `json:"sizeBytes"`
	StoragePath       string          `json:"storagePath"`
	Status            string          `json:"status"`
	UploadedBy        string          `json:"uploadedBy"`
	CreatedAt         time.Time       `json:"createdAt"`
	ExpiresAt         time.Time       `json:"expiresAt"`
}

type createImageryRequest struct {
	Source            string
	CaptureTime       string
	Geometry          string
	CoverageAreaKm2   string
	ResolutionMeters  string
	License           string
	RelatedIncidentID string
	RelatedRiskZoneID string
	MlWorkflowID      string
}

type imageryUploadInput struct {
	Source            string
	CaptureTime       time.Time
	Geometry          json.RawMessage
	CoverageAreaKm2   float64
	ResolutionMeters  float64
	License           string
	RelatedIncidentID string
	RelatedRiskZoneID string
	MlWorkflowID      string
}

type imageryListFilter struct {
	Source            string
	Status            string
	RelatedIncidentID string
	RelatedRiskZoneID string
	Query             string
}

type imageryListResponse struct {
	Imagery     []imageryRecord `json:"imagery"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

type imageryLifecycleResponse struct {
	ExpiredCount int `json:"expiredCount"`
}

type geoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Type       string          `json:"type"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties map[string]any  `json:"properties"`
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

func main() {
	cfg := loadConfig()
	if err := os.MkdirAll(cfg.storagePath, 0o755); err != nil {
		log.Fatalf("ERROR imagery-service storage_path_create_failed path=%s error=%v", cfg.storagePath, err)
	}

	srv := newServer(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/imagery", srv.createImageryHandler)
	mux.HandleFunc("GET /api/v1/imagery", srv.listImageryHandler)
	mux.HandleFunc("GET /api/v1/imagery/geojson", srv.geoJSONHandler)
	mux.HandleFunc("POST /api/v1/imagery/lifecycle/run", srv.runLifecycleHandler)
	mux.HandleFunc("GET /api/v1/imagery/{id}", srv.getImageryHandler)
	mux.HandleFunc("GET /api/v1/imagery/{id}/download", srv.downloadImageryHandler)
	mux.HandleFunc("DELETE /api/v1/imagery/{id}", srv.deleteImageryHandler)
	mux.HandleFunc("POST /api/v1/imagery/{id}/expire", srv.expireImageryHandler)

	addr := envOrDefault("PORT", ":8099")
	log.Printf("INFO imagery-service listening on %s storage=%s retention_days=%d", addr, cfg.storagePath, cfg.retentionDays)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() config {
	retentionDays, err := strconv.Atoi(envOrDefault("DEFAULT_RETENTION_DAYS", strconv.Itoa(defaultRetentionDays)))
	if err != nil || retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}
	return config{
		storagePath:   envOrDefault("IMAGERY_STORAGE_PATH", "./uploads"),
		retentionDays: retentionDays,
	}
}

func newServer(cfg config) *server {
	now := time.Now
	return &server{
		store:  newMemoryStore(now().UTC()),
		now:    now,
		config: cfg,
	}
}

func newMemoryStore(now time.Time) *memoryStore {
	activeGeometry := json.RawMessage(`{"type":"Polygon","coordinates":[[[-0.21,5.55],[-0.18,5.55],[-0.18,5.58],[-0.21,5.58],[-0.21,5.55]]]}`)
	expiringGeometry := json.RawMessage(`{"type":"Polygon","coordinates":[[[-0.25,5.60],[-0.22,5.60],[-0.22,5.63],[-0.25,5.63],[-0.25,5.60]]]}`)

	return &memoryStore{
		records: []imageryRecord{
			{
				ID:               "img_seed_active",
				Reference:        "NADAA-IMG-img_seed_active",
				Source:           "satellite",
				CaptureTime:      now.Add(-24 * time.Hour),
				Geometry:         activeGeometry,
				CoverageAreaKm2:  12.5,
				ResolutionMeters: 10.0,
				License:          "CC-BY-4.0 NADMO",
				FileName:         "seed-active.tif",
				ContentType:      "image/tiff",
				SizeBytes:        1024,
				StoragePath:      "uploads/seed-active.tif",
				Status:           "active",
				UploadedBy:       "usr_imagery_admin",
				CreatedAt:        now.Add(-30 * 24 * time.Hour),
				ExpiresAt:        now.Add(60 * 24 * time.Hour),
			},
			{
				ID:                "img_seed_expiring",
				Reference:         "NADAA-IMG-img_seed_expiring",
				Source:            "drone",
				CaptureTime:       now.Add(-48 * time.Hour),
				Geometry:          expiringGeometry,
				CoverageAreaKm2:   2.1,
				ResolutionMeters:  0.5,
				License:           "Internal",
				RelatedIncidentID: "incident_001",
				FileName:          "seed-expiring.jpg",
				ContentType:       "image/jpeg",
				SizeBytes:         2048,
				StoragePath:       "uploads/seed-expiring.jpg",
				Status:            "active",
				UploadedBy:        "usr_drone_operator",
				CreatedAt:         now.Add(-100 * 24 * time.Hour),
				ExpiresAt:         now.Add(-1 * time.Hour),
			},
		},
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "imagery-service"})
}

func (s *server) createImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	if r.ContentLength > maxUploadSize {
		log.Printf("WARN imagery-service upload_validation_failed code=payload_too_large actor=%s size=%d", ctx.ActorUserID, r.ContentLength)
		writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "file exceeds 20 MB limit")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if strings.Contains(err.Error(), "too large") {
			log.Printf("WARN imagery-service upload_validation_failed code=payload_too_large actor=%s error=%v", ctx.ActorUserID, err)
			writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "file exceeds 20 MB limit")
		} else {
			log.Printf("WARN imagery-service upload_validation_failed code=invalid_multipart actor=%s error=%v", ctx.ActorUserID, err)
			writeError(w, http.StatusBadRequest, "invalid_multipart", "request must be valid multipart/form-data")
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("WARN imagery-service upload_validation_failed code=missing_file actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "missing_file", "multipart field 'file' is required")
		return
	}
	defer file.Close()

	if fileHeader.Size > maxUploadSize {
		log.Printf("WARN imagery-service upload_validation_failed code=payload_too_large actor=%s filename=%s size=%d", ctx.ActorUserID, fileHeader.Filename, fileHeader.Size)
		writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "file exceeds 20 MB limit")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("ERROR imagery-service upload_read_failed actor=%s filename=%s error=%v", ctx.ActorUserID, fileHeader.Filename, err)
		writeError(w, http.StatusInternalServerError, "read_failed", "could not read uploaded file")
		return
	}

	contentType := strings.TrimSpace(strings.ToLower(fileHeader.Header.Get("Content-Type")))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(contentType, "image/") {
		log.Printf("WARN imagery-service upload_validation_failed code=invalid_content_type actor=%s filename=%s content_type=%s", ctx.ActorUserID, fileHeader.Filename, contentType)
		writeError(w, http.StatusBadRequest, "invalid_content_type", "uploaded file must be an image")
		return
	}

	input, code, message := parseCreateImageryRequest(r)
	if code != "" {
		log.Printf("WARN imagery-service upload_validation_failed code=%s actor=%s filename=%s", code, ctx.ActorUserID, fileHeader.Filename)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	id := fmt.Sprintf("img_%03d", len(s.store.records)+1)
	originalName := filepath.Base(fileHeader.Filename)
	if originalName == "" || originalName == "." || originalName == "/" {
		originalName = "unnamed"
	}
	diskName := fmt.Sprintf("%s-%s", id, originalName)
	storagePath := filepath.Join(s.config.storagePath, diskName)

	if err := os.WriteFile(storagePath, data, 0o644); err != nil {
		log.Printf("ERROR imagery-service upload_write_failed actor=%s filename=%s path=%s error=%v", ctx.ActorUserID, fileHeader.Filename, storagePath, err)
		writeError(w, http.StatusInternalServerError, "write_failed", "could not save uploaded file")
		return
	}

	now := s.now().UTC()
	record := imageryRecord{
		ID:                id,
		Reference:         fmt.Sprintf("NADAA-IMG-%s", id),
		Source:            input.Source,
		CaptureTime:       input.CaptureTime,
		Geometry:          input.Geometry,
		CoverageAreaKm2:   input.CoverageAreaKm2,
		ResolutionMeters:  input.ResolutionMeters,
		License:           input.License,
		RelatedIncidentID: input.RelatedIncidentID,
		RelatedRiskZoneID: input.RelatedRiskZoneID,
		MlWorkflowID:      input.MlWorkflowID,
		FileName:          originalName,
		ContentType:       contentType,
		SizeBytes:         int64(len(data)),
		StoragePath:       storagePath,
		Status:            "active",
		UploadedBy:        ctx.ActorUserID,
		CreatedAt:         now,
		ExpiresAt:         now.Add(time.Duration(s.config.retentionDays) * 24 * time.Hour),
	}

	s.store.create(record)
	log.Printf("INFO imagery-service upload_completed id=%s actor=%s source=%s filename=%s size=%d content_type=%s", record.ID, ctx.ActorUserID, record.Source, record.FileName, record.SizeBytes, record.ContentType)
	writeJSON(w, http.StatusCreated, record)
}

func parseCreateImageryRequest(r *http.Request) (imageryUploadInput, string, string) {
	input := imageryUploadInput{
		Source:            normalizeToken(r.FormValue("source")),
		License:           strings.TrimSpace(r.FormValue("license")),
		RelatedIncidentID: strings.TrimSpace(r.FormValue("relatedIncidentId")),
		RelatedRiskZoneID: strings.TrimSpace(r.FormValue("relatedRiskZoneId")),
		MlWorkflowID:      strings.TrimSpace(r.FormValue("mlWorkflowId")),
	}

	if input.Source == "" || !allowedSources[input.Source] {
		return input, "invalid_source", "source must be drone, satellite, or other"
	}

	captureTimeText := strings.TrimSpace(r.FormValue("captureTime"))
	if captureTimeText == "" {
		return input, "missing_capture_time", "captureTime is required"
	}
	captureTime, err := time.Parse(time.RFC3339, captureTimeText)
	if err != nil {
		return input, "invalid_capture_time", "captureTime must be a valid ISO 8601 timestamp"
	}
	input.CaptureTime = captureTime.UTC()

	geometryText := strings.TrimSpace(r.FormValue("geometry"))
	if geometryText == "" {
		return input, "missing_geometry", "geometry is required"
	}
	if !json.Valid([]byte(geometryText)) {
		return input, "invalid_geometry", "geometry must be valid JSON"
	}
	var geometry struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(geometryText), &geometry); err != nil {
		return input, "invalid_geometry", "geometry could not be parsed"
	}
	if normalizeToken(geometry.Type) != "polygon" {
		return input, "invalid_geometry", "geometry type must be Polygon"
	}
	input.Geometry = json.RawMessage(geometryText)

	coverageText := strings.TrimSpace(r.FormValue("coverageAreaKm2"))
	if coverageText == "" {
		return input, "missing_coverage_area", "coverageAreaKm2 is required"
	}
	coverage, err := strconv.ParseFloat(coverageText, 64)
	if err != nil || coverage < 0 {
		return input, "invalid_coverage_area", "coverageAreaKm2 must be a non-negative number"
	}
	input.CoverageAreaKm2 = coverage

	resolutionText := strings.TrimSpace(r.FormValue("resolutionMeters"))
	if resolutionText == "" {
		return input, "missing_resolution", "resolutionMeters is required"
	}
	resolution, err := strconv.ParseFloat(resolutionText, 64)
	if err != nil || resolution < 0 {
		return input, "invalid_resolution", "resolutionMeters must be a non-negative number"
	}
	input.ResolutionMeters = resolution

	if len(input.License) > 200 || unsafeText(input.License) {
		return input, "invalid_license", "license must be 200 safe characters or fewer"
	}
	for _, value := range []struct {
		name  string
		field string
	}{
		{"relatedIncidentId", input.RelatedIncidentID},
		{"relatedRiskZoneId", input.RelatedRiskZoneID},
		{"mlWorkflowId", input.MlWorkflowID},
	} {
		if len(value.field) > 128 || unsafeText(value.field) {
			return input, "invalid_" + value.name, value.name + " must be 128 safe characters or fewer"
		}
	}

	return input, "", ""
}

func (s *server) listImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	filter, ok := parseImageryListFilter(w, r)
	if !ok {
		return
	}

	records := s.store.list(filter)
	log.Printf("INFO imagery-service list count=%d actor=%s source=%s status=%s relatedIncidentId=%s relatedRiskZoneId=%s q=%s", len(records), ctx.ActorUserID, filter.Source, filter.Status, filter.RelatedIncidentID, filter.RelatedRiskZoneID, filter.Query)
	writeJSON(w, http.StatusOK, imageryListResponse{Imagery: records, GeneratedAt: s.now().UTC()})
}

func parseImageryListFilter(w http.ResponseWriter, r *http.Request) (imageryListFilter, bool) {
	query := r.URL.Query()
	filter := imageryListFilter{
		Source:            normalizeToken(query.Get("source")),
		Status:            normalizeToken(query.Get("status")),
		RelatedIncidentID: strings.TrimSpace(query.Get("relatedIncidentId")),
		RelatedRiskZoneID: strings.TrimSpace(query.Get("relatedRiskZoneId")),
		Query:             strings.TrimSpace(query.Get("q")),
	}

	if filter.Source != "" && !allowedSources[filter.Source] {
		writeError(w, http.StatusBadRequest, "invalid_source", "source must be drone, satellite, or other")
		return filter, false
	}
	if filter.Status != "" && !allowedStatuses[filter.Status] {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be active or expired")
		return filter, false
	}

	return filter, true
}

func (s *server) getImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.getByID(id)
	if !found {
		log.Printf("WARN imagery-service get_not_found id=%s actor=%s", id, ctx.ActorUserID)
		writeError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}
	log.Printf("INFO imagery-service get id=%s actor=%s source=%s status=%s", record.ID, ctx.ActorUserID, record.Source, record.Status)
	writeJSON(w, http.StatusOK, record)
}

func (s *server) downloadImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.getByID(id)
	if !found {
		log.Printf("WARN imagery-service download_not_found id=%s actor=%s", id, ctx.ActorUserID)
		writeError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}

	file, err := os.Open(record.StoragePath)
	if err != nil {
		log.Printf("ERROR imagery-service download_open_failed id=%s actor=%s path=%s error=%v", record.ID, ctx.ActorUserID, record.StoragePath, err)
		writeError(w, http.StatusNotFound, "file_not_found", "stored imagery file was not found")
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		log.Printf("ERROR imagery-service download_stat_failed id=%s actor=%s path=%s error=%v", record.ID, ctx.ActorUserID, record.StoragePath, err)
		writeError(w, http.StatusInternalServerError, "file_stat_failed", "could not stat stored imagery file")
		return
	}

	w.Header().Set("Content-Type", record.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", record.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	log.Printf("INFO imagery-service download id=%s actor=%s filename=%s size=%s", record.ID, ctx.ActorUserID, record.FileName, strconv.FormatInt(stat.Size(), 10))
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("ERROR imagery-service download_copy_failed id=%s actor=%s error=%v", record.ID, ctx.ActorUserID, err)
	}
}

func (s *server) deleteImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.getByID(id)
	if !found {
		log.Printf("WARN imagery-service delete_not_found id=%s actor=%s", id, ctx.ActorUserID)
		writeError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}

	if err := os.Remove(record.StoragePath); err != nil && !os.IsNotExist(err) {
		log.Printf("ERROR imagery-service delete_file_failed id=%s actor=%s path=%s error=%v", record.ID, ctx.ActorUserID, record.StoragePath, err)
		writeError(w, http.StatusInternalServerError, "delete_failed", "could not delete stored imagery file")
		return
	}

	s.store.delete(id)
	log.Printf("INFO imagery-service delete id=%s actor=%s filename=%s", record.ID, ctx.ActorUserID, record.FileName)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) expireImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.expire(id, s.now().UTC())
	if !found {
		log.Printf("WARN imagery-service expire_not_found id=%s actor=%s", id, ctx.ActorUserID)
		writeError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}
	log.Printf("INFO imagery-service expire id=%s actor=%s status=%s", record.ID, ctx.ActorUserID, record.Status)
	writeJSON(w, http.StatusOK, record)
}

func (s *server) runLifecycleHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, allowedAuthorityRoles)
	if !ok {
		return
	}

	count := s.store.runLifecycle(s.now().UTC())
	log.Printf("INFO imagery-service lifecycle actor=%s expired_count=%d", ctx.ActorUserID, count)
	writeJSON(w, http.StatusOK, imageryLifecycleResponse{ExpiredCount: count})
}

func (s *server) geoJSONHandler(w http.ResponseWriter, r *http.Request) {
	records := s.store.listActive()
	features := make([]geoJSONFeature, 0, len(records))
	for _, record := range records {
		features = append(features, geoJSONFeature{
			Type:     "Feature",
			Geometry: record.Geometry,
			Properties: map[string]any{
				"id":               record.ID,
				"reference":        record.Reference,
				"source":           record.Source,
				"captureTime":      record.CaptureTime,
				"resolutionMeters": record.ResolutionMeters,
				"downloadUrl":      fmt.Sprintf("%s://%s/api/v1/imagery/%s/download", scheme(r), r.Host, record.ID),
			},
		})
	}
	log.Printf("INFO imagery-service geojson count=%d", len(features))
	writeJSON(w, http.StatusOK, geoJSONFeatureCollection{Type: "FeatureCollection", Features: features})
}

func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func (m *memoryStore) create(record imageryRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, record)
}

func (m *memoryStore) list(filter imageryListFilter) []imageryRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]imageryRecord, 0)
	for _, record := range m.records {
		if filter.Source != "" && record.Source != filter.Source {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if filter.RelatedIncidentID != "" && record.RelatedIncidentID != filter.RelatedIncidentID {
			continue
		}
		if filter.RelatedRiskZoneID != "" && record.RelatedRiskZoneID != filter.RelatedRiskZoneID {
			continue
		}
		if filter.Query != "" && !recordMatchesQuery(record, strings.ToLower(filter.Query)) {
			continue
		}
		records = append(records, copyRecord(record))
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records
}

func (m *memoryStore) listActive() []imageryRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]imageryRecord, 0)
	for _, record := range m.records {
		if record.Status == "active" {
			records = append(records, copyRecord(record))
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records
}

func recordMatchesQuery(record imageryRecord, query string) bool {
	return strings.Contains(strings.ToLower(record.Reference), query) ||
		strings.Contains(strings.ToLower(record.FileName), query) ||
		strings.Contains(strings.ToLower(record.License), query) ||
		strings.Contains(strings.ToLower(record.RelatedIncidentID), query) ||
		strings.Contains(strings.ToLower(record.RelatedRiskZoneID), query) ||
		strings.Contains(strings.ToLower(record.MlWorkflowID), query)
}

func (m *memoryStore) getByID(id string) (imageryRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, record := range m.records {
		if record.ID == id {
			return copyRecord(record), true
		}
	}
	return imageryRecord{}, false
}

func (m *memoryStore) delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index, record := range m.records {
		if record.ID == id {
			m.records = append(m.records[:index], m.records[index+1:]...)
			return true
		}
	}
	return false
}

func (m *memoryStore) expire(id string, now time.Time) (imageryRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.records {
		if m.records[index].ID != id {
			continue
		}
		m.records[index].Status = "expired"
		return copyRecord(m.records[index]), true
	}
	return imageryRecord{}, false
}

func (m *memoryStore) runLifecycle(now time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for index := range m.records {
		if m.records[index].Status != "active" {
			continue
		}
		if now.After(m.records[index].ExpiresAt) {
			m.records[index].Status = "expired"
			count++
		}
	}
	return count
}

func copyRecord(record imageryRecord) imageryRecord {
	record.Geometry = append(json.RawMessage(nil), record.Geometry...)
	return record
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
		log.Printf("WARN imagery-service authority_context_missing requestId=%s path=%s", ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return authorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN imagery-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for imagery operations")
		return authorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		log.Printf("WARN imagery-service authority_forbidden actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to perform imagery operations")
		return authorityContext{}, false
	}
	return ctx, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR imagery-service write_json_response_failed error=%v", err)
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
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

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func normalizeToken(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(value)), "-", "_"), " ", "_")
}

func unsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}
