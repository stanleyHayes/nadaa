package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/utils"
)

// openDataAdminRoles are the authority roles allowed to review access requests.
var openDataAdminRoles = map[string]bool{
	"system_admin":  true,
	"agency_admin":  true,
	"nadmo_officer": true,
}

// bucket tracks token-bucket rate limit state.
type bucket struct {
	tokens  int
	resetAt time.Time
}

// Server holds the HTTP handler dependencies.
type Server struct {
	store      store.Store
	now        func() time.Time
	config     *config.Config
	httpClient *http.Client
	rateLimits map[string]*bucket
	rateMu     sync.Mutex
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, now func() time.Time, cfg *config.Config) *Server {
	return &Server{
		store:      s,
		now:        now,
		config:     cfg,
		httpClient: http.DefaultClient,
		rateLimits: map[string]*bucket{},
	}
}

func (s *Server) clientIP(r *http.Request) string {
	// Proxy-supplied headers are trivially forgeable, so they are honored only
	// when the operator declares the service sits behind a trusted proxy.
	if s.config.TrustProxyHeaders {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
				return first
			}
		}
		if realIP := strings.TrimSpace(r.Header.Get("X-Real-Ip")); realIP != "" {
			return realIP
		}
	}
	// RemoteAddr carries the ephemeral source port; rate limiting keys on the host.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func (s *Server) checkRateLimit(ip string) (models.RateLimitStatus, bool) {
	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	now := s.now().UTC()
	window := time.Duration(s.config.RateLimitWindowSeconds) * time.Second

	// Evict expired buckets so spoofed IPs cannot grow the map without bound.
	for key, existing := range s.rateLimits {
		if now.After(existing.resetAt) {
			delete(s.rateLimits, key)
		}
	}

	b, ok := s.rateLimits[ip]
	if !ok || now.After(b.resetAt) {
		b = &bucket{
			tokens:  s.config.RateLimitRequests,
			resetAt: now.Add(window),
		}
		s.rateLimits[ip] = b
	}

	status := models.RateLimitStatus{
		Limit:   s.config.RateLimitRequests,
		ResetAt: b.resetAt,
	}
	if b.tokens > 0 {
		b.tokens--
		status.Remaining = b.tokens
		return status, true
	}
	status.Remaining = 0
	return status, false
}

// sendAuditEvent forwards an already-persisted audit event to the audit log
// service on a best-effort basis. Local persistence is the source of truth for
// auditLogged; a forwarding failure is logged, never reported as success. The
// forward authenticates with the shared internal service token
// (X-NADAA-Service-Token); the caller's credentials are never forwarded.
func (s *Server) sendAuditEvent(r *http.Request, event models.AuditEvent) {
	if s.config.AuditLogServiceURL == "" {
		return
	}
	// Detach from the request's cancellation so the fire-and-forget attempt can
	// outlive the handler, while still inheriting request-scoped values.
	parent := context.WithoutCancel(r.Context())
	// Production should queue reliably rather than fire-and-forget.
	go func(evt models.AuditEvent) {
		body, err := json.Marshal(evt)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(parent, 2*time.Second)
		defer cancel()
		endpoint := s.config.AuditLogServiceURL + "/api/v1/audit/logs"
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if s.config.InternalServiceToken != "" {
			req.Header.Set("X-NADAA-Service-Token", s.config.InternalServiceToken)
		}
		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("WARN open-data-service audit_forward_failed event=%s target=%s error=%v", utils.SafeLog(evt.ID), utils.SafeLog(evt.TargetID), err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			log.Printf("WARN open-data-service audit_forward_rejected event=%s target=%s status=%d", utils.SafeLog(evt.ID), utils.SafeLog(evt.TargetID), resp.StatusCode)
		}
	}(event)
}

// adminAuthResult carries the outcome of an admin authentication attempt: the
// resolved actor plus the HTTP status and error details to report on failure.
// A 200 status means authorized.
type adminAuthResult struct {
	actor   adminActor
	status  int
	code    string
	message string
}

