package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "alert-service"})
}
