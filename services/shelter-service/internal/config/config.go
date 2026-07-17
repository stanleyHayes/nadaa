package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

// Config holds shelter-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	// TokenSecret verifies auth-service nadaa.<payload>.<sig> bearer tokens
	// (NADAA_AUTH_TOKEN_SECRET). Empty means authority requests are rejected
	// unless AllowMockActors is on.
	TokenSecret string
	// AllowMockActors honors the legacy X-NADAA-Actor-* headers for local dev
	// and smoke tests (NADAA_AUTH_ALLOW_MOCK_ACTORS=true). Off by default.
	AllowMockActors bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:            resolveListenAddr("NADAA_SHELTER_ADDR", ":8093"),
		AllowedOrigins:  utils.AllowedOriginsFromEnv(),
		TokenSecret:     strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		AllowMockActors: strings.TrimSpace(strings.ToLower(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS"))) == "true",
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
