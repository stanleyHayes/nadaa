package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"
)

// withMiddleware applies server-wide middleware to the handler.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}

// requireAuthority validates NADAA authority headers and MFA for protected endpoints.
func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     utils.NormalizeString(r.Header.Get("X-NADAA-Actor-Role")),
		MFACompleted:  utils.NormalizeString(r.Header.Get("X-NADAA-MFA-Completed")) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		log.Printf("WARN damage-claim-service authority_context_missing requestId=%s path=%s", ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN damage-claim-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority operations")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		log.Printf("WARN damage-claim-service authority_forbidden actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}
