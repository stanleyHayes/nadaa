package models

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// This file holds the monetary-donation domain and the payment-gateway
// dependency-injection seam. A PaymentProvider abstracts a checkout backend
// (Paystack today) so the donation flow never depends on a concrete gateway and
// one provider can be swapped for another purely through configuration (see
// handlers.BuildPaymentProvider). Everything defaults to a sandbox provider so
// the platform runs end-to-end before real credentials arrive.
//
// Money-path guardrails (per provider research): a donation is only ever marked
// paid after a server-side Verify call — never from a webhook payload alone —
// and every state transition is idempotent so replayed or duplicated webhooks
// cannot double-credit.

const defaultPaymentTimeout = 12 * time.Second

const maxPaymentResponseBytes = 1 << 20

// Donation is a monetary contribution toward disaster relief.
type Donation struct {
	ID          string     `json:"id"`
	Reference   string     `json:"reference"`
	DonorName   string     `json:"donorName"`
	Email       string     `json:"email"`
	AmountMinor int64      `json:"amountMinor"`
	Currency    string     `json:"currency"`
	Channel     string     `json:"channel,omitempty"`
	Status      string     `json:"status"`
	Provider    string     `json:"provider"`
	ProviderRef string     `json:"providerRef,omitempty"`
	Campaign    string     `json:"campaign,omitempty"`
	Message     string     `json:"message,omitempty"`
	FailureCode string     `json:"failureCode,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PaidAt      *time.Time `json:"paidAt,omitempty"`
}

// CreateDonationRequest is the public payload to start a donation.
type CreateDonationRequest struct {
	DonorName string  `json:"donorName"`
	Email     string  `json:"email"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency,omitempty"`
	Campaign  string  `json:"campaign,omitempty"`
	Message   string  `json:"message,omitempty"`
}

// CreateDonationInput is the normalized donation persisted by the store.
type CreateDonationInput struct {
	DonorName   string
	Email       string
	AmountMinor int64
	Currency    string
	Campaign    string
	Message     string
	Provider    string
}

// CreateDonationResponse is returned after a donation is initialized.
type CreateDonationResponse struct {
	Donation         Donation `json:"donation"`
	AuthorizationURL string   `json:"authorizationUrl,omitempty"`
}

