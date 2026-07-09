// Package config loads imagery-service configuration from the environment.
package config

import (
	"strconv"

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
		Addr:           utils.EnvOrDefault("PORT", ":8099"),
		StoragePath:    utils.EnvOrDefault("IMAGERY_STORAGE_PATH", "./uploads"),
		RetentionDays:  retentionDays,
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
	}
}
