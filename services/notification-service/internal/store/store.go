package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// Store is the persistence interface for notification data.
type Store interface {
	ListAlerts(filters models.AlertFeedFilters, now time.Time) []models.CitizenAlert
	CreateDeliveryAttempts(ctx context.Context, alert models.CitizenAlert, request models.DeliveryRequest, providers map[string]models.NotificationProvider, now time.Time) []models.DeliveryAttempt
	ListDeliveryLogs(filters models.LogFilters) []models.DeliveryAttempt
	CreateVoiceAlertAsset(alert models.CitizenAlert, languages []string, source string, requestedBy string, now time.Time) models.VoiceAlertAsset
	ListVoiceAlertAssets() []models.VoiceAlertAsset
	GetVoiceAlertAsset(id string) (models.VoiceAlertAsset, bool)
	ReviewVoiceAlertAsset(id string, action string, reviewer string, note string, languages []string, now time.Time) (models.VoiceAlertAsset, bool)
	CreateVoiceDeliveryAttempts(ctx context.Context, asset models.VoiceAlertAsset, request models.VoiceDeliveryRequest, providers map[string]models.NotificationProvider, now time.Time) []models.DeliveryAttempt
	CreateCellBroadcastMessage(alert models.CitizenAlert, languages []string, areas []string, requestedBy string, now time.Time) models.CellBroadcastMessage
	ListCellBroadcastMessages() []models.CellBroadcastMessage
	GetCellBroadcastMessage(id string) (models.CellBroadcastMessage, bool)
	ReviewCellBroadcastMessage(id string, action string, reviewer string, note string, languages []string, now time.Time) (models.CellBroadcastMessage, bool)
	CreateCellBroadcastDispatches(ctx context.Context, message models.CellBroadcastMessage, request models.CellBroadcastDeliveryRequest, adapter models.CellBroadcastAdapter, now time.Time) []models.CellBroadcastDispatch
	CreateAccessLog(log models.InclusiveAccessLog) models.InclusiveAccessLog
	ListAccessLogs(filters models.AccessLogFilters) []models.InclusiveAccessLog
	CreateAccessReport(report models.InclusiveAccessReport) models.InclusiveAccessReport
	UpdateAccessReport(report models.InclusiveAccessReport)
	GetOrCreateWhatsAppConversation(key string, phoneRef string, profileID string, linkedProfile bool, language string, now time.Time) models.WhatsAppConversation
	UpdateWhatsAppConversation(conversation models.WhatsAppConversation) models.WhatsAppConversation
	CreateWhatsAppTranscript(transcript models.WhatsAppTranscript) models.WhatsAppTranscript
	WhatsAppTranscripts() []models.WhatsAppTranscript
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu                         sync.RWMutex
	alerts                     []models.CitizenAlert
	deliveryLogs               []models.DeliveryAttempt
	accessLogs                 []models.InclusiveAccessLog
	accessReports              []models.InclusiveAccessReport
	voiceAlerts                []models.VoiceAlertAsset
	cellBroadcasts             []models.CellBroadcastMessage
	whatsappConversations      map[string]models.WhatsAppConversation
	whatsappTranscripts        []models.WhatsAppTranscript
	nextLogID                  int
	nextAccessLogID            int
	nextAccessReportID         int
	nextVoiceAlertID           int
	nextVoiceVariantID         int
	nextCellBroadcastID        int
	nextCellBroadcastSegmentID int
	nextCellBroadcastSerial    int
	nextWhatsAppConversationID int
	nextWhatsAppTranscriptID   int
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	return &MemoryStore{
		alerts:                     seedCitizenAlerts(now),
		whatsappConversations:      map[string]models.WhatsAppConversation{},
		nextLogID:                  1,
		nextAccessLogID:            1,
		nextAccessReportID:         1,
		nextVoiceAlertID:           1,
		nextVoiceVariantID:         1,
		nextCellBroadcastID:        1,
		nextCellBroadcastSegmentID: 1,
		nextCellBroadcastSerial:    1,
		nextWhatsAppConversationID: 1,
		nextWhatsAppTranscriptID:   1,
	}
}

// ListAlerts returns alerts matching the provided filters.
func (m *MemoryStore) ListAlerts(filters models.AlertFeedFilters, now time.Time) []models.CitizenAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]models.CitizenAlert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alert.Status = alertFeedStatus(alert.StartsAt, alert.ExpiresAt, now)
		if alertMatchesFilters(alert, filters, now) {
			alerts = append(alerts, alert)
		}
	}
	sortCitizenAlerts(alerts)
	return alerts
}

