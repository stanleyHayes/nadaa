package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/utils"
)

var authorityRoles = map[string]bool{
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

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}

func requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx := authorityContextFromRequest(r)
	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		log.Printf("WARN imagery-service authority_context_missing requestId=%s path=%s", ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN imagery-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for imagery operations")
		return models.AuthorityContext{}, false
	}
	if !authorityRoles[ctx.ActorRole] {
		log.Printf("WARN imagery-service authority_forbidden actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to perform imagery operations")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

func authorityContextFromRequest(r *http.Request) models.AuthorityContext {
	return models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}
}
