package models

import "time"

// ReporterInfo holds the contact details of the citizen reporting damage.
type ReporterInfo struct {
	Name   string `json:"name"`
	Phone  string `json:"phone,omitempty"`
	Email  string `json:"email,omitempty"`
	UserID string `json:"userId,omitempty"`
}

// ClaimLocation holds WGS84 coordinates and an optional address for the damage.
type ClaimLocation struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address,omitempty"`
}

// DamageClaimRecord is the persisted representation of a damage claim.
type DamageClaimRecord struct {
	ID                  string        `json:"id"`
	Reference           string        `json:"reference"`
	IncidentID          string        `json:"incidentId,omitempty"`
	IncidentReference   string        `json:"incidentReference,omitempty"`
	IncidentLocation    string        `json:"incidentLocation,omitempty"`
	Reporter            ReporterInfo  `json:"reporter"`
	DamageType          string        `json:"damageType"`
	DamageDescription   string        `json:"damageDescription"`
	EstimatedLossAmount string        `json:"estimatedLossAmount"`
	DamagePhotos        []string      `json:"damagePhotos,omitempty"`
	Location            ClaimLocation `json:"location"`
	VerificationStatus  string        `json:"verificationStatus"`
	VerifiedBy          string        `json:"verifiedBy,omitempty"`
	VerifiedAt          *time.Time    `json:"verifiedAt,omitempty"`
	VerificationNotes   string        `json:"verificationNotes,omitempty"`
	Status              string        `json:"status"`
	PrivacyConsent      bool          `json:"privacyConsent"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
}

// CreateClaimRequest is the payload accepted for citizen claim intake.
type CreateClaimRequest struct {
	IncidentID          string        `json:"incidentId,omitempty"`
	Reporter            ReporterInfo  `json:"reporter"`
	DamageType          string        `json:"damageType"`
	DamageDescription   string        `json:"damageDescription"`
	EstimatedLossAmount string        `json:"estimatedLossAmount"`
	DamagePhotos        []string      `json:"damagePhotos,omitempty"`
	Location            ClaimLocation `json:"location"`
	PrivacyConsent      bool          `json:"privacyConsent"`
}

// VerifyClaimRequest is the payload accepted for authority verification.
type VerifyClaimRequest struct {
	VerificationStatus string `json:"verificationStatus"`
	Notes              string `json:"notes,omitempty"`
}

// UpdateClaimRequest is the payload accepted for authority/citizen updates.
type UpdateClaimRequest struct {
	DamageDescription   *string  `json:"damageDescription,omitempty"`
	EstimatedLossAmount *string  `json:"estimatedLossAmount,omitempty"`
	DamagePhotos        []string `json:"damagePhotos,omitempty"`
}

// CloseClaimRequest is the payload accepted when closing a claim.
type CloseClaimRequest struct {
	Reason string `json:"reason"`
}

// ListClaimsFilter holds the supported query parameters for claim list.
type ListClaimsFilter struct {
	Status             string
	VerificationStatus string
	IncidentID         string
	Query              string
}

// ClaimListResponse is the envelope returned by GET /claims.
type ClaimListResponse struct {
	Claims      []DamageClaimRecord `json:"claims"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

// AuthorityContext carries the authenticated authority actor from request headers.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

// APIError is the standard error envelope returned by the service.
type APIError struct {
	Error APIErrorBody `json:"error"`
}

// APIErrorBody describes a single API error.
type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// IncidentLookupResponse is the shape of incident-service detail used to enrich claims.
type IncidentLookupResponse struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
	Location  struct {
		Address string `json:"address"`
	} `json:"location"`
}
