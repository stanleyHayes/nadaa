// Command imagery-service provides HTTP APIs for drone and satellite imagery ingestion.
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

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/store"
)

func main() {
	cfg := config.Load()
	if err := os.MkdirAll(cfg.StoragePath, 0o750); err != nil {
		log.Fatalf("ERROR imagery-service storage_path_create_failed path=%s error=%v", cfg.StoragePath, err)
	}

	s := store.NewMemoryStore(time.Now().UTC(), cfg.RetentionDays)
	srv := handlers.NewServer(s, time.Now, cfg)

	httpServer := &http.Server{
		Addr:         cfg.Addr,
		Handler:      srv.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("INFO imagery-service listening on %s storage=%s retention_days=%d", cfg.Addr, cfg.StoragePath, cfg.RetentionDays)
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
