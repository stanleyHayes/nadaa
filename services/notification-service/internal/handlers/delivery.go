package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

var allowedDeliveryStatuses = map[string]bool{
	"queued":    true,
	"delivered": true,
	"failed":    true,
	"skipped":   true,
}

func (s *Server) listDeliveryLogsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseLogFilters(r)
	if code != "" {
		utils.LogWarn("delivery log list rejected", "code", code, "message", message, "query", r.URL.RawQuery)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	utils.LogInfo("delivery log list requested", "alertId", filters.AlertID, "channel", filters.Channel, "status", filters.Status)
	logs := s.store.ListDeliveryLogs(filters)
	utils.LogInfo("delivery log list completed", "count", len(logs))
	utils.WriteJSON(w, http.StatusOK, models.DeliveryLogListResponse{Logs: logs})
}

func parseLogFilters(r *http.Request) (models.LogFilters, string, string) {
	query := r.URL.Query()
	filters := models.LogFilters{
		AlertID: utils.NormalizeID(query.Get("alertId")),
		Channel: utils.NormalizeQueryValue(query.Get("channel")),
		Status:  utils.NormalizeQueryValue(query.Get("status")),
	}
	if filters.Channel != "" && !allowedLogChannels[filters.Channel] {
		return models.LogFilters{}, "invalid_channel", "channel must be push, sms, voice, or cell_broadcast"
	}
	if filters.Status != "" && !allowedDeliveryStatuses[filters.Status] {
		return models.LogFilters{}, "invalid_status", "status must be queued, delivered, failed, or skipped"
	}
	return filters, "", ""
}
