package config

import (
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

// Config holds integration-service configuration loaded from the environment.
type Config struct {
	Addr               string
	RoadClosureAPIURL  string
	AllowedOrigins     map[string]bool
	SchedulerEnabled   bool
	SchedulerInterval  time.Duration
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:              utils.EnvOrDefault("NADAA_INTEGRATION_ADDR", ":8088"),
		RoadClosureAPIURL: strings.TrimRight(utils.EnvOrDefault("NADAA_ROAD_CLOSURE_SERVICE_URL", "http://localhost:8095"), "/"),
		AllowedOrigins:    utils.AllowedOriginsFromEnv(),
		SchedulerEnabled:  schedulerEnabled(),
		SchedulerInterval: schedulerInterval(),
	}
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