// DonationListResponse returns donations for reconciliation.
type DonationListResponse struct {
	Donations   []Donation `json:"donations"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

// DonationFilter captures accepted query parameters for listing donations.
type DonationFilter struct {
	Status   string
	Campaign string
}

// PaymentInitRequest asks a provider to start a payment.
type PaymentInitRequest struct {
	Reference   string
	AmountMinor int64
	Currency    string
	Email       string
	CallbackURL string
	Metadata    map[string]string
}

// PaymentInitResult is the outcome of starting a payment.
type PaymentInitResult struct {
	Provider         string
	Status           string // "initialized" | "failed" | "skipped"
	AuthorizationURL string
	ProviderRef      string
	Reason           string
}

// PaymentVerifyResult is the outcome of verifying a payment server-side.
type PaymentVerifyResult struct {
	Provider    string
	Status      string // "paid" | "pending" | "failed"
	AmountMinor int64
	Currency    string
	Channel     string
	ProviderRef string
	Reason      string
}

// PaymentProvider abstracts a payment gateway. Implementations must confirm a
// payment through Verify before any donation is credited and must never treat a
// webhook payload as proof of payment on its own.
type PaymentProvider interface {
	Name() string
	Initialize(ctx context.Context, request PaymentInitRequest) PaymentInitResult
	Verify(ctx context.Context, reference string) PaymentVerifyResult
	// VerifyWebhookSignature reports whether a raw webhook body carries a valid
	// provider signature.
	VerifyWebhookSignature(signature string, body []byte) bool
}

// SandboxPaymentProvider simulates a compliant gateway so the donation flow can
// run end-to-end without real credentials. It is the safe default.
type SandboxPaymentProvider struct{}

// Name identifies the provider in donation records and logs.
func (SandboxPaymentProvider) Name() string { return "sandbox_payment" }

// Initialize returns a simulated authorization URL for the reference.
func (SandboxPaymentProvider) Initialize(_ context.Context, request PaymentInitRequest) PaymentInitResult {
	return PaymentInitResult{
		Provider:         "sandbox_payment",
		Status:           "initialized",
		AuthorizationURL: "https://sandbox.nadaa.local/pay/" + request.Reference,
		ProviderRef:      "sandbox_" + request.Reference,
	}
}

// Verify reports the simulated payment as paid so local flows complete.
func (SandboxPaymentProvider) Verify(_ context.Context, reference string) PaymentVerifyResult {
	return PaymentVerifyResult{
		Provider:    "sandbox_payment",
		Status:      "paid",
		Channel:     "sandbox",
		ProviderRef: "sandbox_" + reference,
	}
}

// VerifyWebhookSignature accepts any signature in sandbox mode. This is safe
// only because the sandbox provider is never selected in production; a real
// provider (PaystackProvider) enforces a cryptographic signature.
func (SandboxPaymentProvider) VerifyWebhookSignature(_ string, _ []byte) bool { return true }

// DisabledPaymentProvider is the fail-safe when payments are turned off or a
// real provider is selected without credentials; it never starts or confirms a
// payment.
type DisabledPaymentProvider struct {
	Reason string
}

// Name identifies the disabled provider in logs.
func (DisabledPaymentProvider) Name() string { return "disabled_payment" }

func (p DisabledPaymentProvider) reason() string {
	if strings.TrimSpace(p.Reason) == "" {
		return "payments are not configured"
	}
	return p.Reason
}

// Initialize always skips: no payment gateway is available.
func (p DisabledPaymentProvider) Initialize(_ context.Context, _ PaymentInitRequest) PaymentInitResult {
	return PaymentInitResult{Provider: "disabled_payment", Status: "skipped", Reason: p.reason()}
}

// Verify always reports failed: no payment gateway is available.
func (p DisabledPaymentProvider) Verify(_ context.Context, _ string) PaymentVerifyResult {
	return PaymentVerifyResult{Provider: "disabled_payment", Status: "failed", Reason: p.reason()}
}

// VerifyWebhookSignature always rejects: no gateway means no trusted webhooks.
func (DisabledPaymentProvider) VerifyWebhookSignature(_ string, _ []byte) bool { return false }

// PaystackProvider integrates the Paystack REST API for Ghana mobile money
// (MTN MoMo, Telecel Cash, AirtelTigo Money) and cards. It is a thin,
// stdlib-only client — deliberately not a third-party SDK on the money path —
// and verifies webhooks with HMAC-SHA512 as Paystack specifies.
type PaystackProvider struct {
	SecretKey   string
	BaseURL     string
	CallbackURL string
	HTTPClient  *http.Client
}

// NewPaystackProvider builds a PaystackProvider with default base URL and HTTP
// client when they are not supplied.
func NewPaystackProvider(secretKey, baseURL, callbackURL string, client *http.Client) PaystackProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.paystack.co"
	}
	if client == nil {
		client = &http.Client{Timeout: defaultPaymentTimeout}
	}
	return PaystackProvider{
		SecretKey:   secretKey,
		BaseURL:     strings.TrimRight(baseURL, "/"),
		CallbackURL: strings.TrimSpace(callbackURL),
		HTTPClient:  client,
	}
}

// Name identifies the provider in donation records and logs.
func (PaystackProvider) Name() string { return "paystack" }

// Initialize starts a Paystack transaction and returns its authorization URL.
func (p PaystackProvider) Initialize(ctx context.Context, request PaymentInitRequest) PaymentInitResult {
	const providerID = "paystack"
	if strings.TrimSpace(p.SecretKey) == "" {
		return PaymentInitResult{Provider: providerID, Status: "failed", Reason: "paystack secret key not configured"}
	}

	callback := request.CallbackURL
	if callback == "" {
		callback = p.CallbackURL
	}
	payload := map[string]any{
		"email":     request.Email,
		"amount":    request.AmountMinor,
		"currency":  request.Currency,
		"reference": request.Reference,
	}
	if callback != "" {
		payload["callback_url"] = callback
	}
	if len(request.Metadata) > 0 {
		payload["metadata"] = request.Metadata
	}

	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			AuthorizationURL string `json:"authorization_url"`
			AccessCode       string `json:"access_code"`
			Reference        string `json:"reference"`
		} `json:"data"`
	}
	status, err := p.doJSON(ctx, http.MethodPost, "/transaction/initialize", payload, &parsed)
	if err != nil {
		return PaymentInitResult{Provider: providerID, Status: "failed", Reason: err.Error()}
	}
	if status >= 200 && status < 300 && parsed.Status {
		return PaymentInitResult{
			Provider:         providerID,
			Status:           "initialized",
			AuthorizationURL: parsed.Data.AuthorizationURL,
			ProviderRef:      firstNonEmpty(parsed.Data.Reference, request.Reference),
		}
	}
	return PaymentInitResult{Provider: providerID, Status: "failed", Reason: paystackReason(parsed.Message, status)}
}

// Verify confirms a transaction server-side. It is the authoritative check
// before a donation is credited.
func (p PaystackProvider) Verify(ctx context.Context, reference string) PaymentVerifyResult {
	const providerID = "paystack"
	if strings.TrimSpace(p.SecretKey) == "" {
		return PaymentVerifyResult{Provider: providerID, Status: "failed", Reason: "paystack secret key not configured"}
	}

	var parsed struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Status    string `json:"status"`
			Reference string `json:"reference"`
			Amount    int64  `json:"amount"`
			Currency  string `json:"currency"`
			Channel   string `json:"channel"`
		} `json:"data"`
	}
	status, err := p.doJSON(ctx, http.MethodGet, "/transaction/verify/"+reference, nil, &parsed)
	if err != nil {
		return PaymentVerifyResult{Provider: providerID, Status: "failed", Reason: err.Error()}
	}
	if status < 200 || status >= 300 || !parsed.Status {
		return PaymentVerifyResult{Provider: providerID, Status: "failed", Reason: paystackReason(parsed.Message, status)}
	}

	result := PaymentVerifyResult{
		Provider:    providerID,
		AmountMinor: parsed.Data.Amount,
		Currency:    parsed.Data.Currency,
		Channel:     parsed.Data.Channel,
		ProviderRef: parsed.Data.Reference,
	}
	switch strings.ToLower(parsed.Data.Status) {
	case "success":
		result.Status = "paid"
	case "failed", "abandoned", "reversed":
		result.Status = "failed"
		result.Reason = "paystack transaction " + strings.ToLower(parsed.Data.Status)
	default:
		result.Status = "pending"
	}
	return result
}

// VerifyWebhookSignature verifies the x-paystack-signature header, an
// HMAC-SHA512 of the raw request body keyed with the secret key.
func (p PaystackProvider) VerifyWebhookSignature(signature string, body []byte) bool {
	signature = strings.TrimSpace(signature)
	if signature == "" || strings.TrimSpace(p.SecretKey) == "" {
		return false
	}
	mac := hmac.New(sha512.New, []byte(p.SecretKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(signature)))
}

func (p PaystackProvider) doJSON(ctx context.Context, method, path string, payload any, out any) (int, error) {
	var reader io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("encode paystack request failed")
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.BaseURL+path, reader)
	if err != nil {
		return 0, fmt.Errorf("build paystack request failed")
	}
	req.Header.Set("Authorization", "Bearer "+p.SecretKey)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("paystack request failed")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxPaymentResponseBytes))
	if len(raw) > 0 && out != nil {
		_ = json.Unmarshal(raw, out)
	}
	return resp.StatusCode, nil
}

func paystackReason(message string, status int) string {
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		return trimmed
	}
	return fmt.Sprintf("paystack responded with status %d", status)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
