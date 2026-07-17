package models

import "time"

// CampaignContentBlock is a single unit of campaign content.
type CampaignContentBlock struct {
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Body     string   `json:"body,omitempty"`
	Items    []string `json:"items,omitempty"`
	MediaURL string   `json:"mediaUrl,omitempty"`
}

// CampaignPublishWindow defines when a campaign is considered active.
type CampaignPublishWindow struct {
	StartsAt time.Time `json:"startsAt"`
	EndsAt   time.Time `json:"endsAt"`
}

// Campaign is a preparedness campaign published by an authority.
type Campaign struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	HazardType     string                 `json:"hazardType"`
	TargetRegions  []string               `json:"targetRegions"`
	Languages      []string               `json:"languages"`
	ContentBlocks  []CampaignContentBlock `json:"contentBlocks"`
	PublishWindow  CampaignPublishWindow  `json:"publishWindow"`
	Status         string                 `json:"status"`
	LinkedGuideIDs []string               `json:"linkedGuideIds,omitempty"`
	LinkedAlertIDs []string               `json:"linkedAlertIds,omitempty"`
	CreatedBy      string                 `json:"createdBy,omitempty"`
	UpdatedBy      string                 `json:"updatedBy,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// CampaignMetric holds mocked reach and engagement for a campaign day.
type CampaignMetric struct {
	ID         string    `json:"id"`
	CampaignID string    `json:"campaignId"`
	Date       time.Time `json:"date"`
	Reach      int       `json:"reach"`
	Engagement int       `json:"engagement"`
}

// CampaignTemplate is a reusable seasonal starting point for campaigns.
type CampaignTemplate struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	HazardType     string                 `json:"hazardType"`
	Season         string                 `json:"season"`
	DefaultContent []CampaignContentBlock `json:"defaultContent"`
}

// CampaignFilters captures accepted query parameters for listing campaigns.
type CampaignFilters struct {
	Region     string
	Language   string
	HazardType string
	Status     string
	IncludeAll bool
}

// CreateCampaignRequest is the payload to create a new campaign.
type CreateCampaignRequest struct {
	Title          string                 `json:"title"`
	HazardType     string                 `json:"hazardType"`
	TargetRegions  []string               `json:"targetRegions"`
	Languages      []string               `json:"languages"`
	ContentBlocks  []CampaignContentBlock `json:"contentBlocks"`
	PublishWindow  CampaignPublishWindow  `json:"publishWindow"`
	Status         string                 `json:"status"`
	LinkedGuideIDs []string               `json:"linkedGuideIds,omitempty"`
	LinkedAlertIDs []string               `json:"linkedAlertIds,omitempty"`
}

// UpdateCampaignRequest is the payload to update an existing campaign.
type UpdateCampaignRequest struct {
	Title          string                 `json:"title,omitempty"`
	HazardType     string                 `json:"hazardType,omitempty"`
	TargetRegions  []string               `json:"targetRegions,omitempty"`
	Languages      []string               `json:"languages,omitempty"`
	ContentBlocks  []CampaignContentBlock `json:"contentBlocks,omitempty"`
	PublishWindow  *CampaignPublishWindow `json:"publishWindow,omitempty"`
	Status         string                 `json:"status,omitempty"`
	LinkedGuideIDs []string               `json:"linkedGuideIds,omitempty"`
	LinkedAlertIDs []string               `json:"linkedAlertIds,omitempty"`
}

// CampaignListResponse is the payload returned when listing campaigns.
type CampaignListResponse struct {
	Campaigns   []Campaign `json:"campaigns"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

// CampaignResponse is the payload returned for a single campaign.
type CampaignResponse struct {
	Campaign Campaign `json:"campaign"`
}

// CampaignMetricListResponse is the payload returned for campaign metrics.
type CampaignMetricListResponse struct {
	Metrics     []CampaignMetric `json:"metrics"`
	CampaignID  string           `json:"campaignId"`
	GeneratedAt time.Time        `json:"generatedAt"`
}

// CampaignTemplateListResponse is the payload returned when listing templates.
type CampaignTemplateListResponse struct {
	Templates   []CampaignTemplate `json:"templates"`
	GeneratedAt time.Time          `json:"generatedAt"`
}

// AuthorityContext holds authenticated actor metadata from request headers.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	ActorDistrict string
	MFACompleted  bool
	RequestID     string
}

// TokenClaims mirrors the claims auth-service signs into NADAA bearer tokens
// (nadaa.<payload>.<sig>); the payload JSON tags must stay in sync with it.
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
