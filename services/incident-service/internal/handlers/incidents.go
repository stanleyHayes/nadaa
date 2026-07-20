package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

var (
	allowedHazards = map[string]bool{
		"flood":             true,
		"fire":              true,
		"road_crash":        true,
		"building_collapse": true,
		"medical_emergency": true,
		"security_incident": true,
		"disease_outbreak":  true,
		"electrical_hazard": true,
		"blocked_drain":     true,
		"landslide":         true,
		"marine_accident":   true,
		"storm":             true,
		"tidal_wave":        true,
		"other":             true,
	}
	allowedUrgencies = map[string]bool{
		"low":              true,
		"moderate":         true,
		"high":             true,
		"life_threatening": true,
	}
	allowedAgencyTypes = map[string]bool{
		"nadmo":             true,
		"district_assembly": true,
		"police":            true,
		"fire":              true,
		"ambulance":         true,
		"meteorological":    true,
		"hydrological":      true,
		"hospital":          true,
		"utility":           true,
		"ngo":               true,
		"other":             true,
	}
	allowedAssignmentPriorities = map[string]bool{
		"low":    true,
		"normal": true,
		"high":   true,
		"urgent": true,
	}
	allowedAbuseReviewDecisions = map[string]bool{
		"clear":        true,
		"monitor":      true,
		"false_report": true,
	}
	allowedTriageSeverities = map[string]bool{
		"low":       true,
		"moderate":  true,
		"high":      true,
		"emergency": true,
	}
)

var errValidation = errors.New("validation failed")

func (s *server) createIncidentHandler(w http.ResponseWriter, r *http.Request) {
	clientID := utils.ClientIdentifier(r)
	if !s.rateLimiter.Allow(clientID) {
		utils.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many incident reports submitted; please wait before trying again")
		return
	}

	var request models.CreateIncidentRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	normalized, err := normalizeIncidentRequest(request)
	if errors.Is(err, errValidation) {
		utils.WriteError(w, http.StatusBadRequest, validationCode(normalized), validationMessage(normalized))
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_incident", err.Error())
		return
	}

	record, err := s.store.CreateIncidentWithMedia(normalized, s.now())
	switch {
	case errors.Is(err, store.ErrUnknownMedia):
		utils.WriteError(w, http.StatusBadRequest, "unknown_media", "media references must be created through the upload initiation endpoint before reporting")
		return
	case errors.Is(err, store.ErrMediaAlreadyLinked):
		utils.WriteError(w, http.StatusBadRequest, "media_already_linked", "one or more media references are already linked to another incident")
		return
	case err != nil:
		utils.WriteError(w, http.StatusBadRequest, "invalid_media", err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, models.CreateIncidentResponse{
		ID:                  record.ID,
		Reference:           record.Reference,
		Status:              record.Status,
		Severity:            record.Severity,
		PriorityReview:      record.PriorityReview,
		AbuseSignals:        record.AbuseSignals,
		AbuseScore:          record.AbuseScore,
		AbuseReviewRequired: record.AbuseReviewRequired,
		DuplicateCandidates: record.DuplicateCandidates,
	})
}

func (s *server) listIncidentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, incidentReadRoles)
	if !ok {
		return
	}

	assignedToMe := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("assignedToMe"))) == "true"
	assignedAgencyID := strings.TrimSpace(r.URL.Query().Get("assignedAgencyId"))

	if assignedToMe {
		assignedAgencyID = ctx.ActorAgencyID
	}

	utils.WriteJSON(w, http.StatusOK, models.IncidentListResponse{
		Incidents: sanitizeIncidentsForAuthority(s.store.ListIncidents(assignedAgencyID), ctx),
	})
}

func (s *server) getIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, incidentReadRoles)
	if !ok {
		return
	}

	incident, found := s.store.GetIncident(r.PathValue("id"))
	if !found {
		utils.WriteError(w, http.StatusNotFound, "not_found", "incident was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
}

func (s *server) duplicateReviewHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, incidentReadRoles)
	if !ok {
		return
	}

	payload, code, message := s.store.DuplicateReview(r.PathValue("id"))
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sanitizeDuplicateReviewForAuthority(payload, ctx))
}

func (s *server) listIncidentAuditHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAuthority(w, r, incidentAuditRoles); !ok {
		return
	}

	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be a positive integer")
			return
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}

	utils.WriteJSON(w, http.StatusOK, models.IncidentAuditListResponse{Logs: s.store.ListAudit(limit)})
}

