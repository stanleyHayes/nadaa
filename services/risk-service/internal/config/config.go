package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds risk-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	MLAPIURL       string
	// InternalServiceToken authenticates service-to-service calls (empty disables).
	InternalServiceToken string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                 resolveListenAddr("NADAA_RISK_ADDR", ":8081"),
		AllowedOrigins:       allowedOriginsFromEnv(),
		MLAPIURL:             strings.TrimSpace(os.Getenv("NADAA_ML_API_URL")),
		InternalServiceToken: strings.TrimSpace(os.Getenv("NADAA_INTERNAL_SERVICE_TOKEN")),
	}
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

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for origin := range strings.SplitSeq(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

// Validate fails closed on unsafe configuration. risk-service does not honor
// mock actor headers at all, but the development escape hatch must still
// never leak into a deployed environment.
func (c *Config) Validate() error {
	if os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true" && os.Getenv("NADAA_ENV") != "development" {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
}
