package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{store: newMemoryStore(now), now: func() time.Time { return now }}
}

func TestNearbySheltersReturnsSortedSheltersAndRecoverySupport(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/shelters/nearby?lat=5.5600&lng=-0.2000", nil)

	srv.nearbySheltersHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload nearbyShelterResponse
	decodeResponse(t, response, &payload)
	if len(payload.Shelters) < 2 {
		t.Fatalf("expected nearby shelters, got %#v", payload.Shelters)
	}
	if payload.Shelters[0].ID != "00000000-0000-0000-0000-000000000301" || payload.Shelters[0].DistanceMeters > payload.Shelters[1].DistanceMeters {
		t.Fatalf("expected closest shelter first, got %#v", payload.Shelters)
	}
	if len(payload.RecoverySupport) == 0 || payload.RecoverySupport[0].DistanceMeters <= 0 {
		t.Fatalf("expected nearby recovery support with distances, got %#v", payload.RecoverySupport)
	}
}

func TestNearbySheltersRejectsInvalidCoordinates(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/shelters/nearby?lat=91&lng=-0.2000", nil)

	srv.nearbySheltersHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRecoverySupportNearby(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/recovery-support/nearby?lat=5.5600&lng=-0.2000", nil)

	srv.nearbyRecoverySupportHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload recoverySupportResponse
	decodeResponse(t, response, &payload)
	if len(payload.RecoverySupport) == 0 {
		t.Fatalf("expected recovery support locations, got %#v", payload)
	}
}

func TestUpdateShelterOccupancyRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(occupancyUpdateRequest{}))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestUpdateShelterOccupancy(t *testing.T) {
	srv := newTestServer()
	occupancy := 450
	body := occupancyUpdateRequest{CurrentOccupancy: &occupancy, Notes: "Shelter reached capacity during flood response."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(body))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload shelterUpdateResponse
	decodeResponse(t, response, &payload)
	if payload.Shelter.CurrentOccupancy != 450 || payload.Shelter.Status != "full" || payload.Shelter.UpdatedBy != "usr_shelter_operator" {
		t.Fatalf("expected full updated shelter, got %#v", payload.Shelter)
	}
}

func TestUpdateShelterOccupancyRejectsOverCapacity(t *testing.T) {
	srv := newTestServer()
	capacity := 10
	occupancy := 11
	body := occupancyUpdateRequest{Capacity: &capacity, CurrentOccupancy: &occupancy}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(body))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func authorityRequest(method string, target string, body *bytes.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_shelter_operator")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-shelter-update")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
