package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

var allowedPledgeStatuses = map[string]bool{
	"pledged":   true,
	"delivered": true,
	"cancelled": true,
}

func (s *Server) listRequestPledgesHandler(w http.ResponseWriter, r *http.Request) {
	pledges := s.store.ListPledgesForRequest(r.PathValue("id"))
	log.Printf("INFO donation-service pledge_list_for_request aidRequestId=%s count=%d", r.PathValue("id"), len(pledges))
	utils.WriteJSON(w, http.StatusOK, models.PledgeListResponse{Pledges: pledges, GeneratedAt: s.now().UTC()})
}

func (s *Server) createPledgeHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreatePledgeRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service pledge_create invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreatePledge(request)
	if code != "" {
		log.Printf("WARN donation-service pledge_create validation_failed code=%s", code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.CreatePledge(r.PathValue("id"), normalized, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service pledge_create failed aidRequestId=%s code=%s", r.PathValue("id"), code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service pledge_create completed id=%s reference=%s aidRequestId=%s donorId=%s quantity=%d", pledge.ID, pledge.Reference, pledge.AidRequestID, pledge.DonorID, pledge.QuantityPledged)
	utils.WriteJSON(w, http.StatusCreated, pledge)
}

func (s *Server) listPledgesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	filter := models.PledgeFilter{Status: utils.NormalizeToken(r.URL.Query().Get("status"))}
	if filter.Status != "" && !allowedPledgeStatuses[filter.Status] {
		log.Printf("WARN donation-service pledge_list invalid_status actor=%s status=%s", ctx.ActorUserID, filter.Status)
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be pledged, delivered, or cancelled")
		return
	}

	pledges := s.store.ListPledges(filter)
	log.Printf("INFO donation-service pledge_list count=%d actor=%s status=%s", len(pledges), ctx.ActorUserID, filter.Status)
	utils.WriteJSON(w, http.StatusOK, models.PledgeListResponse{Pledges: pledges, GeneratedAt: s.now().UTC()})
}

func (s *Server) updatePledgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.UpdatePledgeRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service pledge_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdatePledge(request)
	if code != "" {
		log.Printf("WARN donation-service pledge_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.UpdatePledge(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service pledge_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service pledge_update completed id=%s actor=%s status=%s delivered=%d", pledge.ID, ctx.ActorUserID, pledge.Status, pledge.QuantityDelivered)
	utils.WriteJSON(w, http.StatusOK, pledge)
}

func (s *Server) allocatePledgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.AllocateRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service pledge_allocate invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAllocate(request)
	if code != "" {
		log.Printf("WARN donation-service pledge_allocate validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.AllocatePledge(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service pledge_allocate failed aidRequestId=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service pledge_allocate completed id=%s actor=%s quantity=%d", pledge.ID, ctx.ActorUserID, pledge.QuantityDelivered)
	utils.WriteJSON(w, http.StatusOK, pledge)
}

func normalizeCreatePledge(request models.CreatePledgeRequest) (models.CreatePledgeRequest, string, string) {
	request.DonorID = strings.TrimSpace(request.DonorID)
	request.DonorName = strings.TrimSpace(request.DonorName)
	request.ContactEmail = strings.TrimSpace(strings.ToLower(request.ContactEmail))
	request.ContactPhone = strings.TrimSpace(request.ContactPhone)
	request.DeliveryNote = strings.TrimSpace(request.DeliveryNote)

	if request.DonorID == "" {
		return request, "invalid_donor_id", "donorId is required"
	}
	if request.QuantityPledged <= 0 {
		return request, "invalid_quantity_pledged", "quantityPledged must be greater than zero"
	}
	if request.ContactEmail != "" && !utils.ValidEmail(request.ContactEmail) {
		return request, "invalid_contact_email", "contactEmail must be a valid email address"
	}
	if len(request.ContactPhone) > 50 || utils.UnsafeText(request.ContactPhone) {
		return request, "invalid_contact_phone", "contactPhone must be 50 safe characters or fewer"
	}
	if len(request.DeliveryNote) > 500 || utils.UnsafeText(request.DeliveryNote) {
		return request, "invalid_delivery_note", "deliveryNote must be 500 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeUpdatePledge(request models.UpdatePledgeRequest) (models.UpdatePledgeRequest, string, string) {
	request.Status = utils.NormalizeToken(request.Status)
	request.DeliveryNote = strings.TrimSpace(request.DeliveryNote)

	if request.Status == "" && request.QuantityDelivered == 0 && request.DeliveryNote == "" {
		return request, "no_changes", "at least one field must be supplied"
	}
	if request.Status != "" && !allowedPledgeStatuses[request.Status] {
		return request, "invalid_status", "status must be pledged, delivered, or cancelled"
	}
	if request.QuantityDelivered < 0 {
		return request, "invalid_quantity_delivered", "quantityDelivered must be zero or greater"
	}
	if len(request.DeliveryNote) > 500 || utils.UnsafeText(request.DeliveryNote) {
		return request, "invalid_delivery_note", "deliveryNote must be 500 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeAllocate(request models.AllocateRequest) (models.AllocateRequest, string, string) {
	request.PledgeID = strings.TrimSpace(request.PledgeID)

	if request.PledgeID == "" {
		return request, "invalid_pledge_id", "pledgeId is required"
	}
	if request.Quantity <= 0 {
		return request, "invalid_quantity", "quantity must be greater than zero"
	}
	return request, "", ""
}
