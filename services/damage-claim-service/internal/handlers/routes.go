package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("POST /claims", s.createClaimHandler)
	mux.HandleFunc("GET /claims", s.listClaimsHandler)
	mux.HandleFunc("GET /claims/{id}", s.getClaimHandler)
	mux.HandleFunc("PATCH /claims/{id}", s.updateClaimHandler)
	mux.HandleFunc("POST /claims/{id}/verify", s.verifyClaimHandler)
	mux.HandleFunc("POST /claims/{id}/close", s.closeClaimHandler)
	mux.HandleFunc("GET /claims/{id}/export", s.exportClaimHandler)
	return s.withMiddleware(mux)
}
