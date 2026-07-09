package models

import "time"

// Coordinates represents a latitude/longitude pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CreateIncidentRequest is the payload for reporting a new incident.
type CreateIncidentRequest struct {
	Type               string       `json:"type"`
	Description        string       `json:"description"`
	Location           Coordinates  `json:"location"`
	PeopleAffected     int          `json:"peopleAffected"`
	InjuriesReported   bool         `json:"injuriesReported"`
	Urgency            string       `json:"urgency"`
	Anonymous          bool         `json:"anonymous"`
	ContactPermission  bool         `json:"contactPermission"`
	AccessibilityNeeds string       `json:"accessibilityNeeds"`
	Media              []string     `json:"media"`
	Reporter           *ReporterRef `json:"reporter,omitempty"`
}

// ReporterRef identifies the citizen reporter.
type ReporterRef struct {
	UserID string `json:"userId"`
	Phone  string `json:"phone,omitempty"`
}

// IncidentRecord is the persisted incident entity.
type IncidentRecord struct {
	ID                  string               `json:"id"`
	Reference           string               `json:"reference"`
	Type                string               `json:"type"`
	Severity            string               `json:"severity"`
	Status              string               `json:"status"`
	Description         string               `json:"description"`
	Location            Coordinates          `json:"location"`
	PeopleAffected      int                  `json:"peopleAffected"`
	InjuriesReported    bool                 `json:"injuriesReported"`
	Urgency             string               `json:"urgency"`
	Anonymous           bool                 `json:"anonymous"`
	ContactPermission   bool                 `json:"contactPermission"`
	Privacy             IncidentPrivacy      `json:"privacy"`
	AccessibilityNeeds  string               `json:"accessibilityNeeds,omitempty"`
	Media               []string             `json:"media"`
	PriorityReview      bool                 `json:"priorityReview"`
	AbuseSignals        []AbuseSignal        `json:"abuseSignals"`
	AbuseScore          float64              `json:"abuseScore"`
	AbuseReviewRequired bool                 `json:"abuseReviewRequired"`
	AbuseReviewReason   string               `json:"abuseReviewReason,omitempty"`
	AbuseReviewDecision string               `json:"abuseReviewDecision,omitempty"`
	AbuseReviewedBy     string               `json:"abuseReviewedBy,omitempty"`
	AbuseReviewedAt     *time.Time           `json:"abuseReviewedAt,omitempty"`
	DuplicateCandidates []DuplicateCandidate `json:"duplicateCandidates"`
	MergedIncidentIDs   []string             `json:"mergedIncidentIds"`
	ReportedBy          *ReporterRef         `json:"reportedBy,omitempty"`
	Assignments         []IncidentAssignment `json:"assignments"`
	Timeline            []TimelineEvent      `json:"timeline"`
	MergedIntoID        string               `json:"mergedIntoId,omitempty"`
	MergedBy            string               `json:"mergedBy,omitempty"`
	MergedAt            *time.Time           `json:"mergedAt,omitempty"`
	MergeReason         string               `json:"mergeReason,omitempty"`
	VerifiedBy          string               `json:"verifiedBy,omitempty"`
	VerifiedAt          *time.Time           `json:"verifiedAt,omitempty"`
	StatusUpdatedBy     string               `json:"statusUpdatedBy,omitempty"`
	StatusReason        string               `json:"statusReason,omitempty"`
	ResolutionNotes     string               `json:"resolutionNotes,omitempty"`
	ClosedAt            *time.Time           `json:"closedAt,omitempty"`
	CreatedAt           time.Time            `json:"createdAt"`
	UpdatedAt           time.Time            `json:"updatedAt"`
}

// IncidentPrivacy describes what authority users can see about a reporter.
type IncidentPrivacy struct {
	ReporterIdentityVisible bool     `json:"reporterIdentityVisible"`
	ReporterContactVisible  bool     `json:"reporterContactVisible"`
	LocationPrecision       string   `json:"locationPrecision"`
	LocationUse             string   `json:"locationUse"`
	Disclosure              string   `json:"disclosure"`
	Notes                   []string `json:"notes"`
}

// DuplicateCandidate scores a potential duplicate incident.
type DuplicateCandidate struct {
	IncidentID     string   `json:"incidentId"`
	Reference      string   `json:"reference"`
	Score          float64  `json:"score"`
	DistanceMeters float64  `json:"distanceMeters"`
	MinutesApart   int      `json:"minutesApart"`
	Reasons        []string `json:"reasons"`
}

