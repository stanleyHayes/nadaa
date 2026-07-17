package models

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaystackInitialize(t *testing.T) {
	var gotAuth, gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"data":{"authorization_url":"https://checkout.paystack.com/abc","access_code":"ac_1","reference":"GIFT-1"}}`)
	}))
	defer server.Close()

	provider := NewPaystackProvider("sk_test_secret", server.URL, "https://nadaa.gov.gh/thanks", server.Client())
	result := provider.Initialize(context.Background(), PaymentInitRequest{
		Reference:   "GIFT-1",
		AmountMinor: 5000,
		Currency:    "GHS",
		Email:       "donor@example.com",
	})

	if result.Status != "initialized" {
		t.Fatalf("status = %q, want initialized (reason %q)", result.Status, result.Reason)
	}
	if result.AuthorizationURL != "https://checkout.paystack.com/abc" {
		t.Errorf("authorizationUrl = %q", result.AuthorizationURL)
	}
	if result.ProviderRef != "GIFT-1" {
		t.Errorf("providerRef = %q, want GIFT-1", result.ProviderRef)
	}
	if gotAuth != "Bearer sk_test_secret" {
		t.Errorf("authorization = %q, want Bearer sk_test_secret", gotAuth)
	}
	if gotPath != "/transaction/initialize" {
		t.Errorf("path = %q", gotPath)
	}

	var payload struct {
		Email       string `json:"email"`
		Amount      int64  `json:"amount"`
		Currency    string `json:"currency"`
		Reference   string `json:"reference"`
		CallbackURL string `json:"callback_url"`
	}
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("request body not valid json: %v", err)
	}
	if payload.Amount != 5000 || payload.Currency != "GHS" || payload.Email != "donor@example.com" || payload.Reference != "GIFT-1" {
		t.Errorf("unexpected payload: %+v", payload)
	}
	if payload.CallbackURL != "https://nadaa.gov.gh/thanks" {
		t.Errorf("callback_url = %q, want the configured callback", payload.CallbackURL)
	}
}

func TestPaystackVerifyPaid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transaction/verify/GIFT-1" {
			t.Errorf("path = %q, want /transaction/verify/GIFT-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"data":{"status":"success","reference":"GIFT-1","amount":5000,"currency":"GHS","channel":"mobile_money"}}`)
	}))
	defer server.Close()

	provider := NewPaystackProvider("sk_test_secret", server.URL, "", server.Client())
	result := provider.Verify(context.Background(), "GIFT-1")

	if result.Status != "paid" {
		t.Fatalf("status = %q, want paid (reason %q)", result.Status, result.Reason)
	}
	if result.AmountMinor != 5000 {
		t.Errorf("amountMinor = %d, want 5000", result.AmountMinor)
	}
	if result.Channel != "mobile_money" {
		t.Errorf("channel = %q, want mobile_money", result.Channel)
	}
}

func TestPaystackVerifyFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":true,"data":{"status":"abandoned","reference":"GIFT-2","amount":5000,"currency":"GHS"}}`)
	}))
	defer server.Close()

	provider := NewPaystackProvider("sk_test_secret", server.URL, "", server.Client())
	result := provider.Verify(context.Background(), "GIFT-2")

	if result.Status != "failed" {
		t.Fatalf("status = %q, want failed", result.Status)
	}
}

func TestPaystackVerifyPendingOnServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"status":false,"message":"gateway error"}`)
	}))
	defer server.Close()

	provider := NewPaystackProvider("sk_test_secret", server.URL, "", server.Client())
	result := provider.Verify(context.Background(), "GIFT-3")

	if result.Status != "pending" {
		t.Fatalf("status = %q, want pending on a 5xx (transient errors must not fail a donation)", result.Status)
	}
}

func TestPaystackVerifyPendingOnTransportError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // close immediately so every request fails to connect

	provider := NewPaystackProvider("sk_test_secret", server.URL, "", server.Client())
	result := provider.Verify(context.Background(), "GIFT-4")

	if result.Status != "pending" {
		t.Fatalf("status = %q, want pending on a transport error", result.Status)
	}
}

func TestPaystackVerifyWebhookSignature(t *testing.T) {
	const secret = "sk_test_secret"
	provider := NewPaystackProvider(secret, "https://api.paystack.co", "", nil)
	body := []byte(`{"event":"charge.success","data":{"reference":"GIFT-1"}}`)

	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	valid := hex.EncodeToString(mac.Sum(nil))

	if !provider.VerifyWebhookSignature(valid, body) {
		t.Error("valid signature rejected")
	}
	if provider.VerifyWebhookSignature("deadbeef", body) {
		t.Error("invalid signature accepted")
	}
	if provider.VerifyWebhookSignature("", body) {
		t.Error("empty signature accepted")
	}
}

func TestPaystackInitializeWithoutKeyFails(t *testing.T) {
	provider := NewPaystackProvider("", "https://api.paystack.co", "", nil)
	result := provider.Initialize(context.Background(), PaymentInitRequest{Reference: "GIFT-1", AmountMinor: 5000, Currency: "GHS", Email: "d@example.com"})
	if result.Status != "failed" {
		t.Fatalf("status = %q, want failed when secret key is missing", result.Status)
	}
}

func TestSandboxPaymentProvider(t *testing.T) {
	provider := SandboxPaymentProvider{CreditPayments: true}
	init := provider.Initialize(context.Background(), PaymentInitRequest{Reference: "GIFT-9"})
	if init.Status != "initialized" || init.AuthorizationURL == "" {
		t.Fatalf("sandbox initialize = %+v", init)
	}
	verify := provider.Verify(context.Background(), "GIFT-9")
	if verify.Status != "paid" {
		t.Fatalf("sandbox verify status = %q, want paid", verify.Status)
	}
	if !provider.VerifyWebhookSignature("anything", []byte("body")) {
		t.Error("dev sandbox should accept webhook signatures")
	}
}

func TestSandboxPaymentProviderNeverCreditsOutsideDev(t *testing.T) {
	provider := SandboxPaymentProvider{} // CreditPayments defaults to false
	verify := provider.Verify(context.Background(), "GIFT-9")
	if verify.Status != "pending" {
		t.Fatalf("non-dev sandbox verify status = %q, want pending (never paid)", verify.Status)
	}
	if provider.VerifyWebhookSignature("anything", []byte("body")) {
		t.Error("non-dev sandbox must never accept a webhook")
	}
}

func TestDisabledPaymentProvider(t *testing.T) {
	provider := DisabledPaymentProvider{}
	if init := provider.Initialize(context.Background(), PaymentInitRequest{}); init.Status != "skipped" {
		t.Errorf("disabled initialize status = %q, want skipped", init.Status)
	}
	if verify := provider.Verify(context.Background(), "x"); verify.Status != "failed" {
		t.Errorf("disabled verify status = %q, want failed", verify.Status)
	}
	if provider.VerifyWebhookSignature("anything", []byte("body")) {
		t.Error("disabled provider must never accept a webhook")
	}
}
