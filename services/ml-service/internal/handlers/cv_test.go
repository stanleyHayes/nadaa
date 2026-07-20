package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
)

func TestCVAnalyzeFloodImage(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageID: "img_flood_001", ImageName: "flooded-road.jpg"}
	body, _ := json.Marshal(payload)

	req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.CVAnalysisResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Result.ImageID != "img_flood_001" {
		t.Errorf("expected imageId img_flood_001, got %s", resp.Result.ImageID)
	}
	if len(resp.Result.Labels) == 0 {
		t.Error("expected at least one label")
	}
	if resp.Result.Labels[0].Label != "flood_evidence" {
		t.Errorf("expected flood_evidence label, got %s", resp.Result.Labels[0].Label)
	}
	if resp.Result.ModelVersion != "cv-mock-rule-engine-0.1.0" {
		t.Errorf("expected mock model version, got %s", resp.Result.ModelVersion)
	}
	if resp.Result.HumanReviewRequired {
		t.Error("expected no human review required for high-confidence flood")
	}
	if resp.Safety.AutoPublishAllowed {
		t.Error("expected auto-publish blocked")
	}
}

func TestCVAnalyzeFireImage(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageID: "img_fire_001", ImageName: "fire-building.jpg"}
	body, _ := json.Marshal(payload)

	req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	var resp models.CVAnalysisResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Result.Labels[0].Label != "fire_evidence" {
		t.Errorf("expected fire_evidence label, got %s", resp.Result.Labels[0].Label)
	}
}

func TestCVAnalyzeSensitiveImageRequiresReview(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageID: "img_injured_001", ImageName: "injured-person.jpg"}
	body, _ := json.Marshal(payload)

	req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	var resp models.CVAnalysisResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.Result.HumanReviewRequired {
		t.Error("expected human review required for sensitive image")
	}
	if resp.Result.Labels[0].Label != "sensitive" {
		t.Errorf("expected sensitive label, got %s", resp.Result.Labels[0].Label)
	}
}

func TestCVAnalyzeLowConfidenceRequiresReview(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageID: "img_random_001", ImageName: "random-test.jpg"}
	body, _ := json.Marshal(payload)

	req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	var resp models.CVAnalysisResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if !resp.Result.HumanReviewRequired {
		t.Error("expected human review required for low-confidence/unclear image")
	}
}

func TestCVAnalyzeRequiresImageID(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageName: "no-id.jpg"}
	body, _ := json.Marshal(payload)

	req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 got %d", rr.Code)
	}
}

func TestCVGetResult(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageID: "img_cached_001", ImageName: "flood-scene.jpg"}
	body, _ := json.Marshal(payload)

	createReq := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	createRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", createRR.Code)
	}

	getReq := authedRequest(http.MethodGet, "/api/v1/cv/results/img_cached_001", nil)
	getRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", getRR.Code)
	}

	var resp models.CVResultDetailResponse
	if err := json.Unmarshal(getRR.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if resp.Result.ImageID != "img_cached_001" {
		t.Errorf("expected imageId img_cached_001, got %s", resp.Result.ImageID)
	}
}

func TestCVGetResultNotFound(t *testing.T) {
	srv := newTestServer(t)
	req := authedRequest(http.MethodGet, "/api/v1/cv/results/does-not-exist", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 got %d", rr.Code)
	}
}

func TestCVListResults(t *testing.T) {
	srv := newTestServer(t)
	for _, imageID := range []string{"img_a", "img_b"} {
		payload := models.CVAnalysisRequest{ImageID: imageID, ImageName: "flood.jpg"}
		body, _ := json.Marshal(payload)
		req := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
	}

	listReq := authedRequest(http.MethodGet, "/api/v1/cv/results", nil)
	listRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", listRR.Code)
	}

	var resp models.CVResultListResponse
	if err := json.Unmarshal(listRR.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results got %d", len(resp.Results))
	}
}

func TestCVResultCaching(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CVAnalysisRequest{ImageID: "img_cache_test", ImageName: "fire-scene.jpg"}
	body, _ := json.Marshal(payload)

	req1 := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr1 := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr1, req1)

	var resp1 models.CVAnalysisResponse
	if err := json.Unmarshal(rr1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("decode first response: %v", err)
	}

	req2 := authedRequest(http.MethodPost, "/api/v1/cv/analyze", bytes.NewReader(body))
	rr2 := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr2, req2)

	var resp2 models.CVAnalysisResponse
	if err := json.Unmarshal(rr2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("decode second response: %v", err)
	}

	if resp1.Result.ID != resp2.Result.ID {
		t.Error("expected cached result to have same ID")
	}
}
