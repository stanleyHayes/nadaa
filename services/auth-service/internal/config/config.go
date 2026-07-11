package config

import (
	"errors"
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

// Config holds auth-service configuration loaded from the environment.
type Config struct {
	Addr                   string
	AllowedOrigins         map[string]bool
	TokenSecret            string
	MockOTP                string
	ExposeDevOTP           bool
	// AllowMockActorHeaders lets agency endpoints accept the shared X-NADAA-*
	// actor headers instead of a verified session token. Demo/dev only — it
	// trusts client-supplied role headers, so it must stay false in production.
	AllowMockActorHeaders bool
	BootstrapAdminEmail    string
	BootstrapAdminPassword string
	BootstrapAdminPhone    string
	BootstrapAdminName     string
	BootstrapAdminMFACode  string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                   envOrDefault("NADAA_AUTH_ADDR", ":8080"),
		AllowedOrigins:         utils.AllowedOriginsFromEnv(),
		TokenSecret:            envOrDefault("NADAA_AUTH_TOKEN_SECRET", InsecureTokenSecret),
		MockOTP:                strings.TrimSpace(os.Getenv("NADAA_AUTH_MOCK_OTP")),
		ExposeDevOTP:           os.Getenv("NADAA_AUTH_EXPOSE_DEV_OTP") == "true",
		AllowMockActorHeaders:  os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true",
		BootstrapAdminEmail:    strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL")),
		BootstrapAdminPassword: strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD")),
		BootstrapAdminPhone:    envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE", "+233200000001"),
		BootstrapAdminName:     envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_NAME", "NADAA System Admin"),
		BootstrapAdminMFACode:  strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE")),
	}
}

// InsecureTokenSecret is the placeholder token-signing key. It lives in this
// repository, so any build shipping with it can have its bearer tokens forged.
const InsecureTokenSecret = "dev-secret-change-me"

// Validate fails closed on an unsafe token-signing secret. An empty, placeholder,
// or too-short secret is only allowed when NADAA_AUTH_ALLOW_INSECURE_SECRET=true
// (the same opt-in shape as mock actor headers), so production never silently
// signs tokens with a publicly-known key.
func (c *Config) Validate() error {
	if os.Getenv("NADAA_AUTH_ALLOW_INSECURE_SECRET") == "true" {
		return nil
	}
	secret := strings.TrimSpace(c.TokenSecret)
	if secret == "" || secret == InsecureTokenSecret {
		return errors.New("NADAA_AUTH_TOKEN_SECRET is required: set a strong secret, or NADAA_AUTH_ALLOW_INSECURE_SECRET=true for local development")
	}
	if len(secret) < 32 {
		return errors.New("NADAA_AUTH_TOKEN_SECRET must be at least 32 bytes; set NADAA_AUTH_ALLOW_INSECURE_SECRET=true to bypass for local development")
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
