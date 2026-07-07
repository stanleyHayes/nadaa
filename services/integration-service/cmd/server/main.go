package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore(time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC))
	srv := handlers.NewServer(s, &http.Client{Timeout: 15 * time.Second}, cfg.RoadClosureAPIURL)

	if cfg.SchedulerEnabled {
		go startObservationImportScheduler(s, cfg.SchedulerInterval)
	}

	addr := cfg.Addr
	log.Printf("integration-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes(cfg.AllowedOrigins)); err != nil {
		log.Fatal(err)
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
