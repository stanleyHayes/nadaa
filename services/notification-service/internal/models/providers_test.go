package models

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testMessage(dryRun bool, phone, pushToken string) ProviderMessage {
	return ProviderMessage{
		Alert: CitizenAlert{
			ID:      "alert_1",
			Title:   "Flood warning",
			Message: "Move to higher ground now.",
		},
		Request: DeliveryRequest{
			RecipientID: "citizen_1",
			Phone:       phone,
			PushToken:   pushToken,
			DryRun:      dryRun,
		},
		Channel:     "sms",
		AttemptedAt: time.Unix(1700000000, 0).UTC(),
	}
}

func TestArkeselSMSProviderDelivers(t *testing.T) {
	var gotAPIKey, gotBody string
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("api-key")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"success","data":[{"recipient":"233200000000","id":"msg_123"}]}`)
	}))
	defer server.Close()

	provider := NewArkeselSMSProvider("secret-key", "NADAA", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(false, "233200000000", ""))

	if result.Status != "delivered" {
		t.Fatalf("status = %q, want delivered (reason %q)", result.Status, result.Reason)
	}
	if result.Provider != "arkesel_sms" {
		t.Errorf("provider = %q, want arkesel_sms", result.Provider)
	}
	if result.MessageID != "msg_123" {
		t.Errorf("messageID = %q, want msg_123", result.MessageID)
	}
	if gotAPIKey != "secret-key" {
		t.Errorf("api-key header = %q, want secret-key", gotAPIKey)
	}
	if gotPath != "/api/v2/sms/send" {
		t.Errorf("path = %q, want /api/v2/sms/send", gotPath)
	}

	var payload struct {
		Sender     string   `json:"sender"`
		Message    string   `json:"message"`
		Recipients []string `json:"recipients"`
	}
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("request body not valid json: %v", err)
	}
	if payload.Sender != "NADAA" {
		t.Errorf("sender = %q, want NADAA", payload.Sender)
	}
	if len(payload.Recipients) != 1 || payload.Recipients[0] != "233200000000" {
		t.Errorf("recipients = %v, want [233200000000]", payload.Recipients)
	}
	if !strings.Contains(payload.Message, "Flood warning") {
		t.Errorf("message = %q, want it to contain the alert title", payload.Message)
	}
}

func TestArkeselSMSProviderReportsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"status":"error","message":"invalid sender id"}`)
	}))
	defer server.Close()

	provider := NewArkeselSMSProvider("secret-key", "NADAA", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(false, "233200000000", ""))

	if result.Status != "failed" {
		t.Fatalf("status = %q, want failed", result.Status)
	}
	if result.Reason != "invalid sender id" {
		t.Errorf("reason = %q, want the upstream error message", result.Reason)
	}
}

func TestArkeselSMSProviderDryRunSkipsNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("dry run must not call the network")
	}))
	defer server.Close()

	provider := NewArkeselSMSProvider("secret-key", "NADAA", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(true, "233200000000", ""))

	if result.Status != "simulated" {
		t.Fatalf("status = %q, want simulated", result.Status)
	}
}

func TestArkeselSMSProviderSkipsMissingPhone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("must not call the network without a phone number")
	}))
	defer server.Close()

	provider := NewArkeselSMSProvider("secret-key", "NADAA", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(false, "", ""))

	if result.Status != "skipped" {
		t.Fatalf("status = %q, want skipped", result.Status)
	}
}

func TestExpoPushProviderDelivers(t *testing.T) {
	var gotAuth, gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"status":"ok","id":"receipt_1"}}`)
	}))
	defer server.Close()

	provider := NewExpoPushProvider("push-token-secret", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(false, "", "ExponentPushToken[abc]"))

	if result.Status != "delivered" {
		t.Fatalf("status = %q, want delivered (reason %q)", result.Status, result.Reason)
	}
	if result.Provider != "expo_push" {
		t.Errorf("provider = %q, want expo_push", result.Provider)
	}
	if result.MessageID != "receipt_1" {
		t.Errorf("messageID = %q, want receipt_1", result.MessageID)
	}
	if gotAuth != "Bearer push-token-secret" {
		t.Errorf("authorization = %q, want Bearer push-token-secret", gotAuth)
	}
	if gotPath != "/--/api/v2/push/send" {
		t.Errorf("path = %q, want /--/api/v2/push/send", gotPath)
	}
	if !strings.Contains(gotBody, "ExponentPushToken[abc]") {
		t.Errorf("body = %q, want it to target the push token", gotBody)
	}
}

func TestExpoPushProviderReportsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"status":"error","message":"DeviceNotRegistered"}}`)
	}))
	defer server.Close()

	provider := NewExpoPushProvider("", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(false, "", "ExponentPushToken[abc]"))

	if result.Status != "failed" {
		t.Fatalf("status = %q, want failed", result.Status)
	}
	if result.Reason != "DeviceNotRegistered" {
		t.Errorf("reason = %q, want DeviceNotRegistered", result.Reason)
	}
}

func TestExpoPushProviderSkipsMissingToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("must not call the network without a push token")
	}))
	defer server.Close()

	provider := NewExpoPushProvider("", server.URL, server.Client())
	result := provider.Send(context.Background(), testMessage(false, "", ""))

	if result.Status != "skipped" {
		t.Fatalf("status = %q, want skipped", result.Status)
	}
}

func TestExpoPushProviderOmitsAuthWhenNoToken(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"status":"ok","id":"receipt_2"}}`)
	}))
	defer server.Close()

	provider := NewExpoPushProvider("", server.URL, server.Client())
	_ = provider.Send(context.Background(), testMessage(false, "", "ExponentPushToken[abc]"))

	if gotAuth != "" {
		t.Errorf("authorization = %q, want empty when no access token", gotAuth)
	}
}
