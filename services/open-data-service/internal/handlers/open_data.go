package handlers

import (
	"log"
	"net/http"
	"time"

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

	datasets := s.store.ListDatasets(category, status)
	utils.WriteJSON(w, http.StatusOK, models.DatasetListResponse{
		Datasets:    datasets,
		GeneratedAt: s.now().UTC(),
	})
}

func (s *Server) getDatasetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	dataset, ok := s.store.GetDataset(id)
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "dataset_not_found", "dataset not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.DatasetDetailResponse{Dataset: dataset})
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
		log.Printf("WARN open-data-service rate_limit_exceeded ip=%s dataset=%s", ip, id)
		utils.WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded; try again later")
		return
	}

	now := s.now().UTC()
	download := s.store.RecordDownload(id, format, ip, now)

	s.sendAuditEvent(r, models.AuditEvent{
		Action:     "dataset_download",
		TargetType: "dataset",
		TargetID:   id,
		ActorRole:  "public",
		IPAddress:  ip,
		UserAgent:  r.UserAgent(),
		Metadata: map[string]string{
			"downloadId": download.ID,
			"format":     format,
			"size":       "",
		},
		CreatedAt: now,
	})

	log.Printf("INFO open-data-service dataset_download dataset=%s format=%s download=%s ip=%s remaining=%d", id, format, download.ID, ip, status.Remaining)
	utils.WriteJSON(w, http.StatusOK, models.DatasetDownloadResponse{
		Download:    download,
		RateLimit:   status,
		AuditLogged: true,
	})
}

func (s *Server) createRequestHandler(w http.ResponseWriter, r *http.Request) {
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
		ID:            generateRequestID(now),
		DatasetID:     req.DatasetID,
		RequesterInfo: req.RequesterInfo,
		Purpose:       req.Purpose,
		Status:        models.OpenDataRequestPending,
		CreatedAt:     now,
	}

	created := s.store.CreateRequest(record)
	log.Printf("INFO open-data-service access_request_created request=%s dataset=%s requester=%s", created.ID, created.DatasetID, created.RequesterInfo.Email)
	utils.WriteJSON(w, http.StatusCreated, models.OpenDataRequestResponse{Request: created})
}

func (s *Server) listRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.OpenDataRequestListResponse{
		Requests:    s.store.ListRequests(),
		GeneratedAt: s.now().UTC(),
	})
}

func (s *Server) approveRequestHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
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
	record.ReviewedBy = req.Reviewer
	record.ReviewNote = req.Note

	updated := s.store.UpdateRequest(record)
	log.Printf("INFO open-data-service access_request_reviewed request=%s approved=%t reviewer=%s", updated.ID, req.Approved, req.Reviewer)
	utils.WriteJSON(w, http.StatusOK, models.OpenDataRequestResponse{Request: updated})
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

func generateRequestID(now time.Time) string {
	return "odr_" + now.Format("20060102150405") + "_" + "0001"
}
