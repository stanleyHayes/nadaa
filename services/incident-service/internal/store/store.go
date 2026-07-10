package store

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

// Exported constants used by tests and handlers.
const (
	DuplicateCandidateLimit  = 5
	DuplicateDistanceMeters  = 750.0
	DuplicateReviewWindow    = 3 * time.Hour
	DuplicateMinimumScore    = 0.45
	SimilarDescriptionCutoff = 0.25
	EarthRadiusMeters        = 6371000.0
	AbuseReviewThreshold     = 0.55
	ReporterBurstWindow      = 30 * time.Minute
	ReporterBurstPreviousMin = 2

	TriageModelVersion       = "incident-triage-rules-0.1.0"
	TriageFeatureSetVersion  = "incident-features.v1"
	TriageSuggestionLogLimit = 5
)

var (
	// AllowedIncidentStatuses is the set of valid incident statuses.
	AllowedIncidentStatuses = map[string]bool{
		"reported":          true,
		"under_review":      true,
		"verified":          true,
		"assigned":          true,
		"response_en_route": true,
		"on_scene":          true,
		"contained":         true,
		"recovery_ongoing":  true,
		"closed":            true,
		"false_report":      true,
	}
	allowedIncidentTransitions = map[string]map[string]bool{
		"reported": {
			"under_review": true,
			"verified":     true,
			"false_report": true,
		},
		"under_review": {
			"verified":     true,
			"false_report": true,
		},
		"verified": {
			"assigned":          true,
			"response_en_route": true,
			"false_report":      true,
		},
		"assigned": {
			"response_en_route": true,
			"on_scene":          true,
			"contained":         true,
			"recovery_ongoing":  true,
			"closed":            true,
			"false_report":      true,
		},
		"response_en_route": {
			"on_scene":         true,
			"contained":        true,
			"recovery_ongoing": true,
			"closed":           true,
			"false_report":     true,
		},
		"on_scene": {
			"contained":        true,
			"recovery_ongoing": true,
			"closed":           true,
			"false_report":     true,
		},
		"contained": {
			"recovery_ongoing": true,
			"closed":           true,
			"false_report":     true,
		},
		"recovery_ongoing": {
			"closed":       true,
			"false_report": true,
		},
	}
)

// Store is the persistence interface for incident data.
type Store interface {
	CreateIncident(request models.CreateIncidentRequest, now time.Time) models.IncidentRecord
	ListIncidents(assignedAgencyID string) []models.IncidentRecord
	DuplicateReview(id string) (models.DuplicateReviewResponse, string, string)
	ListAudit(limit int) []models.AuditEvent
	SuggestTriage(id string, ctx models.AuthorityContext, now time.Time) (models.TriageSuggestion, string, string)
	RecordTriageOverride(id string, request models.TriageReviewRequest, ctx models.AuthorityContext, now time.Time) (models.IncidentRecord, string, string)
	TransitionIncident(id string, nextStatus string, ctx models.AuthorityContext, request models.IncidentWorkflowRequest, now time.Time) (models.IncidentRecord, string, string)
	MergeIncidents(primaryID string, request models.MergeIncidentsRequest, ctx models.AuthorityContext, now time.Time) (models.MergeIncidentsResponse, string, string)
	ReviewAbuse(id string, request models.AbuseReviewRequest, ctx models.AuthorityContext, now time.Time) (models.IncidentRecord, string, string)
	AssignIncident(id string, request models.AssignmentRequest, ctx models.AuthorityContext, now time.Time) (models.IncidentRecord, string, string)
	RegisterVolunteer(request models.RegisterVolunteerRequest, now time.Time) models.VolunteerProfile
	ListVolunteers(status, district string) []models.VolunteerProfile
	VerifyVolunteer(id string, request models.VerifyVolunteerRequest, ctx models.AuthorityContext, now time.Time) (models.VolunteerProfile, string, string)
	ListVolunteerTasks(volunteerID string) ([]models.VolunteerTaskRecord, string, string)
	AssignVolunteerTask(incidentID string, request models.VolunteerTaskRequest, ctx models.AuthorityContext, now time.Time) (models.VolunteerTaskRecord, string, string)
	UpdateVolunteerTaskStatus(taskID string, request models.VolunteerTaskStatusRequest, now time.Time) (models.VolunteerTaskRecord, string, string)
	AddVolunteerObservation(taskID string, request models.VolunteerObservationRequest, now time.Time) (models.VolunteerTaskRecord, string, string)
	CreateMediaUpload(request models.InitiateMediaUploadRequest, now time.Time) models.MediaRecord
	ListMedia() []models.MediaRecord
	ValidateMediaReferences(mediaIDs []string) error
	LinkMediaToIncident(incidentID string, mediaIDs []string, now time.Time)
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu                    sync.RWMutex
	sequence              int
	volunteerSequence     int
	volunteerTaskSequence int
	incidents             map[string]models.IncidentRecord
	media                 map[string]models.MediaRecord
	volunteers            map[string]models.VolunteerProfile
	volunteerTasks        map[string]models.VolunteerTaskRecord
	triageSuggestions     map[string][]models.TriageSuggestion
	audit                 []models.AuditEvent
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() Store {
	return &MemoryStore{
		incidents:         map[string]models.IncidentRecord{},
		media:             map[string]models.MediaRecord{},
		volunteers:        map[string]models.VolunteerProfile{},
		volunteerTasks:    map[string]models.VolunteerTaskRecord{},
		triageSuggestions: map[string][]models.TriageSuggestion{},
		audit:             []models.AuditEvent{},
	}
}

// CreateIncident persists a new incident record.
func (m *MemoryStore) CreateIncident(request models.CreateIncidentRequest, now time.Time) models.IncidentRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sequence++
	reference := fmt.Sprintf("INC-%06d", m.sequence)
	timestamp := now.UTC()
	record := models.IncidentRecord{
		ID:                 utils.NewID("inc"),
		Reference:          reference,
		Type:               request.Type,
		Severity:           severityFromUrgency(request.Urgency),
		Status:             "reported",
		Description:        request.Description,
		Location:           request.Location,
		PeopleAffected:     request.PeopleAffected,
		InjuriesReported:   request.InjuriesReported,
		Urgency:            request.Urgency,
		Anonymous:          request.Anonymous,
		ContactPermission:  request.ContactPermission,
		AccessibilityNeeds: request.AccessibilityNeeds,
		Media:              append([]string{}, request.Media...),
		PriorityReview:     priorityReview(request),
		ReportedBy:         reportedByFor(request),
		MergedIncidentIDs:  []string{},
		Assignments:        []models.IncidentAssignment{},
		Timeline: []models.TimelineEvent{
			newTimelineEvent("incident.reported", "Citizen report received", models.AuthorityContext{}, map[string]string{
				"reference": reference,
				"hazard":    request.Type,
				"urgency":   request.Urgency,
			}, timestamp),
		},
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
	record.DuplicateCandidates = m.duplicateCandidatesLocked(record)
	record.AbuseSignals = m.abuseSignalsLocked(record)
	record.AbuseScore = abuseScore(record.AbuseSignals)
	record.AbuseReviewRequired = record.AbuseScore >= AbuseReviewThreshold
	if record.AbuseReviewRequired {
		record.AbuseReviewReason = abuseReviewReason(record.AbuseSignals)
		record.Timeline = append(record.Timeline, newTimelineEvent("incident.abuse_flagged", "Suspicious report signals flagged for dispatcher review", models.AuthorityContext{}, map[string]string{
			"score":   fmt.Sprintf("%.2f", record.AbuseScore),
			"signals": strings.Join(abuseSignalCodes(record.AbuseSignals), ","),
		}, timestamp))
	}
	m.incidents[record.ID] = record
	m.linkReverseDuplicateCandidatesLocked(record)
	return record
}

// ListIncidents returns incidents filtered by assigned agency when supplied.
func (m *MemoryStore) ListIncidents(assignedAgencyID string) []models.IncidentRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incidents := make([]models.IncidentRecord, 0, len(m.incidents))
	for _, incident := range m.incidents {
		if assignedAgencyID != "" && !incidentAssignedToAgency(incident, assignedAgencyID) {
			continue
		}
		incidents = append(incidents, incident)
	}
	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt.After(incidents[j].CreatedAt)
	})
	return incidents
}