// AbuseSignal describes a suspicious-report indicator.
type AbuseSignal struct {
	Code   string  `json:"code"`
	Label  string  `json:"label"`
	Detail string  `json:"detail"`
	Weight float64 `json:"weight"`
}

// CreateIncidentResponse is returned after a successful incident report.
type CreateIncidentResponse struct {
	ID                  string               `json:"id"`
	Reference           string               `json:"reference"`
	Status              string               `json:"status"`
	Severity            string               `json:"severity"`
	PriorityReview      bool                 `json:"priorityReview"`
	AbuseSignals        []AbuseSignal        `json:"abuseSignals"`
	AbuseScore          float64              `json:"abuseScore"`
	AbuseReviewRequired bool                 `json:"abuseReviewRequired"`
	DuplicateCandidates []DuplicateCandidate `json:"duplicateCandidates"`
}

// IncidentListResponse is the authority incident feed payload.
type IncidentListResponse struct {
	Incidents []IncidentRecord `json:"incidents"`
}

// DuplicateReviewResponse is the side-by-side duplicate review payload.
type DuplicateReviewResponse struct {
	Incident   IncidentRecord             `json:"incident"`
	Candidates []DuplicateReviewCandidate `json:"candidates"`
}

// DuplicateReviewCandidate pairs a duplicate score with the full incident.
type DuplicateReviewCandidate struct {
	Candidate DuplicateCandidate `json:"candidate"`
	Incident  IncidentRecord     `json:"incident"`
}

// MergeIncidentsRequest requests merging duplicates into a primary incident.
type MergeIncidentsRequest struct {
	DuplicateIncidentIDs []string `json:"duplicateIncidentIds"`
	Note                 string   `json:"note"`
}

// MergeIncidentsResponse returns the updated primary and closed duplicates.
type MergeIncidentsResponse struct {
	Incident        IncidentRecord   `json:"incident"`
	MergedIncidents []IncidentRecord `json:"mergedIncidents"`
}

// AssignmentRequest assigns an agency to an incident.
type AssignmentRequest struct {
	AgencyID      string `json:"agencyId"`
	AgencyName    string `json:"agencyName"`
	AgencyType    string `json:"agencyType"`
	Priority      string `json:"priority"`
	Instructions  string `json:"instructions"`
	ResponderLead string `json:"responderLead"`
}

// IncidentAssignment records an agency assignment.
type IncidentAssignment struct {
	ID            string    `json:"id"`
	AgencyID      string    `json:"agencyId"`
	AgencyName    string    `json:"agencyName"`
	AgencyType    string    `json:"agencyType"`
	Priority      string    `json:"priority"`
	Instructions  string    `json:"instructions"`
	ResponderLead string    `json:"responderLead,omitempty"`
	Status        string    `json:"status"`
	AssignedBy    string    `json:"assignedBy"`
	AssignedAt    time.Time `json:"assignedAt"`
}

// VolunteerProfile is a registered community volunteer record.
type VolunteerProfile struct {
	ID                 string     `json:"id"`
	CitizenUserID      string     `json:"citizenUserId"`
	Name               string     `json:"name"`
	Phone              string     `json:"phone,omitempty"`
	Region             string     `json:"region"`
	District           string     `json:"district"`
	Community          string     `json:"community"`
	GroupID            string     `json:"groupId"`
	Skills             []string   `json:"skills"`
	Languages          []string   `json:"languages"`
	AvailabilityStatus string     `json:"availabilityStatus"`
	VerificationStatus string     `json:"verificationStatus"`
	SafetyNotes        []string   `json:"safetyNotes"`
	VerifiedBy         string     `json:"verifiedBy,omitempty"`
	VerifiedAt         *time.Time `json:"verifiedAt,omitempty"`
	RejectionReason    string     `json:"rejectionReason,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// RegisterVolunteerRequest creates a new volunteer profile.
type RegisterVolunteerRequest struct {
	CitizenUserID      string   `json:"citizenUserId"`
	Name               string   `json:"name"`
	Phone              string   `json:"phone"`
	Region             string   `json:"region"`
	District           string   `json:"district"`
	Community          string   `json:"community"`
	Skills             []string `json:"skills"`
	Languages          []string `json:"languages"`
	AvailabilityStatus string   `json:"availabilityStatus"`
}

// VolunteerProfileResponse wraps a single volunteer.
type VolunteerProfileResponse struct {
	Volunteer VolunteerProfile `json:"volunteer"`
}

// VolunteerListResponse lists volunteers.
type VolunteerListResponse struct {
	Volunteers []VolunteerProfile `json:"volunteers"`
}

// VerifyVolunteerRequest verifies or rejects a volunteer.
type VerifyVolunteerRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note"`
}

