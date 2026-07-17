package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/utils"
)

// Config holds campaign-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	TokenSecret    string
	// AllowMockActorHeaders lets authority endpoints accept the legacy X-NADAA-*
	// actor headers instead of a verified bearer token. Local dev/smoke tests
	// only — it trusts client-supplied identity headers, so it must stay false
	// in production.
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
		Addr:                  resolveListenAddr("NADAA_CAMPAIGN_ADDR", ":8103"),
		AllowedOrigins:        utils.AllowedOriginsFromEnv(),
		TokenSecret:           strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		AllowMockActorHeaders: os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true",
	}
}
