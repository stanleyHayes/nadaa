package models

import "time"

// LastSeenLocation captures where a missing person was last seen.
type LastSeenLocation struct {
	Label    string   `json:"label"`
	Region   string   `json:"region"`
	District string   `json:"district"`
	Lat      *float64 `json:"lat,omitempty"`
	Lng      *float64 `json:"lng,omitempty"`
}

// ReporterContact captures private reporter contact details.
type ReporterContact struct {
	Name                 string `json:"name"`
	Phone                string `json:"phone,omitempty"`
	Email                string `json:"email,omitempty"`
	Relationship         string `json:"relationship"`
	ConsentToContact     bool   `json:"consentToContact"`
	ConsentToPublicShare bool   `json:"consentToPublicShare"`
}

// MissingPerson represents a full sensitive missing-person record.
type MissingPerson struct {
	ID                string           `json:"id"`
	Reference         string           `json:"reference"`
	PersonName        string           `json:"personName"`
	Age               *int             `json:"age,omitempty"`
	Gender            string           `json:"gender,omitempty"`
	Description       string           `json:"description"`
	PhotoURL          string           `json:"photoUrl,omitempty"`
	LastSeenAt        time.Time        `json:"lastSeenAt"`
	LastSeenLocation  LastSeenLocation `json:"lastSeenLocation"`
	RelatedIncidentID string           `json:"relatedIncidentId,omitempty"`
	Reporter          ReporterContact  `json:"reporter"`
	Status            string           `json:"status"`
	ReviewStatus      string           `json:"reviewStatus"`
	PublicVisibility  string           `json:"publicVisibility"`
	PublicSummary     string           `json:"publicSummary,omitempty"`
	ReviewNotes       string           `json:"reviewNotes,omitempty"`
	ClosureType       string           `json:"closureType,omitempty"`
	ClosureNotes      string           `json:"closureNotes,omitempty"`
	CreatedBy         string           `json:"createdBy"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
	ReviewedBy        string           `json:"reviewedBy,omitempty"`
	ReviewedAt        *time.Time       `json:"reviewedAt,omitempty"`
	ClosedBy          string           `json:"closedBy,omitempty"`
	ClosedAt          *time.Time       `json:"closedAt,omitempty"`
}

// PublicMissingPerson is the sanitized record returned to public clients.
type PublicMissingPerson struct {
	ID                string           `json:"id"`
	Reference         string           `json:"reference"`
	PersonName        string           `json:"personName"`
	Age               *int             `json:"age,omitempty"`
	Gender            string           `json:"gender,omitempty"`
	Description       string           `json:"description"`
	PhotoURL          string           `json:"photoUrl,omitempty"`
	LastSeenAt        time.Time        `json:"lastSeenAt"`
	LastSeenLocation  LastSeenLocation `json:"lastSeenLocation"`
	RelatedIncidentID string           `json:"relatedIncidentId,omitempty"`
	Status            string           `json:"status"`
	PublicSummary     string           `json:"publicSummary,omitempty"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

// MissingPersonAuditEntry records sensitive workflow activity.
type MissingPersonAuditEntry struct {
	ID            string    `json:"id"`
	RecordID      string    `json:"recordId"`
	Action        string    `json:"action"`
	ActorUserID   string    `json:"actorUserId,omitempty"`
	ActorAgencyID string    `json:"actorAgencyId,omitempty"`
	ActorRole     string    `json:"actorRole,omitempty"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

// AuthorityContext holds authenticated authority metadata from a verified
// auth-service token, or from X-NADAA-Actor-* headers when mock actors are
// explicitly enabled for local development.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	ActorDistrict string
	MFACompleted  bool
	RequestID     string
}

// TokenClaims is the signed payload of a NADAA access token issued by
// auth-service. It mirrors auth-service's claims; keep the JSON tags in sync.
type TokenClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa"`
	ExpiresAt int64  `json:"exp"`
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

// MissingPersonListResponse is the authority response for listing records.
type MissingPersonListResponse struct {
	Records     []MissingPerson `json:"records"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

// PublicMissingPersonListResponse is the public response for approved records.
type PublicMissingPersonListResponse struct {
	Records     []PublicMissingPerson `json:"records"`
	GeneratedAt time.Time             `json:"generatedAt"`
}

// MissingPersonAuditResponse is the authority response for audit entries.
type MissingPersonAuditResponse struct {
	Entries     []MissingPersonAuditEntry `json:"entries"`
	GeneratedAt time.Time                 `json:"generatedAt"`
}

// CreateMissingPersonRequest is the public intake payload.
type CreateMissingPersonRequest struct {
	PersonName        string           `json:"personName"`
	Age               *int             `json:"age,omitempty"`
	Gender            string           `json:"gender,omitempty"`
	Description       string           `json:"description"`
	PhotoURL          string           `json:"photoUrl,omitempty"`
	LastSeenAt        time.Time        `json:"lastSeenAt"`
	LastSeenLocation  LastSeenLocation `json:"lastSeenLocation"`
	RelatedIncidentID string           `json:"relatedIncidentId,omitempty"`
	Reporter          ReporterContact  `json:"reporter"`
}

// ReviewMissingPersonRequest is the authority review payload.
type ReviewMissingPersonRequest struct {
	Decision      string `json:"decision"`
	PublicSummary string `json:"publicSummary,omitempty"`
	ReviewNotes   string `json:"reviewNotes,omitempty"`
	Status        string `json:"status,omitempty"`
	// ConsentOverride explicitly overrides a reporter's declined
	// consentToPublicShare when approving public visibility; it is recorded
	// in the audit trail.
	ConsentOverride bool `json:"consentOverride,omitempty"`
}

// CloseMissingPersonRequest is the authority closure or reunification payload.
type CloseMissingPersonRequest struct {
	ClosureType        string `json:"closureType"`
	ClosureNotes       string `json:"closureNotes"`
	ReunitedWithFamily bool   `json:"reunitedWithFamily,omitempty"`
}

// MissingPersonFilter is the public/authority filter set.
type MissingPersonFilter struct {
	Query          string
	Status         string
	District       string
	IncludePrivate bool
}
