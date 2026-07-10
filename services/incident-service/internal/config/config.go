package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds incident-service configuration loaded from the environment.
type Config struct {
	Addr           string
	RateLimit      int
	RateWindowSecs int
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           envOrDefault("NADAA_INCIDENT_ADDR", ":8084"),
		RateLimit:      envIntOrDefault("NADAA_INCIDENT_RATE_LIMIT", 60),
		RateWindowSecs: envIntOrDefault("NADAA_INCIDENT_RATE_WINDOW_SECONDS", 60),
		AllowedOrigins: allowedOriginsFromEnv(),
	}
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
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