// VolunteerTaskRequest assigns a task to a volunteer.
type VolunteerTaskRequest struct {
	VolunteerID   string `json:"volunteerId"`
	Type          string `json:"type"`
	Priority      string `json:"priority"`
	Instructions  string `json:"instructions"`
	LocationLabel string `json:"locationLabel"`
}

// VolunteerTaskRecord is a persisted volunteer task.
type VolunteerTaskRecord struct {
	ID                 string                `json:"id"`
	IncidentID         string                `json:"incidentId"`
	IncidentReference  string                `json:"incidentReference"`
	VolunteerID        string                `json:"volunteerId"`
	VolunteerName      string                `json:"volunteerName"`
	GroupID            string                `json:"groupId"`
	Type               string                `json:"type"`
	Priority           string                `json:"priority"`
	Instructions       string                `json:"instructions"`
	LocationLabel      string                `json:"locationLabel"`
	Status             string                `json:"status"`
	SafetyRules        []string              `json:"safetyRules"`
	EscalationRequired bool                  `json:"escalationRequired"`
	AssignedBy         string                `json:"assignedBy"`
	AssignedAt         time.Time             `json:"assignedAt"`
	UpdatedAt          time.Time             `json:"updatedAt"`
	AcceptedAt         *time.Time            `json:"acceptedAt,omitempty"`
	CompletedAt        *time.Time            `json:"completedAt,omitempty"`
	Updates            []VolunteerTaskUpdate `json:"updates"`
}

// VolunteerTaskUpdate records a status change or observation.
type VolunteerTaskUpdate struct {
	ID                  string       `json:"id"`
	Type                string       `json:"type"`
	Status              string       `json:"status,omitempty"`
	Note                string       `json:"note"`
	SafetyStatus        string       `json:"safetyStatus"`
	Location            *Coordinates `json:"location,omitempty"`
	EscalationRequested bool         `json:"escalationRequested"`
	CreatedBy           string       `json:"createdBy"`
	CreatedAt           time.Time    `json:"createdAt"`
}

// VolunteerTaskListResponse lists tasks for a volunteer.
type VolunteerTaskListResponse struct {
	Tasks []VolunteerTaskRecord `json:"tasks"`
}

// VolunteerTaskStatusRequest updates a task status.
type VolunteerTaskStatusRequest struct {
	VolunteerID  string       `json:"volunteerId"`
	Status       string       `json:"status"`
	Note         string       `json:"note"`
	SafetyStatus string       `json:"safetyStatus"`
	Location     *Coordinates `json:"location,omitempty"`
}

// VolunteerObservationRequest submits a field observation.
type VolunteerObservationRequest struct {
	VolunteerID         string       `json:"volunteerId"`
	Observation         string       `json:"observation"`
	SafetyStatus        string       `json:"safetyStatus"`
	Location            *Coordinates `json:"location,omitempty"`
	EscalationRequested bool         `json:"escalationRequested"`
	Media               []string     `json:"media"`
}

