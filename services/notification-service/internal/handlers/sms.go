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

func (s *Server) smsInboundHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireWebhookSecret(w, r, "sms") {
		return
	}

	var request models.SMSInboundRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.LogWarn("sms inbound rejected", "code", "invalid_json", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.From = strings.TrimSpace(request.From)
	request.Body = strings.TrimSpace(request.Body)
	request.Language = utils.NormalizeLanguage(request.Language)
	request.Provider = utils.ProviderOrDefault(request.Provider, "sms_sandbox")
	request.ProfileID = utils.NormalizeID(request.ProfileID)
	request.ProviderError = strings.TrimSpace(request.ProviderError)
	request.ProviderMessageID = utils.NormalizeID(request.ProviderMessageID)

	utils.LogInfo(
		"sms inbound received",
		"provider", request.Provider,
		"phoneRef", utils.PhoneRef(request.From),
		"command", utils.SMSCommandName(request.Body),
		"hasProviderError", request.ProviderError != "",
		"linkedProfileRequested", request.LinkProfile,
	)
	if request.From == "" {
		utils.LogWarn("sms inbound rejected", "code", "missing_from", "provider", request.Provider)
		utils.WriteError(w, http.StatusBadRequest, "missing_from", "from is required")
		return
	}
	if request.Body == "" && request.ProviderError == "" {
		utils.LogWarn("sms inbound rejected", "code", "missing_body", "provider", request.Provider, "phoneRef", utils.PhoneRef(request.From))
		utils.WriteError(w, http.StatusBadRequest, "missing_body", "body is required")
		return
	}

	response := s.handleSMSInbound(r.Context(), request)
	utils.WriteJSON(w, http.StatusAccepted, response)
}

func (s *Server) handleSMSInbound(ctx context.Context, request models.SMSInboundRequest) models.SMSInboundResponse {
	now := s.now()
	phoneRef := utils.PhoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language
	utils.LogInfo(
		"sms inbound handling started",
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"command", utils.SMSCommandName(request.Body),
		"language", language,
		"linkedProfile", linkedProfile,
	)

	if request.ProviderError != "" {
		utils.LogWarn(
			"sms provider error received",
			"provider", request.Provider,
			"providerMessageId", request.ProviderMessageID,
			"phoneRef", phoneRef,
			"errorLength", len(request.ProviderError),
		)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "provider_error",
			Status:            "failed",
			ProviderError:     request.ProviderError,
			CreatedAt:         now,
		})
		return models.SMSInboundResponse{Message: localizedMessage(language, "provider_error"), Log: log}
	}

	command := strings.TrimSpace(request.Body)
	upperCommand := strings.ToUpper(command)
	switch {
	case upperCommand == "ALERT" || upperCommand == "ALERTS":
		alerts, _ := s.listCitizenAlerts(ctx, models.AlertFeedFilters{}, now)
		utils.LogInfo("sms alert summary returned", "provider", request.Provider, "phoneRef", phoneRef, "alertCount", len(alerts))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "current_alerts",
			Status:            "handled",
			CreatedAt:         now,
		})
		return models.SMSInboundResponse{Message: smsAlertMessage(alerts), Log: log}
	case upperCommand == "SHELTER" || upperCommand == "SHELTERS":
		utils.LogInfo("sms shelter guidance returned", "provider", request.Provider, "phoneRef", phoneRef, "language", language)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "shelter_lookup",
			Status:            "handled",
			CreatedAt:         now,
		})
		return models.SMSInboundResponse{Message: shelterMessage(language), Log: log}
	case upperCommand == "HELP" || upperCommand == "112":
		utils.LogInfo("sms 112 guidance returned", "provider", request.Provider, "phoneRef", phoneRef, "language", language, "command", utils.SMSCommandName(request.Body))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "guidance_112",
			Status:            "handled",
			CreatedAt:         now,
		})
		return models.SMSInboundResponse{Message: guidance112Message(language), Log: log}
	case strings.HasPrefix(upperCommand, "REPORT "):
		report, ok, usage := parseSMSReport(request, phoneRef, linkedProfile, now)
		if !ok {
			utils.LogWarn("sms report rejected invalid usage", "provider", request.Provider, "phoneRef", phoneRef, "command", utils.SMSCommandName(request.Body))
			log := s.store.CreateAccessLog(models.InclusiveAccessLog{
				Channel:           "sms",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				PhoneRef:          phoneRef,
				ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
				LinkedProfile:     linkedProfile,
				Language:          language,
				Intent:            "invalid_selection",
				Status:            "handled",
				CreatedAt:         now,
			})
			return models.SMSInboundResponse{Message: usage, Log: log}
		}
		utils.LogInfo(
			"sms report creating access report",
			"provider", request.Provider,
			"phoneRef", phoneRef,
			"hazard", report.Type,
			"urgency", report.Urgency,
			"hasCoordinates", report.Location != nil,
			"locationLabel", utils.LogTextSummary(report.LocationLabel),
			"linkedProfile", linkedProfile,
		)
		report, created := s.store.CreateAccessReport(report)
		if created {
			report = s.submitInclusiveReport(ctx, report, request.From, request.ProfileID, linkedProfile)
		} else {
			utils.LogInfo(
				"sms report webhook deduplicated",
				"provider", request.Provider,
				"phoneRef", phoneRef,
				"reportId", report.ID,
				"providerMessageId", request.ProviderMessageID,
			)
		}
		utils.LogInfo(
			"sms report flow completed",
			"provider", request.Provider,
			"phoneRef", phoneRef,
			"reportId", report.ID,
			"status", report.Status,
			"incidentId", report.IncidentID,
			"incidentReference", report.IncidentReference,
		)
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
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
		return models.SMSInboundResponse{Message: reportConfirmationMessage(language, report), Log: log, Report: &report}
	default:
		utils.LogWarn("sms inbound unknown command", "provider", request.Provider, "phoneRef", phoneRef, "command", utils.SMSCommandName(request.Body))
		log := s.store.CreateAccessLog(models.InclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "invalid_selection",
			Status:            "handled",
			CreatedAt:         now,
		})
		return models.SMSInboundResponse{Message: smsHelpMessage(), Log: log}
	}
}

func parseSMSReport(request models.SMSInboundRequest, phoneRef string, linkedProfile bool, now time.Time) (models.InclusiveAccessReport, bool, string) {
	fields := strings.Fields(request.Body)
	if len(fields) < 3 {
		return models.InclusiveAccessReport{}, false, smsReportUsage()
	}

	hazard := utils.NormalizeSMSHazard(fields[1])
	if !allowedHazards[hazard] {
		return models.InclusiveAccessReport{}, false, smsReportUsage()
	}

	urgency := utils.NormalizeSMSUrgency(fields[2])
	if urgency == "" {
		return models.InclusiveAccessReport{}, false, smsReportUsage()
	}

	description := strings.TrimSpace(strings.Join(fields[3:], " "))
	if description == "" {
		description = fmt.Sprintf("SMS emergency report: %s with %s urgency", hazard, urgency)
	} else {
		description = "SMS emergency report: " + description
	}

	location, locationLabel := utils.InclusiveLocation(request.Location, nil)
	return models.InclusiveAccessReport{
		Channel:           "sms",
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		Type:              hazard,
		Urgency:           urgency,
		Description:       description,
		Location:          location,
		LocationLabel:     locationLabel,
		PhoneRef:          phoneRef,
		ProfileID:         utils.ProfileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Status:            "queued",
		CreatedAt:         now,
	}, true, ""
}
