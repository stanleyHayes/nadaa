package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/campaigns", s.listCampaignsHandler)
	mux.HandleFunc("GET /api/v1/campaigns/{id}", s.getCampaignHandler)
	mux.HandleFunc("POST /api/v1/campaigns", s.createCampaignHandler)
	mux.HandleFunc("PUT /api/v1/campaigns/{id}", s.updateCampaignHandler)
	mux.HandleFunc("GET /api/v1/campaigns/{id}/metrics", s.getCampaignMetricsHandler)
	mux.HandleFunc("GET /api/v1/campaign-templates", s.listCampaignTemplatesHandler)
	return s.withMiddleware(mux)
}
