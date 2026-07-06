package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListGuidesByHazardStageAndLanguage(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?hazard=flood&stage=before&language=en", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload guideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) == 0 {
		t.Fatal("expected at least one guide")
	}

	for _, guide := range payload.Guides {
		if guide.HazardType != "flood" || guide.Stage != "before" || guide.Language != "en" {
			t.Fatalf("unexpected guide returned: %#v", guide)
		}
	}
}

func TestGuideContentCoversInitialTopics(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?language=en", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload guideListResponse
	decodeResponse(t, response, &payload)

	expectedHazards := map[string]bool{
		"flood":             false,
		"fire":              false,
		"road_crash":        false,
		"electrical_hazard": false,
		"disease_outbreak":  false,
		"other":             false,
	}
	expectedTitles := map[string]bool{
		"Safe evacuation":         false,
		"Emergency bag checklist": false,
		"Family emergency plan":   false,
		"Calling 112":             false,
	}

	for _, guide := range payload.Guides {
		if _, ok := expectedHazards[guide.HazardType]; ok {
			expectedHazards[guide.HazardType] = true
		}
		if _, ok := expectedTitles[guide.Title]; ok {
			expectedTitles[guide.Title] = true
		}
	}

	for hazard, found := range expectedHazards {
		if !found {
			t.Fatalf("expected hazard coverage for %s", hazard)
		}
	}
	for title, found := range expectedTitles {
		if !found {
			t.Fatalf("expected guide title %q", title)
		}
	}
}

func TestGuideLanguageFallbackUsesEnglish(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?hazard=fire&stage=during&language=ga", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload guideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) != 1 {
		t.Fatalf("expected one fallback guide, got %#v", payload.Guides)
	}
	if payload.Guides[0].Language != "en" {
		t.Fatalf("expected English fallback, got %#v", payload.Guides[0])
	}
}

func TestGuideExactLanguageMatch(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?hazard=flood&stage=before&language=tw", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload guideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) != 1 {
		t.Fatalf("expected one Twi guide, got %#v", payload.Guides)
	}
	if payload.Guides[0].Language != "tw" {
		t.Fatalf("expected Twi guide, got %#v", payload.Guides[0])
	}
}

func TestGuideOfflineFilter(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?offline=true", nil)

	srv.listGuidesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload guideListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Guides) == 0 {
		t.Fatal("expected offline guides")
	}
	for _, guide := range payload.Guides {
		if !guide.OfflineAvailable {
			t.Fatalf("expected only offline guides, got %#v", guide)
		}
	}
}

func TestListGuidesRejectsInvalidFilter(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/guides?stage=panic", nil)

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
