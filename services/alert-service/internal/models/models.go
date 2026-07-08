package models

import "time"

// AuthorityAlert is an alert issued by an authority.
type AuthorityAlert struct {
	ID                 string                 `json:"id"`
	Title              string                 `json:"title"`
	HazardType         string                 `json:"hazardType"`
	Severity           string                 `json:"severity"`
	Message            string                 `json:"message"`
	Target             AlertTarget            `json:"target"`
	StartsAt           time.Time              `json:"startsAt"`
	ExpiresAt          time.Time              `json:"expiresAt"`
	RecommendedAction  string                 `json:"recommendedAction"`
	EvacuationRequired bool                   `json:"evacuationRequired"`
	ShelterIDs         []string               `json:"shelterIds"`
	IssuingAgencyID    string                 `json:"issuingAgencyId"`
	IssuedBy           string                 `json:"issuedBy"`
	ApprovedBy         string                 `json:"approvedBy,omitempty"`
	RejectedBy         string                 `json:"rejectedBy,omitempty"`
	Status             string                 `json:"status"`
	EmergencyOverride  bool                   `json:"emergencyOverride"`
	StatusReason       string                 `json:"statusReason,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	SubmittedAt        *time.Time             `json:"submittedAt,omitempty"`
	ApprovedAt         *time.Time             `json:"approvedAt,omitempty"`
	RejectedAt         *time.Time             `json:"rejectedAt,omitempty"`
	SourcePrediction   *AlertSourcePrediction `json:"sourcePrediction,omitempty"`
}

// AlertTarget describes the geographic or administrative audience of an alert.
type AlertTarget struct {
	Type                string          `json:"type"`
	IDs                 []string        `json:"ids"`
	Label               string          `json:"label"`
	Center              *Coordinates    `json:"center,omitempty"`
	RadiusMeters        float64         `json:"radiusMeters,omitempty"`
	Geometry            *TargetGeometry `json:"geometry,omitempty"`
	AreaSqKm            float64         `json:"areaSqKm,omitempty"`
	EstimatedPopulation int             `json:"estimatedPopulation,omitempty"`
}

// Coordinates is a lat/lng pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// TargetGeometry is a GeoJSON-like polygon geometry.
type TargetGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// CreateAlertRequest is the payload used to create or update an alert.
type CreateAlertRequest struct {
	Title              string                 `json:"title"`
	HazardType         string                 `json:"hazardType"`
	Severity           string                 `json:"severity"`
	Message            string                 `json:"message"`
	Target             AlertTarget            `json:"target"`
	StartsAt           time.Time              `json:"startsAt"`
	ExpiresAt          time.Time              `json:"expiresAt"`
	RecommendedAction  string                 `json:"recommendedAction"`
	EvacuationRequired bool                   `json:"evacuationRequired"`
	ShelterIDs         []string               `json:"shelterIds"`
	SourcePrediction   *AlertSourcePrediction `json:"sourcePrediction,omitempty"`
}

// AlertSourcePrediction links an alert to an ML prediction.
type AlertSourcePrediction struct {
	PredictionID           string  `json:"predictionId"`
	PredictionLogID        string  `json:"predictionLogId,omitempty"`
	ModelVersion           string  `json:"modelVersion"`
	InputFeatureSetVersion string  `json:"inputFeatureSetVersion"`
	Probability            float64 `json:"probability"`
	Severity               string  `json:"severity"`
	Confidence             string  `json:"confidence"`
	HumanReviewRequired    bool    `json:"humanReviewRequired"`
	AutoPublishAllowed     bool    `json:"autoPublishAllowed"`
	ReviewNote             string  `json:"reviewNote,omitempty"`
}

// WorkflowRequest is the payload for alert workflow transitions.
type WorkflowRequest struct {
	Note   string `json:"note"`
	Reason string `json:"reason"`
}

// AlertListResponse is the payload returned when listing alerts.
type AlertListResponse struct {
	Alerts []AuthorityAlert `json:"alerts"`
}

// TargetPreviewResponse is the payload returned by the target preview endpoint.
type TargetPreviewResponse struct {
	Target   AlertTarget `json:"target"`
	Summary  string      `json:"summary"`
	Warnings []string    `json:"warnings"`
}

// AuditListResponse is the payload returned when listing audit logs.
type AuditListResponse struct {
	Logs []AuditEvent `json:"logs"`
}

// AuditEvent records an action taken by an authority.
type AuditEvent struct {
	ID            string         `json:"id"`
	ActorUserID   string         `json:"actorUserId"`
	ActorAgencyID string         `json:"actorAgencyId"`
	ActorRole     string         `json:"actorRole"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId"`
	RequestID     string         `json:"requestId,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

// AuthorityContext captures the authenticated authority actor from request headers.
type AuthorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
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

// TargetCatalogRecord holds known target metadata for regions, districts, and communities.
type TargetCatalogRecord struct {
	ID                  string
	Type                string
	Label               string
	Center              Coordinates
	RadiusMeters        float64
	AreaSqKm            float64
	EstimatedPopulation int
}