// authenticateAdmin resolves the caller's authority identity and enforces the
// open-data admin policy (verified identity, MFA, allowed role). A verified
// bearer token is the primary path; the legacy X-NADAA-Actor-* headers are
// honored only when NADAA_AUTH_ALLOW_MOCK_ACTORS=true (local dev and smoke
// tests). It writes no response itself.
func (s *Server) authenticateAdmin(r *http.Request) adminAuthResult {
	if token, ok := bearerToken(r); ok {
		claims, err := verifyAuthToken(token, s.config.AuthTokenSecret, s.now())
		if err != nil {
			return adminAuthResult{status: http.StatusUnauthorized, code: "invalid_token", message: "bearer token is invalid or expired"}
		}
		if claims.UserType != "agency" {
			return adminAuthResult{status: http.StatusForbidden, code: "authority_user_required", message: "authority user access is required"}
		}
		actor := adminActor{
			ID:       strings.TrimSpace(claims.Subject),
			Role:     strings.TrimSpace(strings.ToLower(claims.Role)),
			AgencyID: strings.TrimSpace(claims.AgencyID),
			District: strings.TrimSpace(claims.District),
		}
		return s.authorizeAdmin(actor, claims.MFA)
	}

	if s.config.AllowMockActors {
		actor := adminActor{
			ID:       strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
			Role:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
			AgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		}
		mfaCompleted := strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true"
		if actor.ID == "" && actor.AgencyID == "" && actor.Role == "" {
			return adminAuthResult{status: http.StatusUnauthorized, code: "missing_authority_context", message: "Bearer token is required for open-data admin actions"}
		}
		return s.authorizeAdmin(actor, mfaCompleted)
	}

	return adminAuthResult{status: http.StatusUnauthorized, code: "missing_authority_context", message: "Bearer token is required for open-data admin actions"}
}

// authorizeAdmin checks the resolved actor against the admin policy.
func (s *Server) authorizeAdmin(actor adminActor, mfaCompleted bool) adminAuthResult {
	if actor.ID == "" || actor.AgencyID == "" || actor.Role == "" {
		return adminAuthResult{status: http.StatusUnauthorized, code: "missing_authority_context", message: "authority actor id, role, and agency id are required"}
	}
	if !mfaCompleted {
		return adminAuthResult{status: http.StatusForbidden, code: "mfa_required", message: "MFA must be completed for open-data admin actions"}
	}
	if !openDataAdminRoles[actor.Role] {
		return adminAuthResult{status: http.StatusForbidden, code: "forbidden", message: "actor role is not allowed to review open-data access requests"}
	}
	return adminAuthResult{actor: actor, status: http.StatusOK}
}

// isAdmin reports whether the caller is a fully authorized admin, without
// writing any response. Public endpoints use it to widen visibility for
// verified admins while staying public for everyone else.
func (s *Server) isAdmin(r *http.Request) bool {
	return s.authenticateAdmin(r).status == http.StatusOK
}

// requireAdmin enforces the full authority context for open-data admin actions.
// It writes the 401/403 response and logs the actor for audit. It returns the
// verified admin actor and true only when the caller is authorized.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) (adminActor, bool) {
	requestID := utils.SafeLog(strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")))

	result := s.authenticateAdmin(r)
	if result.status != http.StatusOK {
		// #nosec G706 -- request id and path are sanitized with utils.SafeLog; the code is a fixed enum.
		log.Printf("WARN open-data-service admin_access_denied code=%s requestId=%s path=%s", result.code, requestID, utils.SafeLog(r.URL.Path))
		utils.WriteError(w, result.status, result.code, result.message)
		return adminActor{}, false
	}
	actor := result.actor
	// #nosec G706 -- actor, request id, and path are sanitized with utils.SafeLog.
	log.Printf("INFO open-data-service admin_action actor=%s role=%s agency=%s requestId=%s path=%s",
		utils.SafeLog(actor.ID), utils.SafeLog(actor.Role), utils.SafeLog(actor.AgencyID), requestID, utils.SafeLog(r.URL.Path))
	return actor, true
}
