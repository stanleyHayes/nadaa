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

func TestSMSInboundDeduplicatesProviderMessage(t *testing.T) {
	incidentCalls := 0
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		incidentCalls++
		utils.WriteJSON(w, http.StatusCreated, models.IncidentIntakeResponse{ID: "inc_sms_dedup", Reference: "INC-SMS-DEDUP", Status: "reported"})
	}))
	defer incidentServer.Close()

	srv := newTestServer()
	srv.incidentClient = client.NewIncidentServiceClient(incidentServer.URL + "/api/v1")
	body := `{"from":"+233200000030","body":"REPORT FLOOD HIGH road flooded","provider":"sms-test","providerMessageId":"msg-dedup-1","location":{"lat":5.566,"lng":-0.242}}`

	firstResponse := httptest.NewRecorder()
	srv.smsInboundHandler(firstResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body)))
	if firstResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, firstResponse.Code, firstResponse.Body.String())
	}
	var first models.SMSInboundResponse
	decodeResponse(t, firstResponse, &first)
	if first.Report == nil || first.Report.Status != "submitted" {
		t.Fatalf("expected submitted first report, got %#v", first)
	}

	// A provider retry of the same message returns the existing report and must
	// not create a second incident.
	secondResponse := httptest.NewRecorder()
	srv.smsInboundHandler(secondResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body)))
	if secondResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, secondResponse.Code, secondResponse.Body.String())
	}
	var second models.SMSInboundResponse
	decodeResponse(t, secondResponse, &second)
	if second.Report == nil || second.Report.ID != first.Report.ID {
		t.Fatalf("expected the existing report for a retried message, got %#v vs %#v", second.Report, first.Report)
	}
	if incidentCalls != 1 {
		t.Fatalf("expected one incident handoff for a retried message, got %d", incidentCalls)
	}
}

func TestWhatsAppDirectReportDeduplicatesProviderMessage(t *testing.T) {
	incidentCalls := 0
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		incidentCalls++
		utils.WriteJSON(w, http.StatusCreated, models.IncidentIntakeResponse{ID: "inc_wa_dedup", Reference: "INC-WA-DEDUP", Status: "reported"})
	}))
	defer incidentServer.Close()

	srv := newTestServer()
	srv.incidentClient = client.NewIncidentServiceClient(incidentServer.URL + "/api/v1")
	body := `{"from":"+233200000031","body":"REPORT FIRE LIFE shop burning","provider":"wa-test","providerMessageId":"wamid-dedup-1","location":{"lat":5.579,"lng":-0.212}}`

	var firstID string
	for attempt := 1; attempt <= 2; attempt++ {
		response := httptest.NewRecorder()
		srv.whatsappWebhookHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(body)))
		if response.Code != http.StatusAccepted {
			t.Fatalf("attempt %d: expected status %d, got %d: %s", attempt, http.StatusAccepted, response.Code, response.Body.String())
		}
		var payload models.WhatsAppInboundResponse
		decodeResponse(t, response, &payload)
		if payload.Report == nil {
			t.Fatalf("attempt %d: expected report, got %#v", attempt, payload)
		}
		if firstID == "" {
			firstID = payload.Report.ID
		} else if payload.Report.ID != firstID {
			t.Fatalf("expected the existing report for a retried message, got %q vs %q", payload.Report.ID, firstID)
		}
	}
	if incidentCalls != 1 {
		t.Fatalf("expected one incident handoff for a retried message, got %d", incidentCalls)
	}
}

func TestWhatsAppConversationExpiresResetsToIdle(t *testing.T) {
	currentNow := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8090", AllowMockActors: true}
	srv := NewServer(
		store.NewMemoryStore(currentNow),
		nil,
		nil,
		map[string]models.NotificationProvider{},
		models.SandboxCellBroadcastAdapter{},
		func() time.Time { return currentNow },
		cfg,
	)
	from := "+233200000032"

	startResponse := httptest.NewRecorder()
	srv.whatsappWebhookHandler(startResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(`{"from":"`+from+`","body":"REPORT"}`)))
	var startPayload models.WhatsAppInboundResponse
	decodeResponse(t, startResponse, &startPayload)
	if startPayload.Conversation.State != "awaiting_report_hazard" {
		t.Fatalf("expected report flow to start, got %#v", startPayload.Conversation)
	}

	// The session is abandoned past its expiry; a later stray message must not
	// complete the stale report.
	currentNow = currentNow.Add(25 * time.Hour)

	staleResponse := httptest.NewRecorder()
	srv.whatsappWebhookHandler(staleResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(`{"from":"`+from+`","body":"FLOOD"}`)))
	var stalePayload models.WhatsAppInboundResponse
	decodeResponse(t, staleResponse, &stalePayload)
	if stalePayload.Report != nil {
		t.Fatalf("expected no report from an expired conversation, got %#v", stalePayload.Report)
	}
	if stalePayload.Conversation.State != "idle" || stalePayload.Conversation.Intent != "invalid_selection" {
		t.Fatalf("expected expired conversation reset to idle, got %#v", stalePayload.Conversation)
	}
}

