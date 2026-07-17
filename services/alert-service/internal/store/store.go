package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

// Store is the persistence interface for alert data.
type Store interface {
	CreateAlert(request models.CreateAlertRequest, ctx models.AuthorityContext, now time.Time) models.AuthorityAlert
	UpdateAlert(id string, request models.CreateAlertRequest, ctx models.AuthorityContext, now time.Time) (models.AuthorityAlert, string, string)
	TransitionAlert(id string, nextStatus string, ctx models.AuthorityContext, request models.WorkflowRequest, now time.Time) (models.AuthorityAlert, string, string)
	ListAlerts(status string, currentOnly bool, publicOnly bool, targetType string, targetID string, now time.Time) []models.AuthorityAlert
	ListAudit(limit int) []models.AuditEvent
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu     sync.RWMutex
	alerts []models.AuthorityAlert
	audit  []models.AuditEvent
	nextID int
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	return &MemoryStore{
		alerts: seedAlerts(now),
		nextID: 1,
	}
}

// CreateAlert creates a new draft alert.
func (m *MemoryStore) CreateAlert(request models.CreateAlertRequest, ctx models.AuthorityContext, now time.Time) models.AuthorityAlert {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert := models.AuthorityAlert{
		ID:                 fmt.Sprintf("alert_%06d", m.nextID),
		Title:              strings.TrimSpace(request.Title),
		HazardType:         utils.NormalizeQueryValue(request.HazardType),
		Severity:           utils.NormalizeQueryValue(request.Severity),
		Message:            strings.TrimSpace(request.Message),
		Target:             utils.NormalizeTarget(request.Target),
		StartsAt:           request.StartsAt,
		ExpiresAt:          request.ExpiresAt,
		RecommendedAction:  strings.TrimSpace(request.RecommendedAction),
		EvacuationRequired: request.EvacuationRequired,
		ShelterIDs:         utils.CompactStrings(request.ShelterIDs),
		IssuingAgencyID:    ctx.ActorAgencyID,
		IssuedBy:           ctx.ActorUserID,
		Status:             "draft",
		CreatedAt:          now,
		UpdatedAt:          now,
		SourcePrediction:   utils.NormalizeSourcePrediction(request.SourcePrediction),
	}
	m.nextID++
	m.alerts = append(m.alerts, alert)
	m.appendAuditLocked("alert.created", ctx, alert.ID, nil, utils.SnapshotAlert(alert), now)
	return alert
}

// UpdateAlert updates an alert that is in draft or rejected status.
func (m *MemoryStore) UpdateAlert(id string, request models.CreateAlertRequest, ctx models.AuthorityContext, now time.Time) (models.AuthorityAlert, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := m.findAlertIndex(id)
	if index < 0 {
		return models.AuthorityAlert{}, "not_found", "alert was not found"
	}
	alert := m.alerts[index]
	if alert.Status != "draft" && alert.Status != "rejected" {
		return models.AuthorityAlert{}, "invalid_transition", "only draft or rejected alerts can be updated"
	}
	if alert.IssuedBy != ctx.ActorUserID && !utils.ApprovalRoles[ctx.ActorRole] {
		return models.AuthorityAlert{}, "forbidden", "only the drafter or an approver can update this alert"
	}

	before := utils.SnapshotAlert(alert)
	alert.Title = strings.TrimSpace(request.Title)
	alert.HazardType = utils.NormalizeQueryValue(request.HazardType)
	alert.Severity = utils.NormalizeQueryValue(request.Severity)
	alert.Message = strings.TrimSpace(request.Message)
	alert.Target = utils.NormalizeTarget(request.Target)
	alert.StartsAt = request.StartsAt
	alert.ExpiresAt = request.ExpiresAt
	alert.RecommendedAction = strings.TrimSpace(request.RecommendedAction)
	alert.EvacuationRequired = request.EvacuationRequired
	alert.ShelterIDs = utils.CompactStrings(request.ShelterIDs)
	alert.SourcePrediction = utils.NormalizeSourcePrediction(request.SourcePrediction)
	alert.Status = "draft"
	alert.SubmittedAt = nil
	alert.RejectedBy = ""
	alert.StatusReason = ""
	alert.RejectedAt = nil
	alert.UpdatedAt = now
	m.alerts[index] = alert
	m.appendAuditLocked("alert.updated", ctx, alert.ID, before, utils.SnapshotAlert(alert), now)
	return alert, "", ""
}

