package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore(time.Now().UTC(), cfg)
	srv := handlers.NewServer(s, time.Now, cfg)

	addr := cfg.Addr
	log.Printf("auth-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