// CreateDeliveryAttempts creates and persists delivery attempts for an alert.
func (m *MemoryStore) CreateDeliveryAttempts(ctx context.Context, alert models.CitizenAlert, request models.DeliveryRequest, providers map[string]models.NotificationProvider, now time.Time) []models.DeliveryAttempt {
	m.mu.Lock()
	defer m.mu.Unlock()

	attempts := make([]models.DeliveryAttempt, 0, len(request.Channels))
	for _, channel := range request.Channels {
		provider := providers[channel]
		if provider == nil {
			utils.LogError("notification provider missing", "alertId", alert.ID, "channel", channel)
			provider = models.DisabledProvider{Channel: channel, Reason: "provider missing"}
		}
		utils.LogInfo(
			"notification provider send starting",
			"alertId", alert.ID,
			"channel", channel,
			"provider", utils.ProviderName(provider),
			"recipientRef", utils.RecipientRef(request, channel),
			"dryRun", request.DryRun,
		)
		result := provider.Send(ctx, models.ProviderMessage{
			Alert:       alert,
			Request:     request,
			Channel:     channel,
			Recipient:   utils.RecipientRef(request, channel),
			AttemptedAt: now,
		})
		attempt := models.DeliveryAttempt{
			ID:           fmt.Sprintf("delivery_%06d", m.nextLogID),
			AlertID:      alert.ID,
			AlertTitle:   alert.Title,
			Channel:      channel,
			Provider:     result.Provider,
			RecipientRef: utils.RecipientRef(request, channel),
			Status:       result.Status,
			Reason:       result.Reason,
			MessageID:    result.MessageID,
			AttemptedAt:  now,
		}
		m.nextLogID++
		m.deliveryLogs = append(m.deliveryLogs, attempt)
		attempts = append(attempts, attempt)
		utils.LogInfo(
			"delivery attempt stored",
			"attemptId", attempt.ID,
			"alertId", attempt.AlertID,
			"channel", attempt.Channel,
			"provider", attempt.Provider,
			"status", attempt.Status,
			"reason", attempt.Reason,
		)
	}

	return attempts
}

// CreateVoiceAlertAsset creates a new multi-language voice alert asset.
func (m *MemoryStore) CreateVoiceAlertAsset(alert models.CitizenAlert, languages []string, source string, requestedBy string, now time.Time) models.VoiceAlertAsset {
	m.mu.Lock()
	defer m.mu.Unlock()

	asset := models.VoiceAlertAsset{
		ID:                  fmt.Sprintf("voice_alert_%06d", m.nextVoiceAlertID),
		AlertID:             alert.ID,
		AlertTitle:          alert.Title,
		HazardType:          alert.HazardType,
		Severity:            alert.Severity,
		TargetLabel:         alert.TargetLabel,
		Status:              "generated",
		ReviewStatus:        "pending_review",
		Source:              source,
		WorkflowRequestedBy: requestedBy,
		Variants:            make([]models.VoiceVariant, 0, len(languages)),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if asset.TargetLabel == "" {
		asset.TargetLabel = alert.Target.Label
	}
	m.nextVoiceAlertID++

	for _, language := range languages {
		messageText := voiceMessageForAlert(language, alert)
		variant := models.VoiceVariant{
			ID:                  fmt.Sprintf("voice_variant_%06d", m.nextVoiceVariantID),
			Language:            language,
			Locale:              voiceLocale(language),
			VoiceName:           voiceName(language),
			MessageText:         messageText,
			AudioURL:            fmt.Sprintf("voice://%s/%s/%s.mp3", source, alert.ID, language),
			DurationSeconds:     estimateVoiceDurationSeconds(messageText),
			Status:              "generated",
			ReviewStatus:        "pending_review",
			AccessibilityChecks: voiceAccessibilityChecks(messageText, alert),
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		m.nextVoiceVariantID++
		asset.Variants = append(asset.Variants, variant)
	}

	m.voiceAlerts = append(m.voiceAlerts, asset)
	utils.LogInfo(
		"voice alert asset stored",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"variantCount", len(asset.Variants),
		"source", asset.Source,
		"requestedBy", asset.WorkflowRequestedBy,
	)
	return copyVoiceAlertAsset(asset)
}

// ListVoiceAlertAssets returns all stored voice alert assets sorted by creation time.
func (m *MemoryStore) ListVoiceAlertAssets() []models.VoiceAlertAsset {
	m.mu.RLock()
	defer m.mu.RUnlock()

	assets := make([]models.VoiceAlertAsset, 0, len(m.voiceAlerts))
	for _, asset := range m.voiceAlerts {
		assets = append(assets, copyVoiceAlertAsset(asset))
	}
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].CreatedAt.After(assets[j].CreatedAt)
	})
	return assets
}

