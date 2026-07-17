package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/utils"
)

func (s *Server) riskHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}

	risk := s.store.AreaRisk(location)
	if s.mlClient != nil {
		if prediction, err := s.mlClient.predict(r.Context(), location, r.Header.Get("Authorization")); err != nil {
			log.Printf("ml prediction unavailable: %v", err)
		} else {
			risk.MLPrediction = &prediction
			if utils.RiskRank(prediction.Severity) > utils.RiskRank(risk.OverallRisk) {
				risk.OverallRisk = prediction.Severity
				risk.RecommendedActions = utils.RecommendedActions(risk.OverallRisk, utils.RisksFloodLevel(risk.Risks))
			}
		}
	}
	utils.WriteJSON(w, http.StatusOK, risk)
}

func parseLocation(w http.ResponseWriter, r *http.Request) (models.Coordinates, bool) {
	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" || lngText == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng query parameters are required")
		return models.Coordinates{}, false
	}

	lat, latErr := strconv.ParseFloat(latText, 64)
	lng, lngErr := strconv.ParseFloat(lngText, 64)
	if latErr != nil || lngErr != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
		return models.Coordinates{}, false
	}

	location := models.Coordinates{Lat: lat, Lng: lng}
	if !models.ValidCoordinates(location) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
		return models.Coordinates{}, false
	}

	return location, true
}
