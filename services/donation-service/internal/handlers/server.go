package handlers

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/store"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store    store.Store
	payments models.PaymentProvider
	now      func() time.Time
	config   *config.Config
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, payments models.PaymentProvider, now func() time.Time, cfg *config.Config) *Server {
	return &Server{
		store:    s,
		payments: payments,
		now:      now,
		config:   cfg,
	}
}
