package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

func (s *Server) ussdWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var request models.USSDWebhookRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("ussd webhook rejected", "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.SessionID = strings.TrimSpace(request.SessionID)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Language = utils.NormalizeLanguage(request.Language)
	request.Provider = utils.ProviderOrDefault(request.Provider, "ussd_sandbox")
	request.ProfileID = utils.NormalizeID(request.ProfileID)
	request.ProviderError = strings.TrimSpace(request.ProviderError)

	utils.LogInfo(
		"ussd webhook received",
		"sessionId", request.SessionID,
		"provider", request.Provider,
		"phoneRef", utils.PhoneRef(request.Phone),
		"pathDepth", len(ussdTokens(request.Text)),
		"hasProviderError", request.ProviderError != "",
		"linkedProfileRequested", request.LinkProfile,
	)
	if request.SessionID == "" {
		utils.LogWarn("ussd webhook rejected", "code", "missing_session", "provider", request.Provider, "phoneRef", utils.PhoneRef(request.Phone))
		utils.WriteError(w, http.StatusBadRequest, "missing_session", "sessionId is required")
		return
	}
	if request.Phone == "" {
		utils.LogWarn("ussd webhook rejected", "code", "missing_phone", "sessionId", request.SessionID, "provider", request.Provider)
		utils.WriteError(w, http.StatusBadRequest, "missing_phone", "phone is required")
		return
	}

	response := s.handleUSSDRequest(r.Context(), request)
	utils.WriteJSON(w, http.StatusOK, response)
}

