package models

import "time"

// Donor represents a donor offering aid to the NADAA platform.
type Donor struct {
	ID                string    `json:"id"`
	Reference         string    `json:"reference"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	ContactName       string    `json:"contactName"`
	ContactEmail      string    `json:"contactEmail"`
	ContactPhone      string    `json:"contactPhone"`
	Region            string    `json:"region"`
	District          string    `json:"district"`
	ItemsOffered      []string  `json:"itemsOffered"`
	MonetaryPledgeGhs float64   `json:"monetaryPledgeGhs"`
	Status            string    `json:"status"`
	Notes             string    `json:"notes,omitempty"`
	CreatedBy         string    `json:"createdBy,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// AidCatalogItem represents a standardized aid item in the catalog.
type AidCatalogItem struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	DefaultUnit   string  `json:"defaultUnit"`
	PriorityScore float64 `json:"priorityScore"`
}

// AidRequest represents a request for aid from an authority or beneficiary.
type AidRequest struct {
	ID                string    `json:"id"`
	Reference         string    `json:"reference"`
	Title             string    `json:"title"`
	Description       string    `json:"description,omitempty"`
	Category          string    `json:"category"`
	ItemCode          string    `json:"itemCode"`
	QuantityNeeded    int       `json:"quantityNeeded"`
	QuantityFulfilled int       `json:"quantityFulfilled"`
	Unit              string    `json:"unit"`
	Priority          string    `json:"priority"`
	LocationLabel     string    `json:"locationLabel,omitempty"`
	Region            string    `json:"region"`
	District          string    `json:"district"`
	BeneficiaryCount  int       `json:"beneficiaryCount"`
	Status            string    `json:"status"`
	RequestedBy       string    `json:"requestedBy,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// Pledge represents a donor's commitment to fulfill part of an aid request.
type Pledge struct {
	ID                string    `json:"id"`
	Reference         string    `json:"reference"`
	AidRequestID      string    `json:"aidRequestId"`
	DonorID           string    `json:"donorId"`
	DonorName         string    `json:"donorName"`
	QuantityPledged   int       `json:"quantityPledged"`
	QuantityDelivered int       `json:"quantityDelivered"`
	Status            string    `json:"status"`
	DeliveryNote      string    `json:"deliveryNote,omitempty"`
	ContactEmail      string    `json:"contactEmail,omitempty"`
	ContactPhone      string    `json:"contactPhone,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// AuthorityContext holds authenticated authority metadata from request headers.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	ActorDistrict string
	MFACompleted  bool
	RequestID     string
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

// DonorListResponse is the response payload for listing donors.
type DonorListResponse struct {
	Donors      []Donor   `json:"donors"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// AidCatalogResponse is the response payload for listing aid catalog items.
type AidCatalogResponse struct {
	Items       []AidCatalogItem `json:"items"`
	GeneratedAt time.Time        `json:"generatedAt"`
}

// AidRequestListResponse is the response payload for listing aid requests.
type AidRequestListResponse struct {
	Requests    []AidRequest `json:"requests"`
	GeneratedAt time.Time    `json:"generatedAt"`
}

// PledgeListResponse is the response payload for listing pledges.
type PledgeListResponse struct {
	Pledges     []Pledge  `json:"pledges"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// CreateDonorRequest is the payload for creating a donor.
type CreateDonorRequest struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ContactName       string   `json:"contactName,omitempty"`
	ContactEmail      string   `json:"contactEmail,omitempty"`
	ContactPhone      string   `json:"contactPhone,omitempty"`
	Region            string   `json:"region,omitempty"`
	District          string   `json:"district,omitempty"`
	ItemsOffered      []string `json:"itemsOffered,omitempty"`
	MonetaryPledgeGhs float64  `json:"monetaryPledgeGhs,omitempty"`
	Notes             string   `json:"notes,omitempty"`
}

// UpdateDonorRequest is the payload for updating a donor.
type UpdateDonorRequest struct {
	Status string `json:"status,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

// CreateAidRequestRequest is the payload for creating an aid request.
type CreateAidRequestRequest struct {
	Title            string `json:"title"`
	Description      string `json:"description,omitempty"`
	Category         string `json:"category"`
	ItemCode         string `json:"itemCode"`
	QuantityNeeded   int    `json:"quantityNeeded"`
	Unit             string `json:"unit"`
	Priority         string `json:"priority"`
	LocationLabel    string `json:"locationLabel,omitempty"`
	Region           string `json:"region"`
	District         string `json:"district"`
	BeneficiaryCount int    `json:"beneficiaryCount,omitempty"`
}

// UpdateAidRequestRequest is the payload for updating an aid request.
type UpdateAidRequestRequest struct {
	Status         string `json:"status,omitempty"`
	QuantityNeeded int    `json:"quantityNeeded,omitempty"`
}

// CreatePledgeRequest is the payload for creating a pledge.
type CreatePledgeRequest struct {
	DonorID         string `json:"donorId"`
	DonorName       string `json:"donorName,omitempty"`
	QuantityPledged int    `json:"quantityPledged"`
	ContactEmail    string `json:"contactEmail,omitempty"`
	ContactPhone    string `json:"contactPhone,omitempty"`
	DeliveryNote    string `json:"deliveryNote,omitempty"`
}

// UpdatePledgeRequest is the payload for updating a pledge.
type UpdatePledgeRequest struct {
	Status            string `json:"status,omitempty"`
	QuantityDelivered int    `json:"quantityDelivered,omitempty"`
	DeliveryNote      string `json:"deliveryNote,omitempty"`
}

// AllocateRequest is the payload for allocating a pledge to an aid request.
type AllocateRequest struct {
	PledgeID string `json:"pledgeId"`
	Quantity int    `json:"quantity"`
}

// DonorFilter is the set of filters for listing donors.
type DonorFilter struct {
	Type  string
	Query string
}

// AidRequestFilter is the set of filters for listing aid requests.
type AidRequestFilter struct {
	Status   string
	Category string
	Region   string
	Priority string
}

// PledgeFilter is the set of filters for listing pledges.
type PledgeFilter struct {
	Status string
}
