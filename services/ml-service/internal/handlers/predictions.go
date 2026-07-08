package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

func (s *server) createFloodPredictionHandler(w http.ResponseWriter, r *http.Request) {
	var request models.PredictionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	if !utils.ValidCoordinates(request.Location) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_coordinates", "location.lat must be between -90 and 90 and location.lng must be between -180 and 180")
		return
	}

	response, err := s.store.Predict(request, s.now().UTC())
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "prediction_unavailable", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
