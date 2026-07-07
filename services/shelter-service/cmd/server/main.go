package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/store"
)

func main() {
	cfg := config.Load()
	now := time.Now().UTC()
	s := store.NewMemoryStore(now)
	srv := handlers.NewServer(s, time.Now, cfg)

	addr := cfg.Addr
	log.Printf("shelter-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
