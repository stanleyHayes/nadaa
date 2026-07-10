package models

import "time"

// IntegrationContract describes a partner integration agreement.
type IntegrationContract struct {
	ID                     string            `json:"id"`
	Partner                string            `json:"partner"`
	PartnerType            string            `json:"partnerType"`
	Domain                 string            `json:"domain"`
	Direction              string            `json:"direction"`
	DataOwner              string            `json:"dataOwner"`
	Cadence                string            `json:"cadence"`
	Authentication         Authentication    `json:"authentication"`
	Payloads               []PayloadContract `json:"payloads"`
	FailureBehavior        FailureBehavior   `json:"failureBehavior"`
	SourceOfTruth          string            `json:"sourceOfTruth"`
	FreshnessWindowMinutes int               `json:"freshnessWindowMinutes"`
	ContactPoint           string            `json:"contactPoint"`
	Status                 string            `json:"status"`
	Notes                  string            `json:"notes"`
	UpdatedAt              time.Time         `json:"updatedAt"`
}

// Authentication describes how a partner authenticates with NADAA.
type Authentication struct {
	Mode            string   `json:"mode"`
	RequiredHeaders []string `json:"requiredHeaders,omitempty"`
	SecretScope     string   `json:"secretScope,omitempty"`
}

// PayloadContract describes an exchanged payload shape.
type PayloadContract struct {
	Name           string   `json:"name"`
	ContentType    string   `json:"contentType"`
	RequiredFields []string `json:"requiredFields"`
	OptionalFields []string `json:"optionalFields,omitempty"`
	PII            string   `json:"pii"`
	Geometry       string   `json:"geometry,omitempty"`
	ExampleRef     string   `json:"exampleRef"`
}

// FailureBehavior describes retry and fallback behavior.
type FailureBehavior struct {
	Retryable       bool   `json:"retryable"`
	MaxAttempts     int    `json:"maxAttempts"`
	BackoffSeconds  []int  `json:"backoffSeconds"`
	DeadLetterQueue string `json:"deadLetterQueue"`
	ManualFallback  string `json:"manualFallback"`
}

// WeatherHydrologyObservation is a mock weather or hydrology reading.
type WeatherHydrologyObservation struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	Metric      string      `json:"metric"`
	Value       float64     `json:"value"`
	Unit        string      `json:"unit"`
	StationID   string      `json:"stationId"`
	Location    Coordinates `json:"location"`
	ObservedAt  time.Time   `json:"observedAt"`
	ValidFrom   time.Time   `json:"validFrom"`
	ValidTo     time.Time   `json:"validTo"`
	Quality     string      `json:"quality"`
	GeneratedBy string      `json:"generatedBy"`
}

// ImportedWeatherHydrologyObservation is a stored observation after import.
type ImportedWeatherHydrologyObservation struct {
	ID            string            `json:"id"`
	Source        string            `json:"source"`
	Metric        string            `json:"metric"`
	Value         float64           `json:"value"`
	Unit          string            `json:"unit"`
	StationID     string            `json:"stationId"`
	Location      Coordinates       `json:"location"`
	ObservedAt    time.Time         `json:"observedAt"`
	ValidFrom     time.Time         `json:"validFrom"`
	ValidTo       time.Time         `json:"validTo"`
	RainfallMM    *float64          `json:"rainfallMm,omitempty"`
	WaterLevelM   *float64          `json:"waterLevelM,omitempty"`
	Metadata      map[string]string `json:"metadata"`
	ImportJobID   string            `json:"importJobId"`
	ImportedAt    time.Time         `json:"importedAt"`
	SourceRecord  string            `json:"sourceRecord"`
	StorageTarget string            `json:"storageTarget"`
}

// ObservationImportRequest starts or retries an observation import.
type ObservationImportRequest struct {
	AdapterID        string `json:"adapterId"`
	Metric           string `json:"metric,omitempty"`
	SimulateFailure  bool   `json:"simulateFailure,omitempty"`
	FailureMessage   string `json:"failureMessage,omitempty"`
	RequestedBy      string `json:"requestedBy,omitempty"`
	CorrelationID    string `json:"correlationId,omitempty"`
	ForceManualRetry bool   `json:"forceManualRetry,omitempty"`
}

