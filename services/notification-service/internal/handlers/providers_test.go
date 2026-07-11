package handlers

import (
	"testing"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

func TestBuildProvidersDefaultsToSandbox(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{})

	for _, channel := range []string{"push", "sms", "voice"} {
		if _, ok := providers[channel].(models.MockProvider); !ok {
			t.Errorf("channel %q = %T, want MockProvider by default", channel, providers[channel])
		}
	}
}

func TestBuildProvidersSelectsArkeselWithKey(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{
		SMSProvider:   "arkesel",
		ArkeselAPIKey: "secret-key",
		ArkeselSender: "NADAA",
	})

	if _, ok := providers["sms"].(models.ArkeselSMSProvider); !ok {
		t.Fatalf("sms provider = %T, want ArkeselSMSProvider", providers["sms"])
	}
	if name := utils.ProviderName(providers["sms"]); name != "arkesel_sms" {
		t.Errorf("provider name = %q, want arkesel_sms", name)
	}
}

func TestBuildProvidersDisablesArkeselWithoutKey(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{SMSProvider: "arkesel"})

	disabled, ok := providers["sms"].(models.DisabledProvider)
	if !ok {
		t.Fatalf("sms provider = %T, want DisabledProvider when the api key is missing", providers["sms"])
	}
	if disabled.Reason == "" {
		t.Error("disabled provider should explain why arkesel was not used")
	}
}

func TestBuildProvidersSelectsExpo(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{PushProvider: "expo"})

	if _, ok := providers["push"].(models.ExpoPushProvider); !ok {
		t.Fatalf("push provider = %T, want ExpoPushProvider", providers["push"])
	}
	if name := utils.ProviderName(providers["push"]); name != "expo_push" {
		t.Errorf("provider name = %q, want expo_push", name)
	}
}

func TestBuildProvidersDisabledSelection(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{
		SMSProvider:   "disabled",
		PushProvider:  "disabled",
		VoiceProvider: "disabled",
	})

	for _, channel := range []string{"push", "sms", "voice"} {
		if _, ok := providers[channel].(models.DisabledProvider); !ok {
			t.Errorf("channel %q = %T, want DisabledProvider", channel, providers[channel])
		}
	}
}

func TestBuildProvidersUnknownSelectionFailsSafe(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{SMSProvider: "twilio"})

	if _, ok := providers["sms"].(models.DisabledProvider); !ok {
		t.Fatalf("sms provider = %T, want DisabledProvider for an unknown selection", providers["sms"])
	}
}

func TestBuildProvidersVoiceArkeselNotWiredYet(t *testing.T) {
	providers := BuildProviders(config.ProviderConfig{VoiceProvider: "arkesel"})

	if _, ok := providers["voice"].(models.DisabledProvider); !ok {
		t.Fatalf("voice provider = %T, want DisabledProvider until arkesel voice is wired", providers["voice"])
	}
}
