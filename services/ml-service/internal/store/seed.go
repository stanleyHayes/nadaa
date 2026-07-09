package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// FeatureRow is a single flood-risk feature cell used as a simulation grid input.
type FeatureRow struct {
	CellID    string
	Region    string
	District  string
	Community string
	Geometry  models.PolygonGeometry
	Values    map[string]any
}

// loadPredictionStore seeds a MemoryStore from model and fixture artifacts.
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

	features, err := loadFeatureRows(modelDir)
	if err != nil {
		return nil, fmt.Errorf("load feature rows: %w", err)
	}

	return &MemoryStore{
		model:       model,
		predictions: artifact.Predictions,
		features:    features,
	}, nil
}

// loadFeatureRows reads generated feature rows from the flood-risk feature pipeline.
func loadFeatureRows(modelDir string) ([]FeatureRow, error) {
	featurePath := filepath.Join(filepath.Dir(modelDir), "generated", "features.v1.json")

	type featureFile struct {
		FeatureSetVersion string                   `json:"featureSetVersion"`
		Limitations       []string                 `json:"limitations"`
		Rows              []map[string]json.RawMessage `json:"rows"`
	}

	var file featureFile
	if err := utils.ReadJSONFile(featurePath, &file); err != nil {
		return nil, err
	}

	rows := make([]FeatureRow, 0, len(file.Rows))
	for _, raw := range file.Rows {
		row, err := parseFeatureRow(raw)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseFeatureRow(raw map[string]json.RawMessage) (FeatureRow, error) {
	var row FeatureRow
	row.Values = make(map[string]any, len(raw))

	for key, value := range raw {
		switch key {
		case "cell_id":
			_ = json.Unmarshal(value, &row.CellID)
		case "region":
			_ = json.Unmarshal(value, &row.Region)
		case "district":
			_ = json.Unmarshal(value, &row.District)
		case "community":
			_ = json.Unmarshal(value, &row.Community)
		case "geometry":
			_ = json.Unmarshal(value, &row.Geometry)
		default:
			var scalar any
			_ = json.Unmarshal(value, &scalar)
			row.Values[key] = scalar
		}
	}

	if row.CellID == "" {
		return FeatureRow{}, errors.New("feature row missing cell_id")
	}
	return row, nil
}

// featureValue returns the numeric value for a feature key, defaulting to 0.
func featureValue(row FeatureRow, key string) float64 {
	raw, ok := row.Values[key]
	if !ok {
		return 0
	}
	switch v := raw.(type) {
	case float64:
		return v
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}