// GetVoiceAlertAsset returns a voice alert asset by id.
func (m *MemoryStore) GetVoiceAlertAsset(id string) (models.VoiceAlertAsset, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, asset := range m.voiceAlerts {
		if asset.ID == id {
			return copyVoiceAlertAsset(asset), true
		}
	}
	return models.VoiceAlertAsset{}, false
}

// ReviewVoiceAlertAsset updates the review status for a voice alert asset.
func (m *MemoryStore) ReviewVoiceAlertAsset(id string, action string, reviewer string, note string, languages []string, now time.Time) (models.VoiceAlertAsset, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index, asset := range m.voiceAlerts {
		if asset.ID != id {
			continue
		}

		selectedLanguages := map[string]bool{}
		for _, language := range languages {
			selectedLanguages[language] = true
		}
		reviewAll := len(selectedLanguages) == 0
		var reviewStatus string
		var reviewedCount int
		for variantIndex, variant := range asset.Variants {
			if !reviewAll && !selectedLanguages[variant.Language] {
				continue
			}
			asset.Variants[variantIndex].Status = voiceReviewStatus(action)
			asset.Variants[variantIndex].ReviewStatus = voiceReviewStatus(action)
			asset.Variants[variantIndex].UpdatedAt = now
			reviewedCount++
		}

		var approvedCount int
		var rejectedCount int
		for _, variant := range asset.Variants {
			switch variant.ReviewStatus {
			case "approved":
				approvedCount++
			case "rejected":
				rejectedCount++
			}
		}
		switch {
		case approvedCount == len(asset.Variants):
			reviewStatus = "approved"
		case rejectedCount == len(asset.Variants):
			reviewStatus = "rejected"
		default:
			reviewStatus = "partial_review"
		}
		if reviewedCount == 0 {
			reviewStatus = "pending_review"
		}

		asset.ReviewStatus = reviewStatus
		switch reviewStatus {
		case "approved", "rejected":
			asset.Status = reviewStatus
		default:
			asset.Status = "generated"
		}
		asset.Reviewer = reviewer
		asset.ReviewNote = note
		asset.UpdatedAt = now
		asset.ReviewedAt = &now
		m.voiceAlerts[index] = asset
		utils.LogInfo(
			"voice alert asset reviewed",
			"voiceAssetId", asset.ID,
			"alertId", asset.AlertID,
			"action", action,
			"reviewer", reviewer,
			"reviewedCount", reviewedCount,
			"status", asset.Status,
			"reviewStatus", asset.ReviewStatus,
		)
		return copyVoiceAlertAsset(asset), true
	}

	return models.VoiceAlertAsset{}, false
}