func (s *Server) handleUSSDRequest(ctx context.Context, request models.USSDWebhookRequest) models.USSDWebhookResponse {
	now := s.now()
	phoneRef := utils.PhoneRef(request.Phone)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := normalizeUSSDLanguage(request.Language, request.Text)
	utils.LogInfo(
		"ussd session handling started",
		"sessionId", request.SessionID,
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"language", language,
		"pathDepth", len(ussdTokens(request.Text)),
		"linkedProfile", linkedProfile,
	)

	if request.ProviderError != "" {
		utils.LogWarn(
			"ussd provider error received",
			"sessionId", request.SessionID,
			"provider", request.Provider,
			"providerMessageId", request.ProviderMessageID,
			"phoneRef", phoneRef,
			"errorLength", len(request.ProviderError),
		)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "ussd",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         request.SessionID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "provider_error",
			Status:            "failed",
			ProviderError:     request.ProviderError,
			CreatedAt:         now,
		})
		return models.USSDWebhookResponse{
			SessionID: request.SessionID,
			Action:    "end",
			Message:   localizedMessage(language, "provider_error"),
			Language:  language,
			Log:       log,
		}
	}

	tokens := ussdTokens(request.Text)
	if len(tokens) == 0 {
		utils.LogInfo("ussd language menu returned", "sessionId", request.SessionID, "provider", request.Provider, "phoneRef", phoneRef)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "language_menu",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: languageMenu(), Language: language, Log: log}
	}

	if _, ok := ussdLanguageFromToken(tokens[0]); !ok {
		utils.LogWarn(
			"ussd invalid language selection",
			"sessionId", request.SessionID,
			"provider", request.Provider,
			"phoneRef", phoneRef,
			"pathDepth", len(tokens),
		)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: languageMenu(), Language: language, Log: log}
	}

	if len(tokens) == 1 {
		utils.LogInfo("ussd main menu returned", "sessionId", request.SessionID, "provider", request.Provider, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "main_menu",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: mainMenu(language), Language: language, Log: log}
	}

	switch tokens[1] {
	case "1":
		alerts, _ := s.listCitizenAlerts(ctx, models.AlertFeedFilters{}, now)
		utils.LogInfo("ussd current alerts summary returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "alertCount", len(alerts))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "current_alerts",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "end", Message: alertSummaryMessage(language, alerts), Language: language, Log: log}
	case "2":
		utils.LogInfo("ussd report flow selected", "sessionId", request.SessionID, "phoneRef", phoneRef, "pathDepth", len(tokens))
		return s.handleUSSDReport(ctx, request, tokens, language, linkedProfile, now)
	case "3":
		utils.LogInfo("ussd shelter guidance returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "shelter_lookup",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "end", Message: shelterMessage(language), Language: language, Log: log}
	case "4":
		utils.LogInfo("ussd 112 guidance returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "guidance_112",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "end", Message: guidance112Message(language), Language: language, Log: log}
	default:
		utils.LogWarn("ussd invalid main-menu selection", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language, "pathDepth", len(tokens))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: mainMenu(language), Language: language, Log: log}
	}
}

func (s *Server) handleUSSDReport(ctx context.Context, request models.USSDWebhookRequest, tokens []string, language string, linkedProfile bool, now time.Time) models.USSDWebhookResponse {
	phoneRef := utils.PhoneRef(request.Phone)
	if len(tokens) == 2 {
		utils.LogInfo("ussd report hazard menu returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "report_emergency",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: hazardMenu(language), Language: language, Log: log}
	}

	hazard, ok := ussdHazardFromToken(tokens[2])
	if !ok {
		utils.LogWarn("ussd report rejected invalid hazard", "sessionId", request.SessionID, "phoneRef", phoneRef, "pathDepth", len(tokens))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: hazardMenu(language), Language: language, Log: log}
	}

	if len(tokens) == 3 {
		utils.LogInfo("ussd report urgency menu returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "hazard", hazard, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "report_emergency",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: urgencyMenu(language), Language: language, Log: log}
	}

	urgency, ok := ussdUrgencyFromToken(tokens[3])
	if !ok {
		utils.LogWarn("ussd report rejected invalid urgency", "sessionId", request.SessionID, "phoneRef", phoneRef, "hazard", hazard, "pathDepth", len(tokens))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: urgencyMenu(language), Language: language, Log: log}
	}

	location, locationLabel := utils.InclusiveLocation(request.Location, tokens[4:])
	utils.LogInfo(
		"ussd report creating access report",
		"sessionId", request.SessionID,
		"phoneRef", phoneRef,
		"hazard", hazard,
		"urgency", urgency,
		"hasCoordinates", request.Location != nil,
		"locationLabel", utils.LogTextSummary(locationLabel),
		"linkedProfile", linkedProfile,
	)
	report := s.store.CreateAccessReport(models.InclusiveAccessReport{
		Channel:       "ussd",
		Type:          hazard,
		Urgency:       urgency,
		Description:   fmt.Sprintf("USSD emergency report: %s with %s urgency. Location note: %s.", hazard, urgency, locationLabel),
		Location:      location,
		LocationLabel: locationLabel,
		PhoneRef:      phoneRef,
		ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile: linkedProfile,
		Status:        "queued",
		CreatedAt:     now,
	})
	report = s.submitInclusiveReport(ctx, report, request.Phone, request.ProfileID, linkedProfile)
	utils.LogInfo(
		"ussd report flow completed",
		"sessionId", request.SessionID,
		"phoneRef", phoneRef,
		"reportId", report.ID,
		"status", report.Status,
		"incidentId", report.IncidentID,
		"incidentReference", report.IncidentReference,
	)

	log := s.store.CreateAccessLog(models.InclusiveAccessLog{
		Channel:           "ussd",
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		SessionID:         request.SessionID,
		PhoneRef:          phoneRef,
		ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Language:          language,
		Intent:            "report_emergency",
		Status:            report.Status,
		IncidentID:        report.IncidentID,
		IncidentReference: report.IncidentReference,
		CreatedAt:         now,
	})

	message := reportConfirmationMessage(language, report)
	return models.USSDWebhookResponse{SessionID: request.SessionID, Action: "end", Message: message, Language: language, Log: log, Report: &report}
}

func normalizeUSSDLanguage(defaultLanguage string, text string) string {
	tokens := ussdTokens(text)
	if len(tokens) > 0 {
		if language, ok := ussdLanguageFromToken(tokens[0]); ok {
			return language
		}
	}
	return utils.NormalizeLanguage(defaultLanguage)
}

func ussdTokens(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	parts := strings.Split(text, "*")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return tokens
}

func ussdLanguageFromToken(token string) (string, bool) {
	switch token {
	case "1":
		return "en", true
	case "2":
		return "tw", true
	case "3":
		return "ga", true
	case "4":
		return "ee", true
	case "5":
		return "dag", true
	case "6":
		return "ha", true
	default:
		return "", false
	}
}

func ussdHazardFromToken(token string) (string, bool) {
	switch token {
	case "1":
		return "flood", true
	case "2":
		return "fire", true
	case "3":
		return "medical_emergency", true
	case "4":
		return "road_crash", true
	case "5":
		return "other", true
	default:
		return "", false
	}
}

func ussdUrgencyFromToken(token string) (string, bool) {
	switch token {
	case "1":
		return "low", true
	case "2":
		return "moderate", true
	case "3":
		return "high", true
	case "4":
		return "life_threatening", true
	default:
		return "", false
	}
}
