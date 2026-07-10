package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/schools", s.listSchoolsHandler)
	mux.HandleFunc("POST /api/v1/schools", s.createSchoolHandler)
	mux.HandleFunc("GET /api/v1/schools/{id}", s.getSchoolHandler)
	mux.HandleFunc("PUT /api/v1/schools/{id}", s.updateSchoolHandler)
	mux.HandleFunc("GET /api/v1/schools/{id}/drills", s.listDrillsHandler)
	mux.HandleFunc("POST /api/v1/schools/{id}/drills", s.createDrillHandler)
	mux.HandleFunc("GET /api/v1/schools/{id}/readiness", s.getReadinessHandler)
	mux.HandleFunc("POST /api/v1/schools/{id}/readiness", s.createReadinessHandler)
	return s.withMiddleware(mux)
}
