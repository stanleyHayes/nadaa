package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/utils"
)

var authorityRoles = map[string]bool{
	"system_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
	"agency_admin":     true,
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}

func requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx, ok := authorityContextFromRequest(r)
	if !ok {
		log.Printf("WARN missing-person-service authority_context_missing requestId=%s path=%s", ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN missing-person-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority actions")
		return models.AuthorityContext{}, false
	}
	if !authorityRoles[ctx.ActorRole] {
		log.Printf("WARN missing-person-service authority_forbidden actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

func authorityContextFromRequest(r *http.Request) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}
	return ctx, ctx.ActorUserID != "" && ctx.ActorAgencyID != "" && ctx.ActorRole != ""
}
