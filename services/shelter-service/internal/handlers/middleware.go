package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

// requireAuthority builds the authority context from a verified bearer token
// (or, when NADAA_AUTH_ALLOW_MOCK_ACTORS=true, the legacy X-NADAA-Actor-*
// headers) and validates MFA and role membership for protected endpoints.
func (s *server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx, ok := s.authorityContextFromRequest(r)
	if !ok {
		requestID := strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID"))
		// #nosec G706 -- request id and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service authority_context_missing requestId=%s path=%s", utils.SafeLogValue(requestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid bearer token is required for authority access")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		// #nosec G706 -- actor, role, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for shelter updates")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		// #nosec G706 -- actor, role, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service authority_forbidden actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to update shelter capacity")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// authorityContextFromRequest builds the authority context from a verified
// auth-service bearer token. The legacy X-NADAA-Actor-* headers are honored
// only when mock actors are explicitly enabled (local dev and smoke tests).
func (s *server) authorityContextFromRequest(r *http.Request) (models.AuthorityContext, bool) {
	requestID := strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID"))

	if s.config != nil && s.config.AllowMockActors {
		ctx := models.AuthorityContext{
			ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
			ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
			ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
			MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
			RequestID:     requestID,
		}
		if ctx.ActorUserID != "" && ctx.ActorAgencyID != "" && ctx.ActorRole != "" {
			return ctx, true
		}
	}

	if s.config == nil || s.config.TokenSecret == "" {
		return models.AuthorityContext{}, false
	}
	token, ok := bearerToken(r)
	if !ok {
		return models.AuthorityContext{}, false
	}
	claims, ok := verifyToken(token, []byte(s.config.TokenSecret), s.now())
	if !ok {
		return models.AuthorityContext{}, false
	}
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(claims.UserID),
		ActorAgencyID: strings.TrimSpace(claims.AgencyID),
		ActorRole:     strings.TrimSpace(strings.ToLower(claims.Role)),
		ActorDistrict: strings.TrimSpace(claims.District),
		MFACompleted:  claims.MFA,
		RequestID:     requestID,
	}
	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		return models.AuthorityContext{}, false
	}
	return ctx, true
}
