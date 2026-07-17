package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"
)

// withMiddleware applies server-wide middleware to the handler.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}

// verifyAuthToken validates a nadaa.<payload>.<sig> bearer token signed by
// auth-service (HMAC-SHA256 over the base64url claims payload) and returns the
// verified claims. The scheme mirrors auth-service and is duplicated here
// because it is not importable across Go modules.
func verifyAuthToken(token, secret string, now time.Time) (models.TokenClaims, bool) {
	if secret == "" {
		return models.TokenClaims{}, false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return models.TokenClaims{}, false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return models.TokenClaims{}, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return models.TokenClaims{}, false
	}
	var claims models.TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return models.TokenClaims{}, false
	}
	if claims.ExpiresAt <= now.Unix() {
		return models.TokenClaims{}, false
	}
	return claims, true
}

// bearerToken extracts the token from an Authorization: Bearer header.
func bearerToken(r *http.Request) string {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(header) > len("Bearer ") && strings.EqualFold(header[:len("Bearer ")], "Bearer ") {
		return strings.TrimSpace(header[len("Bearer "):])
	}
	return ""
}

// requireAuthority validates the caller's authority identity and MFA for
// protected endpoints. The actor context is built from a verified NADAA bearer
// token; legacy X-NADAA-Actor-* headers are honored only when the service runs
// with NADAA_AUTH_ALLOW_MOCK_ACTORS=true (local development and smoke tests).
func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		RequestID: strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if claims, ok := verifyAuthToken(bearerToken(r), s.config.AuthTokenSecret, s.now()); ok {
		ctx.ActorUserID = strings.TrimSpace(claims.UserID)
		ctx.ActorRole = utils.NormalizeString(claims.Role)
		ctx.ActorAgencyID = strings.TrimSpace(claims.AgencyID)
		ctx.ActorDistrict = strings.TrimSpace(claims.District)
		ctx.MFACompleted = claims.MFA
	} else if s.config.AllowMockActors {
		ctx.ActorUserID = strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID"))
		ctx.ActorAgencyID = strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID"))
		ctx.ActorRole = utils.NormalizeString(r.Header.Get("X-NADAA-Actor-Role"))
		ctx.MFACompleted = utils.NormalizeString(r.Header.Get("X-NADAA-MFA-Completed")) == "true"
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		// #nosec G706 -- request id and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service authority_context_missing requestId=%s path=%s", utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid NADAA bearer token with authority context is required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		// #nosec G706 -- actor, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority operations")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		// #nosec G706 -- actor, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service authority_forbidden actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}
