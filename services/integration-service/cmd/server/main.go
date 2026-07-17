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

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/store"
)

const serviceName = "integration-service"

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore(time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
	srv := handlers.NewServer(s, &http.Client{Timeout: 15 * time.Second}, cfg.RoadClosureAPIURL, cfg.AllowMockActors)

	if cfg.SchedulerEnabled {
		go startObservationImportScheduler(s, cfg.SchedulerInterval)
	}

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      srv.Routes(cfg.AllowedOrigins),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("%s listening on %s", serviceName, cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
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

func startObservationImportScheduler(s store.Store, interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Printf("weather/hydrology import scheduler enabled with interval %s", interval)
	for now := range ticker.C {
		job := s.CreateObservationImportJob(models.ObservationImportRequest{}, "scheduled", now.UTC(), 1)
		log.Printf("scheduled weather/hydrology import %s finished with status %s and %d imported observations", job.ID, job.Status, job.ImportedCount)
	}
}
