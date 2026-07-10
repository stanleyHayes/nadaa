package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"
)

// healthHandler returns a simple service health check.
func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": s.store.Health(), "service": "damage-claim-service"})
}