// CreateVoiceDeliveryAttempts creates and persists voice delivery attempts.
func (m *MemoryStore) CreateVoiceDeliveryAttempts(ctx context.Context, asset models.VoiceAlertAsset, request models.VoiceDeliveryRequest, providers map[string]models.NotificationProvider, now time.Time) []models.DeliveryAttempt {
	m.mu.Lock()
	defer m.mu.Unlock()

	attempts := make([]models.DeliveryAttempt, 0, len(request.Recipients))
	provider := providers["voice"]
	if provider == nil {
		utils.LogError("voice notification provider missing", "voiceAssetId", asset.ID, "alertId", asset.AlertID)
		provider = models.DisabledProvider{Channel: "voice", Reason: "voice provider missing"}
	}

	for _, recipient := range request.Recipients {
		variant, variantFound := voiceVariantForLanguage(asset, recipient.Language)
		result := models.ProviderResult{}
		switch {
		case !variantFound:
			result = models.ProviderResult{Provider: "voice_asset", Status: "skipped", Reason: "approved language variant is missing"}
			utils.LogWarn("voice delivery skipped missing variant", "voiceAssetId", asset.ID, "alertId", asset.AlertID, "language", recipient.Language, "recipientRef", utils.VoiceRecipientRef(recipient))
		case variant.ReviewStatus != "approved":
			result = models.ProviderResult{Provider: "voice_asset", Status: "skipped", Reason: "language variant is not approved"}
			utils.LogWarn("voice delivery skipped unapproved variant", "voiceAssetId", asset.ID, "alertId", asset.AlertID, "language", recipient.Language, "variantStatus", variant.ReviewStatus, "recipientRef", utils.VoiceRecipientRef(recipient))
		default:
			deliveryReq := models.DeliveryRequest{
				AlertID:     asset.AlertID,
				RecipientID: recipient.RecipientID,
				Phone:       recipient.Phone,
				Language:    recipient.Language,
				Channels:    []string{"voice"},
				DryRun:      request.DryRun,
			}
			utils.LogInfo(
				"voice provider send starting",
				"voiceAssetId", asset.ID,
				"alertId", asset.AlertID,
				"language", recipient.Language,
				"provider", utils.ProviderName(provider),
				"recipientRef", utils.VoiceRecipientRef(recipient),
				"dryRun", request.DryRun,
			)
			result = provider.Send(ctx, models.ProviderMessage{
				Alert: models.CitizenAlert{
					ID:          asset.AlertID,
					Title:       asset.AlertTitle,
					HazardType:  asset.HazardType,
					Severity:    asset.Severity,
					TargetLabel: asset.TargetLabel,
				},
				Request:     deliveryReq,
				Channel:     "voice",
				Recipient:   utils.VoiceRecipientRef(recipient),
				AttemptedAt: now,
			})
		}

		attempt := models.DeliveryAttempt{
			ID:           fmt.Sprintf("delivery_%06d", m.nextLogID),
			AlertID:      asset.AlertID,
			AlertTitle:   asset.AlertTitle,
			Channel:      "voice",
			Provider:     result.Provider,
			RecipientRef: utils.VoiceRecipientRef(recipient),
			Status:       result.Status,
			Reason:       result.Reason,
			MessageID:    result.MessageID,
			VoiceAssetID: asset.ID,
			Language:     recipient.Language,
			AttemptedAt:  now,
		}
		if variantFound {
			attempt.AudioURL = variant.AudioURL
		}
		m.nextLogID++
		m.deliveryLogs = append(m.deliveryLogs, attempt)
		attempts = append(attempts, attempt)
		utils.LogInfo(
			"voice delivery attempt stored",
			"attemptId", attempt.ID,
			"voiceAssetId", attempt.VoiceAssetID,
			"alertId", attempt.AlertID,
			"language", attempt.Language,
			"provider", attempt.Provider,
			"status", attempt.Status,
			"reason", attempt.Reason,
		)
	}

	return attempts
}

// CreateAccessLog persists an inclusive access log.
func (m *MemoryStore) CreateAccessLog(log models.InclusiveAccessLog) models.InclusiveAccessLog {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.ID = fmt.Sprintf("access_%06d", m.nextAccessLogID)
	m.nextAccessLogID++
	m.accessLogs = append(m.accessLogs, log)
	utils.LogInfo(
		"inclusive access log stored",
		"logId", log.ID,
		"channel", log.Channel,
		"intent", log.Intent,
		"status", log.Status,
		"provider", log.Provider,
		"phoneRef", log.PhoneRef,
		"linkedProfile", log.LinkedProfile,
	)
	return log
}

// ListAccessLogs returns inclusive access logs matching the filters.
func (m *MemoryStore) ListAccessLogs(filters models.AccessLogFilters) []models.InclusiveAccessLog {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := make([]models.InclusiveAccessLog, 0, len(m.accessLogs))
	for _, log := range m.accessLogs {
		if filters.Channel != "" && log.Channel != filters.Channel {
			continue
		}
		if filters.Intent != "" && log.Intent != filters.Intent {
			continue
		}
		if filters.Status != "" && log.Status != filters.Status {
			continue
		}
		logs = append(logs, log)
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	return logs
}

// CreateAccessReport persists an inclusive access report.
func (m *MemoryStore) CreateAccessReport(report models.InclusiveAccessReport) models.InclusiveAccessReport {
	m.mu.Lock()
	defer m.mu.Unlock()

	report.ID = fmt.Sprintf("access_report_%06d", m.nextAccessReportID)
	m.nextAccessReportID++
	m.accessReports = append(m.accessReports, report)
	utils.LogInfo(
		"inclusive access report stored",
		"reportId", report.ID,
		"channel", report.Channel,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"status", report.Status,
		"phoneRef", report.PhoneRef,
		"linkedProfile", report.LinkedProfile,
	)
	return report
}

// UpdateAccessReport updates an existing inclusive access report.
func (m *MemoryStore) UpdateAccessReport(report models.InclusiveAccessReport) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index, existing := range m.accessReports {
		if existing.ID == report.ID {
			m.accessReports[index] = report
			utils.LogInfo(
				"inclusive access report updated",
				"reportId", report.ID,
				"channel", report.Channel,
				"status", report.Status,
				"incidentId", report.IncidentID,
				"incidentReference", report.IncidentReference,
				"failureReason", utils.LogTextSummary(report.FailureReason),
			)
			return
		}
	}
	m.accessReports = append(m.accessReports, report)
	utils.LogWarn(
		"inclusive access report update appended missing report",
		"reportId", report.ID,
		"channel", report.Channel,
		"status", report.Status,
	)
}

