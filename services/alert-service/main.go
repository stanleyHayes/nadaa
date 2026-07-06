package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	HazardType         string      `json:"hazardType"`
	Severity           string      `json:"severity"`
	Message            string      `json:"message"`
	Target             alertTarget `json:"target"`
	StartsAt           time.Time   `json:"startsAt"`
	ExpiresAt          time.Time   `json:"expiresAt"`
	RecommendedAction  string      `json:"recommendedAction"`
	EvacuationRequired bool        `json:"evacuationRequired"`
	ShelterIDs         []string    `json:"shelterIds"`
	IssuingAgencyID    string      `json:"issuingAgencyId"`
	IssuedBy           string      `json:"issuedBy"`
	ApprovedBy         string      `json:"approvedBy,omitempty"`
	RejectedBy         string      `json:"rejectedBy,omitempty"`
	Status             string      `json:"status"`
	EmergencyOverride  bool        `json:"emergencyOverride"`
	StatusReason       string      `json:"statusReason,omitempty"`
	CreatedAt          time.Time   `json:"createdAt"`
	UpdatedAt          time.Time   `json:"updatedAt"`
	SubmittedAt        *time.Time  `json:"submittedAt,omitempty"`
	ApprovedAt         *time.Time  `json:"approvedAt,omitempty"`
	RejectedAt         *time.Time  `json:"rejectedAt,omitempty"`
}

type alertTarget struct {
	Type  string   `json:"type"`
	IDs   []string `json:"ids"`
	Label string   `json:"label"`
}

type createAlertRequest struct {
	Title              string      `json:"title"`
	HazardType         string      `json:"hazardType"`
	Severity           string      `json:"severity"`
	Message            string      `json:"message"`
	Target             alertTarget `json:"target"`
	StartsAt           time.Time   `json:"startsAt"`
	ExpiresAt          time.Time   `json:"expiresAt"`
	RecommendedAction  string      `json:"recommendedAction"`
	EvacuationRequired bool        `json:"evacuationRequired"`
	ShelterIDs         []string    `json:"shelterIds"`
}

type workflowRequest struct {
	Note   string `json:"note"`
	Reason string `json:"reason"`
}

type alertListResponse struct {
	Alerts []authorityAlert `json:"alerts"`
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

func main() {
	srv := &server{store: newMemoryStore()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/alerts", srv.createAlertHandler)
	mux.HandleFunc("GET /api/v1/alerts", srv.listAlertsHandler)
	mux.HandleFunc("GET /api/v1/alerts/audit", srv.listAuditHandler)
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

	if code, message := validateAlertRequest(request); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	alert := s.store.createAlert(request, ctx, time.Now().UTC())
	writeJSON(w, http.StatusCreated, alert)
}

func (s *server) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	status := normalizeQueryValue(r.URL.Query().Get("status"))
	currentOnly := normalizeQueryValue(r.URL.Query().Get("current")) == "true"
	if status != "" && !validAlertStatus(status) {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be draft, submitted, approved, rejected, published, expired, or cancelled")
		return
	}

	publicOnly := !hasAuthorityHeaders(r)
	writeJSON(w, http.StatusOK, alertListResponse{Alerts: s.store.listAlerts(status, currentOnly, publicOnly, time.Now().UTC())})
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
	if code, message := validateAlertRequest(request); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	alert, code, message := s.store.updateAlert(r.PathValue("id"), request, ctx, time.Now().UTC())
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

func (m *memoryStore) listAlerts(status string, currentOnly bool, publicOnly bool, now time.Time) []authorityAlert {
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
		alerts = append(alerts, alert)
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

func validateAlertRequest(request createAlertRequest) (string, string) {
	title := strings.TrimSpace(request.Title)
	message := strings.TrimSpace(request.Message)
	action := strings.TrimSpace(request.RecommendedAction)
	hazard := normalizeQueryValue(request.HazardType)
	severity := normalizeQueryValue(request.Severity)
	target := normalizeTarget(request.Target)

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
	if !allowedTargetTypes[target.Type] {
		return "invalid_target", "target.type must be national, region, district, radius, community, or custom"
	}
	if target.Type != "national" && len(target.IDs) == 0 {
		return "missing_target_ids", "target.ids are required unless target.type is national"
	}
	if strings.TrimSpace(target.Label) == "" {
		return "missing_target_label", "target.label is required"
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

	return "", ""
}

func normalizeTarget(target alertTarget) alertTarget {
	return alertTarget{
		Type:  normalizeQueryValue(target.Type),
		IDs:   compactStrings(target.IDs),
		Label: strings.TrimSpace(target.Label),
	}
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
