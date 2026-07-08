package config

import (
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
		TokenSecret:            envOrDefault("NADAA_AUTH_TOKEN_SECRET", "dev-secret-change-me"),
		MockOTP:                strings.TrimSpace(os.Getenv("NADAA_AUTH_MOCK_OTP")),
		ExposeDevOTP:           os.Getenv("NADAA_AUTH_EXPOSE_DEV_OTP") == "true",
		BootstrapAdminEmail:    strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL")),
		BootstrapAdminPassword: strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD")),
		BootstrapAdminPhone:    envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE", "+233200000001"),
		BootstrapAdminName:     envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_NAME", "NADAA System Admin"),
		BootstrapAdminMFACode:  strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE")),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
