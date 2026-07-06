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

func TestCreateWeatherHydrologyImportJobStoresObservations(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	body := bytes.NewBufferString(`{
		"adapterId": "mock-weather-hydrology-adapter",
		"metric": "rainfall_mm",
		"requestedBy": "test-runner",
		"correlationId": "corr_import_001"
	}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/weather-hydrology/import-jobs", body)

	srv.createObservationImportJobHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var job observationImportJob
	decodeResponse(t, response, &job)
	if job.Status != "succeeded" || job.ImportedCount == 0 || job.Trigger != "manual" {
		t.Fatalf("unexpected import job: %#v", job)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/weather-hydrology/observations?metric=rainfall_mm", nil)
	srv.listImportedObservationsHandler(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResponse.Code)
	}

	var payload importedObservationListResponse
	decodeResponse(t, listResponse, &payload)
	if len(payload.Observations) != job.ImportedCount {
		t.Fatalf("expected %d imported observations, got %d", job.ImportedCount, len(payload.Observations))
	}
	for _, observation := range payload.Observations {
		if observation.ImportJobID != job.ID || observation.Source == "" || observation.ValidFrom.IsZero() || observation.ValidTo.IsZero() {
			t.Fatalf("imported observation missing source, validity, or job linkage: %#v", observation)
		}
		if observation.RainfallMM == nil || observation.StorageTarget != "weather_observations" {
			t.Fatalf("imported rainfall observation missing normalized storage fields: %#v", observation)
		}
	}
}

func TestWeatherHydrologyImportFailureIsLoggedAndRetryable(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	body := bytes.NewBufferString(`{
		"metric": "water_level_m",
		"simulateFailure": true,
		"failureMessage": "temporary partner timeout",
		"correlationId": "corr_import_failure"
	}`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/weather-hydrology/import-jobs", body)

	srv.createObservationImportJobHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}

	var failedJob observationImportJob
	decodeResponse(t, response, &failedJob)
	if failedJob.Status != "failed" || !failedJob.Retryable || failedJob.NextRetryAt == nil || failedJob.Error == "" {
		t.Fatalf("expected retryable failed import job, got %#v", failedJob)
	}

	logResponse := httptest.NewRecorder()
	logRequest := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/weather-hydrology/import-jobs?status=failed", nil)
	srv.listObservationImportJobsHandler(logResponse, logRequest)
	if logResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, logResponse.Code)
	}
	var logs observationImportJobListResponse
	decodeResponse(t, logResponse, &logs)
	if len(logs.Jobs) != 1 || logs.Jobs[0].ID != failedJob.ID {
		t.Fatalf("expected failed job in import log, got %#v", logs.Jobs)
	}

	retryResponse := httptest.NewRecorder()
	retryRequest := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/weather-hydrology/import-jobs/"+failedJob.ID+"/retry", nil)
	retryRequest.SetPathValue("id", failedJob.ID)
	srv.retryObservationImportJobHandler(retryResponse, retryRequest)
	if retryResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, retryResponse.Code, retryResponse.Body.String())
	}
	var retryJob observationImportJob
	decodeResponse(t, retryResponse, &retryJob)
	if retryJob.Status != "succeeded" || retryJob.Trigger != "retry" || retryJob.Attempts != 2 || retryJob.ImportedCount == 0 {
		t.Fatalf("unexpected retry job: %#v", retryJob)
	}
}

func TestWeatherHydrologyImportRejectsInvalidMetric(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	body := bytes.NewBufferString(`{ "metric": "wind_speed" }`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/weather-hydrology/import-jobs", body)

	srv.createObservationImportJobHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
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
