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
	store          *memoryStore
	alertClient    *alertServiceClient
	incidentClient *incidentServiceClient
	providers      map[string]notificationProvider
	now            func() time.Time
}

type memoryStore struct {
	mu                         sync.RWMutex
	alerts                     []citizenAlert
	deliveryLogs               []deliveryAttempt
	accessLogs                 []inclusiveAccessLog
	accessReports              []inclusiveAccessReport
	voiceAlerts                []voiceAlertAsset
	whatsappConversations      map[string]whatsappConversation
	whatsappTranscripts        []whatsappTranscript
	nextLogID                  int
	nextAccessLogID            int
	nextAccessReportID         int
	nextVoiceAlertID           int
	nextVoiceVariantID         int
	nextWhatsAppConversationID int
	nextWhatsAppTranscriptID   int
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
	VoiceAssetID string    `json:"voiceAssetId,omitempty"`
	Language     string    `json:"language,omitempty"`
	AudioURL     string    `json:"audioUrl,omitempty"`
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

type incidentServiceClient struct {
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

type accessLogFilters struct {
	Channel string
	Intent  string
	Status  string
}

type ussdWebhookRequest struct {
	SessionID         string       `json:"sessionId"`
	Phone             string       `json:"phone"`
	ServiceCode       string       `json:"serviceCode,omitempty"`
	Text              string       `json:"text"`
	Language          string       `json:"language,omitempty"`
	Network           string       `json:"network,omitempty"`
	Provider          string       `json:"provider,omitempty"`
	ProviderMessageID string       `json:"providerMessageId,omitempty"`
	ProviderError     string       `json:"providerError,omitempty"`
	ProfileID         string       `json:"profileId,omitempty"`
	LinkProfile       bool         `json:"linkProfile,omitempty"`
	Location          *coordinates `json:"location,omitempty"`
}

type ussdWebhookResponse struct {
	SessionID string                 `json:"sessionId"`
	Action    string                 `json:"action"`
	Message   string                 `json:"message"`
	Language  string                 `json:"language"`
	Log       inclusiveAccessLog     `json:"log"`
	Report    *inclusiveAccessReport `json:"report,omitempty"`
}

type smsInboundRequest struct {
	From              string       `json:"from"`
	Body              string       `json:"body"`
	Language          string       `json:"language,omitempty"`
	Provider          string       `json:"provider,omitempty"`
	ProviderMessageID string       `json:"providerMessageId,omitempty"`
	ProviderError     string       `json:"providerError,omitempty"`
	ProfileID         string       `json:"profileId,omitempty"`
	LinkProfile       bool         `json:"linkProfile,omitempty"`
	Location          *coordinates `json:"location,omitempty"`
}

type smsInboundResponse struct {
	Message string                 `json:"message"`
	Log     inclusiveAccessLog     `json:"log"`
	Report  *inclusiveAccessReport `json:"report,omitempty"`
}

type whatsappInboundRequest struct {
	From              string          `json:"from"`
	Body              string          `json:"body"`
	Language          string          `json:"language,omitempty"`
	Provider          string          `json:"provider,omitempty"`
	ProviderMessageID string          `json:"providerMessageId,omitempty"`
	ProviderError     string          `json:"providerError,omitempty"`
	ProfileID         string          `json:"profileId,omitempty"`
	LinkProfile       bool            `json:"linkProfile,omitempty"`
	Location          *coordinates    `json:"location,omitempty"`
	Media             []whatsappMedia `json:"media,omitempty"`
}

type whatsappMedia struct {
	ID          string `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Caption     string `json:"caption,omitempty"`
}

type whatsappInboundResponse struct {
	Message       string                 `json:"message"`
	Conversation  whatsappConversation   `json:"conversation"`
	Log           inclusiveAccessLog     `json:"log"`
	Report        *inclusiveAccessReport `json:"report,omitempty"`
	TranscriptIDs []string               `json:"transcriptIds,omitempty"`
}

type whatsappConversation struct {
	ID                 string    `json:"id"`
	Key                string    `json:"-"`
	Channel            string    `json:"channel"`
	PhoneRef           string    `json:"phoneRef"`
	ProfileID          string    `json:"profileId,omitempty"`
	LinkedProfile      bool      `json:"linkedProfile"`
	Language           string    `json:"language"`
	Intent             string    `json:"intent"`
	State              string    `json:"state"`
	Hazard             string    `json:"hazard,omitempty"`
	Urgency            string    `json:"urgency,omitempty"`
	LastMessageSummary string    `json:"lastMessageSummary,omitempty"`
	LastMediaSummary   string    `json:"lastMediaSummary,omitempty"`
	StartedAt          time.Time `json:"startedAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	ExpiresAt          time.Time `json:"expiresAt"`
	RetentionUntil     time.Time `json:"retentionUntil"`
}

type whatsappTranscript struct {
	ID                string    `json:"id"`
	ConversationID    string    `json:"conversationId"`
	Provider          string    `json:"provider"`
	ProviderMessageID string    `json:"providerMessageId,omitempty"`
	PhoneRef          string    `json:"phoneRef"`
	ProfileID         string    `json:"profileId,omitempty"`
	LinkedProfile     bool      `json:"linkedProfile"`
	Direction         string    `json:"direction"`
	Intent            string    `json:"intent"`
	State             string    `json:"state"`
	MessageSummary    string    `json:"messageSummary,omitempty"`
	MediaSummary      string    `json:"mediaSummary,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	RetentionUntil    time.Time `json:"retentionUntil"`
}

type inclusiveAccessLog struct {
	ID                string    `json:"id"`
	Channel           string    `json:"channel"`
	Provider          string    `json:"provider"`
	ProviderMessageID string    `json:"providerMessageId,omitempty"`
	SessionID         string    `json:"sessionId,omitempty"`
	PhoneRef          string    `json:"phoneRef"`
	ProfileID         string    `json:"profileId,omitempty"`
	LinkedProfile     bool      `json:"linkedProfile"`
	Language          string    `json:"language"`
	Intent            string    `json:"intent"`
	Status            string    `json:"status"`
	ProviderError     string    `json:"providerError,omitempty"`
	IncidentID        string    `json:"incidentId,omitempty"`
	IncidentReference string    `json:"incidentReference,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
}

type inclusiveAccessReport struct {
	ID                string      `json:"id"`
	Channel           string      `json:"channel"`
	Type              string      `json:"type"`
	Urgency           string      `json:"urgency"`
	Description       string      `json:"description"`
	Location          coordinates `json:"location"`
	LocationLabel     string      `json:"locationLabel"`
	PhoneRef          string      `json:"phoneRef"`
	ProfileID         string      `json:"profileId,omitempty"`
	LinkedProfile     bool        `json:"linkedProfile"`
	Status            string      `json:"status"`
	Media             []string    `json:"media,omitempty"`
	IncidentID        string      `json:"incidentId,omitempty"`
	IncidentReference string      `json:"incidentReference,omitempty"`
	FailureReason     string      `json:"failureReason,omitempty"`
	CreatedAt         time.Time   `json:"createdAt"`
}

type accessLogListResponse struct {
	Logs []inclusiveAccessLog `json:"logs"`
}

type voiceAlertRequest struct {
	AlertID             string   `json:"alertId"`
	Languages           []string `json:"languages,omitempty"`
	WorkflowRequestedBy string   `json:"workflowRequestedBy,omitempty"`
	Source              string   `json:"source,omitempty"`
}

type voiceAlertResponse struct {
	Asset voiceAlertAsset `json:"asset"`
}

type voiceAlertListResponse struct {
	Assets []voiceAlertAsset `json:"assets"`
}

type voiceReviewRequest struct {
	Action    string   `json:"action"`
	Reviewer  string   `json:"reviewer"`
	Note      string   `json:"note,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

type voiceDeliveryRequest struct {
	Recipients []voiceRecipient `json:"recipients"`
	DryRun     bool             `json:"dryRun,omitempty"`
}

type voiceRecipient struct {
	RecipientID string `json:"recipientId,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Language    string `json:"language"`
}

type voiceDeliveryResponse struct {
	Attempts []deliveryAttempt `json:"attempts"`
}

type voiceAlertAsset struct {
	ID                  string         `json:"id"`
	AlertID             string         `json:"alertId"`
	AlertTitle          string         `json:"alertTitle"`
	HazardType          string         `json:"hazardType"`
	Severity            string         `json:"severity"`
	TargetLabel         string         `json:"targetLabel"`
	Status              string         `json:"status"`
	ReviewStatus        string         `json:"reviewStatus"`
	Source              string         `json:"source"`
	WorkflowRequestedBy string         `json:"workflowRequestedBy,omitempty"`
	Reviewer            string         `json:"reviewer,omitempty"`
	ReviewNote          string         `json:"reviewNote,omitempty"`
	Variants            []voiceVariant `json:"variants"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	ReviewedAt          *time.Time     `json:"reviewedAt,omitempty"`
}

type voiceVariant struct {
	ID                  string    `json:"id"`
	Language            string    `json:"language"`
	Locale              string    `json:"locale"`
	VoiceName           string    `json:"voiceName"`
	MessageText         string    `json:"messageText"`
	AudioURL            string    `json:"audioUrl"`
	DurationSeconds     int       `json:"durationSeconds"`
	Status              string    `json:"status"`
	ReviewStatus        string    `json:"reviewStatus"`
	AccessibilityChecks []string  `json:"accessibilityChecks"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type incidentIntakeRequest struct {
	Type               string       `json:"type"`
	Description        string       `json:"description"`
	Location           coordinates  `json:"location"`
	PeopleAffected     int          `json:"peopleAffected"`
	InjuriesReported   bool         `json:"injuriesReported"`
	Urgency            string       `json:"urgency"`
	Anonymous          bool         `json:"anonymous"`
	ContactPermission  bool         `json:"contactPermission"`
	AccessibilityNeeds string       `json:"accessibilityNeeds"`
	Media              []string     `json:"media"`
	Reporter           *reporterRef `json:"reporter,omitempty"`
}

type reporterRef struct {
	UserID string `json:"userId"`
	Phone  string `json:"phone,omitempty"`
}

type incidentIntakeResponse struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
	Status    string `json:"status"`
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
	"push":  true,
	"sms":   true,
	"voice": true,
}

var allowedVoiceLanguages = map[string]bool{
	"en":  true,
	"tw":  true,
	"ga":  true,
	"ee":  true,
	"dag": true,
	"ha":  true,
}

var allowedAccessChannels = map[string]bool{
	"sms":      true,
	"ussd":     true,
	"whatsapp": true,
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

var allowedAccessStatuses = map[string]bool{
	"handled":   true,
	"failed":    true,
	"queued":    true,
	"submitted": true,
}

var allowedAccessIntents = map[string]bool{
	"language_menu":     true,
	"main_menu":         true,
	"current_alerts":    true,
	"report_emergency":  true,
	"risk_check":        true,
	"emergency_guides":  true,
	"shelter_lookup":    true,
	"guidance_112":      true,
	"provider_error":    true,
	"invalid_selection": true,
}

func main() {
	srv := newServerFromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/notifications/alerts", srv.listAlertsHandler)
	mux.HandleFunc("POST /api/v1/notifications/alerts/{id}/deliver", srv.deliverAlertHandler)
	mux.HandleFunc("GET /api/v1/notifications/delivery-logs", srv.listDeliveryLogsHandler)
	mux.HandleFunc("POST /api/v1/notifications/voice-alerts", srv.createVoiceAlertHandler)
	mux.HandleFunc("GET /api/v1/notifications/voice-alerts", srv.listVoiceAlertsHandler)
	mux.HandleFunc("POST /api/v1/notifications/voice-alerts/{id}/review", srv.reviewVoiceAlertHandler)
	mux.HandleFunc("POST /api/v1/notifications/voice-alerts/{id}/deliver", srv.deliverVoiceAlertHandler)
	mux.HandleFunc("POST /api/v1/notifications/ussd", srv.ussdWebhookHandler)
	mux.HandleFunc("POST /api/v1/notifications/sms/inbound", srv.smsInboundHandler)
	mux.HandleFunc("POST /api/v1/notifications/whatsapp/inbound", srv.whatsappWebhookHandler)
	mux.HandleFunc("POST /api/v1/notifications/whatsapp/webhook", srv.whatsappWebhookHandler)
	mux.HandleFunc("GET /api/v1/notifications/access-logs", srv.listAccessLogsHandler)

	addr := envOrDefault("NADAA_NOTIFICATION_ADDR", ":8090")
	logInfo(
		"notification-service starting",
		"addr", addr,
		"alertClientConfigured", srv.alertClient != nil,
		"incidentClientConfigured", srv.incidentClient != nil,
		"pushProvider", providerName(srv.providers["push"]),
		"smsProvider", providerName(srv.providers["sms"]),
		"voiceProvider", providerName(srv.providers["voice"]),
	)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		logError("notification-service stopped", "addr", addr, "error", err)
		log.Fatal(err)
	}
}

func newServerFromEnv() *server {
	now := time.Now().UTC()
	return &server{
		store:          newMemoryStore(now),
		alertClient:    newAlertServiceClient(envOrDefault("NADAA_ALERT_SERVICE_URL", "http://localhost:8089/api/v1")),
		incidentClient: newIncidentServiceClient(os.Getenv("NADAA_INCIDENT_SERVICE_URL")),
		providers:      providersFromEnv(),
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func newMemoryStore(now time.Time) *memoryStore {
	return &memoryStore{
		alerts:                     seedCitizenAlerts(now),
		whatsappConversations:      map[string]whatsappConversation{},
		nextLogID:                  1,
		nextAccessLogID:            1,
		nextAccessReportID:         1,
		nextVoiceAlertID:           1,
		nextVoiceVariantID:         1,
		nextWhatsAppConversationID: 1,
		nextWhatsAppTranscriptID:   1,
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

func newIncidentServiceClient(rawBaseURL string) *incidentServiceClient {
	rawBaseURL = strings.TrimSpace(rawBaseURL)
	if rawBaseURL == "" {
		return nil
	}
	return &incidentServiceClient{
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

	if envBool("NADAA_VOICE_ENABLED", true) {
		providers["voice"] = mockProvider{channel: "voice"}
	} else {
		providers["voice"] = disabledProvider{channel: "voice", reason: "voice provider disabled"}
	}

	return providers
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "notification-service"})
}

func (s *server) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseAlertFeedFilters(r)
	if code != "" {
		logWarn("citizen alert list rejected", "code", code, "message", message, "query", r.URL.RawQuery)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now()
	logInfo(
		"citizen alert list requested",
		"hazard", filters.Hazard,
		"severity", filters.Severity,
		"status", filters.Status,
		"includeExpired", filters.IncludeExpired,
		"targetType", filters.TargetType,
		"targetId", filters.TargetID,
	)
	alerts, source := s.listCitizenAlerts(r.Context(), filters, now)
	logInfo("citizen alert list completed", "count", len(alerts), "source", source)
	writeJSON(w, http.StatusOK, citizenAlertListResponse{Alerts: alerts, GeneratedAt: now, Source: source})
}

func (s *server) deliverAlertHandler(w http.ResponseWriter, r *http.Request) {
	id := normalizeID(r.PathValue("id"))
	if id == "" {
		logWarn("alert delivery rejected", "code", "missing_alert_id", "path", r.URL.Path)
		writeError(w, http.StatusBadRequest, "missing_alert_id", "alert id is required")
		return
	}

	var request deliveryRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("alert delivery rejected", "alertId", id, "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.AlertID = id
	request.Channels = normalizeChannels(request.Channels)
	request.Language = normalizeLanguage(request.Language)
	logInfo(
		"alert delivery requested",
		"alertId", id,
		"channels", strings.Join(request.Channels, ","),
		"recipientRef", recipientRef(request, preferredRecipientChannel(request.Channels)),
		"dryRun", request.DryRun,
		"language", request.Language,
	)

	if code, message := validateDeliveryRequest(request); code != "" {
		logWarn("alert delivery rejected", "alertId", id, "code", code, "message", message, "channels", strings.Join(request.Channels, ","))
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now()
	alerts, _ := s.listCitizenAlerts(r.Context(), alertFeedFilters{Status: "all"}, now)
	alert, ok := findAlert(alerts, id)
	if !ok {
		logWarn("alert delivery rejected", "alertId", id, "code", "alert_not_found", "availableAlerts", len(alerts))
		writeError(w, http.StatusNotFound, "alert_not_found", "alert was not found in the citizen feed")
		return
	}

	attempts := s.store.createDeliveryAttempts(r.Context(), alert, request, s.providers, now)
	logInfo("alert delivery attempts created", "alertId", id, "attemptCount", len(attempts))
	for _, attempt := range attempts {
		if attempt.Status == "failed" || attempt.Status == "skipped" {
			logWarn(
				"alert delivery attempt did not deliver",
				"attemptId", attempt.ID,
				"alertId", attempt.AlertID,
				"channel", attempt.Channel,
				"provider", attempt.Provider,
				"status", attempt.Status,
				"reason", attempt.Reason,
			)
			continue
		}
		logInfo(
			"alert delivery attempt completed",
			"attemptId", attempt.ID,
			"alertId", attempt.AlertID,
			"channel", attempt.Channel,
			"provider", attempt.Provider,
			"status", attempt.Status,
			"messageId", attempt.MessageID,
		)
	}
	writeJSON(w, http.StatusAccepted, deliveryResponse{Attempts: attempts})
}

func (s *server) listDeliveryLogsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseLogFilters(r)
	if code != "" {
		logWarn("delivery log list rejected", "code", code, "message", message, "query", r.URL.RawQuery)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	logInfo("delivery log list requested", "alertId", filters.AlertID, "channel", filters.Channel, "status", filters.Status)
	logs := s.store.listDeliveryLogs(filters)
	logInfo("delivery log list completed", "count", len(logs))
	writeJSON(w, http.StatusOK, deliveryLogListResponse{Logs: logs})
}

func (s *server) createVoiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	var request voiceAlertRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("voice alert generation rejected", "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.AlertID = normalizeID(request.AlertID)
	request.WorkflowRequestedBy = normalizeID(request.WorkflowRequestedBy)
	languages, code, message := normalizeVoiceLanguages(request.Languages)
	if code != "" {
		logWarn("voice alert generation rejected", "alertId", request.AlertID, "code", code, "message", message)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}
	source, ok := normalizeVoiceSource(request.Source)
	if !ok {
		logWarn("voice alert generation rejected", "alertId", request.AlertID, "code", "invalid_source", "source", request.Source)
		writeError(w, http.StatusBadRequest, "invalid_source", "source must be tts_sandbox or recorded_audio")
		return
	}
	if request.AlertID == "" {
		logWarn("voice alert generation rejected", "code", "missing_alert_id")
		writeError(w, http.StatusBadRequest, "missing_alert_id", "alertId is required")
		return
	}

	now := s.now()
	logInfo(
		"voice alert generation requested",
		"alertId", request.AlertID,
		"languages", strings.Join(languages, ","),
		"source", source,
		"requestedBy", request.WorkflowRequestedBy,
	)
	alerts, _ := s.listCitizenAlerts(r.Context(), alertFeedFilters{Status: "all"}, now)
	alert, found := findAlert(alerts, request.AlertID)
	if !found {
		logWarn("voice alert generation rejected", "alertId", request.AlertID, "code", "alert_not_found", "availableAlerts", len(alerts))
		writeError(w, http.StatusNotFound, "alert_not_found", "alert was not found in the citizen feed")
		return
	}
	if alert.Status == "expired" {
		logWarn("voice alert generation rejected", "alertId", request.AlertID, "code", "alert_not_deliverable", "status", alert.Status)
		writeError(w, http.StatusConflict, "alert_not_deliverable", "voice alerts can only be generated for current or upcoming alerts")
		return
	}

	asset := s.store.createVoiceAlertAsset(alert, languages, source, request.WorkflowRequestedBy, now)
	logInfo(
		"voice alert generation completed",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"variantCount", len(asset.Variants),
		"reviewStatus", asset.ReviewStatus,
	)
	writeJSON(w, http.StatusCreated, voiceAlertResponse{Asset: asset})
}

func (s *server) listVoiceAlertsHandler(w http.ResponseWriter, _ *http.Request) {
	logInfo("voice alert list requested")
	assets := s.store.listVoiceAlertAssets()
	logInfo("voice alert list completed", "count", len(assets))
	writeJSON(w, http.StatusOK, voiceAlertListResponse{Assets: assets})
}

func (s *server) reviewVoiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	id := normalizeID(r.PathValue("id"))
	if id == "" {
		logWarn("voice alert review rejected", "code", "missing_voice_asset_id", "path", r.URL.Path)
		writeError(w, http.StatusBadRequest, "missing_voice_asset_id", "voice alert asset id is required")
		return
	}

	var request voiceReviewRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("voice alert review rejected", "voiceAssetId", id, "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	action := normalizeVoiceReviewAction(request.Action)
	if action == "" {
		logWarn("voice alert review rejected", "voiceAssetId", id, "code", "invalid_action", "action", request.Action)
		writeError(w, http.StatusBadRequest, "invalid_action", "action must be approve or reject")
		return
	}
	request.Reviewer = normalizeID(request.Reviewer)
	if request.Reviewer == "" {
		logWarn("voice alert review rejected", "voiceAssetId", id, "code", "missing_reviewer")
		writeError(w, http.StatusBadRequest, "missing_reviewer", "reviewer is required")
		return
	}

	languages, code, message := normalizeVoiceLanguages(request.Languages)
	if code != "" {
		logWarn("voice alert review rejected", "voiceAssetId", id, "code", code, "message", message)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}
	if len(request.Languages) == 0 {
		languages = nil
	}

	logInfo(
		"voice alert review requested",
		"voiceAssetId", id,
		"action", action,
		"reviewer", request.Reviewer,
		"languages", strings.Join(languages, ","),
	)
	asset, found := s.store.reviewVoiceAlertAsset(id, action, request.Reviewer, strings.TrimSpace(request.Note), languages, s.now())
	if !found {
		logWarn("voice alert review rejected", "voiceAssetId", id, "code", "voice_asset_not_found")
		writeError(w, http.StatusNotFound, "voice_asset_not_found", "voice alert asset was not found")
		return
	}

	logInfo(
		"voice alert review completed",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"status", asset.Status,
		"reviewStatus", asset.ReviewStatus,
		"reviewer", asset.Reviewer,
	)
	writeJSON(w, http.StatusOK, voiceAlertResponse{Asset: asset})
}

func (s *server) deliverVoiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	id := normalizeID(r.PathValue("id"))
	if id == "" {
		logWarn("voice alert delivery rejected", "code", "missing_voice_asset_id", "path", r.URL.Path)
		writeError(w, http.StatusBadRequest, "missing_voice_asset_id", "voice alert asset id is required")
		return
	}

	var request voiceDeliveryRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("voice alert delivery rejected", "voiceAssetId", id, "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	recipients, code, message := normalizeVoiceRecipients(request.Recipients)
	if code != "" {
		logWarn("voice alert delivery rejected", "voiceAssetId", id, "code", code, "message", message)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}
	request.Recipients = recipients

	asset, found := s.store.getVoiceAlertAsset(id)
	if !found {
		logWarn("voice alert delivery rejected", "voiceAssetId", id, "code", "voice_asset_not_found")
		writeError(w, http.StatusNotFound, "voice_asset_not_found", "voice alert asset was not found")
		return
	}
	if asset.Status != "approved" || asset.ReviewStatus != "approved" {
		logWarn(
			"voice alert delivery rejected",
			"voiceAssetId", id,
			"code", "voice_not_approved",
			"status", asset.Status,
			"reviewStatus", asset.ReviewStatus,
		)
		writeError(w, http.StatusConflict, "voice_not_approved", "voice alert asset must be approved before delivery")
		return
	}

	logInfo(
		"voice alert delivery requested",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"recipientCount", len(request.Recipients),
		"dryRun", request.DryRun,
	)
	attempts := s.store.createVoiceDeliveryAttempts(r.Context(), asset, request, s.providers, s.now())
	logInfo("voice alert delivery attempts created", "voiceAssetId", asset.ID, "alertId", asset.AlertID, "attemptCount", len(attempts))
	writeJSON(w, http.StatusAccepted, voiceDeliveryResponse{Attempts: attempts})
}

func (s *server) ussdWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var request ussdWebhookRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("ussd webhook rejected", "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.SessionID = strings.TrimSpace(request.SessionID)
	request.Phone = strings.TrimSpace(request.Phone)
	request.Language = normalizeLanguage(request.Language)
	request.Provider = providerOrDefault(request.Provider, "ussd_sandbox")
	request.ProfileID = normalizeID(request.ProfileID)
	request.ProviderError = strings.TrimSpace(request.ProviderError)

	logInfo(
		"ussd webhook received",
		"sessionId", request.SessionID,
		"provider", request.Provider,
		"phoneRef", phoneRef(request.Phone),
		"pathDepth", len(ussdTokens(request.Text)),
		"hasProviderError", request.ProviderError != "",
		"linkedProfileRequested", request.LinkProfile,
	)
	if request.SessionID == "" {
		logWarn("ussd webhook rejected", "code", "missing_session", "provider", request.Provider, "phoneRef", phoneRef(request.Phone))
		writeError(w, http.StatusBadRequest, "missing_session", "sessionId is required")
		return
	}
	if request.Phone == "" {
		logWarn("ussd webhook rejected", "code", "missing_phone", "sessionId", request.SessionID, "provider", request.Provider)
		writeError(w, http.StatusBadRequest, "missing_phone", "phone is required")
		return
	}

	response := s.handleUSSDRequest(r.Context(), request)
	writeJSON(w, http.StatusOK, response)
}

func (s *server) smsInboundHandler(w http.ResponseWriter, r *http.Request) {
	var request smsInboundRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("sms inbound rejected", "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.From = strings.TrimSpace(request.From)
	request.Body = strings.TrimSpace(request.Body)
	request.Language = normalizeLanguage(request.Language)
	request.Provider = providerOrDefault(request.Provider, "sms_sandbox")
	request.ProfileID = normalizeID(request.ProfileID)
	request.ProviderError = strings.TrimSpace(request.ProviderError)

	logInfo(
		"sms inbound received",
		"provider", request.Provider,
		"phoneRef", phoneRef(request.From),
		"command", smsCommandName(request.Body),
		"hasProviderError", request.ProviderError != "",
		"linkedProfileRequested", request.LinkProfile,
	)
	if request.From == "" {
		logWarn("sms inbound rejected", "code", "missing_from", "provider", request.Provider)
		writeError(w, http.StatusBadRequest, "missing_from", "from is required")
		return
	}
	if request.Body == "" && request.ProviderError == "" {
		logWarn("sms inbound rejected", "code", "missing_body", "provider", request.Provider, "phoneRef", phoneRef(request.From))
		writeError(w, http.StatusBadRequest, "missing_body", "body is required")
		return
	}

	response := s.handleSMSInbound(r.Context(), request)
	writeJSON(w, http.StatusAccepted, response)
}

func (s *server) whatsappWebhookHandler(w http.ResponseWriter, r *http.Request) {
	var request whatsappInboundRequest
	if err := decodeJSON(r, &request); err != nil {
		logWarn("whatsapp webhook rejected", "code", "invalid_json", "error", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.From = strings.TrimSpace(request.From)
	request.Body = strings.TrimSpace(request.Body)
	request.Language = normalizeLanguage(request.Language)
	request.Provider = providerOrDefault(request.Provider, "whatsapp_sandbox")
	request.ProfileID = normalizeID(request.ProfileID)
	request.ProviderError = strings.TrimSpace(request.ProviderError)
	request.Media = normalizeWhatsAppMedia(request.Media)

	logInfo(
		"whatsapp webhook received",
		"provider", request.Provider,
		"phoneRef", phoneRef(request.From),
		"command", smsCommandName(request.Body),
		"hasProviderError", request.ProviderError != "",
		"hasLocation", request.Location != nil,
		"mediaCount", len(request.Media),
		"linkedProfileRequested", request.LinkProfile,
	)
	if request.From == "" {
		logWarn("whatsapp webhook rejected", "code", "missing_from", "provider", request.Provider)
		writeError(w, http.StatusBadRequest, "missing_from", "from is required")
		return
	}
	if request.Body == "" && request.ProviderError == "" && request.Location == nil && len(request.Media) == 0 {
		logWarn("whatsapp webhook rejected", "code", "missing_body", "provider", request.Provider, "phoneRef", phoneRef(request.From))
		writeError(w, http.StatusBadRequest, "missing_body", "body, location, media, or providerError is required")
		return
	}

	response := s.handleWhatsAppInbound(r.Context(), request)
	writeJSON(w, http.StatusAccepted, response)
}

func (s *server) listAccessLogsHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseAccessLogFilters(r)
	if code != "" {
		logWarn("inclusive access log list rejected", "code", code, "message", message, "query", r.URL.RawQuery)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	logInfo("inclusive access log list requested", "channel", filters.Channel, "intent", filters.Intent, "status", filters.Status)
	logs := s.store.listAccessLogs(filters)
	logInfo("inclusive access log list completed", "count", len(logs))
	writeJSON(w, http.StatusOK, accessLogListResponse{Logs: logs})
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
		logInfo("alert-service fetch starting", "baseURL", s.alertClient.baseURL, "fixtureCount", len(fixtureAlerts))
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
			logInfo(
				"alert-service fetch completed",
				"upstreamCount", len(upstreamAlerts),
				"combinedCount", len(combined),
				"source", source,
			)
		} else {
			logWarn("alert-service fetch failed using fixture fallback", "error", err, "fixtureCount", len(fixtureAlerts))
		}
	}

	sortCitizenAlerts(combined)
	return combined, source
}

func (s *server) handleUSSDRequest(ctx context.Context, request ussdWebhookRequest) ussdWebhookResponse {
	now := s.now()
	phoneRef := phoneRef(request.Phone)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := normalizeUSSDLanguage(request.Language, request.Text)
	logInfo(
		"ussd session handling started",
		"sessionId", request.SessionID,
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"language", language,
		"pathDepth", len(ussdTokens(request.Text)),
		"linkedProfile", linkedProfile,
	)

	if request.ProviderError != "" {
		logWarn(
			"ussd provider error received",
			"sessionId", request.SessionID,
			"provider", request.Provider,
			"providerMessageId", request.ProviderMessageID,
			"phoneRef", phoneRef,
			"errorLength", len(request.ProviderError),
		)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "ussd",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         request.SessionID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "provider_error",
			Status:            "failed",
			ProviderError:     request.ProviderError,
			CreatedAt:         now,
		})
		return ussdWebhookResponse{
			SessionID: request.SessionID,
			Action:    "end",
			Message:   localizedMessage(language, "provider_error"),
			Language:  language,
			Log:       log,
		}
	}

	tokens := ussdTokens(request.Text)
	if len(tokens) == 0 {
		logInfo("ussd language menu returned", "sessionId", request.SessionID, "provider", request.Provider, "phoneRef", phoneRef)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "language_menu",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: languageMenu(), Language: language, Log: log}
	}

	if _, ok := ussdLanguageFromToken(tokens[0]); !ok {
		logWarn(
			"ussd invalid language selection",
			"sessionId", request.SessionID,
			"provider", request.Provider,
			"phoneRef", phoneRef,
			"pathDepth", len(tokens),
		)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: languageMenu(), Language: language, Log: log}
	}

	if len(tokens) == 1 {
		logInfo("ussd main menu returned", "sessionId", request.SessionID, "provider", request.Provider, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "main_menu",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: mainMenu(language), Language: language, Log: log}
	}

	switch tokens[1] {
	case "1":
		alerts, _ := s.listCitizenAlerts(ctx, alertFeedFilters{}, now)
		logInfo("ussd current alerts summary returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "alertCount", len(alerts))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "current_alerts",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "end", Message: alertSummaryMessage(language, alerts), Language: language, Log: log}
	case "2":
		logInfo("ussd report flow selected", "sessionId", request.SessionID, "phoneRef", phoneRef, "pathDepth", len(tokens))
		return s.handleUSSDReport(ctx, request, tokens, language, linkedProfile, now)
	case "3":
		logInfo("ussd shelter guidance returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "shelter_lookup",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "end", Message: shelterMessage(language), Language: language, Log: log}
	case "4":
		logInfo("ussd 112 guidance returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "guidance_112",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "end", Message: guidance112Message(language), Language: language, Log: log}
	default:
		logWarn("ussd invalid main-menu selection", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language, "pathDepth", len(tokens))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: mainMenu(language), Language: language, Log: log}
	}
}

func (s *server) handleUSSDReport(ctx context.Context, request ussdWebhookRequest, tokens []string, language string, linkedProfile bool, now time.Time) ussdWebhookResponse {
	phoneRef := phoneRef(request.Phone)
	if len(tokens) == 2 {
		logInfo("ussd report hazard menu returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "report_emergency",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: hazardMenu(language), Language: language, Log: log}
	}

	hazard, ok := ussdHazardFromToken(tokens[2])
	if !ok {
		logWarn("ussd report rejected invalid hazard", "sessionId", request.SessionID, "phoneRef", phoneRef, "pathDepth", len(tokens))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: hazardMenu(language), Language: language, Log: log}
	}

	if len(tokens) == 3 {
		logInfo("ussd report urgency menu returned", "sessionId", request.SessionID, "phoneRef", phoneRef, "hazard", hazard, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "report_emergency",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: urgencyMenu(language), Language: language, Log: log}
	}

	urgency, ok := ussdUrgencyFromToken(tokens[3])
	if !ok {
		logWarn("ussd report rejected invalid urgency", "sessionId", request.SessionID, "phoneRef", phoneRef, "hazard", hazard, "pathDepth", len(tokens))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:       "ussd",
			Provider:      request.Provider,
			SessionID:     request.SessionID,
			PhoneRef:      phoneRef,
			ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile: linkedProfile,
			Language:      language,
			Intent:        "invalid_selection",
			Status:        "handled",
			CreatedAt:     now,
		})
		return ussdWebhookResponse{SessionID: request.SessionID, Action: "continue", Message: urgencyMenu(language), Language: language, Log: log}
	}

	location, locationLabel := inclusiveLocation(request.Location, tokens[4:])
	logInfo(
		"ussd report creating access report",
		"sessionId", request.SessionID,
		"phoneRef", phoneRef,
		"hazard", hazard,
		"urgency", urgency,
		"hasCoordinates", request.Location != nil,
		"locationLabel", logTextSummary(locationLabel),
		"linkedProfile", linkedProfile,
	)
	report := s.store.createAccessReport(inclusiveAccessReport{
		Channel:       "ussd",
		Type:          hazard,
		Urgency:       urgency,
		Description:   fmt.Sprintf("USSD emergency report: %s with %s urgency. Location note: %s.", hazard, urgency, locationLabel),
		Location:      location,
		LocationLabel: locationLabel,
		PhoneRef:      phoneRef,
		ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile: linkedProfile,
		Status:        "queued",
		CreatedAt:     now,
	})
	report = s.submitInclusiveReport(ctx, report, request.Phone, request.ProfileID, linkedProfile)
	logInfo(
		"ussd report flow completed",
		"sessionId", request.SessionID,
		"phoneRef", phoneRef,
		"reportId", report.ID,
		"status", report.Status,
		"incidentId", report.IncidentID,
		"incidentReference", report.IncidentReference,
	)

	log := s.store.createAccessLog(inclusiveAccessLog{
		Channel:           "ussd",
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		SessionID:         request.SessionID,
		PhoneRef:          phoneRef,
		ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Language:          language,
		Intent:            "report_emergency",
		Status:            report.Status,
		IncidentID:        report.IncidentID,
		IncidentReference: report.IncidentReference,
		CreatedAt:         now,
	})

	message := reportConfirmationMessage(language, report)
	return ussdWebhookResponse{SessionID: request.SessionID, Action: "end", Message: message, Language: language, Log: log, Report: &report}
}

func (s *server) handleSMSInbound(ctx context.Context, request smsInboundRequest) smsInboundResponse {
	now := s.now()
	phoneRef := phoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language
	logInfo(
		"sms inbound handling started",
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"command", smsCommandName(request.Body),
		"language", language,
		"linkedProfile", linkedProfile,
	)

	if request.ProviderError != "" {
		logWarn(
			"sms provider error received",
			"provider", request.Provider,
			"providerMessageId", request.ProviderMessageID,
			"phoneRef", phoneRef,
			"errorLength", len(request.ProviderError),
		)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "provider_error",
			Status:            "failed",
			ProviderError:     request.ProviderError,
			CreatedAt:         now,
		})
		return smsInboundResponse{Message: localizedMessage(language, "provider_error"), Log: log}
	}

	command := strings.TrimSpace(request.Body)
	upperCommand := strings.ToUpper(command)
	switch {
	case upperCommand == "ALERT" || upperCommand == "ALERTS":
		alerts, _ := s.listCitizenAlerts(ctx, alertFeedFilters{}, now)
		logInfo("sms alert summary returned", "provider", request.Provider, "phoneRef", phoneRef, "alertCount", len(alerts))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "current_alerts",
			Status:            "handled",
			CreatedAt:         now,
		})
		return smsInboundResponse{Message: smsAlertMessage(alerts), Log: log}
	case upperCommand == "SHELTER" || upperCommand == "SHELTERS":
		logInfo("sms shelter guidance returned", "provider", request.Provider, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "shelter_lookup",
			Status:            "handled",
			CreatedAt:         now,
		})
		return smsInboundResponse{Message: shelterMessage(language), Log: log}
	case upperCommand == "HELP" || upperCommand == "112":
		logInfo("sms 112 guidance returned", "provider", request.Provider, "phoneRef", phoneRef, "language", language, "command", smsCommandName(request.Body))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "guidance_112",
			Status:            "handled",
			CreatedAt:         now,
		})
		return smsInboundResponse{Message: guidance112Message(language), Log: log}
	case strings.HasPrefix(upperCommand, "REPORT "):
		report, ok, usage := parseSMSReport(request, phoneRef, linkedProfile, now)
		if !ok {
			logWarn("sms report rejected invalid usage", "provider", request.Provider, "phoneRef", phoneRef, "command", smsCommandName(request.Body))
			log := s.store.createAccessLog(inclusiveAccessLog{
				Channel:           "sms",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				PhoneRef:          phoneRef,
				ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
				LinkedProfile:     linkedProfile,
				Language:          language,
				Intent:            "invalid_selection",
				Status:            "handled",
				CreatedAt:         now,
			})
			return smsInboundResponse{Message: usage, Log: log}
		}
		logInfo(
			"sms report creating access report",
			"provider", request.Provider,
			"phoneRef", phoneRef,
			"hazard", report.Type,
			"urgency", report.Urgency,
			"hasCoordinates", request.Location != nil,
			"locationLabel", logTextSummary(report.LocationLabel),
			"linkedProfile", linkedProfile,
		)
		report = s.store.createAccessReport(report)
		report = s.submitInclusiveReport(ctx, report, request.From, request.ProfileID, linkedProfile)
		logInfo(
			"sms report flow completed",
			"provider", request.Provider,
			"phoneRef", phoneRef,
			"reportId", report.ID,
			"status", report.Status,
			"incidentId", report.IncidentID,
			"incidentReference", report.IncidentReference,
		)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            report.Status,
			IncidentID:        report.IncidentID,
			IncidentReference: report.IncidentReference,
			CreatedAt:         now,
		})
		return smsInboundResponse{Message: reportConfirmationMessage(language, report), Log: log, Report: &report}
	default:
		logWarn("sms inbound unknown command", "provider", request.Provider, "phoneRef", phoneRef, "command", smsCommandName(request.Body))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "sms",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "invalid_selection",
			Status:            "handled",
			CreatedAt:         now,
		})
		return smsInboundResponse{Message: smsHelpMessage(), Log: log}
	}
}

