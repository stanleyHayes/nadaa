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

	req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
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

	req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 got %d", rr.Code)
	}
}

func TestListAndGetSimulation(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CreateSimulationRequest{Name: "Listable simulation", DurationHours: 2, TimeStepHours: 1}
	body, _ := json.Marshal(payload)

	createReq := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
	createRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(createRR, createReq)

	var created models.SimulationDetailResponse
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}

	listReq := authedRequest(http.MethodGet, "/api/v1/ml/flood/simulations", nil)
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

	getReq := authedRequest(http.MethodGet, "/api/v1/ml/flood/simulations/"+created.Simulation.ID, nil)
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
	req := authedRequest(http.MethodGet, "/api/v1/ml/flood/simulations/does-not-exist", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 got %d", rr.Code)
	}
}

func TestCreateSimulationRejectsExcessiveDuration(t *testing.T) {
	srv := newTestServer(t)
	for _, hours := range []int{169, 1000000000} {
		payload := models.CreateSimulationRequest{Name: "Too long", DurationHours: hours, TimeStepHours: 1}
		body, _ := json.Marshal(payload)

		req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for durationHours=%d got %d", hours, rr.Code)
		}
	}
}

func TestCreateSimulationAcceptsMaxDuration(t *testing.T) {
	srv := newTestServer(t)
	payload := models.CreateSimulationRequest{Name: "One week", DurationHours: 168, TimeStepHours: 24}
	body, _ := json.Marshal(payload)

	req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 got %d: %s", rr.Code, rr.Body.String())
	}
	var resp models.SimulationDetailResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Simulation.Frames) != 7 {
		t.Errorf("expected 7 frames got %d", len(resp.Simulation.Frames))
	}
}

func TestSimulationIDsUniqueWithinSameSecond(t *testing.T) {
	srv := newTestServer(t)

	create := func(name string) models.SimulationRun {
		t.Helper()
		payload := models.CreateSimulationRequest{Name: name, DurationHours: 2, TimeStepHours: 1}
		body, _ := json.Marshal(payload)
		req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status 201 got %d: %s", rr.Code, rr.Body.String())
		}
		var resp models.SimulationDetailResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		return resp.Simulation
	}

	// The test clock is fixed, so both jobs land in the same second.
	first := create("First same-second simulation")
	second := create("Second same-second simulation")

	if first.ID == second.ID {
		t.Fatalf("expected unique simulation IDs, both were %s", first.ID)
	}
	if first.Reference == second.Reference {
		t.Fatalf("expected unique references, both were %s", first.Reference)
	}
	if first.Status != "completed" || second.Status != "completed" {
		t.Fatalf("expected both simulations completed, got %s and %s", first.Status, second.Status)
	}

	// The completion update must not clobber the other same-second run: the
	// second job must be retrievable under its own ID with its own name.
	getReq := authedRequest(http.MethodGet, "/api/v1/ml/flood/simulations/"+second.ID, nil)
	getRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", getRR.Code)
	}
	var got models.SimulationDetailResponse
	if err := json.Unmarshal(getRR.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.Simulation.Name != second.Name || got.Simulation.Status != "completed" {
		t.Fatalf("completion update clobbered another run: got %+v", got.Simulation)
	}

	listReq := authedRequest(http.MethodGet, "/api/v1/ml/flood/simulations", nil)
	listRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(listRR, listReq)
	var list models.SimulationListResponse
	if err := json.Unmarshal(listRR.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, run := range list.Simulations {
		if run.Status != "completed" {
			t.Fatalf("found stuck %q simulation %s", run.Status, run.ID)
		}
	}
}

func TestCreateSimulationRejectsNonPositiveParameters(t *testing.T) {
	srv := newTestServer(t)
	cases := []models.CreateSimulationRequest{
		{Name: "zero duration", DurationHours: 0, TimeStepHours: 1},
		{Name: "negative duration", DurationHours: -5, TimeStepHours: 1},
		{Name: "zero timestep", DurationHours: 2, TimeStepHours: 0},
		{Name: "negative timestep", DurationHours: 2, TimeStepHours: -1},
	}
	for _, payload := range cases {
		body, _ := json.Marshal(payload)
		req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		srv.Routes().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for %+v got %d", payload, rr.Code)
		}
	}
}

func TestCreateSimulationNameLengthCapped(t *testing.T) {
	srv := newTestServer(t)

	tooLong, _ := json.Marshal(models.CreateSimulationRequest{
		Name: strings.Repeat("x", 201), DurationHours: 2, TimeStepHours: 1,
	})
	req := authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(tooLong))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for a 201-character name got %d", rr.Code)
	}

	atCap, _ := json.Marshal(models.CreateSimulationRequest{
		Name: strings.Repeat("x", 200), DurationHours: 2, TimeStepHours: 1,
	})
	req = authedRequest(http.MethodPost, "/api/v1/ml/flood/simulations", bytes.NewReader(atCap))
	rr = httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for a 200-character name got %d: %s", rr.Code, rr.Body.String())
	}
}
