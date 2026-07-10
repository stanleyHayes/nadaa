package config

import "github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"

// Config holds damage-claim-service configuration loaded from the environment.
type Config struct {
	Addr               string
	IncidentServiceURL string
	AllowedOrigins     map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:               utils.EnvOrDefault("PORT", ":8098"),
		IncidentServiceURL: utils.TrimTrailingSlash(utils.EnvOrDefault("INCIDENT_SERVICE_URL", "http://localhost:8081")),
		AllowedOrigins:     utils.AllowedOriginsFromEnv(),
	}
}