// DuplicateReview returns an incident with side-by-side duplicate candidates.
func (m *MemoryStore) DuplicateReview(id string) (models.DuplicateReviewResponse, string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	incident, ok := m.incidents[id]
	if !ok {
		return models.DuplicateReviewResponse{}, "not_found", "incident was not found"
	}

	candidates := make([]models.DuplicateReviewCandidate, 0, len(incident.DuplicateCandidates))
	for _, candidate := range incident.DuplicateCandidates {
		candidateIncident, exists := m.incidents[candidate.IncidentID]
		if !exists || candidateIncident.MergedIntoID != "" || candidateIncident.Status == "false_report" {
			continue
		}
		candidates = append(candidates, models.DuplicateReviewCandidate{
			Candidate: candidate,
			Incident:  candidateIncident,
		})
	}

	return models.DuplicateReviewResponse{Incident: incident, Candidates: candidates}, "", ""
}

// SuggestTriage returns an explainable triage suggestion for an incident and
// logs the suggestion exposure so every model output is reviewable.
func (m *MemoryStore) SuggestTriage(id string, ctx models.AuthorityContext, now time.Time) (models.TriageSuggestion, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return models.TriageSuggestion{}, "not_found", "incident was not found"
	}

	suggestion := suggestTriageForIncident(incident, m.openDuplicateCandidatesLocked(incident))
	suggestion.SuggestionID = utils.NewID("trs")
	timestamp := now.UTC()

	m.triageSuggestions[incident.ID] = append(m.triageSuggestions[incident.ID], suggestion)
	if logged := m.triageSuggestions[incident.ID]; len(logged) > TriageSuggestionLogLimit {
		m.triageSuggestions[incident.ID] = logged[len(logged)-TriageSuggestionLogLimit:]
	}

	after := map[string]any{"triageSuggestion": snapshotTriageSuggestion(suggestion)}
	m.appendAuditLocked("incident.triage_suggested", ctx, incident.ID, nil, after, timestamp)
	return suggestion, "", ""
}

// RecordTriageOverride logs dispatcher acceptance or override of a triage suggestion.
func (m *MemoryStore) RecordTriageOverride(id string, request models.TriageReviewRequest, ctx models.AuthorityContext, now time.Time) (models.IncidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return models.IncidentRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return models.IncidentRecord{}, "invalid_transition", "closed and false-report incidents are terminal"
	}

	suggestion, suggestionSource := m.resolveTriageSuggestionLocked(incident, request.SuggestionID)
	if request.SuggestionID != "" && suggestionSource == "" {
		return models.IncidentRecord{}, "unknown_suggestion", "suggestionId does not match a logged suggestion for this incident"
	}
	before := snapshotIncident(incident)
	timestamp := now.UTC()

	action := "incident.triage_accepted"
	message := "AI triage suggestion accepted by dispatcher"
	metadata := map[string]string{
		"suggestionId":      suggestion.SuggestionID,
		"suggestionSource":  suggestionSource,
		"modelVersion":      suggestion.ModelVersion,
		"featureSetVersion": suggestion.FeatureSetVersion,
		"confidence":        suggestion.Confidence,
	}

	if !request.Accepted || request.OverriddenFields != nil {
		action = "incident.triage_overridden"
		message = "AI triage suggestion overridden by dispatcher"
		metadata["overrideReason"] = request.Reason
		if fields := request.OverriddenFields; fields != nil {
			if fields.Severity != nil {
				metadata["suggestedSeverity"] = suggestion.Severity
				metadata["dispatcherSeverity"] = *fields.Severity
			}
			if fields.SuggestedAgencyType != nil {
				metadata["suggestedAgencyType"] = suggestion.SuggestedAgency.AgencyType
				metadata["dispatcherAgencyType"] = *fields.SuggestedAgencyType
			}
		}
	}

	incident.Timeline = append(incident.Timeline, newTimelineEvent(action, message, ctx, metadata, timestamp))
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident

	after := snapshotIncident(incident)
	after["triageSuggestion"] = snapshotTriageSuggestion(suggestion)
	after["triageSuggestionSource"] = suggestionSource
	if !request.Accepted || request.OverriddenFields != nil {
		after["triageOverride"] = snapshotTriageOverride(request)
	}
	m.appendAuditLocked(action, ctx, incident.ID, before, after, timestamp)
	return incident, "", ""
}

// resolveTriageSuggestionLocked returns the logged suggestion the dispatcher
// reviewed when a suggestionId is supplied, or a fresh recomputation otherwise.
// The second return value is "logged", "recomputed", or "" when the id is unknown.
func (m *MemoryStore) resolveTriageSuggestionLocked(incident models.IncidentRecord, suggestionID string) (models.TriageSuggestion, string) {
	if suggestionID != "" {
		for _, logged := range m.triageSuggestions[incident.ID] {
			if logged.SuggestionID == suggestionID {
				return logged, "logged"
			}
		}
		return models.TriageSuggestion{}, ""
	}
	return suggestTriageForIncident(incident, m.openDuplicateCandidatesLocked(incident)), "recomputed"
}

// openDuplicateCandidatesLocked filters an incident's duplicate candidates with
// the same predicate the duplicate-review endpoint applies, so triage never
// scores candidates dispatchers have already merged or marked as false reports.
func (m *MemoryStore) openDuplicateCandidatesLocked(incident models.IncidentRecord) []models.DuplicateCandidate {
	open := make([]models.DuplicateCandidate, 0, len(incident.DuplicateCandidates))
	for _, candidate := range incident.DuplicateCandidates {
		candidateIncident, exists := m.incidents[candidate.IncidentID]
		if !exists || candidateIncident.MergedIntoID != "" || candidateIncident.Status == "false_report" {
			continue
		}
		open = append(open, candidate)
	}
	return open
}

