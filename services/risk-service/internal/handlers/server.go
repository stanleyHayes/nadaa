package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/utils"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store    store.Store
	mlClient *mlClient
	config   *config.Config
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, cfg *config.Config) *Server {
	srv := &Server{store: s, config: cfg}
	if cfg.MLAPIURL != "" {
		srv.mlClient = newMLClient(cfg.MLAPIURL, http.DefaultClient)
	}
	return srv
}

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/risk", s.riskHandler)
	return utils.WithCORS(s.config.AllowedOrigins, mux)
}
