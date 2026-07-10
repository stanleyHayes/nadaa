package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/open-data/datasets", s.listDatasetsHandler)
	mux.HandleFunc("GET /api/v1/open-data/datasets/{id}", s.getDatasetHandler)
	mux.HandleFunc("GET /api/v1/open-data/datasets/{id}/download", s.downloadDatasetHandler)
	mux.HandleFunc("POST /api/v1/open-data/requests", s.createRequestHandler)
	mux.HandleFunc("GET /api/v1/open-data/requests", s.listRequestsHandler)
	mux.HandleFunc("POST /api/v1/open-data/requests/{id}/approve", s.approveRequestHandler)
	return s.withMiddleware(mux)
}
