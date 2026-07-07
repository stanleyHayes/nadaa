package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/client"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

func TestListAlertFeedIncludesCurrentAndExpiredAlerts(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/alerts?includeExpired=true", nil)

	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.CitizenAlertListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Alerts) < 2 {
		t.Fatalf("expected current and expired alerts, got %#v", payload.Alerts)
	}

	statuses := map[string]bool{}
	for _, alert := range payload.Alerts {
		statuses[alert.Status] = true
		if alert.Target.Label == "" || alert.TargetLabel == "" {
			t.Fatalf("expected target labels, got %#v", alert)
		}
	}
	if !statuses["current"] || !statuses["expired"] {
		t.Fatalf("expected current and expired statuses, got %#v", statuses)
	}
}

func TestDefaultAlertFeedOnlyReturnsCurrentAlerts(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/alerts", nil)

	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.CitizenAlertListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Alerts) == 0 {
		t.Fatal("expected current alerts")
	}
	for _, alert := range payload.Alerts {
		if alert.Status != "current" {
			t.Fatalf("expected current alert, got %#v", alert)
		}
	}
}

func TestDeliverAlertLogsMockPushAndSMS(t *testing.T) {
	srv := newTestServer()
	body := `{"recipientId":"usr_demo_citizen","phone":"+233200000000","pushToken":"ExponentPushToken-demo","channels":["push","sms"]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_current_flood/deliver", bytes.NewBufferString(body))
	request.SetPathValue("id", "alert_feed_current_flood")

	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload models.DeliveryResponse
	decodeResponse(t, response, &payload)
	if len(payload.Attempts) != 2 {
		t.Fatalf("expected two attempts, got %#v", payload.Attempts)
	}
	for _, attempt := range payload.Attempts {
		if attempt.Status != "delivered" {
			t.Fatalf("expected delivered attempt, got %#v", attempt)
		}
		if attempt.Provider != "mock_push" && attempt.Provider != "mock_sms" {
			t.Fatalf("expected mock provider, got %#v", attempt)
		}
	}

	logsResponse := httptest.NewRecorder()
	logsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/delivery-logs?alertId=alert_feed_current_flood", nil)
	srv.listDeliveryLogsHandler(logsResponse, logsRequest)

	var logs models.DeliveryLogListResponse
	decodeResponse(t, logsResponse, &logs)
	if len(logs.Logs) != 2 {
		t.Fatalf("expected persisted delivery logs, got %#v", logs.Logs)
	}
}

func TestSMSDisabledLogsSkippedAttempt(t *testing.T) {
	srv := newTestServer()
	srv.providers["sms"] = models.DisabledProvider{Channel: "sms", Reason: "sms provider disabled"}
	body := `{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_current_flood/deliver", bytes.NewBufferString(body))
	request.SetPathValue("id", "alert_feed_current_flood")

	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload models.DeliveryResponse
	decodeResponse(t, response, &payload)
	if len(payload.Attempts) != 1 {
		t.Fatalf("expected one sms attempt, got %#v", payload.Attempts)
	}
	if payload.Attempts[0].Status != "skipped" || payload.Attempts[0].Provider != "sms_disabled" {
		t.Fatalf("expected skipped sms attempt, got %#v", payload.Attempts[0])
	}
}

func TestUSSDMenuAndAlertSummary(t *testing.T) {
	srv := newTestServer()

	menuResponse := httptest.NewRecorder()
	menuRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/ussd", bytes.NewBufferString(`{"sessionId":"ussd_001","phone":"+233200000001","text":""}`))
	srv.ussdWebhookHandler(menuResponse, menuRequest)

	if menuResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, menuResponse.Code, menuResponse.Body.String())
	}

	var menuPayload models.USSDWebhookResponse
	decodeResponse(t, menuResponse, &menuPayload)
	if menuPayload.Action != "continue" || !strings.Contains(menuPayload.Message, "Select language") {
		t.Fatalf("expected language menu, got %#v", menuPayload)
	}

	alertResponse := httptest.NewRecorder()
	alertRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/ussd", bytes.NewBufferString(`{"sessionId":"ussd_001","phone":"+233200000001","text":"1*1"}`))
	srv.ussdWebhookHandler(alertResponse, alertRequest)

	var alertPayload models.USSDWebhookResponse
	decodeResponse(t, alertResponse, &alertPayload)
	if alertPayload.Action != "end" || !strings.Contains(alertPayload.Message, "Severe flood warning") {
		t.Fatalf("expected current alert summary, got %#v", alertPayload)
	}

	logs := srv.store.ListAccessLogs(models.AccessLogFilters{Channel: "ussd", Intent: "current_alerts"})
	if len(logs) != 1 || logs[0].PhoneRef != "phone:...0001" {
		t.Fatalf("expected current-alerts access log, got %#v", logs)
	}
}

