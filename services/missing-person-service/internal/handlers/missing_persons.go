package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/utils"
)

var allowedStatuses = map[string]bool{
	"pending_review": true,
	"active":         true,
	"located":        true,
	"reunited":       true,
	"closed":         true,
	"rejected":       true,
}

var allowedGenderValues = map[string]bool{
	"":           true,
	"female":     true,
	"male":       true,
	"non_binary": true,
	"unknown":    true,
}

var allowedClosureTypes = map[string]bool{
	"reunited":     true,
	"located_safe": true,
	"duplicate":    true,
	"withdrawn":    true,
	"deceased":     true,
	"other":        true,
}

func (s *Server) listPublicMissingPersonsHandler(w http.ResponseWriter, r *http.Request) {
	filter, code, message := parseMissingPersonFilter(r, false)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	records := s.store.ListPublicMissingPersons(filter)
	log.Printf("INFO missing-person-service public_list count=%d status=%s district=%s queryPresent=%t", len(records), filter.Status, filter.District, filter.Query != "")
	utils.WriteJSON(w, http.StatusOK, models.PublicMissingPersonListResponse{Records: records, GeneratedAt: s.now().UTC()})
}

func (s *Server) createMissingPersonHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreateMissingPersonRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN missing-person-service intake invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized, code, message := normalizeCreateMissingPerson(request)
	if code != "" {
		log.Printf("WARN missing-person-service intake validation_failed code=%s", code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	record := s.store.CreateMissingPerson(normalized, "public", s.now().UTC())
	log.Printf("INFO missing-person-service intake created id=%s reference=%s status=%s visibility=%s", record.ID, record.Reference, record.Status, record.PublicVisibility)
	utils.WriteJSON(w, http.StatusCreated, record)
}

func (s *Server) getPublicMissingPersonHandler(w http.ResponseWriter, r *http.Request) {
	record, ok := s.store.GetPublicMissingPerson(r.PathValue("id"))
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "approved public missing person record was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, record)
}

func (s *Server) listAuthorityMissingPersonsHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := requireAuthority(w, r)
	if !ok {
		return
	}
	filter, code, message := parseMissingPersonFilter(r, true)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	records := s.store.ListMissingPersons(filter)
	log.Printf("INFO missing-person-service authority_list count=%d status=%s district=%s queryPresent=%t", len(records), filter.Status, filter.District, filter.Query != "")
	utils.WriteJSON(w, http.StatusOK, models.MissingPersonListResponse{Records: records, GeneratedAt: s.now().UTC()})
}

func (s *Server) getAuthorityMissingPersonHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := requireAuthority(w, r)
	if !ok {
		return
	}
	record, found := s.store.GetMissingPerson(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "missing person record was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, record)
}

func (s *Server) reviewMissingPersonHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}
	var request models.ReviewMissingPersonRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN missing-person-service review invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized, code, message := normalizeReviewMissingPerson(request)
	if code != "" {
		log.Printf("WARN missing-person-service review validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	record, code, message := s.store.ReviewMissingPerson(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN missing-person-service review failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO missing-person-service review completed id=%s actor=%s reviewStatus=%s visibility=%s status=%s", record.ID, ctx.ActorUserID, record.ReviewStatus, record.PublicVisibility, record.Status)
	utils.WriteJSON(w, http.StatusOK, record)
}

func (s *Server) closeMissingPersonHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}
	var request models.CloseMissingPersonRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN missing-person-service close invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized, code, message := normalizeCloseMissingPerson(request)
	if code != "" {
		log.Printf("WARN missing-person-service close validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	record, code, message := s.store.CloseMissingPerson(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN missing-person-service close failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO missing-person-service close completed id=%s actor=%s closureType=%s status=%s", record.ID, ctx.ActorUserID, record.ClosureType, record.Status)
	utils.WriteJSON(w, http.StatusOK, record)
}

func (s *Server) listAuditHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}
	entries := s.store.ListAudit(r.PathValue("id"))
	log.Printf("INFO missing-person-service audit_list id=%s actor=%s count=%d", r.PathValue("id"), ctx.ActorUserID, len(entries))
	utils.WriteJSON(w, http.StatusOK, models.MissingPersonAuditResponse{Entries: entries, GeneratedAt: s.now().UTC()})
}

