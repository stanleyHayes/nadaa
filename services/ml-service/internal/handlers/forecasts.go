package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

var (
	allowedForecastRiskLevels = map[string]bool{
		"low": true, "moderate": true, "high": true, "severe": true, "emergency": true,
	}
	hazardTypePattern = regexp.MustCompile(`^[a-z][a-z_]{1,39}$`)
)

func (s *server) listForecastsHandler(w http.ResponseWriter, r *http.Request) {
	region := strings.TrimSpace(r.URL.Query().Get("region"))
	now := s.now().UTC()
	utils.WriteJSON(w, http.StatusOK, models.ForecastListResponse{
		Forecasts:   s.store.ListForecasts(region, now),
		GeneratedAt: now.Format(time.RFC3339),
	})
}

func (s *server) getForecastByRegionHandler(w http.ResponseWriter, r *http.Request) {
	region := strings.TrimSpace(r.PathValue("region"))
	if region == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "region path parameter is required")
		return
	}

	now := s.now().UTC()
	forecasts := s.store.ListForecasts(region, now)
	if len(forecasts) == 0 {
		utils.WriteError(w, http.StatusNotFound, "not_found", "no forecasts were found for the given region")
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.ForecastDetailResponse{
		Forecast:    forecasts[0],
		Forecasts:   forecasts,
		GeneratedAt: now.Format(time.RFC3339),
	})
}

func (s *server) listStagingSuggestionsHandler(w http.ResponseWriter, r *http.Request) {
	agencyType := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("agencyType")))
	if agencyType != "" && !hazardTypePattern.MatchString(agencyType) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_agency_type", "agencyType must be a simple lowercase identifier")
		return
	}

	now := s.now().UTC()
	utils.WriteJSON(w, http.StatusOK, models.StagingSuggestionListResponse{
		Suggestions: s.store.StagingSuggestions(agencyType, now),
		GeneratedAt: now.Format(time.RFC3339),
	})
}

func (s *server) compareScenariosHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CompareScenarioRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCompareRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	now := s.now().UTC()
	utils.WriteJSON(w, http.StatusOK, models.CompareScenarioResponse{
		Scenarios:   s.store.CompareScenarios(normalized, now),
		GeneratedAt: now.Format(time.RFC3339),
	})
}

func normalizeCompareRequest(request models.CompareScenarioRequest) (models.CompareScenarioRequest, string, string) {
	request.Region = strings.TrimSpace(request.Region)
	if len(request.Region) > 80 {
		return request, "invalid_region", "region must be at most 80 characters"
	}

	request.RiskLevel = strings.TrimSpace(strings.ToLower(request.RiskLevel))
	if request.RiskLevel != "" && !allowedForecastRiskLevels[request.RiskLevel] {
		return request, "invalid_risk_level", "riskLevel must be low, moderate, high, severe, or emergency"
	}

	if request.HistoricalWeight != 0 && (request.HistoricalWeight < 0.1 || request.HistoricalWeight > 10) {
		return request, "invalid_historical_weight", "historicalWeight must be between 0.1 and 10"
	}
	if request.CapacityFactor != 0 && (request.CapacityFactor < 0.1 || request.CapacityFactor > 10) {
		return request, "invalid_capacity_factor", "capacityFactor must be between 0.1 and 10"
	}
	if request.TimeWindowHours != 0 && (request.TimeWindowHours < 1 || request.TimeWindowHours > 168) {
		return request, "invalid_time_window", "timeWindowHours must be between 1 and 168"
	}

	for i, hazard := range request.HazardTypes {
		hazard = strings.TrimSpace(strings.ToLower(hazard))
		if !hazardTypePattern.MatchString(hazard) {
			return request, "invalid_hazard_type", "hazardTypes must be simple lowercase identifiers"
		}
		request.HazardTypes[i] = hazard
	}

	return request, "", ""
}
