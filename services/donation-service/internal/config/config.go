package config

import (
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

// Config holds donation-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	Payment        PaymentConfig
}

// PaymentConfig selects the payment gateway and carries its credentials.
// Everything defaults to the sandbox provider so the donation flow runs
// end-to-end before real credentials arrive; the live Paystack path is only
// used once "paystack" is selected and its secret key is present (see
// handlers.BuildPaymentProvider).
type PaymentConfig struct {
	Provider string // "sandbox" | "disabled" | "paystack"

	PaystackSecretKey   string
	PaystackBaseURL     string
	PaystackCallbackURL string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           utils.EnvOrDefault("PORT", ":8100"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
		Payment: PaymentConfig{
			Provider:            strings.ToLower(utils.EnvOrDefault("NADAA_PAYMENT_PROVIDER", "sandbox")),
			PaystackSecretKey:   utils.EnvOrDefault("NADAA_PAYSTACK_SECRET_KEY", ""),
			PaystackBaseURL:     utils.EnvOrDefault("NADAA_PAYSTACK_BASE_URL", ""),
			PaystackCallbackURL: utils.EnvOrDefault("NADAA_PAYSTACK_CALLBACK_URL", ""),
		},
	}
}