func parseMissingPersonFilter(r *http.Request, includePrivate bool) (models.MissingPersonFilter, string, string) {
	filter := models.MissingPersonFilter{
		Query:          strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q"))),
		Status:         utils.NormalizeToken(r.URL.Query().Get("status")),
		District:       strings.TrimSpace(strings.ToLower(r.URL.Query().Get("district"))),
		IncludePrivate: includePrivate,
	}
	if len(filter.Query) > 120 || utils.UnsafeText(filter.Query) {
		return filter, "invalid_query", "q must be 120 safe characters or fewer"
	}
	if filter.Status != "" && !allowedStatuses[filter.Status] {
		return filter, "invalid_status", "status is not supported"
	}
	if len(filter.District) > 100 || utils.UnsafeText(filter.District) {
		return filter, "invalid_district", "district filter must be 100 safe characters or fewer"
	}
	return filter, "", ""
}

func normalizeCreateMissingPerson(request models.CreateMissingPersonRequest) (models.CreateMissingPersonRequest, string, string) {
	request = normalizeCreateMissingPersonFields(request)
	if code, message := validateMissingPersonDetails(request); code != "" {
		return request, code, message
	}
	if code, message := validateLastSeenDetails(request); code != "" {
		return request, code, message
	}
	if code, message := validateReporterDetails(request); code != "" {
		return request, code, message
	}
	return request, "", ""
}

func normalizeCreateMissingPersonFields(request models.CreateMissingPersonRequest) models.CreateMissingPersonRequest {
	request.PersonName = strings.TrimSpace(request.PersonName)
	request.Gender = utils.NormalizeToken(request.Gender)
	request.Description = strings.TrimSpace(request.Description)
	request.PhotoURL = strings.TrimSpace(request.PhotoURL)
	request.RelatedIncidentID = strings.TrimSpace(request.RelatedIncidentID)
	request.LastSeenLocation.Label = strings.TrimSpace(request.LastSeenLocation.Label)
	request.LastSeenLocation.Region = strings.TrimSpace(request.LastSeenLocation.Region)
	request.LastSeenLocation.District = strings.TrimSpace(request.LastSeenLocation.District)
	request.Reporter.Name = strings.TrimSpace(request.Reporter.Name)
	request.Reporter.Phone = strings.TrimSpace(request.Reporter.Phone)
	request.Reporter.Email = strings.TrimSpace(request.Reporter.Email)
	request.Reporter.Relationship = strings.TrimSpace(request.Reporter.Relationship)
	return request
}

func validateMissingPersonDetails(request models.CreateMissingPersonRequest) (string, string) {
	if request.PersonName == "" || len(request.PersonName) > 120 || utils.UnsafeText(request.PersonName) {
		return "invalid_person_name", "personName is required and must be 120 safe characters or fewer"
	}
	if request.Age < 0 || request.Age > 120 {
		return "invalid_age", "age must be between 0 and 120"
	}
	if !allowedGenderValues[request.Gender] {
		return "invalid_gender", "gender must be female, male, non_binary, unknown, or omitted"
	}
	if request.Description == "" || len(request.Description) > 1200 || utils.UnsafeText(request.Description) {
		return "invalid_description", "description is required and must be 1200 safe characters or fewer"
	}
	if len(request.PhotoURL) > 500 || utils.UnsafeText(request.PhotoURL) {
		return "invalid_photo_url", "photoUrl must be 500 safe characters or fewer"
	}
	return "", ""
}

