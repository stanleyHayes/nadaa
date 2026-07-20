package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/utils"
)

const maxUploadSize = 20 << 20 // 20 MB

const (
	// maxGeometryBytes caps the raw geometry payload. The geometry is stored
	// and echoed verbatim on the public geojson feed, so it must stay small.
	maxGeometryBytes = 64 << 10 // 64 KiB
	// maxGeometryPositions caps the total number of coordinate positions
	// across all rings of an uploaded polygon.
	maxGeometryPositions = 10000
)

var allowedSources = map[string]bool{
	"drone":     true,
	"satellite": true,
	"other":     true,
}

var allowedStatuses = map[string]bool{
	"active":  true,
	"expired": true,
}

func (s *Server) createImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	if r.ContentLength > maxUploadSize {
		// #nosec G706 -- the actor id is sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service upload_validation_failed code=payload_too_large actor=%s size=%d", utils.SafeLogValue(ctx.ActorUserID), r.ContentLength)
		utils.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "file exceeds 20 MB limit")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		if strings.Contains(err.Error(), "too large") {
			// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
			log.Printf("WARN imagery-service upload_validation_failed code=payload_too_large actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
			utils.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "file exceeds 20 MB limit")
		} else {
			// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
			log.Printf("WARN imagery-service upload_validation_failed code=invalid_multipart actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
			utils.WriteError(w, http.StatusBadRequest, "invalid_multipart", "request must be valid multipart/form-data")
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service upload_validation_failed code=missing_file actor=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), err)
		utils.WriteError(w, http.StatusBadRequest, "missing_file", "multipart field 'file' is required")
		return
	}
	defer func() { _ = file.Close() }()

	if fileHeader.Size > maxUploadSize {
		// #nosec G706 -- actor id and filename are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service upload_validation_failed code=payload_too_large actor=%s filename=%s size=%d", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(fileHeader.Filename), fileHeader.Size)
		utils.WriteError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "file exceeds 20 MB limit")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		// #nosec G706 -- actor id and filename are sanitized with utils.SafeLogValue.
		log.Printf("ERROR imagery-service upload_read_failed actor=%s filename=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(fileHeader.Filename), err)
		utils.WriteError(w, http.StatusInternalServerError, "read_failed", "could not read uploaded file")
		return
	}

	contentType := strings.TrimSpace(strings.ToLower(fileHeader.Header.Get("Content-Type")))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(contentType, "image/") {
		// #nosec G706 -- actor id, filename, and content type are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service upload_validation_failed code=invalid_content_type actor=%s filename=%s content_type=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(fileHeader.Filename), utils.SafeLogValue(contentType))
		utils.WriteError(w, http.StatusBadRequest, "invalid_content_type", "uploaded file must be an image")
		return
	}

	input, code, message := parseCreateImageryRequest(r)
	if code != "" {
		// #nosec G706 -- actor id and filename are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service upload_validation_failed code=%s actor=%s filename=%s", code, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(fileHeader.Filename))
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	originalName := filepath.Base(fileHeader.Filename)
	if originalName == "" || originalName == "." || originalName == "/" {
		originalName = "unnamed"
	}

	now := s.now().UTC()
	record := s.store.Create(input, originalName, contentType, "", ctx.ActorUserID, int64(len(data)), now)
	storagePath := filepath.Join(s.config.StoragePath, fmt.Sprintf("%s-%s", record.ID, originalName))

	if err := os.WriteFile(storagePath, data, 0o600); err != nil {
		// Roll back the metadata record so a failed write never leaves a
		// phantom record that lists but cannot be downloaded.
		s.store.Delete(record.ID)
		_ = os.Remove(storagePath)
		// #nosec G706 -- actor id, filename, and path are sanitized with utils.SafeLogValue.
		log.Printf("ERROR imagery-service upload_write_failed actor=%s filename=%s path=%s error=%v", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(fileHeader.Filename), utils.SafeLogValue(storagePath), err)
		utils.WriteError(w, http.StatusInternalServerError, "write_failed", "could not save uploaded file")
		return
	}

	s.store.SetStoragePath(record.ID, storagePath)
	record.StoragePath = storagePath
	// #nosec G706 -- actor id, filename, and content type are sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service upload_completed id=%s actor=%s source=%s filename=%s size=%d content_type=%s", record.ID, utils.SafeLogValue(ctx.ActorUserID), record.Source, utils.SafeLogValue(record.FileName), record.SizeBytes, utils.SafeLogValue(record.ContentType))
	utils.WriteJSON(w, http.StatusCreated, record)
}

