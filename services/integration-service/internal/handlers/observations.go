package handlers

import (
	"net/http"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

func (s *server) listObservationsHandler(w http.ResponseWriter, r *http.Request) {
	source := utils.NormalizeQueryValue(r.URL.Query().Get("source"))
	metric := utils.NormalizeQueryValue(r.URL.Query().Get("metric"))
	if metric != "" && metric != "rainfall_mm" && metric != "water_level_m" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_metric", "metric must be rainfall_mm or water_level_m")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.ObservationListResponse{Observations: s.store.ListObservations(source, metric)})
}

func (s *server) listImportedObservationsHandler(w http.ResponseWriter, r *http.Request) {
	source := utils.NormalizeQueryValue(r.URL.Query().Get("source"))
	metric := utils.NormalizeQueryValue(r.URL.Query().Get("metric"))
	if metric != "" && metric != "rainfall_mm" && metric != "water_level_m" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_metric", "metric must be rainfall_mm or water_level_m")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.ImportedObservationListResponse{Observations: s.store.ListImportedObservations(source, metric)})
}
