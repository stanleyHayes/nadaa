package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"
)

var authorityRoles = map[string]bool{
	"system_admin":      true,
	"nadmo_officer":     true,
	"district_officer":  true,
	"dispatcher":        true,
	"police":            true,
	"insurance_officer": true,
	"fire":              true,
	"ambulance":         true,
}

var allowedDamageTypes = map[string]bool{
	"structural": true,
	"flood":      true,
	"fire":       true,
	"vehicle":    true,
	"other":      true,
}

var allowedVerificationStatuses = map[string]bool{
	"pending":  true,
	"verified": true,
	"rejected": true,
}

func (s *Server) createClaimHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreateClaimRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN damage-claim-service claim_create invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreateClaim(request)
	if code != "" {
		log.Printf("WARN damage-claim-service claim_create validation_failed code=%s", code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	incidentRef, incidentLocation := s.enrichIncident(r.Context(), normalized.IncidentID)
	claim := s.store.Create(normalized, incidentRef, incidentLocation, s.now().UTC())
	log.Printf("INFO damage-claim-service claim_create completed id=%s reference=%s incidentId=%s", claim.ID, claim.Reference, claim.IncidentID)
	utils.WriteJSON(w, http.StatusCreated, claim)
}

func (s *Server) listClaimsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, authorityRoles)
	if !ok {
		return
	}

	filter := parseListClaimsFilter(r)
	claims := s.store.List(filter)
	log.Printf("INFO damage-claim-service claim_list count=%d status=%s verificationStatus=%s incidentId=%s q=%s actor=%s", len(claims), filter.Status, filter.VerificationStatus, filter.IncidentID, filter.Query, ctx.ActorUserID)
	utils.WriteJSON(w, http.StatusOK, models.ClaimListResponse{Claims: claims, GeneratedAt: s.now().UTC()})
}

func (s *Server) getClaimHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAuthority(w, r, authorityRoles); !ok {
		return
	}

	claim, ok := s.store.Get(r.PathValue("id"))
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "claim was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, claim)
}

func (s *Server) updateClaimHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, authorityRoles)
	if !ok {
		return
	}

	var request models.UpdateClaimRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN damage-claim-service claim_update invalid_json id=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateClaim(request)
	if code != "" {
		log.Printf("WARN damage-claim-service claim_update validation_failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	claim, ok := s.store.Update(r.PathValue("id"), normalized, s.now().UTC())
	if !ok {
		log.Printf("WARN damage-claim-service claim_update failed id=%s actor=%s code=not_found", r.PathValue("id"), ctx.ActorUserID)
		utils.WriteError(w, http.StatusNotFound, "not_found", "claim was not found")
		return
	}
	log.Printf("INFO damage-claim-service claim_update completed id=%s actor=%s", claim.ID, ctx.ActorUserID)
	utils.WriteJSON(w, http.StatusOK, claim)
}

func (s *Server) verifyClaimHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, authorityRoles)
	if !ok {
		return
	}

	var request models.VerifyClaimRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN damage-claim-service claim_verify invalid_json id=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeVerifyClaim(request)
	if code != "" {
		log.Printf("WARN damage-claim-service claim_verify validation_failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	claim, errCode := s.store.Verify(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if errCode != "" {
		log.Printf("WARN damage-claim-service claim_verify failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, errCode)
		status := http.StatusBadRequest
		msg := "claim can only be verified or rejected from pending status"
		if errCode == "not_found" {
			status = http.StatusNotFound
			msg = "claim was not found"
		}
		utils.WriteError(w, status, errCode, msg)
		return
	}
	log.Printf("INFO damage-claim-service claim_verify completed id=%s reference=%s verificationStatus=%s actor=%s", claim.ID, claim.Reference, claim.VerificationStatus, ctx.ActorUserID)
	utils.WriteJSON(w, http.StatusOK, claim)
}

func (s *Server) closeClaimHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, authorityRoles)
	if !ok {
		return
	}

	var request models.CloseClaimRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN damage-claim-service claim_close invalid_json id=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.Reason = strings.TrimSpace(request.Reason)
	if request.Reason == "" {
		log.Printf("WARN damage-claim-service claim_close validation_failed id=%s actor=%s code=missing_reason", r.PathValue("id"), ctx.ActorUserID)
		utils.WriteError(w, http.StatusBadRequest, "missing_reason", "close reason is required")
		return
	}
	if len(request.Reason) > 500 || utils.UnsafeText(request.Reason) {
		log.Printf("WARN damage-claim-service claim_close validation_failed id=%s actor=%s code=invalid_reason", r.PathValue("id"), ctx.ActorUserID)
		utils.WriteError(w, http.StatusBadRequest, "invalid_reason", "reason must be 500 safe characters or fewer")
		return
	}

	claim, ok := s.store.Close(r.PathValue("id"), request.Reason, ctx.ActorUserID, s.now().UTC())
	if !ok {
		log.Printf("WARN damage-claim-service claim_close failed id=%s actor=%s code=not_found", r.PathValue("id"), ctx.ActorUserID)
		utils.WriteError(w, http.StatusNotFound, "not_found", "claim was not found")
		return
	}
	log.Printf("INFO damage-claim-service claim_close completed id=%s reference=%s actor=%s", claim.ID, claim.Reference, ctx.ActorUserID)
	utils.WriteJSON(w, http.StatusOK, claim)
}

