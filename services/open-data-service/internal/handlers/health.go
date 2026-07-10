package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/utils"
)

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "open-data-service"})
}