func TestUSSDReportQueuesWhenIncidentServiceIsNotConfigured(t *testing.T) {
	srv := newTestServer()
	body := `{"sessionId":"ussd_002","phone":"+233200000002","text":"1*2*1*3","profileId":"usr_sms_001","linkProfile":true,"location":{"lat":5.579,"lng":-0.212}}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/ussd", bytes.NewBufferString(body))

	srv.ussdWebhookHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.USSDWebhookResponse
	decodeResponse(t, response, &payload)
	if payload.Report == nil || payload.Report.Status != "queued" || !payload.Report.LinkedProfile {
		t.Fatalf("expected queued linked report, got %#v", payload)
	}
	if payload.Report.ProfileID != "usr_sms_001" || payload.Report.PhoneRef != "phone:...0002" {
		t.Fatalf("expected profile-linked masked report, got %#v", payload.Report)
	}

	logs := srv.store.ListAccessLogs(models.AccessLogFilters{Channel: "ussd", Intent: "report_emergency"})
	if len(logs) != 1 || logs[0].Status != "queued" {
		t.Fatalf("expected queued report access log, got %#v", logs)
	}
}

func TestSMSInboundReportSubmitsToIncidentService(t *testing.T) {
	var incidentPayload models.IncidentIntakeRequest
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/incidents" {
			t.Fatalf("unexpected incident request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&incidentPayload); err != nil {
			t.Fatalf("decode incident payload: %v", err)
		}
		utils.WriteJSON(w, http.StatusCreated, models.IncidentIntakeResponse{ID: "inc_sms_001", Reference: "INC-SMS-001", Status: "reported"})
	}))
	defer incidentServer.Close()

	srv := newTestServer()
	srv.incidentClient = client.NewIncidentServiceClient(incidentServer.URL + "/api/v1")
	body := `{"from":"+233200000003","body":"REPORT FLOOD HIGH road flooded near Kaneshie","profileId":"usr_sms_002","linkProfile":true,"location":{"lat":5.566,"lng":-0.242}}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body))

	srv.smsInboundHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload models.SMSInboundResponse
	decodeResponse(t, response, &payload)
	if payload.Report == nil || payload.Report.Status != "submitted" || payload.Report.IncidentReference != "INC-SMS-001" {
		t.Fatalf("expected submitted SMS report, got %#v", payload)
	}
	if incidentPayload.Type != "flood" || incidentPayload.Urgency != "high" || incidentPayload.Reporter == nil || incidentPayload.Reporter.Phone != "+233200000003" {
		t.Fatalf("unexpected incident handoff payload: %#v", incidentPayload)
	}
	if !strings.Contains(payload.Message, "INC-SMS-001") {
		t.Fatalf("expected incident reference in SMS response, got %q", payload.Message)
	}
}

