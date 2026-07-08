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
	filter := parseReliefPointFilter(r)
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
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.CreateReliefPointRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
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
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.UpdateReliefPointRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
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
		log.Printf("WARN shelter-service relief_point_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service relief_point_update completed id=%s actor=%s status=%s", reliefPoint.ID, ctx.ActorUserID, reliefPoint.Status)
	utils.WriteJSON(w, http.StatusOK, reliefPoint)
}

func (s *server) listReliefPointStockHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history := s.store.ListReliefPointStockHistory(r.PathValue("id"))
	log.Printf("INFO shelter-service relief_point_stock_history reliefPointId=%s count=%d", r.PathValue("id"), len(history))
	utils.WriteJSON(w, http.StatusOK, models.ReliefPointStockHistoryResponse{ReliefPointID: r.PathValue("id"), History: history, GeneratedAt: s.now().UTC()})
}

func parseReliefPointFilter(r *http.Request) models.ReliefPointFilter {
	query := r.URL.Query()
	filter := models.ReliefPointFilter{
		Status:       utils.NormalizeToken(query.Get("status")),
		Type:         utils.NormalizeToken(query.Get("type")),
		RadiusMeters: utils.NearbySearchMeters,
	}
	if latText := strings.TrimSpace(query.Get("lat")); latText != "" {
		if lngText := strings.TrimSpace(query.Get("lng")); lngText != "" {
			lat, latErr := strconv.ParseFloat(latText, 64)
			lng, lngErr := strconv.ParseFloat(lngText, 64)
			if latErr == nil && lngErr == nil {
				location := models.Coordinates{Lat: lat, Lng: lng}
				if utils.ValidCoordinates(location) {
					filter.Location = &location
				}
			}
		}
	}
	if radiusText := strings.TrimSpace(query.Get("radius")); radiusText != "" {
		if radius, err := strconv.ParseFloat(radiusText, 64); err == nil && radius > 0 {
			filter.RadiusMeters = radius
		}
	}
	if limitText := strings.TrimSpace(query.Get("limit")); limitText != "" {
		if limit, err := strconv.Atoi(limitText); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if bboxValue := strings.TrimSpace(query.Get("bbox")); bboxValue != "" {
		if box, ok := utils.ParseBBox(bboxValue); ok {
			filter.BBox = box
		}
	}
	return filter
}
