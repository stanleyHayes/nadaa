package config

import (
	"os"
	"strings"
)

// Config holds risk-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
	MLAPIURL       string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           envOrDefault("NADAA_RISK_ADDR", ":8081"),
		AllowedOrigins: allowedOriginsFromEnv(),
		MLAPIURL:       strings.TrimSpace(os.Getenv("NADAA_ML_API_URL")),
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