func (s *Server) enrichIncident(ctx context.Context, incidentID string) (string, string) {
	incidentID = strings.TrimSpace(incidentID)
	if incidentID == "" {
		return "", ""
	}

	endpoint := fmt.Sprintf("%s/incidents/%s", s.incidentServiceURL, url.PathEscape(incidentID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("WARN damage-claim-service incident_lookup_failed incidentId=%s error=%v", incidentID, err)
		return "", ""
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("WARN damage-claim-service incident_lookup_failed incidentId=%s error=%v", incidentID, err)
		return "", ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("WARN damage-claim-service incident_lookup_failed incidentId=%s status=%d body=%s", incidentID, resp.StatusCode, strings.TrimSpace(string(body)))
		return "", ""
	}

	var lookup models.IncidentLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&lookup); err != nil {
		log.Printf("WARN damage-claim-service incident_lookup_decode_failed incidentId=%s error=%v", incidentID, err)
		return "", ""
	}
	return lookup.Reference, lookup.Location.Address
}

func normalizeCreateClaim(request models.CreateClaimRequest) (models.CreateClaimRequest, string, string) {
	request.IncidentID = strings.TrimSpace(request.IncidentID)
	request.Reporter.Name = strings.TrimSpace(request.Reporter.Name)
	request.Reporter.Phone = strings.TrimSpace(request.Reporter.Phone)
	request.Reporter.Email = strings.TrimSpace(request.Reporter.Email)
	request.Reporter.UserID = strings.TrimSpace(request.Reporter.UserID)
	request.DamageType = utils.NormalizeToken(request.DamageType)
	request.DamageDescription = strings.TrimSpace(request.DamageDescription)
	request.EstimatedLossAmount = strings.TrimSpace(request.EstimatedLossAmount)
	request.Location.Address = strings.TrimSpace(request.Location.Address)

	if request.Reporter.Name == "" || len(request.Reporter.Name) > 200 || utils.UnsafeText(request.Reporter.Name) {
		return request, "invalid_reporter_name", "reporter name is required and must be 200 safe characters or fewer"
	}
	if request.Reporter.Phone == "" {
		return request, "invalid_reporter_phone", "reporter phone is required"
	}
	if len(request.Reporter.Phone) > 50 || utils.UnsafeText(request.Reporter.Phone) {
		return request, "invalid_reporter_phone", "reporter phone must be 50 safe characters or fewer"
	}
	if request.Reporter.Email != "" && (len(request.Reporter.Email) > 200 || !utils.ValidEmail(request.Reporter.Email)) {
		return request, "invalid_reporter_email", "reporter email must be a valid address"
	}
	if !allowedDamageTypes[request.DamageType] {
		return request, "invalid_damage_type", "damageType must be structural, flood, fire, vehicle, or other"
	}
	if request.DamageDescription == "" || len(request.DamageDescription) > 2000 || utils.UnsafeText(request.DamageDescription) {
		return request, "invalid_damage_description", "damageDescription is required and must be 2000 safe characters or fewer"
	}
	if request.EstimatedLossAmount == "" || len(request.EstimatedLossAmount) > 50 || !utils.ValidDecimal(request.EstimatedLossAmount) {
		return request, "invalid_estimated_loss_amount", "estimatedLossAmount is required and must be a valid decimal string"
	}
	if !utils.ValidCoordinates(request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	if len(request.Location.Address) > 300 || utils.UnsafeText(request.Location.Address) {
		return request, "invalid_location_address", "location address must be 300 safe characters or fewer"
	}
	if !request.PrivacyConsent {
		return request, "privacy_consent_required", "privacyConsent must be true to submit a damage claim"
	}

	for i, photo := range request.DamagePhotos {
		request.DamagePhotos[i] = strings.TrimSpace(photo)
		if request.DamagePhotos[i] == "" || len(request.DamagePhotos[i]) > 500 || utils.UnsafeText(request.DamagePhotos[i]) {
			return request, "invalid_damage_photo", "damage photo URLs must be non-empty and 500 safe characters or fewer"
		}
	}

	return request, "", ""
}

func normalizeUpdateClaim(request models.UpdateClaimRequest) (models.UpdateClaimRequest, string, string) {
	if request.DamageDescription != nil {
		*request.DamageDescription = strings.TrimSpace(*request.DamageDescription)
		if *request.DamageDescription == "" || len(*request.DamageDescription) > 2000 || utils.UnsafeText(*request.DamageDescription) {
			return request, "invalid_damage_description", "damageDescription must be 1-2000 safe characters"
		}
	}
	if request.EstimatedLossAmount != nil {
		*request.EstimatedLossAmount = strings.TrimSpace(*request.EstimatedLossAmount)
		if *request.EstimatedLossAmount == "" || len(*request.EstimatedLossAmount) > 50 || !utils.ValidDecimal(*request.EstimatedLossAmount) {
			return request, "invalid_estimated_loss_amount", "estimatedLossAmount must be a valid decimal string"
		}
	}
	if request.DamagePhotos != nil {
		for i, photo := range request.DamagePhotos {
			request.DamagePhotos[i] = strings.TrimSpace(photo)
			if request.DamagePhotos[i] == "" || len(request.DamagePhotos[i]) > 500 || utils.UnsafeText(request.DamagePhotos[i]) {
				return request, "invalid_damage_photo", "damage photo URLs must be non-empty and 500 safe characters or fewer"
			}
		}
	}
	if request.DamageDescription == nil && request.EstimatedLossAmount == nil && request.DamagePhotos == nil {
		return request, "no_changes", "at least one of damageDescription, estimatedLossAmount, or damagePhotos must be supplied"
	}
	return request, "", ""
}

func normalizeVerifyClaim(request models.VerifyClaimRequest) (models.VerifyClaimRequest, string, string) {
	request.VerificationStatus = utils.NormalizeToken(request.VerificationStatus)
	request.Notes = strings.TrimSpace(request.Notes)

	if !allowedVerificationStatuses[request.VerificationStatus] {
		return request, "invalid_verification_status", "verificationStatus must be verified or rejected"
	}
	if len(request.Notes) > 1000 || utils.UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 1000 safe characters or fewer"
	}
	return request, "", ""
}

func parseListClaimsFilter(r *http.Request) models.ListClaimsFilter {
	return models.ListClaimsFilter{
		Status:             utils.NormalizeToken(r.URL.Query().Get("status")),
		VerificationStatus: utils.NormalizeToken(r.URL.Query().Get("verificationStatus")),
		IncidentID:         strings.TrimSpace(r.URL.Query().Get("incidentId")),
		Query:              strings.TrimSpace(r.URL.Query().Get("q")),
	}
}
