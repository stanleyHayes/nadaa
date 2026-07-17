package handlers

import (
	"crypto/hmac"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// serviceTokenHeader carries the shared internal token for service-to-service calls.
//nolint:gosec // G101: header name constant, not a credential.
const serviceTokenHeader = "X-NADAA-Service-Token"

// withMiddleware applies CORS, security headers, and the internal-access gate.
func (s *server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, s.requireInternalAccess(next))
}

// requireInternalAccess gates every non-health endpoint: callers must present a
// verified NADAA bearer token or the internal service token. When no internal
// service token is configured (the local development default, warned about at
// startup), the service-token path stays open.
func (s *server) requireInternalAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || s.hasInternalAccess(r) {
			next.ServeHTTP(w, r)
			return
		}
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "a verified bearer token or internal service token is required")
	})
}

func (s *server) hasInternalAccess(r *http.Request) bool {
	if token, ok := utils.BearerToken(r); ok {
		if _, err := utils.VerifyToken(token, s.config.TokenSecret, s.now()); err == nil {
			return true
		}
	}
	if s.config.InternalServiceToken == "" {
		return true
	}
	if hmac.Equal([]byte(r.Header.Get(serviceTokenHeader)), []byte(s.config.InternalServiceToken)) {
		return true
	}
	// Legacy mock actor headers are honored only when explicitly enabled for
	// local development and smoke tests.
	if s.config.AllowMockActors && r.Header.Get("X-NADAA-Actor-ID") != "" {
		return true
	}
	return false
}
