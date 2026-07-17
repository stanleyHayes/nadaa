package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

var allowedHazards = map[string]bool{
	"flood":             true,
	"fire":              true,
	"road_crash":        true,
	"building_collapse": true,
	"medical_emergency": true,
	"security_incident": true,
	"disease_outbreak":  true,
	"electrical_hazard": true,
	"blocked_drain":     true,
	"landslide":         true,
	"marine_accident":   true,
	"storm":             true,
	"tidal_wave":        true,
	"other":             true,
}

var allowedSeverities = map[string]bool{
	"advisory":       true,
	"watch":          true,
	"warning":        true,
	"severe_warning": true,
	"emergency":      true,
}

var allowedTargetTypes = map[string]bool{
	"national":  true,
	"region":    true,
	"district":  true,
	"radius":    true,
	"community": true,
	"custom":    true,
}

var allowedChannels = map[string]bool{
	"push":  true,
	"sms":   true,
	"voice": true,
}

// allowedLogChannels is the set of channels that may appear in the unified
// delivery-log audit stream. It is broader than allowedChannels because cell
// broadcasts are audited here but must never be deliverable through the generic
// (non-approval-gated) delivery endpoint.
var allowedLogChannels = map[string]bool{
	"push":           true,
	"sms":            true,
	"voice":          true,
	"cell_broadcast": true,
}

var allowedFeedStatuses = map[string]bool{
	"current":  true,
	"expired":  true,
	"upcoming": true,
	"all":      true,
}

func (s *Server) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseAlertFeedFilters(r)
	if code != "" {
		utils.LogWarn("citizen alert list rejected", "code", code, "message", message, "query", r.URL.RawQuery)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now()
	utils.LogInfo(
		"citizen alert list requested",
		"hazard", filters.Hazard,
		"severity", filters.Severity,
		"status", filters.Status,
		"includeExpired", filters.IncludeExpired,
		"targetType", filters.TargetType,
		"targetId", filters.TargetID,
	)
	alerts, source := s.listCitizenAlerts(r.Context(), filters, now)
	utils.LogInfo("citizen alert list completed", "count", len(alerts), "source", source)
	utils.WriteJSON(w, http.StatusOK, models.CitizenAlertListResponse{Alerts: alerts, GeneratedAt: now, Source: source})
}

