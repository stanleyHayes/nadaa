package config

import (
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

// Config holds shelter-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           utils.EnvOrDefault("NADAA_SHELTER_ADDR", ":8093"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
	}
}