// GetOrCreateWhatsAppConversation returns an existing conversation or creates one.
func (m *MemoryStore) GetOrCreateWhatsAppConversation(key string, phoneRef string, profileID string, linkedProfile bool, language string, now time.Time) models.WhatsAppConversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.whatsappConversations == nil {
		m.whatsappConversations = map[string]models.WhatsAppConversation{}
	}
	if conversation, ok := m.whatsappConversations[key]; ok {
		conversation.ProfileID = profileID
		conversation.LinkedProfile = linkedProfile
		conversation.Language = language
		conversation.UpdatedAt = now
		conversation.ExpiresAt = now.Add(24 * time.Hour)
		conversation.RetentionUntil = utils.WhatsAppRetentionUntil(now)
		m.whatsappConversations[key] = conversation
		utils.LogInfo(
			"whatsapp conversation resumed",
			"conversationId", conversation.ID,
			"phoneRef", conversation.PhoneRef,
			"state", conversation.State,
			"intent", conversation.Intent,
		)
		return conversation
	}

	conversation := models.WhatsAppConversation{
		ID:             fmt.Sprintf("whatsapp_%06d", m.nextWhatsAppConversationID),
		Key:            key,
		Channel:        "whatsapp",
		PhoneRef:       phoneRef,
		ProfileID:      profileID,
		LinkedProfile:  linkedProfile,
		Language:       language,
		Intent:         "main_menu",
		State:          "idle",
		StartedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(24 * time.Hour),
		RetentionUntil: utils.WhatsAppRetentionUntil(now),
	}
	m.nextWhatsAppConversationID++
	m.whatsappConversations[key] = conversation
	utils.LogInfo(
		"whatsapp conversation created",
		"conversationId", conversation.ID,
		"phoneRef", conversation.PhoneRef,
		"linkedProfile", conversation.LinkedProfile,
		"retentionUntil", conversation.RetentionUntil,
	)
	return conversation
}

// UpdateWhatsAppConversation updates an existing WhatsApp conversation.
func (m *MemoryStore) UpdateWhatsAppConversation(conversation models.WhatsAppConversation) models.WhatsAppConversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.whatsappConversations == nil {
		m.whatsappConversations = map[string]models.WhatsAppConversation{}
	}
	conversation.UpdatedAt = conversation.UpdatedAt.UTC()
	m.whatsappConversations[conversation.Key] = conversation
	utils.LogInfo(
		"whatsapp conversation updated",
		"conversationId", conversation.ID,
		"phoneRef", conversation.PhoneRef,
		"intent", conversation.Intent,
		"state", conversation.State,
		"hazard", conversation.Hazard,
		"urgency", conversation.Urgency,
	)
	return conversation
}

// WhatsAppTranscripts returns stored WhatsApp transcript summaries.
func (m *MemoryStore) WhatsAppTranscripts() []models.WhatsAppTranscript {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return append([]models.WhatsAppTranscript(nil), m.whatsappTranscripts...)
}

// CreateWhatsAppTranscript persists a WhatsApp transcript summary.
func (m *MemoryStore) CreateWhatsAppTranscript(transcript models.WhatsAppTranscript) models.WhatsAppTranscript {
	m.mu.Lock()
	defer m.mu.Unlock()

	transcript.ID = fmt.Sprintf("whatsapp_transcript_%06d", m.nextWhatsAppTranscriptID)
	m.nextWhatsAppTranscriptID++
	m.whatsappTranscripts = append(m.whatsappTranscripts, transcript)
	utils.LogInfo(
		"whatsapp transcript stored",
		"transcriptId", transcript.ID,
		"conversationId", transcript.ConversationID,
		"direction", transcript.Direction,
		"intent", transcript.Intent,
		"state", transcript.State,
		"phoneRef", transcript.PhoneRef,
		"messageSummary", transcript.MessageSummary,
		"mediaSummary", transcript.MediaSummary,
	)
	return transcript
}

