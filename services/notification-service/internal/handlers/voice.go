package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

var allowedVoiceLanguages = map[string]bool{
	"en":  true,
	"tw":  true,
	"ga":  true,
	"ee":  true,
	"dag": true,
	"ha":  true,
}

func (s *Server) createVoiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	var request models.VoiceAlertRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("voice alert generation rejected", "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.AlertID = utils.NormalizeID(request.AlertID)
	request.WorkflowRequestedBy = utils.NormalizeID(request.WorkflowRequestedBy)
	languages, code, message := normalizeVoiceLanguages(request.Languages)
	if code != "" {
		utils.LogWarn("voice alert generation rejected", "alertId", request.AlertID, "code", code, "message", message)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	source, ok := normalizeVoiceSource(request.Source)
	if !ok {
		utils.LogWarn("voice alert generation rejected", "alertId", request.AlertID, "code", "invalid_source", "source", request.Source)
		utils.WriteError(w, http.StatusBadRequest, "invalid_source", "source must be tts_sandbox or recorded_audio")
		return
	}
	if request.AlertID == "" {
		utils.LogWarn("voice alert generation rejected", "code", "missing_alert_id")
		utils.WriteError(w, http.StatusBadRequest, "missing_alert_id", "alertId is required")
		return
	}

	now := s.now()
	utils.LogInfo(
		"voice alert generation requested",
		"alertId", request.AlertID,
		"languages", strings.Join(languages, ","),
		"source", source,
		"requestedBy", request.WorkflowRequestedBy,
	)
	alerts, _ := s.listCitizenAlerts(r.Context(), models.AlertFeedFilters{Status: "all"}, now)
	alert, found := findAlert(alerts, request.AlertID)
	if !found {
		utils.LogWarn("voice alert generation rejected", "alertId", request.AlertID, "code", "alert_not_found", "availableAlerts", len(alerts))
		utils.WriteError(w, http.StatusNotFound, "alert_not_found", "alert was not found in the citizen feed")
		return
	}
	if alert.Status == "expired" {
		utils.LogWarn("voice alert generation rejected", "alertId", request.AlertID, "code", "alert_not_deliverable", "status", alert.Status)
		utils.WriteError(w, http.StatusConflict, "alert_not_deliverable", "voice alerts can only be generated for current or upcoming alerts")
		return
	}

	asset := s.store.CreateVoiceAlertAsset(alert, languages, source, request.WorkflowRequestedBy, now)
	utils.LogInfo(
		"voice alert generation completed",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"variantCount", len(asset.Variants),
		"reviewStatus", asset.ReviewStatus,
	)
	utils.WriteJSON(w, http.StatusCreated, models.VoiceAlertResponse{Asset: asset})
}

func (s *Server) listVoiceAlertsHandler(w http.ResponseWriter, _ *http.Request) {
	utils.LogInfo("voice alert list requested")
	assets := s.store.ListVoiceAlertAssets()
	utils.LogInfo("voice alert list completed", "count", len(assets))
	utils.WriteJSON(w, http.StatusOK, models.VoiceAlertListResponse{Assets: assets})
}

func (s *Server) reviewVoiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.NormalizeID(r.PathValue("id"))
	if id == "" {
		utils.LogWarn("voice alert review rejected", "code", "missing_voice_asset_id", "path", r.URL.Path)
		utils.WriteError(w, http.StatusBadRequest, "missing_voice_asset_id", "voice alert asset id is required")
		return
	}

	var request models.VoiceReviewRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("voice alert review rejected", "voiceAssetId", id, "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	action := normalizeVoiceReviewAction(request.Action)
	if action == "" {
		utils.LogWarn("voice alert review rejected", "voiceAssetId", id, "code", "invalid_action", "action", request.Action)
		utils.WriteError(w, http.StatusBadRequest, "invalid_action", "action must be approve or reject")
		return
	}
	request.Reviewer = utils.NormalizeID(request.Reviewer)
	if request.Reviewer == "" {
		utils.LogWarn("voice alert review rejected", "voiceAssetId", id, "code", "missing_reviewer")
		utils.WriteError(w, http.StatusBadRequest, "missing_reviewer", "reviewer is required")
		return
	}

	languages, code, message := normalizeVoiceLanguages(request.Languages)
	if code != "" {
		utils.LogWarn("voice alert review rejected", "voiceAssetId", id, "code", code, "message", message)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	if len(request.Languages) == 0 {
		languages = nil
	}

	utils.LogInfo(
		"voice alert review requested",
		"voiceAssetId", id,
		"action", action,
		"reviewer", request.Reviewer,
		"languages", strings.Join(languages, ","),
	)
	asset, found := s.store.ReviewVoiceAlertAsset(id, action, request.Reviewer, strings.TrimSpace(request.Note), languages, s.now())
	if !found {
		utils.LogWarn("voice alert review rejected", "voiceAssetId", id, "code", "voice_asset_not_found")
		utils.WriteError(w, http.StatusNotFound, "voice_asset_not_found", "voice alert asset was not found")
		return
	}

	utils.LogInfo(
		"voice alert review completed",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"status", asset.Status,
		"reviewStatus", asset.ReviewStatus,
		"reviewer", asset.Reviewer,
	)
	utils.WriteJSON(w, http.StatusOK, models.VoiceAlertResponse{Asset: asset})
}

func (s *Server) deliverVoiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.NormalizeID(r.PathValue("id"))
	if id == "" {
		utils.LogWarn("voice alert delivery rejected", "code", "missing_voice_asset_id", "path", r.URL.Path)
		utils.WriteError(w, http.StatusBadRequest, "missing_voice_asset_id", "voice alert asset id is required")
		return
	}

	var request models.VoiceDeliveryRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("voice alert delivery rejected", "voiceAssetId", id, "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	recipients, code, message := normalizeVoiceRecipients(request.Recipients)
	if code != "" {
		utils.LogWarn("voice alert delivery rejected", "voiceAssetId", id, "code", code, "message", message)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	request.Recipients = recipients

	asset, found := s.store.GetVoiceAlertAsset(id)
	if !found {
		utils.LogWarn("voice alert delivery rejected", "voiceAssetId", id, "code", "voice_asset_not_found")
		utils.WriteError(w, http.StatusNotFound, "voice_asset_not_found", "voice alert asset was not found")
		return
	}
	if asset.Status != "approved" || asset.ReviewStatus != "approved" {
		utils.LogWarn(
			"voice alert delivery rejected",
			"voiceAssetId", id,
			"code", "voice_not_approved",
			"status", asset.Status,
			"reviewStatus", asset.ReviewStatus,
		)
		utils.WriteError(w, http.StatusConflict, "voice_not_approved", "voice alert asset must be approved before delivery")
		return
	}

	utils.LogInfo(
		"voice alert delivery requested",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"recipientCount", len(request.Recipients),
		"dryRun", request.DryRun,
	)
	attempts := s.store.CreateVoiceDeliveryAttempts(r.Context(), asset, request, s.providers, s.now())
	utils.LogInfo("voice alert delivery attempts created", "voiceAssetId", asset.ID, "alertId", asset.AlertID, "attemptCount", len(attempts))
	utils.WriteJSON(w, http.StatusAccepted, models.VoiceDeliveryResponse{Attempts: attempts})
}

func normalizeVoiceLanguages(languages []string) ([]string, string, string) {
	defaultLanguages := []string{"en", "tw", "ga", "ee", "dag", "ha"}
	if len(languages) == 0 {
		return append([]string(nil), defaultLanguages...), "", ""
	}

	result := make([]string, 0, len(languages))
	seen := map[string]bool{}
	for _, language := range languages {
		normalized, ok := normalizeVoiceLanguage(language)
		if !ok {
			return nil, "invalid_language", "languages must include only en, tw, ga, ee, dag, or ha"
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return nil, "missing_language", "at least one language is required"
	}
	return result, "", ""
}

func normalizeVoiceLanguage(value string) (string, bool) {
	value = utils.NormalizeQueryValue(value)
	switch value {
	case "", "english":
		value = "en"
	case "ak", "akan", "twi":
		value = "tw"
	case "gaa":
		value = "ga"
	case "ewe":
		value = "ee"
	case "dagbani":
		value = "dag"
	case "hausa":
		value = "ha"
	}
	return value, allowedVoiceLanguages[value]
}

func normalizeVoiceSource(value string) (string, bool) {
	value = utils.NormalizeQueryValue(value)
	if value == "" {
		return "tts_sandbox", true
	}
	if value == "tts_sandbox" || value == "recorded_audio" {
		return value, true
	}
	return "", false
}

func normalizeVoiceReviewAction(value string) string {
	value = utils.NormalizeQueryValue(value)
	switch value {
	case "approve", "approved":
		return "approve"
	case "reject", "rejected":
		return "reject"
	default:
		return ""
	}
}

func normalizeVoiceRecipients(recipients []models.VoiceRecipient) ([]models.VoiceRecipient, string, string) {
	if len(recipients) == 0 {
		return nil, "missing_recipients", "at least one voice recipient is required"
	}

	result := make([]models.VoiceRecipient, 0, len(recipients))
	for _, recipient := range recipients {
		recipient.RecipientID = utils.NormalizeID(recipient.RecipientID)
		recipient.Phone = strings.TrimSpace(recipient.Phone)
		language, ok := normalizeVoiceLanguage(recipient.Language)
		if !ok {
			return nil, "invalid_language", "recipient language must be en, tw, ga, ee, dag, or ha"
		}
		recipient.Language = language
		if recipient.RecipientID == "" && recipient.Phone == "" {
			return nil, "missing_recipient", "each voice recipient requires recipientId or phone"
		}
		result = append(result, recipient)
	}
	return result, "", ""
}
