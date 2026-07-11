package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
)

func TestBuildPaymentProviderDefaultsToSandbox(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{})
	if _, ok := provider.(models.SandboxPaymentProvider); !ok {
		t.Fatalf("default provider = %T, want SandboxPaymentProvider", provider)
	}
}

func TestBuildPaymentProviderSelectsPaystackWithKey(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{Provider: "paystack", PaystackSecretKey: "sk_test"})
	if _, ok := provider.(models.PaystackProvider); !ok {
		t.Fatalf("provider = %T, want PaystackProvider", provider)
	}
	if provider.Name() != "paystack" {
		t.Errorf("name = %q, want paystack", provider.Name())
	}
}

func TestBuildPaymentProviderDisablesPaystackWithoutKey(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{Provider: "paystack"})
	if _, ok := provider.(models.DisabledPaymentProvider); !ok {
		t.Fatalf("provider = %T, want DisabledPaymentProvider when key is missing", provider)
	}
}

func TestBuildPaymentProviderUnknownFailsSafe(t *testing.T) {
	provider := BuildPaymentProvider(config.PaymentConfig{Provider: "stripe"})
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

	// Verify via GET — sandbox reports the payment as paid.
	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/donations/"+created.Donation.Reference, nil)
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
