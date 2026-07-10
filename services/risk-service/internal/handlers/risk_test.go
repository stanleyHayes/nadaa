package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/utils"
)

func newTestServer() *Server {
	return NewServer(store.NewMemoryStore(), &config.Config{Addr: ":8081"})
}

func TestRiskHandlerRequiresCoordinates(t *testing.T) {
	srv := newTestServer()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/risk", nil)
	response := httptest.NewRecorder()

	srv.riskHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRiskHandlerRejectsInvalidCoordinates(t *testing.T) {
	srv := newTestServer()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/risk?lat=91&lng=-0.1870", nil)
	response := httptest.NewRecorder()

	srv.riskHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRiskHandlerReturnsSevereFloodRiskInsideZone(t *testing.T) {
	payload := requestRisk(t, "/api/v1/risk?lat=5.5600&lng=-0.2000")

	if payload.Location != "Accra Metropolitan" {
		t.Fatalf("expected Accra Metropolitan, got %q", payload.Location)
	}
	if payload.OverallRisk != "severe" {
		t.Fatalf("expected severe overall risk, got %#v", payload)
	}
	if len(payload.Risks) == 0 || payload.Risks[0].Type != "flood" || payload.Risks[0].Level != "severe" {
		t.Fatalf("expected severe flood risk first, got %#v", payload.Risks)
	}
	if len(payload.NearestShelters) < 2 {
		t.Fatalf("expected nearby shelters, got %#v", payload.NearestShelters)
	}
	if payload.NearestShelters[0].DistanceMeters > payload.NearestShelters[1].DistanceMeters {
		t.Fatalf("expected shelters sorted by distance, got %#v", payload.NearestShelters)
	}
	if len(payload.NearbyFacilities) < 3 {
		t.Fatalf("expected nearby emergency facilities, got %#v", payload.NearbyFacilities)
	}
}

func TestRiskHandlerReturnsHighFloodRiskNearZoneAndReport(t *testing.T) {
	payload := requestRisk(t, "/api/v1/risk?lat=5.6037&lng=-0.1870")

	if payload.OverallRisk != "high" {
		t.Fatalf("expected high overall risk, got %#v", payload)
	}
	if payload.Risks[0].Type != "flood" || payload.Risks[0].Level != "high" {
		t.Fatalf("expected high flood risk, got %#v", payload.Risks)
	}
	if payload.Risks[0].Probability < 0.6 || payload.Risks[0].Probability > 0.8 {
		t.Fatalf("expected high flood probability band, got %#v", payload.Risks[0])
	}
	if len(payload.RecommendedActions) == 0 {
		t.Fatalf("expected recommended actions, got %#v", payload)
	}
}

func TestRiskHandlerReturnsLowRiskOutsideFixtureZones(t *testing.T) {
	payload := requestRisk(t, "/api/v1/risk?lat=6.6885&lng=-1.6244")

	if payload.Location != "Kumasi area" {
		t.Fatalf("expected Kumasi area, got %q", payload.Location)
	}
	if payload.OverallRisk != "low" {
		t.Fatalf("expected low overall risk, got %#v", payload)
	}
	if payload.Risks[0].Type != "flood" || payload.Risks[0].Level != "low" {
		t.Fatalf("expected low flood risk, got %#v", payload.Risks)
	}
	if len(payload.NearestShelters) != 0 {
		t.Fatalf("expected no nearby fixture shelters outside Accra, got %#v", payload.NearestShelters)
	}
	if len(payload.NearbyFacilities) != 0 {
		t.Fatalf("expected no nearby fixture facilities outside Accra, got %#v", payload.NearbyFacilities)
	}
}

func TestRiskHandlerIncludesMLPredictionWhenConfigured(t *testing.T) {
	mlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ml/flood/predictions" || r.Method != http.MethodPost {
			t.Fatalf("unexpected ML request %s %s", r.Method, r.URL.Path)
		}
		utils.WriteJSON(w, http.StatusOK, models.MLPredictionResponse{
			Prediction: models.MLPredictionPayload{
				ID:                     "pred_grid-accra-central-001",
				ModelVersion:           "flood-logistic-baseline-0.1.0",
				HazardType:             "flood",
				PredictionTime:         "2026-07-06T10:00:00Z",
				TargetTime:             "2026-07-06T12:00:00Z",
				CellID:                 "grid-accra-central-001",
				Region:                 "Greater Accra",
				District:               "Accra Metropolitan",
				Community:              "Accra Central",
				Probability:            0.9993,
				Severity:               "severe",
				ExpectedOnset:          "within_24h",
				Confidence:             "medium",
				InputFeatureSetVersion: "flood-risk-features.v1",
				HumanReviewRequired:    true,
				AutoPublishAllowed:     false,
				Source:                 "baseline_fixture_model",
			},
			Log: models.MLPredictionLog{
				ID:                     "ml_log_test",
				ModelVersion:           "flood-logistic-baseline-0.1.0",
				InputFeatureSetVersion: "flood-risk-features.v1",
			},
		})
	}))
	defer mlServer.Close()

	srv := newTestServer()
	srv.mlClient = newMLClient(mlServer.URL+"/api/v1", mlServer.Client())

	payload := requestRiskFromServer(t, srv, "/api/v1/risk?lat=5.5600&lng=-0.2000")
	if payload.MLPrediction == nil {
		t.Fatalf("expected ML prediction in risk response, got %#v", payload)
	}
	if payload.MLPrediction.ModelVersion != "flood-logistic-baseline-0.1.0" {
		t.Fatalf("expected model version in risk response, got %#v", payload.MLPrediction)
	}
	if !payload.MLPrediction.HumanReviewRequired || payload.MLPrediction.AutoPublishAllowed {
		t.Fatalf("expected human review and no auto-publish, got %#v", payload.MLPrediction)
	}
	if payload.MLPrediction.PredictionLogID != "ml_log_test" {
		t.Fatalf("expected prediction log id, got %#v", payload.MLPrediction)
	}
}

func requestRisk(t *testing.T, path string) models.RiskResponse {
	t.Helper()

	return requestRiskFromServer(t, newTestServer(), path)
}

func requestRiskFromServer(t *testing.T, srv *Server, path string) models.RiskResponse {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()

	srv.riskHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.RiskResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}
