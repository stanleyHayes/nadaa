package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "integration-service"})
}
