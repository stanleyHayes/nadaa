package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/route-service/internal/utils"
)

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
		Addr:                  resolveListenAddr("", ":8096"),
		RoadClosureServiceURL: utils.EnvOrDefault("ROAD_CLOSURE_SERVICE_URL", "http://localhost:8095"),
		ShelterServiceURL:     utils.EnvOrDefault("SHELTER_SERVICE_URL", "http://localhost:8093"),
		RiskServiceURL:        utils.EnvOrDefault("RISK_SERVICE_URL", "http://localhost:8082"),
		AllowedOrigins:        utils.AllowedOriginsFromEnv(),
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
