package store

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// Store is the persistence interface for prediction and simulation data.
type Store interface {
	CVStore
	Predict(req models.PredictionRequest, now time.Time) (models.PredictionResponse, error)
	ListLogs() []models.PredictionLog
	ModelVersion() string
	CreateSimulationJob(req models.CreateSimulationRequest, now time.Time) (models.SimulationRun, error)
	GetSimulationJob(id string) (models.SimulationRun, bool)
	ListSimulationJobs() []models.SimulationRun
	ListForecasts(region string, now time.Time) []models.DemandForecast
	StagingSuggestions(agencyType string, now time.Time) []models.StagingSuggestion
	CompareScenarios(req models.CompareScenarioRequest, now time.Time) []models.ScenarioResult
}

const (
	// maxSimulationDurationHours bounds a simulation window, consistent with
	// the forecast compare endpoint's 1..168 hour window.
	maxSimulationDurationHours = 168
	// maxSimulationSteps caps the computed frame count so a tiny time step
	// cannot exhaust memory.
	maxSimulationSteps = 500
)

// MemoryStore is an in-memory implementation of Store seeded from fixture files.
type MemoryStore struct {
	model             models.ModelArtifact
	predictions       []models.StoredPrediction
	logs              []models.PredictionLog
	logCounter        int
	features          []FeatureRow
	simulations       []models.SimulationRun
	simulationCounter int
	simulationMu      sync.Mutex
	cvCache           *cvResultCache
	mu                sync.Mutex
}

// NewMemoryStore creates an in-memory store seeded from model artifacts in modelDir.
func NewMemoryStore(modelDir string) (Store, error) {
	store, err := loadPredictionStore(modelDir)
	if err != nil {
		return nil, err
	}
	return store, nil
}

// ModelVersion returns the loaded model version.
func (m *MemoryStore) ModelVersion() string {
	return m.model.ModelVersion
}

// Predict returns the nearest stored prediction to the requested location and
// records an audit log entry.
func (m *MemoryStore) Predict(req models.PredictionRequest, now time.Time) (models.PredictionResponse, error) {
	prediction, distanceMeters, err := m.nearestPrediction(req.Location)
	if err != nil {
		return models.PredictionResponse{}, err
	}

	summary := models.PredictionSummary{
		ID:                     prediction.ID,
		ModelVersion:           prediction.ModelVersion,
		HazardType:             prediction.HazardType,
		PredictionTime:         prediction.PredictionTime,
		TargetTime:             prediction.TargetTime,
		CellID:                 prediction.CellID,
		Region:                 prediction.Region,
		District:               prediction.District,
		Community:              prediction.Community,
		Location:               req.Location,
		Geometry:               prediction.Geometry,
		DistanceMeters:         distanceMeters,
		Probability:            prediction.Probability,
		Severity:               prediction.Severity,
		ExpectedOnset:          prediction.ExpectedOnset,
		Confidence:             prediction.Confidence,
		ExplanationFactors:     prediction.ExplanationFactors,
		InputFeatureSetVersion: prediction.InputFeatureSetVersion,
		HumanReviewRequired:    true,
		AutoPublishAllowed:     false,
		Source:                 "baseline_fixture_model",
	}

	logEntry := models.PredictionLog{
		PredictionID:           prediction.ID,
		ModelVersion:           prediction.ModelVersion,
		InputFeatureSetVersion: prediction.InputFeatureSetVersion,
		RequestedBy:            strings.TrimSpace(req.RequestedBy),
		CorrelationID:          strings.TrimSpace(req.CorrelationID),
		Location:               req.Location,
		StorageTarget:          "ml_predictions",
		HumanReviewRequired:    true,
		AutoPublishAllowed:     false,
		CreatedAt:              now.Format(time.RFC3339),
	}

	m.mu.Lock()
	m.logCounter++
	logEntry.ID = fmt.Sprintf("ml_log_%s_%s_%06d", now.Format("20060102150405"), utils.SanitizeID(prediction.CellID), m.logCounter)
	m.logs = append(m.logs, logEntry)
	m.mu.Unlock()

	return models.PredictionResponse{
		Prediction: summary,
		Log:        logEntry,
		Safety: models.SafetyPolicy{
			HumanReviewRequired: true,
			AutoPublishAllowed:  false,
			Message:             "Model output is decision support only and cannot publish alerts without authority review and approval.",
		},
	}, nil
}

