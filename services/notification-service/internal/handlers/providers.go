package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
)

// providerHTTPTimeout bounds every outbound provider request.
const providerHTTPTimeout = 10 * time.Second

// BuildProviders selects a concrete NotificationProvider per channel from
// configuration. It is the single dependency-injection seam for the delivery
// gateway: a real backend (Arkesel for SMS, Expo for push) is chosen only when
// its provider is explicitly selected and its credentials are present. Every
// channel defaults to the sandbox (mock) provider so the platform runs
// end-to-end before real credentials arrive, and it fails safe — a channel that
// selects a real provider without its credentials is disabled with a clear
// reason rather than silently mocking a live selection or crashing. This mirrors
// the CellBroadcastAdapterFromMode fail-safe default.
func BuildProviders(cfg config.ProviderConfig) map[string]models.NotificationProvider {
	httpClient := &http.Client{Timeout: providerHTTPTimeout}
	return map[string]models.NotificationProvider{
		"push":  buildPushProvider(cfg, httpClient),
		"sms":   buildSMSProvider(cfg, httpClient),
		"voice": buildVoiceProvider(cfg),
	}
}

func buildSMSProvider(cfg config.ProviderConfig, client *http.Client) models.NotificationProvider {
	switch normalizeProvider(cfg.SMSProvider) {
	case "arkesel":
		if strings.TrimSpace(cfg.ArkeselAPIKey) == "" {
			return models.DisabledProvider{Channel: "sms", Reason: "sms provider 'arkesel' selected but NADAA_ARKESEL_API_KEY is not set"}
		}
		return models.NewArkeselSMSProvider(cfg.ArkeselAPIKey, cfg.ArkeselSender, cfg.ArkeselBaseURL, client)
	case "", "sandbox", "mock":
		return models.MockProvider{Channel: "sms"}
	case "disabled", "off", "none":
		return models.DisabledProvider{Channel: "sms", Reason: "sms provider disabled"}
	default:
		return models.DisabledProvider{Channel: "sms", Reason: "unknown sms provider '" + normalizeProvider(cfg.SMSProvider) + "'"}
	}
}

func buildPushProvider(cfg config.ProviderConfig, client *http.Client) models.NotificationProvider {
	switch normalizeProvider(cfg.PushProvider) {
	case "expo":
		// Expo push needs no API key (the access token is optional), so there is
		// no credential guard here beyond the explicit selection.
		return models.NewExpoPushProvider(cfg.ExpoAccessToken, cfg.ExpoBaseURL, client)
	case "", "sandbox", "mock":
		return models.MockProvider{Channel: "push"}
	case "disabled", "off", "none":
		return models.DisabledProvider{Channel: "push", Reason: "push provider disabled"}
	default:
		return models.DisabledProvider{Channel: "push", Reason: "unknown push provider '" + normalizeProvider(cfg.PushProvider) + "'"}
	}
}

func buildVoiceProvider(cfg config.ProviderConfig) models.NotificationProvider {
	switch normalizeProvider(cfg.VoiceProvider) {
	case "", "sandbox", "mock":
		return models.MockProvider{Channel: "voice"}
	case "disabled", "off", "none":
		return models.DisabledProvider{Channel: "voice", Reason: "voice provider disabled"}
	case "arkesel":
		// Arkesel Voice is the researched value-for-money choice for Ghana, but
		// its campaign/DTMF API needs a paid pilot before wiring the live path.
		// Until then, fail safe rather than silently mock a live selection.
		return models.DisabledProvider{Channel: "voice", Reason: "voice provider 'arkesel' is not wired yet; use 'sandbox' until the Arkesel Voice campaign API is integrated"}
	default:
		return models.DisabledProvider{Channel: "voice", Reason: "unknown voice provider '" + normalizeProvider(cfg.VoiceProvider) + "'"}
	}
}

func normalizeProvider(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// CellBroadcastAdapterFromMode selects the cell broadcast adapter for the given
// mode. It defaults to a disabled no-op so an unconfigured or unknown mode can
// never emit a live broadcast; "sandbox" enables the simulator for end-to-end
// testing. A real telecom integration would register its adapter here.
func CellBroadcastAdapterFromMode(mode string) models.CellBroadcastAdapter {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "sandbox":
		return models.SandboxCellBroadcastAdapter{}
	default:
		return models.DisabledCellBroadcastAdapter{Reason: "telecom cell broadcast path is not configured"}
	}
}
