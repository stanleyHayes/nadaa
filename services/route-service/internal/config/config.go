package config

import "github.com/stanleyHayes/nadaa/services/route-service/internal/utils"

// Config holds route-service configuration loaded from the environment.
type Config struct {
	Addr                  string
	RoadClosureServiceURL string
	ShelterServiceURL     string
	RiskServiceURL        string
	AllowedOrigins        map[string]bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                  utils.EnvOrDefault("PORT", ":8096"),
		RoadClosureServiceURL: utils.EnvOrDefault("ROAD_CLOSURE_SERVICE_URL", "http://localhost:8095"),
		ShelterServiceURL:     utils.EnvOrDefault("SHELTER_SERVICE_URL", "http://localhost:8093"),
		RiskServiceURL:        utils.EnvOrDefault("RISK_SERVICE_URL", "http://localhost:8082"),
		AllowedOrigins:        utils.AllowedOriginsFromEnv(),
	}
}
