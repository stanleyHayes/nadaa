package config

import (
	"errors"
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/school-service/internal/utils"
)

// Config holds school-service configuration loaded from the environment.
type Config struct {
	Addr            string
	TokenSecret     string
	AllowMockActors bool
	AllowedOrigins  map[string]bool
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

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:            resolveListenAddr("", ":8097"),
		TokenSecret:     utils.EnvOrDefault("NADAA_AUTH_TOKEN_SECRET", ""),
		AllowMockActors: strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS")), "true"),
		AllowedOrigins:  utils.AllowedOriginsFromEnv(),
	}
}

// Validate fails closed on unsafe configuration: honoring self-asserted
// X-NADAA-Actor-* headers is a development-only relaxation, so it is rejected
// unless NADAA_ENV=development.
func (c *Config) Validate() error {
	if c.AllowMockActors && os.Getenv("NADAA_ENV") != "development" {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
}
