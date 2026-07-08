package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

var allowedAccessChannels = map[string]bool{
	"sms":      true,
	"ussd":     true,
	"whatsapp": true,
}

var allowedAccessStatuses = map[string]bool{
	"handled":   true,
	"failed":    true,
	"queued":    true,
	"submitted": true,
}

var allowedAccessIntents = map[string]bool{
	"language_menu":     true,
	"main_menu":         true,
	"current_alerts":    true,
	"report_emergency":  true,
	"risk_check":        true,
	"emergency_guides":  true,
	"shelter_lookup":    true,
	"guidance_112":      true,
	"provider_error":    true,
	"invalid_selection": true,
}

func (s *Server) listAccessLogsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseAccessLogFilters(r)
	if code != "" {
		utils.LogWarn("inclusive access log list rejected", "code", code, "message", message, "query", r.URL.RawQuery)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	utils.LogInfo("inclusive access log list requested", "channel", filters.Channel, "intent", filters.Intent, "status", filters.Status)
	logs := s.store.ListAccessLogs(filters)
	utils.LogInfo("inclusive access log list completed", "count", len(logs))
	utils.WriteJSON(w, http.StatusOK, models.AccessLogListResponse{Logs: logs})
}

func parseAccessLogFilters(r *http.Request) (models.AccessLogFilters, string, string) {
	query := r.URL.Query()
	filters := models.AccessLogFilters{
		Channel: utils.NormalizeQueryValue(query.Get("channel")),
		Intent:  utils.NormalizeQueryValue(query.Get("intent")),
		Status:  utils.NormalizeQueryValue(query.Get("status")),
	}
	if filters.Channel != "" && !allowedAccessChannels[filters.Channel] {
		return models.AccessLogFilters{}, "invalid_channel", "channel must be sms, ussd, or whatsapp"
	}
	if filters.Intent != "" && !allowedAccessIntents[filters.Intent] {
		return models.AccessLogFilters{}, "invalid_intent", "intent must be a supported inclusive access intent"
	}
	if filters.Status != "" && !allowedAccessStatuses[filters.Status] {
		return models.AccessLogFilters{}, "invalid_status", "status must be handled, failed, queued, or submitted"
	}
	return filters, "", ""
}
