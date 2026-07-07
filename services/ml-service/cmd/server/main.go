package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	s, err := store.NewMemoryStore(cfg.ModelDir)
	if err != nil {
		log.Fatal(err)
	}

	srv := handlers.NewServer(s, time.Now, cfg)

	addr := cfg.Addr
	log.Printf("ml-service listening on %s with model %s", addr, s.ModelVersion())
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
