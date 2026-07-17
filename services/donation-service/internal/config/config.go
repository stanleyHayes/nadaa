package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

// Config holds donation-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	Payment        PaymentConfig

	// Env is NADAA_ENV (e.g. "development"); development-only behaviors such
	// as sandbox payment crediting and localhost CORS bypass are gated on it.
	Env string
	// AuthTokenSecret verifies nadaa.<payload>.<sig> bearer tokens issued by
	// auth-service. When empty, authority requests are rejected unless mock
	// actors are allowed.
	AuthTokenSecret string
	// AllowMockActors honors legacy X-NADAA-Actor-* headers for local dev and
	// smoke tests (NADAA_AUTH_ALLOW_MOCK_ACTORS=true); otherwise they are
	// ignored entirely.
	AllowMockActors bool
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

// resolveListenAddr honors a platform-provided PORT (e.g. Render sets a bare
// number like "10000"), normalizing it to ":PORT", then a service-specific
// address override, then the default. This lets the service bind the port the
// host expects while preserving local defaults.
func resolveListenAddr(addrKey, fallback string) string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	if addrKey != "" {
		if value := strings.TrimSpace(os.Getenv(addrKey)); value != "" {
			return value
		}
	}
	return fallback
}

// IsDevelopment reports whether the service runs in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           resolveListenAddr("", ":8100"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
		Payment: PaymentConfig{
			Provider:            strings.ToLower(utils.EnvOrDefault("NADAA_PAYMENT_PROVIDER", "sandbox")),
			PaystackSecretKey:   utils.EnvOrDefault("NADAA_PAYSTACK_SECRET_KEY", ""),
			PaystackBaseURL:     utils.EnvOrDefault("NADAA_PAYSTACK_BASE_URL", ""),
			PaystackCallbackURL: utils.EnvOrDefault("NADAA_PAYSTACK_CALLBACK_URL", ""),
		},
		Env:             strings.ToLower(utils.EnvOrDefault("NADAA_ENV", "")),
		AuthTokenSecret: os.Getenv("NADAA_AUTH_TOKEN_SECRET"),
		AllowMockActors: strings.EqualFold(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS"), "true"),
	}
}
