package handlers

import (
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// ProvidersFromEnv builds the notification provider map from environment flags.
func ProvidersFromEnv() map[string]models.NotificationProvider {
	providers := map[string]models.NotificationProvider{}

	if utils.EnvBool("NADAA_PUSH_ENABLED", true) {
		providers["push"] = models.MockProvider{Channel: "push"}
	} else {
		providers["push"] = models.DisabledProvider{Channel: "push", Reason: "push provider disabled"}
	}

	if utils.EnvBool("NADAA_SMS_ENABLED", true) {
		providers["sms"] = models.MockProvider{Channel: "sms"}
	} else {
		providers["sms"] = models.DisabledProvider{Channel: "sms", Reason: "sms provider disabled"}
	}

	if utils.EnvBool("NADAA_VOICE_ENABLED", true) {
		providers["voice"] = models.MockProvider{Channel: "voice"}
	} else {
		providers["voice"] = models.DisabledProvider{Channel: "voice", Reason: "voice provider disabled"}
	}

	return providers
}
