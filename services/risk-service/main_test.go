package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRiskHandlerRequiresCoordinates(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/risk", nil)
	response := httptest.NewRecorder()

	riskHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRiskHandlerReturnsFloodRisk(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/risk?lat=5.6037&lng=-0.1870", nil)
	response := httptest.NewRecorder()

	riskHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload riskResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.Location != "Accra Central" {
		t.Fatalf("expected Accra Central, got %q", payload.Location)
	}

	if len(payload.Risks) == 0 || payload.Risks[0].Type != "flood" {
		t.Fatalf("expected first risk to be flood, got %#v", payload.Risks)
	}
}

