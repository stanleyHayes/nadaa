package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// withMiddleware applies CORS and security headers to a handler.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}
