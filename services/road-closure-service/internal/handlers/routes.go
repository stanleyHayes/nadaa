package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/road-closures", s.listRoadClosuresHandler)
	mux.HandleFunc("POST /api/v1/road-closures", s.createRoadClosureHandler)
	mux.HandleFunc("PATCH /api/v1/road-closures/{id}", s.updateRoadClosureHandler)
	mux.HandleFunc("POST /api/v1/road-closures/imports/adapter", s.importAdapterHandler)
	return s.withMiddleware(mux)
}
