package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

func (s *Server) createCellBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CellBroadcastRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("cell broadcast generation rejected", "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.AlertID = utils.NormalizeID(request.AlertID)
	request.WorkflowRequestedBy = utils.NormalizeID(request.WorkflowRequestedBy)
	if request.AlertID == "" {
		utils.LogWarn("cell broadcast generation rejected", "code", "missing_alert_id")
		utils.WriteError(w, http.StatusBadRequest, "missing_alert_id", "alertId is required")
		return
	}

	languages, code, message := normalizeVoiceLanguages(request.Languages)
	if code != "" {
		utils.LogWarn("cell broadcast generation rejected", "alertId", request.AlertID, "code", code, "message", message)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	areas := normalizeCellBroadcastAreas(request.Areas)

	now := s.now()
	alerts, _ := s.listCitizenAlerts(r.Context(), models.AlertFeedFilters{Status: "all"}, now)
	alert, found := findAlert(alerts, request.AlertID)
	if !found {
		utils.LogWarn("cell broadcast generation rejected", "alertId", request.AlertID, "code", "alert_not_found", "availableAlerts", len(alerts))
		utils.WriteError(w, http.StatusNotFound, "alert_not_found", "alert was not found in the citizen feed")
		return
	}
	if alert.Status == "expired" {
		utils.LogWarn("cell broadcast generation rejected", "alertId", request.AlertID, "code", "alert_not_deliverable", "status", alert.Status)
		utils.WriteError(w, http.StatusConflict, "alert_not_deliverable", "cell broadcasts can only be generated for current or upcoming alerts")
		return
	}

	generated := s.store.CreateCellBroadcastMessage(alert, languages, areas, request.WorkflowRequestedBy, now)
	utils.LogInfo(
		"cell broadcast generation completed",
		"cellBroadcastId", generated.ID,
		"alertId", generated.AlertID,
		"channel", generated.Channel.MessageIdentifier,
		"segmentCount", len(generated.Segments),
		"reviewStatus", generated.ReviewStatus,
	)
	utils.WriteJSON(w, http.StatusCreated, models.CellBroadcastResponse{Message: generated})
}

func (s *Server) listCellBroadcastHandler(w http.ResponseWriter, _ *http.Request) {
	utils.LogInfo("cell broadcast list requested")
	messages := s.store.ListCellBroadcastMessages()
	utils.LogInfo("cell broadcast list completed", "count", len(messages))
	utils.WriteJSON(w, http.StatusOK, models.CellBroadcastListResponse{Messages: messages})
}

func (s *Server) previewCellBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.NormalizeID(r.PathValue("id"))
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_cell_broadcast_id", "cell broadcast id is required")
		return
	}

	message, found := s.store.GetCellBroadcastMessage(id)
	if !found {
		utils.WriteError(w, http.StatusNotFound, "cell_broadcast_not_found", "cell broadcast was not found")
		return
	}

	previews := make([]models.CellBroadcastSegmentPreview, 0, len(message.Segments))
	for _, segment := range message.Segments {
		previews = append(previews, models.CellBroadcastSegmentPreview{
			Language:         segment.Language,
			Locale:           segment.Locale,
			Channel:          message.Channel.Label,
			HandsetCategory:  message.Channel.HandsetCategory,
			DataCodingScheme: segment.DataCodingScheme,
			MessageText:      segment.MessageText,
			CharacterCount:   segment.CharacterCount,
			Pages:            segment.Pages,
			Truncated:        segment.Truncated,
		})
	}
	utils.LogInfo("cell broadcast preview generated", "cellBroadcastId", message.ID, "segmentCount", len(previews))
	utils.WriteJSON(w, http.StatusOK, models.CellBroadcastPreviewResponse{Message: message, Previews: previews})
}

