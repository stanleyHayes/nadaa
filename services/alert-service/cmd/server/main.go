package main

import (
	"log"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/handlers"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/store"
)

func main() {
	cfg := config.Load()
	s := store.NewMemoryStore(time.Now().UTC())
	srv := handlers.NewServer(s, func() time.Time { return time.Now().UTC() }, cfg)

	addr := cfg.Addr
	log.Printf("alert-service listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
