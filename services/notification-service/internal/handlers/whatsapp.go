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
	if !s.requireWebhookSecret(w, r, "whatsapp") {
		return
	}

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
	request.ProviderMessageID = utils.NormalizeID(request.ProviderMessageID)
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

// whatsappDecision captures the outcome of the atomic conversation transition
// so transcripts, access logs, and the incident handoff can run after the store
// lock is released. report is set when the transition completed a report; its
// status and confirmation message are resolved after the handoff.
type whatsappDecision struct {
	preState string
	intent   string
	status   string
	message  string
	report   *models.InclusiveAccessReport
}

func (s *Server) handleWhatsAppInbound(ctx context.Context, request models.WhatsAppInboundRequest) models.WhatsAppInboundResponse {
	now := s.now()
	phoneRef := utils.PhoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language
	command := utils.SMSCommandName(request.Body)
	conversationKey := utils.WhatsAppConversationKey(request.From, request.ProfileID, linkedProfile)

	// The alert feed is fetched outside the store lock for the commands that
	// need it; it never depends on conversation state.
	var alerts []models.CitizenAlert
	if command == "ALERT" || command == "ALERTS" || command == "RISK" {
		alerts, _ = s.listCitizenAlerts(ctx, models.AlertFeedFilters{}, now)
	}

	seed := models.WhatsAppConversation{
		Key:            conversationKey,
		Channel:        "whatsapp",
		PhoneRef:       phoneRef,
		ProfileID:      utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:  linkedProfile,
		Language:       language,
		Intent:         "main_menu",
		State:          "idle",
		StartedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
		RetentionUntil: utils.WhatsAppRetentionUntil(now),
	}
	decision := whatsappDecision{}
	conversation := s.store.ProcessWhatsAppConversation(conversationKey, seed, now, func(conversation *models.WhatsAppConversation) {
		decision.preState = conversation.State
		transitionWhatsAppConversation(conversation, request, command, alerts, now, &decision)
	})
	utils.LogInfo(
		"whatsapp inbound handling started",
		"conversationId", conversation.ID,
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"command", command,
		"state", decision.preState,
		"language", language,
		"linkedProfile", linkedProfile,
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
		State:             decision.preState,
		MessageSummary:    utils.WhatsAppMessageSummary(request.Body),
		MediaSummary:      utils.WhatsAppMediaSummary(request.Media),
		CreatedAt:         now,
		RetentionUntil:    utils.WhatsAppRetentionUntil(now),
	})

	if decision.report != nil {
		report, created := s.store.CreateAccessReport(*decision.report)
		if created {
			utils.LogInfo(
				"whatsapp report creating access report",
				"conversationId", conversation.ID,
				"phoneRef", report.PhoneRef,
				"hazard", report.Type,
				"urgency", report.Urgency,
				"hasCoordinates", report.Location != nil,
				"locationLabel", utils.LogTextSummary(report.LocationLabel),
				"linkedProfile", report.LinkedProfile,
			)
			report = s.submitInclusiveReport(ctx, report, request.From, request.ProfileID, report.LinkedProfile)
		} else {
			utils.LogInfo(
				"whatsapp report webhook deduplicated",
				"conversationId", conversation.ID,
				"phoneRef", report.PhoneRef,
				"reportId", report.ID,
				"providerMessageId", request.ProviderMessageID,
			)
		}
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
		return s.whatsappResponse(request, conversation, message, log, &report, inboundTranscript.ID, now)
	}

	providerError := ""
	if decision.intent == "provider_error" {
		providerError = request.ProviderError
	}
	log := s.store.CreateAccessLog(models.InclusiveAccessLog{
		Channel:           "whatsapp",
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		SessionID:         conversation.ID,
		PhoneRef:          phoneRef,
		ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Language:          language,
		Intent:            decision.intent,
		Status:            decision.status,
		ProviderError:     providerError,
		CreatedAt:         now,
	})
	return s.whatsappResponse(request, conversation, decision.message, log, nil, inboundTranscript.ID, now)
}