func TestWhatsAppConversationsAreKeyedByFullPhoneNumber(t *testing.T) {
	srv := newTestServer()
	// Both numbers share the same last-4 suffix.
	first := "+233200009999"
	second := "+233555009999"

	startResponse := httptest.NewRecorder()
	srv.whatsappWebhookHandler(startResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(`{"from":"`+first+`","body":"REPORT"}`)))
	var startPayload models.WhatsAppInboundResponse
	decodeResponse(t, startResponse, &startPayload)
	if startPayload.Conversation.State != "awaiting_report_hazard" {
		t.Fatalf("expected report flow to start, got %#v", startPayload.Conversation)
	}

	// A stranger sharing the last-4 suffix must not advance the first user's
	// conversation.
	strangerResponse := httptest.NewRecorder()
	srv.whatsappWebhookHandler(strangerResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(`{"from":"`+second+`","body":"FLOOD"}`)))
	var strangerPayload models.WhatsAppInboundResponse
	decodeResponse(t, strangerResponse, &strangerPayload)
	if strangerPayload.Conversation.State != "idle" || strangerPayload.Conversation.Intent != "invalid_selection" {
		t.Fatalf("expected the stranger's own idle conversation, got %#v", strangerPayload.Conversation)
	}
	if strangerPayload.Conversation.ID == startPayload.Conversation.ID {
		t.Fatalf("expected separate conversations for numbers sharing a last-4 suffix, got %#v", strangerPayload.Conversation)
	}

	continueResponse := httptest.NewRecorder()
	srv.whatsappWebhookHandler(continueResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/whatsapp/inbound", bytes.NewBufferString(`{"from":"`+first+`","body":"FLOOD"}`)))
	var continuePayload models.WhatsAppInboundResponse
	decodeResponse(t, continueResponse, &continuePayload)
	if continuePayload.Conversation.State != "awaiting_report_urgency" || continuePayload.Conversation.Hazard != "flood" {
		t.Fatalf("expected the original conversation to continue unaffected, got %#v", continuePayload.Conversation)
	}
}

func TestUSSDReportWithoutLocationOmitsCoordinates(t *testing.T) {
	var rawPayload map[string]any
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&rawPayload); err != nil {
			t.Fatalf("decode incident payload: %v", err)
		}
		utils.WriteJSON(w, http.StatusCreated, models.IncidentIntakeResponse{ID: "inc_ussd_noloc", Reference: "INC-USSD-NOLOC", Status: "reported"})
	}))
	defer incidentServer.Close()

	srv := newTestServer()
	srv.incidentClient = client.NewIncidentServiceClient(incidentServer.URL + "/api/v1")
	body := `{"sessionId":"ussd_noloc","phone":"+233200000033","text":"1*2*1*3"}`
	response := httptest.NewRecorder()
	srv.ussdWebhookHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/ussd", bytes.NewBufferString(body)))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.USSDWebhookResponse
	decodeResponse(t, response, &payload)
	if payload.Report == nil || payload.Report.Status != "submitted" {
		t.Fatalf("expected submitted location-less report, got %#v", payload)
	}
	if payload.Report.Location != nil {
		t.Fatalf("expected no fabricated coordinates for a location-less report, got %#v", payload.Report.Location)
	}
	if _, ok := rawPayload["location"]; ok {
		t.Fatalf("expected the incident handoff to omit location, got %v", rawPayload["location"])
	}
	if !strings.Contains(payload.Report.Description, "did not provide location") {
		t.Fatalf("expected the missing location noted in the description, got %q", payload.Report.Description)
	}
}

func TestZeroCoordinatesTreatedAsNoLocation(t *testing.T) {
	srv := newTestServer()
	body := `{"from":"+233200000034","body":"REPORT FLOOD HIGH road flooded","location":{"lat":0,"lng":0}}`
	response := httptest.NewRecorder()
	srv.smsInboundHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body)))

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}
	var payload models.SMSInboundResponse
	decodeResponse(t, response, &payload)
	if payload.Report == nil {
		t.Fatalf("expected report, got %#v", payload)
	}
	if payload.Report.Location != nil {
		t.Fatalf("expected (0,0) coordinates treated as no-location, got %#v", payload.Report.Location)
	}
}