func (s *server) verifyIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, verificationRoles)
	if !ok {
		return
	}

	var request models.IncidentWorkflowRequest
	if err := utils.OptionalDecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	request.Note = strings.TrimSpace(request.Note)
	if len(request.Note) > 1000 || utils.UnsafeText(request.Note) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_note", "note must be 1000 safe characters or fewer")
		return
	}

	incident, code, message := s.store.TransitionIncident(r.PathValue("id"), "verified", ctx, request, s.now())
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
}

func (s *server) updateIncidentStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, statusWorkflowRoles)
	if !ok {
		return
	}

	var request models.IncidentStatusRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeIncidentStatusRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	// Verification is restricted to verifier roles; the status endpoint must
	// not let broader workflow roles (e.g. responder) verify incidents.
	if normalized.Status == "verified" && !verificationRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to verify incidents")
		return
	}

	// false_report is owned by the abuse-review flow; the status endpoint must
	// not let broader workflow roles (e.g. responder) mark incidents false.
	if normalized.Status == "false_report" && !abuseReviewRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to mark incidents as false reports")
		return
	}

	incident, code, message := s.store.TransitionIncident(
		r.PathValue("id"),
		normalized.Status,
		ctx,
		models.IncidentWorkflowRequest{Note: normalized.Note, ResolutionNotes: normalized.ResolutionNotes},
		s.now(),
	)
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
}

func (s *server) mergeIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, mergeRoles)
	if !ok {
		return
	}

	var request models.MergeIncidentsRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeMergeRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	payload, code, message := s.store.MergeIncidents(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sanitizeMergeResponseForAuthority(payload, ctx))
}

func (s *server) reviewAbuseHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, abuseReviewRoles)
	if !ok {
		return
	}

	var request models.AbuseReviewRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAbuseReviewRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	incident, code, message := s.store.ReviewAbuse(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sanitizeIncidentForAuthority(incident, ctx))
}

func (s *server) assignIncidentHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, assignmentRoles)
	if !ok {
		return
	}

	var request models.AssignmentRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAssignmentRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	incident, code, message := s.store.AssignIncident(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, sanitizeIncidentForAuthority(incident, ctx))
}

func (s *server) suggestTriageHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, incidentReadRoles)
	if !ok {
		return
	}

	suggestion, code, message := s.store.SuggestTriage(r.PathValue("id"), ctx, s.now())
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.TriageResponse{Suggestion: suggestion})
}

func (s *server) reviewTriageHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, triageReviewRoles)
	if !ok {
		return
	}

	var request models.TriageReviewRequest
	if err := utils.DecodeJSON(w, r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeTriageReviewRequest(request)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	incident, code, message := s.store.RecordTriageOverride(r.PathValue("id"), normalized, ctx, s.now())
	if code != "" {
		utils.WriteError(w, statusForCode(code), code, message)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.TriageReviewResponse{Incident: sanitizeIncidentForAuthority(incident, ctx)})
}

func normalizeIncidentRequest(request models.CreateIncidentRequest) (models.CreateIncidentRequest, error) {
	request.Type = strings.TrimSpace(strings.ToLower(request.Type))
	request.Description = strings.TrimSpace(request.Description)
	request.Urgency = strings.TrimSpace(strings.ToLower(request.Urgency))
	request.AccessibilityNeeds = strings.TrimSpace(request.AccessibilityNeeds)

	if request.Urgency == "" {
		request.Urgency = "moderate"
	}

	if !allowedHazards[request.Type] {
		request.Type = "invalid_type"
		return request, errValidation
	}

	if len(request.Description) < 5 || len(request.Description) > 2000 || utils.UnsafeText(request.Description) {
		request.Type = "invalid_description"
		return request, errValidation
	}

	if request.Location != nil && (!utils.ValidCoordinates(*request.Location) || (request.Location.Lat == 0 && request.Location.Lng == 0)) {
		request.Type = "invalid_location"
		return request, errValidation
	}

	if request.PeopleAffected < 0 || request.PeopleAffected > 1000000 {
		request.Type = "invalid_people_affected"
		return request, errValidation
	}

	if !allowedUrgencies[request.Urgency] {
		request.Type = "invalid_urgency"
		return request, errValidation
	}

	if request.Anonymous {
		request.Reporter = nil
	}

	if request.Reporter != nil {
		request.Reporter.UserID = strings.TrimSpace(request.Reporter.UserID)
		request.Reporter.Phone = strings.TrimSpace(request.Reporter.Phone)
		if request.Reporter.UserID == "" || !utils.MediaRefPattern.MatchString(request.Reporter.UserID) {
			request.Type = "invalid_reporter"
			return request, errValidation
		}
		if len(request.Reporter.Phone) > 32 || utils.UnsafeText(request.Reporter.Phone) {
			request.Type = "invalid_reporter"
			return request, errValidation
		}
	}

	if len(request.Media) > 10 {
		request.Type = "too_many_media"
		return request, errValidation
	}
	for index, mediaRef := range request.Media {
		mediaRef = strings.TrimSpace(mediaRef)
		if !utils.MediaRefPattern.MatchString(mediaRef) {
			request.Type = "invalid_media"
			return request, errValidation
		}
		request.Media[index] = mediaRef
	}

	if len(request.AccessibilityNeeds) > 500 || utils.UnsafeText(request.AccessibilityNeeds) {
		request.Type = "invalid_accessibility_needs"
		return request, errValidation
	}

	return request, nil
}

