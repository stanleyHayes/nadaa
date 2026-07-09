package handlers

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/store"
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
