package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/store"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return NewServer(store.NewMemoryStore(now), nil, "", false)
}

func TestListContractsCoversRequiredPartners(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts", nil)

	srv.listContractsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.ContractListResponse
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
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts?domain=weather&direction=inbound", nil)

	srv.listContractsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.ContractListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Contracts) != 1 {
		t.Fatalf("expected one weather contract, got %#v", payload.Contracts)
	}
	if payload.Contracts[0].Domain != "weather" || payload.Contracts[0].Direction != "inbound" {
		t.Fatalf("unexpected contract returned: %#v", payload.Contracts[0])
	}
}

func TestListContractsRejectsInvalidFilter(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts?domain=aliens", nil)

	srv.listContractsHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestListMockWeatherHydrologyObservations(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/mock/weather-hydrology/observations?metric=rainfall_mm", nil)

	srv.listObservationsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.ObservationListResponse
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
	srv := newTestServer()
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

	var job models.ObservationImportJob
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

	var payload models.ImportedObservationListResponse
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
	srv := newTestServer()
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

	var failedJob models.ObservationImportJob
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
	var logs models.ObservationImportJobListResponse
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
	var retryJob models.ObservationImportJob
	decodeResponse(t, retryResponse, &retryJob)
	if retryJob.Status != "succeeded" || retryJob.Trigger != "retry" || retryJob.Attempts != 2 || retryJob.ImportedCount == 0 {
		t.Fatalf("unexpected retry job: %#v", retryJob)
	}
}

func TestWeatherHydrologyImportRejectsInvalidMetric(t *testing.T) {
	srv := newTestServer()
	body := bytes.NewBufferString(`{ "metric": "wind_speed" }`)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/weather-hydrology/import-jobs", body)

	srv.createObservationImportJobHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateMockSyncEvent(t *testing.T) {
	srv := newTestServer()
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

	var payload models.SyncEvent
	decodeResponse(t, response, &payload)
	if payload.Status != "accepted" || payload.AdapterID != "mock-incident-sync-adapter" {
		t.Fatalf("unexpected sync event: %#v", payload)
	}
}

func TestCreateMockSyncEventRejectsMissingCorrelationID(t *testing.T) {
	srv := newTestServer()
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

func TestCreateMockSyncEventDedupesOnCorrelationID(t *testing.T) {
	srv := newTestServer()
	newBody := func(correlationID string) *bytes.Buffer {
		return bytes.NewBufferString(`{
			"type": "incident",
			"sourceId": "inc_001",
			"reference": "INC-000001",
			"hazardType": "flood",
			"status": "verified",
			"severity": "high",
			"targetAgencyIds": ["00000000-0000-0000-0000-000000000101"],
			"correlationId": "` + correlationID + `"
		}`)
	}

	firstResponse := httptest.NewRecorder()
	srv.createSyncEventHandler(firstResponse, httptest.NewRequest(http.MethodPost, "/api/v1/integrations/mock/sync-events", newBody("corr_dedupe_001")))
	if firstResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, firstResponse.Code, firstResponse.Body.String())
	}
	var first models.SyncEvent
	decodeResponse(t, firstResponse, &first)

	replayResponse := httptest.NewRecorder()
	srv.createSyncEventHandler(replayResponse, httptest.NewRequest(http.MethodPost, "/api/v1/integrations/mock/sync-events", newBody("corr_dedupe_001")))
	if replayResponse.Code != http.StatusOK {
		t.Fatalf("expected idempotent replay status %d, got %d", http.StatusOK, replayResponse.Code)
	}
	var replay models.SyncEvent
	decodeResponse(t, replayResponse, &replay)
	if replay.ID != first.ID {
		t.Fatalf("expected replay to return the original event %s, got %s", first.ID, replay.ID)
	}

	otherResponse := httptest.NewRecorder()
	srv.createSyncEventHandler(otherResponse, httptest.NewRequest(http.MethodPost, "/api/v1/integrations/mock/sync-events", newBody("corr_dedupe_002")))
	if otherResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, otherResponse.Code)
	}
	var other models.SyncEvent
	decodeResponse(t, otherResponse, &other)
	if other.ID == first.ID {
		t.Fatalf("expected unique sync event IDs, got %s twice", other.ID)
	}
}

func roadClosureImportBody(geometry string) *bytes.Buffer {
	return bytes.NewBufferString(`{
		"source": "police",
		"roadName": "N1 Highway",
		"status": "active",
		"geometry": "` + geometry + `",
		"validFrom": "2026-07-17T10:00:00Z"
	}`)
}

func TestImportRoadClosureRejectsInvalidGeometryLocally(t *testing.T) {
	srv := newTestServer()
	for _, geometry := range []string{"POINT(-0.187 5.6037)", "LINESTRING(-0.187 5.6037)", "LINESTRING(-0.187 5.6037, -200 95.0)", "LINESTRING(nope)"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/road-closures/imports", roadClosureImportBody(geometry))

		srv.importRoadClosureHandler(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d for geometry %q, got %d", http.StatusBadRequest, geometry, response.Code)
		}
		var payload models.APIError
		decodeResponse(t, response, &payload)
		if payload.Error.Code != "invalid_geometry" {
			t.Fatalf("expected invalid_geometry for %q, got %q", geometry, payload.Error.Code)
		}
	}
}

