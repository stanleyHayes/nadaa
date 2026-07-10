package models

import "time"

// OpenDataCategory groups datasets by hazard or domain.
type OpenDataCategory string

// Open data category values.
const (
	OpenDataCategoryFlood       OpenDataCategory = "flood"
	OpenDataCategoryFire        OpenDataCategory = "fire"
	OpenDataCategoryRoadClosure OpenDataCategory = "road_closure"
	OpenDataCategoryWeather     OpenDataCategory = "weather"
	OpenDataCategoryShelter     OpenDataCategory = "shelter"
	OpenDataCategoryIncident    OpenDataCategory = "incident"
	OpenDataCategoryRelief      OpenDataCategory = "relief"
	OpenDataCategoryRisk        OpenDataCategory = "risk"
	OpenDataCategoryOther       OpenDataCategory = "other"
)

// PrivacyReviewStatus describes the governance stage of a dataset.
type PrivacyReviewStatus string

// Privacy review status values.
const (
	PrivacyReviewPending  PrivacyReviewStatus = "pending_review"
	PrivacyReviewApproved PrivacyReviewStatus = "approved"
	PrivacyReviewRejected PrivacyReviewStatus = "rejected"
)

// AnonymizationLevel describes how identity is protected in a dataset.
type AnonymizationLevel string

// Anonymization level values.
const (
	AnonymizationNone       AnonymizationLevel = "none"
	AnonymizationAggregated AnonymizationLevel = "aggregated"
	AnonymizationAnonymized AnonymizationLevel = "anonymized"
	AnonymizationSynthetic  AnonymizationLevel = "synthetic"
)

// DatasetLicense is a known open data license.
type DatasetLicense string

// Dataset license values.
const (
	LicenseOpenDataCommonsODCOpen DatasetLicense = "ODC-Open"
	LicenseCreativeCommonsBY40    DatasetLicense = "CC-BY-4.0"
	LicenseCreativeCommonsBYSA40  DatasetLicense = "CC-BY-SA-4.0"
	LicenseGhanaOpenGovernment    DatasetLicense = "Ghana-Open-Government"
	LicensePublicDomain           DatasetLicense = "public-domain"
)

// UpdateFrequency describes how often a dataset is refreshed.
type UpdateFrequency string

// Update frequency values.
const (
	UpdateFrequencyRealtime  UpdateFrequency = "realtime"
	UpdateFrequencyHourly    UpdateFrequency = "hourly"
	UpdateFrequencyDaily     UpdateFrequency = "daily"
	UpdateFrequencyWeekly    UpdateFrequency = "weekly"
	UpdateFrequencyMonthly   UpdateFrequency = "monthly"
	UpdateFrequencyQuarterly UpdateFrequency = "quarterly"
	UpdateFrequencyYearly    UpdateFrequency = "yearly"
	UpdateFrequencyAdHoc     UpdateFrequency = "ad_hoc"
	UpdateFrequencyStatic    UpdateFrequency = "static"
)

// DatasetMetadata holds flexible dataset metadata.
type DatasetMetadata struct {
	Publisher          string            `json:"publisher"`
	ContactEmail       string            `json:"contactEmail,omitempty"`
	RegionCoverage     []string          `json:"regionCoverage,omitempty"`
	TemporalCoverage   string            `json:"temporalCoverage,omitempty"`
	SpatialResolution  string            `json:"spatialResolution,omitempty"`
	Keywords           []string          `json:"keywords,omitempty"`
	SourceSystems      []string          `json:"sourceSystems,omitempty"`
	AnonymizationNotes string            `json:"anonymizationNotes,omitempty"`
	Additional         map[string]string `json:"additional,omitempty"`
}

