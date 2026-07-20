package models

// NumericStandardization holds mean/std used to standardize a feature.
type NumericStandardization struct {
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
}

// Preprocessing describes how raw feature values are transformed before scoring.
type Preprocessing struct {
	NumericStandardization map[string]NumericStandardization `json:"numericStandardization"`
}

// Hyperparameters holds model thresholds and tuning values.
type Hyperparameters struct {
	SeverityThresholds map[string]float64 `json:"severityThresholds"`
}

// ModelArtifact describes a trained ML model used to produce predictions.
type ModelArtifact struct {
	ModelVersion              string             `json:"modelVersion"`
	ModelFamily               string             `json:"modelFamily"`
	HazardType                string             `json:"hazardType"`
	TrainingFeatureSetVersion string             `json:"trainingFeatureSetVersion"`
	FeatureColumns            []string           `json:"featureColumns"`
	Limitations               []string           `json:"limitations"`
	Preprocessing             Preprocessing      `json:"preprocessing"`
	Coefficients              map[string]float64 `json:"coefficients"`
	Hyperparameters           Hyperparameters    `json:"hyperparameters"`
}

// PredictionArtifact is the top-level container for a batch of predictions.
type PredictionArtifact struct {
	ModelVersion      string             `json:"modelVersion"`
	FeatureSetVersion string             `json:"featureSetVersion"`
	GeneratedAt       string             `json:"generatedAt"`
	TargetTime        string             `json:"targetTime"`
	PredictionCount   int                `json:"predictionCount"`
	Predictions       []StoredPrediction `json:"predictions"`
}

// StoredPrediction is a single prediction loaded from a fixture artifact.
type StoredPrediction struct {
	ID                     string              `json:"id"`
	ModelVersion           string              `json:"modelVersion"`
	HazardType             string              `json:"hazardType"`
	PredictionTime         string              `json:"predictionTime"`
	TargetTime             string              `json:"targetTime"`
	CellID                 string              `json:"cellId"`
	Region                 string              `json:"region"`
	District               string              `json:"district"`
	Community              string              `json:"community"`
	Geometry               PolygonGeometry     `json:"geometry"`
	Probability            float64             `json:"probability"`
	Severity               string              `json:"severity"`
	ExpectedOnset          string              `json:"expectedOnset"`
	Confidence             string              `json:"confidence"`
	ExplanationFactors     []ExplanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string              `json:"inputFeatureSetVersion"`
	SourceFeatureRow       string              `json:"sourceFeatureRow"`
}

// PredictionResponse is returned when a prediction is requested.
type PredictionResponse struct {
	Prediction PredictionSummary `json:"prediction"`
	Log        PredictionLog     `json:"log"`
	Safety     SafetyPolicy      `json:"safety"`
}

// PredictionSummary is the human-friendly prediction returned to callers.
type PredictionSummary struct {
	ID                     string              `json:"id"`
	ModelVersion           string              `json:"modelVersion"`
	HazardType             string              `json:"hazardType"`
	PredictionTime         string              `json:"predictionTime"`
	TargetTime             string              `json:"targetTime"`
	CellID                 string              `json:"cellId"`
	Region                 string              `json:"region"`
	District               string              `json:"district"`
	Community              string              `json:"community"`
	Location               Coordinates         `json:"location"`
	Geometry               PolygonGeometry     `json:"geometry"`
	DistanceMeters         int                 `json:"distanceMeters"`
	Probability            float64             `json:"probability"`
	Severity               string              `json:"severity"`
	ExpectedOnset          string              `json:"expectedOnset"`
	Confidence             string              `json:"confidence"`
	ExplanationFactors     []ExplanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string              `json:"inputFeatureSetVersion"`
	HumanReviewRequired    bool                `json:"humanReviewRequired"`
	AutoPublishAllowed     bool                `json:"autoPublishAllowed"`
	Source                 string              `json:"source"`
}

// ExplanationFactor describes one contributor to a prediction.
type ExplanationFactor struct {
	Feature      string  `json:"feature"`
	Label        string  `json:"label"`
	Value        any     `json:"value"`
	Contribution float64 `json:"contribution"`
	Direction    string  `json:"direction"`
}