// TransitionIncident moves an incident to a new status.
func (m *MemoryStore) TransitionIncident(id string, nextStatus string, ctx models.AuthorityContext, request models.IncidentWorkflowRequest, now time.Time) (models.IncidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return models.IncidentRecord{}, "not_found", "incident was not found"
	}

	nextStatus = utils.IncidentStatusSlug(nextStatus)
	if !AllowedIncidentStatuses[nextStatus] {
		return models.IncidentRecord{}, "invalid_status", "status must be a supported incident status"
	}
	if incident.Status == nextStatus {
		return models.IncidentRecord{}, "invalid_transition", "incident is already in that status"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return models.IncidentRecord{}, "invalid_transition", "closed and false-report incidents are terminal"
	}
	if !allowedIncidentTransitions[incident.Status][nextStatus] {
		return models.IncidentRecord{}, "invalid_transition", fmt.Sprintf("cannot move incident from %s to %s", incident.Status, nextStatus)
	}

	note := strings.TrimSpace(request.Note)
	resolutionNotes := strings.TrimSpace(request.ResolutionNotes)
	if utils.RequiresResolutionNotes(nextStatus) && resolutionNotes == "" {
		return models.IncidentRecord{}, "missing_resolution_notes", "resolutionNotes are required for closed and false report statuses"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	incident.Status = nextStatus
	incident.StatusUpdatedBy = ctx.ActorUserID
	incident.StatusReason = note
	incident.UpdatedAt = timestamp

	action := "incident.status_changed"
	if nextStatus == "verified" {
		action = "incident.verified"
		incident.VerifiedBy = ctx.ActorUserID
		if incident.VerifiedAt == nil {
			incident.VerifiedAt = &timestamp
		}
	}
	if utils.RequiresResolutionNotes(nextStatus) {
		incident.ResolutionNotes = resolutionNotes
		incident.ClosedAt = &timestamp
		incident.AbuseReviewRequired = false
		if nextStatus == "closed" {
			action = "incident.closed"
		} else {
			action = "incident.false_reported"
			incident.AbuseReviewDecision = "false_report"
			incident.AbuseReviewReason = note
			incident.AbuseReviewedBy = ctx.ActorUserID
			incident.AbuseReviewedAt = &timestamp
		}
	}

	fromStatus, _ := before["status"].(string)
	incident.Timeline = append(incident.Timeline, newTimelineEvent(action, timelineMessageForStatus(nextStatus, note, resolutionNotes), ctx, map[string]string{
		"fromStatus": fromStatus,
		"toStatus":   nextStatus,
	}, timestamp))
	m.incidents[incident.ID] = incident
	m.appendAuditLocked(action, ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	return incident, "", ""
}

// MergeIncidents merges duplicate incidents into a primary incident.
func (m *MemoryStore) MergeIncidents(primaryID string, request models.MergeIncidentsRequest, ctx models.AuthorityContext, now time.Time) (models.MergeIncidentsResponse, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	primary, ok := m.incidents[primaryID]
	if !ok {
		return models.MergeIncidentsResponse{}, "not_found", "incident was not found"
	}
	if primary.MergedIntoID != "" {
		return models.MergeIncidentsResponse{}, "invalid_merge", "primary incident is already merged into another incident"
	}
	if primary.Status == "closed" || primary.Status == "false_report" {
		return models.MergeIncidentsResponse{}, "invalid_merge", "closed and false-report incidents cannot receive duplicate merges"
	}

	duplicates := make([]models.IncidentRecord, 0, len(request.DuplicateIncidentIDs))
	for _, duplicateID := range request.DuplicateIncidentIDs {
		duplicate, exists := m.incidents[duplicateID]
		if !exists {
			return models.MergeIncidentsResponse{}, "not_found", fmt.Sprintf("duplicate incident %s was not found", duplicateID)
		}
		if duplicate.ID == primary.ID {
			return models.MergeIncidentsResponse{}, "invalid_merge", "primary incident cannot be merged into itself"
		}
		if duplicate.MergedIntoID != "" {
			return models.MergeIncidentsResponse{}, "invalid_merge", fmt.Sprintf("duplicate incident %s is already merged", duplicate.Reference)
		}
		if duplicate.Status == "closed" || duplicate.Status == "false_report" {
			return models.MergeIncidentsResponse{}, "invalid_merge", fmt.Sprintf("duplicate incident %s is terminal", duplicate.Reference)
		}
		if _, ok := duplicateCandidateBetween(primary, duplicate); !ok {
			return models.MergeIncidentsResponse{}, "invalid_duplicate", fmt.Sprintf("incident %s is not a duplicate candidate for %s", duplicate.Reference, primary.Reference)
		}
		duplicates = append(duplicates, duplicate)
	}

	beforePrimary := snapshotIncident(primary)
	timestamp := now.UTC()
	mergedIDs := make([]string, 0, len(duplicates))
	mergedIncidents := make([]models.IncidentRecord, 0, len(duplicates))
	removeIDs := map[string]bool{}
	for _, duplicate := range duplicates {
		removeIDs[duplicate.ID] = true
		mergedIDs = append(mergedIDs, duplicate.ID)
	}

	for _, duplicate := range duplicates {
		beforeDuplicate := snapshotIncident(duplicate)
		duplicate.MergedIntoID = primary.ID
		duplicate.MergedBy = ctx.ActorUserID
		duplicate.MergedAt = &timestamp
		duplicate.MergeReason = request.Note
		duplicate.Status = "closed"
		duplicate.StatusUpdatedBy = ctx.ActorUserID
		duplicate.StatusReason = "Merged into " + primary.Reference
		duplicate.ResolutionNotes = request.Note
		duplicate.ClosedAt = &timestamp
		duplicate.UpdatedAt = timestamp
		duplicate.DuplicateCandidates = filterDuplicateCandidates(duplicate.DuplicateCandidates, map[string]bool{primary.ID: true})
		duplicate.Timeline = append(duplicate.Timeline, newTimelineEvent("incident.merged_into", "Merged into "+primary.Reference, ctx, map[string]string{
			"primaryIncidentId": primary.ID,
			"primaryReference":  primary.Reference,
			"note":              request.Note,
		}, timestamp))

		m.incidents[duplicate.ID] = duplicate
		m.appendAuditLocked("incident.merged_into", ctx, duplicate.ID, beforeDuplicate, snapshotIncident(duplicate), timestamp)
		mergedIncidents = append(mergedIncidents, duplicate)
	}

	primary.MergedIncidentIDs = appendUniqueStrings(primary.MergedIncidentIDs, mergedIDs...)
	sort.Strings(primary.MergedIncidentIDs)
	primary.DuplicateCandidates = filterDuplicateCandidates(primary.DuplicateCandidates, removeIDs)
	primary.StatusUpdatedBy = ctx.ActorUserID
	primary.StatusReason = fmt.Sprintf("Merged %d duplicate report(s)", len(mergedIDs))
	primary.UpdatedAt = timestamp
	primary.Timeline = append(primary.Timeline, newTimelineEvent("incident.merged", fmt.Sprintf("Merged %d duplicate report(s)", len(mergedIDs)), ctx, map[string]string{
		"duplicateIncidentIds": strings.Join(mergedIDs, ","),
		"note":                 request.Note,
	}, timestamp))

	m.incidents[primary.ID] = primary
	m.appendAuditLocked("incident.merged", ctx, primary.ID, beforePrimary, snapshotIncident(primary), timestamp)
	return models.MergeIncidentsResponse{Incident: primary, MergedIncidents: mergedIncidents}, "", ""
}

// ReviewAbuse records an abuse-review decision and may close the incident.
func (m *MemoryStore) ReviewAbuse(id string, request models.AbuseReviewRequest, ctx models.AuthorityContext, now time.Time) (models.IncidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return models.IncidentRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return models.IncidentRecord{}, "invalid_transition", "closed and false-report incidents are terminal"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	incident.AbuseReviewDecision = request.Decision
	incident.AbuseReviewReason = request.Note
	incident.AbuseReviewedBy = ctx.ActorUserID
	incident.AbuseReviewedAt = &timestamp
	incident.StatusReason = request.Note
	incident.UpdatedAt = timestamp

	action := "incident.abuse_reviewed"
	message := "Suspicious report review updated"
	metadata := map[string]string{"decision": request.Decision}

	switch request.Decision {
	case "clear":
		incident.AbuseReviewRequired = false
		action = "incident.abuse_cleared"
		message = "Suspicious report signals cleared"
	case "monitor":
		incident.AbuseReviewRequired = true
		action = "incident.abuse_monitored"
		message = "Suspicious report kept under dispatcher monitoring"
	case "false_report":
		if !allowedIncidentTransitions[incident.Status]["false_report"] {
			return models.IncidentRecord{}, "invalid_transition", fmt.Sprintf("cannot move incident from %s to false_report", incident.Status)
		}
		incident.Status = "false_report"
		incident.StatusUpdatedBy = ctx.ActorUserID
		incident.ResolutionNotes = request.ResolutionNotes
		incident.ClosedAt = &timestamp
		incident.AbuseReviewRequired = false
		action = "incident.false_reported"
		message = "Incident marked as false report after abuse review"
		metadata["resolutionNotes"] = request.ResolutionNotes
	}

	incident.Timeline = append(incident.Timeline, newTimelineEvent(action, message, ctx, metadata, timestamp))
	m.incidents[incident.ID] = incident
	m.appendAuditLocked(action, ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	return incident, "", ""
}

// AssignIncident assigns an agency to an incident.
func (m *MemoryStore) AssignIncident(id string, request models.AssignmentRequest, ctx models.AuthorityContext, now time.Time) (models.IncidentRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[id]
	if !ok {
		return models.IncidentRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "reported" || incident.Status == "under_review" {
		return models.IncidentRecord{}, "invalid_transition", "incident must be verified before assignment"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return models.IncidentRecord{}, "invalid_transition", "closed and false-report incidents cannot be assigned"
	}
	if ctx.ActorRole == "agency_admin" && ctx.ActorAgencyID != request.AgencyID {
		return models.IncidentRecord{}, "forbidden", "agency admins can assign only to their own agency"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	assignment := models.IncidentAssignment{
		ID:            fmt.Sprintf("asg_%06d", len(incident.Assignments)+1),
		AgencyID:      request.AgencyID,
		AgencyName:    request.AgencyName,
		AgencyType:    request.AgencyType,
		Priority:      request.Priority,
		Instructions:  request.Instructions,
		ResponderLead: request.ResponderLead,
		Status:        "active",
		AssignedBy:    ctx.ActorUserID,
		AssignedAt:    timestamp,
	}

	incident.Assignments = append(incident.Assignments, assignment)
	if incident.Status == "verified" {
		incident.Status = "assigned"
	}
	incident.StatusUpdatedBy = ctx.ActorUserID
	incident.StatusReason = "Assigned to " + assignment.AgencyName
	incident.UpdatedAt = timestamp
	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.assigned", "Assigned to "+assignment.AgencyName, ctx, map[string]string{
		"assignmentId": assignment.ID,
		"agencyId":     assignment.AgencyID,
		"agencyName":   assignment.AgencyName,
		"agencyType":   assignment.AgencyType,
		"priority":     assignment.Priority,
	}, timestamp))

	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.assigned", ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	return incident, "", ""
}

// RegisterVolunteer creates a pending volunteer profile.
func (m *MemoryStore) RegisterVolunteer(request models.RegisterVolunteerRequest, now time.Time) models.VolunteerProfile {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.volunteerSequence++
	timestamp := now.UTC()
	volunteer := models.VolunteerProfile{
		ID:                 fmt.Sprintf("vol_%06d", m.volunteerSequence),
		CitizenUserID:      request.CitizenUserID,
		Name:               request.Name,
		Phone:              request.Phone,
		Region:             request.Region,
		District:           request.District,
		Community:          request.Community,
		GroupID:            volunteerGroupID(request.Region, request.District, request.Community),
		Skills:             append([]string{}, request.Skills...),
		Languages:          append([]string{}, request.Languages...),
		AvailabilityStatus: request.AvailabilityStatus,
		VerificationStatus: "pending",
		SafetyNotes:        VolunteerSafetyRules(),
		CreatedAt:          timestamp,
		UpdatedAt:          timestamp,
	}
	m.volunteers[volunteer.ID] = volunteer
	m.appendAuditForTargetLocked("volunteer.registered", volunteerActorContext(volunteer), "volunteer_profile", volunteer.ID, nil, snapshotVolunteer(volunteer), timestamp)
	return volunteer
}

// ListVolunteers returns volunteers filtered by status and district.
func (m *MemoryStore) ListVolunteers(status string, district string) []models.VolunteerProfile {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volunteers := make([]models.VolunteerProfile, 0, len(m.volunteers))
	for _, volunteer := range m.volunteers {
		if status != "" && volunteer.VerificationStatus != status {
			continue
		}
		if district != "" && !strings.EqualFold(volunteer.District, district) {
			continue
		}
		volunteers = append(volunteers, volunteer)
	}
	sort.Slice(volunteers, func(i, j int) bool {
		if volunteers[i].UpdatedAt.Equal(volunteers[j].UpdatedAt) {
			return volunteers[i].ID < volunteers[j].ID
		}
		return volunteers[i].UpdatedAt.After(volunteers[j].UpdatedAt)
	})
	return volunteers
}

// VerifyVolunteer updates a volunteer's verification status.
func (m *MemoryStore) VerifyVolunteer(id string, request models.VerifyVolunteerRequest, ctx models.AuthorityContext, now time.Time) (models.VolunteerProfile, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	volunteer, ok := m.volunteers[id]
	if !ok {
		return models.VolunteerProfile{}, "not_found", "volunteer profile was not found"
	}

	before := snapshotVolunteer(volunteer)
	timestamp := now.UTC()
	volunteer.VerifiedBy = ctx.ActorUserID
	volunteer.VerifiedAt = &timestamp
	volunteer.UpdatedAt = timestamp
	volunteer.RejectionReason = ""
	switch request.Decision {
	case "verify":
		volunteer.VerificationStatus = "verified"
	case "reject":
		volunteer.VerificationStatus = "rejected"
		volunteer.RejectionReason = request.Note
	case "suspend":
		volunteer.VerificationStatus = "suspended"
		volunteer.RejectionReason = request.Note
	}
	m.volunteers[volunteer.ID] = volunteer
	m.appendAuditForTargetLocked("volunteer."+volunteer.VerificationStatus, ctx, "volunteer_profile", volunteer.ID, before, snapshotVolunteer(volunteer), timestamp)
	return volunteer, "", ""
}

// ListVolunteerTasks returns tasks for a volunteer.
func (m *MemoryStore) ListVolunteerTasks(volunteerID string) ([]models.VolunteerTaskRecord, string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, ok := m.volunteers[volunteerID]; !ok {
		return nil, "not_found", "volunteer profile was not found"
	}

	tasks := make([]models.VolunteerTaskRecord, 0, len(m.volunteerTasks))
	for _, task := range m.volunteerTasks {
		if task.VolunteerID == volunteerID {
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})
	return tasks, "", ""
}

// AssignVolunteerTask assigns a task to a verified volunteer.
func (m *MemoryStore) AssignVolunteerTask(incidentID string, request models.VolunteerTaskRequest, ctx models.AuthorityContext, now time.Time) (models.VolunteerTaskRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	incident, ok := m.incidents[incidentID]
	if !ok {
		return models.VolunteerTaskRecord{}, "not_found", "incident was not found"
	}
	if incident.Status == "reported" || incident.Status == "under_review" {
		return models.VolunteerTaskRecord{}, "invalid_transition", "incident must be verified before volunteer tasks can be assigned"
	}
	if incident.Status == "closed" || incident.Status == "false_report" {
		return models.VolunteerTaskRecord{}, "invalid_transition", "closed and false-report incidents cannot receive volunteer tasks"
	}
	volunteer, ok := m.volunteers[request.VolunteerID]
	if !ok {
		return models.VolunteerTaskRecord{}, "not_found", "volunteer profile was not found"
	}
	if volunteer.VerificationStatus != "verified" {
		return models.VolunteerTaskRecord{}, "volunteer_not_verified", "volunteer must be verified before task assignment"
	}
	if volunteer.AvailabilityStatus != "available" {
		return models.VolunteerTaskRecord{}, "volunteer_unavailable", "volunteer must be available before task assignment"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	m.volunteerTaskSequence++
	task := models.VolunteerTaskRecord{
		ID:                 fmt.Sprintf("vtask_%06d", m.volunteerTaskSequence),
		IncidentID:         incident.ID,
		IncidentReference:  incident.Reference,
		VolunteerID:        volunteer.ID,
		VolunteerName:      volunteer.Name,
		GroupID:            volunteer.GroupID,
		Type:               request.Type,
		Priority:           request.Priority,
		Instructions:       request.Instructions,
		LocationLabel:      request.LocationLabel,
		Status:             "assigned",
		SafetyRules:        VolunteerSafetyRules(),
		EscalationRequired: false,
		AssignedBy:         ctx.ActorUserID,
		AssignedAt:         timestamp,
		UpdatedAt:          timestamp,
		Updates:            []models.VolunteerTaskUpdate{},
	}
	m.volunteerTasks[task.ID] = task

	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_assigned", "Volunteer task assigned to "+volunteer.Name, ctx, map[string]string{
		"taskId":      task.ID,
		"volunteerId": volunteer.ID,
		"groupId":     volunteer.GroupID,
		"taskType":    task.Type,
		"priority":    task.Priority,
	}, timestamp))
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.volunteer_assigned", ctx, incident.ID, before, snapshotIncident(incident), timestamp)
	m.appendAuditForTargetLocked("volunteer_task.assigned", ctx, "volunteer_task", task.ID, nil, snapshotVolunteerTask(task), timestamp)
	return task, "", ""
}

