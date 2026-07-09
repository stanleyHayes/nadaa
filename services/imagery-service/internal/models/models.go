// Package models defines the domain types for the imagery-service.
package models

import (
	"encoding/json"
	"time"
)

// ImageryRecord represents a stored drone or satellite imagery metadata record.
type ImageryRecord struct {
	ID                string          `json:"id"`
	Reference         string          `json:"reference"`
	Source            string          `json:"source"`
	CaptureTime       time.Time       `json:"captureTime"`
	Geometry          json.RawMessage `json:"geometry"`
	CoverageAreaKm2   float64         `json:"coverageAreaKm2"`
	ResolutionMeters  float64         `json:"resolutionMeters"`
	License           string          `json:"license,omitempty"`
	RelatedIncidentID string          `json:"relatedIncidentId,omitempty"`
	RelatedRiskZoneID string          `json:"relatedRiskZoneId,omitempty"`
	MlWorkflowID      string          `json:"mlWorkflowId,omitempty"`
	FileName          string          `json:"fileName"`
	ContentType       string          `json:"contentType"`
	SizeBytes         int64           `json:"sizeBytes"`
	StoragePath       string          `json:"storagePath"`
	Status            string          `json:"status"`
	UploadedBy        string          `json:"uploadedBy"`
	CreatedAt         time.Time       `json:"createdAt"`
	ExpiresAt         time.Time       `json:"expiresAt"`
}

// ImageryUploadInput is the parsed and validated input for imagery creation.
type ImageryUploadInput struct {
	Source            string
	CaptureTime       time.Time
	Geometry          json.RawMessage
	CoverageAreaKm2   float64
	ResolutionMeters  float64
	License           string
	RelatedIncidentID string
	RelatedRiskZoneID string
	MlWorkflowID      string
}

// ImageryListFilter is the set of filters for listing imagery records.
type ImageryListFilter struct {
	Source            string
	Status            string
	RelatedIncidentID string
	RelatedRiskZoneID string
	Query             string
}

// ImageryListResponse is the response payload for listing imagery records.
type ImageryListResponse struct {
	Imagery     []ImageryRecord `json:"imagery"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

// ImageryLifecycleResponse reports the result of a lifecycle run.
type ImageryLifecycleResponse struct {
	ExpiredCount int `json:"expiredCount"`
}

// GeoJSONFeatureCollection is a minimal GeoJSON FeatureCollection wrapper.
type GeoJSONFeatureCollection struct {
	Type     string          `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

// GeoJSONFeature is a minimal GeoJSON Feature wrapper.
type GeoJSONFeature struct {
	Type       string          `json:"type"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties map[string]any  `json:"properties"`
}

// AuthorityContext holds authenticated authority metadata from request headers.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}
