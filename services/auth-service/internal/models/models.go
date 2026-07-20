package models

import "time"

const (
	// DefaultAgencyID is the fixed ID of the bootstrap NADMO agency fixture.
	DefaultAgencyID = "00000000-0000-0000-0000-000000000101"

	// RoleCitizen is the default role for public citizen accounts.
	RoleCitizen = "citizen"
	// RoleAgencyViewer is a read-only authority role.
	RoleAgencyViewer = "agency_viewer"
	// RoleDispatcher can triage, verify, and assign incidents.
	RoleDispatcher = "dispatcher"
	// RoleResponder can update assigned incident status.
	RoleResponder = "responder"
	// RoleNADMOOfficer can manage alerts and incidents across agencies.
	RoleNADMOOfficer = "nadmo_officer"
	// RoleDistrictOfficer can coordinate response within a district.
	RoleDistrictOfficer = "district_officer"
	// RoleAgencyAdmin can manage users within their agency.
	RoleAgencyAdmin = "agency_admin"
	// RoleSystemAdmin has full platform governance access.
	RoleSystemAdmin = "system_admin"
)

// AgencyRoles is the set of authority roles that may be assigned to agency users.
var AgencyRoles = map[string]bool{
	RoleAgencyViewer:    true,
	RoleDispatcher:      true,
	RoleResponder:       true,
	RoleNADMOOfficer:    true,
	RoleDistrictOfficer: true,
	RoleAgencyAdmin:     true,
	RoleSystemAdmin:     true,
}

// Coordinates represents a geographic point.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// AgencyRecord is the persisted representation of a response agency.
type AgencyRecord struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Region        string    `json:"region"`
	District      string    `json:"district"`
	ContactNumber string    `json:"contactNumber,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AgencySummary is the public view of an agency embedded in user profiles.
type AgencySummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Region        string `json:"region"`
	District      string `json:"district"`
	ContactNumber string `json:"contactNumber,omitempty"`
}

// CitizenProfile is the persisted public profile of a registered citizen.
type CitizenProfile struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	Phone             string       `json:"phone"`
	Role              string       `json:"role"`
	PreferredLanguage string       `json:"preferredLanguage"`
	HomeLocation      *Coordinates `json:"homeLocation,omitempty"`
	ContactPermission bool         `json:"contactPermission"`
	CreatedAt         time.Time    `json:"createdAt"`
}

// AgencyUser is the persisted record for an authority user.
type AgencyUser struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	Role         string
	AgencyID     string
	MFARequired  bool
	MFAEnabled   bool
	PasswordHash string
	MFASecret    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AgencyUserProfile is the public profile of an authority user.
type AgencyUserProfile struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	Phone       string        `json:"phone"`
	Role        string        `json:"role"`
	Agency      AgencySummary `json:"agency"`
	MFARequired bool          `json:"mfaRequired"`
	MFAEnabled  bool          `json:"mfaEnabled"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// OTPChallenge tracks a citizen login challenge sent to a phone number.
type OTPChallenge struct {
	ID        string
	Phone     string
	Code      string
	ExpiresAt time.Time
}

// MFAChallenge tracks a pending TOTP enrollment for an agency user. The
// secret is the challenge: verification proves possession of the
// authenticator enrolled with it, so no separate code is stored.
type MFAChallenge struct {
	ID        string
	UserID    string
	Secret    string
	ExpiresAt time.Time
}

// RegisterCitizenRequest is the payload for citizen registration.
type RegisterCitizenRequest struct {
	Name              string       `json:"name"`
	Phone             string       `json:"phone"`
	PreferredLanguage string       `json:"preferredLanguage"`
	HomeLocation      *Coordinates `json:"homeLocation"`
	ContactPermission bool         `json:"contactPermission"`
}

// RegisterCitizenResponse is returned after a citizen registers.
type RegisterCitizenResponse struct {
	UserID      string `json:"userId"`
	Phone       string `json:"phone"`
	ChallengeID string `json:"challengeId"`
	OTPDelivery string `json:"otpDelivery"`
	DevOTP      string `json:"devOtp,omitempty"`
}

// RequestCitizenOTPRequest is the payload requesting a fresh login challenge
// for an already-registered citizen phone.
type RequestCitizenOTPRequest struct {
	Phone string `json:"phone"`
}

// RequestCitizenOTPResponse is returned after a login challenge is issued.
type RequestCitizenOTPResponse struct {
	Phone       string `json:"phone"`
	ChallengeID string `json:"challengeId"`
	OTPDelivery string `json:"otpDelivery"`
	DevOTP      string `json:"devOtp,omitempty"`
}

// LoginCitizenRequest is the payload for citizen login.
type LoginCitizenRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

// LoginCitizenResponse is returned after a successful citizen login.
type LoginCitizenResponse struct {
	AccessToken string         `json:"accessToken"`
	TokenType   string         `json:"tokenType"`
	ExpiresAt   time.Time      `json:"expiresAt"`
	User        CitizenProfile `json:"user"`
}

