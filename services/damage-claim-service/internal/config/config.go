package config

import (
	"errors"
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"
)

// Config holds damage-claim-service configuration loaded from the environment.
type Config struct {
	Addr                 string
	IncidentServiceURL   string
	AllowedOrigins       map[string]bool
	AuthTokenSecret      string
	InternalServiceToken string
	AllowMockActors      bool
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

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                 resolveListenAddr("", ":8098"),
		IncidentServiceURL:   utils.TrimTrailingSlash(utils.EnvOrDefault("INCIDENT_SERVICE_URL", "http://localhost:8084/api/v1")),
		AllowedOrigins:       utils.AllowedOriginsFromEnv(),
		AuthTokenSecret:      os.Getenv("NADAA_AUTH_TOKEN_SECRET"),
		InternalServiceToken: strings.TrimSpace(os.Getenv("NADAA_INTERNAL_SERVICE_TOKEN")),
		AllowMockActors:      strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS")), "true"),
	}
}

// Validate rejects configuration combinations that are unsafe outside local
// development. Self-asserted X-NADAA-Actor-* headers enable full impersonation,
// so they must never be honored in a deployed environment.
func (c *Config) Validate() error {
	if c.AllowMockActors && !strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_ENV")), "development") {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
}
