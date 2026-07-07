package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

// Routes returns the configured HTTP handler with middleware applied.
func (s *server) Routes(allowedOrigins map[string]bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/integrations/contracts", s.listContractsHandler)
	mux.HandleFunc("GET /api/v1/integrations/mock/weather-hydrology/observations", s.listObservationsHandler)
	mux.HandleFunc("GET /api/v1/integrations/weather-hydrology/observations", s.listImportedObservationsHandler)
	mux.HandleFunc("POST /api/v1/integrations/weather-hydrology/import-jobs", s.createObservationImportJobHandler)
	mux.HandleFunc("GET /api/v1/integrations/weather-hydrology/import-jobs", s.listObservationImportJobsHandler)
	mux.HandleFunc("POST /api/v1/integrations/weather-hydrology/import-jobs/{id}/retry", s.retryObservationImportJobHandler)
	mux.HandleFunc("POST /api/v1/integrations/mock/sync-events", s.createSyncEventHandler)
	mux.HandleFunc("GET /api/v1/integrations/mock/sync-events", s.listSyncEventsHandler)
	mux.HandleFunc("POST /api/v1/integrations/road-closures/imports", s.importRoadClosureHandler)
	mux.HandleFunc("GET /api/v1/integrations/road-closures/imports", s.listRoadClosureImportsHandler)

	return utils.WithCORS(allowedOrigins, mux)
}