// UpdateVolunteerTaskStatus records a status change from the volunteer.
func (m *MemoryStore) UpdateVolunteerTaskStatus(taskID string, request models.VolunteerTaskStatusRequest, now time.Time) (models.VolunteerTaskRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.volunteerTasks[taskID]
	if !ok {
		return models.VolunteerTaskRecord{}, "not_found", "volunteer task was not found"
	}
	if task.VolunteerID != request.VolunteerID {
		return models.VolunteerTaskRecord{}, "forbidden", "volunteer can update only their own tasks"
	}
	incident, ok := m.incidents[task.IncidentID]
	if !ok {
		return models.VolunteerTaskRecord{}, "not_found", "linked incident was not found"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	task.Status = request.Status
	task.UpdatedAt = timestamp
	if request.Status == "accepted" && task.AcceptedAt == nil {
		task.AcceptedAt = &timestamp
	}
	if request.Status == "completed" {
		task.CompletedAt = &timestamp
	}
	if request.Status == "needs_escalation" || request.SafetyStatus == "unsafe" || request.SafetyStatus == "needs_authority" {
		task.EscalationRequired = true
	}
	update := models.VolunteerTaskUpdate{
		ID:                  fmt.Sprintf("vtup_%06d", len(task.Updates)+1),
		Type:                "status",
		Status:              request.Status,
		Note:                request.Note,
		SafetyStatus:        request.SafetyStatus,
		Location:            request.Location,
		EscalationRequested: task.EscalationRequired,
		CreatedBy:           request.VolunteerID,
		CreatedAt:           timestamp,
	}
	task.Updates = append(task.Updates, update)
	m.volunteerTasks[task.ID] = task

	actor := volunteerActorContextByID(request.VolunteerID)
	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_status_updated", "Volunteer task "+request.Status, actor, map[string]string{
		"taskId":       task.ID,
		"volunteerId":  request.VolunteerID,
		"status":       request.Status,
		"safetyStatus": request.SafetyStatus,
	}, timestamp))
	if task.EscalationRequired {
		incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_escalation", "Volunteer requested authority escalation", actor, map[string]string{
			"taskId":       task.ID,
			"volunteerId":  request.VolunteerID,
			"safetyStatus": request.SafetyStatus,
		}, timestamp))
	}
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.volunteer_status_updated", actor, incident.ID, before, snapshotIncident(incident), timestamp)
	return task, "", ""
}

