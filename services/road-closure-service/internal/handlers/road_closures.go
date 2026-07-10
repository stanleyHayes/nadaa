package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

var closureUpdateRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

var allowedClosureStatuses = map[string]bool{
	"active":    true,
	"scheduled": true,
	"lifted":    true,
	"cancelled": true,
}

var allowedClosureSeverities = map[string]bool{
	"low":       true,
	"moderate":  true,
	"high":      true,
	"severe":    true,
	"emergency": true,
}

func (s *Server) listRoadClosuresHandler(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseListFilter(w, r)
	if !ok {
		return
	}
	closures := s.store.ListClosures(filter, s.now().UTC())
	log.Printf("INFO road-closure-service closure_list count=%d status=%s hasLocation=%t bbox=%t", len(closures), filter.Status, filter.Location != nil, filter.BBox != nil)
	utils.WriteJSON(w, http.StatusOK, models.RoadClosureListResponse{Closures: closures, GeneratedAt: s.now().UTC()})
}

func (s *Server) createRoadClosureHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, closureUpdateRoles)
	if !ok {
		return
	}

	var request models.CreateRoadClosureRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN road-closure-service create_closure invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreate(request, s.now().UTC())
	if code != "" {
		log.Printf("WARN road-closure-service create_closure validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	closure := s.store.CreateClosure(normalized, ctx, s.now().UTC())
	log.Printf("INFO road-closure-service create_closure completed id=%s actor=%s source=%s", closure.ID, ctx.ActorUserID, closure.Source)
	utils.WriteJSON(w, http.StatusCreated, models.RoadClosureResponse{Closure: closure})
}

func (s *Server) updateRoadClosureHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, closureUpdateRoles)
	if !ok {
		return
	}

	var request models.UpdateRoadClosureRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN road-closure-service update_closure invalid_json id=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdate(request)
	if code != "" {
		log.Printf("WARN road-closure-service update_closure validation_failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	closure, code, message := s.store.UpdateClosure(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN road-closure-service update_closure failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO road-closure-service update_closure completed id=%s actor=%s status=%s", closure.ID, ctx.ActorUserID, closure.Status)
	utils.WriteJSON(w, http.StatusOK, models.RoadClosureResponse{Closure: closure})
}

func parseListFilter(w http.ResponseWriter, r *http.Request) (models.ListFilter, bool) {
	filter := models.ListFilter{
		Status:         utils.NormalizeQueryValue(r.URL.Query().Get("status")),
		RadiusMeters:   utils.NearbySearchMeters,
		Limit:          utils.DefaultLimit,
		IncludeExpired: utils.NormalizeQueryValue(r.URL.Query().Get("includeExpired")) == "true",
	}

	if filter.Status != "" && !allowedClosureStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be active, scheduled, lifted, or cancelled")
		return filter, false
	}

	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 100")
			return filter, false
		}
		filter.Limit = parsed
	}

	if value := strings.TrimSpace(r.URL.Query().Get("radius")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || parsed <= 0 || parsed > 100000 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_radius", "radius must be between 1 and 100000 meters")
			return filter, false
		}
		filter.RadiusMeters = parsed
	}

	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText != "" || lngText != "" {
		if latText == "" || lngText == "" {
			utils.WriteError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng must be supplied together")
			return filter, false
		}
		lat, latErr := strconv.ParseFloat(latText, 64)
		lng, lngErr := strconv.ParseFloat(lngText, 64)
		if latErr != nil || lngErr != nil {
			utils.WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
			return filter, false
		}
		loc := models.Coordinates{Lat: lat, Lng: lng}
		if !utils.ValidCoordinates(loc) {
			utils.WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
			return filter, false
		}
		filter.Location = &loc
	}

	if value := strings.TrimSpace(r.URL.Query().Get("bbox")); value != "" {
		parts := strings.Split(value, ",")
		if len(parts) != 4 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_bbox", "bbox must be minLng,minLat,maxLng,maxLat")
			return filter, false
		}
		var floats [4]float64
		for i := range 4 {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, "invalid_bbox", "bbox values must be valid decimal coordinates")
				return filter, false
			}
			floats[i] = parsed
		}
		//nolint:gosec // G602 false positive: floats is a fixed [4]float64 array, so indices 0-3 are always in range.
		filter.BBox = &models.BBox{MinLng: floats[0], MinLat: floats[1], MaxLng: floats[2], MaxLat: floats[3]}
	}

	return filter, true
}

