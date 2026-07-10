package models

// ForecastFactor describes one contributor to a demand forecast.
type ForecastFactor struct {
	Name      string  `json:"name"`
	Label     string  `json:"label"`
	Value     float64 `json:"value"`
	Weight    float64 `json:"weight"`
	Direction string  `json:"direction"`
}

// DemandForecast is a per-district predicted emergency demand estimate.
type DemandForecast struct {
	ID                     string           `json:"id"`
	Region                 string           `json:"region"`
	District               string           `json:"district"`
	TimeWindowStart        string           `json:"timeWindowStart"`
	TimeWindowEnd          string           `json:"timeWindowEnd"`
	PredictedIncidentCount int              `json:"predictedIncidentCount"`
	HazardType             string           `json:"hazardType"`
	Confidence             string           `json:"confidence"`
	ConfidenceScore        float64          `json:"confidenceScore"`
	Factors                []ForecastFactor `json:"factors"`
	RiskLevel              string           `json:"riskLevel"`
	GeneratedAt            string           `json:"generatedAt"`
}

// ForecastListResponse is the payload for listing demand forecasts.
type ForecastListResponse struct {
	Forecasts   []DemandForecast `json:"forecasts"`
	GeneratedAt string           `json:"generatedAt"`
}

// ForecastDetailResponse returns the top forecast for a region plus all of them.
type ForecastDetailResponse struct {
	Forecast    DemandForecast   `json:"forecast"`
	Forecasts   []DemandForecast `json:"forecasts"`
	GeneratedAt string           `json:"generatedAt"`
}

// StagingSuggestion recommends a staging position for a responder agency.
type StagingSuggestion struct {
	ID                     string      `json:"id"`
	Location               Coordinates `json:"location"`
	LocationLabel          string      `json:"locationLabel"`
	AgencyType             string      `json:"agencyType"`
	Reason                 string      `json:"reason"`
	Confidence             string      `json:"confidence"`
	ConfidenceScore        float64     `json:"confidenceScore"`
	OperationalConstraints []string    `json:"operationalConstraints"`
	RecommendedUnits       int         `json:"recommendedUnits"`
	RadiusMeters           int         `json:"radiusMeters"`
	GeneratedAt            string      `json:"generatedAt"`
}

// StagingSuggestionListResponse is the payload for listing staging suggestions.
type StagingSuggestionListResponse struct {
	Suggestions []StagingSuggestion `json:"suggestions"`
	GeneratedAt string              `json:"generatedAt"`
}

// CompareScenarioRequest overrides the baseline forecasting parameters.
type CompareScenarioRequest struct {
	Region           string   `json:"region,omitempty"`
	RiskLevel        string   `json:"riskLevel,omitempty"`
	HistoricalWeight float64  `json:"historicalWeight,omitempty"`
	CapacityFactor   float64  `json:"capacityFactor,omitempty"`
	HazardTypes      []string `json:"hazardTypes,omitempty"`
	TimeWindowHours  int      `json:"timeWindowHours,omitempty"`
}

// ScenarioSummary aggregates a scenario's forecasts.
type ScenarioSummary struct {
	TotalPredictedIncidents int     `json:"totalPredictedIncidents"`
	AverageConfidenceScore  float64 `json:"averageConfidenceScore"`
	HighestRiskRegion       string  `json:"highestRiskRegion"`
	HighestRiskHazard       string  `json:"highestRiskHazard"`
}

// ScenarioResult is one named scenario within a comparison.
type ScenarioResult struct {
	Name       string                 `json:"name"`
	Parameters CompareScenarioRequest `json:"parameters"`
	Forecasts  []DemandForecast       `json:"forecasts"`
	Summary    ScenarioSummary        `json:"summary"`
}

// CompareScenarioResponse returns the baseline and adjusted scenarios.
type CompareScenarioResponse struct {
	Scenarios   []ScenarioResult `json:"scenarios"`
	GeneratedAt string           `json:"generatedAt"`
}