func TestImportRoadClosureAccepted(t *testing.T) {
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer nadaa.test.sig" {
			t.Errorf("expected Authorization header forwarded, got %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("X-NADAA-Actor-ID") != "" {
			t.Errorf("expected no X-NADAA-Actor-ID when a bearer token is present, got %q", r.Header.Get("X-NADAA-Actor-ID"))
		}
		if r.Header.Get("X-NADAA-Request-ID") != "req_001" {
			t.Errorf("expected X-NADAA-Request-ID passthrough, got %q", r.Header.Get("X-NADAA-Request-ID"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer downstream.Close()

	srv := NewServer(store.NewMemoryStore(time.Now().UTC()), downstream.Client(), downstream.URL, false)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/road-closures/imports", roadClosureImportBody("LINESTRING(-0.187 5.6037, -0.20 5.61)"))
	request.Header.Set("Authorization", "Bearer nadaa.test.sig")
	request.Header.Set("X-NADAA-Actor-ID", "forged-actor")
	request.Header.Set("X-NADAA-Request-ID", "req_001")

	srv.importRoadClosureHandler(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
	}
	var payload models.RoadClosureImportResponse
	decodeResponse(t, response, &payload)
	if payload.Imported != 1 || payload.Record.ID == "" {
		t.Fatalf("unexpected import response: %#v", payload)
	}
}

func TestImportRoadClosureForwardsMockActorHeadersOnlyWhenEnabled(t *testing.T) {
	newDownstream := func(t *testing.T, expectHeaders bool) *httptest.Server {
		t.Helper()
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectHeaders && r.Header.Get("X-NADAA-Actor-ID") != "actor_001" {
				t.Errorf("expected X-NADAA-Actor-ID forwarded with mock actors enabled, got %q", r.Header.Get("X-NADAA-Actor-ID"))
			}
			if !expectHeaders && r.Header.Get("X-NADAA-Actor-ID") != "" {
				t.Errorf("expected no X-NADAA-Actor-ID with mock actors disabled, got %q", r.Header.Get("X-NADAA-Actor-ID"))
			}
			w.WriteHeader(http.StatusOK)
		}))
	}

	for _, tc := range []struct {
		name         string
		allowMock    bool
		expectHeader bool
	}{
		{"mock actors enabled", true, true},
		{"mock actors disabled", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			downstream := newDownstream(t, tc.expectHeader)
			defer downstream.Close()

			srv := NewServer(store.NewMemoryStore(time.Now().UTC()), downstream.Client(), downstream.URL, tc.allowMock)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/road-closures/imports", roadClosureImportBody("LINESTRING(-0.187 5.6037, -0.20 5.61)"))
			request.Header.Set("X-NADAA-Actor-ID", "actor_001")
			request.Header.Set("X-NADAA-Actor-Role", "dispatcher")
			request.Header.Set("X-NADAA-Agency-ID", "agency_001")
			request.Header.Set("X-NADAA-MFA-Completed", "true")

			srv.importRoadClosureHandler(response, request)

			if response.Code != http.StatusAccepted {
				t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, response.Code, response.Body.String())
			}
		})
	}
}

func TestImportRoadClosureMapsDownstreamClientErrors(t *testing.T) {
	for _, tc := range []struct {
		name             string
		downstreamStatus int
		downstreamCode   string
	}{
		{"validation error", http.StatusBadRequest, "invalid_road_name"},
		{"missing authority", http.StatusUnauthorized, "missing_authority_context"},
		{"forbidden role", http.StatusForbidden, "forbidden"},
		{"not found", http.StatusNotFound, "not_found"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.downstreamStatus)
				_, _ = w.Write([]byte(`{"error":{"code":"` + tc.downstreamCode + `","message":"downstream said no"}}`))
			}))
			defer downstream.Close()

			srv := NewServer(store.NewMemoryStore(time.Now().UTC()), downstream.Client(), downstream.URL, false)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/road-closures/imports", roadClosureImportBody("LINESTRING(-0.187 5.6037, -0.20 5.61)"))

			srv.importRoadClosureHandler(response, request)

			if response.Code != tc.downstreamStatus {
				t.Fatalf("expected status %d, got %d: %s", tc.downstreamStatus, response.Code, response.Body.String())
			}
			var payload models.APIError
			decodeResponse(t, response, &payload)
			if payload.Error.Code != tc.downstreamCode {
				t.Fatalf("expected downstream code %q propagated, got %q", tc.downstreamCode, payload.Error.Code)
			}
		})
	}
}

func TestImportRoadClosureReturns502OnDownstreamFailure(t *testing.T) {
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer downstream.Close()

	srv := NewServer(store.NewMemoryStore(time.Now().UTC()), downstream.Client(), downstream.URL, false)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/integrations/road-closures/imports", roadClosureImportBody("LINESTRING(-0.187 5.6037, -0.20 5.61)"))

	srv.importRoadClosureHandler(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, response.Code)
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "road_closure_service_unavailable" {
		t.Fatalf("expected road_closure_service_unavailable, got %q", payload.Error.Code)
	}
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