// CreateAgencyUserRequest is the payload for creating an authority user.
type CreateAgencyUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	AgencyID string `json:"agencyId"`
	Role     string `json:"role"`
}

// CreateAgencyUserResponse is returned after an authority user is created.
type CreateAgencyUserResponse struct {
	User              AgencyUserProfile `json:"user"`
	TemporaryPassword string            `json:"temporaryPassword"`
	MFASetupRequired  bool              `json:"mfaSetupRequired"`
}

// AgencyMFASetupRequest starts MFA setup for an agency user.
type AgencyMFASetupRequest struct {
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporaryPassword"`
}

// AgencyMFASetupResponse returns the pending TOTP enrollment. Secret and
// OTPAuthURL are shown once so the user can enroll an authenticator; DevCode
// carries the current TOTP code only in development (NADAA_ENV=development
// with NADAA_AUTH_EXPOSE_DEV_OTP=true) so automated tests can complete the
// flow.
type AgencyMFASetupResponse struct {
	UserID      string    `json:"userId"`
	ChallengeID string    `json:"challengeId"`
	Method      string    `json:"method"`
	Secret      string    `json:"secret"`
	OTPAuthURL  string    `json:"otpauthUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
	DevCode     string    `json:"devCode,omitempty"`
}

// AgencyMFAVerifyRequest verifies an MFA setup challenge.
type AgencyMFAVerifyRequest struct {
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporaryPassword"`
	Code              string `json:"code"`
}

// AgencyMFAVerifyResponse is returned after MFA verification.
type AgencyMFAVerifyResponse struct {
	User AgencyUserProfile `json:"user"`
}

// LoginAgencyRequest is the payload for authority login.
type LoginAgencyRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MFACode  string `json:"mfaCode"`
}

// LoginAgencyResponse is returned after a successful authority login.
type LoginAgencyResponse struct {
	AccessToken string            `json:"accessToken"`
	TokenType   string            `json:"tokenType"`
	ExpiresAt   time.Time         `json:"expiresAt"`
	User        AgencyUserProfile `json:"user"`
}

// AgencyListResponse is the payload returned when listing the agency directory.
type AgencyListResponse struct {
	Agencies []AgencySummary `json:"agencies"`
}

// ChangeAgencyPasswordRequest is the payload for an agency user password change.
type ChangeAgencyPasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangeAgencyPasswordResponse is returned after a successful password change.
type ChangeAgencyPasswordResponse struct {
	OK bool `json:"ok"`
}

// AgencyUserDirectoryEntry is the sanitized directory view of an agency user.
// It never carries password hashes or MFA secrets.
type AgencyUserDirectoryEntry struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	AgencyID    string     `json:"agencyId"`
	MFAEnabled  bool       `json:"mfaEnabled"`
	LockedUntil *time.Time `json:"lockedUntil,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// AgencyUserListResponse is the payload returned when listing agency users.
type AgencyUserListResponse struct {
	Users []AgencyUserDirectoryEntry `json:"users"`
}

// IngestAuditLogRequest is the payload for service-to-service audit ingestion.
type IngestAuditLogRequest struct {
	EventType    string         `json:"eventType"`
	ActorID      string         `json:"actorId,omitempty"`
	ActorRole    string         `json:"actorRole,omitempty"`
	ResourceType string         `json:"resourceType,omitempty"`
	ResourceID   string         `json:"resourceId,omitempty"`
	Summary      string         `json:"summary,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// IngestAuditLogResponse is returned after an audit event is ingested.
type IngestAuditLogResponse struct {
	ID string `json:"id"`
}

// AuditLogRecord is a single audit event.
type AuditLogRecord struct {
	ID            string         `json:"id"`
	ActorUserID   string         `json:"actorUserId,omitempty"`
	ActorAgencyID string         `json:"actorAgencyId,omitempty"`
	ActorRole     string         `json:"actorRole,omitempty"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId,omitempty"`
	RequestID     string         `json:"requestId,omitempty"`
	IPAddress     string         `json:"ipAddress,omitempty"`
	UserAgent     string         `json:"userAgent,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

// AuditLogListResponse is the payload returned when listing audit logs.
type AuditLogListResponse struct {
	Logs []AuditLogRecord `json:"logs"`
}

// AuditActor identifies who performed an auditable action.
type AuditActor struct {
	UserID   string
	AgencyID string
	Role     string
}

// AuditTarget identifies the object of an auditable action.
type AuditTarget struct {
	Type string
	ID   string
}

// AuditRequestContext captures request metadata for audit logging.
type AuditRequestContext struct {
	RequestID string
	IPAddress string
	UserAgent string
}

// TokenClaims is the signed payload of a NADAA access token.
type TokenClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa,omitempty"`
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
