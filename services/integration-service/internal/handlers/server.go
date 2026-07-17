package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/store"
)

// server holds the HTTP handler dependencies.
type server struct {
	store             store.Store
	httpClient        *http.Client
	roadClosureAPIURL string
	allowMockActors   bool
}

// NewServer creates a new server with the given dependencies.
func NewServer(s store.Store, httpClient *http.Client, roadClosureAPIURL string, allowMockActors bool) *server {
	return &server{
		store:             s,
		httpClient:        httpClient,
		roadClosureAPIURL: roadClosureAPIURL,
		allowMockActors:   allowMockActors,
	}
}
