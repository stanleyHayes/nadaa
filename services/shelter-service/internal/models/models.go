package models

import "time"

// Coordinates represents a latitude/longitude pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Shelter is an evacuation or relief shelter record.
type Shelter struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Type             string      `json:"type"`
	Region           string      `json:"region"`
	District         string      `json:"district"`
	Address          string      `json:"address"`
	Location         Coordinates `json:"location"`
	Capacity         int         `json:"capacity"`
	CurrentOccupancy int         `json:"currentOccupancy"`
	Status           string      `json:"status"`
	Contact          string      `json:"contact"`
	Facilities       []string    `json:"facilities"`
	Notes            string      `json:"notes,omitempty"`
	DistanceMeters   int         `json:"distanceMeters,omitempty"`
	UpdatedBy        string      `json:"updatedBy,omitempty"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

// RecoverySupportLocation is a recovery desk or support point.
type RecoverySupportLocation struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Region         string      `json:"region"`
	District       string      `json:"district"`
	Address        string      `json:"address"`
	Location       Coordinates `json:"location"`
	Contact        string      `json:"contact"`
	Services       []string    `json:"services"`
	Hours          string      `json:"hours"`
	Status         string      `json:"status"`
	DistanceMeters int         `json:"distanceMeters,omitempty"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

// ReliefPoint is a distribution point for food, water, or other aid.
type ReliefPoint struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	Region          string                `json:"region"`
	District        string                `json:"district"`
	Address         string                `json:"address"`
	Location        Coordinates           `json:"location"`
	Contact         string                `json:"contact"`
	OperatingHours  string                `json:"operatingHours"`
	Eligibility     string                `json:"eligibility"`
	Schedule        string                `json:"schedule"`
	StockCategories []ReliefStockCategory `json:"stockCategories"`
	Status          string                `json:"status"`
	Source          string                `json:"source"`
	SourceRef       string                `json:"sourceRef,omitempty"`
	DistanceMeters  int                   `json:"distanceMeters,omitempty"`
	CreatedBy       string                `json:"createdBy,omitempty"`
	UpdatedBy       string                `json:"updatedBy,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

// ReliefStockCategory describes a stock line at a relief point.
type ReliefStockCategory struct {
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	Unit        string    `json:"unit"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// ReliefStockHistory captures stock category changes over time.
type ReliefStockHistory struct {
	ID              string                `json:"id"`
	ReliefPointID   string                `json:"reliefPointId"`
	ChangedBy       string                `json:"changedBy"`
	ChangedAt       time.Time             `json:"changedAt"`
	Note            string                `json:"note,omitempty"`
	StockCategories []ReliefStockCategory `json:"stockCategories"`
}

