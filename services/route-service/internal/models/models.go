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

// RouteSegment describes one leg of a planned route.
type RouteSegment struct {
	Start          Coordinates `json:"start"`
	End            Coordinates `json:"end"`
	DistanceMeters int         `json:"distanceMeters"`
	Mode           string      `json:"mode"`
}

// Shelter is a minimal shelter representation returned by the shelter-service.
type Shelter struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Location Coordinates `json:"location"`
	Status   string      `json:"status"`
}

// NearbyShelterResponse is the payload returned by the shelter-service nearby endpoint.
type NearbyShelterResponse struct {
	Shelters    []Shelter `json:"shelters"`
	GeneratedAt time.Time `json:"generatedAt"`
}

// RoadClosure is a minimal road closure representation.
type RoadClosure struct {
	ID       string             `json:"id"`
	Status   string             `json:"status"`
	Severity string             `json:"severity"`
	Geometry LineStringGeometry `json:"geometry"`
}

// RoadClosureListResponse is the payload returned by the road-closure-service.
type RoadClosureListResponse struct {
	Closures    []RoadClosure `json:"closures"`
	GeneratedAt time.Time     `json:"generatedAt"`
}

// RiskArea is a minimal risk zone representation.
type RiskArea struct {
	ID           string        `json:"id"`
	RiskLevel    string        `json:"riskLevel"`
	Polygon      []Coordinates `json:"polygon,omitempty"`
	Center       Coordinates   `json:"center,omitzero"`
	RadiusMeters float64       `json:"radiusMeters,omitempty"`
}

// RiskAreaListResponse is the payload returned by the risk-service.
type RiskAreaListResponse struct {
	Areas       []RiskArea `json:"areas"`
	GeneratedAt time.Time  `json:"generatedAt"`
}

// RoutePlanRequest is the payload to plan an evacuation route.
type RoutePlanRequest struct {
	Origin              Coordinates  `json:"origin"`
	Destination         *Coordinates `json:"destination,omitempty"`
	WaypointType        string       `json:"waypointType"`
	AvoidRiskLevels     []string     `json:"avoidRiskLevels,omitempty"`
	ClosureBufferMeters float64      `json:"closureBufferMeters,omitempty"`
}

// RoutePlanResponse is the payload returned for a planned route.
type RoutePlanResponse struct {
	Route                    []Coordinates  `json:"route"`
	Segments                 []RouteSegment `json:"segments"`
	DistanceMeters           int            `json:"distanceMeters"`
	EstimatedDurationMinutes int            `json:"estimatedDurationMinutes"`
	TargetShelter            *Shelter       `json:"targetShelter,omitempty"`
	AvoidedClosures          []string       `json:"avoidedClosures"`
	AvoidedRiskZones         []string       `json:"avoidedRiskZones"`
	Disclaimer               string         `json:"disclaimer"`
	GeneratedAt              time.Time      `json:"generatedAt"`
	DecisionSupport          bool           `json:"decisionSupport"`
}

// OptionsResponse returns supported enum values.
type OptionsResponse struct {
	WaypointTypes []string  `json:"waypointTypes"`
	GeneratedAt   time.Time `json:"generatedAt"`
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