// AddVolunteerObservation records a field observation from the volunteer.
func (m *MemoryStore) AddVolunteerObservation(taskID string, request models.VolunteerObservationRequest, now time.Time) (models.VolunteerTaskRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.volunteerTasks[taskID]
	if !ok {
		return models.VolunteerTaskRecord{}, "not_found", "volunteer task was not found"
	}
	if task.VolunteerID != request.VolunteerID {
		return models.VolunteerTaskRecord{}, "forbidden", "volunteer can add observations only to their own tasks"
	}
	incident, ok := m.incidents[task.IncidentID]
	if !ok {
		return models.VolunteerTaskRecord{}, "not_found", "linked incident was not found"
	}

	before := snapshotIncident(incident)
	timestamp := now.UTC()
	if request.EscalationRequested || request.SafetyStatus == "unsafe" || request.SafetyStatus == "needs_authority" {
		task.EscalationRequired = true
		task.Status = "needs_escalation"
	}
	update := models.VolunteerTaskUpdate{
		ID:                  fmt.Sprintf("vtup_%06d", len(task.Updates)+1),
		Type:                "observation",
		Status:              task.Status,
		Note:                request.Observation,
		SafetyStatus:        request.SafetyStatus,
		Location:            request.Location,
		EscalationRequested: request.EscalationRequested,
		CreatedBy:           request.VolunteerID,
		CreatedAt:           timestamp,
	}
	task.Updates = append(task.Updates, update)
	task.UpdatedAt = timestamp
	m.volunteerTasks[task.ID] = task

	actor := volunteerActorContextByID(request.VolunteerID)
	incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_observation", "Volunteer field observation received", actor, map[string]string{
		"taskId":       task.ID,
		"volunteerId":  request.VolunteerID,
		"safetyStatus": request.SafetyStatus,
		"mediaCount":   strconv.Itoa(len(request.Media)),
	}, timestamp))
	if task.EscalationRequired {
		incident.Timeline = append(incident.Timeline, newTimelineEvent("incident.volunteer_escalation", "Volunteer observation requires authority review", actor, map[string]string{
			"taskId":       task.ID,
			"volunteerId":  request.VolunteerID,
			"safetyStatus": request.SafetyStatus,
		}, timestamp))
	}
	incident.UpdatedAt = timestamp
	m.incidents[incident.ID] = incident
	m.appendAuditLocked("incident.volunteer_observation", actor, incident.ID, before, snapshotIncident(incident), timestamp)
	return task, "", ""
}

// ListAudit returns recent audit events up to limit.
func (m *MemoryStore) ListAudit(limit int) []models.AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logs := append([]models.AuditEvent(nil), m.audit...)
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].CreatedAt.Equal(logs[j].CreatedAt) {
			return logs[i].ID > logs[j].ID
		}
		return logs[i].CreatedAt.After(logs[j].CreatedAt)
	})
	if len(logs) > limit {
		return logs[:limit]
	}
	return logs
}

// CreateMediaUpload persists a pending media upload record.
func (m *MemoryStore) CreateMediaUpload(request models.InitiateMediaUploadRequest, now time.Time) models.MediaRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	timestamp := now.UTC()
	id := utils.NewID("media")
	record := models.MediaRecord{
		ID:          id,
		Purpose:     request.Purpose,
		FileName:    request.FileName,
		ContentType: request.ContentType,
		SizeBytes:   request.SizeBytes,
		UploadedBy:  request.UploadedBy,
		Access:      "private",
		Status:      "pending_upload",
		UploadURL:   fmt.Sprintf("/dev/uploads/%s/%s", id, request.FileName),
		ExpiresAt:   timestamp.Add(15 * time.Minute),
		CreatedAt:   timestamp,
	}
	m.media[record.ID] = record
	return record
}

// Media reference errors.
var (
	ErrUnknownMedia       = errors.New("unknown media")
	ErrMediaAlreadyLinked = errors.New("media already linked")
	ErrDuplicateMediaRef  = errors.New("duplicate media reference")
)

// ValidateMediaReferences ensures media IDs exist and are not already linked.
func (m *MemoryStore) ValidateMediaReferences(mediaIDs []string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := map[string]bool{}
	for _, mediaID := range mediaIDs {
		if seen[mediaID] {
			return ErrDuplicateMediaRef
		}
		seen[mediaID] = true

		record, ok := m.media[mediaID]
		if !ok {
			return ErrUnknownMedia
		}
		if record.IncidentID != "" || record.Status == "linked" {
			return ErrMediaAlreadyLinked
		}
	}
	return nil
}

// LinkMediaToIncident marks media records as linked to an incident.
func (m *MemoryStore) LinkMediaToIncident(incidentID string, mediaIDs []string, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	linkedAt := now.UTC()
	for _, mediaID := range mediaIDs {
		record := m.media[mediaID]
		record.IncidentID = incidentID
		record.Status = "linked"
		record.LinkedAt = &linkedAt
		m.media[mediaID] = record
	}
}

// ListMedia returns all media records.
func (m *MemoryStore) ListMedia() []models.MediaRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	media := make([]models.MediaRecord, 0, len(m.media))
	for _, record := range m.media {
		media = append(media, record)
	}
	sort.Slice(media, func(i, j int) bool {
		return media[i].CreatedAt.After(media[j].CreatedAt)
	})
	return media
}

func (m *MemoryStore) appendAuditLocked(action string, ctx models.AuthorityContext, targetID string, before map[string]any, after map[string]any, now time.Time) {
	m.appendAuditForTargetLocked(action, ctx, "incident", targetID, before, after, now)
}

func (m *MemoryStore) appendAuditForTargetLocked(action string, ctx models.AuthorityContext, targetType string, targetID string, before map[string]any, after map[string]any, now time.Time) {
	m.audit = append(m.audit, models.AuditEvent{
		ID:            fmt.Sprintf("aud_%06d", len(m.audit)+1),
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Action:        action,
		TargetType:    targetType,
		TargetID:      targetID,
		RequestID:     ctx.RequestID,
		Before:        before,
		After:         after,
		CreatedAt:     now,
	})
}

