package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

var allowedRequestStatuses = map[string]bool{
	"open":                true,
	"partially_fulfilled": true,
	"fulfilled":           true,
	"closed":              true,
}

var allowedPriorities = map[string]bool{
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

var allowedCatalogCategories = map[string]bool{
	"food":       true,
	"water":      true,
	"medical":    true,
	"shelter":    true,
	"sanitation": true,
}

func (s *Server) listAidRequestsHandler(w http.ResponseWriter, r *http.Request) {
	filter := models.AidRequestFilter{
		Status:   utils.NormalizeToken(r.URL.Query().Get("status")),
		Category: utils.NormalizeToken(r.URL.Query().Get("category")),
		Region:   strings.TrimSpace(strings.ToLower(r.URL.Query().Get("region"))),
		Priority: utils.NormalizeToken(r.URL.Query().Get("priority")),
	}
	if filter.Status != "" && !allowedRequestStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be open, partially_fulfilled, fulfilled, or closed")
		return
	}
	if filter.Priority != "" && !allowedPriorities[filter.Priority] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_priority", "priority must be low, medium, high, or critical")
		return
	}

	requests := s.store.ListAidRequests(filter)
	log.Printf("INFO donation-service aid_request_list count=%d status=%s category=%s region=%s priority=%s", len(requests), filter.Status, filter.Category, filter.Region, filter.Priority)
	utils.WriteJSON(w, http.StatusOK, models.AidRequestListResponse{Requests: requests, GeneratedAt: s.now().UTC()})
}

func (s *Server) createAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.CreateAidRequestRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service aid_request_create invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreateAidRequest(request)
	if code != "" {
		log.Printf("WARN donation-service aid_request_create validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	req := s.store.CreateAidRequest(normalized, ctx.ActorUserID, s.now().UTC())
	log.Printf("INFO donation-service aid_request_create completed id=%s reference=%s actor=%s", req.ID, req.Reference, ctx.ActorUserID)
	utils.WriteJSON(w, http.StatusCreated, req)
}

func (s *Server) getAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := s.store.GetAidRequest(r.PathValue("id"))
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "aid request was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, req)
}

func (s *Server) updateAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.UpdateAidRequestRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service aid_request_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateAidRequest(request)
	if code != "" {
		log.Printf("WARN donation-service aid_request_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	req, code, message := s.store.UpdateAidRequest(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service aid_request_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service aid_request_update completed id=%s actor=%s status=%s fulfilled=%d/%d", req.ID, ctx.ActorUserID, req.Status, req.QuantityFulfilled, req.QuantityNeeded)
	utils.WriteJSON(w, http.StatusOK, req)
}

func normalizeCreateAidRequest(request models.CreateAidRequestRequest) (models.CreateAidRequestRequest, string, string) {
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	request.Category = utils.NormalizeToken(request.Category)
	request.ItemCode = strings.TrimSpace(request.ItemCode)
	request.Unit = strings.TrimSpace(request.Unit)
	request.Priority = utils.NormalizeToken(request.Priority)
	request.LocationLabel = strings.TrimSpace(request.LocationLabel)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)

	if request.Title == "" || len(request.Title) > 200 || utils.UnsafeText(request.Title) {
		return request, "invalid_title", "title is required and must be 200 safe characters or fewer"
	}
	if len(request.Description) > 1000 || utils.UnsafeText(request.Description) {
		return request, "invalid_description", "description must be 1000 safe characters or fewer"
	}
	if request.Category == "" || !allowedCatalogCategories[request.Category] {
		return request, "invalid_category", "category must be food, water, medical, shelter, or sanitation"
	}
	if request.ItemCode == "" || len(request.ItemCode) > 100 || utils.UnsafeText(request.ItemCode) {
		return request, "invalid_item_code", "itemCode is required and must be 100 safe characters or fewer"
	}
	if request.QuantityNeeded <= 0 {
		return request, "invalid_quantity_needed", "quantityNeeded must be greater than zero"
	}
	if request.Unit == "" || len(request.Unit) > 50 || utils.UnsafeText(request.Unit) {
		return request, "invalid_unit", "unit is required and must be 50 safe characters or fewer"
	}
	if !allowedPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, medium, high, or critical"
	}
	if len(request.LocationLabel) > 200 || utils.UnsafeText(request.LocationLabel) {
		return request, "invalid_location_label", "locationLabel must be 200 safe characters or fewer"
	}
	if request.Region == "" || len(request.Region) > 100 || utils.UnsafeText(request.Region) {
		return request, "invalid_region", "region is required and must be 100 safe characters or fewer"
	}
	if request.District == "" || len(request.District) > 100 || utils.UnsafeText(request.District) {
		return request, "invalid_district", "district is required and must be 100 safe characters or fewer"
	}
	if request.BeneficiaryCount < 0 {
		return request, "invalid_beneficiary_count", "beneficiaryCount must be zero or greater"
	}
	return request, "", ""
}

func normalizeUpdateAidRequest(request models.UpdateAidRequestRequest) (models.UpdateAidRequestRequest, string, string) {
	request.Status = utils.NormalizeToken(request.Status)

	if request.Status == "" && request.QuantityNeeded == 0 {
		return request, "no_changes", "at least one of status or quantityNeeded must be supplied"
	}
	if request.Status != "" && !allowedRequestStatuses[request.Status] {
		return request, "invalid_status", "status must be open, partially_fulfilled, fulfilled, or closed"
	}
	if request.QuantityNeeded < 0 {
		return request, "invalid_quantity_needed", "quantityNeeded must be zero or greater"
	}
	return request, "", ""
}
