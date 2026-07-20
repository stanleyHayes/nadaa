package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

var authorityRoles = map[string]bool{
	"system_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
	"ngo":              true,
	"agency_admin":     true,
	"agency_viewer":    true,
}

// withMiddleware applies server-wide middleware to the handler.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, s.config.IsDevelopment(), next)
}

// requireAuthority validates that the request includes a valid authority context.
func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx, ok := s.authorityContextFromRequest(r)
	if !ok {
		log.Printf("WARN donation-service authority_context_missing requestId=%s path=%s", utils.LogSafe(ctx.RequestID), utils.LogSafe(r.URL.Path)) // #nosec G706 -- values sanitized by utils.LogSafe (strips \n and \r)
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid bearer token is required for authority actions")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN donation-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", utils.LogSafe(ctx.ActorUserID), utils.LogSafe(ctx.ActorRole), utils.LogSafe(ctx.RequestID), utils.LogSafe(r.URL.Path)) // #nosec G706 -- values sanitized by utils.LogSafe (strips \n and \r)
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority actions")
		return models.AuthorityContext{}, false
	}
	if !authorityRoles[ctx.ActorRole] {
		log.Printf("WARN donation-service authority_forbidden actor=%s role=%s requestId=%s path=%s", utils.LogSafe(ctx.ActorUserID), utils.LogSafe(ctx.ActorRole), utils.LogSafe(ctx.RequestID), utils.LogSafe(r.URL.Path)) // #nosec G706 -- values sanitized by utils.LogSafe (strips \n and \r)
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// verifiedAuthorityCaller reports whether the request carries a complete
// authority context (verified bearer token, or mock actor headers when
// explicitly allowed) with MFA completed and an authority role. Unlike
// requireAuthority it writes no response, so public endpoints can branch on it.
func (s *Server) verifiedAuthorityCaller(r *http.Request) bool {
	ctx, ok := s.authorityContextFromRequest(r)
	return ok && ctx.MFACompleted && authorityRoles[ctx.ActorRole]
}

// authorityContextFromRequest builds the authority context from a verified
// bearer token's claims. Legacy X-NADAA-Actor-* headers are honored only when
// NADAA_AUTH_ALLOW_MOCK_ACTORS=true (local dev and smoke tests); otherwise they
// are ignored entirely.
func (s *Server) authorityContextFromRequest(r *http.Request) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		RequestID: strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if claims, ok := verifyAuthToken(s.config.AuthTokenSecret, bearerToken(r), s.now()); ok {
		ctx.ActorUserID = strings.TrimSpace(claims.Sub)
		ctx.ActorAgencyID = strings.TrimSpace(claims.AgencyID)
		ctx.ActorRole = strings.TrimSpace(strings.ToLower(claims.Role))
		ctx.ActorDistrict = strings.TrimSpace(claims.District)
		ctx.MFACompleted = claims.MFA
		return ctx, ctx.ActorUserID != "" && ctx.ActorAgencyID != "" && ctx.ActorRole != ""
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
