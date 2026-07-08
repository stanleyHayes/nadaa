package handlers

import (
	"net/http"
	"strconv"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/guide-service/internal/utils"
)

const (
	hazardTypesKey = "hazard"
	stageKey       = "stage"
	offlineKey     = "offline"
	languageKey    = "language"
)

var allowedHazards = map[string]bool{
	"flood":             true,
	"fire":              true,
	"road_crash":        true,
	"building_collapse": true,
	"medical_emergency": true,
	"security_incident": true,
	"disease_outbreak":  true,
	"electrical_hazard": true,
	"blocked_drain":     true,
	"landslide":         true,
	"marine_accident":   true,
	"storm":             true,
	"tidal_wave":        true,
	"other":             true,
}

var allowedStages = map[string]bool{
	"before":   true,
	"during":   true,
	"after":    true,
	"recovery": true,
}

// listGuidesHandler returns the guide catalog filtered by query parameters.
func (s *Server) listGuidesHandler(w http.ResponseWriter, r *http.Request) {
	filters, code, message := parseGuideFilters(r)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.GuideListResponse{Guides: s.store.ListGuides(r.Context(), filters)})
}

// parseGuideFilters validates and normalizes guide list query parameters.
func parseGuideFilters(r *http.Request) (models.GuideFilters, string, string) {
	query := r.URL.Query()
	filters := models.GuideFilters{
		HazardType: utils.NormalizeQueryValue(query.Get(hazardTypesKey)),
		Stage:      utils.NormalizeQueryValue(query.Get(stageKey)),
		Language:   utils.NormalizeLanguage(query.Get(languageKey)),
	}

	if filters.HazardType != "" && !allowedHazards[filters.HazardType] {
		return models.GuideFilters{}, "invalid_hazard", "hazard must be a supported NADAA hazard type"
	}
	if filters.Stage != "" && !allowedStages[filters.Stage] {
		return models.GuideFilters{}, "invalid_stage", "stage must be before, during, after, or recovery"
	}
	if offlineRaw := utils.NormalizeQueryValue(query.Get(offlineKey)); offlineRaw != "" {
		offline, err := strconv.ParseBool(offlineRaw)
		if err != nil {
			return models.GuideFilters{}, "invalid_offline", "offline must be true or false"
		}
		filters.Offline = &offline
	}

	return filters, "", ""
}
