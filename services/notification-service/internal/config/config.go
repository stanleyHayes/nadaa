package config

import (
	"errors"
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// Config holds notification-service configuration loaded from the environment.
type Config struct {
	Addr              string
	AllowedOrigins    map[string]bool
	CellBroadcastMode string
	Env               string
	TokenSecret       string
	AllowMockActors   bool
	// AllowFixtureAlerts re-enables serving seeded fixture alerts outside
	// NADAA_ENV=development. Fixture alerts never went through the alert
	// approval workflow, so production must leave this off.
	AllowFixtureAlerts bool
	WebhookSecrets     WebhookSecretConfig
	Providers          ProviderConfig
}

// WebhookSecretConfig carries the per-channel shared secrets inbound provider
// webhooks must present in the X-NADAA-Webhook-Secret header. An empty secret
// means the channel accepts unauthenticated webhooks (local development
// default; a WARN is logged at startup).
type WebhookSecretConfig struct {
	SMS      string
	USSD     string
	WhatsApp string
	Voice    string
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
		Addr:               resolveListenAddr("NADAA_NOTIFICATION_ADDR", ":8090"),
		AllowedOrigins:     utils.AllowedOriginsFromEnv(),
		CellBroadcastMode:  utils.EnvOrDefault("NADAA_CELL_BROADCAST_MODE", "disabled"),
		Env:                strings.TrimSpace(os.Getenv("NADAA_ENV")),
		TokenSecret:        strings.TrimSpace(os.Getenv("NADAA_AUTH_TOKEN_SECRET")),
		AllowMockActors:    strings.TrimSpace(os.Getenv("NADAA_AUTH_ALLOW_MOCK_ACTORS")) == "true",
		AllowFixtureAlerts: utils.EnvBool("NADAA_NOTIFICATION_ALLOW_FIXTURE_ALERTS", false),
		WebhookSecrets: WebhookSecretConfig{
			SMS:      strings.TrimSpace(os.Getenv("NADAA_SMS_WEBHOOK_SECRET")),
			USSD:     strings.TrimSpace(os.Getenv("NADAA_USSD_WEBHOOK_SECRET")),
			WhatsApp: strings.TrimSpace(os.Getenv("NADAA_WHATSAPP_WEBHOOK_SECRET")),
			Voice:    strings.TrimSpace(os.Getenv("NADAA_VOICE_WEBHOOK_SECRET")),
		},
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

// FixtureAlertsEnabled reports whether seeded fixture alerts may be merged into
// the citizen feed and delivery path. Fixtures are demo/smoke data that never
// went through the alert approval workflow, so they are limited to
// NADAA_ENV=development unless explicitly re-enabled with
// NADAA_NOTIFICATION_ALLOW_FIXTURE_ALERTS=true.
func (c *Config) FixtureAlertsEnabled() bool {
	return c.AllowFixtureAlerts || utils.NormalizeQueryValue(c.Env) == "development"
}

// Validate fails closed on unsafe configuration. Mock actor headers trust
// self-asserted X-NADAA-Actor-* identity, so they are only allowed when
// NADAA_ENV=development (local development and smoke tests).
func (c *Config) Validate() error {
	if c.AllowMockActors && utils.NormalizeQueryValue(c.Env) != "development" {
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
