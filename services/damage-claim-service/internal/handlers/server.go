package handlers

import (
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/store"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store              store.Store
	now                func() time.Time
	config             *config.Config
	httpClient         *http.Client
	incidentServiceURL string
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, now func() time.Time, cfg *config.Config) *Server {
	return &Server{
		store:              s,
		now:                now,
		config:             cfg,
		httpClient:         http.DefaultClient,
		incidentServiceURL: cfg.IncidentServiceURL,
	}
}
