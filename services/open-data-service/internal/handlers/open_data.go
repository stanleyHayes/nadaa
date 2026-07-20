package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/utils"
)

func (s *Server) listDatasetsHandler(w http.ResponseWriter, r *http.Request) {
	category := utils.NormalizeString(r.URL.Query().Get("category"))
	statusParam := utils.NormalizeString(r.URL.Query().Get("privacyReviewStatus"))

	var status models.PrivacyReviewStatus
	switch statusParam {
	case "approved":
		status = models.PrivacyReviewApproved
	case "pending_review":
		status = models.PrivacyReviewPending
	case "rejected":
		status = models.PrivacyReviewRejected
	}

	// Anonymous reads are approved-only; only a verified admin may list or
	// filter across every review status.
	if !s.isAdmin(r) {
		status = models.PrivacyReviewApproved
	}

	datasets := s.store.ListDatasets(category, status)
	for i := range datasets {
		scrubNonApproved(&datasets[i])
	}
	utils.WriteJSON(w, http.StatusOK, models.DatasetListResponse{
		Datasets:    datasets,
		GeneratedAt: s.now().UTC(),
	})
}

func (s *Server) getDatasetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dataset, ok := s.store.GetDataset(id)
	// Non-approved datasets are invisible to anonymous callers, same as missing.
	if !ok || (dataset.PrivacyReviewStatus != models.PrivacyReviewApproved && !s.isAdmin(r)) {
		utils.WriteError(w, http.StatusNotFound, "dataset_not_found", "dataset not found")
		return
	}

	scrubNonApproved(&dataset)
	utils.WriteJSON(w, http.StatusOK, models.DatasetDetailResponse{Dataset: dataset})
}

// scrubNonApproved strips sample rows and column layout from datasets that have
// not passed privacy review; they leak the same class of data downloads gate.
func scrubNonApproved(dataset *models.Dataset) {
	if dataset.PrivacyReviewStatus == models.PrivacyReviewApproved {
		return
	}
	dataset.SampleRows = nil
	dataset.Columns = nil
}

func (s *Server) downloadDatasetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	format := utils.NormalizeString(r.URL.Query().Get("format"))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "json" {
		utils.WriteError(w, http.StatusBadRequest, "unsupported_format", "format must be csv or json")
		return
	}

	dataset, ok := s.store.GetDataset(id)
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "dataset_not_found", "dataset not found")
		return
	}

	if dataset.PrivacyReviewStatus != models.PrivacyReviewApproved {
		utils.WriteError(w, http.StatusForbidden, "dataset_not_approved", "dataset is not approved for download")
		return
	}

	ip := s.clientIP(r)
	status, allowed := s.checkRateLimit(ip)
	if !allowed {
		// #nosec G706 -- ip and dataset id are sanitized with utils.SafeLog.
		log.Printf("WARN open-data-service rate_limit_exceeded ip=%s dataset=%s", utils.SafeLog(ip), utils.SafeLog(id))
		utils.WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded; try again later")
		return
	}

	payload, contentType := serializeDatasetRows(dataset, format)

	now := s.now().UTC()
	// The recorded artifact size is the exact byte size served below.
	download := s.store.RecordDownload(id, format, ip, int64(len(payload)), now)

	// Local persistence is the audit trail of record; forwarding to the audit
	// log service is best-effort on top and never claimed here.
	persisted := s.store.RecordAuditEvent(models.AuditEvent{
		Action:     "dataset_download",
		TargetType: "dataset",
		TargetID:   id,
		ActorRole:  "public",
		IPAddress:  ip,
		UserAgent:  r.UserAgent(),
		Metadata: map[string]string{
			"downloadId": download.ID,
			"format":     format,
			"size":       strconv.FormatInt(download.Size, 10),
		},
		CreatedAt: now,
	})
	s.sendAuditEvent(r, persisted)

	// #nosec G706 -- dataset id, format, and ip are sanitized with utils.SafeLog.
	log.Printf("INFO open-data-service dataset_download dataset=%s format=%s download=%s ip=%s remaining=%d",
		utils.SafeLog(id), utils.SafeLog(format), download.ID, utils.SafeLog(ip), status.Remaining)

	// The download record, rate limit state, and local-audit outcome travel in
	// headers because the body carries the dataset bytes themselves.
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", id+"."+format))
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(status.Limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(status.Remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(status.ResetAt.Unix(), 10))
	w.Header().Set("X-NADAA-Download-Id", download.ID)
	w.Header().Set("X-NADAA-Audit-Logged", strconv.FormatBool(persisted.ID != ""))
	w.WriteHeader(http.StatusOK)
	// #nosec G705 -- dataset rows are served as a file attachment
	// (Content-Disposition) with a non-HTML content type, never as markup.
	if _, err := w.Write(payload); err != nil {
		// #nosec G706 -- dataset id is sanitized with utils.SafeLog; the error is a client-connection failure, not request data.
		log.Printf("ERROR open-data-service write_download_failed dataset=%s error=%v", utils.SafeLog(id), err)
	}
}

