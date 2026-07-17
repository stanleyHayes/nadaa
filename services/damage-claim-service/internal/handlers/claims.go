package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
	"verified": true,
	"rejected": true,
}

func (s *Server) createClaimHandler(w http.ResponseWriter, r *http.Request) {
	// The create endpoint is public; cap the request body at ~1 MiB so a
	// single unauthenticated POST cannot exhaust memory.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

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

	incidentRef, incidentLocation := s.enrichIncident(r.Context(), strings.TrimSpace(r.Header.Get("Authorization")), normalized.IncidentID)
	claim := s.store.Create(normalized, incidentRef, incidentLocation, s.now().UTC())
	log.Printf("INFO damage-claim-service claim_create completed id=%s reference=%s incidentId=%s", claim.ID, claim.Reference, utils.SafeLogValue(claim.IncidentID))
	utils.WriteJSON(w, http.StatusCreated, claim)
}

func (s *Server) listClaimsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, authorityRoles)
	if !ok {
		return
	}

	filter := parseListClaimsFilter(r)
	claims := s.store.List(filter)
	// #nosec G706 -- filter values are sanitized with utils.SafeLogValue.
	log.Printf("INFO damage-claim-service claim_list count=%d status=%s verificationStatus=%s incidentId=%s q=%s actor=%s", len(claims), utils.SafeLogValue(filter.Status), utils.SafeLogValue(filter.VerificationStatus), utils.SafeLogValue(filter.IncidentID), utils.SafeLogValue(filter.Query), ctx.ActorUserID)
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
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_update invalid_json id=%s actor=%s error=%v", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateClaim(request)
	if code != "" {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_update validation_failed id=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	claim, ok := s.store.Update(r.PathValue("id"), normalized, s.now().UTC())
	if !ok {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_update failed id=%s actor=%s code=not_found", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID)
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
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_verify invalid_json id=%s actor=%s error=%v", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeVerifyClaim(request)
	if code != "" {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_verify validation_failed id=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	claim, errCode := s.store.Verify(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if errCode != "" {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_verify failed id=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, errCode)
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
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_close invalid_json id=%s actor=%s error=%v", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	request.Reason = strings.TrimSpace(request.Reason)
	if request.Reason == "" {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_close validation_failed id=%s actor=%s code=missing_reason", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID)
		utils.WriteError(w, http.StatusBadRequest, "missing_reason", "close reason is required")
		return
	}
	if len(request.Reason) > 500 || utils.UnsafeText(request.Reason) {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_close validation_failed id=%s actor=%s code=invalid_reason", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID)
		utils.WriteError(w, http.StatusBadRequest, "invalid_reason", "reason must be 500 safe characters or fewer")
		return
	}

	claim, errCode := s.store.Close(r.PathValue("id"), request.Reason, ctx.ActorUserID, s.now().UTC())
	if errCode != "" {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_close failed id=%s actor=%s code=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, errCode)
		status := http.StatusBadRequest
		msg := "claim is already closed"
		if errCode == "not_found" {
			status = http.StatusNotFound
			msg = "claim was not found"
		}
		utils.WriteError(w, status, errCode, msg)
		return
	}
	log.Printf("INFO damage-claim-service claim_close completed id=%s reference=%s actor=%s", claim.ID, claim.Reference, ctx.ActorUserID)
	utils.WriteJSON(w, http.StatusOK, claim)
}

func (s *Server) enrichIncident(ctx context.Context, authorization, incidentID string) (string, string) {
	incidentID = strings.TrimSpace(incidentID)
	if incidentID == "" {
		return "", ""
	}

	endpoint := fmt.Sprintf("%s/incidents/%s", s.incidentServiceURL, url.PathEscape(incidentID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		log.Printf("WARN damage-claim-service incident_lookup_failed incidentId=%s error=%v", utils.SafeLogValue(incidentID), err)
		return "", ""
	}
	// Forward the caller's credentials so incident-service can authorize the
	// lookup; never fabricate X-NADAA-Actor-* headers outbound. Public claim
	// intake sends no caller token, so fall back to the internal
	// service-to-service token when one is configured.
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	} else if s.config.InternalServiceToken != "" {
		req.Header.Set("X-NADAA-Service-Token", s.config.InternalServiceToken)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("WARN damage-claim-service incident_lookup_failed incidentId=%s error=%v", utils.SafeLogValue(incidentID), err)
		return "", ""
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Printf("WARN damage-claim-service incident_lookup_failed incidentId=%s status=%d body=%s", utils.SafeLogValue(incidentID), resp.StatusCode, utils.SafeLogValue(strings.TrimSpace(string(body))))
		return "", ""
	}

	var lookup models.IncidentLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&lookup); err != nil {
		log.Printf("WARN damage-claim-service incident_lookup_decode_failed incidentId=%s error=%v", utils.SafeLogValue(incidentID), err)
		return "", ""
	}

	reference := strings.TrimSpace(lookup.Reference)
	location := ""
	if lookup.Location != nil {
		location = formatIncidentLocation(*lookup.Location)
	}
	if reference == "" && location == "" {
		log.Printf("WARN damage-claim-service incident_lookup_empty incidentId=%s", utils.SafeLogValue(incidentID))
		return "", ""
	}
	return reference, location
}

// formatIncidentLocation renders incident-service's coordinate-only location
// as a "lat,lng" string for the claim's incidentLocation field.
func formatIncidentLocation(coords models.IncidentCoordinates) string {
	return strconv.FormatFloat(coords.Lat, 'f', -1, 64) + "," + strconv.FormatFloat(coords.Lng, 'f', -1, 64)
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

	if len(request.DamagePhotos) > 20 {
		return request, "too_many_damage_photos", "damagePhotos must contain 20 or fewer URLs"
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
		if len(request.DamagePhotos) > 20 {
			return request, "too_many_damage_photos", "damagePhotos must contain 20 or fewer URLs"
		}
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
