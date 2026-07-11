package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

// Config holds road-closure-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
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
	}
}
