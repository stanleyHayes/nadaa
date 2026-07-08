package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

func (s *server) listPredictionLogsHandler(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, models.PredictionLogListResponse{Logs: s.store.ListLogs()})
}
