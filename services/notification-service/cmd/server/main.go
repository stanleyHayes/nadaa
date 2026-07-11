package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/client"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

const serviceName = "notification-service"

func main() {
	cfg := config.Load()
	now := time.Now().UTC()
	s := store.NewMemoryStore(now)
	alertClient := client.NewAlertServiceClient(utils.EnvOrDefault("NADAA_ALERT_SERVICE_URL", "http://localhost:8089/api/v1"))
	incidentClient := client.NewIncidentServiceClient(os.Getenv("NADAA_INCIDENT_SERVICE_URL"))
	providers := handlers.BuildProviders(cfg.Providers)
	cellBroadcast := handlers.CellBroadcastAdapterFromMode(cfg.CellBroadcastMode)
	srv := handlers.NewServer(s, alertClient, incidentClient, providers, cellBroadcast, func() time.Time { return time.Now().UTC() }, cfg)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      srv.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		utils.LogInfo(
			serviceName+" starting",
			"addr", cfg.Addr,
			"alertClientConfigured", alertClient != nil,
			"incidentClientConfigured", incidentClient != nil,
			"pushProvider", utils.ProviderName(providers["push"]),
			"smsProvider", utils.ProviderName(providers["sms"]),
			"voiceProvider", utils.ProviderName(providers["voice"]),
			"cellBroadcastAdapter", cellBroadcast.Name(),
		)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			utils.LogError(serviceName+" stopped", "addr", cfg.Addr, "error", err)
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
