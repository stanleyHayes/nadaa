package store

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

func loadPredictionStore(modelDir string) (*MemoryStore, error) {
	modelPath := filepath.Join(modelDir, "baseline-logistic.v1.json")
	predictionPath := filepath.Join(modelDir, "sample-predictions.v1.json")

	var model models.ModelArtifact
	if err := utils.ReadJSONFile(modelPath, &model); err != nil {
		return nil, fmt.Errorf("load model artifact: %w", err)
	}

	var artifact models.PredictionArtifact
	if err := utils.ReadJSONFile(predictionPath, &artifact); err != nil {
		return nil, fmt.Errorf("load prediction artifact: %w", err)
	}

	if model.ModelVersion == "" || model.ModelVersion != artifact.ModelVersion {
		return nil, errors.New("model and prediction artifacts must share a modelVersion")
	}
	if artifact.FeatureSetVersion == "" || artifact.PredictionCount != len(artifact.Predictions) {
		return nil, errors.New("prediction artifact has invalid feature set or row count")
	}

	return &MemoryStore{model: model, predictions: artifact.Predictions}, nil
}
