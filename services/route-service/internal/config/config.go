package config

import (
	"errors"
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
	// TokenSecret verifies nadaa.<payload>.<sig> bearer tokens issued by
	// auth-service (NADAA_AUTH_TOKEN_SECRET).
	TokenSecret string
	// AllowMockActors honors legacy X-NADAA-Actor-* headers for local dev and
	// smoke tests (NADAA_AUTH_ALLOW_MOCK_ACTORS=true); off by default.
	AllowMockActors bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                  resolveListenAddr("", ":8096"),
		RoadClosureServiceURL: utils.EnvOrDefault("ROAD_CLOSURE_SERVICE_URL", "http://localhost:8095"),
		ShelterServiceURL:     utils.EnvOrDefault("SHELTER_SERVICE_URL", "http://localhost:8093"),
		RiskServiceURL:        utils.EnvOrDefault("RISK_SERVICE_URL", "http://localhost:8081"),
		AllowedOrigins:        utils.AllowedOriginsFromEnv(),
		TokenSecret:           os.Getenv("NADAA_AUTH_TOKEN_SECRET"),
		AllowMockActors:       os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true",
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

// Validate fails closed on unsafe configuration: honoring self-asserted
// X-NADAA-Actor-* headers is a development-only relaxation, so it is rejected
// unless NADAA_ENV=development.
func (c *Config) Validate() error {
	if c.AllowMockActors && os.Getenv("NADAA_ENV") != "development" {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
}
