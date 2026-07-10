package models

import "time"

// Coordinates is a WGS84 latitude/longitude pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// EmergencyContact is a named point of contact for a school.
type EmergencyContact struct {
	Name      string `json:"name"`
	Role      string `json:"role"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	IsPrimary bool   `json:"isPrimary"`
}

// EvacuationPoint is a designated assembly or exit point.
type EvacuationPoint struct {
	Label       string      `json:"label"`
	Location    Coordinates `json:"location"`
	Capacity    int         `json:"capacity,omitempty"`
	Description string      `json:"description,omitempty"`
}

// SchoolProfile is the full emergency-preparedness record for a school.
type SchoolProfile struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Location          Coordinates        `json:"location"`
	Region            string             `json:"region"`
	District          string             `json:"district"`
	Address           string             `json:"address,omitempty"`
	StudentPopulation int                `json:"studentPopulation"`
	EmergencyContacts []EmergencyContact `json:"emergencyContacts"`
	Hazards           []string           `json:"hazards"`
	EvacuationPoints  []EvacuationPoint  `json:"evacuationPoints"`
	CreatedBy         string             `json:"createdBy"`
	UpdatedBy         string             `json:"updatedBy"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`
}

// SchoolSummary is a district-officer list view that omits sensitive details.
type SchoolSummary struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Location          Coordinates `json:"location"`
	District          string      `json:"district"`
	StudentPopulation int         `json:"studentPopulation"`
	ReadinessStatus   string      `json:"readinessStatus"`
	LastDrillDate     *time.Time  `json:"lastDrillDate,omitempty"`
	UpdatedAt         time.Time   `json:"updatedAt"`
}

// DrillRecord tracks an emergency drill conducted at a school.
type DrillRecord struct {
	ID           string    `json:"id"`
	SchoolID     string    `json:"schoolId"`
	Date         time.Time `json:"date"`
	Type         string    `json:"type"`
	Participants int       `json:"participants"`
	Notes        string    `json:"notes,omitempty"`
	Completed    bool      `json:"completed"`
	CreatedBy    string    `json:"createdBy"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ChecklistItem is one item in a readiness checklist.
type ChecklistItem struct {
	Label    string `json:"label"`
	Checked  bool   `json:"checked"`
	Category string `json:"category"`
}

// ReadinessCheck is a periodic preparedness assessment for a school.
type ReadinessCheck struct {
	ID             string          `json:"id"`
	SchoolID       string          `json:"schoolId"`
	CheckDate      time.Time       `json:"checkDate"`
	RiskLevel      string          `json:"riskLevel"`
	AreaRiskRef    string          `json:"areaRiskRef,omitempty"`
	ChecklistItems []ChecklistItem `json:"checklistItems"`
	OverallStatus  string          `json:"overallStatus"`
	Notes          string          `json:"notes,omitempty"`
	CheckedBy      string          `json:"checkedBy"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
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

// SchoolFilter is the authority list filter set.
type SchoolFilter struct {
	District string
	Query    string
}

// SchoolListResponse is the authority response for listing schools.
type SchoolListResponse struct {
	Schools     []SchoolSummary `json:"schools"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

// SchoolDetailResponse wraps a full school profile.
type SchoolDetailResponse struct {
	School      SchoolProfile `json:"school"`
	GeneratedAt time.Time     `json:"generatedAt"`
}

// DrillListResponse is the response for listing drills.
type DrillListResponse struct {
	Drills      []DrillRecord `json:"drills"`
	GeneratedAt time.Time     `json:"generatedAt"`
}

// ReadinessResponse is the response for the latest readiness check.
type ReadinessResponse struct {
	Readiness   *ReadinessCheck `json:"readiness,omitempty"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

// CreateSchoolRequest is the payload to create a school profile.
type CreateSchoolRequest struct {
	Name              string             `json:"name"`
	Location          Coordinates        `json:"location"`
	Region            string             `json:"region"`
	District          string             `json:"district"`
	Address           string             `json:"address,omitempty"`
	StudentPopulation int                `json:"studentPopulation"`
	EmergencyContacts []EmergencyContact `json:"emergencyContacts"`
	Hazards           []string           `json:"hazards"`
	EvacuationPoints  []EvacuationPoint  `json:"evacuationPoints"`
}

// UpdateSchoolRequest is the payload to update a school profile.
type UpdateSchoolRequest struct {
	Name              string             `json:"name,omitempty"`
	Location          *Coordinates       `json:"location,omitempty"`
	Region            string             `json:"region,omitempty"`
	District          string             `json:"district,omitempty"`
	Address           string             `json:"address,omitempty"`
	StudentPopulation *int               `json:"studentPopulation,omitempty"`
	EmergencyContacts []EmergencyContact `json:"emergencyContacts,omitempty"`
	Hazards           []string           `json:"hazards,omitempty"`
	EvacuationPoints  []EvacuationPoint  `json:"evacuationPoints,omitempty"`
}

// CreateDrillRequest is the payload to add a drill record.
type CreateDrillRequest struct {
	Date         time.Time `json:"date"`
	Type         string    `json:"type"`
	Participants int       `json:"participants"`
	Notes        string    `json:"notes,omitempty"`
	Completed    bool      `json:"completed"`
}

// CreateReadinessRequest is the payload to submit a readiness check.
type CreateReadinessRequest struct {
	CheckDate      time.Time       `json:"checkDate"`
	RiskLevel      string          `json:"riskLevel"`
	AreaRiskRef    string          `json:"areaRiskRef,omitempty"`
	ChecklistItems []ChecklistItem `json:"checklistItems"`
	OverallStatus  string          `json:"overallStatus"`
	Notes          string          `json:"notes,omitempty"`
}