// transitionWhatsAppConversation is the pure WhatsApp state machine. It runs
// inside the store lock (see ProcessWhatsAppConversation), so it must only
// mutate the conversation and the decision — no store calls, no network I/O.
func transitionWhatsAppConversation(conversation *models.WhatsAppConversation, request models.WhatsAppInboundRequest, command string, alerts []models.CitizenAlert, now time.Time, decision *whatsappDecision) {
	language := request.Language
	setSummaries := func() {
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
	}
	resetIdle := func(intent string) {
		conversation.Intent = intent
		conversation.State = "idle"
		conversation.Hazard = ""
		conversation.Urgency = ""
	}

	if request.ProviderError != "" {
		resetIdle("provider_error")
		setSummaries()
		decision.intent = "provider_error"
		decision.status = "failed"
		decision.message = localizedMessage(language, "provider_error")
		return
	}

	if command == "CANCEL" || command == "MENU" || command == "START" || command == "HI" || command == "HELLO" {
		resetIdle("main_menu")
		setSummaries()
		decision.intent = "main_menu"
		decision.status = "handled"
		decision.message = whatsappHelpMessage()
		return
	}

	if conversation.State != "" && conversation.State != "idle" && !utils.IsWhatsAppTopLevelCommand(command) {
		transitionWhatsAppReport(conversation, request, now, decision)
		return
	}

	switch command {
	case "ALERT", "ALERTS":
		resetIdle("current_alerts")
		setSummaries()
		decision.intent = "current_alerts"
		decision.status = "handled"
		decision.message = alertSummaryMessage(language, alerts)
	case "RISK":
		resetIdle("risk_check")
		setSummaries()
		decision.intent = "risk_check"
		decision.status = "handled"
		decision.message = riskCheckMessage(language, request.Location != nil, alerts)
	case "GUIDE", "GUIDES":
		hazard := utils.WhatsAppCommandArg(request.Body, 1)
		if hazard == "" {
			hazard = "flood"
		}
		hazard = utils.NormalizeSMSHazard(hazard)
		if !allowedHazards[hazard] {
			hazard = "flood"
		}
		resetIdle("emergency_guides")
		setSummaries()
		decision.intent = "emergency_guides"
		decision.status = "handled"
		decision.message = emergencyGuideMessage(language, hazard)
	case "SHELTER", "SHELTERS":
		resetIdle("shelter_lookup")
		setSummaries()
		decision.intent = "shelter_lookup"
		decision.status = "handled"
		decision.message = shelterMessage(language)
	case "HELP", "112":
		resetIdle("guidance_112")
		setSummaries()
		decision.intent = "guidance_112"
		decision.status = "handled"
		decision.message = guidance112Message(language)
	case "REPORT":
		report, ok, usage := parseWhatsAppDirectReport(request, conversation.PhoneRef, conversation.LinkedProfile, now)
		if ok {
			resetIdle("report_emergency")
			setSummaries()
			decision.intent = "report_emergency"
			decision.report = &report
			return
		}
		if len(strings.Fields(request.Body)) > 1 {
			resetIdle("invalid_selection")
			setSummaries()
			decision.intent = "invalid_selection"
			decision.status = "handled"
			decision.message = usage
			return
		}
		conversation.Intent = "report_emergency"
		conversation.State = "awaiting_report_hazard"
		conversation.Hazard = ""
		conversation.Urgency = ""
		setSummaries()
		decision.intent = "report_emergency"
		decision.status = "handled"
		decision.message = whatsappHazardPrompt()
	default:
		resetIdle("invalid_selection")
		setSummaries()
		decision.intent = "invalid_selection"
		decision.status = "handled"
		decision.message = whatsappHelpMessage()
	}
}

// transitionWhatsAppReport advances an in-progress report conversation. Like
// transitionWhatsAppConversation it runs inside the store lock and must stay
// free of store calls and network I/O.
func transitionWhatsAppReport(conversation *models.WhatsAppConversation, request models.WhatsAppInboundRequest, now time.Time, decision *whatsappDecision) {
	setSummaries := func() {
		conversation.LastMessageSummary = utils.WhatsAppMessageSummary(request.Body)
		conversation.LastMediaSummary = utils.WhatsAppMediaSummary(request.Media)
	}
	completeReport := func(details string) {
		report := buildWhatsAppReport(request, *conversation, conversation.PhoneRef, conversation.LinkedProfile, now, details)
		conversation.Intent = "report_emergency"
		conversation.State = "idle"
		conversation.Hazard = ""
		conversation.Urgency = ""
		setSummaries()
		decision.intent = "report_emergency"
		decision.report = &report
	}

	switch conversation.State {
	case "awaiting_report_hazard":
		hazard := utils.NormalizeSMSHazard(utils.FirstToken(request.Body))
		if !allowedHazards[hazard] {
			decision.intent = "invalid_selection"
			decision.status = "handled"
			decision.message = whatsappHazardPrompt()
			return
		}
		conversation.Intent = "report_emergency"
		conversation.State = "awaiting_report_urgency"
		conversation.Hazard = hazard
		setSummaries()
		decision.intent = "report_emergency"
		decision.status = "handled"
		decision.message = whatsappUrgencyPrompt()
	case "awaiting_report_urgency":
		fields := strings.Fields(request.Body)
		urgency := utils.NormalizeSMSUrgency(utils.FirstToken(request.Body))
		if urgency == "" {
			decision.intent = "invalid_selection"
			decision.status = "handled"
			decision.message = whatsappUrgencyPrompt()
			return
		}
		conversation.Intent = "report_emergency"
		conversation.Urgency = urgency
		details := strings.TrimSpace(strings.Join(fields[1:], " "))
		if details != "" || request.Location != nil || len(request.Media) > 0 {
			completeReport(details)
			return
		}
		conversation.State = "awaiting_report_location"
		setSummaries()
		decision.intent = "report_emergency"
		decision.status = "handled"
		decision.message = whatsappLocationPrompt()
	case "awaiting_report_location":
		if strings.TrimSpace(request.Body) == "" && request.Location == nil && len(request.Media) == 0 {
			decision.intent = "invalid_selection"
			decision.status = "handled"
			decision.message = whatsappLocationPrompt()
			return
		}
		completeReport(request.Body)
	default:
		conversation.Intent = "invalid_selection"
		conversation.State = "idle"
		decision.intent = "invalid_selection"
		decision.status = "handled"
		decision.message = whatsappHelpMessage()
	}
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
		description = fmt.Sprintf("%s Media attachments received: %d (%s).", description, len(mediaRefs), mediaRefsText(mediaRefs))
	}
	// Media refs are folded into the description as text instead of the Media
	// field: incident-service only accepts media registered through its own
	// upload endpoint and rejects anything else with a 400.
	return models.InclusiveAccessReport{
		Channel:           "whatsapp",
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		Type:              conversation.Hazard,
		Urgency:           conversation.Urgency,
		Description:       description,
		Location:          location,
		LocationLabel:     locationLabel,
		PhoneRef:          phoneRef,
		ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Status:            "queued",
		CreatedAt:         now,
	}
}

// mediaRefsText renders unregistered media references as bounded description
// text so long provider URLs cannot push the description past incident-service
// validation limits.
func mediaRefsText(refs []string) string {
	text := strings.Join(refs, ", ")
	if len(text) > 500 {
		text = text[:500] + "..."
	}
	return text
}
