package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *memoryStore
}

type memoryStore struct {
	mu     sync.RWMutex
	alerts []authorityAlert
	audit  []auditEvent
	nextID int
}

type authorityAlert struct {
	ID                 string                 `json:"id"`
	Title              string                 `json:"title"`
	HazardType         string                 `json:"hazardType"`
	Severity           string                 `json:"severity"`
	Message            string                 `json:"message"`
	Target             alertTarget            `json:"target"`
	StartsAt           time.Time              `json:"startsAt"`
	ExpiresAt          time.Time              `json:"expiresAt"`
	RecommendedAction  string                 `json:"recommendedAction"`
	EvacuationRequired bool                   `json:"evacuationRequired"`
	ShelterIDs         []string               `json:"shelterIds"`
	IssuingAgencyID    string                 `json:"issuingAgencyId"`
	IssuedBy           string                 `json:"issuedBy"`
	ApprovedBy         string                 `json:"approvedBy,omitempty"`
	RejectedBy         string                 `json:"rejectedBy,omitempty"`
	Status             string                 `json:"status"`
	EmergencyOverride  bool                   `json:"emergencyOverride"`
	StatusReason       string                 `json:"statusReason,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	SubmittedAt        *time.Time             `json:"submittedAt,omitempty"`
	ApprovedAt         *time.Time             `json:"approvedAt,omitempty"`
	RejectedAt         *time.Time             `json:"rejectedAt,omitempty"`
	SourcePrediction   *alertSourcePrediction `json:"sourcePrediction,omitempty"`
}

type alertTarget struct {
	Type                string          `json:"type"`
	IDs                 []string        `json:"ids"`
	Label               string          `json:"label"`
	Center              *coordinates    `json:"center,omitempty"`
	RadiusMeters        float64         `json:"radiusMeters,omitempty"`
	Geometry            *targetGeometry `json:"geometry,omitempty"`
	AreaSqKm            float64         `json:"areaSqKm,omitempty"`
	EstimatedPopulation int             `json:"estimatedPopulation,omitempty"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type targetGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type createAlertRequest struct {
	Title              string                 `json:"title"`
	HazardType         string                 `json:"hazardType"`
	Severity           string                 `json:"severity"`
	Message            string                 `json:"message"`
	Target             alertTarget            `json:"target"`
	StartsAt           time.Time              `json:"startsAt"`
	ExpiresAt          time.Time              `json:"expiresAt"`
	RecommendedAction  string                 `json:"recommendedAction"`
	EvacuationRequired bool                   `json:"evacuationRequired"`
	ShelterIDs         []string               `json:"shelterIds"`
	SourcePrediction   *alertSourcePrediction `json:"sourcePrediction,omitempty"`
}

type alertSourcePrediction struct {
	PredictionID           string  `json:"predictionId"`
	PredictionLogID        string  `json:"predictionLogId,omitempty"`
	ModelVersion           string  `json:"modelVersion"`
	InputFeatureSetVersion string  `json:"inputFeatureSetVersion"`
	Probability            float64 `json:"probability"`
	Severity               string  `json:"severity"`
	Confidence             string  `json:"confidence"`
	HumanReviewRequired    bool    `json:"humanReviewRequired"`
	AutoPublishAllowed     bool    `json:"autoPublishAllowed"`
	ReviewNote             string  `json:"reviewNote,omitempty"`
}

type workflowRequest struct {
	Note   string `json:"note"`
	Reason string `json:"reason"`
}

type alertListResponse struct {
	Alerts []authorityAlert `json:"alerts"`
}

type targetPreviewResponse struct {
	Target   alertTarget `json:"target"`
	Summary  string      `json:"summary"`
	Warnings []string    `json:"warnings"`
}

type auditListResponse struct {
	Logs []auditEvent `json:"logs"`
}

type auditEvent struct {
	ID            string         `json:"id"`
	ActorUserID   string         `json:"actorUserId"`
	ActorAgencyID string         `json:"actorAgencyId"`
	ActorRole     string         `json:"actorRole"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId"`
	RequestID     string         `json:"requestId,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type authorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type targetCatalogRecord struct {
	ID                  string
	Type                string
	Label               string
	Center              coordinates
	RadiusMeters        float64
	AreaSqKm            float64
	EstimatedPopulation int
}

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

var allowedRiskLevels = map[string]bool{
	"low":       true,
	"moderate":  true,
	"high":      true,
	"severe":    true,
	"emergency": true,
}

var allowedTargetTypes = map[string]bool{
	"national":  true,
	"region":    true,
	"district":  true,
	"radius":    true,
	"community": true,
	"custom":    true,
}

var draftRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

var approvalRoles = map[string]bool{
	"system_admin":  true,
	"agency_admin":  true,
	"nadmo_officer": true,
}

var overrideRoles = map[string]bool{
	"system_admin":  true,
	"nadmo_officer": true,
}

var targetCatalog = map[string]targetCatalogRecord{
	"region:greater-accra": {
		ID:                  "greater-accra",
		Type:                "region",
		Label:               "Greater Accra Region",
		Center:              coordinates{Lat: 5.75, Lng: -0.11},
		RadiusMeters:        52000,
		AreaSqKm:            3245,
		EstimatedPopulation: 5455000,
	},
	"district:accra-metropolitan": {
		ID:                  "accra-metropolitan",
		Type:                "district",
		Label:               "Accra Metropolitan",
		Center:              coordinates{Lat: 5.56, Lng: -0.2},
		RadiusMeters:        9000,
		AreaSqKm:            61,
		EstimatedPopulation: 284000,
	},
	"district:tema-metropolitan": {
		ID:                  "tema-metropolitan",
		Type:                "district",
		Label:               "Tema Metropolitan",
		Center:              coordinates{Lat: 5.642, Lng: -0.028},
		RadiusMeters:        12000,
		AreaSqKm:            565,
		EstimatedPopulation: 402000,
	},
	"district:ablekuma-west": {
		ID:                  "ablekuma-west",
		Type:                "district",
		Label:               "Ablekuma West",
		Center:              coordinates{Lat: 5.601, Lng: -0.286},
		RadiusMeters:        7000,
		AreaSqKm:            15,
		EstimatedPopulation: 220000,
	},
	"community:accra-central": {
		ID:                  "accra-central",
		Type:                "community",
		Label:               "Accra Central",
		Center:              coordinates{Lat: 5.556, Lng: -0.202},
		RadiusMeters:        3000,
		AreaSqKm:            8,
		EstimatedPopulation: 75000,
	},
}

func main() {
	srv := &server{store: newMemoryStore()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/alerts", srv.createAlertHandler)
	mux.HandleFunc("GET /api/v1/alerts", srv.listAlertsHandler)
	mux.HandleFunc("GET /api/v1/alerts/audit", srv.listAuditHandler)
	mux.HandleFunc("POST /api/v1/alerts/targets/preview", srv.previewTargetHandler)
	mux.HandleFunc("PATCH /api/v1/alerts/{id}", srv.updateAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/submit", srv.submitAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/approve", srv.approveAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/reject", srv.rejectAlertHandler)
	mux.HandleFunc("POST /api/v1/alerts/{id}/emergency-override", srv.emergencyOverrideHandler)

	addr := envOrDefault("NADAA_ALERT_ADDR", ":8089")
	log.Printf("alert-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newMemoryStore() *memoryStore {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &memoryStore{
		alerts: []authorityAlert{
			{
				ID:                 "alert_fixture_submitted",
				Title:              "Accra flood watch",
				HazardType:         "flood",
				Severity:           "warning",
				Message:            "Heavy rainfall may cause flooding in low-lying communities.",
				Target:             alertTarget{Type: "district", IDs: []string{"accra-metropolitan"}, Label: "Accra Metropolitan"},
				StartsAt:           now.Add(30 * time.Minute),
				ExpiresAt:          now.Add(12 * time.Hour),
				RecommendedAction:  "Avoid flooded roads and prepare to move to higher ground.",
				EvacuationRequired: false,
				ShelterIDs:         []string{"00000000-0000-0000-0000-000000000301"},
				IssuingAgencyID:    "00000000-0000-0000-0000-000000000101",
				IssuedBy:           "usr_dispatcher_fixture",
				Status:             "submitted",
				CreatedAt:          now.Add(-45 * time.Minute),
				UpdatedAt:          now.Add(-15 * time.Minute),
				SubmittedAt:        timePtr(now.Add(-15 * time.Minute)),
			},
		},
		nextID: 1,
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "alert-service"})
}

func (s *server) createAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, draftRoles)
	if !ok {
		return
	}

	var request createAlertRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized := normalizeAlertRequest(request)
	if code, message := validateAlertRequest(normalized); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	alert := s.store.createAlert(normalized, ctx, time.Now().UTC())
	writeJSON(w, http.StatusCreated, alert)
}

func (s *server) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	status := normalizeQueryValue(r.URL.Query().Get("status"))
	currentOnly := normalizeQueryValue(r.URL.Query().Get("current")) == "true"
	targetType := normalizeQueryValue(r.URL.Query().Get("targetType"))
	targetID := normalizeQueryValue(r.URL.Query().Get("targetId"))
	if status != "" && !validAlertStatus(status) {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be draft, submitted, approved, rejected, published, expired, or cancelled")
		return
	}
	if targetType != "" && !allowedTargetTypes[targetType] {
		writeError(w, http.StatusBadRequest, "invalid_target_type", "targetType must be national, region, district, radius, community, or custom")
		return
	}

	publicOnly := !hasAuthorityHeaders(r)
	writeJSON(w, http.StatusOK, alertListResponse{Alerts: s.store.listAlerts(status, currentOnly, publicOnly, targetType, targetID, time.Now().UTC())})
}

func (s *server) listAuditHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r, approvalRoles); !ok {
		return
	}

	limit := 50
	if raw := normalizeQueryValue(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}

	writeJSON(w, http.StatusOK, auditListResponse{Logs: s.store.listAudit(limit)})
}

func (s *server) previewTargetHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r, draftRoles); !ok {
		return
	}

	var target alertTarget
	if err := decodeJSON(r, &target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized := normalizeTarget(target)
	if code, message := validateTarget(normalized); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	writeJSON(w, http.StatusOK, targetPreviewResponse{
		Target:   normalized,
		Summary:  targetSummary(normalized),
		Warnings: targetWarnings(normalized),
	})
}

func (s *server) updateAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, draftRoles)
	if !ok {
		return
	}

	var request createAlertRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized := normalizeAlertRequest(request)
	if code, message := validateAlertRequest(normalized); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	alert, code, message := s.store.updateAlert(r.PathValue("id"), normalized, ctx, time.Now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (s *server) submitAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, draftRoles)
	if !ok {
		return
	}

	alert, code, message := s.store.transitionAlert(r.PathValue("id"), "submitted", ctx, workflowRequest{}, time.Now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (s *server) approveAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, approvalRoles)
	if !ok {
		return
	}

	var request workflowRequest
	if err := optionalDecodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	alert, code, message := s.store.transitionAlert(r.PathValue("id"), "approved", ctx, request, time.Now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (s *server) rejectAlertHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, approvalRoles)
	if !ok {
		return
	}

	var request workflowRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if strings.TrimSpace(request.Reason) == "" {
		writeError(w, http.StatusBadRequest, "missing_reason", "reason is required when rejecting an alert")
		return
	}

	alert, code, message := s.store.transitionAlert(r.PathValue("id"), "rejected", ctx, request, time.Now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (s *server) emergencyOverrideHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, overrideRoles)
	if !ok {
		return
	}

	var request workflowRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if strings.TrimSpace(request.Reason) == "" {
		writeError(w, http.StatusBadRequest, "missing_reason", "reason is required for emergency override")
		return
	}

	alert, code, message := s.store.transitionAlert(r.PathValue("id"), "emergency_override", ctx, request, time.Now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, alert)
}

func (m *memoryStore) createAlert(request createAlertRequest, ctx authorityContext, now time.Time) authorityAlert {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert := authorityAlert{
		ID:                 fmt.Sprintf("alert_%06d", m.nextID),
		Title:              strings.TrimSpace(request.Title),
		HazardType:         normalizeQueryValue(request.HazardType),
		Severity:           normalizeQueryValue(request.Severity),
		Message:            strings.TrimSpace(request.Message),
		Target:             normalizeTarget(request.Target),
		StartsAt:           request.StartsAt,
		ExpiresAt:          request.ExpiresAt,
		RecommendedAction:  strings.TrimSpace(request.RecommendedAction),
		EvacuationRequired: request.EvacuationRequired,
		ShelterIDs:         compactStrings(request.ShelterIDs),
		IssuingAgencyID:    ctx.ActorAgencyID,
		IssuedBy:           ctx.ActorUserID,
		Status:             "draft",
		CreatedAt:          now,
		UpdatedAt:          now,
		SourcePrediction:   normalizeSourcePrediction(request.SourcePrediction),
	}
	m.nextID++
	m.alerts = append(m.alerts, alert)
	m.appendAuditLocked("alert.created", ctx, alert.ID, nil, snapshotAlert(alert), now)
	return alert
}

func (m *memoryStore) updateAlert(id string, request createAlertRequest, ctx authorityContext, now time.Time) (authorityAlert, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := m.findAlertIndex(id)
	if index < 0 {
		return authorityAlert{}, "not_found", "alert was not found"
	}
	alert := m.alerts[index]
	if alert.Status != "draft" && alert.Status != "rejected" {
		return authorityAlert{}, "invalid_transition", "only draft or rejected alerts can be updated"
	}
	if alert.IssuedBy != ctx.ActorUserID && !approvalRoles[ctx.ActorRole] {
		return authorityAlert{}, "forbidden", "only the drafter or an approver can update this alert"
	}

	before := snapshotAlert(alert)
	alert.Title = strings.TrimSpace(request.Title)
	alert.HazardType = normalizeQueryValue(request.HazardType)
	alert.Severity = normalizeQueryValue(request.Severity)
	alert.Message = strings.TrimSpace(request.Message)
	alert.Target = normalizeTarget(request.Target)
	alert.StartsAt = request.StartsAt
	alert.ExpiresAt = request.ExpiresAt
	alert.RecommendedAction = strings.TrimSpace(request.RecommendedAction)
	alert.EvacuationRequired = request.EvacuationRequired
	alert.ShelterIDs = compactStrings(request.ShelterIDs)
	alert.SourcePrediction = normalizeSourcePrediction(request.SourcePrediction)
	alert.Status = "draft"
	alert.RejectedBy = ""
	alert.StatusReason = ""
	alert.RejectedAt = nil
	alert.UpdatedAt = now
	m.alerts[index] = alert
	m.appendAuditLocked("alert.updated", ctx, alert.ID, before, snapshotAlert(alert), now)
	return alert, "", ""
}

func (m *memoryStore) transitionAlert(id string, nextStatus string, ctx authorityContext, request workflowRequest, now time.Time) (authorityAlert, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := m.findAlertIndex(id)
	if index < 0 {
		return authorityAlert{}, "not_found", "alert was not found"
	}

	alert := m.alerts[index]
	before := snapshotAlert(alert)

	switch nextStatus {
	case "submitted":
		if alert.Status != "draft" {
			return authorityAlert{}, "invalid_transition", "only draft alerts can be submitted"
		}
		if alert.IssuedBy != ctx.ActorUserID && !approvalRoles[ctx.ActorRole] {
			return authorityAlert{}, "forbidden", "only the drafter or an approver can submit this alert"
		}
		alert.Status = "submitted"
		alert.SubmittedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Note)
		m.appendAuditLocked("alert.submitted", ctx, alert.ID, before, snapshotAlert(alert), now)
	case "approved":
		if alert.Status != "submitted" {
			return authorityAlert{}, "invalid_transition", "only submitted alerts can be approved"
		}
		if alert.IssuedBy == ctx.ActorUserID && ctx.ActorRole != "system_admin" {
			return authorityAlert{}, "separation_of_duties", "approver must be different from drafter unless actor is system_admin"
		}
		alert.Status = "approved"
		alert.ApprovedBy = ctx.ActorUserID
		alert.ApprovedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Note)
		m.appendAuditLocked("alert.approved", ctx, alert.ID, before, snapshotAlert(alert), now)
	case "rejected":
		if alert.Status != "submitted" {
			return authorityAlert{}, "invalid_transition", "only submitted alerts can be rejected"
		}
		alert.Status = "rejected"
		alert.RejectedBy = ctx.ActorUserID
		alert.RejectedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Reason)
		m.appendAuditLocked("alert.rejected", ctx, alert.ID, before, snapshotAlert(alert), now)
	case "emergency_override":
		if alert.Status == "approved" || alert.Status == "published" {
			return authorityAlert{}, "invalid_transition", "approved or published alerts do not need override"
		}
		alert.Status = "approved"
		alert.EmergencyOverride = true
		alert.ApprovedBy = ctx.ActorUserID
		alert.ApprovedAt = &now
		alert.StatusReason = strings.TrimSpace(request.Reason)
		m.appendAuditLocked("alert.emergency_override", ctx, alert.ID, before, snapshotAlert(alert), now)
	default:
		return authorityAlert{}, "invalid_transition", "unsupported alert transition"
	}

	alert.UpdatedAt = now
	m.alerts[index] = alert
	return alert, "", ""
}

func (m *memoryStore) listAlerts(status string, currentOnly bool, publicOnly bool, targetType string, targetID string, now time.Time) []authorityAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]authorityAlert, 0, len(m.alerts))
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
		if targetID != "" && !containsString(alert.Target.IDs, targetID) {
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

func (m *memoryStore) listAudit(limit int) []auditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := append([]auditEvent(nil), m.audit...)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	if len(logs) > limit {
		return logs[:limit]
	}
	return logs
}

func (m *memoryStore) findAlertIndex(id string) int {
	for index, alert := range m.alerts {
		if alert.ID == id {
			return index
		}
	}
	return -1
}

func (m *memoryStore) appendAuditLocked(action string, ctx authorityContext, targetID string, before map[string]any, after map[string]any, now time.Time) {
	m.audit = append(m.audit, auditEvent{
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

func requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (authorityContext, bool) {
	ctx := authorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     normalizeQueryValue(r.Header.Get("X-NADAA-Actor-Role")),
		MFACompleted:  normalizeQueryValue(r.Header.Get("X-NADAA-MFA-Completed")) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		writeError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return authorityContext{}, false
	}
	if !ctx.MFACompleted {
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for alert workflow actions")
		return authorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this alert workflow action")
		return authorityContext{}, false
	}

	return ctx, true
}

func hasAuthorityHeaders(r *http.Request) bool {
	return strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")) != "" &&
		strings.TrimSpace(r.Header.Get("X-NADAA-Actor-Role")) != "" &&
		strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")) != ""
}

func normalizeAlertRequest(request createAlertRequest) createAlertRequest {
	request.Title = strings.TrimSpace(request.Title)
	request.HazardType = normalizeQueryValue(request.HazardType)
	request.Severity = normalizeQueryValue(request.Severity)
	request.Message = strings.TrimSpace(request.Message)
	request.Target = normalizeTarget(request.Target)
	request.RecommendedAction = strings.TrimSpace(request.RecommendedAction)
	request.ShelterIDs = compactStrings(request.ShelterIDs)
	request.SourcePrediction = normalizeSourcePrediction(request.SourcePrediction)
	return request
}

func validateAlertRequest(request createAlertRequest) (string, string) {
	title := strings.TrimSpace(request.Title)
	message := strings.TrimSpace(request.Message)
	action := strings.TrimSpace(request.RecommendedAction)
	hazard := normalizeQueryValue(request.HazardType)
	severity := normalizeQueryValue(request.Severity)

	if len(title) < 4 || len(title) > 140 {
		return "invalid_title", "title must be 4 to 140 characters"
	}
	if !allowedHazards[hazard] {
		return "invalid_hazard", "hazardType must be a supported NADAA hazard type"
	}
	if !allowedSeverities[severity] {
		return "invalid_severity", "severity must be advisory, watch, warning, severe_warning, or emergency"
	}
	if len(message) < 10 || len(message) > 1000 {
		return "invalid_message", "message must be 10 to 1000 characters"
	}
	if code, message := validateTarget(request.Target); code != "" {
		return code, message
	}
	if request.StartsAt.IsZero() {
		return "missing_starts_at", "startsAt is required"
	}
	if request.ExpiresAt.IsZero() || !request.ExpiresAt.After(request.StartsAt) {
		return "invalid_expiry", "expiresAt must be after startsAt"
	}
	if action == "" {
		return "missing_recommended_action", "recommendedAction is required"
	}
	if code, message := validateSourcePrediction(request.SourcePrediction); code != "" {
		return code, message
	}

	return "", ""
}

func normalizeSourcePrediction(source *alertSourcePrediction) *alertSourcePrediction {
	if source == nil {
		return nil
	}
	return &alertSourcePrediction{
		PredictionID:           strings.TrimSpace(source.PredictionID),
		PredictionLogID:        strings.TrimSpace(source.PredictionLogID),
		ModelVersion:           strings.TrimSpace(source.ModelVersion),
		InputFeatureSetVersion: strings.TrimSpace(source.InputFeatureSetVersion),
		Probability:            roundProbability(source.Probability),
		Severity:               normalizeQueryValue(source.Severity),
		Confidence:             normalizeQueryValue(source.Confidence),
		HumanReviewRequired:    source.HumanReviewRequired,
		AutoPublishAllowed:     source.AutoPublishAllowed,
		ReviewNote:             strings.TrimSpace(source.ReviewNote),
	}
}

func validateSourcePrediction(source *alertSourcePrediction) (string, string) {
	if source == nil {
		return "", ""
	}
	if source.PredictionID == "" {
		return "missing_prediction_id", "sourcePrediction.predictionId is required"
	}
	if source.ModelVersion == "" || source.InputFeatureSetVersion == "" {
		return "missing_prediction_model", "sourcePrediction model and feature set versions are required"
	}
	if source.Probability < 0 || source.Probability > 1 {
		return "invalid_prediction_probability", "sourcePrediction.probability must be between 0 and 1"
	}
	if !source.HumanReviewRequired || source.AutoPublishAllowed {
		return "invalid_prediction_safety", "sourcePrediction must require human review and disallow auto-publish"
	}
	if !allowedRiskLevels[source.Severity] {
		return "invalid_prediction_severity", "sourcePrediction.severity must be a supported risk level"
	}
	if source.Confidence != "low" && source.Confidence != "medium" && source.Confidence != "high" {
		return "invalid_prediction_confidence", "sourcePrediction.confidence must be low, medium, or high"
	}
	if len(source.ReviewNote) > 400 {
		return "invalid_prediction_review_note", "sourcePrediction.reviewNote must be 400 characters or fewer"
	}
	return "", ""
}

func normalizeTarget(target alertTarget) alertTarget {
	normalized := alertTarget{
		Type:                normalizeQueryValue(target.Type),
		IDs:                 normalizeTargetIDs(target.IDs),
		Label:               strings.TrimSpace(target.Label),
		Center:              normalizeCenter(target.Center),
		RadiusMeters:        roundMeters(target.RadiusMeters),
		Geometry:            normalizeGeometry(target.Geometry),
		AreaSqKm:            roundArea(target.AreaSqKm),
		EstimatedPopulation: target.EstimatedPopulation,
	}

	switch normalized.Type {
	case "national":
		normalized.IDs = []string{"ghana"}
		if normalized.Label == "" {
			normalized.Label = "Ghana"
		}
		normalized.Center = &coordinates{Lat: 7.9465, Lng: -1.0232}
		normalized.RadiusMeters = 365000
		normalized.AreaSqKm = 238533
		normalized.EstimatedPopulation = 33480000
		normalized.Geometry = geometryFromBounds(4.54, -3.26, 11.18, 1.2)
	case "region", "district", "community":
		applyCatalogTarget(&normalized)
	case "radius":
		if normalized.IDs == nil {
			normalized.IDs = []string{"radius"}
		}
		if normalized.Label == "" {
			normalized.Label = "Radius target"
		}
		if normalized.RadiusMeters > 0 {
			normalized.AreaSqKm = roundArea(math.Pi * math.Pow(normalized.RadiusMeters/1000, 2))
		}
		if normalized.AreaSqKm > 0 && normalized.EstimatedPopulation == 0 {
			normalized.EstimatedPopulation = int(math.Round(normalized.AreaSqKm * 4500))
		}
	case "custom":
		if normalized.IDs == nil {
			normalized.IDs = []string{"custom"}
		}
		if normalized.Label == "" {
			normalized.Label = "Custom geometry"
		}
		if normalized.Geometry != nil {
			normalized.Center = polygonCenter(normalized.Geometry)
			normalized.AreaSqKm = polygonAreaSqKm(normalized.Geometry)
			if normalized.AreaSqKm > 0 && normalized.EstimatedPopulation == 0 {
				normalized.EstimatedPopulation = int(math.Round(normalized.AreaSqKm * 5000))
			}
		}
	}

	return normalized
}

func validateTarget(target alertTarget) (string, string) {
	if !allowedTargetTypes[target.Type] {
		return "invalid_target", "target.type must be national, region, district, radius, community, or custom"
	}
	if target.Type != "national" && len(target.IDs) == 0 {
		return "missing_target_ids", "target.ids are required unless target.type is national"
	}
	if strings.TrimSpace(target.Label) == "" {
		return "missing_target_label", "target.label is required"
	}
	switch target.Type {
	case "region", "district", "community":
		for _, id := range target.IDs {
			if _, ok := targetCatalog[target.Type+":"+id]; !ok {
				return "unknown_target_id", "target.ids must match a supported region, district, or community"
			}
		}
		if target.Geometry == nil || target.Center == nil {
			return "missing_target_geometry", "target geometry could not be resolved"
		}
	case "radius":
		if target.Center == nil || !validCoordinates(*target.Center) {
			return "invalid_target_center", "radius targets require a valid center"
		}
		if target.RadiusMeters < 250 || target.RadiusMeters > 100000 {
			return "invalid_target_radius", "radiusMeters must be between 250 and 100000"
		}
	case "custom":
		if target.Geometry == nil || !validPolygonGeometry(*target.Geometry) {
			return "invalid_target_geometry", "custom targets require a closed polygon geometry"
		}
		if target.AreaSqKm <= 0 || target.AreaSqKm > 50000 {
			return "invalid_target_area", "custom target area must be greater than 0 and at most 50000 square kilometers"
		}
	}
	return "", ""
}

func applyCatalogTarget(target *alertTarget) {
	records := make([]targetCatalogRecord, 0, len(target.IDs))
	for _, id := range target.IDs {
		record, ok := targetCatalog[target.Type+":"+id]
		if !ok {
			continue
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return
	}

	if target.Label == "" {
		labels := make([]string, 0, len(records))
		for _, record := range records {
			labels = append(labels, record.Label)
		}
		target.Label = strings.Join(labels, ", ")
	}

	lat := 0.0
	lng := 0.0
	area := 0.0
	population := 0
	for _, record := range records {
		lat += record.Center.Lat
		lng += record.Center.Lng
		area += record.AreaSqKm
		population += record.EstimatedPopulation
	}
	target.Center = &coordinates{Lat: roundCoordinate(lat / float64(len(records))), Lng: roundCoordinate(lng / float64(len(records)))}
	target.RadiusMeters = maxCatalogRadius(records)
	target.AreaSqKm = roundArea(area)
	target.EstimatedPopulation = population
	target.Geometry = geometryFromCatalogRecords(records)
}

func normalizeCenter(center *coordinates) *coordinates {
	if center == nil {
		return nil
	}
	return &coordinates{Lat: roundCoordinate(center.Lat), Lng: roundCoordinate(center.Lng)}
}

func normalizeGeometry(geometry *targetGeometry) *targetGeometry {
	if geometry == nil {
		return nil
	}
	return &targetGeometry{
		Type:        strings.TrimSpace(geometry.Type),
		Coordinates: geometry.Coordinates,
	}
}

func normalizeTargetIDs(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = normalizeQueryValue(value)
		if value != "" {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func geometryFromCatalogRecords(records []targetCatalogRecord) *targetGeometry {
	if len(records) == 0 {
		return nil
	}

	minLat := 90.0
	minLng := 180.0
	maxLat := -90.0
	maxLng := -180.0
	for _, record := range records {
		deltaLat, deltaLng := degreeDeltas(record.Center, record.RadiusMeters)
		minLat = math.Min(minLat, record.Center.Lat-deltaLat)
		maxLat = math.Max(maxLat, record.Center.Lat+deltaLat)
		minLng = math.Min(minLng, record.Center.Lng-deltaLng)
		maxLng = math.Max(maxLng, record.Center.Lng+deltaLng)
	}
	return geometryFromBounds(minLat, minLng, maxLat, maxLng)
}

func geometryFromBounds(minLat float64, minLng float64, maxLat float64, maxLng float64) *targetGeometry {
	return &targetGeometry{
		Type: "Polygon",
		Coordinates: [][][]float64{{
			{roundCoordinate(minLng), roundCoordinate(minLat)},
			{roundCoordinate(maxLng), roundCoordinate(minLat)},
			{roundCoordinate(maxLng), roundCoordinate(maxLat)},
			{roundCoordinate(minLng), roundCoordinate(maxLat)},
			{roundCoordinate(minLng), roundCoordinate(minLat)},
		}},
	}
}

func maxCatalogRadius(records []targetCatalogRecord) float64 {
	maxRadius := 0.0
	for _, record := range records {
		if record.RadiusMeters > maxRadius {
			maxRadius = record.RadiusMeters
		}
	}
	return roundMeters(maxRadius)
}

func degreeDeltas(center coordinates, radiusMeters float64) (float64, float64) {
	latDelta := radiusMeters / 111320
	lngDelta := radiusMeters / (111320 * math.Cos(center.Lat*math.Pi/180))
	if math.IsInf(lngDelta, 0) || math.IsNaN(lngDelta) {
		lngDelta = latDelta
	}
	return latDelta, lngDelta
}

func validCoordinates(center coordinates) bool {
	return center.Lat >= -90 && center.Lat <= 90 && center.Lng >= -180 && center.Lng <= 180
}

func validPolygonGeometry(geometry targetGeometry) bool {
	if geometry.Type != "Polygon" || len(geometry.Coordinates) != 1 {
		return false
	}
	ring := geometry.Coordinates[0]
	if len(ring) < 4 {
		return false
	}
	first := ring[0]
	last := ring[len(ring)-1]
	if len(first) != 2 || len(last) != 2 || first[0] != last[0] || first[1] != last[1] {
		return false
	}
	for _, point := range ring {
		if len(point) != 2 {
			return false
		}
		if !validCoordinates(coordinates{Lat: point[1], Lng: point[0]}) {
			return false
		}
	}
	return true
}

func polygonCenter(geometry *targetGeometry) *coordinates {
	if geometry == nil || len(geometry.Coordinates) == 0 || len(geometry.Coordinates[0]) == 0 {
		return nil
	}
	ring := geometry.Coordinates[0]
	lat := 0.0
	lng := 0.0
	count := 0
	for index, point := range ring {
		if index == len(ring)-1 {
			continue
		}
		if len(point) != 2 {
			return nil
		}
		lat += point[1]
		lng += point[0]
		count++
	}
	if count == 0 {
		return nil
	}
	return &coordinates{Lat: roundCoordinate(lat / float64(count)), Lng: roundCoordinate(lng / float64(count))}
}

func polygonAreaSqKm(geometry *targetGeometry) float64 {
	if geometry == nil || len(geometry.Coordinates) == 0 || len(geometry.Coordinates[0]) < 4 {
		return 0
	}
	center := polygonCenter(geometry)
	if center == nil {
		return 0
	}

	ring := geometry.Coordinates[0]
	sum := 0.0
	for index := 0; index < len(ring)-1; index++ {
		if len(ring[index]) != 2 || len(ring[index+1]) != 2 {
			return 0
		}
		x1, y1 := lonLatToMeters(ring[index][0], ring[index][1], center.Lat)
		x2, y2 := lonLatToMeters(ring[index+1][0], ring[index+1][1], center.Lat)
		sum += x1*y2 - x2*y1
	}
	return roundArea(math.Abs(sum) / 2 / 1000000)
}

func lonLatToMeters(lng float64, lat float64, referenceLat float64) (float64, float64) {
	x := lng * 111320 * math.Cos(referenceLat*math.Pi/180)
	y := lat * 110540
	return x, y
}

func targetSummary(target alertTarget) string {
	switch target.Type {
	case "radius":
		return fmt.Sprintf("%s radius target, approximately %.1f sq km and %d people.", metersLabel(target.RadiusMeters), target.AreaSqKm, target.EstimatedPopulation)
	case "custom":
		return fmt.Sprintf("Custom polygon target, approximately %.1f sq km and %d people.", target.AreaSqKm, target.EstimatedPopulation)
	default:
		return fmt.Sprintf("%s target covering approximately %.1f sq km and %d people.", target.Label, target.AreaSqKm, target.EstimatedPopulation)
	}
}

func targetWarnings(target alertTarget) []string {
	warnings := []string{}
	if target.Type == "national" {
		warnings = append(warnings, "National alerts should be reserved for broad life-safety threats.")
	}
	if target.AreaSqKm > 1000 {
		warnings = append(warnings, "Large target area may increase alert fatigue; confirm scope before approval.")
	}
	if target.Type == "custom" {
		warnings = append(warnings, "Custom geometry should be reviewed against official district boundaries before publishing.")
	}
	return warnings
}

func metersLabel(value float64) string {
	if value >= 1000 {
		return fmt.Sprintf("%.1f km", value/1000)
	}
	return fmt.Sprintf("%.0f m", value)
}

func roundMeters(value float64) float64 {
	return math.Round(value)
}

func roundCoordinate(value float64) float64 {
	return math.Round(value*1000000) / 1000000
}

func roundArea(value float64) float64 {
	return math.Round(value*10) / 10
}

func roundProbability(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func snapshotAlert(alert authorityAlert) map[string]any {
	return map[string]any{
		"id":                alert.ID,
		"title":             alert.Title,
		"hazardType":        alert.HazardType,
		"severity":          alert.Severity,
		"target":            alert.Target,
		"issuingAgencyId":   alert.IssuingAgencyID,
		"issuedBy":          alert.IssuedBy,
		"approvedBy":        alert.ApprovedBy,
		"rejectedBy":        alert.RejectedBy,
		"status":            alert.Status,
		"emergencyOverride": alert.EmergencyOverride,
		"statusReason":      alert.StatusReason,
		"sourcePrediction":  alert.SourcePrediction,
	}
}

func validAlertStatus(status string) bool {
	switch status {
	case "draft", "submitted", "approved", "rejected", "published", "expired", "cancelled":
		return true
	default:
		return false
	}
}

func statusForCode(code string) int {
	switch code {
	case "not_found":
		return http.StatusNotFound
	case "forbidden", "separation_of_duties":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func optionalDecodeJSON(r *http.Request, target any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	return decodeJSON(r, target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NADAA-Actor-ID, X-NADAA-Actor-Role, X-NADAA-Agency-ID, X-NADAA-MFA-Completed, X-NADAA-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
