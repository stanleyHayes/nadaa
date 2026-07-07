package handlers

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/store"
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
