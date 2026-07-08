package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/route-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/route-service/internal/store"
)

func newTestConfig() *config.Config {
	return &config.Config{
		Addr:                  ":8096",
		RoadClosureServiceURL: "http://closures.test",
		ShelterServiceURL:     "http://shelters.test",
		RiskServiceURL:        "http://risk.test",
		AllowedOrigins:        nil,
	}
}

func newTestServer(responses map[string]mockResponse) *Server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := newTestConfig()
	srv := NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
	srv.httpClient = &http.Client{Transport: &mockTransport{responses: responses}}
	return srv
}

type mockResponse struct {
	status int
	body   string
}

type mockTransport struct {
	responses map[string]mockResponse
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	response, ok := m.responses[path]
	if !ok {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"error":"not found"}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	}
	return &http.Response{
		StatusCode: response.status,
		Body:       io.NopCloser(strings.NewReader(response.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func TestHealthHandler(t *testing.T) {
	srv := newTestServer(nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestOptionsHandler(t *testing.T) {
	srv := newTestServer(nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/routes/options", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.OptionsResponse
	decodeResponse(t, response, &payload)
	if len(payload.WaypointTypes) != 3 {
		t.Fatalf("expected 3 waypoint types, got %#v", payload.WaypointTypes)
	}
}

func TestPlanRouteValidReturnsRoute(t *testing.T) {
	responses := map[string]mockResponse{
		"/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/risk/areas":    {status: http.StatusOK, body: `{"areas":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin: models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		Destination: &models.Coordinates{Lat: 5.6100, Lng: -0.1800},
		WaypointType: "manual",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if len(payload.Route) < 2 {
		t.Fatalf("expected route with at least origin and destination, got %#v", payload.Route)
	}
	if payload.DistanceMeters <= 0 {
		t.Fatalf("expected positive distance, got %d", payload.DistanceMeters)
	}
	if !payload.DecisionSupport {
		t.Fatalf("expected decisionSupport to be true")
	}
	if payload.Disclaimer == "" {
		t.Fatalf("expected disclaimer")
	}
}

func TestPlanRouteInvalidRequestReturns400(t *testing.T) {
	srv := newTestServer(nil)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		WaypointType: "invalid_type",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestPlanRouteMissingDestinationTargetsShelter(t *testing.T) {
	responses := map[string]mockResponse{
		"/shelters/nearby": {
			status: http.StatusOK,
			body: `{
				"shelters":[
					{"id":"shelter_001","name":"Tema Community Shelter","location":{"lat":5.6500,"lng":-0.1700},"status":"open"}
				],
				"generatedAt":"2026-07-06T12:00:00Z"
			}`,
		},
		"/road-closures": {status: http.StatusOK, body: `{"closures":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
		"/risk/areas":    {status: http.StatusOK, body: `{"areas":[],"generatedAt":"2026-07-06T12:00:00Z"}`},
	}
	srv := newTestServer(responses)

	body := models.RoutePlanRequest{
		Origin:       models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		WaypointType: "shelter",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/routes/plan", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.RoutePlanResponse
	decodeResponse(t, response, &payload)
	if payload.TargetShelter == nil {
		t.Fatalf("expected target shelter, got nil")
	}
	if payload.TargetShelter.ID != "shelter_001" {
		t.Fatalf("expected shelter_001, got %s", payload.TargetShelter.ID)
	}
	if len(payload.Route) < 2 {
		t.Fatalf("expected route with at least two waypoints, got %#v", payload.Route)
	}
	last := payload.Route[len(payload.Route)-1]
	if last.Lat != payload.TargetShelter.Location.Lat || last.Lng != payload.TargetShelter.Location.Lng {
		t.Fatalf("expected route to end at shelter, got %#v", last)
	}
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
