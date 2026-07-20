// Package config loads imagery-service configuration from the environment.
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

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
	// PublicBaseURL is the externally reachable base URL (scheme + host) used
	// to build absolute download URLs in the public geojson feed. When empty,
	// the request's scheme and Host header are used as a fallback.
	PublicBaseURL string
	// TokenSecret verifies NADAA bearer tokens (NADAA_AUTH_TOKEN_SECRET).
	TokenSecret string
	// AllowMockActors honors legacy X-NADAA-Actor-* headers (local dev only).
	AllowMockActors bool
	// Development enables local-dev relaxations (NADAA_ENV=development).
	Development bool
}

// Load reads configuration from environment variables.
func Load() *Config {
	retentionDays, err := strconv.Atoi(utils.EnvOrDefault("DEFAULT_RETENTION_DAYS", strconv.Itoa(DefaultRetentionDays)))
	if err != nil || retentionDays <= 0 {
		retentionDays = DefaultRetentionDays
	}
	return &Config{
		Addr:            resolveListenAddr("", ":8099"),
		StoragePath:     utils.EnvOrDefault("IMAGERY_STORAGE_PATH", "./uploads"),
		RetentionDays:   retentionDays,
		AllowedOrigins:  utils.AllowedOriginsFromEnv(),
		PublicBaseURL:   strings.TrimSuffix(utils.EnvOrDefault("NADAA_IMAGERY_PUBLIC_BASE_URL", ""), "/"),
		TokenSecret:     utils.EnvOrDefault("NADAA_AUTH_TOKEN_SECRET", ""),
		AllowMockActors: os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true",
		Development:     os.Getenv("NADAA_ENV") == "development",
	}
}

// Validate fails closed on unsafe configuration: mock actor headers trust
// client-supplied identity, so they are only allowed with NADAA_ENV=development
// and can never leak into a deployed environment.
func (c *Config) Validate() error {
	if c.AllowMockActors && !c.Development {
		return errors.New("NADAA_AUTH_ALLOW_MOCK_ACTORS is only allowed when NADAA_ENV=development")
	}
	return nil
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
