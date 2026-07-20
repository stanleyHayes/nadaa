package config

import (
	"errors"
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

// Config holds road-closure-service configuration loaded from the environment.
type Config struct {
	Addr                  string
	AllowedOrigins        map[string]bool
	AuthTokenSecret       string
	AllowMockActorHeaders bool
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
		Addr:           resolveListenAddr("NADAA_ROAD_CLOSURE_ADDR", ":8095"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
		// Shared HMAC key verifying nadaa.<payload>.<sig> tokens from
		// auth-service. Empty means bearer verification is disabled and
		// authority requests fail unless mock actor headers are allowed.
		AuthTokenSecret: strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		// Local dev / smoke-test escape hatch: honor legacy X-NADAA-Actor-*
		// headers instead of a verified bearer token.
		AllowMockActorHeaders: os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true",
	}
}

// Validate fails closed on unsafe configuration: honoring self-asserted
// X-NADAA-Actor-* headers is a development-only relaxation, so it is rejected
// unless NADAA_ENV=development.
func (c *Config) Validate() error {
	if c.AllowMockActorHeaders && os.Getenv("NADAA_ENV") != "development" {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
}
