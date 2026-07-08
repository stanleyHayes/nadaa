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

func (s *Server) whatsappWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var request models.WhatsAppInboundRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("whatsapp webhook rejected", "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.From = strings.TrimSpace(request.From)
	request.Body = strings.TrimSpace(request.Body)
	request.Language = utils.NormalizeLanguage(request.Language)
	request.Provider = utils.ProviderOrDefault(request.Provider, "whatsapp_sandbox")
	request.ProfileID = utils.NormalizeID(request.ProfileID)
	request.ProviderError = strings.TrimSpace(request.ProviderError)
	request.Media = utils.NormalizeWhatsAppMedia(request.Media)

	utils.LogInfo(
		"whatsapp webhook received",
		"provider", request.Provider,
		"phoneRef", utils.PhoneRef(request.From),
		"command", utils.SMSCommandName(request.Body),
		"hasProviderError", request.ProviderError != "",
		"hasLocation", request.Location != nil,
		"mediaCount", len(request.Media),
		"linkedProfileRequested", request.LinkProfile,
	)
	if request.From == "" {
		utils.LogWarn("whatsapp webhook rejected", "code", "missing_from", "provider", request.Provider)
		utils.WriteError(w, http.StatusBadRequest, "missing_from", "from is required")
		return
	}
	if request.Body == "" && request.ProviderError == "" && request.Location == nil && len(request.Media) == 0 {
		utils.LogWarn("whatsapp webhook rejected", "code", "missing_body", "provider", request.Provider, "phoneRef", utils.PhoneRef(request.From))
		utils.WriteError(w, http.StatusBadRequest, "missing_body", "body, location, media, or providerError is required")
		return
	}

	response := s.handleWhatsAppInbound(r.Context(), request)
	utils.WriteJSON(w, http.StatusAccepted, response)
}

