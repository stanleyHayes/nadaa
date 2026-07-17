package config

import (
	"os"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

// Config holds integration-service configuration loaded from the environment.
type Config struct {
	Addr              string
	RoadClosureAPIURL string
	AllowedOrigins    map[string]bool
	AllowMockActors   bool
	SchedulerEnabled  bool
	SchedulerInterval time.Duration
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:              resolveListenAddr("NADAA_INTEGRATION_ADDR", ":8088"),
		RoadClosureAPIURL: strings.TrimRight(utils.EnvOrDefault("NADAA_ROAD_CLOSURE_SERVICE_URL", "http://localhost:8095"), "/"),
		AllowedOrigins:    utils.AllowedOriginsFromEnv(),
		AllowMockActors:   mockActorsEnabled(),
		SchedulerEnabled:  schedulerEnabled(),
		SchedulerInterval: schedulerInterval(),
	}
}

// mockActorsEnabled gates forwarding of legacy X-NADAA-Actor-* headers for local
// dev and smoke tests; production forwards the caller's bearer token instead.
func mockActorsEnabled() bool {
	value := utils.NormalizeQueryValue(utils.EnvOrDefault("NADAA_AUTH_ALLOW_MOCK_ACTORS", ""))
	return value == "true" || value == "1" || value == "yes"
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

func schedulerEnabled() bool {
	value := utils.NormalizeQueryValue(utils.EnvOrDefault("NADAA_IMPORT_SCHEDULER_ENABLED", ""))
	return value == "true" || value == "1" || value == "yes"
}

func schedulerInterval() time.Duration {
	value := strings.TrimSpace(utils.EnvOrDefault("NADAA_IMPORT_SCHEDULER_INTERVAL", ""))
	if value == "" {
		return 15 * time.Minute
	}
	interval, err := time.ParseDuration(value)
	if err != nil || interval <= 0 {
		return 15 * time.Minute
	}
	return interval
}
