package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.healthHandler)
	mux.HandleFunc("GET /routes/options", s.optionsHandler)
	mux.HandleFunc("POST /routes/plan", s.planRouteHandler)
	return s.withMiddleware(mux)
}
