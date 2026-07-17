package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

func (s *Server) createAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.DraftRoles)
	if !ok {
		return
	}

	var request models.CreateAlertRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized := utils.NormalizeAlertRequest(request)
	if code, message := utils.ValidateAlertRequest(normalized); code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	alert := s.store.CreateAlert(normalized, ctx, s.now().UTC())
	utils.WriteJSON(w, http.StatusCreated, alert)
}

func (s *Server) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	status := utils.NormalizeQueryValue(r.URL.Query().Get("status"))
	currentOnly := utils.NormalizeQueryValue(r.URL.Query().Get("current")) == "true"
	targetType := utils.NormalizeQueryValue(r.URL.Query().Get("targetType"))
	targetID := utils.NormalizeQueryValue(r.URL.Query().Get("targetId"))
	if status != "" && !utils.ValidAlertStatus(status) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be draft, submitted, approved, rejected, published, expired, or cancelled")
		return
	}
	if targetType != "" && !utils.AllowedTargetTypes[targetType] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_target_type", "targetType must be national, region, district, radius, community, or custom")
		return
	}

	_, isAuthority := s.authorityContext(r)
	publicOnly := !isAuthority
	utils.WriteJSON(w, http.StatusOK, models.AlertListResponse{Alerts: s.store.ListAlerts(status, currentOnly, publicOnly, targetType, targetID, s.now().UTC())})
}

func (s *Server) updateAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.DraftRoles)
	if !ok {
		return
	}

	var request models.CreateAlertRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized := utils.NormalizeAlertRequest(request)
	if code, message := utils.ValidateAlertRequest(normalized); code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	alert, code, message := s.store.UpdateAlert(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alert)
}

func (s *Server) submitAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.DraftRoles)
	if !ok {
		return
	}

	alert, code, message := s.store.TransitionAlert(r.PathValue("id"), "submitted", ctx, models.WorkflowRequest{}, s.now().UTC())
	if code != "" {
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alert)
}

func (s *Server) approveAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ApprovalRoles)
	if !ok {
		return
	}

	var request models.WorkflowRequest
	if err := utils.OptionalDecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	alert, code, message := s.store.TransitionAlert(r.PathValue("id"), "approved", ctx, request, s.now().UTC())
	if code != "" {
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alert)
}

func (s *Server) rejectAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ApprovalRoles)
	if !ok {
		return
	}

	var request models.WorkflowRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if strings.TrimSpace(request.Reason) == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_reason", "reason is required when rejecting an alert")
		return
	}

	alert, code, message := s.store.TransitionAlert(r.PathValue("id"), "rejected", ctx, request, s.now().UTC())
	if code != "" {
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alert)
}

func (s *Server) emergencyOverrideHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.OverrideRoles)
	if !ok {
		return
	}

	var request models.WorkflowRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if strings.TrimSpace(request.Reason) == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_reason", "reason is required for emergency override")
		return
	}

	alert, code, message := s.store.TransitionAlert(r.PathValue("id"), "emergency_override", ctx, request, s.now().UTC())
	if code != "" {
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, alert)
}

func (s *Server) listAuditHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAuthority(w, r, utils.ApprovalRoles); !ok {
		return
	}

	limit := 50
	if raw := utils.NormalizeQueryValue(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}

	utils.WriteJSON(w, http.StatusOK, models.AuditListResponse{Logs: s.store.ListAudit(limit)})
}