func (s *Server) handleWhatsAppInbound(ctx context.Context, request models.WhatsAppInboundRequest) models.WhatsAppInboundResponse {
	now := s.now()
	phoneRef := utils.PhoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language
	conversationKey := utils.WhatsAppConversationKey(phoneRef, request.ProfileID, linkedProfile)
	conversation := s.store.GetOrCreateWhatsAppConversation(
		conversationKey,
		phoneRef,
		utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		linkedProfile,
		language,
		now,
	)
	inboundTranscript := s.store.CreateWhatsAppTranscript(models.WhatsAppTranscript{
		ConversationID:    conversation.ID,
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		PhoneRef:          phoneRef,
		ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Direction:         "inbound",
		Intent:            "incoming",
		State:             conversation.State,
		MessageSummary:    utils.WhatsAppMessageSummary(request.Body),
		MediaSummary:      utils.WhatsAppMediaSummary(request.Media),
		CreatedAt:         now,
		RetentionUntil:    utils.WhatsAppRetentionUntil(now),
	})
	utils.LogInfo(
		"whatsapp inbound handling started",
		"conversationId", conversation.ID,
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"command", utils.SMSCommandName(request.Body),
		"state", conversation.State,
		"language", language,
		"linkedProfile", linkedProfile,
	)

	if request.ProviderError != "" {
		utils.LogWarn(
			"whatsapp provider error received",
			"conversationId", conversation.ID,
			"provider", request.Provider,
			"providerMessageId", request.ProviderMessageID,
			"phoneRef", phoneRef,
			"errorLength", len(request.ProviderError),
		)
		conversation.Intent = "provider_error"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "provider_error",
			Status:            "failed",
			ProviderError:     request.ProviderError,
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, localizedMessage(language, "provider_error"), log, nil, inboundTranscript.ID, now)
	}

	command := utils.SMSCommandName(request.Body)
	if command == "CANCEL" || command == "MENU" || command == "START" || command == "HI" || command == "HELLO" {
		conversation.Intent = "main_menu"
		conversation.State = "idle"
		conversation.Hazard = ""
		conversation.Urgency = ""
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp main menu returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "command", command)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "main_menu",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHelpMessage(), log, nil, inboundTranscript.ID, now)
	}

	if conversation.State != "" && conversation.State != "idle" && !utils.IsWhatsAppTopLevelCommand(command) {
		return s.handleWhatsAppReportConversation(ctx, request, conversation, inboundTranscript.ID, now)
	}

	switch command {
	case "ALERT", "ALERTS":
		alerts, _ := s.listCitizenAlerts(ctx, models.AlertFeedFilters{}, now)
		conversation.Intent = "current_alerts"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp alert summary returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "alertCount", len(alerts))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "current_alerts",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, alertSummaryMessage(language, alerts), log, nil, inboundTranscript.ID, now)
	case "RISK":
		alerts, _ := s.listCitizenAlerts(ctx, models.AlertFeedFilters{}, now)
		conversation.Intent = "risk_check"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp risk guidance returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "hasLocation", request.Location != nil, "alertCount", len(alerts))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "risk_check",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, riskCheckMessage(language, request.Location != nil, alerts), log, nil, inboundTranscript.ID, now)
	case "GUIDE", "GUIDES":
		hazard := utils.WhatsAppCommandArg(request.Body, 1)
		if hazard == "" {
			hazard = "flood"
		}
		hazard = utils.NormalizeSMSHazard(hazard)
		if !allowedHazards[hazard] {
			hazard = "flood"
		}
		conversation.Intent = "emergency_guides"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp emergency guide returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", hazard)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "emergency_guides",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, emergencyGuideMessage(language, hazard), log, nil, inboundTranscript.ID, now)
	case "SHELTER", "SHELTERS":
		conversation.Intent = "shelter_lookup"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp shelter guidance returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "shelter_lookup",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, shelterMessage(language), log, nil, inboundTranscript.ID, now)
	case "HELP", "112":
		conversation.Intent = "guidance_112"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp 112 guidance returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "guidance_112",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, guidance112Message(language), log, nil, inboundTranscript.ID, now)
	case "REPORT":
		report, ok, usage := parseWhatsAppDirectReport(request, phoneRef, linkedProfile, now)
		if ok {
			return s.completeWhatsAppReport(ctx, request, conversation, report, inboundTranscript.ID, now)
		}
		if len(strings.Fields(request.Body)) > 1 {
			utils.LogWarn("whatsapp report rejected invalid direct usage", "conversationId", conversation.ID, "phoneRef", phoneRef, "command", command)
			conversation.Intent = "invalid_selection"
			conversation.State = "idle"
			conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
			conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
			conversation.UpdatedAt = now
			conversation = s.store.UpdateWhatsAppConversation(conversation)
			log := s.store.CreateAccessLog(models.InclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
				LinkedProfile:     linkedProfile,
				Language:          language,
				Intent:            "invalid_selection",
				Status:            "handled",
				CreatedAt:         now,
			})
			return s.whatsappResponse(request, conversation, usage, log, nil, inboundTranscript.ID, now)
		}
		conversation.Intent = "report_emergency"
		conversation.State = "awaiting_report_hazard"
		conversation.Hazard = ""
		conversation.Urgency = ""
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp report flow started", "conversationId", conversation.ID, "phoneRef", phoneRef)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHazardPrompt(), log, nil, inboundTranscript.ID, now)
	default:
		if request.Location != nil || len(request.Media) > 0 {
			utils.LogWarn("whatsapp location or media received without active report", "conversationId", conversation.ID, "phoneRef", phoneRef, "mediaCount", len(request.Media), "hasLocation", request.Location != nil)
		} else {
			utils.LogWarn("whatsapp inbound unknown command", "conversationId", conversation.ID, "phoneRef", phoneRef, "command", command)
		}
		conversation.Intent = "invalid_selection"
		conversation.State = "idle"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "invalid_selection",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHelpMessage(), log, nil, inboundTranscript.ID, now)
	}
}