// ListDeliveryLogs returns persisted delivery attempts matching the filters.
func (m *MemoryStore) ListDeliveryLogs(filters models.LogFilters) []models.DeliveryAttempt {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := make([]models.DeliveryAttempt, 0, len(m.deliveryLogs))
	for _, log := range m.deliveryLogs {
		if filters.AlertID != "" && log.AlertID != filters.AlertID {
			continue
		}
		if filters.Channel != "" && log.Channel != filters.Channel {
			continue
		}
		if filters.Status != "" && log.Status != filters.Status {
			continue
		}
		logs = append(logs, log)
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].AttemptedAt.After(logs[j].AttemptedAt)
	})
	return logs
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

func copyVoiceAlertAsset(asset models.VoiceAlertAsset) models.VoiceAlertAsset {
	asset.Variants = append([]models.VoiceVariant(nil), asset.Variants...)
	return asset
}

func voiceVariantForLanguage(asset models.VoiceAlertAsset, language string) (models.VoiceVariant, bool) {
	for _, variant := range asset.Variants {
		if variant.Language == language {
			return variant, true
		}
	}
	return models.VoiceVariant{}, false
}

func voiceReviewStatus(action string) string {
	if action == "approve" {
		return "approved"
	}
	return "rejected"
}

func voiceLocale(language string) string {
	switch language {
	case "tw":
		return "ak-GH"
	case "ga":
		return "gaa-GH"
	case "ee":
		return "ee-GH"
	case "dag":
		return "dag-GH"
	case "ha":
		return "ha-GH"
	default:
		return "en-GH"
	}
}

func voiceName(language string) string {
	switch language {
	case "tw":
		return "nadaa-twi-sandbox"
	case "ga":
		return "nadaa-ga-sandbox"
	case "ee":
		return "nadaa-ewe-sandbox"
	case "dag":
		return "nadaa-dagbani-sandbox"
	case "ha":
		return "nadaa-hausa-sandbox"
	default:
		return "nadaa-english-sandbox"
	}
}

func voiceMessageForAlert(language string, alert models.CitizenAlert) string {
	title := strings.TrimSpace(alert.Title)
	target := strings.TrimSpace(alert.TargetLabel)
	if target == "" {
		target = strings.TrimSpace(alert.Target.Label)
	}
	action := strings.TrimSpace(alert.RecommendedAction)
	if action == "" {
		action = strings.TrimSpace(alert.Message)
	}
	switch language {
	case "tw":
		return fmt.Sprintf("NADAA Twi alert. %s. Area: %s. %s Call 112 if life is in danger.", title, target, action)
	case "ga":
		return fmt.Sprintf("NADAA Ga alert. %s. Area: %s. %s Call 112 if life is in danger.", title, target, action)
	case "ee":
		return fmt.Sprintf("NADAA Ewe alert. %s. Area: %s. %s Call 112 if life is in danger.", title, target, action)
	case "dag":
		return fmt.Sprintf("NADAA Dagbani alert. %s. Area: %s. %s Call 112 if life is in danger.", title, target, action)
	case "ha":
		return fmt.Sprintf("NADAA Hausa alert. %s. Area: %s. %s Call 112 if life is in danger.", title, target, action)
	default:
		return fmt.Sprintf("NADAA alert. %s. Area: %s. %s Call 112 if life is in danger.", title, target, action)
	}
}

func estimateVoiceDurationSeconds(message string) int {
	words := len(strings.Fields(message))
	seconds := (words * 60) / 130
	if seconds < 8 {
		return 8
	}
	return seconds
}

func voiceAccessibilityChecks(message string, alert models.CitizenAlert) []string {
	checks := []string{"plain_language", "action_oriented"}
	if strings.TrimSpace(alert.TargetLabel) != "" || strings.TrimSpace(alert.Target.Label) != "" {
		checks = append(checks, "target_area_included")
	}
	if strings.Contains(message, "112") {
		checks = append(checks, "includes_112_guidance")
	}
	if len(strings.Fields(message)) <= 65 {
		checks = append(checks, "low_literacy_length")
	}
	return checks
}
