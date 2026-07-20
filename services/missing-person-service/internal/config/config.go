package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/utils"
)

// Config holds missing-person-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	// AuthTokenSecret verifies NADAA auth-service bearer tokens; when empty,
	// token-based authority authentication always fails.
	AuthTokenSecret string
	// AllowMockActors honors legacy X-NADAA-Actor-* headers for local
	// development and smoke tests (NADAA_AUTH_ALLOW_MOCK_ACTORS=true).
	AllowMockActors bool
	// RateLimitRequests bounds public intake requests per client IP per window;
	// a value <= 0 disables the limiter.
	RateLimitRequests int
	// RateLimitWindowSeconds is the rate limit window in seconds.
	RateLimitWindowSeconds int
	// MaxRecords caps the number of stored records; public intake is refused
	// once the cap is reached. A value <= 0 means no cap.
	MaxRecords int
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                   resolveListenAddr("", ":8101"),
		AllowedOrigins:         utils.AllowedOriginsFromEnv(),
		AuthTokenSecret:        strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		AllowMockActors:        strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS")), "true"),
		RateLimitRequests:      utils.EnvOrDefaultInt("RATE_LIMIT_REQUESTS", 10),
		RateLimitWindowSeconds: utils.EnvOrDefaultInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		MaxRecords:             utils.EnvOrDefaultInt("MISSING_PERSON_MAX_RECORDS", 10000),
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
