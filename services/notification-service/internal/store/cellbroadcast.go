package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// Cell broadcast page-encoding limits (3GPP TS 23.041). A CB page is 82 octets,
// which yields 93 GSM-7 septets or 40 UCS-2 characters, and a message may carry
// at most 15 concatenated pages.
const (
	cellBroadcastGSM7CharsPerPage = 93
	cellBroadcastUCS2CharsPerPage = 40
	cellBroadcastMaxPages         = 15
)

// CreateCellBroadcastMessage generates a review-gated cell broadcast message set
// for an already human-approved citizen alert.
func (m *MemoryStore) CreateCellBroadcastMessage(alert models.CitizenAlert, languages []string, areas []string, requestedBy string, now time.Time) models.CellBroadcastMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	channel := cellBroadcastChannelForSeverity(alert.Severity)
	targetLabel := strings.TrimSpace(alert.TargetLabel)
	if targetLabel == "" {
		targetLabel = strings.TrimSpace(alert.Target.Label)
	}
	if len(areas) == 0 && targetLabel != "" {
		areas = []string{targetLabel}
	}

	message := models.CellBroadcastMessage{
		ID:                  fmt.Sprintf("cell_broadcast_%06d", m.nextCellBroadcastID),
		AlertID:             alert.ID,
		AlertTitle:          alert.Title,
		HazardType:          alert.HazardType,
		Severity:            alert.Severity,
		TargetLabel:         targetLabel,
		Areas:               append([]string(nil), areas...),
		Channel:             channel,
		Protocol:            cellBroadcastProtocol(alert.HazardType, alert.Severity),
		Status:              "generated",
		ReviewStatus:        "pending_review",
		EmergencyOverride:   channel.MessageIdentifier == cellBroadcastPresidentialChannel,
		WorkflowRequestedBy: requestedBy,
		Segments:            make([]models.CellBroadcastSegment, 0, len(languages)),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	m.nextCellBroadcastID++

	for _, language := range languages {
		text, charCount, pages, truncated, dcs := renderCellBroadcastSegment(language, alert, targetLabel)
		segment := models.CellBroadcastSegment{
			ID:               fmt.Sprintf("cb_segment_%06d", m.nextCellBroadcastSegmentID),
			Language:         language,
			Locale:           voiceLocale(language),
			DataCodingScheme: dcs,
			MessageText:      text,
			CharacterCount:   charCount,
			Pages:            pages,
			Truncated:        truncated,
			Status:           "generated",
			ReviewStatus:     "pending_review",
			ComplianceChecks: cellBroadcastComplianceChecks(text, pages, truncated, targetLabel, channel),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		m.nextCellBroadcastSegmentID++
		message.Segments = append(message.Segments, segment)
	}

	m.cellBroadcasts = append(m.cellBroadcasts, message)
	utils.LogInfo(
		"cell broadcast message generated",
		"cellBroadcastId", message.ID,
		"alertId", message.AlertID,
		"channel", message.Channel.MessageIdentifier,
		"segmentCount", len(message.Segments),
		"emergencyOverride", message.EmergencyOverride,
		"requestedBy", message.WorkflowRequestedBy,
	)
	return copyCellBroadcastMessage(message)
}

// ListCellBroadcastMessages returns all cell broadcast sets, newest first.
func (m *MemoryStore) ListCellBroadcastMessages() []models.CellBroadcastMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	messages := make([]models.CellBroadcastMessage, 0, len(m.cellBroadcasts))
	for _, message := range m.cellBroadcasts {
		messages = append(messages, copyCellBroadcastMessage(message))
	}
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})
	return messages
}

// GetCellBroadcastMessage returns a cell broadcast set by id.
func (m *MemoryStore) GetCellBroadcastMessage(id string) (models.CellBroadcastMessage, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, message := range m.cellBroadcasts {
		if message.ID == id {
			return copyCellBroadcastMessage(message), true
		}
	}
	return models.CellBroadcastMessage{}, false
}

