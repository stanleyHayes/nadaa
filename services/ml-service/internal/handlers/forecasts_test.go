package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
)

func TestListForecasts(t *testing.T) {
	srv := newTestServer(t)

	req := authedRequest(http.MethodGet, "/api/v1/forecasts", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.ForecastListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Forecasts) == 0 {
		t.Fatal("expected forecasts")
	}
	// Deterministic ordering: forecasts are sorted by predicted demand desc.
	for i := 1; i < len(resp.Forecasts); i++ {
		if resp.Forecasts[i-1].PredictedIncidentCount < resp.Forecasts[i].PredictedIncidentCount {
			t.Fatalf("forecasts not sorted by demand desc: %+v", resp.Forecasts)
		}
	}
	top := resp.Forecasts[0]
	if top.HazardType != "flood" || top.PredictedIncidentCount <= 0 {
		t.Fatalf("unexpected top forecast: %+v", top)
	}
	if top.Confidence == "" || top.RiskLevel == "" || len(top.Factors) == 0 {
		t.Fatalf("forecast missing explainability fields: %+v", top)
	}
	if top.TimeWindowStart == "" || top.TimeWindowEnd == "" {
		t.Fatalf("forecast missing time window: %+v", top)
	}
}

func TestListForecastsDeterministic(t *testing.T) {
	srv := newTestServer(t)

	first := forecastBody(t, srv, "/api/v1/forecasts")
	second := forecastBody(t, srv, "/api/v1/forecasts")
	if !bytes.Equal(first, second) {
		t.Fatal("forecast output is not deterministic across identical requests")
	}
}

func TestForecastsRegionFilter(t *testing.T) {
	srv := newTestServer(t)

	req := authedRequest(http.MethodGet, "/api/v1/forecasts?region=Ashanti", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var resp models.ForecastListResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	for _, f := range resp.Forecasts {
		if f.Region != "Ashanti" {
			t.Fatalf("region filter leaked non-Ashanti forecast: %+v", f)
		}
	}
}

func TestForecastByRegion(t *testing.T) {
	srv := newTestServer(t)

	req := authedRequest(http.MethodGet, "/api/v1/forecasts/Greater%20Accra", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.ForecastDetailResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Forecast.Region != "Greater Accra" || len(resp.Forecasts) == 0 {
		t.Fatalf("unexpected detail response: %+v", resp)
	}
}

func TestForecastByRegionNotFound(t *testing.T) {
	srv := newTestServer(t)

	req := authedRequest(http.MethodGet, "/api/v1/forecasts/Nowhere", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", rr.Code)
	}
}

func TestStagingSuggestions(t *testing.T) {
	srv := newTestServer(t)

	req := authedRequest(http.MethodGet, "/api/v1/staging-suggestions", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.StagingSuggestionListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Suggestions) == 0 {
		t.Fatal("expected staging suggestions")
	}
	for _, s := range resp.Suggestions {
		if s.RecommendedUnits < 1 || s.RadiusMeters <= 0 {
			t.Fatalf("invalid staging units/radius: %+v", s)
		}
		if len(s.OperationalConstraints) == 0 {
			t.Fatalf("staging suggestion missing operational constraints: %+v", s)
		}
		if s.Location.Lat == 0 || s.Location.Lng == 0 {
			t.Fatalf("staging suggestion missing location: %+v", s)
		}
	}
}

func TestStagingSuggestionsAgencyFilter(t *testing.T) {
	srv := newTestServer(t)

	req := authedRequest(http.MethodGet, "/api/v1/staging-suggestions?agencyType=ambulance", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	var resp models.StagingSuggestionListResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Suggestions) == 0 {
		t.Fatal("expected ambulance suggestions")
	}
	for _, s := range resp.Suggestions {
		if s.AgencyType != "ambulance" {
			t.Fatalf("agency filter leaked %s", s.AgencyType)
		}
	}
}

func TestCompareScenarios(t *testing.T) {
	srv := newTestServer(t)

	payload := models.CompareScenarioRequest{HistoricalWeight: 2.0, TimeWindowHours: 48}
	body, _ := json.Marshal(payload)
	req := authedRequest(http.MethodPost, "/api/v1/forecasts/compare", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.CompareScenarioResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Scenarios) != 2 {
		t.Fatalf("expected 2 scenarios got %d", len(resp.Scenarios))
	}
	if resp.Scenarios[0].Name != "Current conditions" || resp.Scenarios[1].Name != "Adjusted scenario" {
		t.Fatalf("unexpected scenario names: %+v", resp.Scenarios)
	}
	// A higher historical weight and longer window must not reduce total demand.
	if resp.Scenarios[1].Summary.TotalPredictedIncidents < resp.Scenarios[0].Summary.TotalPredictedIncidents {
		t.Fatalf("adjusted scenario demand should be >= baseline: %+v", resp.Scenarios)
	}
}

func TestCompareScenariosRiskFilterIsSymmetric(t *testing.T) {
	srv := newTestServer(t)

	// riskLevel scopes BOTH scenarios, so a filtered comparison must remain
	// consistent: the baseline and adjusted cover the same districts and the
	// adjusted (heavier history) total is never below the baseline.
	payload := models.CompareScenarioRequest{RiskLevel: "severe", HistoricalWeight: 2.0}
	body, _ := json.Marshal(payload)
	req := authedRequest(http.MethodPost, "/api/v1/forecasts/compare", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.CompareScenarioResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	base := resp.Scenarios[0]
	adjusted := resp.Scenarios[1]
	if len(base.Forecasts) != len(adjusted.Forecasts) {
		t.Fatalf("filtered scenarios must cover the same districts: base %d adjusted %d",
			len(base.Forecasts), len(adjusted.Forecasts))
	}
	if len(base.Forecasts) == 0 {
		t.Fatal("severe threshold should include at least one seeded district")
	}
	for _, f := range adjusted.Forecasts {
		if f.RiskLevel != "severe" && f.RiskLevel != "emergency" {
			t.Fatalf("riskLevel threshold leaked a below-severe district: %s", f.RiskLevel)
		}
	}
	if adjusted.Summary.TotalPredictedIncidents < base.Summary.TotalPredictedIncidents {
		t.Fatalf("adjusted total %d must be >= baseline %d under a shared filter",
			adjusted.Summary.TotalPredictedIncidents, base.Summary.TotalPredictedIncidents)
	}
}

func TestCompareScenariosRejectsInvalidInput(t *testing.T) {
	srv := newTestServer(t)

	cases := []models.CompareScenarioRequest{
		{HistoricalWeight: 50},
		{CapacityFactor: -1},
		{TimeWindowHours: 500},
		{RiskLevel: "catastrophic"},
	}
	for _, payload := range cases {
		body, _ := json.Marshal(payload)
		req := authedRequest(http.MethodPost, "/api/v1/forecasts/compare", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %+v got %d", payload, rr.Code)
		}
	}
}

func forecastBody(t *testing.T, srv *server, target string) []byte {
	t.Helper()
	req := authedRequest(http.MethodGet, target, nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	return rr.Body.Bytes()
}
