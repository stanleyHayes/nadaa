package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
)

const paymentHTTPTimeout = 12 * time.Second

// BuildPaymentProvider selects the payment gateway from configuration. It is
// the dependency-injection seam for donations: the live Paystack path is chosen
// only when it is explicitly selected and its secret key is present. It defaults
// to the sandbox provider so the flow runs end-to-end before real credentials
// arrive, and it fails safe — selecting "paystack" without a key yields a
// disabled provider with a clear reason rather than a broken live path.
func BuildPaymentProvider(cfg config.PaymentConfig) models.PaymentProvider {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "paystack":
		if strings.TrimSpace(cfg.PaystackSecretKey) == "" {
			return models.DisabledPaymentProvider{Reason: "payment provider 'paystack' selected but NADAA_PAYSTACK_SECRET_KEY is not set"}
		}
		return models.NewPaystackProvider(
			cfg.PaystackSecretKey,
			cfg.PaystackBaseURL,
			cfg.PaystackCallbackURL,
			&http.Client{Timeout: paymentHTTPTimeout},
		)
	case "", "sandbox", "mock":
		return models.SandboxPaymentProvider{}
	case "disabled", "off", "none":
		return models.DisabledPaymentProvider{Reason: "payments disabled"}
	default:
		return models.DisabledPaymentProvider{Reason: "unknown payment provider '" + strings.ToLower(strings.TrimSpace(cfg.Provider)) + "'"}
	}
}
