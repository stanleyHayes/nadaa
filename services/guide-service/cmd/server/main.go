package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore(time.Now().UTC())
	srv := handlers.NewServer(s, time.Now, cfg)

	addr := cfg.Addr
	log.Printf("guide-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
