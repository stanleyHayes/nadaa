package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
)

func TestBuildPaymentProviderDefaultsToSandbox(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{}, true)
	if _, ok := provider.(models.SandboxPaymentProvider); !ok {
		t.Fatalf("default provider = %T, want SandboxPaymentProvider", provider)
	}
}

func TestBuildPaymentProviderSandboxCreditsOnlyInDev(t *testing.T) {
	dev := BuildPaymentProvider(config.PaymentConfig{}, true)
	if verify := dev.Verify(t.Context(), "GIFT-1"); verify.Status != "paid" {
		t.Errorf("dev sandbox verify status = %q, want paid", verify.Status)
	}
	if !dev.VerifyWebhookSignature("anything", []byte("body")) {
		t.Error("dev sandbox should accept webhook signatures")
	}

	prod := BuildPaymentProvider(config.PaymentConfig{}, false)
	if verify := prod.Verify(t.Context(), "GIFT-1"); verify.Status != "pending" {
		t.Errorf("non-dev sandbox verify status = %q, want pending (never paid)", verify.Status)
	}
	if prod.VerifyWebhookSignature("anything", []byte("body")) {
		t.Error("non-dev sandbox must never accept a webhook")
	}
}

func TestBuildPaymentProviderSelectsPaystackWithKey(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{Provider: "paystack", PaystackSecretKey: "sk_test"}, false)
	if _, ok := provider.(models.PaystackProvider); !ok {
		t.Fatalf("provider = %T, want PaystackProvider", provider)
	}
	if provider.Name() != "paystack" {
		t.Errorf("name = %q, want paystack", provider.Name())
	}
}

func TestBuildPaymentProviderDisablesPaystackWithoutKey(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{Provider: "paystack"}, false)
	if _, ok := provider.(models.DisabledPaymentProvider); !ok {
		t.Fatalf("provider = %T, want DisabledPaymentProvider when key is missing", provider)
	}
}

func TestBuildPaymentProviderUnknownFailsSafe(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{Provider: "stripe"}, false)
	if _, ok := provider.(models.DisabledPaymentProvider); !ok {
		t.Fatalf("provider = %T, want DisabledPaymentProvider for an unknown selection", provider)
	}
}

func TestCreateDonationFlow(t *testing.T) {
	srv := newTestServer() // sandbox payment provider

	// Start a donation.
	body, _ := json.Marshal(models.CreateDonationRequest{
		DonorName: "Ama",
		Email:     "ama@example.com",
		Amount:    50,
		Campaign:  "flood_relief",
	})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/donations", bytes.NewReader(body))
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201: %s", response.Code, response.Body.String())
	}
	var created models.CreateDonationResponse
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.Donation.Status != "pending" {
		t.Errorf("status = %q, want pending", created.Donation.Status)
	}
	if created.Donation.AmountMinor != 5000 {
		t.Errorf("amountMinor = %d, want 5000", created.Donation.AmountMinor)
	}
	if created.AuthorizationURL == "" {
		t.Error("expected an authorization URL")
	}

	// Verify via GET — sandbox (dev mode) reports the payment as paid.
	getResponse := httptest.NewRecorder()
	getRequest := authorityRequest(http.MethodGet, "/api/v1/donations/"+created.Donation.Reference, nil)
	srv.Routes().ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", getResponse.Code)
	}
	var fetched models.Donation
	if err := json.Unmarshal(getResponse.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if fetched.Status != "paid" {
		t.Errorf("status = %q, want paid after verification", fetched.Status)
	}
	if fetched.PaidAt == nil {
		t.Error("paidAt should be set once paid")
	}
}

func TestCreateDonationRejectsBadEmail(t *testing.T) {
	srv := newTestServer()
	body, _ := json.Marshal(models.CreateDonationRequest{DonorName: "Ama", Email: "not-an-email", Amount: 50})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/donations", bytes.NewReader(body))
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for an invalid email", response.Code)
	}
}

func TestCreateDonationRejectsTinyAmount(t *testing.T) {
	srv := newTestServer()
	body, _ := json.Marshal(models.CreateDonationRequest{Email: "ama@example.com", Amount: 0.5})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/donations", bytes.NewReader(body))
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for an amount below the minimum", response.Code)
	}
}

func TestWebhookRejectsBadSignatureWithPaystack(t *testing.T) {
	// A Paystack-backed server must reject an unsigned webhook.
	srv := newTestServer()
	srv.payments = models.NewPaystackProvider("sk_test_secret", "https://api.paystack.co", "", nil)

	body := []byte(`{"event":"charge.success","data":{"reference":"GIFT-1"}}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/paystack", bytes.NewReader(body))
	request.Header.Set("x-paystack-signature", "deadbeef")
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 for an invalid signature", response.Code)
	}
}

type stubVerifyProvider struct {
	result models.PaymentVerifyResult
}

func (p stubVerifyProvider) Name() string { return "stub_payment" }
func (p stubVerifyProvider) Initialize(_ context.Context, _ models.PaymentInitRequest) models.PaymentInitResult {
	return models.PaymentInitResult{Provider: "stub_payment", Status: "initialized", AuthorizationURL: "https://stub.local/pay", ProviderRef: "stub_ref"}
}
func (p stubVerifyProvider) Verify(_ context.Context, _ string) models.PaymentVerifyResult {
	return p.result
}
func (p stubVerifyProvider) VerifyWebhookSignature(_ string, _ []byte) bool { return false }

func TestReconcileDonationFailsOnCurrencyMismatch(t *testing.T) {
	srv := newTestServer()
	srv.payments = stubVerifyProvider{result: models.PaymentVerifyResult{
		Provider:    "stub_payment",
		Status:      "paid",
		AmountMinor: 5000,
		Currency:    "USD", // donation is recorded in GHS
	}}

	body, _ := json.Marshal(models.CreateDonationRequest{DonorName: "Ama", Email: "ama@example.com", Amount: 50})
	createResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/donations", bytes.NewReader(body)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201: %s", createResponse.Code, createResponse.Body.String())
	}
	var created models.CreateDonationResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	getResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getResponse, authorityRequest(http.MethodGet, "/api/v1/donations/"+created.Donation.Reference, nil))
	if getResponse.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200: %s", getResponse.Code, getResponse.Body.String())
	}
	var fetched models.Donation
	if err := json.Unmarshal(getResponse.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if fetched.Status != "failed" || fetched.FailureCode != "currency_mismatch" {
		t.Fatalf("expected failed/currency_mismatch, got status=%q failureCode=%q", fetched.Status, fetched.FailureCode)
	}
}
