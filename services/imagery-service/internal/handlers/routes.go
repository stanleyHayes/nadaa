package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("POST /api/v1/imagery", s.createImageryHandler)
	mux.HandleFunc("GET /api/v1/imagery", s.listImageryHandler)
	mux.HandleFunc("GET /api/v1/imagery/geojson", s.geoJSONHandler)
	mux.HandleFunc("POST /api/v1/imagery/lifecycle/run", s.runLifecycleHandler)
	mux.HandleFunc("GET /api/v1/imagery/{id}", s.getImageryHandler)
	mux.HandleFunc("GET /api/v1/imagery/{id}/download", s.downloadImageryHandler)
	mux.HandleFunc("DELETE /api/v1/imagery/{id}", s.deleteImageryHandler)
	mux.HandleFunc("POST /api/v1/imagery/{id}/expire", s.expireImageryHandler)
	return s.withMiddleware(mux)
}
