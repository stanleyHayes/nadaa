package models

import "time"

// Coordinates is a WGS84 latitude/longitude pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// LineStringGeometry is a GeoJSON-style LineString.
type LineStringGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// RoadClosureRecord is the persisted road closure entity.
type RoadClosureRecord struct {
	ID             string             `json:"id"`
	RoadName       string             `json:"roadName"`
	Reason         string             `json:"reason,omitempty"`
	Status         string             `json:"status"`
	Severity       string             `json:"severity"`
	Source         string             `json:"source"`
	SourceRef      string             `json:"sourceRef,omitempty"`
	Geometry       LineStringGeometry `json:"geometry"`
	ValidFrom      time.Time          `json:"validFrom"`
	ValidTo        *time.Time         `json:"validTo,omitempty"`
	DetourNote     string             `json:"detourNote,omitempty"`
	DistanceMeters int                `json:"distanceMeters,omitempty"`
	CreatedBy      string             `json:"createdBy,omitempty"`
	UpdatedBy      string             `json:"updatedBy,omitempty"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
}

// RoadClosureListResponse is the payload returned when listing closures.
type RoadClosureListResponse struct {
	Closures    []RoadClosureRecord `json:"closures"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

// RoadClosureResponse is the payload returned for a single closure.
type RoadClosureResponse struct {
	Closure RoadClosureRecord `json:"closure"`
}

// CreateRoadClosureRequest is the payload to create a new closure.
type CreateRoadClosureRequest struct {
	RoadName   string             `json:"roadName"`
	Reason     string             `json:"reason,omitempty"`
	Status     string             `json:"status"`
	Severity   string             `json:"severity"`
	Source     string             `json:"source,omitempty"`
	SourceRef  string             `json:"sourceRef,omitempty"`
	Geometry   LineStringGeometry `json:"geometry"`
	ValidFrom  *time.Time         `json:"validFrom,omitempty"`
	ValidTo    *time.Time         `json:"validTo,omitempty"`
	DetourNote string             `json:"detourNote,omitempty"`
}

// UpdateRoadClosureRequest is the payload to update an existing closure.
type UpdateRoadClosureRequest struct {
	RoadName   string              `json:"roadName,omitempty"`
	Reason     string              `json:"reason,omitempty"`
	Status     string              `json:"status,omitempty"`
	Severity   string              `json:"severity,omitempty"`
	Source     string              `json:"source,omitempty"`
	SourceRef  string              `json:"sourceRef,omitempty"`
	Geometry   *LineStringGeometry `json:"geometry,omitempty"`
	ValidFrom  *time.Time          `json:"validFrom,omitempty"`
	ValidTo    *time.Time          `json:"validTo,omitempty"`
	DetourNote string              `json:"detourNote,omitempty"`
}

// AdapterImportRequest is the payload to import a closure from an external adapter.
type AdapterImportRequest struct {
	Source    string     `json:"source"`
	SourceRef string     `json:"sourceRef,omitempty"`
	RoadName  string     `json:"roadName"`
	Status    string     `json:"status"`
	Reason    string     `json:"reason,omitempty"`
	Geometry  string     `json:"geometry"`
	ValidFrom time.Time  `json:"validFrom"`
	ValidTo   *time.Time `json:"validTo,omitempty"`
	Detour    string     `json:"detour,omitempty"`
}

// AdapterImportResponse is the payload returned after an adapter import.
type AdapterImportResponse struct {
	Imported    int                 `json:"imported"`
	Closures    []RoadClosureRecord `json:"closures"`
	GeneratedAt time.Time           `json:"generatedAt"`
	Source      string              `json:"source"`
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

// TokenClaims mirrors the claims auth-service signs into bearer tokens. It is
// duplicated here because auth-service is a separate Go module.
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

// BBox is a longitude/latitude bounding box.
type BBox struct {
	MinLat float64
	MinLng float64
	MaxLat float64
	MaxLng float64
}

// ListFilter captures accepted query parameters for listing closures.
type ListFilter struct {
	Status         string
	Location       *Coordinates
	RadiusMeters   float64
	BBox           *BBox
	Limit          int
	IncludeExpired bool
}
