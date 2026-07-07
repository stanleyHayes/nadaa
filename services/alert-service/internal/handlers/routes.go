package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("POST /api/v1/alerts", s.createAlertHandler)
	mux.HandleFunc("GET /api/v1/alerts", s.listAlertsHandler)
	mux.HandleFunc("GET /api/v1/alerts/audit", s.listAuditHandler)
	mux.HandleFunc("POST /api/v1/alerts/targets/preview", s.previewTargetHandler)
	mux.HandleFunc("PATCH /api/v1/alerts/{id}", s.updateAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/submit", s.submitAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/approve", s.approveAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/reject", s.rejectAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/emergency-override", s.emergencyOverrideHandler)
	return utils.WithCORS(s.config.AllowedOrigins, mux)
}
