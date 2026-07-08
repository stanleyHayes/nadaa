package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

// withMiddleware applies server-wide middleware to the handler.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}