func (s *Server) deliverAlertHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAuthority(w, r, deliveryRoles)
	if !ok {
		return
	}

	id := utils.NormalizeID(r.PathValue("id"))
	if id == "" {
		utils.LogWarn("alert delivery rejected", "code", "missing_alert_id", "path", r.URL.Path)
		utils.WriteError(w, http.StatusBadRequest, "missing_alert_id", "alert id is required")
		return
	}

	var request models.DeliveryRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("alert delivery rejected", "alertId", id, "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.AlertID = id
	request.Channels = normalizeChannels(request.Channels)
	request.Language = utils.NormalizeLanguage(request.Language)
	utils.LogInfo(
		"alert delivery requested",
		"alertId", id,
		"channels", strings.Join(request.Channels, ","),
		"recipientRef", utils.RecipientRef(request, utils.PreferredRecipientChannel(request.Channels)),
		"dryRun", request.DryRun,
		"language", request.Language,
		"actorId", actor.ActorUserID,
		"actorRole", actor.ActorRole,
	)

	if code, message := validateDeliveryRequest(request); code != "" {
		utils.LogWarn("alert delivery rejected", "alertId", id, "code", code, "message", message, "channels", strings.Join(request.Channels, ","))
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now()
	alerts, _ := s.listCitizenAlerts(r.Context(), models.AlertFeedFilters{Status: "all"}, now)
	alert, ok := findAlert(alerts, id)
	if !ok {
		utils.LogWarn("alert delivery rejected", "alertId", id, "code", "alert_not_found", "availableAlerts", len(alerts))
		utils.WriteError(w, http.StatusNotFound, "alert_not_found", "alert was not found in the citizen feed")
		return
	}
	if alert.Status == "expired" {
		utils.LogWarn("alert delivery rejected", "alertId", id, "code", "alert_not_deliverable", "status", alert.Status)
		utils.WriteError(w, http.StatusConflict, "alert_not_deliverable", "alerts can only be delivered while current or upcoming")
		return
	}

	attempts := s.store.CreateDeliveryAttempts(r.Context(), alert, request, s.providers, now)
	utils.LogInfo("alert delivery attempts created", "alertId", id, "attemptCount", len(attempts))
	for _, attempt := range attempts {
		if attempt.Status == "failed" || attempt.Status == "skipped" {
			utils.LogWarn(
				"alert delivery attempt did not deliver",
				"attemptId", attempt.ID,
				"alertId", attempt.AlertID,
				"channel", attempt.Channel,
				"provider", attempt.Provider,
				"status", attempt.Status,
				"reason", attempt.Reason,
			)
			continue
		}
		utils.LogInfo(
			"alert delivery attempt completed",
			"attemptId", attempt.ID,
			"alertId", attempt.AlertID,
			"channel", attempt.Channel,
			"provider", attempt.Provider,
			"status", attempt.Status,
			"messageId", attempt.MessageID,
		)
	}
	utils.WriteJSON(w, http.StatusAccepted, models.DeliveryResponse{Attempts: attempts})
}

func (s *Server) listCitizenAlerts(ctx context.Context, filters models.AlertFeedFilters, now time.Time) ([]models.CitizenAlert, string) {
	fixtureAlerts := s.store.ListAlerts(filters, now)
	combined := make([]models.CitizenAlert, 0, len(fixtureAlerts))
	seen := map[string]bool{}
	for _, alert := range fixtureAlerts {
		combined = append(combined, alert)
		seen[alert.ID] = true
	}

	source := "fixture"
	if s.alertClient != nil {
		utils.LogInfo("alert-service fetch starting", "baseURL", s.alertClient.BaseURL, "fixtureCount", len(fixtureAlerts))
		upstreamAlerts, err := s.alertClient.ListAlerts(ctx, now)
		if err == nil {
			for _, alert := range upstreamAlerts {
				if seen[alert.ID] {
					continue
				}
				if alertMatchesFilters(alert, filters, now) {
					combined = append(combined, alert)
					seen[alert.ID] = true
				}
			}
			if len(upstreamAlerts) > 0 {
				source = "alert-service+fixture"
			}
			utils.LogInfo(
				"alert-service fetch completed",
				"upstreamCount", len(upstreamAlerts),
				"combinedCount", len(combined),
				"source", source,
			)
		} else {
			utils.LogWarn("alert-service fetch failed using fixture fallback", "error", err, "fixtureCount", len(fixtureAlerts))
		}
	}

	sortCitizenAlerts(combined)
	return combined, source
}

func parseAlertFeedFilters(r *http.Request) (models.AlertFeedFilters, string, string) {
	query := r.URL.Query()
	filters := models.AlertFeedFilters{
		Hazard:         utils.NormalizeQueryValue(query.Get("hazard")),
		Severity:       utils.NormalizeQueryValue(query.Get("severity")),
		Status:         utils.NormalizeQueryValue(query.Get("status")),
		TargetType:     utils.NormalizeQueryValue(query.Get("targetType")),
		TargetID:       utils.NormalizeID(query.Get("targetId")),
		IncludeExpired: utils.NormalizeQueryValue(query.Get("includeExpired")) == "true",
	}

	if filters.Hazard != "" && !allowedHazards[filters.Hazard] {
		return models.AlertFeedFilters{}, "invalid_hazard", "hazard must be a supported NADAA hazard type"
	}
	if filters.Severity != "" && !allowedSeverities[filters.Severity] {
		return models.AlertFeedFilters{}, "invalid_severity", "severity must be advisory, watch, warning, severe_warning, or emergency"
	}
	if filters.Status != "" && !allowedFeedStatuses[filters.Status] {
		return models.AlertFeedFilters{}, "invalid_status", "status must be current, expired, upcoming, or all"
	}
	if filters.TargetType != "" && !allowedTargetTypes[filters.TargetType] {
		return models.AlertFeedFilters{}, "invalid_target_type", "targetType must be national, region, district, radius, community, or custom"
	}

	return filters, "", ""
}

func validateDeliveryRequest(request models.DeliveryRequest) (string, string) {
	if request.RecipientID == "" && request.Phone == "" && request.PushToken == "" {
		return "missing_recipient", "recipientId, phone, or pushToken is required"
	}
	if len(request.Channels) == 0 {
		return "missing_channels", "channels must include push, sms, or both"
	}
	for _, channel := range request.Channels {
		if !allowedChannels[channel] {
			return "invalid_channel", "channels must include only push, sms, or voice"
		}
		if channel == "voice" {
			return "voice_requires_asset", "voice delivery requires an approved voice alert asset"
		}
		if channel == "push" && request.PushToken == "" && request.RecipientID == "" {
			return "missing_push_target", "push delivery requires pushToken or recipientId"
		}
		if channel == "sms" && request.Phone == "" {
			return "missing_sms_phone", "sms delivery requires phone"
		}
	}
	return "", ""
}

func normalizeChannels(channels []string) []string {
	if len(channels) == 0 {
		return []string{"push", "sms"}
	}
	result := make([]string, 0, len(channels))
	seen := map[string]bool{}
	for _, channel := range channels {
		channel = utils.NormalizeQueryValue(channel)
		if channel == "" || seen[channel] {
			continue
		}
		seen[channel] = true
		result = append(result, channel)
	}
	return result
}

func findAlert(alerts []models.CitizenAlert, id string) (models.CitizenAlert, bool) {
	for _, alert := range alerts {
		if alert.ID == id {
			return alert, true
		}
	}
	return models.CitizenAlert{}, false
}

func alertMatchesFilters(alert models.CitizenAlert, filters models.AlertFeedFilters, now time.Time) bool {
	alert.Status = alertFeedStatus(alert.StartsAt, alert.ExpiresAt, now)
	if filters.Hazard != "" && alert.HazardType != filters.Hazard {
		return false
	}
	if filters.Severity != "" && alert.Severity != filters.Severity {
		return false
	}
	if filters.TargetType != "" && alert.Target.Type != filters.TargetType {
		return false
	}
	if filters.TargetID != "" && !utils.ContainsString(alert.Target.IDs, filters.TargetID) {
		return false
	}
	if filters.Status == "all" {
		return true
	}
	if filters.Status != "" {
		return alert.Status == filters.Status
	}
	if filters.IncludeExpired {
		return alert.Status == "current" || alert.Status == "expired"
	}
	return alert.Status == "current"
}

func alertFeedStatus(startsAt time.Time, expiresAt time.Time, now time.Time) string {
	if now.Before(startsAt) {
		return "upcoming"
	}
	if !expiresAt.After(now) {
		return "expired"
	}
	return "current"
}

func sortCitizenAlerts(alerts []models.CitizenAlert) {
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Status != alerts[j].Status {
			return feedStatusRank(alerts[i].Status) < feedStatusRank(alerts[j].Status)
		}
		if alerts[i].Severity != alerts[j].Severity {
			return severityRank(alerts[i].Severity) > severityRank(alerts[j].Severity)
		}
		return alerts[i].StartsAt.After(alerts[j].StartsAt)
	})
}

func feedStatusRank(status string) int {
	switch status {
	case "current":
		return 0
	case "upcoming":
		return 1
	case "expired":
		return 2
	default:
		return 3
	}
}

func severityRank(severity string) int {
	switch severity {
	case "emergency":
		return 5
	case "severe_warning":
		return 4
	case "warning":
		return 3
	case "watch":
		return 2
	case "advisory":
		return 1
	default:
		return 0
	}
}
