package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

func (s *server) listHospitalCapacityHandler(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseHospitalCapacityFilter(w, r)
	if !ok {
		return
	}
	facilities := s.store.ListHospitalCapacity(filter, s.now().UTC())
	log.Printf("INFO shelter-service hospital_capacity_list count=%d service=%s emergencyCapacity=%s minAvailableBeds=%d includeStale=%t", len(facilities), filter.Service, filter.EmergencyCapacity, filter.MinAvailableBeds, filter.IncludeStale)
	utils.WriteJSON(w, http.StatusOK, models.HospitalCapacityResponse{
		Facilities:            facilities,
		GeneratedAt:           s.now().UTC(),
		StaleThresholdMinutes: utils.HospitalCapacityStaleAfterMinutes,
	})
}

func (s *server) updateHospitalCapacityHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	// Sanitize user-controlled path values before logging (G706).
	facilityIDLog := utils.SafeLogValue(r.PathValue("id"))

	var request models.HospitalCapacityUpdateRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service hospital_capacity_update invalid_json facilityId=%s actor=%s error=%v", facilityIDLog, ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
	log.Printf("INFO shelter-service hospital_capacity_update received facilityId=%s actor=%s source=%s", facilityIDLog, ctx.ActorUserID, request.Source)

	normalized, code, message := utils.NormalizeHospitalCapacityUpdate(request)
	if code != "" {
		// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service hospital_capacity_update validation_failed facilityId=%s actor=%s code=%s", facilityIDLog, ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	facility, code, message := s.store.UpdateHospitalCapacity(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service hospital_capacity_update failed facilityId=%s actor=%s code=%s", facilityIDLog, ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service hospital_capacity_update completed facilityId=%s availableBeds=%d emergencyCapacity=%s source=%s", facility.ID, facility.AvailableBeds, facility.EmergencyCapacity, facility.Source)
	utils.WriteJSON(w, http.StatusOK, models.HospitalCapacityUpdateResponse{Facility: facility})
}

func (s *server) importHospitalCapacityFixtureHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.HospitalCapacityImportRequest
	if err := utils.OptionalDecodeJSON(w, r, &request); err != nil {
		log.Printf("WARN shelter-service hospital_capacity_fixture_import invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized, code, message := utils.NormalizeHospitalCapacityImport(request)
	if code != "" {
		log.Printf("WARN shelter-service hospital_capacity_fixture_import validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	facilities, imported := s.store.ImportHospitalCapacityFixture(normalized, ctx, s.now().UTC())
	log.Printf("INFO shelter-service hospital_capacity_fixture_import completed actor=%s source=%s imported=%d", ctx.ActorUserID, normalized.Source, imported)
	utils.WriteJSON(w, http.StatusOK, models.HospitalCapacityImportResponse{
		Imported:    imported,
		Facilities:  facilities,
		GeneratedAt: s.now().UTC(),
		Source:      normalized.Source,
	})
}

func parseHospitalCapacityFilter(w http.ResponseWriter, r *http.Request) (models.HospitalCapacityFilter, bool) {
	filter := models.HospitalCapacityFilter{
		EmergencyCapacity: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("emergencyCapacity"))),
		IncludeStale:      strings.TrimSpace(strings.ToLower(r.URL.Query().Get("includeStale"))) != "false",
		Limit:             utils.DefaultNearbyLimit,
		Service:           utils.NormalizeToken(r.URL.Query().Get("service")),
	}
	if filter.EmergencyCapacity != "" && !utils.AllowedEmergencyCapacityStatuses[filter.EmergencyCapacity] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_emergency_capacity", "emergencyCapacity must be available, limited, full, offline, or unknown")
		return filter, false
	}
	if value := strings.TrimSpace(r.URL.Query().Get("minAvailableBeds")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_min_available_beds", "minAvailableBeds must be zero or greater")
			return filter, false
		}
		filter.MinAvailableBeds = parsed
	}
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 50 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 50")
			return filter, false
		}
		filter.Limit = parsed
	}

	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" && lngText == "" {
		return filter, true
	}
	if latText == "" || lngText == "" {
		utils.WriteError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng must be supplied together")
		return filter, false
	}
	location, ok := utils.ParseLocation(w, r)
	if !ok {
		return filter, false
	}
	filter.Location = &location
	return filter, true
}