func (m *MemoryStore) duplicateCandidatesLocked(record models.IncidentRecord) []models.DuplicateCandidate {
	candidates := make([]models.DuplicateCandidate, 0, DuplicateCandidateLimit)
	for _, existing := range m.incidents {
		candidate, ok := scoreDuplicateCandidate(record, existing)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].DistanceMeters < candidates[j].DistanceMeters
		}
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > DuplicateCandidateLimit {
		return candidates[:DuplicateCandidateLimit]
	}
	return candidates
}

func (m *MemoryStore) linkReverseDuplicateCandidatesLocked(record models.IncidentRecord) {
	for _, candidate := range record.DuplicateCandidates {
		existing := m.incidents[candidate.IncidentID]
		if hasDuplicateCandidate(existing.DuplicateCandidates, record.ID) {
			continue
		}

		existing.DuplicateCandidates = append(existing.DuplicateCandidates, reverseDuplicateCandidate(record, candidate))
		sort.Slice(existing.DuplicateCandidates, func(i, j int) bool {
			if existing.DuplicateCandidates[i].Score == existing.DuplicateCandidates[j].Score {
				return existing.DuplicateCandidates[i].DistanceMeters < existing.DuplicateCandidates[j].DistanceMeters
			}
			return existing.DuplicateCandidates[i].Score > existing.DuplicateCandidates[j].Score
		})
		if len(existing.DuplicateCandidates) > DuplicateCandidateLimit {
			existing.DuplicateCandidates = existing.DuplicateCandidates[:DuplicateCandidateLimit]
		}
		existing.UpdatedAt = record.UpdatedAt
		m.incidents[existing.ID] = existing
	}
}

func (m *MemoryStore) abuseSignalsLocked(record models.IncidentRecord) []models.AbuseSignal {
	signals := abuseSignalsForDescription(record.Description)

	reporterKey := reporterAbuseKey(record.ReportedBy)
	if reporterKey != "" {
		recentReports := 0
		cutoff := record.CreatedAt.Add(-ReporterBurstWindow)
		for _, existing := range m.incidents {
			if existing.ID == record.ID || existing.CreatedAt.Before(cutoff) {
				continue
			}
			if reporterAbuseKey(existing.ReportedBy) == reporterKey {
				recentReports++
			}
		}
		if recentReports >= ReporterBurstPreviousMin {
			signals = append(signals, models.AbuseSignal{
				Code:   "reporter_burst",
				Label:  "Reporter burst",
				Detail: fmt.Sprintf("Reporter submitted %d other report(s) in the last %d minutes.", recentReports, int(ReporterBurstWindow.Minutes())),
				Weight: 0.55,
			})
		}
	}

	sort.Slice(signals, func(i, j int) bool {
		if signals[i].Weight == signals[j].Weight {
			return signals[i].Code < signals[j].Code
		}
		return signals[i].Weight > signals[j].Weight
	})
	return signals
}

func scoreDuplicateCandidate(record models.IncidentRecord, existing models.IncidentRecord) (models.DuplicateCandidate, bool) {
	if record.ID == existing.ID || record.Type != existing.Type || existing.Status == "false_report" {
		return models.DuplicateCandidate{}, false
	}

	timeApart := absoluteDuration(record.CreatedAt.Sub(existing.CreatedAt))
	if timeApart > DuplicateReviewWindow {
		return models.DuplicateCandidate{}, false
	}

	distance := haversineMeters(record.Location, existing.Location)
	if distance > DuplicateDistanceMeters {
		return models.DuplicateCandidate{}, false
	}

	descriptionScore := descriptionSimilarity(record.Description, existing.Description)
	distanceScore := clamp01(1 - distance/DuplicateDistanceMeters)
	timeScore := clamp01(1 - timeApart.Seconds()/DuplicateReviewWindow.Seconds())
	score := roundScore(0.50*distanceScore + 0.30*timeScore + 0.20*descriptionScore)
	if score < DuplicateMinimumScore {
		return models.DuplicateCandidate{}, false
	}

	reasons := []string{"same_hazard", "nearby_location", "recent_report"}
	if descriptionScore >= SimilarDescriptionCutoff {
		reasons = append(reasons, "similar_description")
	}

	return models.DuplicateCandidate{
		IncidentID:     existing.ID,
		Reference:      existing.Reference,
		Score:          score,
		DistanceMeters: math.Round(distance),
		MinutesApart:   int(math.Round(timeApart.Minutes())),
		Reasons:        reasons,
	}, true
}

func reverseDuplicateCandidate(record models.IncidentRecord, candidate models.DuplicateCandidate) models.DuplicateCandidate {
	return models.DuplicateCandidate{
		IncidentID:     record.ID,
		Reference:      record.Reference,
		Score:          candidate.Score,
		DistanceMeters: candidate.DistanceMeters,
		MinutesApart:   candidate.MinutesApart,
		Reasons:        append([]string{}, candidate.Reasons...),
	}
}

func hasDuplicateCandidate(candidates []models.DuplicateCandidate, incidentID string) bool {
	for _, candidate := range candidates {
		if candidate.IncidentID == incidentID {
			return true
		}
	}
	return false
}

