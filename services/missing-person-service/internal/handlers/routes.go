package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/missing-persons", s.listPublicMissingPersonsHandler)
	mux.HandleFunc("POST /api/v1/missing-persons", s.createMissingPersonHandler)
	mux.HandleFunc("GET /api/v1/missing-persons/{id}", s.getPublicMissingPersonHandler)
	mux.HandleFunc("GET /api/v1/authority/missing-persons", s.listAuthorityMissingPersonsHandler)
	mux.HandleFunc("GET /api/v1/authority/missing-persons/{id}", s.getAuthorityMissingPersonHandler)
	mux.HandleFunc("PATCH /api/v1/authority/missing-persons/{id}/review", s.reviewMissingPersonHandler)
	mux.HandleFunc("PATCH /api/v1/authority/missing-persons/{id}/close", s.closeMissingPersonHandler)
	mux.HandleFunc("GET /api/v1/authority/missing-persons/{id}/audit", s.listAuditHandler)
	return s.withMiddleware(mux)
}
