package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

func (s *server) createSimulationHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreateSimulationRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	run, err := s.store.CreateSimulationJob(request, s.now().UTC())
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "simulation_validation_failed", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, models.SimulationDetailResponse{Simulation: run})
}

func (s *server) listSimulationsHandler(w http.ResponseWriter, r *http.Request) {
	limit, offset, ok := parsePagination(w, r)
	if !ok {
		return
	}
	jobs, total := s.store.ListSimulationJobs(limit, offset)
	utils.WriteJSON(w, http.StatusOK, models.SimulationListResponse{
		Simulations: jobs,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
		GeneratedAt: s.now().UTC().Format(time.RFC3339),
	})
}

func (s *server) getSimulationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, ok := s.store.GetSimulationJob(id)
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "simulation job was not found")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.SimulationDetailResponse{Simulation: job})
}
