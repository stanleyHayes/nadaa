package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/utils"
)

// Config holds open-data-service configuration loaded from the environment.
type Config struct {
	Addr                   string
	AuditLogServiceURL     string
	AllowedOrigins         map[string]bool
	RateLimitRequests      int
	RateLimitWindowSeconds int
	AuthTokenSecret        string
	AllowMockActors        bool
	TrustProxyHeaders      bool
	// InternalServiceToken authenticates service-to-service calls (sent as
	// X-NADAA-Service-Token) when forwarding audit events to the audit log
	// service.
	InternalServiceToken string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:                   resolveListenAddr("", ":8102"),
		AuditLogServiceURL:     utils.EnvOrDefault("AUDIT_LOG_SERVICE_URL", "http://localhost:8080"),
		AllowedOrigins:         utils.AllowedOriginsFromEnv(),
		RateLimitRequests:      utils.EnvOrDefaultInt("RATE_LIMIT_REQUESTS", 10),
		RateLimitWindowSeconds: utils.EnvOrDefaultInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		AuthTokenSecret:        os.Getenv("NADAA_AUTH_TOKEN_SECRET"),
		AllowMockActors:        os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS") == "true",
		TrustProxyHeaders:      os.Getenv("NADAA_TRUST_PROXY_HEADERS") == "true",
		InternalServiceToken:   strings.TrimSpace(os.Getenv("NADAA_INTERNAL_SERVICE_TOKEN")),
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