func TestSMSProviderErrorIsLogged(t *testing.T) {
	srv := newTestServer()
	body := `{"from":"+233200000004","provider":"sms-test","providerError":"signature check failed"}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body))

	srv.smsInboundHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload models.SMSInboundResponse
	decodeResponse(t, response, &payload)
	if payload.Log.Status != "failed" || payload.Log.Intent != "provider_error" || payload.Log.ProviderError == "" {
		t.Fatalf("expected failed provider-error log, got %#v", payload)
	}
}

func TestWhatsAppConversationReportQueuesWithLocationAndMedia(t *testing.T) {
	srv := newTestServer()

	startResponse := httptest.NewRecorder()
	startRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/webhook", bytes.NewBufferString(`{"from":"+233200000005","body":"REPORT","provider":"wa-test"}`))
	srv.whatsappWebhookHandler(startResponse, startRequest)

	if startResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, startResponse.Code, startResponse.Body.String())
	}
	var startPayload models.WhatsAppInboundResponse
	decodeResponse(t, startResponse, &startPayload)
	if startPayload.Conversation.State != "awaiting_report_hazard" || !strings.Contains(startPayload.Message, "What type") {
		t.Fatalf("expected hazard prompt, got %#v", startPayload)
	}

	hazardResponse := httptest.NewRecorder()
	hazardRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/webhook", bytes.NewBufferString(`{"from":"+233200000005","body":"FLOOD","provider":"wa-test"}`))
	srv.whatsappWebhookHandler(hazardResponse, hazardRequest)

	var hazardPayload models.WhatsAppInboundResponse
	decodeResponse(t, hazardResponse, &hazardPayload)
	if hazardPayload.Conversation.State != "awaiting_report_urgency" || hazardPayload.Conversation.Hazard != "flood" {
		t.Fatalf("expected urgency prompt state, got %#v", hazardPayload.Conversation)
	}

	urgencyResponse := httptest.NewRecorder()
	urgencyRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/webhook", bytes.NewBufferString(`{"from":"+233200000005","body":"HIGH","provider":"wa-test"}`))
	srv.whatsappWebhookHandler(urgencyResponse, urgencyRequest)

	var urgencyPayload models.WhatsAppInboundResponse
	decodeResponse(t, urgencyResponse, &urgencyPayload)
	if urgencyPayload.Conversation.State != "awaiting_report_location" || urgencyPayload.Conversation.Urgency != "high" {
		t.Fatalf("expected location prompt state, got %#v", urgencyPayload.Conversation)
	}

	completeResponse := httptest.NewRecorder()
	completeBody := `{"from":"+233200000005","body":"water rising near Circle","provider":"wa-test","profileId":"usr_wa_001","linkProfile":true,"location":{"lat":5.566,"lng":-0.242},"media":[{"id":"wa_media_001","contentType":"image/jpeg"}]}`
	completeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/webhook", bytes.NewBufferString(completeBody))
	srv.whatsappWebhookHandler(completeResponse, completeRequest)

	if completeResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, completeResponse.Code, completeResponse.Body.String())
	}
	var completePayload models.WhatsAppInboundResponse
	decodeResponse(t, completeResponse, &completePayload)
	if completePayload.Report == nil || completePayload.Report.Status != "queued" || completePayload.Report.Channel != "whatsapp" {
		t.Fatalf("expected queued WhatsApp report, got %#v", completePayload)
	}
	if completePayload.Conversation.State != "idle" || len(completePayload.Report.Media) != 1 || completePayload.Report.Media[0] != "wa_media_001" {
		t.Fatalf("expected completed conversation with media, got %#v", completePayload)
	}
	if len(completePayload.TranscriptIDs) != 2 {
		t.Fatalf("expected inbound/outbound transcript ids, got %#v", completePayload.TranscriptIDs)
	}

	logs := srv.store.ListAccessLogs(models.AccessLogFilters{Channel: "whatsapp", Intent: "report_emergency"})
	hasQueuedLog := false
	for _, log := range logs {
		if log.Status == "queued" {
			hasQueuedLog = true
		}
	}
	if len(logs) < 4 || !hasQueuedLog {
		t.Fatalf("expected WhatsApp report access logs, got %#v", logs)
	}
	if len(srv.store.WhatsAppTranscripts()) < 8 {
		t.Fatalf("expected privacy-safe transcript summaries, got %#v", srv.store.WhatsAppTranscripts())
	}
}

func TestWhatsAppDirectReportSubmitsToIncidentService(t *testing.T) {
	var incidentPayload models.IncidentIntakeRequest
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/incidents" {
			t.Fatalf("unexpected incident request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&incidentPayload); err != nil {
			t.Fatalf("decode incident payload: %v", err)
		}
		utils.WriteJSON(w, http.StatusCreated, models.IncidentIntakeResponse{ID: "inc_wa_001", Reference: "INC-WA-001", Status: "reported"})
	}))
	defer incidentServer.Close()

	srv := newTestServer()
	srv.incidentClient = client.NewIncidentServiceClient(incidentServer.URL + "/api/v1")
	body := `{"from":"+233200000006","body":"REPORT FLOOD HIGH water entering homes near Odaw","provider":"wa-test","profileId":"usr_wa_002","linkProfile":true,"location":{"lat":5.579,"lng":-0.212},"media":[{"id":"wa_media_002","contentType":"image/jpeg"}]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(body))

	srv.whatsappWebhookHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload models.WhatsAppInboundResponse
	decodeResponse(t, response, &payload)
	if payload.Report == nil || payload.Report.Status != "submitted" || payload.Report.IncidentReference != "INC-WA-001" {
		t.Fatalf("expected submitted WhatsApp report, got %#v", payload)
	}
	if incidentPayload.Type != "flood" || incidentPayload.Urgency != "high" || incidentPayload.Reporter == nil || incidentPayload.Reporter.Phone != "+233200000006" {
		t.Fatalf("unexpected incident handoff payload: %#v", incidentPayload)
	}
	if len(incidentPayload.Media) != 1 || incidentPayload.Media[0] != "wa_media_002" {
		t.Fatalf("expected WhatsApp media handoff, got %#v", incidentPayload.Media)
	}
}