// TimelineEvent records a significant incident event.
type TimelineEvent struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Message       string            `json:"message"`
	ActorUserID   string            `json:"actorUserId,omitempty"`
	ActorAgencyID string            `json:"actorAgencyId,omitempty"`
	ActorRole     string            `json:"actorRole,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
}

// IncidentWorkflowRequest captures optional notes for workflow actions.
type IncidentWorkflowRequest struct {
	Note            string `json:"note"`
	ResolutionNotes string `json:"resolutionNotes"`
}

// IncidentStatusRequest updates incident status.
type IncidentStatusRequest struct {
	Status          string `json:"status"`
	Note            string `json:"note"`
	ResolutionNotes string `json:"resolutionNotes"`
}

// AbuseReviewRequest reviews suspicious-report signals.
type AbuseReviewRequest struct {
	Decision        string `json:"decision"`
	Note            string `json:"note"`
	ResolutionNotes string `json:"resolutionNotes"`
}

// IncidentAuditListResponse returns audit log entries.
type IncidentAuditListResponse struct {
	Logs []AuditEvent `json:"logs"`
}

// TriageExplanationFactor describes one input to the triage suggestion.
type TriageExplanationFactor struct {
	Feature      string  `json:"feature"`
	Label        string  `json:"label"`
	Value        any     `json:"value"`
	Contribution float64 `json:"contribution"`
	Direction    string  `json:"direction"`
}

// TriageAgencySuggestion recommends a responder agency type and optional id.
type TriageAgencySuggestion struct {
	AgencyType string `json:"agencyType"`
	AgencyID   string `json:"agencyId,omitempty"`
	Name       string `json:"name"`
	Reason     string `json:"reason"`
}

// TriageSuggestion is an explainable, human-reviewed incident triage output.
type TriageSuggestion struct {
	SuggestionID            string                    `json:"suggestionId"`
	Severity                string                    `json:"severity"`
	DuplicateLikelihood     float64                   `json:"duplicateLikelihood"`
	TopDuplicateIncidentIDs []string                  `json:"topDuplicateIncidentIds"`
	AffectedPopulation      int                       `json:"affectedPopulation"`
	SuggestedAgency         TriageAgencySuggestion    `json:"suggestedAgency"`
	Confidence              string                    `json:"confidence"`
	ModelVersion            string                    `json:"modelVersion"`
	FeatureSetVersion       string                    `json:"featureSetVersion"`
	ExplanationFactors      []TriageExplanationFactor `json:"explanationFactors"`
	HumanReviewRequired     bool                      `json:"humanReviewRequired"`
	AutoPublishAllowed      bool                      `json:"autoPublishAllowed"`
}

// TriageOverrideFields captures dispatcher edits to a triage suggestion.
// Pointer fields distinguish "not overridden" from an explicit zero or empty value.
type TriageOverrideFields struct {
	Severity            *string `json:"severity,omitempty"`
	AffectedPopulation  *int    `json:"affectedPopulation,omitempty"`
	SuggestedAgencyType *string `json:"suggestedAgencyType,omitempty"`
	SuggestedAgencyID   *string `json:"suggestedAgencyId,omitempty"`
}

// TriageReviewRequest records whether a dispatcher accepted or overrode a suggestion.
type TriageReviewRequest struct {
	Accepted         bool                  `json:"accepted"`
	SuggestionID     string                `json:"suggestionId,omitempty"`
	OverriddenFields *TriageOverrideFields `json:"overriddenFields,omitempty"`
	Reason           string                `json:"reason,omitempty"`
}

// TriageResponse returns a triage suggestion for an incident.
type TriageResponse struct {
	Suggestion TriageSuggestion `json:"suggestion"`
}

// TriageReviewResponse returns the incident after a triage review is recorded.
type TriageReviewResponse struct {
	Incident IncidentRecord `json:"incident"`
}

// AuditEvent records an authority action.
type AuditEvent struct {
	ID            string         `json:"id"`
	ActorUserID   string         `json:"actorUserId"`
	ActorAgencyID string         `json:"actorAgencyId"`
	ActorRole     string         `json:"actorRole"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId"`
	RequestID     string         `json:"requestId,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

// AuthorityContext carries the authenticated authority actor.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

// InitiateMediaUploadRequest starts a presigned media upload.
type InitiateMediaUploadRequest struct {
	Purpose     string `json:"purpose"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	UploadedBy  string `json:"uploadedBy,omitempty"`
}

// MediaUploadResponse returns upload metadata.
type MediaUploadResponse struct {
	MediaID      string            `json:"mediaId"`
	UploadURL    string            `json:"uploadUrl"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	ExpiresAt    time.Time         `json:"expiresAt"`
	MaxSizeBytes int64             `json:"maxSizeBytes"`
	Access       string            `json:"access"`
}

// MediaRecord is a persisted media upload reference.
type MediaRecord struct {
	ID          string     `json:"id"`
	Purpose     string     `json:"purpose"`
	FileName    string     `json:"fileName"`
	ContentType string     `json:"contentType"`
	SizeBytes   int64      `json:"sizeBytes"`
	UploadedBy  string     `json:"uploadedBy,omitempty"`
	IncidentID  string     `json:"incidentId,omitempty"`
	Access      string     `json:"access"`
	Status      string     `json:"status"`
	UploadURL   string     `json:"uploadUrl"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	LinkedAt    *time.Time `json:"linkedAt,omitempty"`
}

// MediaListResponse lists media records.
type MediaListResponse struct {
	Media []MediaRecord `json:"media"`
}

// APIError is the standard error response envelope.
type APIError struct {
	Error APIErrorBody `json:"error"`
}

// APIErrorBody is the standard error response body.
type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
