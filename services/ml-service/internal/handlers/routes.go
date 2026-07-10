package handlers

import "net/http"

// routes registers the ml-service HTTP handlers on a fresh ServeMux.
func (s *server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("POST /api/v1/ml/flood/predictions", s.createFloodPredictionHandler)
	mux.HandleFunc("GET /api/v1/ml/prediction-logs", s.listPredictionLogsHandler)
	mux.HandleFunc("POST /api/v1/ml/flood/simulations", s.createSimulationHandler)
	mux.HandleFunc("GET /api/v1/ml/flood/simulations", s.listSimulationsHandler)
	mux.HandleFunc("GET /api/v1/ml/flood/simulations/{id}", s.getSimulationHandler)
	mux.HandleFunc("POST /api/v1/cv/analyze", s.createCVAnalysisHandler)
	mux.HandleFunc("GET /api/v1/cv/results/{imageId}", s.getCVResultHandler)
	mux.HandleFunc("GET /api/v1/cv/results", s.listCVResultsHandler)
	mux.HandleFunc("GET /api/v1/forecasts", s.listForecastsHandler)
	mux.HandleFunc("POST /api/v1/forecasts/compare", s.compareScenariosHandler)
	mux.HandleFunc("GET /api/v1/forecasts/{region}", s.getForecastByRegionHandler)
	mux.HandleFunc("GET /api/v1/staging-suggestions", s.listStagingSuggestionsHandler)
	return mux
}
