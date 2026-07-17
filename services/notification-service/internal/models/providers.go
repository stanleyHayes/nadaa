package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// This file holds the real (network-backed) NotificationProvider implementations
// that sit behind the same models.NotificationProvider seam as MockProvider and
// DisabledProvider. Each provider is a plain struct with its credentials and an
// injected *http.Client, so the delivery path never depends on a concrete
// backend and one provider can be swapped for another purely through
// configuration (see handlers.BuildProviders). Providers are intentionally
// fail-safe: a dry run never touches the network, a missing recipient is
// skipped rather than errored, and a missing credential fails loudly instead of
// pretending to deliver.

// defaultProviderTimeout bounds any single outbound provider request.
const defaultProviderTimeout = 10 * time.Second

// maxProviderResponseBytes caps how much of a provider response body we read so
// a hostile or misbehaving upstream cannot exhaust memory.
const maxProviderResponseBytes = 1 << 20

// ArkeselSMSProvider delivers SMS through the Arkesel SMS API v2
// (https://developers.arkesel.com). Arkesel is the confirmed SMS backend for
// Ghana; it authenticates with an `api-key` header and accepts a JSON body of
// a sender id, a message, and a list of recipient MSISDNs.
type ArkeselSMSProvider struct {
	APIKey     string
	Sender     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewArkeselSMSProvider builds an ArkeselSMSProvider with sane defaults for the
// base URL, sender id, and HTTP client when they are not supplied.
func NewArkeselSMSProvider(apiKey, sender, baseURL string, client *http.Client) ArkeselSMSProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://sms.arkesel.com"
	}
	if strings.TrimSpace(sender) == "" {
		sender = "NADAA"
	}
	if client == nil {
		client = &http.Client{Timeout: defaultProviderTimeout}
	}
	return ArkeselSMSProvider{
		APIKey:     apiKey,
		Sender:     sender,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: client,
	}
}

// Send delivers the alert as an SMS to the request's phone number.
func (p ArkeselSMSProvider) Send(ctx context.Context, message ProviderMessage) ProviderResult {
	const providerID = "arkesel_sms"

	if message.Request.DryRun {
		return ProviderResult{Provider: providerID, Status: "simulated", Reason: "dry run; no live sms sent"}
	}
	phone := strings.TrimSpace(message.Request.Phone)
	if phone == "" {
		return ProviderResult{Provider: providerID, Status: "skipped", Reason: "no phone number for sms recipient"}
	}
	if strings.TrimSpace(p.APIKey) == "" {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "arkesel api key not configured"}
	}

	body, err := json.Marshal(map[string]any{
		"sender":     p.Sender,
		"message":    smsMessageText(message.Alert),
		"recipients": []string{phone},
	})
	if err != nil {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "encode arkesel request failed"}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/api/v2/sms/send", bytes.NewReader(body))
	if err != nil {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "build arkesel request failed"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-key", p.APIKey)

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "arkesel request failed"}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxProviderResponseBytes))

	var parsed struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    []struct {
			Recipient string `json:"recipient"`
			ID        string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(raw, &parsed)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && strings.EqualFold(parsed.Status, "success") {
		messageID := ""
		if len(parsed.Data) > 0 {
			messageID = parsed.Data[0].ID
		}
		return ProviderResult{Provider: providerID, Status: "delivered", MessageID: messageID}
	}

	reason := strings.TrimSpace(parsed.Message)
	if reason == "" {
		reason = fmt.Sprintf("arkesel responded with status %d", resp.StatusCode)
	}
	return ProviderResult{Provider: providerID, Status: "failed", Reason: reason}
}

// ExpoPushProvider delivers push notifications through the Expo push service
// (https://exp.host/--/api/v2/push/send). Expo is the value-for-money default
// for the NADAA mobile apps because they are Expo apps: it is free, delivers to
// both APNs and FCM behind one token, and honors the Android notification
// channels and iOS critical-alert configuration the apps already declare. An
// access token is optional (used for enhanced push security) and, when present,
// is sent as a bearer token.
type ExpoPushProvider struct {
	AccessToken string
	BaseURL     string
	HTTPClient  *http.Client
}

// NewExpoPushProvider builds an ExpoPushProvider with default base URL and HTTP
// client when they are not supplied.
func NewExpoPushProvider(accessToken, baseURL string, client *http.Client) ExpoPushProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://exp.host"
	}
	if client == nil {
		client = &http.Client{Timeout: defaultProviderTimeout}
	}
	return ExpoPushProvider{
		AccessToken: accessToken,
		BaseURL:     strings.TrimRight(baseURL, "/"),
		HTTPClient:  client,
	}
}

// Send delivers the alert as a push notification to the request's push token.
func (p ExpoPushProvider) Send(ctx context.Context, message ProviderMessage) ProviderResult {
	const providerID = "expo_push"

	if message.Request.DryRun {
		return ProviderResult{Provider: providerID, Status: "simulated", Reason: "dry run; no live push sent"}
	}
	token := strings.TrimSpace(message.Request.PushToken)
	if token == "" {
		return ProviderResult{Provider: providerID, Status: "skipped", Reason: "no push token for recipient"}
	}

	body, err := json.Marshal(map[string]any{
		"to":       token,
		"title":    pushTitle(message.Alert),
		"body":     pushBody(message.Alert),
		"priority": "high",
		"sound":    "default",
	})
	if err != nil {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "encode expo request failed"}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/--/api/v2/push/send", bytes.NewReader(body))
	if err != nil {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "build expo request failed"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if strings.TrimSpace(p.AccessToken) != "" {
		req.Header.Set("Authorization", "Bearer "+p.AccessToken)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return ProviderResult{Provider: providerID, Status: "failed", Reason: "expo request failed"}
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxProviderResponseBytes))

	var parsed struct {
		Data struct {
			Status  string `json:"status"`
			ID      string `json:"id"`
			Message string `json:"message"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	_ = json.Unmarshal(raw, &parsed)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && strings.EqualFold(parsed.Data.Status, "ok") {
		// An "ok" ticket only means Expo accepted the push; delivery receipts
		// are never fetched, so record the attempt as sent, not delivered.
		return ProviderResult{Provider: providerID, Status: "sent", MessageID: parsed.Data.ID}
	}

	reason := strings.TrimSpace(parsed.Data.Message)
	if reason == "" && len(parsed.Errors) > 0 {
		reason = strings.TrimSpace(parsed.Errors[0].Message)
	}
	if reason == "" {
		reason = fmt.Sprintf("expo responded with status %d", resp.StatusCode)
	}
	return ProviderResult{Provider: providerID, Status: "failed", Reason: reason}
}

// smsMessageText renders a concise, branded SMS body from an alert.
func smsMessageText(alert CitizenAlert) string {
	title := strings.TrimSpace(alert.Title)
	body := strings.TrimSpace(alert.Message)
	switch {
	case title != "" && body != "":
		return fmt.Sprintf("NADAA %s: %s", title, body)
	case title != "":
		return "NADAA: " + title
	case body != "":
		return "NADAA: " + body
	default:
		return "NADAA emergency alert"
	}
}

// pushTitle renders the push notification title from an alert.
func pushTitle(alert CitizenAlert) string {
	if title := strings.TrimSpace(alert.Title); title != "" {
		return title
	}
	return "NADAA alert"
}

// pushBody renders the push notification body from an alert.
func pushBody(alert CitizenAlert) string {
	if body := strings.TrimSpace(alert.Message); body != "" {
		return body
	}
	return strings.TrimSpace(alert.RecommendedAction)
}