func (m *MemoryStore) nearestPrediction(location models.Coordinates) (models.StoredPrediction, int, error) {
	if len(m.predictions) == 0 {
		return models.StoredPrediction{}, 0, errors.New("no predictions are loaded")
	}

	type candidate struct {
		prediction models.StoredPrediction
		distance   float64
	}
	candidates := make([]candidate, 0, len(m.predictions))
	for _, prediction := range m.predictions {
		centroid, ok := prediction.Geometry.Centroid()
		if !ok {
			continue
		}
		candidates = append(candidates, candidate{
			prediction: prediction,
			distance:   utils.HaversineMeters(location, centroid),
		})
	}
	if len(candidates) == 0 {
		return models.StoredPrediction{}, 0, errors.New("loaded predictions do not include usable geometry")
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})
	return candidates[0].prediction, int(math.Round(candidates[0].distance)), nil
}

// ListLogs returns a copy of the stored prediction logs sorted newest first.
func (m *MemoryStore) ListLogs() []models.PredictionLog {
	m.mu.Lock()
	defer m.mu.Unlock()

	logs := append([]models.PredictionLog(nil), m.logs...)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt > logs[j].CreatedAt
	})
	return logs
}

// CreateSimulationJob runs a deterministic flood simulation and stores the result.
func (m *MemoryStore) CreateSimulationJob(req models.CreateSimulationRequest, now time.Time) (models.SimulationRun, error) {
	if strings.TrimSpace(req.Name) == "" {
		return models.SimulationRun{}, errors.New("simulation name is required")
	}
	if req.DurationHours <= 0 {
		req.DurationHours = 6
	}
	if req.DurationHours > maxSimulationDurationHours {
		return models.SimulationRun{}, fmt.Errorf("durationHours must be between 1 and %d", maxSimulationDurationHours)
	}
	if req.TimeStepHours <= 0 {
		req.TimeStepHours = 1
	}
	if req.TimeStepHours > req.DurationHours {
		req.TimeStepHours = req.DurationHours
	}
	if steps := req.DurationHours / req.TimeStepHours; steps > maxSimulationSteps {
		return models.SimulationRun{}, fmt.Errorf("durationHours and timeStepHours must yield at most %d steps", maxSimulationSteps)
	}

	scenario := models.SimulationScenario{
		DurationHours: req.DurationHours,
		TimeStepHours: req.TimeStepHours,
	}
	if req.RainfallMmOverride != 0 {
		scenario.RainfallMmOverride = &req.RainfallMmOverride
	}
	if req.WaterLevelTrendCmOverride != 0 {
		scenario.WaterLevelTrendCmOverride = &req.WaterLevelTrendCmOverride
	}

	run := models.SimulationRun{
		Name:              strings.TrimSpace(req.Name),
		Status:            "running",
		Scenario:          scenario,
		Assumptions:       defaultSimulationAssumptions(),
		Limitations:       append([]string(nil), m.model.Limitations...),
		ModelVersion:      m.model.ModelVersion,
		FeatureSetVersion: m.model.TrainingFeatureSetVersion,
		CreatedAt:         now.Format(time.RFC3339),
		UpdatedAt:         now.Format(time.RFC3339),
		Safety: models.SafetyPolicy{
			HumanReviewRequired: true,
			AutoPublishAllowed:  false,
			Message:             "Simulation output is decision support only and cannot publish alerts without authority review and approval.",
		},
	}

	// Mint the ID and reference and append under the lock so same-second jobs
	// never collide and the completion update below cannot clobber another run.
	m.simulationMu.Lock()
	m.simulationCounter++
	run.ID = fmt.Sprintf("sim_%s_%06d", now.Format("20060102150405"), m.simulationCounter)
	run.Reference = fmt.Sprintf("FS-%s-%05d", now.Format("2006"), len(m.simulations)+1)
	m.simulations = append(m.simulations, run)
	m.simulationMu.Unlock()

	frames := m.runSimulation(scenario, now)
	run.Frames = frames
	run.Status = "completed"
	run.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	m.simulationMu.Lock()
	for i := range m.simulations {
		if m.simulations[i].ID == run.ID {
			m.simulations[i] = run
			break
		}
	}
	m.simulationMu.Unlock()

	return run, nil
}

// GetSimulationJob returns a simulation job by id.
func (m *MemoryStore) GetSimulationJob(id string) (models.SimulationRun, bool) {
	m.simulationMu.Lock()
	defer m.simulationMu.Unlock()

	for _, run := range m.simulations {
		if run.ID == id {
			return run, true
		}
	}
	return models.SimulationRun{}, false
}

// ListSimulationJobs returns simulation jobs sorted newest first.
func (m *MemoryStore) ListSimulationJobs() []models.SimulationRun {
	m.simulationMu.Lock()
	defer m.simulationMu.Unlock()

	jobs := append([]models.SimulationRun(nil), m.simulations...)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt > jobs[j].CreatedAt
	})
	return jobs
}