func parseCreateImageryRequest(r *http.Request) (models.ImageryUploadInput, string, string) {
	input := models.ImageryUploadInput{
		Source:            utils.NormalizeToken(r.FormValue("source")),
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
	if len(geometryText) > maxGeometryBytes {
		return input, "invalid_geometry", "geometry payload must be 64 KiB or smaller"
	}
	if !json.Valid([]byte(geometryText)) {
		return input, "invalid_geometry", "geometry must be valid JSON"
	}
	var geometry struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal([]byte(geometryText), &geometry); err != nil {
		return input, "invalid_geometry", "geometry could not be parsed"
	}
	if utils.NormalizeToken(geometry.Type) != "polygon" {
		return input, "invalid_geometry", "geometry type must be Polygon"
	}
	if !validPolygonCoordinates(geometry.Coordinates) {
		return input, "invalid_geometry", "geometry coordinates must be a non-empty array of closed linear rings with at least 4 positions each"
	}
	if countPolygonPositions(geometry.Coordinates) > maxGeometryPositions {
		return input, "invalid_geometry", "geometry coordinates must contain 10000 positions or fewer"
	}
	input.Geometry = json.RawMessage(geometryText)

	coverageText := strings.TrimSpace(r.FormValue("coverageAreaKm2"))
	if coverageText == "" {
		return input, "missing_coverage_area", "coverageAreaKm2 is required"
	}
	coverage, err := strconv.ParseFloat(coverageText, 64)
	if err != nil || coverage < 0 || math.IsNaN(coverage) || math.IsInf(coverage, 0) {
		return input, "invalid_coverage_area", "coverageAreaKm2 must be a finite non-negative number"
	}
	input.CoverageAreaKm2 = coverage

	resolutionText := strings.TrimSpace(r.FormValue("resolutionMeters"))
	if resolutionText == "" {
		return input, "missing_resolution", "resolutionMeters is required"
	}
	resolution, err := strconv.ParseFloat(resolutionText, 64)
	if err != nil || resolution < 0 || math.IsNaN(resolution) || math.IsInf(resolution, 0) {
		return input, "invalid_resolution", "resolutionMeters must be a finite non-negative number"
	}
	input.ResolutionMeters = resolution

	if len(input.License) > 200 || utils.UnsafeText(input.License) {
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
		if len(value.field) > 128 || utils.UnsafeText(value.field) {
			return input, "invalid_" + value.name, value.name + " must be 128 safe characters or fewer"
		}
	}

	return input, "", ""
}

// validPolygonCoordinates reports whether rings is a usable GeoJSON Polygon:
// a non-empty array of linear rings, each with at least 4 positions of at
// least 2 numbers, where the first and last positions are identical (closed).
func validPolygonCoordinates(rings [][][]float64) bool {
	if len(rings) == 0 {
		return false
	}
	for _, ring := range rings {
		if len(ring) < 4 {
			return false
		}
		for _, position := range ring {
			if len(position) < 2 {
				return false
			}
		}
		first, last := ring[0], ring[len(ring)-1]
		if len(first) != len(last) {
			return false
		}
		for i := range first {
			if first[i] != last[i] {
				return false
			}
		}
	}
	return true
}

// countPolygonPositions returns the total number of coordinate positions
// across all rings of a polygon.
func countPolygonPositions(rings [][][]float64) int {
	count := 0
	for _, ring := range rings {
		count += len(ring)
	}
	return count
}

func (s *Server) listImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	filter, ok := parseImageryListFilter(w, r)
	if !ok {
		return
	}

	records := s.store.List(filter)
	// #nosec G706 -- actor id and filter values are sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service list count=%d actor=%s source=%s status=%s relatedIncidentId=%s relatedRiskZoneId=%s q=%s", len(records), utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(filter.Source), utils.SafeLogValue(filter.Status), utils.SafeLogValue(filter.RelatedIncidentID), utils.SafeLogValue(filter.RelatedRiskZoneID), utils.SafeLogValue(filter.Query))
	utils.WriteJSON(w, http.StatusOK, models.ImageryListResponse{Imagery: records, GeneratedAt: s.now().UTC()})
}

func parseImageryListFilter(w http.ResponseWriter, r *http.Request) (models.ImageryListFilter, bool) {
	query := r.URL.Query()
	filter := models.ImageryListFilter{
		Source:            utils.NormalizeToken(query.Get("source")),
		Status:            utils.NormalizeToken(query.Get("status")),
		RelatedIncidentID: strings.TrimSpace(query.Get("relatedIncidentId")),
		RelatedRiskZoneID: strings.TrimSpace(query.Get("relatedRiskZoneId")),
		Query:             strings.TrimSpace(query.Get("q")),
	}

	if filter.Source != "" && !allowedSources[filter.Source] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_source", "source must be drone, satellite, or other")
		return filter, false
	}
	if filter.Status != "" && !allowedStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be active or expired")
		return filter, false
	}

	return filter, true
}

