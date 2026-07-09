package config

import "github.com/stanleyHayes/nadaa/services/missing-person-service/internal/utils"

// Config holds missing-person-service configuration loaded from the environment.
type Config struct {
	Addr           string
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           utils.EnvOrDefault("PORT", ":8101"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
	}
}