// ReviewCellBroadcastMessage approves or rejects a cell broadcast set (or a
// subset of its language segments) and recomputes the aggregate review status.
func (m *MemoryStore) ReviewCellBroadcastMessage(id string, action string, reviewer string, note string, languages []string, now time.Time) (models.CellBroadcastMessage, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index, message := range m.cellBroadcasts {
		if message.ID != id {
			continue
		}

		selected := map[string]bool{}
		for _, language := range languages {
			selected[language] = true
		}
		reviewAll := len(selected) == 0
		status := cellBroadcastReviewStatus(action)

		var reviewedCount int
		for segmentIndex, segment := range message.Segments {
			if !reviewAll && !selected[segment.Language] {
				continue
			}
			message.Segments[segmentIndex].Status = status
			message.Segments[segmentIndex].ReviewStatus = status
			message.Segments[segmentIndex].UpdatedAt = now
			reviewedCount++
		}

		message.ReviewStatus = aggregateCellBroadcastReviewStatus(message.Segments, reviewedCount)
		switch message.ReviewStatus {
		case "approved", "rejected":
			message.Status = message.ReviewStatus
		default:
			message.Status = "generated"
		}
		message.Reviewer = reviewer
		message.ReviewNote = note
		message.UpdatedAt = now
		message.ReviewedAt = &now
		m.cellBroadcasts[index] = message

		utils.LogInfo(
			"cell broadcast message reviewed",
			"cellBroadcastId", message.ID,
			"alertId", message.AlertID,
			"action", action,
			"reviewer", reviewer,
			"reviewedCount", reviewedCount,
			"status", message.Status,
			"reviewStatus", message.ReviewStatus,
		)
		return copyCellBroadcastMessage(message), true
	}
	return models.CellBroadcastMessage{}, false
}

// CreateCellBroadcastDispatches broadcasts every approved language segment of an
// approved cell broadcast set through the configured adapter, records an audit
// entry per segment, and returns the dispatch outcomes.
func (m *MemoryStore) CreateCellBroadcastDispatches(ctx context.Context, message models.CellBroadcastMessage, request models.CellBroadcastDeliveryRequest, adapter models.CellBroadcastAdapter, now time.Time) []models.CellBroadcastDispatch {
	m.mu.Lock()
	defer m.mu.Unlock()

	if adapter == nil {
		adapter = models.DisabledCellBroadcastAdapter{Reason: "cell broadcast adapter missing"}
	}
	areas := request.Areas
	if len(areas) == 0 {
		areas = message.Areas
	}

	dispatches := make([]models.CellBroadcastDispatch, 0, len(message.Segments))
	for _, segment := range message.Segments {
		serial := m.nextCellBroadcastSerial
		m.nextCellBroadcastSerial++

		dispatch := models.CellBroadcastDispatch{
			ID:                fmt.Sprintf("cb_dispatch_%06d", serial),
			MessageID:         message.ID,
			AlertID:           message.AlertID,
			Language:          segment.Language,
			MessageIdentifier: message.Channel.MessageIdentifier,
			SerialNumber:      serial,
			Areas:             append([]string(nil), areas...),
			Adapter:           adapter.Name(),
			Pages:             segment.Pages,
			DataCodingScheme:  segment.DataCodingScheme,
			DryRun:            request.DryRun,
			BroadcastAt:       now,
		}

		if segment.ReviewStatus != "approved" {
			dispatch.Status = "skipped"
			dispatch.Reason = "language segment is not approved"
			utils.LogWarn(
				"cell broadcast segment skipped",
				"cellBroadcastId", message.ID,
				"alertId", message.AlertID,
				"language", segment.Language,
				"segmentStatus", segment.ReviewStatus,
			)
		} else {
			result := adapter.Broadcast(ctx, models.CellBroadcastDispatchRequest{
				MessageID:         message.ID,
				AlertID:           message.AlertID,
				Language:          segment.Language,
				MessageIdentifier: message.Channel.MessageIdentifier,
				SerialNumber:      serial,
				Areas:             areas,
				DataCodingScheme:  segment.DataCodingScheme,
				Pages:             segment.Pages,
				MessageText:       segment.MessageText,
				EmergencyOverride: message.EmergencyOverride,
				DryRun:            request.DryRun,
			})
			dispatch.Status = result.Status
			dispatch.Reason = result.Reason
			utils.LogInfo(
				"cell broadcast segment dispatched",
				"cellBroadcastId", message.ID,
				"alertId", message.AlertID,
				"language", segment.Language,
				"adapter", adapter.Name(),
				"channel", message.Channel.MessageIdentifier,
				"serial", serial,
				"status", dispatch.Status,
				"dryRun", request.DryRun,
			)
		}

		m.recordCellBroadcastAuditLocked(message, dispatch)
		dispatches = append(dispatches, dispatch)
	}
	return dispatches
}