func (s *Server) reviewCellBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.NormalizeID(r.PathValue("id"))
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_cell_broadcast_id", "cell broadcast id is required")
		return
	}

	var request models.CellBroadcastReviewRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("cell broadcast review rejected", "cellBroadcastId", id, "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	action := normalizeVoiceReviewAction(request.Action)
	if action == "" {
		utils.LogWarn("cell broadcast review rejected", "cellBroadcastId", id, "code", "invalid_action", "action", request.Action)
		utils.WriteError(w, http.StatusBadRequest, "invalid_action", "action must be approve or reject")
		return
	}
	request.Reviewer = utils.NormalizeID(request.Reviewer)
	if request.Reviewer == "" {
		utils.LogWarn("cell broadcast review rejected", "cellBroadcastId", id, "code", "missing_reviewer")
		utils.WriteError(w, http.StatusBadRequest, "missing_reviewer", "reviewer is required")
		return
	}

	languages, code, message := normalizeVoiceLanguages(request.Languages)
	if code != "" {
		utils.LogWarn("cell broadcast review rejected", "cellBroadcastId", id, "code", code, "message", message)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	if len(request.Languages) == 0 {
		languages = nil
	}

	reviewed, found := s.store.ReviewCellBroadcastMessage(id, action, request.Reviewer, strings.TrimSpace(request.Note), languages, s.now())
	if !found {
		utils.LogWarn("cell broadcast review rejected", "cellBroadcastId", id, "code", "cell_broadcast_not_found")
		utils.WriteError(w, http.StatusNotFound, "cell_broadcast_not_found", "cell broadcast was not found")
		return
	}

	utils.LogInfo(
		"cell broadcast review completed",
		"cellBroadcastId", reviewed.ID,
		"alertId", reviewed.AlertID,
		"status", reviewed.Status,
		"reviewStatus", reviewed.ReviewStatus,
		"reviewer", reviewed.Reviewer,
	)
	utils.WriteJSON(w, http.StatusOK, models.CellBroadcastResponse{Message: reviewed})
}

func (s *Server) deliverCellBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.NormalizeID(r.PathValue("id"))
	if id == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_cell_broadcast_id", "cell broadcast id is required")
		return
	}

	var request models.CellBroadcastDeliveryRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("cell broadcast delivery rejected", "cellBroadcastId", id, "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.Areas = normalizeCellBroadcastAreas(request.Areas)

	message, found := s.store.GetCellBroadcastMessage(id)
	if !found {
		utils.LogWarn("cell broadcast delivery rejected", "cellBroadcastId", id, "code", "cell_broadcast_not_found")
		utils.WriteError(w, http.StatusNotFound, "cell_broadcast_not_found", "cell broadcast was not found")
		return
	}
	if message.Status != "approved" || message.ReviewStatus != "approved" {
		utils.LogWarn(
			"cell broadcast delivery rejected",
			"cellBroadcastId", id,
			"code", "cell_broadcast_not_approved",
			"status", message.Status,
			"reviewStatus", message.ReviewStatus,
		)
		utils.WriteError(w, http.StatusConflict, "cell_broadcast_not_approved", "cell broadcast must be approved before delivery")
		return
	}

	utils.LogInfo(
		"cell broadcast delivery requested",
		"cellBroadcastId", message.ID,
		"alertId", message.AlertID,
		"adapter", s.cellBroadcastAdapter().Name(),
		"channel", message.Channel.MessageIdentifier,
		"dryRun", request.DryRun,
	)
	dispatches := s.store.CreateCellBroadcastDispatches(r.Context(), message, request, s.cellBroadcastAdapter(), s.now())
	utils.LogInfo("cell broadcast dispatch completed", "cellBroadcastId", message.ID, "alertId", message.AlertID, "dispatchCount", len(dispatches))
	utils.WriteJSON(w, http.StatusAccepted, models.CellBroadcastDeliveryResponse{Dispatches: dispatches})
}

// cellBroadcastAdapter returns the configured adapter, defaulting to a disabled
// no-op so an unconfigured deployment can never emit a live broadcast.
func (s *Server) cellBroadcastAdapter() models.CellBroadcastAdapter {
	if s.cellBroadcast == nil {
		return models.DisabledCellBroadcastAdapter{Reason: "cell broadcast adapter not configured"}
	}
	return s.cellBroadcast
}

func normalizeCellBroadcastAreas(areas []string) []string {
	result := make([]string, 0, len(areas))
	seen := map[string]bool{}
	for _, area := range areas {
		trimmed := strings.TrimSpace(area)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, trimmed)
	}
	return result
}