// TransitionAlert moves an alert through the workflow.
func (m *MemoryStore) TransitionAlert(id string, nextStatus string, ctx models.AuthorityContext, request models.WorkflowRequest, now time.Time) (models.AuthorityAlert, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := m.findAlertIndex(id)
	if index < 0 {
		return models.AuthorityAlert{}, "not_found", "alert was not found"
	}

	alert := m.alerts[index]
	before := utils.SnapshotAlert(alert)

	switch nextStatus {
	case "submitted":
		if alert.Status != "draft" {
			return models.AuthorityAlert{}, "invalid_transition", "only draft alerts can be submitted"
		}
		if alert.IssuedBy != ctx.ActorUserID && !utils.ApprovalRoles[ctx.ActorRole] {
			return models.AuthorityAlert{}, "forbidden", "only the drafter or an approver can submit this alert"
		}
		alert.Status = "submitted"
		alert.SubmittedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Note)
		m.appendAuditLocked("alert.submitted", ctx, alert.ID, before, utils.SnapshotAlert(alert), now)
	case "approved":
		if alert.Status != "submitted" {
			return models.AuthorityAlert{}, "invalid_transition", "only submitted alerts can be approved"
		}
		if alert.IssuedBy == ctx.ActorUserID && ctx.ActorRole != "system_admin" {
			return models.AuthorityAlert{}, "separation_of_duties", "approver must be different from drafter unless actor is system_admin"
		}
		alert.Status = "approved"
		alert.ApprovedBy = ctx.ActorUserID
		alert.ApprovedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Note)
		m.appendAuditLocked("alert.approved", ctx, alert.ID, before, utils.SnapshotAlert(alert), now)
	case "rejected":
		if alert.Status != "submitted" {
			return models.AuthorityAlert{}, "invalid_transition", "only submitted alerts can be rejected"
		}
		alert.Status = "rejected"
		alert.RejectedBy = ctx.ActorUserID
		alert.RejectedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Reason)
		m.appendAuditLocked("alert.rejected", ctx, alert.ID, before, utils.SnapshotAlert(alert), now)
	case "emergency_override":
		if alert.Status == "approved" || alert.Status == "published" {
			return models.AuthorityAlert{}, "invalid_transition", "approved or published alerts do not need override"
		}
		alert.Status = "approved"
		alert.EmergencyOverride = true
		alert.ApprovedBy = ctx.ActorUserID
		alert.ApprovedAt = &now
		alert.RejectedBy = ""
		alert.RejectedAt = nil
		alert.StatusReason = strings.TrimSpace(request.Reason)
		m.appendAuditLocked("alert.emergency_override", ctx, alert.ID, before, utils.SnapshotAlert(alert), now)
	default:
		return models.AuthorityAlert{}, "invalid_transition", "unsupported alert transition"
	}

	alert.UpdatedAt = now
	m.alerts[index] = alert
	return alert, "", ""
}

// ListAlerts returns alerts matching the provided filters.
func (m *MemoryStore) ListAlerts(status string, currentOnly bool, publicOnly bool, targetType string, targetID string, now time.Time) []models.AuthorityAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]models.AuthorityAlert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		if publicOnly && alert.Status != "approved" && alert.Status != "published" {
			continue
		}
		if status != "" && alert.Status != status {
			continue
		}
		if currentOnly && (alert.StartsAt.After(now) || !alert.ExpiresAt.After(now)) {
			continue
		}
		if targetType != "" && alert.Target.Type != targetType {
			continue
		}
		if targetID != "" && !utils.ContainsString(alert.Target.IDs, targetID) {
			continue
		}
		responseAlert := alert
		if publicOnly {
			responseAlert.SourcePrediction = nil
		}
		alerts = append(alerts, responseAlert)
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].UpdatedAt.After(alerts[j].UpdatedAt)
	})
	return alerts
}

// ListAudit returns the most recent audit logs up to limit.
func (m *MemoryStore) ListAudit(limit int) []models.AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := append([]models.AuditEvent(nil), m.audit...)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	if len(logs) > limit {
		return logs[:limit]
	}
	return logs
}

func (m *MemoryStore) findAlertIndex(id string) int {
	for index, alert := range m.alerts {
		if alert.ID == id {
			return index
		}
	}
	return -1
}

func (m *MemoryStore) appendAuditLocked(action string, ctx models.AuthorityContext, targetID string, before map[string]any, after map[string]any, now time.Time) {
	m.audit = append(m.audit, models.AuditEvent{
		ID:            fmt.Sprintf("aud_%06d", len(m.audit)+1),
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Action:        action,
		TargetType:    "alert",
		TargetID:      targetID,
		RequestID:     ctx.RequestID,
		Before:        before,
		After:         after,
		CreatedAt:     now,
	})
}
