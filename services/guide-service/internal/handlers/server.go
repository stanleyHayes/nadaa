// Package handlers provides the HTTP handlers for the guide service.
package handlers

import (
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/utils"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store  store.Store
	now    func() time.Time
	config *config.Config
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, now func() time.Time, cfg *config.Config) *Server {
	return &Server{store: s, now: now, config: cfg}
}

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/guides", s.listGuidesHandler)
	return utils.WithCORS(s.config.AllowedOrigins, mux)
}
