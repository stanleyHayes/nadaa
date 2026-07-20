package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

func (s *server) listReliefPointsHandler(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseReliefPointFilter(w, r)
	if !ok {
		return
	}
	reliefPoints := s.store.ListReliefPoints(filter)
	log.Printf("INFO shelter-service relief_point_list count=%d status=%s type=%s hasLocation=%t bbox=%t", len(reliefPoints), filter.Status, filter.Type, filter.Location != nil, filter.BBox != nil)
	utils.WriteJSON(w, http.StatusOK, models.ReliefPointListResponse{ReliefPoints: reliefPoints, GeneratedAt: s.now().UTC()})
}

func (s *server) nearbyReliefPointsHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := utils.ParseLocation(w, r)
	if !ok {
		return
	}
	reliefPoints := s.store.NearbyReliefPoints(location, utils.DefaultNearbyLimit)
	log.Printf("INFO shelter-service relief_point_nearby count=%d lat=%.3f lng=%.3f", len(reliefPoints), location.Lat, location.Lng)
	utils.WriteJSON(w, http.StatusOK, models.ReliefPointNearbyResponse{ReliefPoints: reliefPoints, GeneratedAt: s.now().UTC()})
}

func (s *server) createReliefPointHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.CreateReliefPointRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		log.Printf("WARN shelter-service relief_point_create invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeCreateReliefPoint(request)
	if code != "" {
		log.Printf("WARN shelter-service relief_point_create validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	reliefPoint := s.store.CreateReliefPoint(normalized, ctx, s.now().UTC())
	log.Printf("INFO shelter-service relief_point_create completed id=%s actor=%s source=%s", reliefPoint.ID, ctx.ActorUserID, reliefPoint.Source)
	utils.WriteJSON(w, http.StatusCreated, reliefPoint)
}

func (s *server) updateReliefPointHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.UpdateReliefPointRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		log.Printf("WARN shelter-service relief_point_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeUpdateReliefPoint(request)
	if code != "" {
		log.Printf("WARN shelter-service relief_point_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	reliefPoint, code, message := s.store.UpdateReliefPoint(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service relief_point_update failed id=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service relief_point_update completed id=%s actor=%s status=%s", reliefPoint.ID, ctx.ActorUserID, reliefPoint.Status)
	utils.WriteJSON(w, http.StatusOK, reliefPoint)
}

func (s *server) deleteReliefPointHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, utils.ShelterDeleteRoles)
	if !ok {
		return
	}

	id := r.PathValue("id")
	// Sanitize user-controlled path values before logging (G706).
	idLog := utils.SafeLogValue(id)
	if !s.store.DeleteReliefPoint(id) {
		// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
		log.Printf("WARN shelter-service relief_point_delete not_found id=%s actor=%s", idLog, ctx.ActorUserID)
		utils.WriteError(w, http.StatusNotFound, "not_found", "relief point was not found")
		return
	}
	// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
	log.Printf("INFO shelter-service relief_point_delete completed id=%s actor=%s", idLog, ctx.ActorUserID)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listReliefPointStockHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history := s.store.ListReliefPointStockHistory(r.PathValue("id"))
	// #nosec G706 -- path values are sanitized with utils.SafeLogValue.
	log.Printf("INFO shelter-service relief_point_stock_history reliefPointId=%s count=%d", utils.SafeLogValue(r.PathValue("id")), len(history))
	utils.WriteJSON(w, http.StatusOK, models.ReliefPointStockHistoryResponse{ReliefPointID: r.PathValue("id"), History: history, GeneratedAt: s.now().UTC()})
}

func parseReliefPointFilter(w http.ResponseWriter, r *http.Request) (models.ReliefPointFilter, bool) {
	query := r.URL.Query()
	filter := models.ReliefPointFilter{
		Status:       utils.NormalizeToken(query.Get("status")),
		Type:         utils.NormalizeToken(query.Get("type")),
		RadiusMeters: utils.NearbySearchMeters,
	}
	if filter.Status != "" && !utils.AllowedReliefPointStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be open, limited, closed, or paused")
		return filter, false
	}
	if filter.Type != "" && !utils.AllowedReliefPointTypes[filter.Type] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_type", "type must be food, water, medical, hygiene, blankets, cash, or mixed")
		return filter, false
	}
	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 100")
			return filter, false
		}
		filter.Limit = parsed
	}
	if value := strings.TrimSpace(query.Get("radius")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || parsed <= 0 || parsed > 100000 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_radius", "radius must be greater than zero and no more than 100000")
			return filter, false
		}
		filter.RadiusMeters = parsed
	}
	if bboxValue := strings.TrimSpace(query.Get("bbox")); bboxValue != "" {
		box, ok := utils.ParseBBox(bboxValue)
		if !ok {
			utils.WriteError(w, http.StatusBadRequest, "invalid_bbox", "bbox must be minLng,minLat,maxLng,maxLat")
			return filter, false
		}
		filter.BBox = box
	}

	latText := strings.TrimSpace(query.Get("lat"))
	lngText := strings.TrimSpace(query.Get("lng"))
	if latText == "" && lngText == "" {
		return filter, true
	}
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
	location := models.Coordinates{Lat: lat, Lng: lng}
	if !utils.ValidCoordinates(location) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
		return filter, false
	}
	filter.Location = &location
	return filter, true
}
