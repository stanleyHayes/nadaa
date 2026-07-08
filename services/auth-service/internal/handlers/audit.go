package handlers

import (
	"net/http"

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