func normalizeIncidentStatusRequest(request models.IncidentStatusRequest) (models.IncidentStatusRequest, string, string) {
	request.Status = utils.IncidentStatusSlug(request.Status)
	request.Note = strings.TrimSpace(request.Note)
	request.ResolutionNotes = strings.TrimSpace(request.ResolutionNotes)

	if !store.AllowedIncidentStatuses[request.Status] {
		return request, "invalid_status", "status must be reported, under_review, verified, assigned, response_en_route, on_scene, contained, recovery_ongoing, closed, or false_report"
	}
	if len(request.Note) > 1000 || utils.UnsafeText(request.Note) {
		return request, "invalid_note", "note must be 1000 safe characters or fewer"
	}
	if len(request.ResolutionNotes) > 2000 || utils.UnsafeText(request.ResolutionNotes) {
		return request, "invalid_resolution_notes", "resolutionNotes must be 2000 safe characters or fewer"
	}
	if utils.RequiresResolutionNotes(request.Status) && request.ResolutionNotes == "" {
		return request, "missing_resolution_notes", "resolutionNotes are required for closed and false report statuses"
	}
	return request, "", ""
}

func normalizeMergeRequest(request models.MergeIncidentsRequest) (models.MergeIncidentsRequest, string, string) {
	request.Note = strings.TrimSpace(request.Note)

	if len(request.DuplicateIncidentIDs) == 0 {
		return request, "missing_duplicates", "duplicateIncidentIds must include at least one incident"
	}
	if len(request.DuplicateIncidentIDs) > store.DuplicateCandidateLimit {
		return request, "too_many_duplicates", fmt.Sprintf("duplicateIncidentIds can include at most %d incidents", store.DuplicateCandidateLimit)
	}

	normalizedIDs := make([]string, 0, len(request.DuplicateIncidentIDs))
	seen := map[string]bool{}
	for _, incidentID := range request.DuplicateIncidentIDs {
		incidentID = strings.TrimSpace(incidentID)
		if incidentID == "" || !utils.MediaRefPattern.MatchString(incidentID) {
			return request, "invalid_duplicate_id", "duplicateIncidentIds must contain safe incident references"
		}
		if seen[incidentID] {
			return request, "duplicate_duplicate_id", "duplicateIncidentIds must not contain the same incident more than once"
		}
		seen[incidentID] = true
		normalizedIDs = append(normalizedIDs, incidentID)
	}

	if len(request.Note) < 5 || len(request.Note) > 1000 || utils.UnsafeText(request.Note) {
		return request, "invalid_note", "note must be 5 to 1000 safe characters"
	}

	request.DuplicateIncidentIDs = normalizedIDs
	return request, "", ""
}

func normalizeAbuseReviewRequest(request models.AbuseReviewRequest) (models.AbuseReviewRequest, string, string) {
	request.Decision = utils.IncidentStatusSlug(request.Decision)
	request.Note = strings.TrimSpace(request.Note)
	request.ResolutionNotes = strings.TrimSpace(request.ResolutionNotes)

	if !allowedAbuseReviewDecisions[request.Decision] {
		return request, "invalid_decision", "decision must be clear, monitor, or false_report"
	}
	if len(request.Note) < 5 || len(request.Note) > 1000 || utils.UnsafeText(request.Note) {
		return request, "invalid_note", "note must be 5 to 1000 safe characters"
	}
	if len(request.ResolutionNotes) > 2000 || utils.UnsafeText(request.ResolutionNotes) {
		return request, "invalid_resolution_notes", "resolutionNotes must be 2000 safe characters or fewer"
	}
	if request.Decision == "false_report" && request.ResolutionNotes == "" {
		return request, "missing_resolution_notes", "resolutionNotes are required when an abuse review marks a false report"
	}

	return request, "", ""
}

