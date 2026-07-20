package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/store"
)

var tokenTestNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

func newTokenTestServer(secret string) *Server {
	cfg := &config.Config{Addr: ":8090", TokenSecret: secret, Env: "development"}
	return NewServer(
		store.NewMemoryStore(tokenTestNow),
		nil,
		nil,
		map[string]models.NotificationProvider{
			"push":  models.MockProvider{Channel: "push"},
			"sms":   models.MockProvider{Channel: "sms"},
			"voice": models.MockProvider{Channel: "voice"},
		},
		models.SandboxCellBroadcastAdapter{},
		func() time.Time { return tokenTestNow },
		cfg,
	)
}

func signTestToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	return "nadaa." + encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func authorityClaims(role string, mfa bool) map[string]any {
	return map[string]any{
		"sub":      "usr_dispatcher_1",
		"typ":      "agency",
		"role":     role,
		"agencyId": "agency_test",
		"district": "accra-metropolitan",
		"mfa":      mfa,
		"exp":      tokenTestNow.Add(time.Hour).Unix(),
	}
}

func newDeliverRequest(body string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_current_flood/deliver", bytes.NewBufferString(body))
	request.SetPathValue("id", "alert_feed_current_flood")
	return request
}

func TestDeliverAlertRequiresAuthorityWithoutToken(t *testing.T) {
	srv := newTokenTestServer("test-secret")
	body := `{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`

	// No token at all → 401.
	response := httptest.NewRecorder()
	srv.deliverAlertHandler(response, newDeliverRequest(body))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "missing_authority_context") {
		t.Fatalf("expected missing_authority_context error, got %s", response.Body.String())
	}

	// Legacy mock-actor headers are ignored when mock actors are disabled.
	headerRequest := newDeliverRequest(body)
	withAuthority(headerRequest)
	headerResponse := httptest.NewRecorder()
	srv.deliverAlertHandler(headerResponse, headerRequest)
	if headerResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected mock-actor headers to be ignored with status %d, got %d", http.StatusUnauthorized, headerResponse.Code)
	}
}

func TestDeliverAlertWithValidBearerToken(t *testing.T) {
	srv := newTokenTestServer("test-secret")
	request := newDeliverRequest(`{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, "test-secret", authorityClaims("dispatcher", true)))

	response := httptest.NewRecorder()
	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}
	var payload models.DeliveryResponse
	decodeResponse(t, response, &payload)
	if len(payload.Attempts) != 1 || payload.Attempts[0].Status != "delivered" {
		t.Fatalf("expected one delivered mock sms attempt, got %#v", payload.Attempts)
	}
}

func TestDeliverAlertRejectsTamperedTokenSignature(t *testing.T) {
	srv := newTokenTestServer("test-secret")
	request := newDeliverRequest(`{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, "other-secret", authorityClaims("dispatcher", true)))

	response := httptest.NewRecorder()
	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestDeliverAlertRequiresCompletedMFA(t *testing.T) {
	srv := newTokenTestServer("test-secret")
	request := newDeliverRequest(`{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, "test-secret", authorityClaims("dispatcher", false)))

	response := httptest.NewRecorder()
	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "mfa_required") {
		t.Fatalf("expected mfa_required error, got %s", response.Body.String())
	}
}

func TestDeliverAlertRejectsDisallowedRole(t *testing.T) {
	srv := newTokenTestServer("test-secret")
	request := newDeliverRequest(`{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, "test-secret", authorityClaims("agency_viewer", true)))

	response := httptest.NewRecorder()
	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "forbidden") {
		t.Fatalf("expected forbidden error, got %s", response.Body.String())
	}
}

func TestDeliverExpiredAlertConflict(t *testing.T) {
	srv := newTestServer()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_expired_road/deliver", bytes.NewBufferString(`{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`))
	request.SetPathValue("id", "alert_feed_expired_road")
	withAuthority(request)

	response := httptest.NewRecorder()
	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "alert_not_deliverable") {
		t.Fatalf("expected alert_not_deliverable error, got %s", response.Body.String())
	}
}

func TestVoiceDeliveryRevalidatesUnderlyingAlert(t *testing.T) {
	currentNow := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8090", AllowMockActors: true, Env: "development"}
	srv := NewServer(
		store.NewMemoryStore(currentNow),
		nil,
		nil,
		map[string]models.NotificationProvider{"voice": models.MockProvider{Channel: "voice"}},
		models.SandboxCellBroadcastAdapter{},
		func() time.Time { return currentNow },
		cfg,
	)

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts", bytes.NewBufferString(`{"alertId":"alert_feed_current_flood","languages":["en"]}`))
	srv.createVoiceAlertHandler(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.VoiceAlertResponse
	decodeResponse(t, createResponse, &created)

	reviewResponse := httptest.NewRecorder()
	reviewRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts/"+created.Asset.ID+"/review", bytes.NewBufferString(`{"action":"approve","reviewer":"nadmo_voice_reviewer"}`))
	reviewRequest.SetPathValue("id", created.Asset.ID)
	srv.reviewVoiceAlertHandler(reviewResponse, reviewRequest)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected review status %d, got %d: %s", http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
	}

	// The underlying alert expires before the delivery is attempted.
	currentNow = currentNow.Add(8 * 24 * time.Hour)

	deliverResponse := httptest.NewRecorder()
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/voice-alerts/"+created.Asset.ID+"/deliver", bytes.NewBufferString(`{"recipients":[{"phone":"+233200000010","language":"en"}]}`))
	deliverRequest.SetPathValue("id", created.Asset.ID)
	withAuthority(deliverRequest)
	srv.deliverVoiceAlertHandler(deliverResponse, deliverRequest)

	if deliverResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d: %s", http.StatusConflict, deliverResponse.Code, deliverResponse.Body.String())
	}
	if !strings.Contains(deliverResponse.Body.String(), "alert_not_deliverable") {
		t.Fatalf("expected alert_not_deliverable error, got %s", deliverResponse.Body.String())
	}
}

func TestWebhookSecretRequiredWhenConfigured(t *testing.T) {
	srv := newTestServer()
	srv.config.WebhookSecrets.SMS = "webhook-secret"
	body := `{"from":"+233200000020","body":"ALERTS"}`

	// Missing header → 401.
	missingResponse := httptest.NewRecorder()
	srv.smsInboundHandler(missingResponse, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body)))
	if missingResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, missingResponse.Code)
	}
	if !strings.Contains(missingResponse.Body.String(), "invalid_webhook_secret") {
		t.Fatalf("expected invalid_webhook_secret error, got %s", missingResponse.Body.String())
	}

	// Wrong header → 401.
	wrongRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body))
	wrongRequest.Header.Set("X-NADAA-Webhook-Secret", "wrong-secret")
	wrongResponse := httptest.NewRecorder()
	srv.smsInboundHandler(wrongResponse, wrongRequest)
	if wrongResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, wrongResponse.Code)
	}

	// Matching header → accepted.
	okRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body))
	okRequest.Header.Set("X-NADAA-Webhook-Secret", "webhook-secret")
	okResponse := httptest.NewRecorder()
	srv.smsInboundHandler(okResponse, okRequest)
	if okResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, okResponse.Code, okResponse.Body.String())
	}
}

func TestWebhookWithoutSecretConfiguredStaysOpen(t *testing.T) {
	srv := newTestServer()
	body := `{"from":"+233200000021","body":"ALERTS"}`
	response := httptest.NewRecorder()
	srv.smsInboundHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/sms/inbound", bytes.NewBufferString(body)))
	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}
}
