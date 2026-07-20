package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

func (s *Server) listAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAgencyRole(w, r, models.RoleSystemAdmin)
	if !ok {
		return
	}

	limit := utils.ParseAuditLimit(r.URL.Query().Get("limit"))
	logs := s.store.ListAuditLogs(limit)
	utils.WriteJSON(w, http.StatusOK, models.AuditLogListResponse{Logs: logs})

	s.recordAudit(r, utils.AuditActorFromAgency(actor), "audit.logs.viewed", models.AuditTarget{Type: "audit_logs"}, nil, map[string]any{
		"limit": limit,
		"count": len(logs),
	})
}

// ingestAuditLogHandler accepts audit events forwarded by other services. It
// is gated on the shared service-to-service token only — never on end-user
// credentials — and stays closed when NADAA_INTERNAL_SERVICE_TOKEN is unset.
func (s *Server) ingestAuditLogHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.InternalServiceToken == "" || !utils.SecureCompare(strings.TrimSpace(r.Header.Get(serviceTokenHeader)), s.config.InternalServiceToken) {
		utils.WriteError(w, http.StatusUnauthorized, "invalid_service_token", "a valid service token is required")
		return
	}

	var request models.IngestAuditLogRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.EventType = strings.TrimSpace(request.EventType)
	if request.EventType == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_event", "eventType is required")
		return
	}

	var after map[string]any
	if request.Summary != "" || request.Metadata != nil {
		after = map[string]any{}
		if request.Summary != "" {
			after["summary"] = request.Summary
		}
		if request.Metadata != nil {
			after["metadata"] = request.Metadata
		}
	}

	record := s.recordAudit(r, models.AuditActor{
		UserID: strings.TrimSpace(request.ActorID),
		Role:   strings.TrimSpace(request.ActorRole),
	}, request.EventType, models.AuditTarget{
		Type: strings.TrimSpace(request.ResourceType),
		ID:   strings.TrimSpace(request.ResourceID),
	}, nil, after)
	utils.WriteJSON(w, http.StatusCreated, models.IngestAuditLogResponse{ID: record.ID})
}

func (s *Server) recordAudit(r *http.Request, actor models.AuditActor, action string, target models.AuditTarget, before map[string]any, after map[string]any) models.AuditLogRecord {
	context := utils.AuditContextFromRequest(r)
	record := models.AuditLogRecord{
		ID:            utils.NewID("aud"),
		ActorUserID:   actor.UserID,
		ActorAgencyID: actor.AgencyID,
		ActorRole:     actor.Role,
		Action:        action,
		TargetType:    target.Type,
		TargetID:      target.ID,
		RequestID:     context.RequestID,
		IPAddress:     context.IPAddress,
		UserAgent:     context.UserAgent,
		Before:        before,
		After:         after,
		CreatedAt:     s.now().UTC(),
	}
	return s.store.AppendAuditLog(record)
}
