package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/models"
)

type mlClient struct {
	baseURL    string
	httpClient *http.Client
}

func newMLClient(baseURL string, httpClient *http.Client) *mlClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &mlClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Transport:     httpClient.Transport,
			CheckRedirect: httpClient.CheckRedirect,
			Jar:           httpClient.Jar,
			Timeout:       2 * time.Second,
		},
	}
}

func (c *mlClient) predict(ctx context.Context, location models.Coordinates) (models.MLPrediction, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	requestPayload := models.MLPredictionRequest{
		Location:      location,
		RequestedBy:   "risk-service",
		CorrelationID: fmt.Sprintf("risk_%0.4f_%0.4f", location.Lat, location.Lng),
	}
	body, err := json.Marshal(requestPayload)
	if err != nil {
		return models.MLPrediction{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/ml/flood/predictions", bytes.NewReader(body))
	if err != nil {
		return models.MLPrediction{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return models.MLPrediction{}, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return models.MLPrediction{}, fmt.Errorf("ml service returned %d", response.StatusCode)
	}

	var payload models.MLPredictionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return models.MLPrediction{}, err
	}
	if payload.Prediction.ModelVersion == "" {
		return models.MLPrediction{}, errors.New("ml service returned an empty modelVersion")
	}

	return models.MLPrediction{
		ID:                     payload.Prediction.ID,
		ModelVersion:           payload.Prediction.ModelVersion,
		HazardType:             payload.Prediction.HazardType,
		PredictionTime:         payload.Prediction.PredictionTime,
		TargetTime:             payload.Prediction.TargetTime,
		CellID:                 payload.Prediction.CellID,
		Region:                 payload.Prediction.Region,
		District:               payload.Prediction.District,
		Community:              payload.Prediction.Community,
		Probability:            payload.Prediction.Probability,
		Severity:               payload.Prediction.Severity,
		ExpectedOnset:          payload.Prediction.ExpectedOnset,
		Confidence:             payload.Prediction.Confidence,
		ExplanationFactors:     payload.Prediction.ExplanationFactors,
		InputFeatureSetVersion: payload.Prediction.InputFeatureSetVersion,
		PredictionLogID:        payload.Log.ID,
		HumanReviewRequired:    payload.Prediction.HumanReviewRequired,
		AutoPublishAllowed:     payload.Prediction.AutoPublishAllowed,
		Source:                 payload.Prediction.Source,
	}, nil
}
