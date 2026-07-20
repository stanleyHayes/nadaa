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

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/store"
)

const serviceName = "ml-service"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	s, err := store.NewMemoryStore(cfg.ModelDir)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.InternalServiceToken == "" {
		log.Printf("WARN %s NADAA_INTERNAL_SERVICE_TOKEN is not set; the service-token path is closed and only verified agency bearer tokens (or mock actors in development) are accepted", serviceName)
	}

	srv := handlers.NewServer(s, time.Now, cfg)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      srv.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("%s listening on %s with model %s", serviceName, cfg.Addr, s.ModelVersion())
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
