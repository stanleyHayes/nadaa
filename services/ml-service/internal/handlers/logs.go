package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

func (s *server) listPredictionLogsHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := parsePagination(w, r)
	if !ok {
		return
	}
	logs, total := s.store.ListLogs(limit, offset)
	utils.WriteJSON(w, http.StatusOK, models.PredictionLogListResponse{
		Logs:   logs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
