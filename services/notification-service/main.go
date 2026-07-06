package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type server struct {
	store       *memoryStore
	alertClient *alertServiceClient
	providers   map[string]notificationProvider
	now         func() time.Time
}

type memoryStore struct {
	mu           sync.RWMutex
	alerts       []citizenAlert
	deliveryLogs []deliveryAttempt
	nextLogID    int
}

type citizenAlert struct {
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	HazardType         string      `json:"hazardType"`
	Severity           string      `json:"severity"`
	Message            string      `json:"message"`
	Target             alertTarget `json:"target"`
	TargetLabel        string      `json:"targetLabel"`
	StartsAt           time.Time   `json:"startsAt"`
	ExpiresAt          time.Time   `json:"expiresAt"`
	Status             string      `json:"status"`
	RecommendedAction  string      `json:"recommendedAction"`
	EvacuationRequired bool        `json:"evacuationRequired"`
	ShelterIDs         []string    `json:"shelterIds"`
	Source             string      `json:"source"`
	UpdatedAt          time.Time   `json:"updatedAt"`
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
	Status             string      `json:"status"`
	UpdatedAt          time.Time   `json:"updatedAt"`
}

type alertTarget struct {
	Type                string       `json:"type"`
	IDs                 []string     `json:"ids"`
	Label               string       `json:"label"`
	Center              *coordinates `json:"center,omitempty"`
	RadiusMeters        float64      `json:"radiusMeters,omitempty"`
	AreaSqKm            float64      `json:"areaSqKm,omitempty"`
	EstimatedPopulation int          `json:"estimatedPopulation,omitempty"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type citizenAlertListResponse struct {
	Alerts      []citizenAlert `json:"alerts"`
	GeneratedAt time.Time      `json:"generatedAt"`
	Source      string         `json:"source"`
}

type authorityAlertListResponse struct {
	Alerts []authorityAlert `json:"alerts"`
}

type deliveryRequest struct {
	AlertID     string   `json:"alertId,omitempty"`
	RecipientID string   `json:"recipientId"`
	Phone       string   `json:"phone,omitempty"`
	PushToken   string   `json:"pushToken,omitempty"`
	Language    string   `json:"language,omitempty"`
	Channels    []string `json:"channels"`
	DryRun      bool     `json:"dryRun,omitempty"`
}

type deliveryResponse struct {
	Attempts []deliveryAttempt `json:"attempts"`
}

type deliveryAttempt struct {
	ID           string    `json:"id"`
	AlertID      string    `json:"alertId"`
	AlertTitle   string    `json:"alertTitle"`
	Channel      string    `json:"channel"`
	Provider     string    `json:"provider"`
	RecipientRef string    `json:"recipientRef"`
	Status       string    `json:"status"`
	Reason       string    `json:"reason,omitempty"`
	MessageID    string    `json:"messageId,omitempty"`
	AttemptedAt  time.Time `json:"attemptedAt"`
}

type deliveryLogListResponse struct {
	Logs []deliveryAttempt `json:"logs"`
}

type providerMessage struct {
	Alert       citizenAlert
	Request     deliveryRequest
	Channel     string
	Recipient   string
	AttemptedAt time.Time
}

type providerResult struct {
	Provider  string
	Status    string
	Reason    string
	MessageID string
}

type notificationProvider interface {
	Send(context.Context, providerMessage) providerResult
}

type mockProvider struct {
	channel string
}

type disabledProvider struct {
	channel string
	reason  string
}

type alertServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

type alertFeedFilters struct {
	Hazard         string
	Severity       string
	Status         string
	IncludeExpired bool
	TargetType     string
	TargetID       string
}

type logFilters struct {
	AlertID string
	Channel string
	Status  string
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

var allowedChannels = map[string]bool{
	"push": true,
	"sms":  true,
}

var allowedDeliveryStatuses = map[string]bool{
	"queued":    true,
	"delivered": true,
	"failed":    true,
	"skipped":   true,
}

var allowedFeedStatuses = map[string]bool{
	"current":  true,
	"expired":  true,
	"upcoming": true,
	"all":      true,
}

func main() {
	srv := newServerFromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/notifications/alerts", srv.listAlertsHandler)
	mux.HandleFunc("POST /api/v1/notifications/alerts/{id}/deliver", srv.deliverAlertHandler)
	mux.HandleFunc("GET /api/v1/notifications/delivery-logs", srv.listDeliveryLogsHandler)

	addr := envOrDefault("NADAA_NOTIFICATION_ADDR", ":8090")
	log.Printf("notification-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServerFromEnv() *server {
	now := time.Now().UTC()
	return &server{
		store:       newMemoryStore(now),
		alertClient: newAlertServiceClient(envOrDefault("NADAA_ALERT_SERVICE_URL", "http://localhost:8089/api/v1")),
		providers:   providersFromEnv(),
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func newMemoryStore(now time.Time) *memoryStore {
	return &memoryStore{
		alerts:    seedCitizenAlerts(now),
		nextLogID: 1,
	}
}

func newAlertServiceClient(rawBaseURL string) *alertServiceClient {
	rawBaseURL = strings.TrimSpace(rawBaseURL)
	if rawBaseURL == "" {
		return nil
	}
	return &alertServiceClient{
		baseURL:    strings.TrimRight(rawBaseURL, "/"),
		httpClient: &http.Client{Timeout: 2 * time.Second},
	}
}

func providersFromEnv() map[string]notificationProvider {
	providers := map[string]notificationProvider{}

	if envBool("NADAA_PUSH_ENABLED", true) {
		providers["push"] = mockProvider{channel: "push"}
	} else {
		providers["push"] = disabledProvider{channel: "push", reason: "push provider disabled"}
	}

	if envBool("NADAA_SMS_ENABLED", true) {
		providers["sms"] = mockProvider{channel: "sms"}
	} else {
		providers["sms"] = disabledProvider{channel: "sms", reason: "sms provider disabled"}
	}

	return providers
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "notification-service"})
}

func (s *server) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseAlertFeedFilters(r)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now()
	alerts, source := s.listCitizenAlerts(r.Context(), filters, now)
	writeJSON(w, http.StatusOK, citizenAlertListResponse{Alerts: alerts, GeneratedAt: now, Source: source})
}

func (s *server) deliverAlertHandler(w http.ResponseWriter, r *http.Request) {
	id := normalizeID(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing_alert_id", "alert id is required")
		return
	}

	var request deliveryRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.AlertID = id
	request.Channels = normalizeChannels(request.Channels)
	request.Language = normalizeLanguage(request.Language)

	if code, message := validateDeliveryRequest(request); code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now()
	alerts, _ := s.listCitizenAlerts(r.Context(), alertFeedFilters{Status: "all"}, now)
	alert, ok := findAlert(alerts, id)
	if !ok {
		writeError(w, http.StatusNotFound, "alert_not_found", "alert was not found in the citizen feed")
		return
	}

	attempts := s.store.createDeliveryAttempts(r.Context(), alert, request, s.providers, now)
	writeJSON(w, http.StatusAccepted, deliveryResponse{Attempts: attempts})
}

func (s *server) listDeliveryLogsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseLogFilters(r)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	writeJSON(w, http.StatusOK, deliveryLogListResponse{Logs: s.store.listDeliveryLogs(filters)})
}

func (s *server) listCitizenAlerts(ctx context.Context, filters alertFeedFilters, now time.Time) ([]citizenAlert, string) {
	fixtureAlerts := s.store.listAlerts(filters, now)
	combined := make([]citizenAlert, 0, len(fixtureAlerts))
	seen := map[string]bool{}
	for _, alert := range fixtureAlerts {
		combined = append(combined, alert)
		seen[alert.ID] = true
	}

	source := "fixture"
	if s.alertClient != nil {
		upstreamAlerts, err := s.alertClient.listAlerts(ctx, now)
		if err == nil {
			for _, alert := range upstreamAlerts {
				if seen[alert.ID] {
					continue
				}
				if alertMatchesFilters(alert, filters, now) {
					combined = append(combined, alert)
					seen[alert.ID] = true
				}
			}
			if len(upstreamAlerts) > 0 {
				source = "alert-service+fixture"
			}
		}
	}

	sortCitizenAlerts(combined)
	return combined, source
}

func (c *alertServiceClient) listAlerts(ctx context.Context, now time.Time) ([]citizenAlert, error) {
	parsed, err := url.Parse(c.baseURL + "/alerts")
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("alert-service returned %d", response.StatusCode)
	}

	var payload authorityAlertListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	alerts := make([]citizenAlert, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		if alert.Status != "approved" && alert.Status != "published" {
			continue
		}
		alerts = append(alerts, citizenAlert{
			ID:                 alert.ID,
			Title:              alert.Title,
			HazardType:         alert.HazardType,
			Severity:           alert.Severity,
			Message:            alert.Message,
			Target:             alert.Target,
			TargetLabel:        alert.Target.Label,
			StartsAt:           alert.StartsAt,
			ExpiresAt:          alert.ExpiresAt,
			Status:             alertFeedStatus(alert.StartsAt, alert.ExpiresAt, now),
			RecommendedAction:  alert.RecommendedAction,
			EvacuationRequired: alert.EvacuationRequired,
			ShelterIDs:         alert.ShelterIDs,
			Source:             "alert-service",
			UpdatedAt:          alert.UpdatedAt,
		})
	}

	return alerts, nil
}

func parseAlertFeedFilters(r *http.Request) (alertFeedFilters, string, string) {
	query := r.URL.Query()
	filters := alertFeedFilters{
		Hazard:         normalizeQueryValue(query.Get("hazard")),
		Severity:       normalizeQueryValue(query.Get("severity")),
		Status:         normalizeQueryValue(query.Get("status")),
		TargetType:     normalizeQueryValue(query.Get("targetType")),
		TargetID:       normalizeID(query.Get("targetId")),
		IncludeExpired: normalizeQueryValue(query.Get("includeExpired")) == "true",
	}

	if filters.Hazard != "" && !allowedHazards[filters.Hazard] {
		return alertFeedFilters{}, "invalid_hazard", "hazard must be a supported NADAA hazard type"
	}
	if filters.Severity != "" && !allowedSeverities[filters.Severity] {
		return alertFeedFilters{}, "invalid_severity", "severity must be advisory, watch, warning, severe_warning, or emergency"
	}
	if filters.Status != "" && !allowedFeedStatuses[filters.Status] {
		return alertFeedFilters{}, "invalid_status", "status must be current, expired, upcoming, or all"
	}
	if filters.TargetType != "" && !allowedTargetTypes[filters.TargetType] {
		return alertFeedFilters{}, "invalid_target_type", "targetType must be national, region, district, radius, community, or custom"
	}

	return filters, "", ""
}

func parseLogFilters(r *http.Request) (logFilters, string, string) {
	query := r.URL.Query()
	filters := logFilters{
		AlertID: normalizeID(query.Get("alertId")),
		Channel: normalizeQueryValue(query.Get("channel")),
		Status:  normalizeQueryValue(query.Get("status")),
	}
	if filters.Channel != "" && !allowedChannels[filters.Channel] {
		return logFilters{}, "invalid_channel", "channel must be push or sms"
	}
	if filters.Status != "" && !allowedDeliveryStatuses[filters.Status] {
		return logFilters{}, "invalid_status", "status must be queued, delivered, failed, or skipped"
	}
	return filters, "", ""
}

func validateDeliveryRequest(request deliveryRequest) (string, string) {
	if request.RecipientID == "" && request.Phone == "" && request.PushToken == "" {
		return "missing_recipient", "recipientId, phone, or pushToken is required"
	}
	if len(request.Channels) == 0 {
		return "missing_channels", "channels must include push, sms, or both"
	}
	for _, channel := range request.Channels {
		if !allowedChannels[channel] {
			return "invalid_channel", "channels must include only push or sms"
		}
		if channel == "push" && request.PushToken == "" && request.RecipientID == "" {
			return "missing_push_target", "push delivery requires pushToken or recipientId"
		}
		if channel == "sms" && request.Phone == "" {
			return "missing_sms_phone", "sms delivery requires phone"
		}
	}
	return "", ""
}

func (m *memoryStore) listAlerts(filters alertFeedFilters, now time.Time) []citizenAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]citizenAlert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alert.Status = alertFeedStatus(alert.StartsAt, alert.ExpiresAt, now)
		if alertMatchesFilters(alert, filters, now) {
			alerts = append(alerts, alert)
		}
	}
	sortCitizenAlerts(alerts)
	return alerts
}

func (m *memoryStore) createDeliveryAttempts(ctx context.Context, alert citizenAlert, request deliveryRequest, providers map[string]notificationProvider, now time.Time) []deliveryAttempt {
	m.mu.Lock()
	defer m.mu.Unlock()

	attempts := make([]deliveryAttempt, 0, len(request.Channels))
	for _, channel := range request.Channels {
		provider := providers[channel]
		result := provider.Send(ctx, providerMessage{
			Alert:       alert,
			Request:     request,
			Channel:     channel,
			Recipient:   recipientRef(request, channel),
			AttemptedAt: now,
		})
		attempt := deliveryAttempt{
			ID:           fmt.Sprintf("delivery_%06d", m.nextLogID),
			AlertID:      alert.ID,
			AlertTitle:   alert.Title,
			Channel:      channel,
			Provider:     result.Provider,
			RecipientRef: recipientRef(request, channel),
			Status:       result.Status,
			Reason:       result.Reason,
			MessageID:    result.MessageID,
			AttemptedAt:  now,
		}
		m.nextLogID++
		m.deliveryLogs = append(m.deliveryLogs, attempt)
		attempts = append(attempts, attempt)
	}

	return attempts
}

func (m *memoryStore) listDeliveryLogs(filters logFilters) []deliveryAttempt {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := make([]deliveryAttempt, 0, len(m.deliveryLogs))
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

func (p mockProvider) Send(_ context.Context, message providerMessage) providerResult {
	providerID := "mock_push"
	if p.channel == "sms" {
		providerID = "mock_sms"
	}
	return providerResult{
		Provider:  providerID,
		Status:    "delivered",
		MessageID: fmt.Sprintf("%s_%s_%d", providerID, message.Alert.ID, message.AttemptedAt.Unix()),
	}
}

func (p disabledProvider) Send(_ context.Context, _ providerMessage) providerResult {
	providerID := p.channel + "_disabled"
	return providerResult{
		Provider: providerID,
		Status:   "skipped",
		Reason:   p.reason,
	}
}

func alertMatchesFilters(alert citizenAlert, filters alertFeedFilters, now time.Time) bool {
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
	if filters.TargetID != "" && !containsString(alert.Target.IDs, filters.TargetID) {
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

func alertFeedStatus(startsAt time.Time, expiresAt time.Time, now time.Time) string {
	if now.Before(startsAt) {
		return "upcoming"
	}
	if !expiresAt.After(now) {
		return "expired"
	}
	return "current"
}

func sortCitizenAlerts(alerts []citizenAlert) {
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

func findAlert(alerts []citizenAlert, id string) (citizenAlert, bool) {
	for _, alert := range alerts {
		if alert.ID == id {
			return alert, true
		}
	}
	return citizenAlert{}, false
}

func seedCitizenAlerts(now time.Time) []citizenAlert {
	return []citizenAlert{
		newCitizenAlert(
			"alert_feed_current_flood",
			"Severe flood warning",
			"flood",
			"severe_warning",
			"Heavy rainfall and rising drains may flood low-lying parts of Accra Metro and Tema.",
			alertTarget{Type: "district", IDs: []string{"accra-metropolitan", "tema-metropolitan"}, Label: "Accra Metro and Tema"},
			now.Add(-30*time.Minute),
			now.Add(5*time.Hour),
			"Move away from drains, avoid flooded roads, and prepare to go to a shelter if directed.",
			true,
			[]string{"shelter-ama-001", "shelter-osu-002"},
			"fixture",
			now.Add(-20*time.Minute),
			now,
		),
		newCitizenAlert(
			"alert_feed_current_fire",
			"Market fire watch",
			"fire",
			"watch",
			"Responders are monitoring dense market areas after smoke reports near electrical kiosks.",
			alertTarget{Type: "community", IDs: []string{"accra-central"}, Label: "Accra Central"},
			now.Add(-20*time.Minute),
			now.Add(3*time.Hour),
			"Keep access lanes open, avoid overloaded sockets, and call 112 if you see flames or heavy smoke.",
			false,
			nil,
			"fixture",
			now.Add(-15*time.Minute),
			now,
		),
		newCitizenAlert(
			"alert_feed_expired_road",
			"Road hazard resolved",
			"road_crash",
			"advisory",
			"Earlier congestion near Kaneshie Market Road has cleared after responders reopened the lane.",
			alertTarget{Type: "radius", IDs: []string{"kaneshie-market-road"}, Label: "Kaneshie Market Road", Center: &coordinates{Lat: 5.566, Lng: -0.242}, RadiusMeters: 1500},
			now.Add(-8*time.Hour),
			now.Add(-2*time.Hour),
			"Continue to drive carefully and give way to emergency vehicles.",
			false,
			nil,
			"fixture",
			now.Add(-2*time.Hour),
			now,
		),
	}
}

func newCitizenAlert(id string, title string, hazard string, severity string, message string, target alertTarget, startsAt time.Time, expiresAt time.Time, action string, evacuation bool, shelters []string, source string, updatedAt time.Time, now time.Time) citizenAlert {
	return citizenAlert{
		ID:                 id,
		Title:              title,
		HazardType:         hazard,
		Severity:           severity,
		Message:            message,
		Target:             target,
		TargetLabel:        target.Label,
		StartsAt:           startsAt,
		ExpiresAt:          expiresAt,
		Status:             alertFeedStatus(startsAt, expiresAt, now),
		RecommendedAction:  action,
		EvacuationRequired: evacuation,
		ShelterIDs:         shelters,
		Source:             source,
		UpdatedAt:          updatedAt,
	}
}

func normalizeChannels(channels []string) []string {
	if len(channels) == 0 {
		return []string{"push", "sms"}
	}
	result := make([]string, 0, len(channels))
	seen := map[string]bool{}
	for _, channel := range channels {
		channel = normalizeQueryValue(channel)
		if channel == "" || seen[channel] {
			continue
		}
		seen[channel] = true
		result = append(result, channel)
	}
	return result
}

func recipientRef(request deliveryRequest, channel string) string {
	if request.RecipientID != "" {
		return request.RecipientID
	}
	if channel == "sms" && request.Phone != "" {
		if len(request.Phone) <= 4 {
			return "phone:" + request.Phone
		}
		return "phone:..." + request.Phone[len(request.Phone)-4:]
	}
	if channel == "push" && request.PushToken != "" {
		if len(request.PushToken) <= 6 {
			return "push:" + request.PushToken
		}
		return "push:..." + request.PushToken[len(request.PushToken)-6:]
	}
	return "anonymous"
}

func normalizeQueryValue(value string) string {
	return strings.Trim(strings.ToLower(strings.ReplaceAll(value, " ", "_")), "_")
}

func normalizeID(value string) string {
	return strings.TrimSpace(value)
}

func normalizeLanguage(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "en"
	}
	return value
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := normalizeQueryValue(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on", "enabled", "mock":
		return true
	case "0", "false", "no", "off", "disabled":
		return false
	default:
		return fallback
	}
}