func (m *MemoryStore) runSimulation(scenario models.SimulationScenario, now time.Time) []models.SimulationFrame {
	steps := scenario.DurationHours / scenario.TimeStepHours
	if steps <= 0 {
		steps = 1
	}

	frames := make([]models.SimulationFrame, 0, steps)
	for step := 1; step <= steps; step++ {
		progress := float64(step) / float64(steps)
		targetTime := now.Add(time.Duration(step*scenario.TimeStepHours) * time.Hour).UTC()

		cells := make([]models.SimulationCell, 0, len(m.features))
		for _, row := range m.features {
			cell := m.scoreCell(row, scenario, progress)
			cells = append(cells, cell)
		}

		frames = append(frames, models.SimulationFrame{
			TargetTime: targetTime.Format(time.RFC3339),
			Cells:      cells,
		})
	}
	return frames
}

func (m *MemoryStore) scoreCell(row FeatureRow, scenario models.SimulationScenario, progress float64) models.SimulationCell {
	values := make(map[string]float64, len(row.Values))
	for key := range row.Values {
		values[key] = featureValue(row, key)
	}

	if scenario.RainfallMmOverride != nil {
		values["rainfall_forecast_24h_mm"] += *scenario.RainfallMmOverride * progress
	}
	if scenario.WaterLevelTrendCmOverride != nil {
		values["water_level_trend_cm"] += *scenario.WaterLevelTrendCmOverride * progress
	}

	probability := m.logisticProbability(values)
	severity := m.severityFromProbability(probability, values)
	confidence := "medium"

	cell := models.SimulationCell{
		CellID:             row.CellID,
		Region:             row.Region,
		District:           row.District,
		Community:          row.Community,
		Geometry:           row.Geometry,
		Probability:        probability,
		Severity:           severity,
		DepthBand:          depthBandForSeverity(severity),
		Confidence:         confidence,
		ExplanationFactors: m.explanationFactors(values),
	}
	return cell
}

func (m *MemoryStore) logisticProbability(values map[string]float64) float64 {
	intercept := m.model.Coefficients["intercept"]
	z := intercept
	for _, feature := range m.model.FeatureColumns {
		coef, ok := m.model.Coefficients[feature]
		if !ok {
			continue
		}
		std, ok := m.model.Preprocessing.NumericStandardization[feature]
		if !ok {
			continue
		}
		raw := values[feature]
		standardized := (raw - std.Mean) / std.Std
		z += coef * standardized
	}
	return 1.0 / (1.0 + math.Exp(-z))
}

func (m *MemoryStore) severityFromProbability(probability float64, values map[string]float64) string {
	thresholds := m.model.Hyperparameters.SeverityThresholds
	if thresholds == nil {
		thresholds = map[string]float64{"severe": 0.78, "high": 0.58, "moderate": 0.35}
	}
	insideZone := values["inside_known_flood_zone"] >= 0.5
	rainfall := values["rainfall_forecast_24h_mm"]

	if probability >= thresholds["severe"] && insideZone && rainfall >= 55 {
		return "severe"
	}
	if probability >= thresholds["high"] {
		return "high"
	}
	if probability >= thresholds["moderate"] {
		return "moderate"
	}
	return "low"
}

func (m *MemoryStore) explanationFactors(values map[string]float64) []models.ExplanationFactor {
	factors := make([]models.ExplanationFactor, 0, len(m.model.FeatureColumns))
	for _, feature := range m.model.FeatureColumns {
		coef, ok := m.model.Coefficients[feature]
		if !ok {
			continue
		}
		std, ok := m.model.Preprocessing.NumericStandardization[feature]
		if !ok {
			continue
		}
		raw := values[feature]
		standardized := (raw - std.Mean) / std.Std
		contribution := coef * standardized
		direction := "reduces_risk"
		if contribution > 0 {
			direction = "increases_risk"
		}
		factors = append(factors, models.ExplanationFactor{
			Feature:      feature,
			Label:        feature,
			Value:        raw,
			Contribution: contribution,
			Direction:    direction,
		})
	}
	sort.Slice(factors, func(i, j int) bool {
		return math.Abs(factors[j].Contribution) < math.Abs(factors[i].Contribution)
	})
	if len(factors) > 5 {
		factors = factors[:5]
	}
	return factors
}

func defaultSimulationAssumptions() []string {
	return []string{
		"Simulation applies user rainfall and water-level overrides linearly across the requested time window.",
		"Depth bands are estimated from severity class, not from a hydrodynamic model.",
		"Only the 24-hour rainfall forecast and water-level trend features are perturbed; terrain, drainage, and exposure remain static.",
		"Cells use simplified bounding-box geometries from the NADAA-070 feature pipeline.",
	}
}

func depthBandForSeverity(severity string) string {
	switch severity {
	case "severe":
		return "> 1.2 m"
	case "high":
		return "0.6 - 1.2 m"
	case "moderate":
		return "0.3 - 0.6 m"
	default:
		return "< 0.3 m"
	}
}
