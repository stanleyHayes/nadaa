package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListContractsCoversRequiredPartners(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts", nil)

	srv.listContractsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload contractListResponse
	decodeResponse(t, response, &payload)

	expectedPartners := map[string]bool{
		"Ghana Meteorological Agency":     false,
		"Ghana Hydrological Authority":    false,
		"NADMO National Operations":       false,
		"Ghana Police Service":            false,
		"Ghana National Fire Service":     false,
		"National Ambulance Service":      false,
		"District Assemblies":             false,
		"Hospitals And Health Facilities": false,
		"Utilities And Power Providers":   false,
	}

	for _, contract := range payload.Contracts {
		if _, ok := expectedPartners[contract.Partner]; ok {
			expectedPartners[contract.Partner] = true
		}
		if contract.DataOwner == "" || contract.Cadence == "" || len(contract.Payloads) == 0 {
			t.Fatalf("contract is missing ownership, cadence, or payloads: %#v", contract)
		}
		if contract.Authentication.Mode == "" || contract.FailureBehavior.MaxAttempts == 0 {
			t.Fatalf("contract is missing auth or failure behavior: %#v", contract)
		}
	}

	for partner, found := range expectedPartners {
		if !found {
			t.Fatalf("expected contract coverage for %s", partner)
		}
	}
}

func TestListContractsFiltersByDomainAndDirection(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts?domain=weather&direction=inbound", nil)

	srv.listContractsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload contractListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Contracts) != 1 {
		t.Fatalf("expected one weather contract, got %#v", payload.Contracts)
	}
	if payload.Contracts[0].Domain != "weather" || payload.Contracts[0].Direction != "inbound" {
		t.Fatalf("unexpected contract returned: %#v", payload.Contracts[0])
	}
}

func TestListContractsRejectsInvalidFilter(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts?domain=aliens", nil)

	srv.listContractsHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestListMockWeatherHydrologyObservations(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/mock/weather-hydrology/observations?metric=rainfall_mm", nil)

	srv.listObservationsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload observationListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Observations) == 0 {
		t.Fatal("expected rainfall observations")
	}
	for _, observation := range payload.Observations {
		if observation.Metric != "rainfall_mm" || observation.GeneratedBy != "mock_adapter" {
			t.Fatalf("unexpected observation: %#v", observation)
		}
	}
}

func TestCreateMockSyncEvent(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	body := bytes.NewBufferString(`{
		"type": "incident",
		"sourceId": "inc_001",
		"reference": "INC-000001",
		"hazardType": "flood",
		"status": "verified",
		"severity": "high",
		"summary": "Flooded road near market",
		"location": { "lat": 5.6037, "lng": -0.187 },
		"targetAgencyIds": ["00000000-0000-0000-0000-000000000101"],
		"correlationId": "corr_001"
	}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/mock/sync-events", body)

	srv.createSyncEventHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var payload syncEvent
	decodeResponse(t, response, &payload)
	if payload.Status != "accepted" || payload.AdapterID != "mock-incident-sync-adapter" {
		t.Fatalf("unexpected sync event: %#v", payload)
	}
}

func TestCreateMockSyncEventRejectsMissingCorrelationID(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	body := bytes.NewBufferString(`{
		"type": "alert",
		"sourceId": "alert_001",
		"reference": "ALT-000001",
		"hazardType": "flood",
		"severity": "warning",
		"targetAgencyIds": ["00000000-0000-0000-0000-000000000101"]
	}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/mock/sync-events", body)

	srv.createSyncEventHandler(response, request)

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
