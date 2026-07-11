package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/donors", s.listDonorsHandler)
	mux.HandleFunc("POST /api/v1/donors", s.createDonorHandler)
	mux.HandleFunc("GET /api/v1/donors/{id}", s.getDonorHandler)
	mux.HandleFunc("PATCH /api/v1/donors/{id}", s.updateDonorHandler)
	mux.HandleFunc("GET /api/v1/aid-catalog", s.listCatalogHandler)
	mux.HandleFunc("GET /api/v1/aid-requests", s.listAidRequestsHandler)
	mux.HandleFunc("POST /api/v1/aid-requests", s.createAidRequestHandler)
	mux.HandleFunc("GET /api/v1/aid-requests/{id}", s.getAidRequestHandler)
	mux.HandleFunc("PATCH /api/v1/aid-requests/{id}", s.updateAidRequestHandler)
	mux.HandleFunc("GET /api/v1/aid-requests/{id}/pledges", s.listRequestPledgesHandler)
	mux.HandleFunc("POST /api/v1/aid-requests/{id}/pledges", s.createPledgeHandler)
	mux.HandleFunc("GET /api/v1/pledges", s.listPledgesHandler)
	mux.HandleFunc("PATCH /api/v1/pledges/{id}", s.updatePledgeHandler)
	mux.HandleFunc("POST /api/v1/aid-requests/{id}/allocate", s.allocatePledgeHandler)
	mux.HandleFunc("GET /api/v1/donations", s.listDonationsHandler)
	mux.HandleFunc("POST /api/v1/donations", s.createDonationHandler)
	mux.HandleFunc("GET /api/v1/donations/{reference}", s.getDonationHandler)
	mux.HandleFunc("POST /api/v1/webhooks/paystack", s.paystackWebhookHandler)
	return s.withMiddleware(mux)
}
