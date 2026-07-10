package config

import "github.com/stanleyHayes/nadaa/services/school-service/internal/utils"

// Config holds school-service configuration loaded from the environment.
type Config struct {
	Addr           string
	RiskServiceURL string
	AllowedOrigins map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:           utils.EnvOrDefault("PORT", ":8097"),
		RiskServiceURL: utils.EnvOrDefault("RISK_SERVICE_URL", "http://localhost:8082"),
		AllowedOrigins: utils.AllowedOriginsFromEnv(),
	}
}
