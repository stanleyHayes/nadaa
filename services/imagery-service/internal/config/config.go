// Package config loads imagery-service configuration from the environment.
package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/utils"
)

// DefaultRetentionDays is the default imagery retention period.
const DefaultRetentionDays = 90

// Config holds imagery-service configuration loaded from the environment.
type Config struct {
	Addr           string
	StoragePath    string
	RetentionDays  int
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	retentionDays, err := strconv.Atoi(utils.EnvOrDefault("DEFAULT_RETENTION_DAYS", strconv.Itoa(DefaultRetentionDays)))
	if err != nil || retentionDays <= 0 {
		retentionDays = DefaultRetentionDays
	}
	return &Config{
		Addr:           resolveListenAddr("", ":8099"),
		StoragePath:    utils.EnvOrDefault("IMAGERY_STORAGE_PATH", "./uploads"),
		RetentionDays:  retentionDays,
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
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
