package handlers

import (
	"net/http"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

var allowedImportStatuses = map[string]bool{
	"running":   true,
	"succeeded": true,
	"failed":    true,
}

func (s *server) createObservationImportJobHandler(w http.ResponseWriter, r *http.Request) {
	request, ok := decodeOptionalObservationImportRequest(w, r)
	if !ok {
		return
	}
	if code, message := validateObservationImportRequest(request); code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	job := s.store.CreateObservationImportJob(request, "manual", time.Now().UTC(), 1)
	utils.WriteJSON(w, http.StatusAccepted, job)
}

func (s *server) listObservationImportJobsHandler(w http.ResponseWriter, r *http.Request) {
	status := utils.NormalizeQueryValue(r.URL.Query().Get("status"))
	if status != "" && !allowedImportStatuses[status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be succeeded, failed, or running")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.ObservationImportJobListResponse{Jobs: s.store.ListObservationImportJobs(status)})
}

func (s *server) retryObservationImportJobHandler(w http.ResponseWriter, r *http.Request) {
	job, ok, conflict := s.store.RetryObservationImportJob(r.PathValue("id"), time.Now().UTC())
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "import_job_not_found", "import job was not found")
		return
	}
	if conflict != "" {
		utils.WriteError(w, http.StatusConflict, "import_job_not_retryable", conflict)
		return
	}

	utils.WriteJSON(w, http.StatusAccepted, job)
}

func decodeOptionalObservationImportRequest(w http.ResponseWriter, r *http.Request) (models.ObservationImportRequest, bool) {
	if r.Body == nil || r.ContentLength == 0 {
		return models.ObservationImportRequest{}, true
	}

	var request models.ObservationImportRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return models.ObservationImportRequest{}, false
	}
	return request, true
}

func validateObservationImportRequest(request models.ObservationImportRequest) (string, string) {
	metric := utils.NormalizeQueryValue(request.Metric)
	if metric != "" && metric != "rainfall_mm" && metric != "water_level_m" {
		return "invalid_metric", "metric must be rainfall_mm or water_level_m"
	}
	return "", ""
}
