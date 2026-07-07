package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// withMiddleware applies CORS and security headers to the handler.
func (s *server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}