// serializeDatasetRows renders the dataset's actual rows in the requested
// format and returns the bytes served to the caller together with the
// matching Content-Type.
func serializeDatasetRows(dataset models.Dataset, format string) ([]byte, string) {
	if format == "json" {
		rows := dataset.SampleRows
		if rows == nil {
			rows = []map[string]any{}
		}
		payload, err := json.Marshal(rows)
		if err != nil {
			// map[string]any rows seeded from literals cannot fail to marshal;
			// fall back to an empty array rather than fabricating content.
			return []byte("[]"), "application/json"
		}
		return payload, "application/json"
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	header := make([]string, 0, len(dataset.Columns))
	for _, column := range dataset.Columns {
		header = append(header, column.Name)
	}
	_ = writer.Write(header)
	for _, row := range dataset.SampleRows {
		record := make([]string, 0, len(dataset.Columns))
		for _, column := range dataset.Columns {
			record = append(record, csvCell(row[column.Name]))
		}
		_ = writer.Write(record)
	}
	writer.Flush()
	return buf.Bytes(), "text/csv"
}

// csvCell renders a row value as a CSV cell string.
func csvCell(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func (s *Server) createRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Unauthenticated request creation is rate-limited with the same token
	// bucket as downloads so it cannot grow the store without bound.
	ip := s.clientIP(r)
	if _, allowed := s.checkRateLimit(ip); !allowed {
		// #nosec G706 -- ip and path are sanitized with utils.SafeLog.
		log.Printf("WARN open-data-service rate_limit_exceeded ip=%s path=%s", utils.SafeLog(ip), utils.SafeLog(r.URL.Path))
		utils.WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded; try again later")
		return
	}

	var req models.CreateOpenDataRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		log.Printf("WARN open-data-service create_request invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	if _, ok := s.store.GetDataset(req.DatasetID); !ok {
		utils.WriteError(w, http.StatusNotFound, "dataset_not_found", "dataset not found")
		return
	}

	if err := validateRequesterInfo(req.RequesterInfo); err != "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_requester_info", err)
		return
	}

	if req.Purpose == "" || len(req.Purpose) < 10 {
		utils.WriteError(w, http.StatusBadRequest, "invalid_purpose", "purpose must be at least 10 characters")
		return
	}

	req.RequesterInfo.Email = utils.SanitizeEmail(req.RequesterInfo.Email)
	now := s.now().UTC()
	record := models.OpenDataRequest{
		// ID is assigned by the store under its lock to guarantee uniqueness.
		DatasetID:     req.DatasetID,
		RequesterInfo: req.RequesterInfo,
		Purpose:       req.Purpose,
		Status:        models.OpenDataRequestPending,
		CreatedAt:     now,
	}

	created := s.store.CreateRequest(record)
	log.Printf("INFO open-data-service access_request_created request=%s dataset=%s requester=%s",
		created.ID, utils.SafeLog(created.DatasetID), utils.SafeLog(created.RequesterInfo.Email))
	utils.WriteJSON(w, http.StatusCreated, models.OpenDataRequestResponse{Request: created})
}

func (s *Server) listRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.OpenDataRequestListResponse{
		Requests:    s.store.ListRequests(),
		GeneratedAt: s.now().UTC(),
	})
}

func (s *Server) approveRequestHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	record, ok := s.store.GetRequest(id)
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "request_not_found", "request not found")
		return
	}

	// A decided request is final: re-review would silently overwrite the
	// original decision and its audit trail.
	if record.Status != models.OpenDataRequestPending {
		utils.WriteError(w, http.StatusConflict, "request_already_reviewed", "request has already been reviewed and cannot be reviewed again")
		return
	}

	var req models.ReviewOpenDataRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		log.Printf("WARN open-data-service approve_request invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	now := s.now().UTC()
	decision := "rejected"
	if req.Approved {
		record.Status = models.OpenDataRequestApproved
		decision = "approved"
	} else {
		record.Status = models.OpenDataRequestRejected
	}
	record.ReviewedAt = &now
	// Attribution comes from the verified admin actor, never the request body.
	record.ReviewedBy = actor.ID
	record.ReviewNote = req.Note

	updated := s.store.UpdateRequest(record)

	// Every review decision is audit-logged with the verified admin actor.
	// Local persistence is the record of truth; forwarding is best-effort.
	persisted := s.store.RecordAuditEvent(models.AuditEvent{
		Action:     "access_request_review",
		TargetType: "access_request",
		TargetID:   updated.ID,
		ActorRole:  actor.Role,
		IPAddress:  s.clientIP(r),
		UserAgent:  r.UserAgent(),
		Metadata: map[string]string{
			"datasetId": updated.DatasetID,
			"decision":  decision,
			"reviewer":  actor.ID,
		},
		CreatedAt: now,
	})
	s.sendAuditEvent(r, persisted)

	log.Printf("INFO open-data-service access_request_reviewed request=%s approved=%t reviewer=%s", updated.ID, req.Approved, utils.SafeLog(actor.ID))
	utils.WriteJSON(w, http.StatusOK, models.OpenDataRequestResponse{Request: updated})
}

func (s *Server) listAuditEventsHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.AuditEventListResponse{
		Events:      s.store.ListAuditEvents(),
		GeneratedAt: s.now().UTC(),
	})
}

func validateRequesterInfo(info models.RequesterInfo) string {
	if info.Name == "" || len(info.Name) < 2 {
		return "name must be at least 2 characters"
	}
	if info.Email == "" || !utils.ValidEmail(info.Email) {
		return "valid email is required"
	}
	if info.UseCase == "" {
		return "useCase is required"
	}
	if utils.UnsafeText(info.Name) || utils.UnsafeText(info.UseCase) || utils.UnsafeText(info.Organization) {
		return "input contains unsafe text"
	}
	return ""
}
