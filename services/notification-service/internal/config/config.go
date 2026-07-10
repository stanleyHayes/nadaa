package config

import "github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"

// Config holds notification-service configuration loaded from the environment.
type Config struct {
	Addr              string
	AllowedOrigins    map[string]bool
	CellBroadcastMode string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:              utils.EnvOrDefault("NADAA_NOTIFICATION_ADDR", ":8090"),
		AllowedOrigins:    utils.AllowedOriginsFromEnv(),
		CellBroadcastMode: utils.EnvOrDefault("NADAA_CELL_BROADCAST_MODE", "disabled"),
	}
}
