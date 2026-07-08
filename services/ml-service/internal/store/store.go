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

// Store is the persistence interface for prediction data.
type Store interface {
	Predict(req models.PredictionRequest, now time.Time) (models.PredictionResponse, error)
	ListLogs() []models.PredictionLog
	ModelVersion() string
}

// MemoryStore is an in-memory implementation of Store seeded from fixture files.
type MemoryStore struct {
	model       models.ModelArtifact
	predictions []models.StoredPrediction
	logs        []models.PredictionLog
	mu          sync.Mutex
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
		ID:                     fmt.Sprintf("ml_log_%s_%s", now.Format("20060102150405"), utils.SanitizeID(prediction.CellID)),
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