func validateLastSeenDetails(request models.CreateMissingPersonRequest) (string, string) {
	if request.LastSeenAt.IsZero() || request.LastSeenAt.After(time.Now().Add(24*time.Hour)) {
		return "invalid_last_seen_at", "lastSeenAt is required and cannot be in the far future"
	}
	if request.LastSeenLocation.Label == "" || len(request.LastSeenLocation.Label) > 200 || utils.UnsafeText(request.LastSeenLocation.Label) {
		return "invalid_last_seen_location", "lastSeenLocation.label is required and must be 200 safe characters or fewer"
	}
	if request.LastSeenLocation.Region == "" || len(request.LastSeenLocation.Region) > 100 || utils.UnsafeText(request.LastSeenLocation.Region) {
		return "invalid_region", "lastSeenLocation.region is required and must be 100 safe characters or fewer"
	}
	if request.LastSeenLocation.District == "" || len(request.LastSeenLocation.District) > 100 || utils.UnsafeText(request.LastSeenLocation.District) {
		return "invalid_district", "lastSeenLocation.district is required and must be 100 safe characters or fewer"
	}
	if request.LastSeenLocation.Lat != nil && (*request.LastSeenLocation.Lat < -90 || *request.LastSeenLocation.Lat > 90) {
		return "invalid_latitude", "latitude must be between -90 and 90"
	}
	if request.LastSeenLocation.Lng != nil && (*request.LastSeenLocation.Lng < -180 || *request.LastSeenLocation.Lng > 180) {
		return "invalid_longitude", "longitude must be between -180 and 180"
	}
	if len(request.RelatedIncidentID) > 120 || utils.UnsafeText(request.RelatedIncidentID) {
		return "invalid_related_incident", "relatedIncidentId must be 120 safe characters or fewer"
	}
	return "", ""
}

func validateReporterDetails(request models.CreateMissingPersonRequest) (string, string) {
	if request.Reporter.Name == "" || len(request.Reporter.Name) > 120 || utils.UnsafeText(request.Reporter.Name) {
		return "invalid_reporter_name", "reporter.name is required and must be 120 safe characters or fewer"
	}
	if request.Reporter.Phone == "" || len(request.Reporter.Phone) > 40 || utils.UnsafeText(request.Reporter.Phone) {
		return "invalid_reporter_phone", "reporter.phone is required and must be 40 safe characters or fewer"
	}
	if len(request.Reporter.Email) > 200 || utils.UnsafeText(request.Reporter.Email) {
		return "invalid_reporter_email", "reporter.email must be 200 safe characters or fewer"
	}
	if request.Reporter.Relationship == "" || len(request.Reporter.Relationship) > 80 || utils.UnsafeText(request.Reporter.Relationship) {
		return "invalid_relationship", "reporter.relationship is required and must be 80 safe characters or fewer"
	}
	if !request.Reporter.ConsentToContact {
		return "contact_consent_required", "reporter.consentToContact is required for case follow-up"
	}
	return "", ""
}

func normalizeReviewMissingPerson(request models.ReviewMissingPersonRequest) (models.ReviewMissingPersonRequest, string, string) {
	request.Decision = utils.NormalizeToken(request.Decision)
	request.Status = utils.NormalizeToken(request.Status)
	request.PublicSummary = strings.TrimSpace(request.PublicSummary)
	request.ReviewNotes = strings.TrimSpace(request.ReviewNotes)

	if request.Decision != "approve_public" && request.Decision != "approve_private" && request.Decision != "reject" {
		return request, "invalid_decision", "decision must be approve_public, approve_private, or reject"
	}
	if request.Status != "" && !allowedStatuses[request.Status] {
		return request, "invalid_status", "status is not supported"
	}
	if len(request.PublicSummary) > 500 || utils.UnsafeText(request.PublicSummary) {
		return request, "invalid_public_summary", "publicSummary must be 500 safe characters or fewer"
	}
	if len(request.ReviewNotes) > 800 || utils.UnsafeText(request.ReviewNotes) {
		return request, "invalid_review_notes", "reviewNotes must be 800 safe characters or fewer"
	}
	if request.Decision == "approve_public" && request.PublicSummary == "" {
		return request, "public_summary_required", "publicSummary is required when approving public visibility"
	}
	if request.Decision == "reject" && request.ReviewNotes == "" {
		return request, "review_notes_required", "reviewNotes is required when rejecting a record"
	}
	return request, "", ""
}

func normalizeCloseMissingPerson(request models.CloseMissingPersonRequest) (models.CloseMissingPersonRequest, string, string) {
	request.ClosureType = utils.NormalizeToken(request.ClosureType)
	request.ClosureNotes = strings.TrimSpace(request.ClosureNotes)

	if !allowedClosureTypes[request.ClosureType] {
		return request, "invalid_closure_type", "closureType must be reunited, located_safe, duplicate, withdrawn, deceased, or other"
	}
	if request.ClosureNotes == "" || len(request.ClosureNotes) > 1000 || utils.UnsafeText(request.ClosureNotes) {
		return request, "invalid_closure_notes", "closureNotes is required and must be 1000 safe characters or fewer"
	}
	return request, "", ""
}
