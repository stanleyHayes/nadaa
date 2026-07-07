package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/client"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

func main() {
	cfg := config.Load()
	now := time.Now().UTC()
	s := store.NewMemoryStore(now)
	alertClient := client.NewAlertServiceClient(utils.EnvOrDefault("NADAA_ALERT_SERVICE_URL", "http://localhost:8089/api/v1"))
	incidentClient := client.NewIncidentServiceClient(os.Getenv("NADAA_INCIDENT_SERVICE_URL"))
	providers := handlers.ProvidersFromEnv()
	srv := handlers.NewServer(s, alertClient, incidentClient, providers, func() time.Time { return time.Now().UTC() }, cfg)

	addr := cfg.Addr
	utils.LogInfo(
		"notification-service starting",
		"addr", addr,
		"alertClientConfigured", alertClient != nil,
		"incidentClientConfigured", incidentClient != nil,
		"pushProvider", utils.ProviderName(providers["push"]),
		"smsProvider", utils.ProviderName(providers["sms"]),
		"voiceProvider", utils.ProviderName(providers["voice"]),
	)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		utils.LogError("notification-service stopped", "addr", addr, "error", err)
		log.Fatal(err)
	}
}
