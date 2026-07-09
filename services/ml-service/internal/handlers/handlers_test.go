package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
)

func TestCreateSimulation(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CreateSimulationRequest{Name: "Test simulation", RainfallMmOverride: 30, DurationHours: 3, TimeStepHours: 1}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 got %d: %s", rr.Code, rr.Body.String())
	}

	var resp models.SimulationDetailResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Simulation.ID == "" {
		t.Error("expected simulation id")
	}
	if !strings.HasPrefix(resp.Simulation.Reference, "FS-") {
		t.Errorf("expected FS- reference got %s", resp.Simulation.Reference)
	}
	if len(resp.Simulation.Frames) != 3 {
		t.Errorf("expected 3 frames got %d", len(resp.Simulation.Frames))
	}
	if !resp.Simulation.Safety.HumanReviewRequired || resp.Simulation.Safety.AutoPublishAllowed {
		t.Error("expected safety policy to require review and block auto-publish")
	}
	if len(resp.Simulation.Limitations) == 0 {
		t.Error("expected limitations")
	}
}

func TestCreateSimulationRequiresName(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CreateSimulationRequest{DurationHours: 2}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 got %d", rr.Code)
	}
}

func TestListAndGetSimulation(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CreateSimulationRequest{Name: "Listable simulation", DurationHours: 2}
	body, _ := json.Marshal(payload)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
	createRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(createRR, createReq)

	var created models.SimulationDetailResponse
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/ml/flood/simulations", nil)
	listRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", listRR.Code)
	}
	var list models.SimulationListResponse
	if err := json.Unmarshal(listRR.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list.Simulations) != 1 {
		t.Errorf("expected 1 simulation got %d", len(list.Simulations))
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/ml/flood/simulations/"+created.Simulation.ID, nil)
	getRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", getRR.Code)
	}
	var got models.SimulationDetailResponse
	if err := json.Unmarshal(getRR.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.Simulation.ID != created.Simulation.ID {
		t.Error("get returned wrong simulation")
	}
}

func TestGetSimulationNotFound(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ml/flood/simulations/does-not-exist", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 got %d", rr.Code)
	}
}
