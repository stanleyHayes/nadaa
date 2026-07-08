package models

import (
	"math"
)

// Coordinates represents a latitude/longitude point.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// BoundingBox describes a rectangular geographic area.
type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}

// Contains reports whether the bounding box contains the given coordinates.
func (b BoundingBox) Contains(location Coordinates) bool {
	return location.Lat >= b.MinLat && location.Lat <= b.MaxLat && location.Lng >= b.MinLng && location.Lng <= b.MaxLng
}

// DistanceMeters returns the shortest distance from the coordinates to the box edge, or 0 if inside.
func (b BoundingBox) DistanceMeters(location Coordinates) float64 {
	if b.Contains(location) {
		return 0
	}

	nearest := Coordinates{
		Lat: Clamp(location.Lat, b.MinLat, b.MaxLat),
		Lng: Clamp(location.Lng, b.MinLng, b.MaxLng),
	}
	return HaversineMeters(location, nearest)
}

// ValidCoordinates reports whether the coordinates are within valid ranges.
func ValidCoordinates(location Coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

// HaversineMeters returns the great-circle distance between two coordinates in meters.
func HaversineMeters(a, b Coordinates) float64 {
	const earthRadiusMeters = 6371000.0

	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)

	sinLat := math.Sin(deltaLat / 2)
	sinLng := math.Sin(deltaLng / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

// Clamp restricts value to the inclusive range [minimum, maximum].
func Clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

// RiskZone is an internal store record for a geographic risk zone.
type RiskZone struct {
	ID          string
	HazardType  string
	RiskLevel   string
	Bounds      BoundingBox
	Probability float64
	Explanation string
}

// HistoricalReport is an internal store record for a past incident report.
type HistoricalReport struct {
	HazardType string
	Location   Coordinates
	Severity   string
}

// ShelterSummary is a nearby shelter returned in risk responses.
type ShelterSummary struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Location         Coordinates `json:"location"`
	Capacity         int         `json:"capacity,omitempty"`
	CurrentOccupancy int         `json:"currentOccupancy,omitempty"`
	Contact          string      `json:"contact,omitempty"`
	DistanceMeters   int         `json:"distanceMeters,omitempty"`
	Status           string      `json:"status,omitempty"`
	Facilities       []string    `json:"facilities,omitempty"`
}

// FacilitySummary is a nearby emergency facility returned in risk responses.
type FacilitySummary struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Location       Coordinates `json:"location"`
	Region         string      `json:"region,omitempty"`
	District       string      `json:"district,omitempty"`
	Contact        string      `json:"contact,omitempty"`
	DistanceMeters int         `json:"distanceMeters,omitempty"`
}

// RiskSummary is a single hazard risk inside a risk response.
type RiskSummary struct {
	Type        string  `json:"type"`
	Level       string  `json:"level"`
	Probability float64 `json:"probability"`
	Reason      string  `json:"reason"`
}

// RiskResponse is the payload returned by the risk endpoint.
type RiskResponse struct {
	Location           string            `json:"location"`
	OverallRisk        string            `json:"overallRisk"`
	Risks              []RiskSummary     `json:"risks"`
	MLPrediction       *MLPrediction     `json:"mlPrediction,omitempty"`
	NearestShelters    []ShelterSummary  `json:"nearestShelters"`
	NearbyFacilities   []FacilitySummary `json:"nearbyFacilities"`
	RecommendedActions []string          `json:"recommendedActions"`
}

// MLPrediction is a machine-learning prediction embedded in a risk response.
type MLPrediction struct {
	ID                     string                `json:"id"`
	ModelVersion           string                `json:"modelVersion"`
	HazardType             string                `json:"hazardType"`
	PredictionTime         string                `json:"predictionTime"`
	TargetTime             string                `json:"targetTime"`
	CellID                 string                `json:"cellId"`
	Region                 string                `json:"region"`
	District               string                `json:"district"`
	Community              string                `json:"community"`
	Probability            float64               `json:"probability"`
	Severity               string                `json:"severity"`
	ExpectedOnset          string                `json:"expectedOnset"`
	Confidence             string                `json:"confidence"`
	ExplanationFactors     []MLExplanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string                `json:"inputFeatureSetVersion"`
	PredictionLogID        string                `json:"predictionLogId"`
	HumanReviewRequired    bool                  `json:"humanReviewRequired"`
	AutoPublishAllowed     bool                  `json:"autoPublishAllowed"`
	Source                 string                `json:"source"`
}

// MLExplanationFactor explains one feature contribution in an ML prediction.
type MLExplanationFactor struct {
	Feature      string  `json:"feature"`
	Label        string  `json:"label"`
	Value        any     `json:"value"`
	Contribution float64 `json:"contribution"`
	Direction    string  `json:"direction"`
}

// MLPredictionPayload is the prediction object returned by the ML service.
type MLPredictionPayload struct {
	ID                     string                `json:"id"`
	ModelVersion           string                `json:"modelVersion"`
	HazardType             string                `json:"hazardType"`
	PredictionTime         string                `json:"predictionTime"`
	TargetTime             string                `json:"targetTime"`
	CellID                 string                `json:"cellId"`
	Region                 string                `json:"region"`
	District               string                `json:"district"`
	Community              string                `json:"community"`
	Probability            float64               `json:"probability"`
	Severity               string                `json:"severity"`
	ExpectedOnset          string                `json:"expectedOnset"`
	Confidence             string                `json:"confidence"`
	ExplanationFactors     []MLExplanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string                `json:"inputFeatureSetVersion"`
	HumanReviewRequired    bool                  `json:"humanReviewRequired"`
	AutoPublishAllowed     bool                  `json:"autoPublishAllowed"`
	Source                 string                `json:"source"`
}

// MLPredictionResponse is the top-level response returned by the ML service.
type MLPredictionResponse struct {
	Prediction MLPredictionPayload `json:"prediction"`
	Log        MLPredictionLog     `json:"log"`
}

// MLPredictionLog is the log object returned alongside an ML prediction.
type MLPredictionLog struct {
	ID                     string `json:"id"`
	ModelVersion           string `json:"modelVersion"`
	InputFeatureSetVersion string `json:"inputFeatureSetVersion"`
}

// MLPredictionRequest is the request body sent to the ML service.
type MLPredictionRequest struct {
	Location      Coordinates `json:"location"`
	RequestedBy   string      `json:"requestedBy"`
	CorrelationID string      `json:"correlationId"`
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
