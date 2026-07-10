// Package config loads guide-service configuration from the environment.
package config

import (
	"os"
	"strings"
)

// Config holds guide-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           envOrDefault("NADAA_GUIDE_ADDR", ":8086"),
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
