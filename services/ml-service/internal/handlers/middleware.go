package handlers

import (
	"crypto/hmac"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// serviceTokenHeader carries the shared internal token for service-to-service calls.
//
//nolint:gosec // G101: header name constant, not a credential.
const serviceTokenHeader = "X-NADAA-Service-Token"

// internalAccessRoles mirrors the incident-service authority role allowlists:
// only verified agency/dispatcher role claims may reach ML decision-support
// endpoints. Citizen tokens verify but never authorize.
var internalAccessRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

// withMiddleware applies CORS, security headers, and the internal-access gate.
func (s *server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, s.requireInternalAccess(next))
}

// requireInternalAccess gates every non-health endpoint: callers must present
// a verified NADAA bearer token carrying an agency/dispatcher role or the
// internal service token; mock actor headers work only when explicitly
// enabled for local development. The service-token path fails closed when
// NADAA_INTERNAL_SERVICE_TOKEN is unset.
func (s *server) requireInternalAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		switch s.checkInternalAccess(r) {
		case http.StatusOK:
			next.ServeHTTP(w, r)
		case http.StatusForbidden:
			utils.WriteError(w, http.StatusForbidden, "forbidden", "the verified token does not carry an agency or dispatcher role")
		default:
			utils.WriteError(w, http.StatusUnauthorized, "unauthorized", "a verified agency bearer token or internal service token is required")
		}
	})
}

// checkInternalAccess returns http.StatusOK when the request is authorized,
// http.StatusForbidden when it is authenticated but under-privileged, and
// http.StatusUnauthorized otherwise.
func (s *server) checkInternalAccess(r *http.Request) int {
	// Service-to-service calls authenticate with the shared internal token.
	// When no token is configured the header is ignored entirely: the gate
	// fails closed instead of the old development-default open behavior.
	if s.config.InternalServiceToken != "" &&
		hmac.Equal([]byte(r.Header.Get(serviceTokenHeader)), []byte(s.config.InternalServiceToken)) {
		return http.StatusOK
	}
	if token, ok := utils.BearerToken(r); ok {
		if claims, err := utils.VerifyToken(token, s.config.TokenSecret, s.now()); err == nil {
			if internalAccessRoles[normalizeRole(claims.Role)] {
				return http.StatusOK
			}
			// A verified token without an authority role (e.g. a citizen
			// token) is authenticated but not authorized.
			return http.StatusForbidden
		}
	}
	// Legacy mock actor headers are honored only when explicitly enabled for
	// local development and smoke tests.
	if s.config.AllowMockActors && r.Header.Get("X-NADAA-Actor-ID") != "" {
		return http.StatusOK
	}
	return http.StatusUnauthorized
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}
