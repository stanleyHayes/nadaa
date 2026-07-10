package config

import "github.com/stanleyHayes/nadaa/services/open-data-service/internal/utils"

// Config holds open-data-service configuration loaded from the environment.
type Config struct {
	Addr                   string
	AuditLogServiceURL     string
	AllowedOrigins         map[string]bool
	RateLimitRequests      int
	RateLimitWindowSeconds int
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                   utils.EnvOrDefault("PORT", ":8102"),
		AuditLogServiceURL:     utils.EnvOrDefault("AUDIT_LOG_SERVICE_URL", "http://localhost:8080"),
		AllowedOrigins:         utils.AllowedOriginsFromEnv(),
		RateLimitRequests:      utils.EnvOrDefaultInt("RATE_LIMIT_REQUESTS", 10),
		RateLimitWindowSeconds: utils.EnvOrDefaultInt("RATE_LIMIT_WINDOW_SECONDS", 60),
	}
}
