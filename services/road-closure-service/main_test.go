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

func TestHealthHandler(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestListRoadClosuresDefaultsToActive(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload roadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 2 {
		t.Fatalf("expected 2 active closures, got %d", len(payload.Closures))
	}
}

func TestListRoadClosuresFiltersByStatus(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?status=scheduled", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload roadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 1 || payload.Closures[0].Status != "scheduled" {
		t.Fatalf("expected 1 scheduled closure, got %#v", payload.Closures)
	}
}

func TestListRoadClosuresNearby(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?lat=5.570&lng=-0.200", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload roadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) == 0 {
		t.Fatalf("expected nearby closures, got %#v", payload.Closures)
	}
	if payload.Closures[0].DistanceMeters <= 0 {
		t.Fatalf("expected distance populated, got %#v", payload.Closures[0])
	}
}

func TestListRoadClosuresByBBox(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?bbox=-0.30,5.50,-0.15,5.60", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload roadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 2 {
		t.Fatalf("expected 2 closures in bbox, got %d", len(payload.Closures))
	}
}

func TestCreateRoadClosureRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(createRoadClosureRequest{RoadName: "Test Road"}))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateRoadClosure(t *testing.T) {
	srv := newTestServer()
	body := createRoadClosureRequest{
		RoadName: "Ring Road Central",
		Reason:   "Flooding",
		Status:   "active",
		Severity: "high",
		Geometry: lineStringGeometry{
			Type: "LineString",
			Coordinates: [][]float64{
				{-0.210, 5.550},
				{-0.200, 5.552},
			},
		},
		DetourNote: "Use Independence Avenue",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload roadClosureResponse
	decodeResponse(t, response, &payload)
	if payload.Closure.RoadName != "Ring Road Central" || payload.Closure.Status != "active" || payload.Closure.CreatedBy != "usr_road_closure_officer" {
		t.Fatalf("expected created closure, got %#v", payload.Closure)
	}
}

func TestCreateRoadClosureRejectsInvalidGeometry(t *testing.T) {
	srv := newTestServer()
	body := createRoadClosureRequest{
		RoadName: "Bad Road",
		Status:   "active",
		Geometry: lineStringGeometry{
			Type: "Polygon",
			Coordinates: [][]float64{
				{-0.210, 5.550},
			},
		},
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestUpdateRoadClosure(t *testing.T) {
	srv := newTestServer()
	body := updateRoadClosureRequest{
		Status: "lifted",
		Reason: "Water receded",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/road-closures/road_closure_001", jsonBody(body))
	request.SetPathValue("id", "road_closure_001")

	srv.updateRoadClosureHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload roadClosureResponse
	decodeResponse(t, response, &payload)
	if payload.Closure.Status != "lifted" || payload.Closure.UpdatedBy != "usr_road_closure_officer" {
		t.Fatalf("expected lifted closure, got %#v", payload.Closure)
	}
}

func TestUpdateRoadClosureNotFound(t *testing.T) {
	srv := newTestServer()
	body := updateRoadClosureRequest{Status: "lifted"}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/road-closures/missing", jsonBody(body))
	request.SetPathValue("id", "missing")

	srv.updateRoadClosureHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestImportAdapter(t *testing.T) {
	srv := newTestServer()
	body := adapterImportRequest{
		Source:    "ghana-police",
		SourceRef: "police-road-closure-feed",
		RoadName:  "Sample Market Road",
		Status:    "active",
		Reason:    "Flooding",
		Geometry:  "LINESTRING(-0.20 5.56, -0.19 5.57)",
		ValidFrom: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC),
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures/imports/adapter", jsonBody(body))

	srv.importAdapterHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload adapterImportResponse
	decodeResponse(t, response, &payload)
	if payload.Imported != 1 || payload.Source != "ghana-police" {
		t.Fatalf("expected one imported closure from ghana-police, got %#v", payload)
	}
	if payload.Closures[0].Geometry.Type != "LineString" || len(payload.Closures[0].Geometry.Coordinates) != 2 {
		t.Fatalf("expected parsed LineString geometry, got %#v", payload.Closures[0].Geometry)
	}
}

func TestImportAdapterRejectsInvalidWKT(t *testing.T) {
	srv := newTestServer()
	body := adapterImportRequest{
		Source:    "ghana-police",
		RoadName:  "Bad Road",
		Status:    "active",
		Geometry:  "POINT(0 0)",
		ValidFrom: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC),
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures/imports/adapter", jsonBody(body))

	srv.importAdapterHandler(response, request)

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
	request.Header.Set("X-NADAA-Actor-ID", "usr_road_closure_officer")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-road-closure")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