// PredictionRequest is the body of a flood prediction request.
type PredictionRequest struct {
	Location      Coordinates `json:"location"`
	RequestedBy   string      `json:"requestedBy,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

// PredictionLog records each prediction request for audit purposes.
type PredictionLog struct {
	ID                     string      `json:"id"`
	PredictionID           string      `json:"predictionId"`
	ModelVersion           string      `json:"modelVersion"`
	InputFeatureSetVersion string      `json:"inputFeatureSetVersion"`
	RequestedBy            string      `json:"requestedBy,omitempty"`
	CorrelationID          string      `json:"correlationId,omitempty"`
	Location               Coordinates `json:"location"`
	StorageTarget          string      `json:"storageTarget"`
	HumanReviewRequired    bool        `json:"humanReviewRequired"`
	AutoPublishAllowed     bool        `json:"autoPublishAllowed"`
	CreatedAt              string      `json:"createdAt"`
}

// PredictionLogListResponse is the payload returned when listing prediction logs.
type PredictionLogListResponse struct {
	Logs   []PredictionLog `json:"logs"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// SafetyPolicy communicates the review and publication rules for model output.
type SafetyPolicy struct {
	HumanReviewRequired bool   `json:"humanReviewRequired"`
	AutoPublishAllowed  bool   `json:"autoPublishAllowed"`
	Message             string `json:"message"`
}

// PolygonGeometry is a simple GeoJSON-style polygon.
type PolygonGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// Coordinates is a lat/lng point.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
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

// Centroid returns the arithmetic center of the polygon's first ring.
func (g PolygonGeometry) Centroid() (Coordinates, bool) {
	if len(g.Coordinates) == 0 || len(g.Coordinates[0]) == 0 {
		return Coordinates{}, false
	}

	ring := g.Coordinates[0]
	if len(ring) > 1 && samePoint(ring[0], ring[len(ring)-1]) {
		ring = ring[:len(ring)-1]
	}
	if len(ring) == 0 {
		return Coordinates{}, false
	}

	var latSum float64
	var lngSum float64
	for _, point := range ring {
		if len(point) < 2 {
			return Coordinates{}, false
		}
		lngSum += point[0]
		latSum += point[1]
	}

	return Coordinates{Lat: latSum / float64(len(ring)), Lng: lngSum / float64(len(ring))}, true
}

func samePoint(a []float64, b []float64) bool {
	return len(a) >= 2 && len(b) >= 2 && a[0] == b[0] && a[1] == b[1]
}

// SimulationScenario describes the user-supplied overrides for a simulation run.
type SimulationScenario struct {
	RainfallMmOverride        *float64 `json:"rainfallMmOverride,omitempty"`
	WaterLevelTrendCmOverride *float64 `json:"waterLevelTrendCmOverride,omitempty"`
	DurationHours             int      `json:"durationHours"`
	TimeStepHours             int      `json:"timeStepHours"`
}

// CreateSimulationRequest is the body of a flood simulation creation request.
type CreateSimulationRequest struct {
	Name                      string  `json:"name"`
	RainfallMmOverride        float64 `json:"rainfallMmOverride,omitempty"`
	WaterLevelTrendCmOverride float64 `json:"waterLevelTrendCmOverride,omitempty"`
	DurationHours             int     `json:"durationHours,omitempty"`
	TimeStepHours             int     `json:"timeStepHours,omitempty"`
}

// SimulationCell is one grid cell inside a simulation frame.
type SimulationCell struct {
	CellID             string              `json:"cellId"`
	Region             string              `json:"region"`
	District           string              `json:"district"`
	Community          string              `json:"community"`
	Geometry           PolygonGeometry     `json:"geometry"`
	Probability        float64             `json:"probability"`
	Severity           string              `json:"severity"`
	DepthBand          string              `json:"depthBand"`
	Confidence         string              `json:"confidence"`
	ExplanationFactors []ExplanationFactor `json:"explanationFactors"`
}

// SimulationFrame represents the simulated state at a single target time.
type SimulationFrame struct {
	TargetTime string           `json:"targetTime"`
	Cells      []SimulationCell `json:"cells"`
}

// SimulationRun is the persisted result of a flood simulation job.
type SimulationRun struct {
	ID                string             `json:"id"`
	Reference         string             `json:"reference"`
	Name              string             `json:"name"`
	Status            string             `json:"status"`
	Scenario          SimulationScenario `json:"scenario"`
	Frames            []SimulationFrame  `json:"frames"`
	Assumptions       []string           `json:"assumptions"`
	Limitations       []string           `json:"limitations"`
	ModelVersion      string             `json:"modelVersion"`
	FeatureSetVersion string             `json:"featureSetVersion"`
	CreatedAt         string             `json:"createdAt"`
	UpdatedAt         string             `json:"updatedAt"`
	Safety            SafetyPolicy       `json:"safety"`
}

// SimulationListResponse is returned when listing simulation jobs.
type SimulationListResponse struct {
	Simulations []SimulationRun `json:"simulations"`
	Total       int             `json:"total"`
	Limit       int             `json:"limit"`
	Offset      int             `json:"offset"`
	GeneratedAt string          `json:"generatedAt"`
}

// SimulationDetailResponse is returned when fetching a single simulation job.
type SimulationDetailResponse struct {
	Simulation SimulationRun `json:"simulation"`
}
