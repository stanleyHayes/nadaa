package main

import (
	"log"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore()
	srv := handlers.NewServer(s, cfg)

	addr := cfg.Addr
	log.Printf("risk-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