func normalizeAssignmentRequest(request models.AssignmentRequest) (models.AssignmentRequest, string, string) {
	request.AgencyID = strings.TrimSpace(request.AgencyID)
	request.AgencyName = strings.TrimSpace(request.AgencyName)
	request.AgencyType = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(request.AgencyType)), "-", "_"), " ", "_")
	request.Priority = strings.TrimSpace(strings.ToLower(request.Priority))
	request.Instructions = strings.TrimSpace(request.Instructions)
	request.ResponderLead = strings.TrimSpace(request.ResponderLead)

	if request.AgencyID == "" || !utils.MediaRefPattern.MatchString(request.AgencyID) {
		return request, "invalid_agency_id", "agencyId is required and must be a safe agency reference"
	}
	if len(request.AgencyName) < 2 || len(request.AgencyName) > 140 || utils.UnsafeText(request.AgencyName) {
		return request, "invalid_agency_name", "agencyName must be 2 to 140 safe characters"
	}
	if !allowedAgencyTypes[request.AgencyType] {
		return request, "invalid_agency_type", "agencyType must be police, fire, ambulance, nadmo, district_assembly, or another supported agency type"
	}
	if request.Priority == "" {
		request.Priority = "normal"
	}
	if !allowedAssignmentPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, normal, high, or urgent"
	}
	if len(request.Instructions) < 5 || len(request.Instructions) > 1000 || utils.UnsafeText(request.Instructions) {
		return request, "invalid_instructions", "instructions must be 5 to 1000 safe characters"
	}
	if len(request.ResponderLead) > 140 || utils.UnsafeText(request.ResponderLead) {
		return request, "invalid_responder_lead", "responderLead must be 140 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeTriageReviewRequest(request models.TriageReviewRequest) (models.TriageReviewRequest, string, string) {
	request.Reason = strings.TrimSpace(request.Reason)
	request.SuggestionID = strings.TrimSpace(request.SuggestionID)

	if request.SuggestionID != "" && !utils.MediaRefPattern.MatchString(request.SuggestionID) {
		return request, "invalid_suggestion_id", "suggestionId must be a safe suggestion reference"
	}

	if fields := request.OverriddenFields; fields != nil {
		if code, message := normalizeTriageOverrideFields(fields); code != "" {
			return request, code, message
		}
	}

	if !request.Accepted || request.OverriddenFields != nil {
		if len(request.Reason) < 5 || len(request.Reason) > 1000 || utils.UnsafeText(request.Reason) {
			return request, "invalid_reason", "reason must be 5 to 1000 safe characters when overriding or rejecting a suggestion"
		}
	}

	return request, "", ""
}

func normalizeTriageOverrideFields(fields *models.TriageOverrideFields) (string, string) {
	if fields.Severity != nil {
		*fields.Severity = strings.TrimSpace(strings.ToLower(*fields.Severity))
		if !allowedTriageSeverities[*fields.Severity] {
			return "invalid_severity", "severity must be low, moderate, high, or emergency"
		}
	}
	if fields.SuggestedAgencyType != nil {
		*fields.SuggestedAgencyType = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(*fields.SuggestedAgencyType)), "-", "_"), " ", "_")
		if !allowedAgencyTypes[*fields.SuggestedAgencyType] {
			return "invalid_agency_type", "suggestedAgencyType must be a supported agency type"
		}
	}
	if fields.SuggestedAgencyID != nil {
		*fields.SuggestedAgencyID = strings.TrimSpace(*fields.SuggestedAgencyID)
		if !utils.MediaRefPattern.MatchString(*fields.SuggestedAgencyID) {
			return "invalid_agency_id", "suggestedAgencyId must be a safe agency reference"
		}
	}
	if fields.AffectedPopulation != nil {
		if *fields.AffectedPopulation < 0 || *fields.AffectedPopulation > 1000000 {
			return "invalid_people_affected", "affectedPopulation must be between 0 and 1000000"
		}
	}
	if fields.Severity == nil && fields.AffectedPopulation == nil && fields.SuggestedAgencyType == nil && fields.SuggestedAgencyID == nil {
		return "empty_override", "overriddenFields must include at least one field when supplied"
	}
	return "", ""
}

