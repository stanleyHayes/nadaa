package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/store"
)

func newTestServer() *Server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8086"}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
}

func TestHealthHandler(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestListGuides(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GuideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) == 0 {
		t.Fatalf("expected guides, got %#v", payload)
	}
}

func TestListGuidesFiltersByHazard(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?hazard=flood", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GuideListResponse
	decodeResponse(t, response, &payload)
	for _, guide := range payload.Guides {
		if guide.HazardType != "flood" {
			t.Fatalf("expected only flood guides, got %#v", guide)
		}
	}
}

func TestListGuidesRejectsInvalidHazard(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?hazard=unknown", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
