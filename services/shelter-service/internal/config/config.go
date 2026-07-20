package config

import (
	"errors"
	"os"
	"strconv"
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
	// PledgeRateLimit caps how many aid pledges one client may create per
	// PledgeRateWindowSecs (NADAA_PLEDGE_RATE_LIMIT /
	// NADAA_PLEDGE_RATE_WINDOW_SECONDS); pledge creation is unauthenticated,
	// so it is throttled per client.
	PledgeRateLimit      int
	PledgeRateWindowSecs int
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                 resolveListenAddr("NADAA_SHELTER_ADDR", ":8093"),
		AllowedOrigins:       utils.AllowedOriginsFromEnv(),
		TokenSecret:          strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		AllowMockActors:      strings.TrimSpace(strings.ToLower(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS"))) == "true",
		PledgeRateLimit:      envIntOrDefault("NADAA_PLEDGE_RATE_LIMIT", 10),
		PledgeRateWindowSecs: envIntOrDefault("NADAA_PLEDGE_RATE_WINDOW_SECONDS", 60),
	}
}

// Validate fails closed on unsafe configuration: mock actor headers trust
// client-supplied identity, so they are rejected unless NADAA_ENV=development.
func (c *Config) Validate() error {
	if c.AllowMockActors && !strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_ENV")), "development") {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
}

func envIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
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
