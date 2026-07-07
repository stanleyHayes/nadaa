package handlers

import "net/http"

// routes registers the ml-service HTTP handlers on a fresh ServeMux.
func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("POST /api/v1/ml/flood/predictions", s.createFloodPredictionHandler)
	mux.HandleFunc("GET /api/v1/ml/prediction-logs", s.listPredictionLogsHandler)
	return mux
}