// Dataset is an approved, anonymized open data catalog entry.
type Dataset struct {
	ID                  string              `json:"id"`
	Title               string              `json:"title"`
	Description         string              `json:"description"`
	Category            OpenDataCategory    `json:"category"`
	License             DatasetLicense      `json:"license"`
	UpdateFrequency     UpdateFrequency     `json:"updateFrequency"`
	PrivacyReviewStatus PrivacyReviewStatus `json:"privacyReviewStatus"`
	AnonymizationLevel  AnonymizationLevel  `json:"anonymizationLevel"`
	Metadata            DatasetMetadata     `json:"metadata"`
	SampleRows          []map[string]any    `json:"sampleRows,omitempty"`
	Columns             []DatasetColumn     `json:"columns,omitempty"`
	AccessRestriction   string              `json:"accessRestriction,omitempty"`
	CreatedAt           time.Time           `json:"createdAt"`
	UpdatedAt           time.Time           `json:"updatedAt"`
}

// DatasetColumn describes a column in a dataset.
type DatasetColumn struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Nullable    bool   `json:"nullable"`
}

// DatasetDownload is a downloadable artifact for a dataset.
type DatasetDownload struct {
	ID        string    `json:"id"`
	DatasetID string    `json:"datasetId"`
	Format    string    `json:"format"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"createdAt"`
}

// OpenDataRequestStatus is the lifecycle state of an access request.
type OpenDataRequestStatus string

// Open data request status values.
const (
	OpenDataRequestPending  OpenDataRequestStatus = "pending"
	OpenDataRequestApproved OpenDataRequestStatus = "approved"
	OpenDataRequestRejected OpenDataRequestStatus = "rejected"
	OpenDataRequestExpired  OpenDataRequestStatus = "expired"
)

// OpenDataRequest is a request to access a restricted dataset.
type OpenDataRequest struct {
	ID            string                `json:"id"`
	DatasetID     string                `json:"datasetId"`
	RequesterInfo RequesterInfo         `json:"requesterInfo"`
	Purpose       string                `json:"purpose"`
	Status        OpenDataRequestStatus `json:"status"`
	CreatedAt     time.Time             `json:"createdAt"`
	ReviewedAt    *time.Time            `json:"reviewedAt,omitempty"`
	ReviewedBy    string                `json:"reviewedBy,omitempty"`
	ReviewNote    string                `json:"reviewNote,omitempty"`
}

// RequesterInfo captures minimal requester details without storing PII.
type RequesterInfo struct {
	Name         string `json:"name"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email"`
	UseCase      string `json:"useCase"`
}

// DatasetListResponse returns available datasets.
type DatasetListResponse struct {
	Datasets    []Dataset `json:"datasets"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// DatasetDetailResponse returns a single dataset.
type DatasetDetailResponse struct {
	Dataset Dataset `json:"dataset"`
}

// DatasetDownloadResponse returns the download artifact.
type DatasetDownloadResponse struct {
	Download    DatasetDownload `json:"download"`
	RateLimit   RateLimitStatus `json:"rateLimit"`
	AuditLogged bool            `json:"auditLogged"`
}

// RateLimitStatus reports current rate limit state.
type RateLimitStatus struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"resetAt"`
}

// CreateOpenDataRequest is the payload to request dataset access.
type CreateOpenDataRequest struct {
	DatasetID     string        `json:"datasetId"`
	RequesterInfo RequesterInfo `json:"requesterInfo"`
	Purpose       string        `json:"purpose"`
}

// OpenDataRequestResponse returns a created request.
type OpenDataRequestResponse struct {
	Request OpenDataRequest `json:"request"`
}

// OpenDataRequestListResponse returns requests for admin review.
type OpenDataRequestListResponse struct {
	Requests    []OpenDataRequest `json:"requests"`
	GeneratedAt time.Time         `json:"generatedAt"`
}

// ReviewOpenDataRequest is the payload to approve or reject a request.
type ReviewOpenDataRequest struct {
	Reviewer string `json:"reviewer"`
	Approved bool   `json:"approved"`
	Note     string `json:"note,omitempty"`
}

// AuditEvent is a lightweight audit event sent to the audit log service.
type AuditEvent struct {
	Action     string            `json:"action"`
	TargetType string            `json:"targetType"`
	TargetID   string            `json:"targetId"`
	ActorRole  string            `json:"actorRole,omitempty"`
	IPAddress  string            `json:"ipAddress,omitempty"`
	UserAgent  string            `json:"userAgent,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
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
