package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore()
	srv := handlers.NewServer(s, time.Now, cfg)

	addr := cfg.Addr
	log.Printf("incident-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes(cfg.AllowedOrigins)); err != nil {
		log.Fatal(err)
	}
}
