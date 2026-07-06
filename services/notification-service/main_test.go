package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListAlertFeedIncludesCurrentAndExpiredAlerts(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/alerts?includeExpired=true", nil)

	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload citizenAlertListResponse
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

	var payload citizenAlertListResponse
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

	var payload deliveryResponse
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

	var logs deliveryLogListResponse
	decodeResponse(t, logsResponse, &logs)
	if len(logs.Logs) != 2 {
		t.Fatalf("expected persisted delivery logs, got %#v", logs.Logs)
	}
}

func TestSMSDisabledLogsSkippedAttempt(t *testing.T) {
	srv := newTestServer()
	srv.providers["sms"] = disabledProvider{channel: "sms", reason: "sms provider disabled"}
	body := `{"recipientId":"usr_demo_citizen","phone":"+233200000000","channels":["sms"]}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/alerts/alert_feed_current_flood/deliver", bytes.NewBufferString(body))
	request.SetPathValue("id", "alert_feed_current_flood")

	srv.deliverAlertHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload deliveryResponse
	decodeResponse(t, response, &payload)
	if len(payload.Attempts) != 1 {
		t.Fatalf("expected one sms attempt, got %#v", payload.Attempts)
	}
	if payload.Attempts[0].Status != "skipped" || payload.Attempts[0].Provider != "sms_disabled" {
		t.Fatalf("expected skipped sms attempt, got %#v", payload.Attempts[0])
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

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{
		store:       newMemoryStore(now),
		alertClient: nil,
		providers: map[string]notificationProvider{
			"push": mockProvider{channel: "push"},
			"sms":  mockProvider{channel: "sms"},
		},
		now: func() time.Time { return now },
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
