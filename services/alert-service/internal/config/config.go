package config

import (
	"os"
	"strings"
)

// Config holds alert-service configuration loaded from the environment.
type Config struct {
	Addr            string
	AllowedOrigins  map[string]bool
	TokenSecret     string
	AllowMockActors bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:            resolveListenAddr("NADAA_ALERT_ADDR", ":8089"),
		AllowedOrigins:  allowedOriginsFromEnv(),
		TokenSecret:     strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		AllowMockActors: strings.TrimSpace(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS")) == "true",
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
