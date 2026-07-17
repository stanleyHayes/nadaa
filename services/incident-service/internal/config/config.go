package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds incident-service configuration loaded from the environment.
type Config struct {
	Addr                 string
	RateLimit            int
	RateWindowSecs       int
	AllowedOrigins       map[string]bool
	TokenSecret          string
	InternalServiceToken string
	AllowMockActors      bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                 resolveListenAddr("NADAA_INCIDENT_ADDR", ":8084"),
		RateLimit:            envIntOrDefault("NADAA_INCIDENT_RATE_LIMIT", 60),
		RateWindowSecs:       envIntOrDefault("NADAA_INCIDENT_RATE_WINDOW_SECONDS", 60),
		AllowedOrigins:       allowedOriginsFromEnv(),
		TokenSecret:          strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		InternalServiceToken: strings.TrimSpace(os.Getenv("NADAA_INTERNAL_SERVICE_TOKEN")),
		AllowMockActors:      strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS")), "true"),
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

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
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