func haversineMeters(a models.Coordinates, b models.Coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)

	sinLat := math.Sin(deltaLat / 2)
	sinLng := math.Sin(deltaLng / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	return EarthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func descriptionSimilarity(a string, b string) float64 {
	aTokens := tokenSet(a)
	bTokens := tokenSet(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}

	intersection := 0
	union := map[string]bool{}
	for token := range aTokens {
		union[token] = true
		if bTokens[token] {
			intersection++
		}
	}
	for token := range bTokens {
		union[token] = true
	}

	return float64(intersection) / float64(len(union))
}

func tokenSet(value string) map[string]bool {
	tokens := utils.WordPattern.FindAllString(strings.ToLower(value), -1)
	set := map[string]bool{}
	for _, token := range tokens {
		if len(token) < 3 {
			continue
		}
		set[token] = true
	}
	return set
}

func repeatedTokenRatio(value string) float64 {
	tokens := utils.WordPattern.FindAllString(strings.ToLower(value), -1)
	if len(tokens) < 6 {
		return 0
	}

	counts := map[string]int{}
	maxCount := 0
	for _, token := range tokens {
		if len(token) < 3 {
			continue
		}
		counts[token]++
		if counts[token] > maxCount {
			maxCount = counts[token]
		}
	}
	if len(counts) == 0 {
		return 0
	}
	return float64(maxCount) / float64(len(tokens))
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func reporterAbuseKey(reporter *models.ReporterRef) string {
	if reporter == nil {
		return ""
	}
	if reporter.UserID != "" {
		return "user:" + strings.ToLower(reporter.UserID)
	}
	if reporter.Phone != "" {
		return "phone:" + strings.ToLower(reporter.Phone)
	}
	return ""
}

func abuseScore(signals []models.AbuseSignal) float64 {
	score := 0.0
	for _, signal := range signals {
		score += signal.Weight
	}
	return roundScore(clamp01(score))
}

func abuseReviewReason(signals []models.AbuseSignal) string {
	if len(signals) == 0 {
		return ""
	}
	labels := make([]string, 0, len(signals))
	for _, signal := range signals {
		labels = append(labels, signal.Label)
	}
	return "Review requested: " + strings.Join(labels, ", ")
}

func abuseSignalCodes(signals []models.AbuseSignal) []string {
	codes := make([]string, 0, len(signals))
	for _, signal := range signals {
		codes = append(codes, signal.Code)
	}
	return codes
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func absoluteDuration(value time.Duration) time.Duration {
	if value < 0 {
		return -value
	}
	return value
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func roundScore(value float64) float64 {
	return math.Round(value*100) / 100
}

func abuseSignalsForDescription(description string) []models.AbuseSignal {
	lower := strings.ToLower(description)
	signals := []models.AbuseSignal{}

	if strings.Contains(lower, "http://") ||
		strings.Contains(lower, "https://") ||
		strings.Contains(lower, "www.") ||
		strings.Contains(lower, "bit.ly") {
		signals = append(signals, models.AbuseSignal{
			Code:   "external_link",
			Label:  "External link",
			Detail: "Description includes a public link, which can indicate spam in citizen reporting.",
			Weight: 0.45,
		})
	}

	if containsAny(lower, []string{"free money", "promo", "promotion", "discount", "loan offer", "click here", "whatsapp me"}) {
		signals = append(signals, models.AbuseSignal{
			Code:   "promotional_language",
			Label:  "Promotional wording",
			Detail: "Description includes marketing or solicitation language uncommon in emergency reports.",
			Weight: 0.35,
		})
	}

	if repeatedTokenRatio(lower) >= 0.50 {
		signals = append(signals, models.AbuseSignal{
			Code:   "repeated_language",
			Label:  "Repeated language",
			Detail: "Description repeats the same terms unusually often.",
			Weight: 0.25,
		})
	}

	if len([]rune(strings.TrimSpace(description))) < 24 {
		signals = append(signals, models.AbuseSignal{
			Code:   "low_detail",
			Label:  "Low detail",
			Detail: "Description is very short and may need dispatcher confirmation.",
			Weight: 0.20,
		})
	}

	return signals
}

func severityFromUrgency(urgency string) string {
	switch urgency {
	case "low":
		return "low"
	case "high":
		return "high"
	case "life_threatening":
		return "emergency"
	default:
		return "moderate"
	}
}

func priorityReview(request models.CreateIncidentRequest) bool {
	return request.Urgency == "life_threatening" || request.InjuriesReported
}

func reportedByFor(request models.CreateIncidentRequest) *models.ReporterRef {
	if request.Anonymous || request.Reporter == nil {
		return nil
	}

	reporter := *request.Reporter
	if !request.ContactPermission {
		reporter.Phone = ""
	}
	return &reporter
}

func incidentAssignedToAgency(incident models.IncidentRecord, agencyID string) bool {
	for _, assignment := range incident.Assignments {
		if assignment.AgencyID == agencyID && assignment.Status == "active" {
			return true
		}
	}
	return false
}

func assignmentAgencyIDs(assignments []models.IncidentAssignment) []string {
	ids := make([]string, 0, len(assignments))
	seen := map[string]bool{}
	for _, assignment := range assignments {
		if assignment.AgencyID == "" || seen[assignment.AgencyID] {
			continue
		}
		seen[assignment.AgencyID] = true
		ids = append(ids, assignment.AgencyID)
	}
	sort.Strings(ids)
	return ids
}

func duplicateCandidateBetween(primary models.IncidentRecord, duplicate models.IncidentRecord) (models.DuplicateCandidate, bool) {
	for _, candidate := range primary.DuplicateCandidates {
		if candidate.IncidentID == duplicate.ID {
			return candidate, true
		}
	}
	for _, candidate := range duplicate.DuplicateCandidates {
		if candidate.IncidentID == primary.ID {
			return reverseDuplicateCandidate(duplicate, candidate), true
		}
	}
	return models.DuplicateCandidate{}, false
}

func filterDuplicateCandidates(candidates []models.DuplicateCandidate, removeIDs map[string]bool) []models.DuplicateCandidate {
	if len(candidates) == 0 || len(removeIDs) == 0 {
		return candidates
	}
	filtered := make([]models.DuplicateCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if removeIDs[candidate.IncidentID] {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered
}

func appendUniqueStrings(values []string, additions ...string) []string {
	seen := map[string]bool{}
	next := make([]string, 0, len(values)+len(additions))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		next = append(next, value)
	}
	for _, value := range additions {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		next = append(next, value)
	}
	return next
}

func snapshotIncident(incident models.IncidentRecord) map[string]any {
	return map[string]any{
		"id":                  incident.ID,
		"reference":           incident.Reference,
		"type":                incident.Type,
		"severity":            incident.Severity,
		"status":              incident.Status,
		"priorityReview":      incident.PriorityReview,
		"verifiedBy":          incident.VerifiedBy,
		"statusUpdatedBy":     incident.StatusUpdatedBy,
		"statusReason":        incident.StatusReason,
		"resolutionNotes":     incident.ResolutionNotes,
		"abuseScore":          incident.AbuseScore,
		"abuseReviewRequired": incident.AbuseReviewRequired,
		"abuseReviewDecision": incident.AbuseReviewDecision,
		"mergedIncidentIds":   append([]string{}, incident.MergedIncidentIDs...),
		"mergedIntoId":        incident.MergedIntoID,
		"mergeReason":         incident.MergeReason,
		"duplicateCount":      len(incident.DuplicateCandidates),
		"assignmentCount":     len(incident.Assignments),
		"assignedAgencyIds":   assignmentAgencyIDs(incident.Assignments),
	}
}

func snapshotVolunteer(volunteer models.VolunteerProfile) map[string]any {
	return map[string]any{
		"id":                 volunteer.ID,
		"citizenUserId":      volunteer.CitizenUserID,
		"groupId":            volunteer.GroupID,
		"district":           volunteer.District,
		"community":          volunteer.Community,
		"skills":             append([]string{}, volunteer.Skills...),
		"availabilityStatus": volunteer.AvailabilityStatus,
		"verificationStatus": volunteer.VerificationStatus,
		"verifiedBy":         volunteer.VerifiedBy,
		"rejectionReason":    volunteer.RejectionReason,
	}
}

func snapshotVolunteerTask(task models.VolunteerTaskRecord) map[string]any {
	return map[string]any{
		"id":                 task.ID,
		"incidentId":         task.IncidentID,
		"incidentReference":  task.IncidentReference,
		"volunteerId":        task.VolunteerID,
		"groupId":            task.GroupID,
		"type":               task.Type,
		"priority":           task.Priority,
		"status":             task.Status,
		"escalationRequired": task.EscalationRequired,
		"updateCount":        len(task.Updates),
	}
}

func newTimelineEvent(eventType string, message string, ctx models.AuthorityContext, metadata map[string]string, now time.Time) models.TimelineEvent {
	return models.TimelineEvent{
		ID:            utils.NewID("tle"),
		Type:          eventType,
		Message:       message,
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Metadata:      metadata,
		CreatedAt:     now,
	}
}

func timelineMessageForStatus(status string, note string, resolutionNotes string) string {
	switch status {
	case "verified":
		return "Incident verified"
	case "closed":
		if resolutionNotes != "" {
			return "Incident closed with resolution notes"
		}
		return "Incident closed"
	case "false_report":
		return "Incident marked as false report"
	default:
		if note != "" {
			return fmt.Sprintf("Status changed to %s: %s", status, note)
		}
		return "Status changed to " + status
	}
}

func volunteerGroupID(region string, district string, community string) string {
	return fmt.Sprintf("grp_%s_%s_%s", slugRef(region), slugRef(district), slugRef(community))
}

func slugRef(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	previousDash := false
	for _, char := range lower {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			previousDash = false
			continue
		}
		if !previousDash && builder.Len() > 0 {
			builder.WriteByte('-')
			previousDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func volunteerActorContext(volunteer models.VolunteerProfile) models.AuthorityContext {
	return volunteerActorContextByID(volunteer.ID)
}

func volunteerActorContextByID(volunteerID string) models.AuthorityContext {
	return models.AuthorityContext{
		ActorUserID: volunteerID,
		ActorRole:   "citizen",
	}
}

func suggestTriageForIncident(incident models.IncidentRecord, openCandidates []models.DuplicateCandidate) models.TriageSuggestion {
	severity := triageSeverity(incident)
	duplicateLikelihood, topDuplicates := triageDuplicateSignal(openCandidates)
	affectedPopulation := triageAffectedPopulation(incident, openCandidates)
	agency := triageSuggestedAgency(incident)
	confidence := triageConfidence(incident, duplicateLikelihood)

	factors := []models.TriageExplanationFactor{
		{
			Feature:      "urgency",
			Label:        "Reported urgency",
			Value:        incident.Urgency,
			Contribution: triageUrgencyContribution(incident.Urgency),
			Direction:    "increases_risk",
		},
		{
			Feature:      "people_affected",
			Label:        "People directly affected",
			Value:        incident.PeopleAffected,
			Contribution: triagePeopleContribution(incident.PeopleAffected),
			Direction:    "increases_risk",
		},
		{
			Feature:      "hazard_type",
			Label:        "Hazard type",
			Value:        incident.Type,
			Contribution: triageHazardContribution(incident.Type),
			Direction:    "increases_risk",
		},
	}

	if incident.InjuriesReported {
		factors = append(factors, models.TriageExplanationFactor{
			Feature:      "injuries_reported",
			Label:        "Injuries reported",
			Value:        true,
			Contribution: 0.25,
			Direction:    "increases_risk",
		})
	}

	if len(openCandidates) > 0 {
		factors = append(factors, models.TriageExplanationFactor{
			Feature:      "duplicate_candidates",
			Label:        "Duplicate report candidates",
			Value:        len(openCandidates),
			Contribution: roundScore(duplicateLikelihood),
			Direction:    "increases_risk",
		})
	} else {
		factors = append(factors, models.TriageExplanationFactor{
			Feature:      "duplicate_candidates",
			Label:        "No duplicate candidates",
			Value:        0,
			Contribution: -0.10,
			Direction:    "reduces_risk",
		})
	}

	if incident.AbuseReviewRequired {
		factors = append(factors, models.TriageExplanationFactor{
			Feature:      "abuse_review_required",
			Label:        "Suspicious report signals flagged",
			Value:        incident.AbuseScore,
			Contribution: -0.15,
			Direction:    "reduces_risk",
		})
	}

	return models.TriageSuggestion{
		Severity:                severity,
		DuplicateLikelihood:     roundScore(duplicateLikelihood),
		TopDuplicateIncidentIDs: topDuplicates,
		AffectedPopulation:      affectedPopulation,
		SuggestedAgency:         agency,
		Confidence:              confidence,
		ModelVersion:            TriageModelVersion,
		FeatureSetVersion:       TriageFeatureSetVersion,
		ExplanationFactors:      factors,
		HumanReviewRequired:     true,
		AutoPublishAllowed:      false,
	}
}

func triageSeverity(incident models.IncidentRecord) string {
	if incident.Urgency == "life_threatening" || incident.InjuriesReported {
		return "emergency"
	}
	if incident.Urgency == "high" || incident.PeopleAffected >= 20 {
		return "high"
	}
	if incident.Urgency == "moderate" || incident.PeopleAffected >= 5 {
		return "moderate"
	}
	return "low"
}

func triageDuplicateSignal(openCandidates []models.DuplicateCandidate) (float64, []string) {
	if len(openCandidates) == 0 {
		return 0, []string{}
	}
	maxScore := 0.0
	for _, candidate := range openCandidates {
		if candidate.Score > maxScore {
			maxScore = candidate.Score
		}
	}
	limit := min(len(openCandidates), 3)
	ids := make([]string, 0, limit)
	for index := range limit {
		ids = append(ids, openCandidates[index].IncidentID)
	}
	return maxScore, ids
}

func triageAffectedPopulation(incident models.IncidentRecord, openCandidates []models.DuplicateCandidate) int {
	base := incident.PeopleAffected
	if base <= 0 {
		base = 1
	}
	if incident.Urgency == "life_threatening" {
		base *= 2
	}
	if len(openCandidates) > 0 {
		base += len(openCandidates) * 3
	}
	if base > 1000000 {
		return 1000000
	}
	return base
}

func triageSuggestedAgency(incident models.IncidentRecord) models.TriageAgencySuggestion {
	switch incident.Type {
	case "fire", "electrical_hazard", "building_collapse":
		return models.TriageAgencySuggestion{
			AgencyType: "fire",
			AgencyID:   "00000000-0000-0000-0000-000000000201",
			Name:       "Ghana National Fire Service",
			Reason:     "Primary responder for fire and structural collapse incidents.",
		}
	case "road_crash":
		return models.TriageAgencySuggestion{
			AgencyType: "police",
			AgencyID:   "00000000-0000-0000-0000-000000000203",
			Name:       "Ghana Police Service",
			Reason:     "Traffic and scene control for road crashes; ambulance should be co-dispatched for casualties.",
		}
	case "medical_emergency", "disease_outbreak":
		return models.TriageAgencySuggestion{
			AgencyType: "ambulance",
			AgencyID:   "00000000-0000-0000-0000-000000000202",
			Name:       "National Ambulance Service",
			Reason:     "Primary responder for medical and health incidents.",
		}
	case "blocked_drain":
		return models.TriageAgencySuggestion{
			AgencyType: "district_assembly",
			AgencyID:   "00000000-0000-0000-0000-000000000204",
			Name:       "Accra Metropolitan Assembly",
			Reason:     "Local sanitation and drainage works responsibility.",
		}
	case "security_incident":
		return models.TriageAgencySuggestion{
			AgencyType: "police",
			AgencyID:   "00000000-0000-0000-0000-000000000203",
			Name:       "Ghana Police Service",
			Reason:     "Law enforcement lead for security incidents.",
		}
	default:
		return models.TriageAgencySuggestion{
			AgencyType: "nadmo",
			AgencyID:   "00000000-0000-0000-0000-000000000101",
			Name:       "NADMO Accra Metro",
			Reason:     "NADMO coordinates multi-hazard disaster response for this report.",
		}
	}
}

func triageConfidence(incident models.IncidentRecord, duplicateLikelihood float64) string {
	if incident.PeopleAffected > 0 && (incident.Urgency != "low" || duplicateLikelihood > 0) {
		return "high"
	}
	if incident.PeopleAffected > 0 || incident.Urgency != "low" {
		return "medium"
	}
	return "low"
}

func triageUrgencyContribution(urgency string) float64 {
	switch urgency {
	case "life_threatening":
		return 0.90
	case "high":
		return 0.60
	case "moderate":
		return 0.30
	default:
		return 0.10
	}
}

func triagePeopleContribution(peopleAffected int) float64 {
	if peopleAffected >= 50 {
		return 0.70
	}
	if peopleAffected >= 20 {
		return 0.50
	}
	if peopleAffected >= 5 {
		return 0.30
	}
	if peopleAffected > 0 {
		return 0.10
	}
	return 0
}

func triageHazardContribution(hazard string) float64 {
	switch hazard {
	case "fire", "medical_emergency", "building_collapse":
		return 0.40
	case "flood", "road_crash", "electrical_hazard", "security_incident":
		return 0.30
	case "landslide", "storm", "tidal_wave":
		return 0.35
	default:
		return 0.20
	}
}

func snapshotTriageSuggestion(suggestion models.TriageSuggestion) map[string]any {
	return map[string]any{
		"suggestionId":            suggestion.SuggestionID,
		"severity":                suggestion.Severity,
		"duplicateLikelihood":     suggestion.DuplicateLikelihood,
		"topDuplicateIncidentIds": append([]string{}, suggestion.TopDuplicateIncidentIDs...),
		"affectedPopulation":      suggestion.AffectedPopulation,
		"suggestedAgencyType":     suggestion.SuggestedAgency.AgencyType,
		"suggestedAgencyId":       suggestion.SuggestedAgency.AgencyID,
		"confidence":              suggestion.Confidence,
		"modelVersion":            suggestion.ModelVersion,
		"featureSetVersion":       suggestion.FeatureSetVersion,
	}
}

// snapshotTriageOverride records only the fields the dispatcher actually
// supplied, so an unedited field is never audited as an explicit zero value.
func snapshotTriageOverride(request models.TriageReviewRequest) map[string]any {
	snapshot := map[string]any{
		"accepted": request.Accepted,
		"reason":   request.Reason,
	}
	if fields := request.OverriddenFields; fields != nil {
		overridden := map[string]any{}
		if fields.Severity != nil {
			overridden["severity"] = *fields.Severity
		}
		if fields.AffectedPopulation != nil {
			overridden["affectedPopulation"] = *fields.AffectedPopulation
		}
		if fields.SuggestedAgencyType != nil {
			overridden["suggestedAgencyType"] = *fields.SuggestedAgencyType
		}
		if fields.SuggestedAgencyID != nil {
			overridden["suggestedAgencyId"] = *fields.SuggestedAgencyID
		}
		snapshot["overriddenFields"] = overridden
	}
	return snapshot
}
