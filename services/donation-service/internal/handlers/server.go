package handlers

import (
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/store"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store       store.Store
	payments    models.PaymentProvider
	rateLimiter *rateLimiter
	now         func() time.Time
	config      *config.Config
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, payments models.PaymentProvider, now func() time.Time, cfg *config.Config) *Server {
	return &Server{
		store:       s,
		payments:    payments,
		rateLimiter: newRateLimiter(cfg.DonationRateLimit, time.Duration(cfg.DonationRateWindowSecs)*time.Second, now),
		now:         now,
		config:      cfg,
	}
}

type rateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	requests  map[string][]time.Time
	lastSweep time.Time
	now       func() time.Time
}

func newRateLimiter(limit int, window time.Duration, now func() time.Time) *rateLimiter {
	if limit <= 0 {
		limit = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &rateLimiter{limit: limit, window: window, requests: map[string][]time.Time{}, now: now}
}

func (r *rateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	cutoff := now.Add(-r.window)

	// Periodically evict expired keys so rotated/spoofed identifiers cannot
	// grow the map without bound.
	if now.Sub(r.lastSweep) >= r.window {
		for existing, events := range r.requests {
			kept := events[:0]
			for _, event := range events {
				if event.After(cutoff) {
					kept = append(kept, event)
				}
			}
			if len(kept) == 0 {
				delete(r.requests, existing)
			} else {
				r.requests[existing] = kept
			}
		}
		r.lastSweep = now
	}

	events := r.requests[key]
	kept := events[:0]
	for _, event := range events {
		if event.After(cutoff) {
			kept = append(kept, event)
		}
	}

	if len(kept) >= r.limit {
		r.requests[key] = kept
		return false
	}

	kept = append(kept, now)
	r.requests[key] = kept
	return true
}