func validationCode(request models.CreateIncidentRequest) string {
	switch request.Type {
	case "invalid_type":
		return "unsupported_hazard"
	case "invalid_description":
		return "invalid_description"
	case "invalid_location":
		return "invalid_location"
	case "invalid_people_affected":
		return "invalid_people_affected"
	case "invalid_urgency":
		return "invalid_urgency"
	case "invalid_reporter":
		return "invalid_reporter"
	case "too_many_media":
		return "too_many_media"
	case "invalid_media":
		return "invalid_media"
	case "invalid_accessibility_needs":
		return "invalid_accessibility_needs"
	default:
		return "invalid_incident"
	}
}

func validationMessage(request models.CreateIncidentRequest) string {
	switch validationCode(request) {
	case "unsupported_hazard":
		return "type must be a supported hazard"
	case "invalid_description":
		return "description must be 5 to 2000 safe characters"
	case "invalid_location":
		return "location is optional, but when supplied it must contain valid lat and lng values and cannot be 0,0"
	case "invalid_people_affected":
		return "peopleAffected must be between 0 and 1000000"
	case "invalid_urgency":
		return "urgency must be low, moderate, high, or life_threatening"
	case "invalid_reporter":
		return "reporter.userId must be a safe user reference and reporter.phone must be 32 safe characters or fewer"
	case "too_many_media":
		return "a report can reference at most 10 media items"
	case "invalid_media":
		return "media references must be 3 to 128 characters using letters, numbers, underscores, or dashes"
	case "invalid_accessibility_needs":
		return "accessibilityNeeds must be 500 safe characters or fewer"
	default:
		return "incident request is invalid"
	}
}

func sanitizeIncidentsForAuthority(incidents []models.IncidentRecord, ctx models.AuthorityContext) []models.IncidentRecord {
	sanitized := make([]models.IncidentRecord, 0, len(incidents))
	for _, incident := range incidents {
		sanitized = append(sanitized, sanitizeIncidentForAuthority(incident, ctx))
	}
	return sanitized
}

func sanitizeDuplicateReviewForAuthority(payload models.DuplicateReviewResponse, ctx models.AuthorityContext) models.DuplicateReviewResponse {
	payload.Incident = sanitizeIncidentForAuthority(payload.Incident, ctx)
	for index := range payload.Candidates {
		payload.Candidates[index].Incident = sanitizeIncidentForAuthority(payload.Candidates[index].Incident, ctx)
	}
	return payload
}

func sanitizeMergeResponseForAuthority(payload models.MergeIncidentsResponse, ctx models.AuthorityContext) models.MergeIncidentsResponse {
	payload.Incident = sanitizeIncidentForAuthority(payload.Incident, ctx)
	payload.MergedIncidents = sanitizeIncidentsForAuthority(payload.MergedIncidents, ctx)
	return payload
}

func sanitizeIncidentForAuthority(incident models.IncidentRecord, ctx models.AuthorityContext) models.IncidentRecord {
	privacy := privacyForIncident(incident, ctx)
	incident.Privacy = privacy

	if !privacy.ReporterIdentityVisible {
		incident.ReportedBy = nil
		return incident
	}

	if !privacy.ReporterContactVisible && incident.ReportedBy != nil {
		reporter := *incident.ReportedBy
		reporter.Phone = ""
		incident.ReportedBy = &reporter
	}
	return incident
}

func privacyForIncident(incident models.IncidentRecord, ctx models.AuthorityContext) models.IncidentPrivacy {
	canViewContact := reporterContactRoles[ctx.ActorRole]
	hasReporter := incident.ReportedBy != nil
	reporterIdentityVisible := hasReporter && !incident.Anonymous && incident.ContactPermission && canViewContact
	reporterContactVisible := reporterIdentityVisible && incident.ReportedBy.Phone != ""

	notes := []string{
		"Exact incident location is available only to MFA-verified authority users for emergency response coordination.",
	}
	if incident.Anonymous {
		notes = append(notes, "Reporter chose anonymous reporting; citizen identity is hidden in authority views.")
	}
	if !incident.ContactPermission {
		notes = append(notes, "Reporter did not grant contact permission; contact details are hidden.")
	}
	if hasReporter && !canViewContact {
		notes = append(notes, "Current authority role receives a standard operational view without reporter contact details.")
	}

	return models.IncidentPrivacy{
		ReporterIdentityVisible: reporterIdentityVisible,
		ReporterContactVisible:  reporterContactVisible,
		LocationPrecision:       "exact",
		LocationUse:             "emergency_response",
		Disclosure:              "Location is used to route emergency response, detect duplicates, and coordinate verified authority actions.",
		Notes:                   notes,
	}
}
