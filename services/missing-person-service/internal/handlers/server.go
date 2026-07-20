package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/store"
)

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
	rateLimits map[string]*bucket
	rateMu     sync.Mutex
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, now func() time.Time, cfg *config.Config) *Server {
	return &Server{store: s, now: now, config: cfg, rateLimits: map[string]*bucket{}}
}

// clientIP derives the rate-limit key from the connection's remote address.
// Proxy-supplied headers are trivially forgeable, so they are never consulted.
func (s *Server) clientIP(r *http.Request) string {
	// RemoteAddr carries the ephemeral source port; rate limiting keys on the host.
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// checkRateLimit consumes one token from the caller's per-IP bucket and
// reports whether the request may proceed. A configured limit <= 0 disables
// the limiter.
func (s *Server) checkRateLimit(ip string) bool {
	if s.config.RateLimitRequests <= 0 {
		return true
	}

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

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}
