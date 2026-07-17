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

func TestListGuidesFillsUncoveredGroupsWithEnglish(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?language=tw&hazard=fire", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GuideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) != 1 {
		t.Fatalf("expected 1 English filler guide, got %#v", payload.Guides)
	}
	guide := payload.Guides[0]
	if guide.Language != "en" || guide.HazardType != "fire" {
		t.Fatalf("expected English fire filler, got %#v", guide)
	}
}

func TestListGuidesSkipsEnglishFillerForCoveredGroup(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?language=tw&hazard=flood&stage=before", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GuideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) != 1 {
		t.Fatalf("expected only the Twi guide, got %#v", payload.Guides)
	}
	if payload.Guides[0].Language != "tw" {
		t.Fatalf("expected the Twi guide, got %#v", payload.Guides[0])
	}
}

func TestListGuidesLanguageFallbackCatalogWide(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?language=tw", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GuideListResponse
	decodeResponse(t, response, &payload)

	// 11 seeded English guides minus the (flood, before) filler covered by the
	// Twi translation, plus the Twi guide itself.
	if len(payload.Guides) != 11 {
		t.Fatalf("expected 11 guides, got %d: %#v", len(payload.Guides), payload.Guides)
	}

	floodBefore := 0
	for _, guide := range payload.Guides {
		if guide.Language != "tw" && guide.Language != "en" {
			t.Fatalf("unexpected language %q in %#v", guide.Language, guide)
		}
		if guide.HazardType == "flood" && guide.Stage == "before" {
			floodBefore++
			if guide.Language != "tw" {
				t.Fatalf("expected no English filler for covered group, got %#v", guide)
			}
		}
	}
	if floodBefore != 1 {
		t.Fatalf("expected exactly one (flood, before) guide, got %d", floodBefore)
	}
}

func TestListGuidesFullEnglishFallbackForUntranslatedLanguage(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?language=fr", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GuideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) != 11 {
		t.Fatalf("expected all 11 English guides, got %d", len(payload.Guides))
	}
	for _, guide := range payload.Guides {
		if guide.Language != "en" {
			t.Fatalf("expected only English guides, got %#v", guide)
		}
	}
}

func TestCORSLocalhostOriginGatedByEnvironment(t *testing.T) {
	cfg := &config.Config{Addr: ":8086", AllowedOrigins: map[string]bool{"https://nadaa.gov.gh": true}}
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)

	corsHeader := func() string {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/guides", nil)
		request.Header.Set("Origin", "http://localhost:5173")
		response := httptest.NewRecorder()
		srv.Routes().ServeHTTP(response, request)
		return response.Header().Get("Access-Control-Allow-Origin")
	}

	if got := corsHeader(); got != "" {
		t.Fatalf("expected localhost origin rejected outside development, got %q", got)
	}

	t.Setenv("NADAA_ENV", "development")
	if got := corsHeader(); got != "http://localhost:5173" {
		t.Fatalf("expected localhost origin allowed in development, got %q", got)
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