func (s *server) handleWhatsAppInbound(ctx context.Context, request whatsappInboundRequest) whatsappInboundResponse {
	now := s.now()
	phoneRef := phoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language
	conversationKey := whatsappConversationKey(phoneRef, request.ProfileID, linkedProfile)
	conversation := s.store.getOrCreateWhatsAppConversation(
		conversationKey,
		phoneRef,
		profileIDForLog(request.ProfileID, linkedProfile),
		linkedProfile,
		language,
		now,
	)
	inboundTranscript := s.store.createWhatsAppTranscript(whatsappTranscript{
		ConversationID:    conversation.ID,
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		PhoneRef:          phoneRef,
		ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile:     linkedProfile,
		Direction:         "inbound",
		Intent:            "incoming",
		State:             conversation.State,
		MessageSummary:    whatsappMessageSummary(request.Body),
		MediaSummary:      whatsappMediaSummary(request.Media),
		CreatedAt:         now,
		RetentionUntil:    whatsappRetentionUntil(now),
	})
	logInfo(
		"whatsapp inbound handling started",
		"conversationId", conversation.ID,
		"provider", request.Provider,
		"phoneRef", phoneRef,
		"command", smsCommandName(request.Body),
		"state", conversation.State,
		"language", language,
		"linkedProfile", linkedProfile,
	)

	if request.ProviderError != "" {
		logWarn(
			"whatsapp provider error received",
			"conversationId", conversation.ID,
			"provider", request.Provider,
			"providerMessageId", request.ProviderMessageID,
			"phoneRef", phoneRef,
			"errorLength", len(request.ProviderError),
		)
		conversation.Intent = "provider_error"
		conversation.State = "idle"
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "provider_error",
			Status:            "failed",
			ProviderError:     request.ProviderError,
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, localizedMessage(language, "provider_error"), log, nil, inboundTranscript.ID, now)
	}

	command := smsCommandName(request.Body)
	if command == "CANCEL" || command == "MENU" || command == "START" || command == "HI" || command == "HELLO" {
		conversation.Intent = "main_menu"
		conversation.State = "idle"
		conversation.Hazard = ""
		conversation.Urgency = ""
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp main menu returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "command", command)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "main_menu",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHelpMessage(), log, nil, inboundTranscript.ID, now)
	}

	if conversation.State != "" && conversation.State != "idle" && !isWhatsAppTopLevelCommand(command) {
		return s.handleWhatsAppReportConversation(ctx, request, conversation, inboundTranscript.ID, now)
	}

	switch command {
	case "ALERT", "ALERTS":
		alerts, _ := s.listCitizenAlerts(ctx, alertFeedFilters{}, now)
		conversation.Intent = "current_alerts"
		conversation.State = "idle"
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp alert summary returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "alertCount", len(alerts))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "current_alerts",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, alertSummaryMessage(language, alerts), log, nil, inboundTranscript.ID, now)
	case "RISK":
		alerts, _ := s.listCitizenAlerts(ctx, alertFeedFilters{}, now)
		conversation.Intent = "risk_check"
		conversation.State = "idle"
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp risk guidance returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "hasLocation", request.Location != nil, "alertCount", len(alerts))
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "risk_check",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, riskCheckMessage(language, request.Location != nil, alerts), log, nil, inboundTranscript.ID, now)
	case "GUIDE", "GUIDES":
		hazard := whatsappCommandArg(request.Body, 1)
		if hazard == "" {
			hazard = "flood"
		}
		hazard = normalizeSMSHazard(hazard)
		if !allowedHazards[hazard] {
			hazard = "flood"
		}
		conversation.Intent = "emergency_guides"
		conversation.State = "idle"
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp emergency guide returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", hazard)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp shelter guidance returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp 112 guidance returned", "conversationId", conversation.ID, "phoneRef", phoneRef, "language", language)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
			logWarn("whatsapp report rejected invalid direct usage", "conversationId", conversation.ID, "phoneRef", phoneRef, "command", command)
			conversation.Intent = "invalid_selection"
			conversation.State = "idle"
			conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
			conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
			conversation.UpdatedAt = now
			conversation = s.store.updateWhatsAppConversation(conversation)
			log := s.store.createAccessLog(inclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp report flow started", "conversationId", conversation.ID, "phoneRef", phoneRef)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHazardPrompt(), log, nil, inboundTranscript.ID, now)
	default:
		if request.Location != nil || len(request.Media) > 0 {
			logWarn("whatsapp location or media received without active report", "conversationId", conversation.ID, "phoneRef", phoneRef, "mediaCount", len(request.Media), "hasLocation", request.Location != nil)
		} else {
			logWarn("whatsapp inbound unknown command", "conversationId", conversation.ID, "phoneRef", phoneRef, "command", command)
		}
		conversation.Intent = "invalid_selection"
		conversation.State = "idle"
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "invalid_selection",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHelpMessage(), log, nil, inboundTranscript.ID, now)
	}
}