func (s *Server) getImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.GetByID(id)
	if !found {
		// #nosec G706 -- path id and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service get_not_found id=%s actor=%s", utils.SafeLogValue(id), utils.SafeLogValue(ctx.ActorUserID))
		utils.WriteError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}
	// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service get id=%s actor=%s source=%s status=%s", record.ID, utils.SafeLogValue(ctx.ActorUserID), record.Source, record.Status)
	utils.WriteJSON(w, http.StatusOK, record)
}

func (s *Server) downloadImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.GetByID(id)
	if !found {
		// #nosec G706 -- path id and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service download_not_found id=%s actor=%s", utils.SafeLogValue(id), utils.SafeLogValue(ctx.ActorUserID))
		utils.WriteError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}

	// Expired imagery is past its retention window: its file has been removed
	// by the retention lifecycle and must not be served.
	if record.Status != "active" {
		// #nosec G706 -- path id and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service download_expired id=%s actor=%s", utils.SafeLogValue(id), utils.SafeLogValue(ctx.ActorUserID))
		utils.WriteError(w, http.StatusGone, "imagery_expired", "imagery record has expired and its file has been removed")
		return
	}

	file, err := os.Open(record.StoragePath)
	if err != nil {
		// #nosec G706 -- actor id and path are sanitized with utils.SafeLogValue.
		log.Printf("ERROR imagery-service download_open_failed id=%s actor=%s path=%s error=%v", record.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(record.StoragePath), err)
		utils.WriteError(w, http.StatusNotFound, "file_not_found", "stored imagery file was not found")
		return
	}
	defer func() { _ = file.Close() }()

	stat, err := file.Stat()
	if err != nil {
		// #nosec G706 -- actor id and path are sanitized with utils.SafeLogValue.
		log.Printf("ERROR imagery-service download_stat_failed id=%s actor=%s path=%s error=%v", record.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(record.StoragePath), err)
		utils.WriteError(w, http.StatusInternalServerError, "file_stat_failed", "could not stat stored imagery file")
		return
	}

	w.Header().Set("Content-Type", record.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", record.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	// #nosec G706 -- actor id and filename are sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service download id=%s actor=%s filename=%s size=%s", record.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(record.FileName), strconv.FormatInt(stat.Size(), 10))
	if _, err := io.Copy(w, file); err != nil {
		// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
		log.Printf("ERROR imagery-service download_copy_failed id=%s actor=%s error=%v", record.ID, utils.SafeLogValue(ctx.ActorUserID), err)
	}
}

func (s *Server) deleteImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.GetByID(id)
	if !found {
		// #nosec G706 -- path id and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service delete_not_found id=%s actor=%s", utils.SafeLogValue(id), utils.SafeLogValue(ctx.ActorUserID))
		utils.WriteError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}

	if err := os.Remove(record.StoragePath); err != nil && !os.IsNotExist(err) {
		// #nosec G706 -- actor id and path are sanitized with utils.SafeLogValue.
		log.Printf("ERROR imagery-service delete_file_failed id=%s actor=%s path=%s error=%v", record.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(record.StoragePath), err)
		utils.WriteError(w, http.StatusInternalServerError, "delete_failed", "could not delete stored imagery file")
		return
	}

	s.store.Delete(id)
	// #nosec G706 -- actor id and filename are sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service delete id=%s actor=%s filename=%s", record.ID, utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(record.FileName))
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) expireImageryHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	record, found := s.store.Expire(id, s.now().UTC())
	if !found {
		// #nosec G706 -- path id and actor id are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service expire_not_found id=%s actor=%s", utils.SafeLogValue(id), utils.SafeLogValue(ctx.ActorUserID))
		utils.WriteError(w, http.StatusNotFound, "not_found", "imagery record was not found")
		return
	}
	// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service expire id=%s actor=%s status=%s", record.ID, utils.SafeLogValue(ctx.ActorUserID), record.Status)
	utils.WriteJSON(w, http.StatusOK, record)
}

func (s *Server) runLifecycleHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r)
	if !ok {
		return
	}

	count := s.store.RunLifecycle(s.now().UTC())
	// #nosec G706 -- actor id is sanitized with utils.SafeLogValue.
	log.Printf("INFO imagery-service lifecycle actor=%s expired_count=%d", utils.SafeLogValue(ctx.ActorUserID), count)
	utils.WriteJSON(w, http.StatusOK, models.ImageryLifecycleResponse{ExpiredCount: count})
}
