package handlers

import (
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/store"
)

// server holds the HTTP handler dependencies.
type server struct {
	store  store.Store
	now    func() time.Time
	config *config.Config
}

// NewServer creates a new server with the given dependencies.
func NewServer(s store.Store, now func() time.Time, cfg *config.Config) *server {
	return &server{store: s, now: now, config: cfg}
}

// Routes returns the configured HTTP handler with middleware applied.
func (s *server) Routes() http.Handler {
	return s.withMiddleware(s.routes())
}
