package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
)

func TestPredictionLogsAreCappedWithFIFOEviction(t *testing.T) {
	srv := newTestServer(t)
	var firstID, lastID string
	for i := range 501 {
		resp, err := srv.store.Predict(models.PredictionRequest{
			Location: models.Coordinates{Lat: 5.5600, Lng: -0.2000},
		}, testNow)
		if err != nil {
			t.Fatalf("predict %d: %v", i, err)
		}
		if i == 0 {
			firstID = resp.Log.ID
		}
		lastID = resp.Log.ID
	}

	logs, total := srv.store.ListLogs(1000, 0)
	if total != 500 {
		t.Fatalf("expected 500 retained prediction logs, got %d", total)
	}
	if len(logs) != 500 {
		t.Fatalf("expected a full page of 500 logs, got %d", len(logs))
	}
	for _, entry := range logs {
		if entry.ID == firstID {
			t.Fatal("expected the oldest prediction log to be evicted FIFO")
		}
	}
	if logs[0].ID != lastID {
		t.Fatalf("expected the newest log first, got %s want %s", logs[0].ID, lastID)
	}
}

func TestSimulationRunsAreCappedWithFIFOEviction(t *testing.T) {
	srv := newTestServer(t)
	var firstID, lastID string
	for i := range 101 {
		run, err := srv.store.CreateSimulationJob(models.CreateSimulationRequest{
			Name: fmt.Sprintf("eviction-sim-%d", i), DurationHours: 1, TimeStepHours: 1,
		}, testNow)
		if err != nil {
			t.Fatalf("create simulation %d: %v", i, err)
		}
		if i == 0 {
			firstID = run.ID
		}
		lastID = run.ID
	}

	runs, total := srv.store.ListSimulationJobs(200, 0)
	if total != 100 {
		t.Fatalf("expected 100 retained simulation runs, got %d", total)
	}
	for _, run := range runs {
		if run.ID == firstID {
			t.Fatal("expected the oldest simulation run to be evicted FIFO")
		}
	}
	if runs[0].ID != lastID {
		t.Fatalf("expected the newest run first, got %s want %s", runs[0].ID, lastID)
	}
}

func TestCVResultsAreCappedWithFIFOEviction(t *testing.T) {
	srv := newTestServer(t)
	for i := range 501 {
		if _, err := srv.store.AnalyzeImage(models.CVAnalysisRequest{
			ImageID: fmt.Sprintf("img_evict_%03d", i), ImageName: "flood.jpg",
		}, testNow); err != nil {
			t.Fatalf("analyze %d: %v", i, err)
		}
	}

	results, total := srv.store.ListCVResults(600, 0)
	if total != 500 {
		t.Fatalf("expected 500 retained CV results, got %d", total)
	}
	for _, result := range results {
		if result.ImageID == "img_evict_000" {
			t.Fatal("expected the oldest CV result to be evicted FIFO")
		}
	}
	if results[0].ImageID != "img_evict_500" {
		t.Fatalf("expected the newest result first, got %s", results[0].ImageID)
	}
}

func TestPredictionLogsPagination(t *testing.T) {
	srv := newTestServer(t)
	for range 3 {
		postPrediction(t, srv, models.PredictionRequest{Location: models.Coordinates{Lat: 5.5600, Lng: -0.2000}})
	}

	get := func(target string) models.PredictionLogListResponse {
		t.Helper()
		req := authedRequest(http.MethodGet, target, nil)
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
		}
		var resp models.PredictionLogListResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return resp
	}

	first := get("/api/v1/ml/prediction-logs?limit=2")
	if len(first.Logs) != 2 || first.Total != 3 || first.Limit != 2 || first.Offset != 0 {
		t.Fatalf("unexpected first page: %+v", first)
	}
	second := get("/api/v1/ml/prediction-logs?limit=2&offset=2")
	if len(second.Logs) != 1 || second.Total != 3 || second.Offset != 2 {
		t.Fatalf("unexpected second page: %+v", second)
	}
	if first.Logs[0].ID == second.Logs[0].ID || first.Logs[1].ID == second.Logs[0].ID {
		t.Fatal("expected pages to be disjoint")
	}
	if past := get("/api/v1/ml/prediction-logs?limit=2&offset=99"); len(past.Logs) != 0 || past.Total != 3 {
		t.Fatalf("expected an empty page past the end, got %+v", past)
	}

	for _, target := range []string{
		"/api/v1/ml/prediction-logs?limit=0",
		"/api/v1/ml/prediction-logs?limit=5000",
		"/api/v1/ml/prediction-logs?offset=-1",
	} {
		req := authedRequest(http.MethodGet, target, nil)
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s got %d", target, rr.Code)
		}
	}
}

func TestSimulationsPagination(t *testing.T) {
	srv := newTestServer(t)
	for i := range 3 {
		body, _ := json.Marshal(models.CreateSimulationRequest{
			Name: fmt.Sprintf("paged-sim-%d", i), DurationHours: 1, TimeStepHours: 1,
		})
		req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("create: expected 201 got %d: %s", rr.Code, rr.Body.String())
		}
	}

	req := authedRequest(http.MethodGet, "/api/v1/ml/flood/simulations?limit=2&offset=2", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.SimulationListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Simulations) != 1 || resp.Total != 3 || resp.Limit != 2 || resp.Offset != 2 {
		t.Fatalf("unexpected simulations page: %+v", resp)
	}
}

func TestCVResultsPagination(t *testing.T) {
	srv := newTestServer(t)
	for _, imageID := range []string{"img_page_a", "img_page_b", "img_page_c"} {
		body, _ := json.Marshal(models.CVAnalysisRequest{ImageID: imageID, ImageName: "flood.jpg"})
		req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
	}

	req := authedRequest(http.MethodGet, "/api/v1/cv/results?limit=2&offset=2", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.CVResultListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Results) != 1 || resp.Total != 3 || resp.Limit != 2 || resp.Offset != 2 {
		t.Fatalf("unexpected CV results page: %+v", resp)
	}
	if resp.Results[0].ImageID != "img_page_a" {
		t.Fatalf("expected the oldest result on the last page, got %s", resp.Results[0].ImageID)
	}
}