func (s *Server) handleWhatsAppReportConversation(ctx context.Context, request models.WhatsAppInboundRequest, conversation models.WhatsAppConversation, inboundTranscriptID string, now time.Time) models.WhatsAppInboundResponse {
	phoneRef := utils.PhoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language

	switch conversation.State {
	case "awaiting_report_hazard":
		hazard := utils.NormalizeSMSHazard(utils.FirstToken(request.Body))
		if !allowedHazards[hazard] {
			utils.LogWarn("whatsapp report rejected invalid hazard", "conversationId", conversation.ID, "phoneRef", phoneRef, "state", conversation.State)
			log := s.store.CreateAccessLog(models.InclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
				LinkedProfile:     linkedProfile,
				Language:          language,
				Intent:            "invalid_selection",
				Status:            "handled",
				CreatedAt:         now,
			})
			return s.whatsappResponse(request, conversation, whatsappHazardPrompt(), log, nil, inboundTranscriptID, now)
		}
		conversation.Intent = "report_emergency"
		conversation.State = "awaiting_report_urgency"
		conversation.Hazard = hazard
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp report hazard captured", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", hazard)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappUrgencyPrompt(), log, nil, inboundTranscriptID, now)
	case "awaiting_report_urgency":
		fields := strings.Fields(request.Body)
		urgency := utils.NormalizeSMSUrgency(utils.FirstToken(request.Body))
		if urgency == "" {
			utils.LogWarn("whatsapp report rejected invalid urgency", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", conversation.Hazard)
			log := s.store.CreateAccessLog(models.InclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
				LinkedProfile:     linkedProfile,
				Language:          language,
				Intent:            "invalid_selection",
				Status:            "handled",
				CreatedAt:         now,
			})
			return s.whatsappResponse(request, conversation, whatsappUrgencyPrompt(), log, nil, inboundTranscriptID, now)
		}
		conversation.Intent = "report_emergency"
		conversation.Urgency = urgency
		details := strings.TrimSpace(strings.Join(fields[1:], " "))
		if details != "" || request.Location != nil || len(request.Media) > 0 {
			report := buildWhatsAppReport(request, conversation, phoneRef, linkedProfile, now, details)
			return s.completeWhatsAppReport(ctx, request, conversation, report, inboundTranscriptID, now)
		}
		conversation.State = "awaiting_report_location"
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		utils.LogInfo("whatsapp report urgency captured", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", conversation.Hazard, "urgency", urgency)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappLocationPrompt(), log, nil, inboundTranscriptID, now)
	case "awaiting_report_location":
		if strings.TrimSpace(request.Body) == "" && request.Location == nil && len(request.Media) == 0 {
			utils.LogWarn("whatsapp report still missing location details", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", conversation.Hazard, "urgency", conversation.Urgency)
			log := s.store.CreateAccessLog(models.InclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
				LinkedProfile:     linkedProfile,
				Language:          language,
				Intent:            "invalid_selection",
				Status:            "handled",
				CreatedAt:         now,
			})
			return s.whatsappResponse(request, conversation, whatsappLocationPrompt(), log, nil, inboundTranscriptID, now)
		}
		report := buildWhatsAppReport(request, conversation, phoneRef, linkedProfile, now, request.Body)
		return s.completeWhatsAppReport(ctx, request, conversation, report, inboundTranscriptID, now)
	default:
		utils.LogWarn("whatsapp conversation state unknown", "conversationId", conversation.ID, "phoneRef", phoneRef, "state", conversation.State)
		conversation.Intent = "invalid_selection"
		conversation.State = "idle"
		conversation.UpdatedAt = now
		conversation = s.store.UpdateWhatsAppConversation(conversation)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "invalid_selection",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHelpMessage(), log, nil, inboundTranscriptID, now)
	}
}

func (s *Server) completeWhatsAppReport(ctx context.Context, request models.WhatsAppInboundRequest, conversation models.WhatsAppConversation, report models.InclusiveAccessReport, inboundTranscriptID string, now time.Time) models.WhatsAppInboundResponse {
	utils.LogInfo(
		"whatsapp report creating access report",
		"conversationId", conversation.ID,
		"phoneRef", report.PhoneRef,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"hasCoordinates", request.Location != nil,
		"mediaCount", len(report.Media),
		"locationLabel", utils.LogTextSummary(report.LocationLabel),
		"linkedProfile", report.LinkedProfile,
	)
	report = s.store.CreateAccessReport(report)
	report = s.submitInclusiveReport(ctx, report, request.From, request.ProfileID, report.LinkedProfile)
	conversation.Intent = "report_emergency"
	conversation.State = "idle"
	conversation.Hazard = ""
	conversation.Urgency = ""
	conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
	conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
	conversation.UpdatedAt = now
	conversation = s.store.UpdateWhatsAppConversation(conversation)
	utils.LogInfo(
		"whatsapp report flow completed",
		"conversationId", conversation.ID,
		"phoneRef", report.PhoneRef,
		"reportId", report.ID,
		"status", report.Status,
		"incidentId", report.IncidentID,
		"incidentReference", report.IncidentReference,
	)

	log := s.store.CreateAccessLog(models.InclusiveAccessLog{
		Channel:           "whatsapp",
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		SessionID:         conversation.ID,
		PhoneRef:          report.PhoneRef,
		ProfileID:         report.ProfileID,
		LinkedProfile:     report.LinkedProfile,
		Language:          request.Language,
		Intent:            "report_emergency",
		Status:            report.Status,
		IncidentID:        report.IncidentID,
		IncidentReference: report.IncidentReference,
		CreatedAt:         now,
	})
	message := reportConfirmationMessage(request.Language, report)
	return s.whatsappResponse(request, conversation, message, log, &report, inboundTranscriptID, now)
}

