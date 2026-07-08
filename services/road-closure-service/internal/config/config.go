package config

import (
	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

// Config holds road-closure-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           utils.EnvOrDefault("NADAA_ROAD_CLOSURE_ADDR", ":8095"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
	}
}
