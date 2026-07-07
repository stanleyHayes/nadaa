package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/utils"
)

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "guide-service"})
}
