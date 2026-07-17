package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/store"
)

func TestPredictionHandlerReturnsFixtureModelPrediction(t *testing.T) {
	srv := newTestServer(t)
	response := postPrediction(t, srv, models.PredictionRequest{
		Location:      models.Coordinates{Lat: 5.5600, Lng: -0.2000},
		RequestedBy:   "risk-service",
		CorrelationID: "risk-test-001",
	})

	if response.Prediction.ModelVersion != "flood-logistic-baseline-0.1.0" {
		t.Fatalf("expected baseline model version, got %#v", response.Prediction)
	}
	if response.Prediction.Severity != "severe" {
		t.Fatalf("expected severe Accra Central prediction, got %#v", response.Prediction)
	}
	if response.Prediction.Confidence != "medium" {
		t.Fatalf("expected fixture confidence cap, got %#v", response.Prediction)
	}
	if !response.Prediction.HumanReviewRequired || response.Prediction.AutoPublishAllowed {
		t.Fatalf("expected human review and no auto-publish, got %#v", response.Prediction)
	}
	if response.Log.ModelVersion != response.Prediction.ModelVersion || response.Log.StorageTarget != "ml_predictions" {
		t.Fatalf("expected prediction log with model version and storage target, got %#v", response.Log)
	}
	if response.Safety.AutoPublishAllowed || !response.Safety.HumanReviewRequired {
		t.Fatalf("expected safety policy in response, got %#v", response.Safety)
	}
}

func TestPredictionHandlerRejectsInvalidCoordinates(t *testing.T) {
	srv := newTestServer(t)
	requestBody := []byte(`{"location":{"lat":91,"lng":-0.2}}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ml/flood/predictions", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()

	srv.createFloodPredictionHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestPredictionLogsListStoredPredictions(t *testing.T) {
	srv := newTestServer(t)
	postPrediction(t, srv, models.PredictionRequest{Location: models.Coordinates{Lat: 5.6037, Lng: -0.1870}})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/ml/prediction-logs", nil)
	response := httptest.NewRecorder()
	srv.listPredictionLogsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.PredictionLogListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode logs: %v", err)
	}
	if len(payload.Logs) != 1 {
		t.Fatalf("expected one prediction log, got %#v", payload.Logs)
	}
	if payload.Logs[0].ModelVersion != "flood-logistic-baseline-0.1.0" {
		t.Fatalf("expected logged model version, got %#v", payload.Logs[0])
	}
}

func newTestServer(t *testing.T) *server {
	t.Helper()
	cfg := &config.Config{Addr: ":8094"}
	s, err := store.NewMemoryStore("../../../../data/flood-risk/models")
	if err != nil {
		t.Fatalf("new memory store: %v", err)
	}
	return NewServer(s, func() time.Time { return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC) }, cfg)
}

func postPrediction(t *testing.T, srv *server, payload models.PredictionRequest) models.PredictionResponse {
	t.Helper()
	requestBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/ml/flood/predictions", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	srv.createFloodPredictionHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var result models.PredictionResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return result
}

func TestPredictionLogIDsUniqueWithinSameSecond(t *testing.T) {
	srv := newTestServer(t)
	// The test clock is fixed, so both requests hit the same cell in the same second.
	first := postPrediction(t, srv, models.PredictionRequest{Location: models.Coordinates{Lat: 5.5600, Lng: -0.2000}})
	second := postPrediction(t, srv, models.PredictionRequest{Location: models.Coordinates{Lat: 5.5600, Lng: -0.2000}})

	if first.Log.ID == "" || second.Log.ID == "" {
		t.Fatal("expected prediction log IDs")
	}
	if first.Log.ID == second.Log.ID {
		t.Fatalf("expected unique prediction log IDs, both were %s", first.Log.ID)
	}
}