func TestVoiceAlertGenerationReviewAndDelivery(t *testing.T) {
	srv := newTestServer()

	createResponse := httptest.NewRecorder()
	createBody := `{"alertId":"alert_feed_current_flood","languages":["en","tw","ha"],"workflowRequestedBy":"dispatcher_001"}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts", bytes.NewBufferString(createBody))
	srv.createVoiceAlertHandler(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var created models.VoiceAlertResponse
	decodeResponse(t, createResponse, &created)
	if created.Asset.ID == "" || created.Asset.ReviewStatus != "pending_review" || len(created.Asset.Variants) != 3 {
		t.Fatalf("expected generated pending voice asset with variants, got %#v", created.Asset)
	}
	for _, variant := range created.Asset.Variants {
		if variant.AudioURL == "" || variant.DurationSeconds == 0 || !strings.Contains(variant.MessageText, "112") {
			t.Fatalf("expected accessible generated variant, got %#v", variant)
		}
	}

	reviewResponse := httptest.NewRecorder()
	reviewBody := `{"action":"approve","reviewer":"nadmo_voice_reviewer","note":"Checked low-literacy script"}`
	reviewRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts/"+created.Asset.ID+"/review", bytes.NewBufferString(reviewBody))
	reviewRequest.SetPathValue("id", created.Asset.ID)
	srv.reviewVoiceAlertHandler(reviewResponse, reviewRequest)

	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
	}

	var reviewed models.VoiceAlertResponse
	decodeResponse(t, reviewResponse, &reviewed)
	if reviewed.Asset.Status != "approved" || reviewed.Asset.ReviewStatus != "approved" || reviewed.Asset.Reviewer != "nadmo_voice_reviewer" {
		t.Fatalf("expected approved voice asset, got %#v", reviewed.Asset)
	}

	deliverResponse := httptest.NewRecorder()
	deliverBody := `{"recipients":[{"phone":"+233200000010","language":"en"},{"recipientId":"usr_voice_002","phone":"+233200000011","language":"tw"}]}`
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts/"+created.Asset.ID+"/deliver", bytes.NewBufferString(deliverBody))
	deliverRequest.SetPathValue("id", created.Asset.ID)
	srv.deliverVoiceAlertHandler(deliverResponse, deliverRequest)

	if deliverResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, deliverResponse.Code, deliverResponse.Body.String())
	}

	var delivered models.VoiceDeliveryResponse
	decodeResponse(t, deliverResponse, &delivered)
	if len(delivered.Attempts) != 2 {
		t.Fatalf("expected two voice attempts, got %#v", delivered.Attempts)
	}
	for _, attempt := range delivered.Attempts {
		if attempt.Channel != "voice" || attempt.Provider != "mock_voice" || attempt.Status != "delivered" {
			t.Fatalf("expected delivered mock voice attempt, got %#v", attempt)
		}
		if attempt.VoiceAssetID != created.Asset.ID || attempt.AudioURL == "" || attempt.Language == "" {
			t.Fatalf("expected voice delivery metadata, got %#v", attempt)
		}
	}

	logsResponse := httptest.NewRecorder()
	logsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/delivery-logs?channel=voice&alertId=alert_feed_current_flood", nil)
	srv.listDeliveryLogsHandler(logsResponse, logsRequest)

	var logs models.DeliveryLogListResponse
	decodeResponse(t, logsResponse, &logs)
	if len(logs.Logs) != 2 {
		t.Fatalf("expected persisted voice delivery logs, got %#v", logs.Logs)
	}
}

func TestVoiceDeliveryRequiresApprovedAsset(t *testing.T) {
	srv := newTestServer()

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts", bytes.NewBufferString(`{"alertId":"alert_feed_current_flood","languages":["en"]}`))
	srv.createVoiceAlertHandler(createResponse, createRequest)

	var created models.VoiceAlertResponse
	decodeResponse(t, createResponse, &created)

	deliverResponse := httptest.NewRecorder()
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts/"+created.Asset.ID+"/deliver", bytes.NewBufferString(`{"recipients":[{"phone":"+233200000010","language":"en"}]}`))
	deliverRequest.SetPathValue("id", created.Asset.ID)
	srv.deliverVoiceAlertHandler(deliverResponse, deliverRequest)

	if deliverResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, deliverResponse.Code, deliverResponse.Body.String())
	}
}

func TestGenericDeliveryRejectsVoiceChannel(t *testing.T) {
	srv := newTestServer()
	body := `{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["voice"]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_current_flood/deliver", bytes.NewBufferString(body))
	request.SetPathValue("id", "alert_feed_current_flood")

	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if !strings.Contains(response.Body.String(), "voice_requires_asset") {
		t.Fatalf("expected voice_requires_asset error, got %s", response.Body.String())
	}
}

func TestDeliverRejectsUnsupportedChannel(t *testing.T) {
	srv := newTestServer()
	body := `{"recipientId":"usr_demo_citizen","channels":["email"]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_current_flood/deliver", bytes.NewBufferString(body))
	request.SetPathValue("id", "alert_feed_current_flood")

	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestAlertFeedRejectsInvalidStatus(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/alerts?status=missing", nil)

	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func newTestServer() *Server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8090"}
	return NewServer(
		store.NewMemoryStore(now),
		nil,
		nil,
		map[string]models.NotificationProvider{
			"push":  models.MockProvider{Channel: "push"},
			"sms":   models.MockProvider{Channel: "sms"},
			"voice": models.MockProvider{Channel: "voice"},
		},
		func() time.Time { return now },
		cfg,
	)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
