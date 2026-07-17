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
	return utils.WithCORS(s.config.AllowedOrigins, s.config.Development, next)
}

func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx := s.authorityContext(r)
	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		// #nosec G706 -- request id and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service authority_context_missing requestId=%s path=%s", utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid authority bearer token is required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		// #nosec G706 -- actor, role, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for imagery operations")
		return models.AuthorityContext{}, false
	}
	if !authorityRoles[ctx.ActorRole] {
		// #nosec G706 -- actor, role, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN imagery-service authority_forbidden actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to perform imagery operations")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// authorityContext builds the actor context from a verified bearer token.
// Legacy X-NADAA-Actor-* headers are honored only when mock actors are
// enabled (NADAA_AUTH_ALLOW_MOCK_ACTORS=true, local dev and smoke tests);
// otherwise they are ignored entirely.
func (s *Server) authorityContext(r *http.Request) models.AuthorityContext {
	ctx := models.AuthorityContext{
		RequestID: strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if s.config.AllowMockActors && hasMockActorHeaders(r) {
		ctx.ActorUserID = strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID"))
		ctx.ActorAgencyID = strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID"))
		ctx.ActorRole = strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role")))
		ctx.MFACompleted = strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true"
		return ctx
	}

	token, ok := bearerToken(r)
	if !ok {
		return ctx
	}
	claims, valid := verifyToken(token, s.config.TokenSecret, s.now())
	if !valid {
		return ctx
	}
	ctx.ActorUserID = claims.UserID
	ctx.ActorRole = strings.TrimSpace(strings.ToLower(claims.Role))
	ctx.ActorAgencyID = claims.AgencyID
	ctx.ActorDistrict = claims.District
	ctx.MFACompleted = claims.MFA
	return ctx
}

func hasMockActorHeaders(r *http.Request) bool {
	return r.Header.Get("X-NADAA-Actor-ID") != "" ||
		r.Header.Get("X-NADAA-Actor-Role") != "" ||
		r.Header.Get("X-NADAA-Agency-ID") != ""
}
