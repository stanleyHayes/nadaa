package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

func (s *server) listAidRequestsHandler(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseAidRequestFilter(w, r)
	if !ok {
		return
	}
	aidRequests := s.store.ListAidRequests(filter)
	log.Printf("INFO shelter-service aid_request_list count=%d status=%s category=%s priority=%s includePrivate=%t", len(aidRequests), filter.Status, filter.Category, filter.Priority, filter.IncludePrivate)
	utils.WriteJSON(w, http.StatusOK, models.AidRequestListResponse{AidRequests: aidRequests, GeneratedAt: s.now().UTC()})
}

func (s *server) createAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.CreateAidRequestRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service aid_request_create invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeCreateAidRequest(request, s.now().UTC())
	if code != "" {
		log.Printf("WARN shelter-service aid_request_create validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	aidRequest := s.store.CreateAidRequest(normalized, ctx, s.now().UTC())
	log.Printf("INFO shelter-service aid_request_create completed id=%s actor=%s category=%s priority=%s", aidRequest.ID, ctx.ActorUserID, aidRequest.Category, aidRequest.Priority)
	utils.WriteJSON(w, http.StatusCreated, aidRequest)
}

func (s *server) reviewAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.ReviewAidRequestRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service aid_request_review invalid_json actor=%s requestId=%s error=%v", ctx.ActorUserID, r.PathValue("id"), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeReviewAidRequest(request)
	if code != "" {
		log.Printf("WARN shelter-service aid_request_review validation_failed actor=%s requestId=%s code=%s", ctx.ActorUserID, r.PathValue("id"), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	aidRequest, code, message := s.store.ReviewAidRequest(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN shelter-service aid_request_review failed actor=%s requestId=%s code=%s", ctx.ActorUserID, r.PathValue("id"), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service aid_request_review completed id=%s actor=%s status=%s", aidRequest.ID, ctx.ActorUserID, aidRequest.Status)
	utils.WriteJSON(w, http.StatusOK, aidRequest)
}

func (s *server) deleteAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	id := r.PathValue("id")
	if !s.store.DeleteAidRequest(id) {
		log.Printf("WARN shelter-service aid_request_delete not_found id=%s actor=%s", id, ctx.ActorUserID)
		utils.WriteError(w, http.StatusNotFound, "not_found", "aid request was not found")
		return
	}
	log.Printf("INFO shelter-service aid_request_delete completed id=%s actor=%s", id, ctx.ActorUserID)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) listAidPledgesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	pledges, code, message := s.store.ListAidPledges(r.PathValue("id"))
	if code != "" {
		log.Printf("WARN shelter-service aid_pledge_list failed actor=%s requestId=%s code=%s", ctx.ActorUserID, r.PathValue("id"), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service aid_pledge_list requestId=%s count=%d", r.PathValue("id"), len(pledges))
	utils.WriteJSON(w, http.StatusOK, models.AidPledgeListResponse{AidRequestID: r.PathValue("id"), Pledges: pledges, GeneratedAt: s.now().UTC()})
}

func (s *server) createAidPledgeHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreateAidPledgeRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service aid_pledge_create invalid_json requestId=%s error=%v", r.PathValue("id"), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeCreateAidPledge(request)
	if code != "" {
		log.Printf("WARN shelter-service aid_pledge_create validation_failed requestId=%s code=%s", r.PathValue("id"), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.CreateAidPledge(r.PathValue("id"), normalized, s.now().UTC())
	if code != "" {
		log.Printf("WARN shelter-service aid_pledge_create failed requestId=%s code=%s", r.PathValue("id"), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service aid_pledge_create completed requestId=%s pledgeId=%s donorType=%s quantity=%d", pledge.AidRequestID, pledge.ID, pledge.DonorType, pledge.Quantity)
	utils.WriteJSON(w, http.StatusCreated, pledge)
}

func (s *server) reviewAidPledgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	var request models.ReviewAidPledgeRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service aid_pledge_review invalid_json actor=%s requestId=%s pledgeId=%s error=%v", ctx.ActorUserID, r.PathValue("id"), r.PathValue("pledgeId"), err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := utils.NormalizeReviewAidPledge(request)
	if code != "" {
		log.Printf("WARN shelter-service aid_pledge_review validation_failed actor=%s requestId=%s pledgeId=%s code=%s", ctx.ActorUserID, r.PathValue("id"), r.PathValue("pledgeId"), code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.ReviewAidPledge(r.PathValue("id"), r.PathValue("pledgeId"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN shelter-service aid_pledge_review failed actor=%s requestId=%s pledgeId=%s code=%s", ctx.ActorUserID, r.PathValue("id"), r.PathValue("pledgeId"), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service aid_pledge_review completed requestId=%s pledgeId=%s actor=%s status=%s reviewStatus=%s", pledge.AidRequestID, pledge.ID, ctx.ActorUserID, pledge.Status, pledge.ReviewStatus)
	utils.WriteJSON(w, http.StatusOK, pledge)
}

func (s *server) exportAidRequestsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, utils.ShelterUpdateRoles)
	if !ok {
		return
	}

	filter := models.AidRequestFilter{IncludePrivate: true}
	if status := utils.NormalizeToken(r.URL.Query().Get("status")); status != "" {
		if !utils.AllowedAidRequestStatuses[status] {
			utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status is not supported")
			return
		}
		filter.Status = status
	}
	aidRequests := s.store.ListAidRequests(filter)
	log.Printf("INFO shelter-service aid_request_export actor=%s count=%d status=%s", ctx.ActorUserID, len(aidRequests), filter.Status)

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="nadaa-aid-requests.csv"`)
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "id,title,category,priority,status,district,receivingOrganization,quantityNeeded,quantityPledged,quantityUnit,pledgeCount,neededBy"); err != nil {
		log.Printf("ERROR shelter-service aid_request_export header_failed error=%v", err)
		return
	}
	for _, request := range aidRequests {
		if _, err := fmt.Fprintf(
			w,
			"%s,%q,%s,%s,%s,%q,%q,%d,%d,%s,%d,%s\n",
			request.ID,
			request.Title,
			request.Category,
			request.Priority,
			request.Status,
			request.District,
			request.ReceivingOrganization,
			request.QuantityNeeded,
			request.QuantityPledged,
			request.QuantityUnit,
			len(request.Pledges),
			request.NeededBy.Format(time.RFC3339),
		); err != nil {
			log.Printf("ERROR shelter-service aid_request_export row_failed id=%s error=%v", request.ID, err)
			return
		}
	}
}

func parseAidRequestFilter(w http.ResponseWriter, r *http.Request) (models.AidRequestFilter, bool) {
	query := r.URL.Query()
	filter := models.AidRequestFilter{
		Category:       utils.NormalizeToken(query.Get("category")),
		District:       utils.NormalizeToken(query.Get("district")),
		IncludePrivate: strings.TrimSpace(strings.ToLower(query.Get("includePrivate"))) == "true",
		Priority:       utils.NormalizeToken(query.Get("priority")),
		RadiusMeters:   utils.NearbySearchMeters,
		Region:         utils.NormalizeToken(query.Get("region")),
		Status:         utils.NormalizeToken(query.Get("status")),
	}
	if filter.Category != "" && !utils.AllowedAidRequestCategories[filter.Category] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_category", "category is not supported")
		return filter, false
	}
	if filter.Priority != "" && !utils.AllowedAidRequestPriorities[filter.Priority] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_priority", "priority is not supported")
		return filter, false
	}
	if filter.Status != "" && !utils.AllowedAidRequestStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status is not supported")
		return filter, false
	}
	if !filter.IncludePrivate && filter.Status != "" && !utils.PublicAidRequestStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "private_status_requires_authority", "private aid request statuses require includePrivate=true and authority context")
		return filter, false
	}
	if filter.IncludePrivate {
		if _, ok := requireAuthority(w, r, utils.ShelterUpdateRoles); !ok {
			return filter, false
		}
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
