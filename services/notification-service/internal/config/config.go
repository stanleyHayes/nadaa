package config

import (
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// Config holds notification-service configuration loaded from the environment.
type Config struct {
	Addr              string
	AllowedOrigins    map[string]bool
	CellBroadcastMode string
	Providers         ProviderConfig
}

// ProviderConfig selects the delivery provider per channel and carries the
// credentials each real provider needs. Everything defaults to the sandbox
// (mock) provider so the platform runs end-to-end before real credentials
// arrive; a channel only reaches a live backend once its provider is explicitly
// selected and its credentials are present (see handlers.BuildProviders).
type ProviderConfig struct {
	// Per-channel provider selection: "sandbox" (mock), "disabled", or a real
	// provider id such as "arkesel" (sms) or "expo" (push).
	SMSProvider   string
	PushProvider  string
	VoiceProvider string

	// Arkesel (SMS today; Voice/USSD share the same account per the provider
	// research) credentials.
	ArkeselAPIKey  string
	ArkeselSender  string
	ArkeselBaseURL string

	// Expo push credentials. AccessToken is optional (enhanced push security);
	// BaseURL is overridable for tests.
	ExpoAccessToken string
	ExpoBaseURL     string
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Addr:              resolveListenAddr("NADAA_NOTIFICATION_ADDR", ":8090"),
		AllowedOrigins:    utils.AllowedOriginsFromEnv(),
		CellBroadcastMode: utils.EnvOrDefault("NADAA_CELL_BROADCAST_MODE", "disabled"),
		Providers: ProviderConfig{
			SMSProvider:   providerSelection("NADAA_SMS_PROVIDER", "NADAA_SMS_ENABLED"),
			PushProvider:  providerSelection("NADAA_PUSH_PROVIDER", "NADAA_PUSH_ENABLED"),
			VoiceProvider: providerSelection("NADAA_VOICE_PROVIDER", "NADAA_VOICE_ENABLED"),

			ArkeselAPIKey:  utils.EnvOrDefault("NADAA_ARKESEL_API_KEY", ""),
			ArkeselSender:  utils.EnvOrDefault("NADAA_ARKESEL_SENDER", "NADAA"),
			ArkeselBaseURL: utils.EnvOrDefault("NADAA_ARKESEL_BASE_URL", ""),

			ExpoAccessToken: utils.EnvOrDefault("NADAA_EXPO_ACCESS_TOKEN", ""),
			ExpoBaseURL:     utils.EnvOrDefault("NADAA_EXPO_BASE_URL", ""),
		},
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

// providerSelection resolves a channel's provider id. The legacy
// NADAA_<CHANNEL>_ENABLED flag still wins when set to a falsey value so existing
// deployments that disable a channel keep working; otherwise the explicit
// NADAA_<CHANNEL>_PROVIDER value is used, defaulting to the sandbox provider.
func providerSelection(providerKey, enabledKey string) string {
	if !utils.EnvBool(enabledKey, true) {
		return "disabled"
	}
	return strings.ToLower(utils.EnvOrDefault(providerKey, "sandbox"))
}
