package handlers

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/client"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/store"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store          store.Store
	alertClient    *client.AlertServiceClient
	incidentClient *client.IncidentServiceClient
	providers      map[string]models.NotificationProvider
	cellBroadcast  models.CellBroadcastAdapter
	now            func() time.Time
	config         *config.Config
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, alertClient *client.AlertServiceClient, incidentClient *client.IncidentServiceClient, providers map[string]models.NotificationProvider, cellBroadcast models.CellBroadcastAdapter, now func() time.Time, cfg *config.Config) *Server {
	return &Server{
		store:          s,
		alertClient:    alertClient,
		incidentClient: incidentClient,
		providers:      providers,
		cellBroadcast:  cellBroadcast,
		now:            now,
		config:         cfg,
	}
}
