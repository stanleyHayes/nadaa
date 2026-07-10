package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
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
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}

func (s *Server) checkRateLimit(ip string) (models.RateLimitStatus, bool) {
	s.rateMu.Lock()
	defer s.rateMu.Unlock()

	now := s.now().UTC()
	window := time.Duration(s.config.RateLimitWindowSeconds) * time.Second
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
		req.Header.Set("X-NADAA-Actor-Role", evt.ActorRole)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return
		}
		_ = resp.Body.Close()
	}(event)
}

// requireAdmin enforces the full authority context for open-data admin actions:
// actor id, agency id, and role must be present, MFA must be completed, and the
// role must be an allowed admin role. It writes the 401/403 response and logs the
// actor for audit. Returns true only when the caller is authorized.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	actorID := strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID"))
	agencyID := strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID"))
	role := strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role")))
	mfaCompleted := strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true"
	requestID := strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID"))

	if actorID == "" || agencyID == "" || role == "" {
		log.Printf("WARN open-data-service authority_context_missing requestId=%s path=%s", requestID, r.URL.Path)
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return false
	}
	if !mfaCompleted {
		log.Printf("WARN open-data-service mfa_required actor=%s role=%s requestId=%s path=%s", actorID, role, requestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for open-data admin actions")
		return false
	}
	if !openDataAdminRoles[role] {
		log.Printf("WARN open-data-service forbidden actor=%s role=%s requestId=%s path=%s", actorID, role, requestID, r.URL.Path)
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to review open-data access requests")
		return false
	}
	log.Printf("INFO open-data-service admin_action actor=%s role=%s agency=%s requestId=%s path=%s", actorID, role, agencyID, requestID, r.URL.Path)
	return true
}
