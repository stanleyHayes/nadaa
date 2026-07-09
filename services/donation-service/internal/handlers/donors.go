package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

var allowedDonorTypes = map[string]bool{
	"individual":   true,
	"organization": true,
	"ngo":          true,
	"government":   true,
	"other":        true,
}

var allowedDonorStatuses = map[string]bool{
	"active":   true,
	"inactive": true,
}

func (s *Server) listDonorsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	filter := models.DonorFilter{
		Type:  utils.NormalizeToken(r.URL.Query().Get("type")),
		Query: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q"))),
	}
	if filter.Type != "" && !allowedDonorTypes[filter.Type] {
		log.Printf("WARN donation-service donor_list invalid_type actor=%s type=%s", ctx.ActorUserID, filter.Type)
		utils.WriteError(w, http.StatusBadRequest, "invalid_type", "type must be individual, organization, ngo, government, or other")
		return
	}

	donors := s.store.ListDonors(filter)
	log.Printf("INFO donation-service donor_list count=%d actor=%s type=%s q=%t", len(donors), ctx.ActorUserID, filter.Type, filter.Query != "")
	utils.WriteJSON(w, http.StatusOK, models.DonorListResponse{Donors: donors, GeneratedAt: s.now().UTC()})
}

func (s *Server) createDonorHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreateDonorRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service donor_create invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreateDonor(request)
	if code != "" {
		log.Printf("WARN donation-service donor_create validation_failed code=%s", code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	createdBy := "public"
	if ctx, ok := authorityContextFromRequest(r); ok {
		createdBy = ctx.ActorUserID
	}

	donor := s.store.CreateDonor(normalized, createdBy, s.now().UTC())
	log.Printf("INFO donation-service donor_create completed id=%s reference=%s createdBy=%s", donor.ID, donor.Reference, donor.CreatedBy)
	utils.WriteJSON(w, http.StatusCreated, donor)
}

func (s *Server) getDonorHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r); !ok {
		return
	}

	donor, ok := s.store.GetDonor(r.PathValue("id"))
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "donor was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, donor)
}

func (s *Server) updateDonorHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.UpdateDonorRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service donor_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateDonor(request)
	if code != "" {
		log.Printf("WARN donation-service donor_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	donor, code, message := s.store.UpdateDonor(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service donor_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service donor_update completed id=%s actor=%s status=%s", donor.ID, ctx.ActorUserID, donor.Status)
	utils.WriteJSON(w, http.StatusOK, donor)
}

func normalizeCreateDonor(request models.CreateDonorRequest) (models.CreateDonorRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = utils.NormalizeToken(request.Type)
	request.ContactName = strings.TrimSpace(request.ContactName)
	request.ContactEmail = strings.TrimSpace(strings.ToLower(request.ContactEmail))
	request.ContactPhone = strings.TrimSpace(request.ContactPhone)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Notes = strings.TrimSpace(request.Notes)

	if request.Name == "" || len(request.Name) > 200 || utils.UnsafeText(request.Name) {
		return request, "invalid_name", "name is required and must be 200 safe characters or fewer"
	}
	if !allowedDonorTypes[request.Type] {
		return request, "invalid_type", "type must be individual, organization, ngo, government, or other"
	}
	if len(request.ContactName) > 200 || utils.UnsafeText(request.ContactName) {
		return request, "invalid_contact_name", "contactName must be 200 safe characters or fewer"
	}
	if request.ContactEmail != "" && !utils.ValidEmail(request.ContactEmail) {
		return request, "invalid_contact_email", "contactEmail must be a valid email address"
	}
	if len(request.ContactPhone) > 50 || utils.UnsafeText(request.ContactPhone) {
		return request, "invalid_contact_phone", "contactPhone must be 50 safe characters or fewer"
	}
	if len(request.Region) > 100 || utils.UnsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || utils.UnsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if request.MonetaryPledgeGhs < 0 {
		return request, "invalid_monetary_pledge", "monetaryPledgeGhs must be zero or greater"
	}
	if len(request.Notes) > 700 || utils.UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 700 safe characters or fewer"
	}

	cleaned := make([]string, 0, len(request.ItemsOffered))
	for _, item := range request.ItemsOffered {
		item = strings.TrimSpace(item)
		if item != "" && !utils.UnsafeText(item) {
			cleaned = append(cleaned, item)
		}
	}
	request.ItemsOffered = cleaned
	return request, "", ""
}

func normalizeUpdateDonor(request models.UpdateDonorRequest) (models.UpdateDonorRequest, string, string) {
	request.Status = utils.NormalizeToken(request.Status)
	request.Notes = strings.TrimSpace(request.Notes)

	if request.Status == "" && request.Notes == "" {
		return request, "no_changes", "at least one of status or notes must be supplied"
	}
	if request.Status != "" && !allowedDonorStatuses[request.Status] {
		return request, "invalid_status", "status must be active or inactive"
	}
	if len(request.Notes) > 700 || utils.UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 700 safe characters or fewer"
	}
	return request, "", ""
}