func (s *Server) whatsappResponse(request models.WhatsAppInboundRequest, conversation models.WhatsAppConversation, message string, accessLog models.InclusiveAccessLog, report *models.InclusiveAccessReport, inboundTranscriptID string, now time.Time) models.WhatsAppInboundResponse {
	outboundTranscript := s.store.CreateWhatsAppTranscript(models.WhatsAppTranscript{
		ConversationID:    conversation.ID,
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		PhoneRef:          conversation.PhoneRef,
		ProfileID:         conversation.ProfileID,
		LinkedProfile:     conversation.LinkedProfile,
		Direction:         "outbound",
		Intent:            accessLog.Intent,
		State:             conversation.State,
		MessageSummary:    utils.WhatsAppMessageSummary(message),
		MediaSummary:      "",
		CreatedAt:         now,
		RetentionUntil:    conversation.RetentionUntil,
	})
	transcriptIDs := []string{outboundTranscript.ID}
	if inboundTranscriptID != "" {
		transcriptIDs = append([]string{inboundTranscriptID}, transcriptIDs...)
	}
	return models.WhatsAppInboundResponse{
		Message:       message,
		Conversation:  conversation,
		Log:           accessLog,
		Report:        report,
		TranscriptIDs: transcriptIDs,
	}
}

func parseWhatsAppDirectReport(request models.WhatsAppInboundRequest, phoneRef string, linkedProfile bool, now time.Time) (models.InclusiveAccessReport, bool, string) {
	fields := strings.Fields(request.Body)
	if len(fields) < 3 {
		return models.InclusiveAccessReport{}, false, whatsappReportUsage()
	}

	hazard := utils.NormalizeSMSHazard(fields[1])
	if !allowedHazards[hazard] {
		return models.InclusiveAccessReport{}, false, whatsappReportUsage()
	}

	urgency := utils.NormalizeSMSUrgency(fields[2])
	if urgency == "" {
		return models.InclusiveAccessReport{}, false, whatsappReportUsage()
	}

	conversation := models.WhatsAppConversation{
		Channel:       "whatsapp",
		PhoneRef:      phoneRef,
		ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile: linkedProfile,
		Language:      request.Language,
		Intent:        "report_emergency",
		State:         "idle",
		Hazard:        hazard,
		Urgency:       urgency,
	}
	report := buildWhatsAppReport(request, conversation, phoneRef, linkedProfile, now, strings.Join(fields[3:], " "))
	return report, true, ""
}

func buildWhatsAppReport(request models.WhatsAppInboundRequest, conversation models.WhatsAppConversation, phoneRef string, linkedProfile bool, now time.Time, details string) models.InclusiveAccessReport {
	details = strings.TrimSpace(details)
	location, locationLabel := utils.InclusiveLocation(request.Location, strings.Fields(details))
	mediaRefs := utils.WhatsAppMediaRefs(request.Media)
	description := fmt.Sprintf("WhatsApp emergency report: %s with %s urgency. Location note: %s.", conversation.Hazard, conversation.Urgency, locationLabel)
	if details != "" {
		description = fmt.Sprintf("WhatsApp emergency report: %s with %s urgency. Details: %s.", conversation.Hazard, conversation.Urgency, details)
	}
	if len(mediaRefs) > 0 {
		description = fmt.Sprintf("%s Media attachments received: %d.", description, len(mediaRefs))
	}
	return models.InclusiveAccessReport{
		Channel:       "whatsapp",
		Type:          conversation.Hazard,
		Urgency:       conversation.Urgency,
		Description:   description,
		Location:      location,
		LocationLabel: locationLabel,
		PhoneRef:      phoneRef,
		ProfileID:     utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile: linkedProfile,
		Status:        "queued",
		Media:         mediaRefs,
		CreatedAt:     now,
	}
}