func normalizeCreate(request models.CreateRoadClosureRequest, now time.Time) (models.CreateRoadClosureRequest, string, string) {
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Status = utils.NormalizeQueryValue(request.Status)
	request.Severity = utils.NormalizeQueryValue(request.Severity)
	request.Source = utils.NormalizeQueryValue(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.DetourNote = strings.TrimSpace(request.DetourNote)

	if request.RoadName == "" || len(request.RoadName) > 200 || utils.UnsafeText(request.RoadName) {
		return request, "invalid_road_name", "roadName is required and must be 200 safe characters or fewer"
	}
	if request.Status == "" {
		request.Status = "active"
	}
	if !allowedClosureStatuses[request.Status] {
		return request, "invalid_status", "status must be active, scheduled, lifted, or cancelled"
	}
	if request.Severity == "" {
		request.Severity = utils.SeverityFromStatus(request.Status)
	}
	if !allowedClosureSeverities[request.Severity] {
		return request, "invalid_severity", "severity must be low, moderate, high, severe, or emergency"
	}
	if errCode, errMsg := utils.ValidateGeometry(request.Geometry); errCode != "" {
		return request, errCode, errMsg
	}
	if request.Source == "" {
		request.Source = "manual"
	}
	if len(request.Source) > 80 || utils.UnsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || utils.UnsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	if len(request.Reason) > 200 || utils.UnsafeText(request.Reason) {
		return request, "invalid_reason", "reason must be 200 safe characters or fewer"
	}
	if len(request.DetourNote) > 500 || utils.UnsafeText(request.DetourNote) {
		return request, "invalid_detour_note", "detourNote must be 500 safe characters or fewer"
	}
	if request.ValidFrom != nil && request.ValidTo != nil && request.ValidTo.Before(*request.ValidFrom) {
		return request, "invalid_valid_to", "validTo must be after validFrom"
	}
	return request, "", ""
}

func normalizeUpdate(request models.UpdateRoadClosureRequest) (models.UpdateRoadClosureRequest, string, string) {
	request.RoadName = strings.TrimSpace(request.RoadName)
	request.Reason = strings.TrimSpace(request.Reason)
	request.Status = utils.NormalizeQueryValue(request.Status)
	request.Severity = utils.NormalizeQueryValue(request.Severity)
	request.Source = utils.NormalizeQueryValue(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	request.DetourNote = strings.TrimSpace(request.DetourNote)

	if request.RoadName != "" && (len(request.RoadName) > 200 || utils.UnsafeText(request.RoadName)) {
		return request, "invalid_road_name", "roadName must be 200 safe characters or fewer"
	}
	if request.Status != "" && !allowedClosureStatuses[request.Status] {
		return request, "invalid_status", "status must be active, scheduled, lifted, or cancelled"
	}
	if request.Severity != "" && !allowedClosureSeverities[request.Severity] {
		return request, "invalid_severity", "severity must be low, moderate, high, severe, or emergency"
	}
	if request.Geometry != nil {
		if errCode, errMsg := utils.ValidateGeometry(*request.Geometry); errCode != "" {
			return request, errCode, errMsg
		}
	}
	if request.Source != "" && (len(request.Source) > 80 || utils.UnsafeText(request.Source)) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if request.SourceRef != "" && (len(request.SourceRef) > 120 || utils.UnsafeText(request.SourceRef)) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	if len(request.Reason) > 200 || utils.UnsafeText(request.Reason) {
		return request, "invalid_reason", "reason must be 200 safe characters or fewer"
	}
	if len(request.DetourNote) > 500 || utils.UnsafeText(request.DetourNote) {
		return request, "invalid_detour_note", "detourNote must be 500 safe characters or fewer"
	}
	if request.ValidFrom != nil && request.ValidTo != nil && request.ValidTo.Before(*request.ValidFrom) {
		return request, "invalid_valid_to", "validTo must be after validFrom"
	}
	if request.RoadName == "" && request.Reason == "" && request.Status == "" && request.Severity == "" &&
		request.Source == "" && request.SourceRef == "" && request.Geometry == nil &&
		request.ValidFrom == nil && request.ValidTo == nil && request.DetourNote == "" {
		return request, "no_changes", "at least one closure field must be supplied"
	}
	return request, "", ""
}

func requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     utils.NormalizeQueryValue(r.Header.Get("X-NADAA-Actor-Role")),
		MFACompleted:  utils.NormalizeQueryValue(r.Header.Get("X-NADAA-MFA-Completed")) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for road closure updates")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to update road closures")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}
