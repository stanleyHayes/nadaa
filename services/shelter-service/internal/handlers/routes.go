package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

// Routes returns the configured HTTP handler with middleware applied.
func (s *server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/shelters", s.listSheltersHandler)
	mux.HandleFunc("GET /api/v1/shelters/nearby", s.nearbySheltersHandler)
	mux.HandleFunc("GET /api/v1/recovery-support/nearby", s.nearbyRecoverySupportHandler)
	mux.HandleFunc("PATCH /api/v1/shelters/{id}/occupancy", s.updateShelterOccupancyHandler)
	mux.HandleFunc("GET /api/v1/hospitals/capacity", s.listHospitalCapacityHandler)
	mux.HandleFunc("PATCH /api/v1/hospitals/{id}/capacity", s.updateHospitalCapacityHandler)
	mux.HandleFunc("POST /api/v1/hospitals/capacity/imports/fixture", s.importHospitalCapacityFixtureHandler)
	mux.HandleFunc("GET /api/v1/relief-points", s.listReliefPointsHandler)
	mux.HandleFunc("GET /api/v1/relief-points/nearby", s.nearbyReliefPointsHandler)
	mux.HandleFunc("POST /api/v1/relief-points", s.createReliefPointHandler)
	mux.HandleFunc("PATCH /api/v1/relief-points/{id}", s.updateReliefPointHandler)
	mux.HandleFunc("GET /api/v1/relief-points/{id}/stock-history", s.listReliefPointStockHistoryHandler)
	mux.HandleFunc("GET /api/v1/aid-requests", s.listAidRequestsHandler)
	mux.HandleFunc("POST /api/v1/aid-requests", s.createAidRequestHandler)
	mux.HandleFunc("GET /api/v1/aid-requests/report.csv", s.exportAidRequestsHandler)
	mux.HandleFunc("PATCH /api/v1/aid-requests/{id}/review", s.reviewAidRequestHandler)
	mux.HandleFunc("GET /api/v1/aid-requests/{id}/pledges", s.listAidPledgesHandler)
	mux.HandleFunc("POST /api/v1/aid-requests/{id}/pledges", s.createAidPledgeHandler)
	mux.HandleFunc("PATCH /api/v1/aid-requests/{id}/pledges/{pledgeId}/review", s.reviewAidPledgeHandler)

	return utils.WithCORS(s.config.AllowedOrigins, mux)
}
