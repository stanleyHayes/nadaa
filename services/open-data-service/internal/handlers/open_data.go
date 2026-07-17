package handlers

import (
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

	now := s.now().UTC()
	download := s.store.RecordDownload(id, format, ip, now)

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
	utils.WriteJSON(w, http.StatusOK, models.DatasetDownloadResponse{
		Download:    download,
		RateLimit:   status,
		AuditLogged: persisted.ID != "",
	})
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

	var req models.ReviewOpenDataRequest
	if err := utils.DecodeJSON(r, &req); err != nil {
		log.Printf("WARN open-data-service approve_request invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	now := s.now().UTC()
	if req.Approved {
		record.Status = models.OpenDataRequestApproved
	} else {
		record.Status = models.OpenDataRequestRejected
	}
	record.ReviewedAt = &now
	// Attribution comes from the verified admin actor, never the request body.
	record.ReviewedBy = actor.ID
	record.ReviewNote = req.Note

	updated := s.store.UpdateRequest(record)
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