// recordCellBroadcastAuditLocked writes a compact entry into the unified delivery
// log so cell broadcasts appear in the same audit stream as other channels. The
// caller must hold m.mu.
func (m *MemoryStore) recordCellBroadcastAuditLocked(message models.CellBroadcastMessage, dispatch models.CellBroadcastDispatch) {
	attempt := models.DeliveryAttempt{
		ID:           fmt.Sprintf("delivery_%06d", m.nextLogID),
		AlertID:      message.AlertID,
		AlertTitle:   message.AlertTitle,
		Channel:      "cell_broadcast",
		Provider:     dispatch.Adapter,
		RecipientRef: "broadcast:" + strings.Join(dispatch.Areas, "|"),
		Status:       dispatch.Status,
		Reason:       dispatch.Reason,
		MessageID:    dispatch.ID,
		Language:     dispatch.Language,
		AttemptedAt:  dispatch.BroadcastAt,
	}
	m.nextLogID++
	m.deliveryLogs = append(m.deliveryLogs, attempt)
}

const cellBroadcastPresidentialChannel = 4370

// cellBroadcastChannelForSeverity maps an alert severity onto a CMAS/WEA message
// identifier. Cell broadcast is reserved for severe-and-above hazards, so the
// lowest tier still maps to the "severe threat" channel. Matching is substring
// based so compound severities (e.g. "severe_warning") classify correctly.
func cellBroadcastChannelForSeverity(severity string) models.CellBroadcastChannel {
	s := utils.NormalizeQueryValue(severity)
	switch {
	case containsAny(s, "extreme", "critical", "catastrophic"):
		return models.CellBroadcastChannel{MessageIdentifier: cellBroadcastPresidentialChannel, Label: "presidential", HandsetCategory: "Presidential Alert"}
	case containsAny(s, "severe", "high"):
		return models.CellBroadcastChannel{MessageIdentifier: 4371, Label: "extreme", HandsetCategory: "Extreme Alert"}
	default:
		return models.CellBroadcastChannel{MessageIdentifier: 4373, Label: "severe", HandsetCategory: "Severe Alert"}
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

// cellBroadcastProtocol derives the CAP-aligned classification operators require.
func cellBroadcastProtocol(hazardType, severity string) models.CellBroadcastProtocol {
	return models.CellBroadcastProtocol{
		Standard:    "3GPP-CBS",
		Category:    capCategory(hazardType),
		Urgency:     capUrgency(severity),
		CAPSeverity: capSeverity(severity),
		Certainty:   "Likely",
	}
}

func capCategory(hazardType string) string {
	h := utils.NormalizeQueryValue(hazardType)
	switch {
	case containsAny(h, "flood", "storm", "rain", "weather", "cyclone"):
		return "Met"
	case containsAny(h, "fire", "wildfire", "bushfire"):
		return "Fire"
	case containsAny(h, "earthquake", "landslide", "tsunami"):
		return "Geo"
	case containsAny(h, "disease", "epidemic", "health", "cholera"):
		return "Health"
	default:
		return "Safety"
	}
}

func capUrgency(severity string) string {
	s := utils.NormalizeQueryValue(severity)
	switch {
	case containsAny(s, "extreme", "critical", "catastrophic", "severe", "high"):
		return "Immediate"
	case containsAny(s, "moderate", "warning", "medium"):
		return "Expected"
	default:
		return "Future"
	}
}

func capSeverity(severity string) string {
	s := utils.NormalizeQueryValue(severity)
	switch {
	case containsAny(s, "extreme", "critical", "catastrophic"):
		return "Extreme"
	case containsAny(s, "severe", "high"):
		return "Severe"
	case containsAny(s, "moderate", "warning", "medium"):
		return "Moderate"
	default:
		return "Minor"
	}
}

// renderCellBroadcastSegment builds a terse, page-bounded broadcast string for a
// language and returns its text, character count, page count, whether it was
// truncated to the 15-page limit, and the data coding scheme used.
func renderCellBroadcastSegment(language string, alert models.CitizenAlert, targetLabel string) (string, int, int, bool, string) {
	text := cellBroadcastMessageText(language, alert, targetLabel)
	dcs, charsPerPage := cellBroadcastEncoding(text)

	runes := []rune(text)
	maxChars := charsPerPage * cellBroadcastMaxPages
	truncated := false
	if len(runes) > maxChars {
		runes = runes[:maxChars]
		text = string(runes)
		truncated = true
	}

	charCount := len(runes)
	pages := max((charCount+charsPerPage-1)/charsPerPage, 1)
	return text, charCount, pages, truncated, dcs
}

func cellBroadcastMessageText(language string, alert models.CitizenAlert, targetLabel string) string {
	title := strings.TrimSpace(alert.Title)
	action := strings.TrimSpace(alert.RecommendedAction)
	if action == "" {
		action = strings.TrimSpace(alert.Message)
	}
	severity := strings.ToUpper(capSeverity(alert.Severity))
	hazard := strings.TrimSpace(alert.HazardType)
	prefix := "NADAA"
	switch language {
	case "tw":
		prefix = "NADAA (Twi)"
	case "ga":
		prefix = "NADAA (Ga)"
	case "ee":
		prefix = "NADAA (Ewe)"
	case "dag":
		prefix = "NADAA (Dagbani)"
	case "ha":
		prefix = "NADAA (Hausa)"
	}
	return fmt.Sprintf("%s %s %s alert: %s. Area: %s. %s Call 112.", prefix, severity, hazard, title, targetLabel, action)
}

// cellBroadcastEncoding picks GSM-7 when the text is representable in 7-bit
// ASCII, otherwise UCS-2, and returns the matching characters-per-page budget.
// (ASCII is a conservative proxy for the GSM 03.38 basic alphabet.)
func cellBroadcastEncoding(text string) (string, int) {
	for _, r := range text {
		if r > 127 {
			return "UCS-2", cellBroadcastUCS2CharsPerPage
		}
	}
	return "GSM-7", cellBroadcastGSM7CharsPerPage
}

func cellBroadcastComplianceChecks(text string, pages int, truncated bool, targetLabel string, channel models.CellBroadcastChannel) []string {
	checks := []string{"cap_classified", "channel_assigned"}
	if !truncated && pages <= cellBroadcastMaxPages {
		checks = append(checks, "within_page_limit")
	}
	if strings.Contains(text, "112") {
		checks = append(checks, "includes_emergency_number")
	}
	if strings.TrimSpace(targetLabel) != "" {
		checks = append(checks, "target_area_specified")
	}
	if channel.MessageIdentifier == cellBroadcastPresidentialChannel {
		checks = append(checks, "presidential_override")
	}
	return checks
}

func cellBroadcastReviewStatus(action string) string {
	if action == "approve" {
		return "approved"
	}
	return "rejected"
}

func aggregateCellBroadcastReviewStatus(segments []models.CellBroadcastSegment, reviewedCount int) string {
	if reviewedCount == 0 {
		return "pending_review"
	}
	var approvedCount, rejectedCount int
	for _, segment := range segments {
		switch segment.ReviewStatus {
		case "approved":
			approvedCount++
		case "rejected":
			rejectedCount++
		}
	}
	switch {
	case approvedCount == len(segments):
		return "approved"
	case rejectedCount == len(segments):
		return "rejected"
	default:
		return "partial_review"
	}
}

func copyCellBroadcastMessage(message models.CellBroadcastMessage) models.CellBroadcastMessage {
	clone := message
	clone.Areas = append([]string(nil), message.Areas...)
	clone.Segments = make([]models.CellBroadcastSegment, len(message.Segments))
	for i, segment := range message.Segments {
		segmentCopy := segment
		segmentCopy.ComplianceChecks = append([]string(nil), segment.ComplianceChecks...)
		clone.Segments[i] = segmentCopy
	}
	if message.ReviewedAt != nil {
		reviewedAt := *message.ReviewedAt
		clone.ReviewedAt = &reviewedAt
	}
	return clone
}
