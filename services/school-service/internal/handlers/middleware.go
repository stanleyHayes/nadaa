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

	"github.com/stanleyHayes/nadaa/services/school-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/utils"
)

var (
	schoolReadRoles = map[string]bool{
		"system_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"agency_admin":     true,
		"agency_viewer":    true,
	}
	schoolWriteRoles = map[string]bool{
		"system_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"agency_admin":     true,
	}
)

// tokenClaims is the signed payload of a NADAA access token, mirroring
// auth-service's scheme (the packages are not importable across Go modules).
type tokenClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa"`
	ExpiresAt int64  `json:"exp"`
}

// verifyTokenClaims verifies a `nadaa.<payload>.<sig>` token against secret
// and returns its claims. An empty secret never validates, so authority
// requests fail closed when NADAA_AUTH_TOKEN_SECRET is unset.
func verifyTokenClaims(secret []byte, token string, now time.Time) (tokenClaims, bool) {
	if len(secret) == 0 {
		return tokenClaims{}, false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return tokenClaims{}, false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return tokenClaims{}, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, false
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, false
	}
	if claims.ExpiresAt <= now.Unix() {
		return tokenClaims{}, false
	}
	return claims, true
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}

func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx, ok := s.authorityContextFromRequest(r)
	if !ok {
		// #nosec G706 -- request id and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service authority_context_missing requestId=%s path=%s", utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid bearer token is required for authority actions")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		// #nosec G706 -- actor, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority actions")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		// #nosec G706 -- actor, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service authority_forbidden actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return models.AuthorityContext{}, false
	}
	if !isUnscopedRole(ctx.ActorRole) && ctx.ActorDistrict == "" {
		// #nosec G706 -- actor, request id, and path are sanitized with utils.SafeLogValue.
		log.Printf("WARN school-service authority_district_required actor=%s role=%s requestId=%s path=%s", utils.SafeLogValue(ctx.ActorUserID), utils.SafeLogValue(ctx.ActorRole), utils.SafeLogValue(ctx.RequestID), utils.SafeLogValue(r.URL.Path))
		utils.WriteError(w, http.StatusForbidden, "district_scope_required", "actor district is required for district-scoped roles")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// authorityContextFromRequest builds the authority context from a verified
// bearer token. Legacy X-NADAA-Actor-* headers are honored only when the
// service runs with NADAA_AUTH_ALLOW_MOCK_ACTORS=true (local dev and smoke
// tests); otherwise they are ignored entirely.
func (s *Server) authorityContextFromRequest(r *http.Request) (models.AuthorityContext, bool) {
	requestID := strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID"))
	if claims, ok := s.bearerClaims(r); ok {
		ctx := models.AuthorityContext{
			ActorUserID:   strings.TrimSpace(claims.UserID),
			ActorAgencyID: strings.TrimSpace(claims.AgencyID),
			ActorRole:     strings.TrimSpace(strings.ToLower(claims.Role)),
			ActorDistrict: strings.TrimSpace(strings.ToLower(claims.District)),
			MFACompleted:  claims.MFA,
			RequestID:     requestID,
		}
		return ctx, ctx.ActorUserID != "" && ctx.ActorRole != ""
	}
	if !s.config.AllowMockActors {
		return models.AuthorityContext{RequestID: requestID}, false
	}
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		ActorDistrict: strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-District"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     requestID,
	}
	return ctx, ctx.ActorUserID != "" && ctx.ActorAgencyID != "" && ctx.ActorRole != ""
}

func (s *Server) bearerClaims(r *http.Request) (tokenClaims, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(header) <= len("bearer ") || !strings.EqualFold(header[:len("bearer ")], "bearer ") {
		return tokenClaims{}, false
	}
	return verifyTokenClaims([]byte(s.config.TokenSecret), strings.TrimSpace(header[len("bearer "):]), s.now())
}

func isSystemAdmin(ctx models.AuthorityContext) bool {
	return ctx.ActorRole == "system_admin"
}

// isUnscopedRole reports whether a role may access schools across districts.
// Only system_admin and nadmo_officer are unscoped; every other role is
// restricted to its own district.
func isUnscopedRole(role string) bool {
	return role == "system_admin" || role == "nadmo_officer"
}

func scopedDistrict(ctx models.AuthorityContext) string {
	if isUnscopedRole(ctx.ActorRole) {
		return ""
	}
	return ctx.ActorDistrict
}