// AidRequest is a request for aid from a receiving organization.
type AidRequest struct {
	ID                    string      `json:"id"`
	Title                 string      `json:"title"`
	Category              string      `json:"category"`
	Priority              string      `json:"priority"`
	Status                string      `json:"status"`
	Region                string      `json:"region"`
	District              string      `json:"district"`
	Location              Coordinates `json:"location"`
	ReceivingOrganization string      `json:"receivingOrganization"`
	Contact               string      `json:"contact"`
	QuantityNeeded        int         `json:"quantityNeeded"`
	QuantityUnit          string      `json:"quantityUnit"`
	QuantityPledged       int         `json:"quantityPledged"`
	Description           string      `json:"description"`
	NeededBy              time.Time   `json:"neededBy"`
	Visibility            string      `json:"visibility"`
	SourceReliefPointID   string      `json:"sourceReliefPointId,omitempty"`
	AgencyID              string      `json:"agencyId,omitempty"`
	CreatedBy             string      `json:"createdBy"`
	ApprovedBy            string      `json:"approvedBy,omitempty"`
	ApprovalNotes         string      `json:"approvalNotes,omitempty"`
	AntiFraudNotes        string      `json:"antiFraudNotes,omitempty"`
	Pledges               []AidPledge `json:"pledges"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
}

// AidPledge is a donor pledge against an aid request.
type AidPledge struct {
	ID               string    `json:"id"`
	AidRequestID     string    `json:"aidRequestId"`
	DonorName        string    `json:"donorName"`
	DonorType        string    `json:"donorType"`
	Contact          string    `json:"contact"`
	Quantity         int       `json:"quantity"`
	Unit             string    `json:"unit"`
	Note             string    `json:"note,omitempty"`
	Status           string    `json:"status"`
	ReviewStatus     string    `json:"reviewStatus"`
	FraudReviewNotes string    `json:"fraudReviewNotes,omitempty"`
	ReviewedBy       string    `json:"reviewedBy,omitempty"`
	PledgedAt        time.Time `json:"pledgedAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// HospitalCapacity tracks bed and emergency capacity for a hospital.
type HospitalCapacity struct {
	ID                     string      `json:"id"`
	Name                   string      `json:"name"`
	Type                   string      `json:"type"`
	Region                 string      `json:"region"`
	District               string      `json:"district"`
	Address                string      `json:"address"`
	Location               Coordinates `json:"location"`
	Contact                string      `json:"contact"`
	Services               []string    `json:"services"`
	TotalBeds              int         `json:"totalBeds"`
	AvailableBeds          int         `json:"availableBeds"`
	ICUBedsAvailable       int         `json:"icuBedsAvailable"`
	MaternityBedsAvailable int         `json:"maternityBedsAvailable"`
	PediatricBedsAvailable int         `json:"pediatricBedsAvailable"`
	IsolationBedsAvailable int         `json:"isolationBedsAvailable"`
	EmergencyCapacity      string      `json:"emergencyCapacity"`
	EmergencyUnitStatus    string      `json:"emergencyUnitStatus"`
	AmbulancesAvailable    int         `json:"ambulancesAvailable"`
	OxygenAvailable        bool        `json:"oxygenAvailable"`
	Notes                  string      `json:"notes,omitempty"`
	Source                 string      `json:"source"`
	SourceRef              string      `json:"sourceRef,omitempty"`
	UpdatedBy              string      `json:"updatedBy,omitempty"`
	UpdatedAt              time.Time   `json:"updatedAt"`
	DistanceMeters         int         `json:"distanceMeters,omitempty"`
	Stale                  bool        `json:"stale"`
	StaleReason            string      `json:"staleReason,omitempty"`
}

// ShelterListResponse is the payload for listing shelters.
type ShelterListResponse struct {
	Shelters    []Shelter `json:"shelters"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// NearbyShelterResponse returns nearby shelters and recovery support locations.
type NearbyShelterResponse struct {
	Shelters        []Shelter                 `json:"shelters"`
	RecoverySupport []RecoverySupportLocation `json:"recoverySupport"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
}

// RecoverySupportResponse returns nearby recovery support locations.
type RecoverySupportResponse struct {
	RecoverySupport []RecoverySupportLocation `json:"recoverySupport"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
}

// ReliefPointListResponse is the payload for listing relief points.
type ReliefPointListResponse struct {
	ReliefPoints []ReliefPoint `json:"reliefPoints"`
	GeneratedAt  time.Time     `json:"generatedAt"`
}

// ReliefPointNearbyResponse is the payload for nearby relief points.
type ReliefPointNearbyResponse struct {
	ReliefPoints []ReliefPoint `json:"reliefPoints"`
	GeneratedAt  time.Time     `json:"generatedAt"`
}

// ReliefPointStockHistoryResponse returns stock history for a relief point.
type ReliefPointStockHistoryResponse struct {
	ReliefPointID string               `json:"reliefPointId"`
	History       []ReliefStockHistory `json:"history"`
	GeneratedAt   time.Time            `json:"generatedAt"`
}

// AidRequestListResponse is the payload for listing aid requests.
type AidRequestListResponse struct {
	AidRequests []AidRequest `json:"aidRequests"`
	GeneratedAt time.Time    `json:"generatedAt"`
}

// PublicAidRequest is the anonymous-safe view of an aid request. It carries
// only the fields donors need and omits pledge records (donor contact PII and
// fraud-review internals) and authority-only review metadata.
type PublicAidRequest struct {
	ID                    string      `json:"id"`
	Title                 string      `json:"title"`
	Category              string      `json:"category"`
	Priority              string      `json:"priority"`
	Status                string      `json:"status"`
	Region                string      `json:"region"`
	District              string      `json:"district"`
	Location              Coordinates `json:"location"`
	ReceivingOrganization string      `json:"receivingOrganization"`
	Contact               string      `json:"contact"`
	QuantityNeeded        int         `json:"quantityNeeded"`
	QuantityUnit          string      `json:"quantityUnit"`
	QuantityPledged       int         `json:"quantityPledged"`
	Description           string      `json:"description"`
	NeededBy              time.Time   `json:"neededBy"`
	Visibility            string      `json:"visibility"`
	SourceReliefPointID   string      `json:"sourceReliefPointId,omitempty"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
}

// PublicAidRequestListResponse is the payload for anonymous aid request listings.
type PublicAidRequestListResponse struct {
	AidRequests []PublicAidRequest `json:"aidRequests"`
	GeneratedAt time.Time          `json:"generatedAt"`
}

// AidPledgeListResponse is the payload for listing pledges.
type AidPledgeListResponse struct {
	AidRequestID string      `json:"aidRequestId"`
	Pledges      []AidPledge `json:"pledges"`
	GeneratedAt  time.Time   `json:"generatedAt"`
}

// CreateReliefPointRequest creates a new relief point.
type CreateReliefPointRequest struct {
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	Region          string                `json:"region,omitempty"`
	District        string                `json:"district,omitempty"`
	Address         string                `json:"address,omitempty"`
	Location        Coordinates           `json:"location"`
	Contact         string                `json:"contact,omitempty"`
	OperatingHours  string                `json:"operatingHours,omitempty"`
	Eligibility     string                `json:"eligibility,omitempty"`
	Schedule        string                `json:"schedule,omitempty"`
	StockCategories []ReliefStockCategory `json:"stockCategories,omitempty"`
	Status          string                `json:"status,omitempty"`
	Source          string                `json:"source,omitempty"`
	SourceRef       string                `json:"sourceRef,omitempty"`
}

// UpdateReliefPointRequest updates an existing relief point.
type UpdateReliefPointRequest struct {
	Name            string                `json:"name,omitempty"`
	Type            string                `json:"type,omitempty"`
	Region          string                `json:"region,omitempty"`
	District        string                `json:"district,omitempty"`
	Address         string                `json:"address,omitempty"`
	Location        *Coordinates          `json:"location,omitempty"`
	Contact         string                `json:"contact,omitempty"`
	OperatingHours  string                `json:"operatingHours,omitempty"`
	Eligibility     string                `json:"eligibility,omitempty"`
	Schedule        string                `json:"schedule,omitempty"`
	StockCategories []ReliefStockCategory `json:"stockCategories,omitempty"`
	Status          string                `json:"status,omitempty"`
	SourceRef       string                `json:"sourceRef,omitempty"`
}

// CreateAidRequestRequest creates a new aid request.
type CreateAidRequestRequest struct {
	Title                 string      `json:"title"`
	Category              string      `json:"category"`
	Priority              string      `json:"priority"`
	Region                string      `json:"region,omitempty"`
	District              string      `json:"district,omitempty"`
	Location              Coordinates `json:"location"`
	ReceivingOrganization string      `json:"receivingOrganization"`
	Contact               string      `json:"contact,omitempty"`
	QuantityNeeded        int         `json:"quantityNeeded"`
	QuantityUnit          string      `json:"quantityUnit"`
	Description           string      `json:"description"`
	NeededBy              time.Time   `json:"neededBy"`
	Visibility            string      `json:"visibility,omitempty"`
	SourceReliefPointID   string      `json:"sourceReliefPointId,omitempty"`
}

// ReviewAidRequestRequest reviews an aid request.
type ReviewAidRequestRequest struct {
	Status         string `json:"status"`
	ApprovalNotes  string `json:"approvalNotes,omitempty"`
	AntiFraudNotes string `json:"antiFraudNotes,omitempty"`
}

// CreateAidPledgeRequest creates a pledge against an aid request.
type CreateAidPledgeRequest struct {
	DonorName string `json:"donorName"`
	DonorType string `json:"donorType"`
	Contact   string `json:"contact"`
	Quantity  int    `json:"quantity"`
	Unit      string `json:"unit"`
	Note      string `json:"note,omitempty"`
}

// ReviewAidPledgeRequest reviews an aid pledge.
type ReviewAidPledgeRequest struct {
	Status           string `json:"status,omitempty"`
	ReviewStatus     string `json:"reviewStatus,omitempty"`
	FraudReviewNotes string `json:"fraudReviewNotes,omitempty"`
}

// OccupancyUpdateRequest updates shelter occupancy.
type OccupancyUpdateRequest struct {
	Capacity         *int   `json:"capacity,omitempty"`
	CurrentOccupancy *int   `json:"currentOccupancy,omitempty"`
	Status           string `json:"status,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

// ShelterUpdateResponse returns an updated shelter.
type ShelterUpdateResponse struct {
	Shelter Shelter `json:"shelter"`
}

// HospitalCapacityResponse returns hospital capacity listings.
type HospitalCapacityResponse struct {
	Facilities            []HospitalCapacity `json:"facilities"`
	GeneratedAt           time.Time          `json:"generatedAt"`
	StaleThresholdMinutes int                `json:"staleThresholdMinutes"`
}

// HospitalCapacityUpdateRequest updates hospital capacity.
type HospitalCapacityUpdateRequest struct {
	TotalBeds              *int   `json:"totalBeds,omitempty"`
	AvailableBeds          *int   `json:"availableBeds,omitempty"`
	ICUBedsAvailable       *int   `json:"icuBedsAvailable,omitempty"`
	MaternityBedsAvailable *int   `json:"maternityBedsAvailable,omitempty"`
	PediatricBedsAvailable *int   `json:"pediatricBedsAvailable,omitempty"`
	IsolationBedsAvailable *int   `json:"isolationBedsAvailable,omitempty"`
	EmergencyCapacity      string `json:"emergencyCapacity,omitempty"`
	EmergencyUnitStatus    string `json:"emergencyUnitStatus,omitempty"`
	AmbulancesAvailable    *int   `json:"ambulancesAvailable,omitempty"`
	OxygenAvailable        *bool  `json:"oxygenAvailable,omitempty"`
	Notes                  string `json:"notes,omitempty"`
	Source                 string `json:"source,omitempty"`
	SourceRef              string `json:"sourceRef,omitempty"`
}

// HospitalCapacityUpdateResponse returns an updated hospital.
type HospitalCapacityUpdateResponse struct {
	Facility HospitalCapacity `json:"facility"`
}

// HospitalCapacityImportRequest imports fixture hospital capacity data.
type HospitalCapacityImportRequest struct {
	Source    string                          `json:"source,omitempty"`
	SourceRef string                          `json:"sourceRef,omitempty"`
	Records   []HospitalCapacityFixtureRecord `json:"records,omitempty"`
}

// HospitalCapacityFixtureRecord is a single fixture record for import.
type HospitalCapacityFixtureRecord struct {
	FacilityID             string `json:"facilityId"`
	AvailableBeds          int    `json:"availableBeds"`
	ICUBedsAvailable       int    `json:"icuBedsAvailable,omitempty"`
	MaternityBedsAvailable int    `json:"maternityBedsAvailable,omitempty"`
	PediatricBedsAvailable int    `json:"pediatricBedsAvailable,omitempty"`
	IsolationBedsAvailable int    `json:"isolationBedsAvailable,omitempty"`
	EmergencyCapacity      string `json:"emergencyCapacity"`
	EmergencyUnitStatus    string `json:"emergencyUnitStatus,omitempty"`
	AmbulancesAvailable    int    `json:"ambulancesAvailable,omitempty"`
	OxygenAvailable        *bool  `json:"oxygenAvailable,omitempty"`
	Notes                  string `json:"notes,omitempty"`
}

// HospitalCapacityImportResponse returns imported hospital capacity data.
type HospitalCapacityImportResponse struct {
	Imported    int                `json:"imported"`
	Facilities  []HospitalCapacity `json:"facilities"`
	GeneratedAt time.Time          `json:"generatedAt"`
	Source      string             `json:"source"`
}

// AuthorityContext carries authenticated authority metadata.
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

// AidRequestFilter captures accepted query parameters for listing aid requests.
type AidRequestFilter struct {
	Category       string
	Priority       string
	Status         string
	Region         string
	District       string
	IncludePrivate bool
	Location       *Coordinates
	RadiusMeters   float64
	Limit          int
	// ViewerRole and ViewerAgencyID scope private (non-public) results when
	// IncludePrivate is set: privileged roles see every private request, while
	// agency roles only see private requests owned by their own agency.
	ViewerRole     string
	ViewerAgencyID string
}

// HospitalCapacityFilter captures accepted query parameters for hospital capacity.
type HospitalCapacityFilter struct {
	Location          *Coordinates
	Service           string
	EmergencyCapacity string
	MinAvailableBeds  int
	IncludeStale      bool
	Limit             int
}

// BoundingBox describes a geographic bounding box.
type BoundingBox struct {
	MinLat float64
	MinLng float64
	MaxLat float64
	MaxLng float64
}

// ReliefPointFilter captures accepted query parameters for listing relief points.
type ReliefPointFilter struct {
	Status       string
	Type         string
	Location     *Coordinates
	RadiusMeters float64
	BBox         *BoundingBox
	Limit        int
}
