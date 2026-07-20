package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/register", s.registerCitizenHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/login/otp", s.requestCitizenOTPHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/login", s.loginCitizenHandler)
	mux.HandleFunc("GET /api/v1/auth/me", s.meHandler)
	mux.HandleFunc("POST /api/v1/auth/agency-users", s.createAgencyUserHandler)
	mux.HandleFunc("GET /api/v1/auth/agency-users", s.listAgencyUsersHandler)
	mux.HandleFunc("POST /api/v1/auth/agency-users/{id}/mfa/setup", s.setupAgencyMFAHandler)
	mux.HandleFunc("POST /api/v1/auth/agency-users/{id}/mfa/verify", s.verifyAgencyMFAHandler)
	mux.HandleFunc("POST /api/v1/auth/agency/login", s.loginAgencyHandler)
	mux.HandleFunc("POST /api/v1/auth/agency/password", s.changeAgencyPasswordHandler)
	mux.HandleFunc("GET /api/v1/auth/agencies", s.listAgenciesHandler)
	mux.HandleFunc("GET /api/v1/audit/logs", s.listAuditLogsHandler)
	mux.HandleFunc("POST /api/v1/audit/logs", s.ingestAuditLogHandler)

	return utils.WithCORS(s.config.AllowedOrigins, mux)
}
