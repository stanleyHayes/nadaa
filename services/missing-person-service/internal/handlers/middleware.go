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

func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx, ok := s.authorityContextFromRequest(r)
	if !ok {
		// #nosec G706 -- request id and path are sanitized with utils.SanitizeLogValue.
		log.Printf("WARN missing-person-service authority_context_missing requestId=%s path=%s", utils.SanitizeLogValue(ctx.RequestID), utils.SanitizeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid bearer token is required for authority actions")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		// #nosec G706 -- actor, role, request id, and path are sanitized with utils.SanitizeLogValue.
		log.Printf("WARN missing-person-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", utils.SanitizeLogValue(ctx.ActorUserID), utils.SanitizeLogValue(ctx.ActorRole), utils.SanitizeLogValue(ctx.RequestID), utils.SanitizeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority actions")
		return models.AuthorityContext{}, false
	}
	if !authorityRoles[ctx.ActorRole] {
		// #nosec G706 -- actor, role, request id, and path are sanitized with utils.SanitizeLogValue.
		log.Printf("WARN missing-person-service authority_forbidden actor=%s role=%s requestId=%s path=%s", utils.SanitizeLogValue(ctx.ActorUserID), utils.SanitizeLogValue(ctx.ActorRole), utils.SanitizeLogValue(ctx.RequestID), utils.SanitizeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// authorityContextFromRequest builds the authority context from a verified
// auth-service bearer token. Legacy X-NADAA-Actor-* headers are honored only
// when mock actors are explicitly enabled (NADAA_AUTH_ALLOW_MOCK_ACTORS=true)
// for local development and smoke tests.
func (s *Server) authorityContextFromRequest(r *http.Request) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		RequestID: strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}
	if claims, ok := verifyBearerToken(r, s.config.AuthTokenSecret, s.now()); ok {
		ctx.ActorUserID = strings.TrimSpace(claims.UserID)
		ctx.ActorAgencyID = strings.TrimSpace(claims.AgencyID)
		ctx.ActorRole = strings.TrimSpace(strings.ToLower(claims.Role))
		ctx.ActorDistrict = strings.TrimSpace(claims.District)
		ctx.MFACompleted = claims.MFA
		return ctx, ctx.ActorUserID != "" && ctx.ActorRole != ""
	}
	if !s.config.AllowMockActors {
		return ctx, false
	}
	ctx.ActorUserID = strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID"))
	ctx.ActorAgencyID = strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID"))
	ctx.ActorRole = strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role")))
	ctx.MFACompleted = strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true"
	return ctx, ctx.ActorUserID != "" && ctx.ActorAgencyID != "" && ctx.ActorRole != ""
}