func (s *server) handleWhatsAppReportConversation(ctx context.Context, request whatsappInboundRequest, conversation whatsappConversation, inboundTranscriptID string, now time.Time) whatsappInboundResponse {
	phoneRef := phoneRef(request.From)
	linkedProfile := request.LinkProfile && request.ProfileID != ""
	language := request.Language

	switch conversation.State {
	case "awaiting_report_hazard":
		hazard := normalizeSMSHazard(firstToken(request.Body))
		if !allowedHazards[hazard] {
			logWarn("whatsapp report rejected invalid hazard", "conversationId", conversation.ID, "phoneRef", phoneRef, "state", conversation.State)
			log := s.store.createAccessLog(inclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp report hazard captured", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", hazard)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappUrgencyPrompt(), log, nil, inboundTranscriptID, now)
	case "awaiting_report_urgency":
		fields := strings.Fields(request.Body)
		urgency := normalizeSMSUrgency(firstToken(request.Body))
		if urgency == "" {
			logWarn("whatsapp report rejected invalid urgency", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", conversation.Hazard)
			log := s.store.createAccessLog(inclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
		conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
		conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		logInfo("whatsapp report urgency captured", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", conversation.Hazard, "urgency", urgency)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "report_emergency",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappLocationPrompt(), log, nil, inboundTranscriptID, now)
	case "awaiting_report_location":
		if strings.TrimSpace(request.Body) == "" && request.Location == nil && len(request.Media) == 0 {
			logWarn("whatsapp report still missing location details", "conversationId", conversation.ID, "phoneRef", phoneRef, "hazard", conversation.Hazard, "urgency", conversation.Urgency)
			log := s.store.createAccessLog(inclusiveAccessLog{
				Channel:           "whatsapp",
				Provider:          request.Provider,
				ProviderMessageID: request.ProviderMessageID,
				SessionID:         conversation.ID,
				PhoneRef:          phoneRef,
				ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
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
		logWarn("whatsapp conversation state unknown", "conversationId", conversation.ID, "phoneRef", phoneRef, "state", conversation.State)
		conversation.Intent = "invalid_selection"
		conversation.State = "idle"
		conversation.UpdatedAt = now
		conversation = s.store.updateWhatsAppConversation(conversation)
		log := s.store.createAccessLog(inclusiveAccessLog{
			Channel:           "whatsapp",
			Provider:          request.Provider,
			ProviderMessageID: request.ProviderMessageID,
			SessionID:         conversation.ID,
			PhoneRef:          phoneRef,
			ProfileID:         profileIDForLog(request.ProfileID, linkedProfile),
			LinkedProfile:     linkedProfile,
			Language:          language,
			Intent:            "invalid_selection",
			Status:            "handled",
			CreatedAt:         now,
		})
		return s.whatsappResponse(request, conversation, whatsappHelpMessage(), log, nil, inboundTranscriptID, now)
	}
}

func (s *server) completeWhatsAppReport(ctx context.Context, request whatsappInboundRequest, conversation whatsappConversation, report inclusiveAccessReport, inboundTranscriptID string, now time.Time) whatsappInboundResponse {
	logInfo(
		"whatsapp report creating access report",
		"conversationId", conversation.ID,
		"phoneRef", report.PhoneRef,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"hasCoordinates", request.Location != nil,
		"mediaCount", len(report.Media),
		"locationLabel", logTextSummary(report.LocationLabel),
		"linkedProfile", report.LinkedProfile,
	)
	report = s.store.createAccessReport(report)
	report = s.submitInclusiveReport(ctx, report, request.From, request.ProfileID, report.LinkedProfile)
	conversation.Intent = "report_emergency"
	conversation.State = "idle"
	conversation.Hazard = ""
	conversation.Urgency = ""
	conversation.LastMessageSummary = whatsappMessageSummary(request.Body)
	conversation.LastMediaSummary = whatsappMediaSummary(request.Media)
	conversation.UpdatedAt = now
	conversation = s.store.updateWhatsAppConversation(conversation)
	logInfo(
		"whatsapp report flow completed",
		"conversationId", conversation.ID,
		"phoneRef", report.PhoneRef,
		"reportId", report.ID,
		"status", report.Status,
		"incidentId", report.IncidentID,
		"incidentReference", report.IncidentReference,
	)

	log := s.store.createAccessLog(inclusiveAccessLog{
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

func (s *server) whatsappResponse(request whatsappInboundRequest, conversation whatsappConversation, message string, accessLog inclusiveAccessLog, report *inclusiveAccessReport, inboundTranscriptID string, now time.Time) whatsappInboundResponse {
	outboundTranscript := s.store.createWhatsAppTranscript(whatsappTranscript{
		ConversationID:    conversation.ID,
		Provider:          request.Provider,
		ProviderMessageID: request.ProviderMessageID,
		PhoneRef:          conversation.PhoneRef,
		ProfileID:         conversation.ProfileID,
		LinkedProfile:     conversation.LinkedProfile,
		Direction:         "outbound",
		Intent:            accessLog.Intent,
		State:             conversation.State,
		MessageSummary:    whatsappMessageSummary(message),
		MediaSummary:      "",
		CreatedAt:         now,
		RetentionUntil:    conversation.RetentionUntil,
	})
	transcriptIDs := []string{outboundTranscript.ID}
	if inboundTranscriptID != "" {
		transcriptIDs = append([]string{inboundTranscriptID}, transcriptIDs...)
	}
	return whatsappInboundResponse{
		Message:       message,
		Conversation:  conversation,
		Log:           accessLog,
		Report:        report,
		TranscriptIDs: transcriptIDs,
	}
}

func (s *server) submitInclusiveReport(ctx context.Context, report inclusiveAccessReport, rawPhone string, profileID string, linkedProfile bool) inclusiveAccessReport {
	if s.incidentClient == nil {
		report.Status = "queued"
		report.FailureReason = "incident-service handoff is not configured"
		logWarn(
			"inclusive report queued without incident-service",
			"reportId", report.ID,
			"channel", report.Channel,
			"hazard", report.Type,
			"urgency", report.Urgency,
			"phoneRef", report.PhoneRef,
		)
		s.store.updateAccessReport(report)
		return report
	}

	logInfo(
		"inclusive report incident handoff starting",
		"reportId", report.ID,
		"channel", report.Channel,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"phoneRef", report.PhoneRef,
		"linkedProfile", linkedProfile,
	)
	response, err := s.incidentClient.createIncident(ctx, report, rawPhone, profileID, linkedProfile)
	if err != nil {
		report.Status = "queued"
		report.FailureReason = err.Error()
		logWarn(
			"inclusive report incident handoff failed",
			"reportId", report.ID,
			"channel", report.Channel,
			"phoneRef", report.PhoneRef,
			"error", err,
		)
		s.store.updateAccessReport(report)
		return report
	}

	report.Status = "submitted"
	report.IncidentID = response.ID
	report.IncidentReference = response.Reference
	report.FailureReason = ""
	logInfo(
		"inclusive report incident handoff completed",
		"reportId", report.ID,
		"channel", report.Channel,
		"phoneRef", report.PhoneRef,
		"incidentId", report.IncidentID,
		"incidentReference", report.IncidentReference,
	)
	s.store.updateAccessReport(report)
	return report
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

func (c *incidentServiceClient) createIncident(ctx context.Context, report inclusiveAccessReport, rawPhone string, profileID string, linkedProfile bool) (incidentIntakeResponse, error) {
	parsed, err := url.Parse(c.baseURL + "/incidents")
	if err != nil {
		logError("incident-service handoff url invalid", "baseURL", c.baseURL, "reportId", report.ID, "error", err)
		return incidentIntakeResponse{}, err
	}
	logInfo(
		"incident-service handoff request prepared",
		"reportId", report.ID,
		"channel", report.Channel,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"endpoint", parsed.String(),
		"linkedProfile", linkedProfile,
	)

	payload := incidentIntakeRequest{
		Type:               report.Type,
		Description:        report.Description,
		Location:           report.Location,
		PeopleAffected:     0,
		InjuriesReported:   report.Urgency == "life_threatening",
		Urgency:            report.Urgency,
		Anonymous:          !linkedProfile,
		ContactPermission:  linkedProfile,
		AccessibilityNeeds: "Inclusive access channel report",
		Media:              report.Media,
	}
	if linkedProfile {
		payload.Reporter = &reporterRef{UserID: profileID, Phone: rawPhone}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logError("incident-service handoff payload marshal failed", "reportId", report.ID, "error", err)
		return incidentIntakeResponse{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), strings.NewReader(string(body)))
	if err != nil {
		logError("incident-service handoff request creation failed", "reportId", report.ID, "error", err)
		return incidentIntakeResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		logWarn("incident-service handoff request failed", "reportId", report.ID, "endpoint", parsed.String(), "error", err)
		return incidentIntakeResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		logWarn("incident-service handoff returned non-success", "reportId", report.ID, "endpoint", parsed.String(), "statusCode", response.StatusCode)
		return incidentIntakeResponse{}, fmt.Errorf("incident-service returned %d", response.StatusCode)
	}

	var result incidentIntakeResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		logError("incident-service handoff response decode failed", "reportId", report.ID, "statusCode", response.StatusCode, "error", err)
		return incidentIntakeResponse{}, err
	}
	logInfo("incident-service handoff response decoded", "reportId", report.ID, "incidentId", result.ID, "incidentReference", result.Reference)
	return result, nil
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
		return logFilters{}, "invalid_channel", "channel must be push, sms, or voice"
	}
	if filters.Status != "" && !allowedDeliveryStatuses[filters.Status] {
		return logFilters{}, "invalid_status", "status must be queued, delivered, failed, or skipped"
	}
	return filters, "", ""
}

func parseAccessLogFilters(r *http.Request) (accessLogFilters, string, string) {
	query := r.URL.Query()
	filters := accessLogFilters{
		Channel: normalizeQueryValue(query.Get("channel")),
		Intent:  normalizeQueryValue(query.Get("intent")),
		Status:  normalizeQueryValue(query.Get("status")),
	}
	if filters.Channel != "" && !allowedAccessChannels[filters.Channel] {
		return accessLogFilters{}, "invalid_channel", "channel must be sms, ussd, or whatsapp"
	}
	if filters.Intent != "" && !allowedAccessIntents[filters.Intent] {
		return accessLogFilters{}, "invalid_intent", "intent must be a supported inclusive access intent"
	}
	if filters.Status != "" && !allowedAccessStatuses[filters.Status] {
		return accessLogFilters{}, "invalid_status", "status must be handled, failed, queued, or submitted"
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
			return "invalid_channel", "channels must include only push, sms, or voice"
		}
		if channel == "voice" {
			return "voice_requires_asset", "voice delivery requires an approved voice alert asset"
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
		if provider == nil {
			logError("notification provider missing", "alertId", alert.ID, "channel", channel)
			provider = disabledProvider{channel: channel, reason: "provider missing"}
		}
		logInfo(
			"notification provider send starting",
			"alertId", alert.ID,
			"channel", channel,
			"provider", providerName(provider),
			"recipientRef", recipientRef(request, channel),
			"dryRun", request.DryRun,
		)
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
		logInfo(
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

func (m *memoryStore) createVoiceAlertAsset(alert citizenAlert, languages []string, source string, requestedBy string, now time.Time) voiceAlertAsset {
	m.mu.Lock()
	defer m.mu.Unlock()

	asset := voiceAlertAsset{
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
		Variants:            make([]voiceVariant, 0, len(languages)),
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if asset.TargetLabel == "" {
		asset.TargetLabel = alert.Target.Label
	}
	m.nextVoiceAlertID++

	for _, language := range languages {
		messageText := voiceMessageForAlert(language, alert)
		variant := voiceVariant{
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
	logInfo(
		"voice alert asset stored",
		"voiceAssetId", asset.ID,
		"alertId", asset.AlertID,
		"variantCount", len(asset.Variants),
		"source", asset.Source,
		"requestedBy", asset.WorkflowRequestedBy,
	)
	return copyVoiceAlertAsset(asset)
}

func (m *memoryStore) listVoiceAlertAssets() []voiceAlertAsset {
	m.mu.RLock()
	defer m.mu.RUnlock()

	assets := make([]voiceAlertAsset, 0, len(m.voiceAlerts))
	for _, asset := range m.voiceAlerts {
		assets = append(assets, copyVoiceAlertAsset(asset))
	}
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].CreatedAt.After(assets[j].CreatedAt)
	})
	return assets
}

func (m *memoryStore) getVoiceAlertAsset(id string) (voiceAlertAsset, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, asset := range m.voiceAlerts {
		if asset.ID == id {
			return copyVoiceAlertAsset(asset), true
		}
	}
	return voiceAlertAsset{}, false
}

func (m *memoryStore) reviewVoiceAlertAsset(id string, action string, reviewer string, note string, languages []string, now time.Time) (voiceAlertAsset, bool) {
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
		reviewStatus := "approved"
		var reviewedCount int
		for variantIndex, variant := range asset.Variants {
			if !reviewAll && !selectedLanguages[variant.Language] {
				if variant.ReviewStatus != "approved" {
					reviewStatus = "partial_review"
				}
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
		logInfo(
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

	return voiceAlertAsset{}, false
}

func (m *memoryStore) createVoiceDeliveryAttempts(ctx context.Context, asset voiceAlertAsset, request voiceDeliveryRequest, providers map[string]notificationProvider, now time.Time) []deliveryAttempt {
	m.mu.Lock()
	defer m.mu.Unlock()

	attempts := make([]deliveryAttempt, 0, len(request.Recipients))
	provider := providers["voice"]
	if provider == nil {
		logError("voice notification provider missing", "voiceAssetId", asset.ID, "alertId", asset.AlertID)
		provider = disabledProvider{channel: "voice", reason: "voice provider missing"}
	}

	for _, recipient := range request.Recipients {
		variant, variantFound := voiceVariantForLanguage(asset, recipient.Language)
		result := providerResult{}
		if !variantFound {
			result = providerResult{Provider: "voice_asset", Status: "skipped", Reason: "approved language variant is missing"}
			logWarn("voice delivery skipped missing variant", "voiceAssetId", asset.ID, "alertId", asset.AlertID, "language", recipient.Language, "recipientRef", voiceRecipientRef(recipient))
		} else if variant.ReviewStatus != "approved" {
			result = providerResult{Provider: "voice_asset", Status: "skipped", Reason: "language variant is not approved"}
			logWarn("voice delivery skipped unapproved variant", "voiceAssetId", asset.ID, "alertId", asset.AlertID, "language", recipient.Language, "variantStatus", variant.ReviewStatus, "recipientRef", voiceRecipientRef(recipient))
		} else {
			deliveryReq := deliveryRequest{
				AlertID:     asset.AlertID,
				RecipientID: recipient.RecipientID,
				Phone:       recipient.Phone,
				Language:    recipient.Language,
				Channels:    []string{"voice"},
				DryRun:      request.DryRun,
			}
			logInfo(
				"voice provider send starting",
				"voiceAssetId", asset.ID,
				"alertId", asset.AlertID,
				"language", recipient.Language,
				"provider", providerName(provider),
				"recipientRef", voiceRecipientRef(recipient),
				"dryRun", request.DryRun,
			)
			result = provider.Send(ctx, providerMessage{
				Alert: citizenAlert{
					ID:          asset.AlertID,
					Title:       asset.AlertTitle,
					HazardType:  asset.HazardType,
					Severity:    asset.Severity,
					TargetLabel: asset.TargetLabel,
				},
				Request:     deliveryReq,
				Channel:     "voice",
				Recipient:   voiceRecipientRef(recipient),
				AttemptedAt: now,
			})
		}

		attempt := deliveryAttempt{
			ID:           fmt.Sprintf("delivery_%06d", m.nextLogID),
			AlertID:      asset.AlertID,
			AlertTitle:   asset.AlertTitle,
			Channel:      "voice",
			Provider:     result.Provider,
			RecipientRef: voiceRecipientRef(recipient),
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
		logInfo(
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

func (m *memoryStore) createAccessLog(log inclusiveAccessLog) inclusiveAccessLog {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.ID = fmt.Sprintf("access_%06d", m.nextAccessLogID)
	m.nextAccessLogID++
	m.accessLogs = append(m.accessLogs, log)
	logInfo(
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

func (m *memoryStore) listAccessLogs(filters accessLogFilters) []inclusiveAccessLog {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := make([]inclusiveAccessLog, 0, len(m.accessLogs))
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

func (m *memoryStore) createAccessReport(report inclusiveAccessReport) inclusiveAccessReport {
	m.mu.Lock()
	defer m.mu.Unlock()

	report.ID = fmt.Sprintf("access_report_%06d", m.nextAccessReportID)
	m.nextAccessReportID++
	m.accessReports = append(m.accessReports, report)
	logInfo(
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

func (m *memoryStore) updateAccessReport(report inclusiveAccessReport) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index, existing := range m.accessReports {
		if existing.ID == report.ID {
			m.accessReports[index] = report
			logInfo(
				"inclusive access report updated",
				"reportId", report.ID,
				"channel", report.Channel,
				"status", report.Status,
				"incidentId", report.IncidentID,
				"incidentReference", report.IncidentReference,
				"failureReason", logTextSummary(report.FailureReason),
			)
			return
		}
	}
	m.accessReports = append(m.accessReports, report)
	logWarn(
		"inclusive access report update appended missing report",
		"reportId", report.ID,
		"channel", report.Channel,
		"status", report.Status,
	)
}

func (m *memoryStore) getOrCreateWhatsAppConversation(key string, phoneRef string, profileID string, linkedProfile bool, language string, now time.Time) whatsappConversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.whatsappConversations == nil {
		m.whatsappConversations = map[string]whatsappConversation{}
	}
	if conversation, ok := m.whatsappConversations[key]; ok {
		conversation.ProfileID = profileID
		conversation.LinkedProfile = linkedProfile
		conversation.Language = language
		conversation.UpdatedAt = now
		conversation.ExpiresAt = now.Add(24 * time.Hour)
		conversation.RetentionUntil = whatsappRetentionUntil(now)
		m.whatsappConversations[key] = conversation
		logInfo(
			"whatsapp conversation resumed",
			"conversationId", conversation.ID,
			"phoneRef", conversation.PhoneRef,
			"state", conversation.State,
			"intent", conversation.Intent,
		)
		return conversation
	}

	conversation := whatsappConversation{
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
		RetentionUntil: whatsappRetentionUntil(now),
	}
	m.nextWhatsAppConversationID++
	m.whatsappConversations[key] = conversation
	logInfo(
		"whatsapp conversation created",
		"conversationId", conversation.ID,
		"phoneRef", conversation.PhoneRef,
		"linkedProfile", conversation.LinkedProfile,
		"retentionUntil", conversation.RetentionUntil,
	)
	return conversation
}

func (m *memoryStore) updateWhatsAppConversation(conversation whatsappConversation) whatsappConversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.whatsappConversations == nil {
		m.whatsappConversations = map[string]whatsappConversation{}
	}
	conversation.UpdatedAt = conversation.UpdatedAt.UTC()
	m.whatsappConversations[conversation.Key] = conversation
	logInfo(
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

func (m *memoryStore) createWhatsAppTranscript(transcript whatsappTranscript) whatsappTranscript {
	m.mu.Lock()
	defer m.mu.Unlock()

	transcript.ID = fmt.Sprintf("whatsapp_transcript_%06d", m.nextWhatsAppTranscriptID)
	m.nextWhatsAppTranscriptID++
	m.whatsappTranscripts = append(m.whatsappTranscripts, transcript)
	logInfo(
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
	switch p.channel {
	case "sms":
		providerID = "mock_sms"
	case "voice":
		providerID = "mock_voice"
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

func preferredRecipientChannel(channels []string) string {
	for _, channel := range channels {
		if channel == "sms" {
			return "sms"
		}
	}
	if len(channels) > 0 {
		return channels[0]
	}
	return ""
}

func normalizeVoiceLanguages(languages []string) ([]string, string, string) {
	defaultLanguages := []string{"en", "tw", "ga", "ee", "dag", "ha"}
	if len(languages) == 0 {
		return append([]string(nil), defaultLanguages...), "", ""
	}

	result := make([]string, 0, len(languages))
	seen := map[string]bool{}
	for _, language := range languages {
		normalized, ok := normalizeVoiceLanguage(language)
		if !ok {
			return nil, "invalid_language", "languages must include only en, tw, ga, ee, dag, or ha"
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return nil, "missing_language", "at least one language is required"
	}
	return result, "", ""
}

func normalizeVoiceLanguage(value string) (string, bool) {
	value = normalizeQueryValue(value)
	switch value {
	case "", "english":
		value = "en"
	case "ak", "akan", "twi":
		value = "tw"
	case "gaa":
		value = "ga"
	case "ewe":
		value = "ee"
	case "dagbani":
		value = "dag"
	case "hausa":
		value = "ha"
	}
	return value, allowedVoiceLanguages[value]
}

func normalizeVoiceSource(value string) (string, bool) {
	value = normalizeQueryValue(value)
	if value == "" {
		return "tts_sandbox", true
	}
	if value == "tts_sandbox" || value == "recorded_audio" {
		return value, true
	}
	return "", false
}

func normalizeVoiceReviewAction(value string) string {
	value = normalizeQueryValue(value)
	switch value {
	case "approve", "approved":
		return "approve"
	case "reject", "rejected":
		return "reject"
	default:
		return ""
	}
}

func voiceReviewStatus(action string) string {
	if action == "approve" {
		return "approved"
	}
	return "rejected"
}

func normalizeVoiceRecipients(recipients []voiceRecipient) ([]voiceRecipient, string, string) {
	if len(recipients) == 0 {
		return nil, "missing_recipients", "at least one voice recipient is required"
	}

	result := make([]voiceRecipient, 0, len(recipients))
	for _, recipient := range recipients {
		recipient.RecipientID = normalizeID(recipient.RecipientID)
		recipient.Phone = strings.TrimSpace(recipient.Phone)
		language, ok := normalizeVoiceLanguage(recipient.Language)
		if !ok {
			return nil, "invalid_language", "recipient language must be en, tw, ga, ee, dag, or ha"
		}
		recipient.Language = language
		if recipient.RecipientID == "" && recipient.Phone == "" {
			return nil, "missing_recipient", "each voice recipient requires recipientId or phone"
		}
		result = append(result, recipient)
	}
	return result, "", ""
}

func voiceRecipientRef(recipient voiceRecipient) string {
	if recipient.RecipientID != "" {
		return recipient.RecipientID
	}
	if recipient.Phone != "" {
		if len(recipient.Phone) <= 4 {
			return "phone:" + recipient.Phone
		}
		return "phone:..." + recipient.Phone[len(recipient.Phone)-4:]
	}
	return "anonymous"
}

func copyVoiceAlertAsset(asset voiceAlertAsset) voiceAlertAsset {
	asset.Variants = append([]voiceVariant(nil), asset.Variants...)
	return asset
}

func voiceVariantForLanguage(asset voiceAlertAsset, language string) (voiceVariant, bool) {
	for _, variant := range asset.Variants {
		if variant.Language == language {
			return variant, true
		}
	}
	return voiceVariant{}, false
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

func voiceMessageForAlert(language string, alert citizenAlert) string {
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

func voiceAccessibilityChecks(message string, alert citizenAlert) []string {
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

func providerName(provider notificationProvider) string {
	if provider == nil {
		return "missing"
	}
	switch provider := provider.(type) {
	case mockProvider:
		return "mock_" + provider.channel
	case disabledProvider:
		return provider.channel + "_disabled"
	default:
		return fmt.Sprintf("%T", provider)
	}
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

func normalizeUSSDLanguage(defaultLanguage string, text string) string {
	tokens := ussdTokens(text)
	if len(tokens) > 0 {
		if language, ok := ussdLanguageFromToken(tokens[0]); ok {
			return language
		}
	}
	return normalizeLanguage(defaultLanguage)
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

func inclusiveLocation(location *coordinates, locationTokens []string) (coordinates, string) {
	if location != nil && location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180 {
		return *location, "caller shared approximate coordinates"
	}
	locationLabel := strings.TrimSpace(strings.Join(locationTokens, " "))
	if locationLabel == "" {
		locationLabel = "caller did not provide location details; use phone follow-up"
	}
	return coordinates{Lat: 5.5600, Lng: -0.2057}, locationLabel
}

func parseSMSReport(request smsInboundRequest, phoneRef string, linkedProfile bool, now time.Time) (inclusiveAccessReport, bool, string) {
	fields := strings.Fields(request.Body)
	if len(fields) < 3 {
		return inclusiveAccessReport{}, false, smsReportUsage()
	}

	hazard := normalizeSMSHazard(fields[1])
	if !allowedHazards[hazard] {
		return inclusiveAccessReport{}, false, smsReportUsage()
	}

	urgency := normalizeSMSUrgency(fields[2])
	if urgency == "" {
		return inclusiveAccessReport{}, false, smsReportUsage()
	}

	description := strings.TrimSpace(strings.Join(fields[3:], " "))
	if description == "" {
		description = fmt.Sprintf("SMS emergency report: %s with %s urgency", hazard, urgency)
	} else {
		description = fmt.Sprintf("SMS emergency report: %s", description)
	}

	location, locationLabel := inclusiveLocation(request.Location, nil)
	return inclusiveAccessReport{
		Channel:       "sms",
		Type:          hazard,
		Urgency:       urgency,
		Description:   description,
		Location:      location,
		LocationLabel: locationLabel,
		PhoneRef:      phoneRef,
		ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile: linkedProfile,
		Status:        "queued",
		CreatedAt:     now,
	}, true, ""
}

func parseWhatsAppDirectReport(request whatsappInboundRequest, phoneRef string, linkedProfile bool, now time.Time) (inclusiveAccessReport, bool, string) {
	fields := strings.Fields(request.Body)
	if len(fields) < 3 {
		return inclusiveAccessReport{}, false, whatsappReportUsage()
	}

	hazard := normalizeSMSHazard(fields[1])
	if !allowedHazards[hazard] {
		return inclusiveAccessReport{}, false, whatsappReportUsage()
	}

	urgency := normalizeSMSUrgency(fields[2])
	if urgency == "" {
		return inclusiveAccessReport{}, false, whatsappReportUsage()
	}

	conversation := whatsappConversation{
		Channel:       "whatsapp",
		PhoneRef:      phoneRef,
		ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
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

func buildWhatsAppReport(request whatsappInboundRequest, conversation whatsappConversation, phoneRef string, linkedProfile bool, now time.Time, details string) inclusiveAccessReport {
	details = strings.TrimSpace(details)
	location, locationLabel := inclusiveLocation(request.Location, strings.Fields(details))
	mediaRefs := whatsappMediaRefs(request.Media)
	description := fmt.Sprintf("WhatsApp emergency report: %s with %s urgency. Location note: %s.", conversation.Hazard, conversation.Urgency, locationLabel)
	if details != "" {
		description = fmt.Sprintf("WhatsApp emergency report: %s with %s urgency. Details: %s.", conversation.Hazard, conversation.Urgency, details)
	}
	if len(mediaRefs) > 0 {
		description = fmt.Sprintf("%s Media attachments received: %d.", description, len(mediaRefs))
	}
	return inclusiveAccessReport{
		Channel:       "whatsapp",
		Type:          conversation.Hazard,
		Urgency:       conversation.Urgency,
		Description:   description,
		Location:      location,
		LocationLabel: locationLabel,
		PhoneRef:      phoneRef,
		ProfileID:     profileIDForLog(request.ProfileID, linkedProfile),
		LinkedProfile: linkedProfile,
		Status:        "queued",
		Media:         mediaRefs,
		CreatedAt:     now,
	}
}

func normalizeWhatsAppMedia(media []whatsappMedia) []whatsappMedia {
	result := make([]whatsappMedia, 0, len(media))
	for _, item := range media {
		item.ID = strings.TrimSpace(item.ID)
		item.URL = strings.TrimSpace(item.URL)
		item.ContentType = strings.TrimSpace(item.ContentType)
		item.Caption = strings.TrimSpace(item.Caption)
		if item.ID == "" && item.URL == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}

func whatsappMediaRefs(media []whatsappMedia) []string {
	refs := make([]string, 0, len(media))
	for _, item := range media {
		switch {
		case item.ID != "":
			refs = append(refs, item.ID)
		case item.URL != "":
			refs = append(refs, item.URL)
		}
	}
	return refs
}

func whatsappConversationKey(phoneRef string, profileID string, linkedProfile bool) string {
	return phoneRef
}

func isWhatsAppTopLevelCommand(command string) bool {
	switch command {
	case "ALERT", "ALERTS", "RISK", "GUIDE", "GUIDES", "SHELTER", "SHELTERS", "HELP", "112", "REPORT", "CANCEL", "MENU", "START", "HI", "HELLO":
		return true
	default:
		return false
	}
}

func whatsappCommandArg(body string, index int) string {
	fields := strings.Fields(body)
	if index < 0 || index >= len(fields) {
		return ""
	}
	return fields[index]
}

func firstToken(body string) string {
	return whatsappCommandArg(body, 0)
}

func whatsappRetentionUntil(now time.Time) time.Time {
	return now.Add(90 * 24 * time.Hour)
}

func whatsappMessageSummary(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	command := smsCommandName(body)
	if command == "" {
		return fmt.Sprintf("provided:%d_chars", len(body))
	}
	return fmt.Sprintf("command:%s:%d_chars", command, len(body))
}

func whatsappMediaSummary(media []whatsappMedia) string {
	if len(media) == 0 {
		return ""
	}
	return fmt.Sprintf("attachments:%d", len(media))
}

func whatsappHelpMessage() string {
	return "NADAA WhatsApp commands: ALERTS, RISK, REPORT, SHELTER, GUIDE FLOOD, HELP, or 112. To report: REPORT FLOOD HIGH your location/details, or send REPORT and answer the prompts."
}

func whatsappReportUsage() string {
	return "Use: REPORT FLOOD HIGH your location/details. Hazards: FLOOD FIRE MEDICAL ROAD OTHER. Urgency: LOW MODERATE HIGH LIFE."
}

func whatsappHazardPrompt() string {
	return "What type of emergency are you reporting? Reply FLOOD, FIRE, MEDICAL, ROAD, SECURITY, STORM, or OTHER."
}

func whatsappUrgencyPrompt() string {
	return "How urgent is it? Reply LOW, MODERATE, HIGH, or LIFE. Call 112 now if life is in immediate danger."
}

func whatsappLocationPrompt() string {
	return "Please send your location pin, nearest landmark, or a short description. You can attach a photo or voice note if safe."
}

func riskCheckMessage(language string, hasLocation bool, alerts []citizenAlert) string {
	prefix := "Share your location pin for a more specific risk check. "
	if hasLocation {
		prefix = "Location received for this WhatsApp risk check. "
	}
	if len(alerts) == 0 {
		return prefix + "No current NADAA alerts are active in the notification feed. Stay alert and call 112 for immediate danger."
	}
	alert := alerts[0]
	return fmt.Sprintf("%sCurrent NADAA signal: %s for %s. %s", prefix, alert.Title, alert.TargetLabel, alert.RecommendedAction)
}

func emergencyGuideMessage(language string, hazard string) string {
	switch hazard {
	case "fire":
		return "Fire guide: leave the area, keep exits clear, avoid smoke, do not use lifts, and call 112 if flames or heavy smoke are visible."
	case "medical_emergency":
		return "Medical guide: call 112 for serious injury, keep the person still, share location clearly, and do not move them unless the area is unsafe."
	case "road_crash":
		return "Road crash guide: move away from traffic if safe, call 112, warn approaching vehicles, and do not move injured people unless there is immediate danger."
	case "storm":
		return "Storm guide: stay indoors, avoid trees and power lines, secure loose items if safe, and follow NADAA alerts."
	default:
		if language == "tw" {
			return "Nsuyiri akwankyerɛ: kɔ baabi a ɛkorɔn, kwati nsuo a ɛsen, sie nkrataa, na frɛ 112 sɛ nkwa wɔ asiane mu."
		}
		return "Flood guide: move to higher ground, avoid drains and floodwater, keep documents dry, follow official alerts, and call 112 for immediate danger."
	}
}

func normalizeSMSHazard(value string) string {
	value = normalizeQueryValue(value)
	switch value {
	case "road", "crash", "accident":
		return "road_crash"
	case "medical", "ambulance":
		return "medical_emergency"
	case "security":
		return "security_incident"
	default:
		return value
	}
}

func normalizeSMSUrgency(value string) string {
	value = normalizeQueryValue(value)
	switch value {
	case "life", "life_threatening", "emergency":
		return "life_threatening"
	case "high", "moderate", "low":
		return value
	default:
		return ""
	}
}

func providerOrDefault(provider string, fallback string) string {
	provider = normalizeQueryValue(provider)
	if provider == "" {
		return fallback
	}
	return provider
}

func profileIDForLog(profileID string, linked bool) string {
	if linked {
		return profileID
	}
	return ""
}

func phoneRef(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "phone:unknown"
	}
	if len(phone) <= 4 {
		return "phone:" + phone
	}
	return "phone:..." + phone[len(phone)-4:]
}

func languageMenu() string {
	return "Select language / Paw kasa:\n1 English\n2 Twi\n3 Ga\n4 Ewe\n5 Dagbani\n6 Hausa"
}

func mainMenu(language string) string {
	switch language {
	case "tw":
		return "NADAA menu:\n1 Kɔkɔbɔ\n2 Bɔ amanneɛ\n3 Dwanekɔbea\n4 Frɛ 112"
	case "ga":
		return "NADAA menu:\n1 Alerts\n2 Report emergency\n3 Shelter\n4 Call 112"
	default:
		return "NADAA menu:\n1 Current alerts\n2 Report emergency\n3 Find shelter\n4 112 guidance"
	}
}

func hazardMenu(language string) string {
	switch language {
	case "tw":
		return "Paw asiane:\n1 Nsuyiri\n2 Ogya\n3 Ayaresa\n4 Kar akwanhyia\n5 Foforo"
	default:
		return "Select emergency type:\n1 Flood\n2 Fire\n3 Medical\n4 Road crash\n5 Other"
	}
}

func urgencyMenu(language string) string {
	switch language {
	case "tw":
		return "Ɛyɛ den sɛn?\n1 Kakra\n2 Mfinimfini\n3 Den\n4 Ɛhaw nkwa"
	default:
		return "Select urgency:\n1 Low\n2 Moderate\n3 High\n4 Life-threatening"
	}
}

func localizedMessage(language string, key string) string {
	switch key {
	case "provider_error":
		if language == "tw" {
			return "Nkitahodie no nni hɔ seesei. Sɛ ɛyɛ asianeɛ a, frɛ 112."
		}
		return "The channel is temporarily unavailable. If this is life-threatening, call 112."
	default:
		return smsHelpMessage()
	}
}

func alertSummaryMessage(language string, alerts []citizenAlert) string {
	if len(alerts) == 0 {
		if language == "tw" {
			return "Kɔkɔbɔ foforo biara nni hɔ seesei. Sɛ ɛyɛ asianeɛ a, frɛ 112."
		}
		return "No current NADAA alerts. If this is life-threatening, call 112."
	}
	alert := alerts[0]
	return fmt.Sprintf("%s: %s. %s", alert.Title, alert.TargetLabel, alert.RecommendedAction)
}

func smsAlertMessage(alerts []citizenAlert) string {
	return alertSummaryMessage("en", alerts)
}

func shelterMessage(language string) string {
	if language == "tw" {
		return "Bɛn dwanekɔbea: Accra Sports Hall, Osu Community Centre. Fa wo ho kɔ baabi a ɛyɛ banbɔ. Frɛ 112 sɛ ɛyɛ asianeɛ."
	}
	return "Nearby shelters: Accra Sports Hall, Osu Community Centre. Move to safe high ground when directed. Call 112 for immediate danger."
}

func guidance112Message(language string) string {
	if language == "tw" {
		return "Sɛ nkwa wɔ asiane mu a, frɛ 112 ntɛm. Ka baabi a wowɔ, asiane no, ne nnipa dodow."
	}
	return "If life is in immediate danger, call 112 now. Share your location, emergency type, and people affected."
}

func reportConfirmationMessage(language string, report inclusiveAccessReport) string {
	reference := report.ID
	if report.IncidentReference != "" {
		reference = report.IncidentReference
	}
	if language == "tw" {
		return fmt.Sprintf("Yɛagye wo amanneɛ no: %s. Frɛ 112 sɛ nkwa wɔ asiane mu.", reference)
	}
	return fmt.Sprintf("NADAA report received: %s. Call 112 if life is in immediate danger.", reference)
}

func smsHelpMessage() string {
	return "NADAA SMS commands: ALERTS, SHELTER, HELP, or REPORT FLOOD HIGH your location/details. Call 112 for immediate danger."
}

func smsReportUsage() string {
	return "Use: REPORT FLOOD HIGH your location/details. Hazards: FLOOD FIRE MEDICAL ROAD OTHER. Urgency: LOW MODERATE HIGH LIFE."
}

func smsCommandName(body string) string {
	parts := strings.Fields(strings.TrimSpace(body))
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}

func logTextSummary(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("provided:%d_chars", len(value))
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
		logError("write json response failed", "status", status, "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func logInfo(message string, fields ...any) {
	logWithLevel("INFO", message, fields...)
}

func logWarn(message string, fields ...any) {
	logWithLevel("WARN", message, fields...)
}

func logError(message string, fields ...any) {
	logWithLevel("ERROR", message, fields...)
}

func logWithLevel(level string, message string, fields ...any) {
	pairs := make([]string, 0, (len(fields)+1)/2)
	for index := 0; index < len(fields); index += 2 {
		key := strings.TrimSpace(fmt.Sprint(fields[index]))
		if key == "" {
			key = fmt.Sprintf("field_%d", index/2)
		}
		key = strings.ReplaceAll(key, " ", "_")

		value := "missing_value"
		if index+1 < len(fields) {
			value = fmt.Sprint(fields[index+1])
		}
		value = strings.ReplaceAll(value, "\n", " ")
		value = strings.ReplaceAll(value, "\r", " ")
		pairs = append(pairs, fmt.Sprintf("%s=%q", key, value))
	}

	if len(pairs) == 0 {
		log.Printf("level=%s msg=%q", level, message)
		return
	}
	log.Printf("level=%s msg=%q %s", level, message, strings.Join(pairs, " "))
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsFromEnv()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Cache-Control", "no-store")
}

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
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