// ObservationImportJob tracks an observation import run.
type ObservationImportJob struct {
	ID            string     `json:"id"`
	AdapterID     string     `json:"adapterId"`
	Source        string     `json:"source"`
	Metric        string     `json:"metric,omitempty"`
	Status        string     `json:"status"`
	Trigger       string     `json:"trigger"`
	Attempts      int        `json:"attempts"`
	Retryable     bool       `json:"retryable"`
	StartedAt     time.Time  `json:"startedAt"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
	NextRetryAt   *time.Time `json:"nextRetryAt,omitempty"`
	ImportedCount int        `json:"importedCount"`
	FailedCount   int        `json:"failedCount"`
	Error         string     `json:"error,omitempty"`
	Message       string     `json:"message"`
	RequestedBy   string     `json:"requestedBy,omitempty"`
	CorrelationID string     `json:"correlationId,omitempty"`
}

// Coordinates is a WGS84 point.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// SyncRequest requests an outbound incident or alert sync event.
type SyncRequest struct {
	Type            string      `json:"type"`
	SourceID        string      `json:"sourceId"`
	Reference       string      `json:"reference"`
	HazardType      string      `json:"hazardType"`
	Status          string      `json:"status"`
	Severity        string      `json:"severity"`
	Title           string      `json:"title,omitempty"`
	Summary         string      `json:"summary,omitempty"`
	Message         string      `json:"message,omitempty"`
	Location        Coordinates `json:"location,omitzero"`
	TargetLabel     string      `json:"targetLabel,omitempty"`
	TargetAgencyIDs []string    `json:"targetAgencyIds"`
	CorrelationID   string      `json:"correlationId"`
	OccurredAt      time.Time   `json:"occurredAt"`
}

// SyncEvent is an accepted sync event.
type SyncEvent struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	SourceID        string    `json:"sourceId"`
	Reference       string    `json:"reference"`
	TargetAgencyIDs []string  `json:"targetAgencyIds"`
	CorrelationID   string    `json:"correlationId"`
	Status          string    `json:"status"`
	AdapterID       string    `json:"adapterId"`
	QueuedAt        time.Time `json:"queuedAt"`
	Retryable       bool      `json:"retryable"`
}

// RoadClosureImportRecord is a persisted road closure import.
type RoadClosureImportRecord struct {
	ID         string     `json:"id"`
	Source     string     `json:"source"`
	SourceRef  string     `json:"sourceRef,omitempty"`
	RoadName   string     `json:"roadName"`
	Status     string     `json:"status"`
	Reason     string     `json:"reason,omitempty"`
	Geometry   string     `json:"geometry"`
	ValidFrom  time.Time  `json:"validFrom"`
	ValidTo    *time.Time `json:"validTo,omitempty"`
	Detour     string     `json:"detour,omitempty"`
	ImportedAt time.Time  `json:"importedAt"`
}

// RoadClosureImportRequest requests a road closure import.
type RoadClosureImportRequest struct {
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

// RoadClosureImportResponse is the response to a successful import.
type RoadClosureImportResponse struct {
	Imported   int                     `json:"imported"`
	Record     RoadClosureImportRecord `json:"record"`
	AcceptedAt time.Time               `json:"acceptedAt"`
}

// ContractListResponse lists integration contracts.
type ContractListResponse struct {
	Contracts []IntegrationContract `json:"contracts"`
}

// ObservationListResponse lists mock observations.
type ObservationListResponse struct {
	Observations []WeatherHydrologyObservation `json:"observations"`
}

// ImportedObservationListResponse lists imported observations.
type ImportedObservationListResponse struct {
	Observations []ImportedWeatherHydrologyObservation `json:"observations"`
}

// ObservationImportJobListResponse lists import jobs.
type ObservationImportJobListResponse struct {
	Jobs []ObservationImportJob `json:"jobs"`
}

// SyncEventListResponse lists sync events.
type SyncEventListResponse struct {
	Events []SyncEvent `json:"events"`
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
