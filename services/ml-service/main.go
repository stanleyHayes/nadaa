package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *predictionStore
}

type predictionStore struct {
	model       modelArtifact
	predictions []storedPrediction
	logs        []predictionLog
	mu          sync.Mutex
}

type modelArtifact struct {
	ModelVersion              string   `json:"modelVersion"`
	ModelFamily               string   `json:"modelFamily"`
	HazardType                string   `json:"hazardType"`
	TrainingFeatureSetVersion string   `json:"trainingFeatureSetVersion"`
	Limitations               []string `json:"limitations"`
}

type predictionArtifact struct {
	ModelVersion      string             `json:"modelVersion"`
	FeatureSetVersion string             `json:"featureSetVersion"`
	GeneratedAt       string             `json:"generatedAt"`
	TargetTime        string             `json:"targetTime"`
	PredictionCount   int                `json:"predictionCount"`
	Predictions       []storedPrediction `json:"predictions"`
}

type storedPrediction struct {
	ID                     string              `json:"id"`
	ModelVersion           string              `json:"modelVersion"`
	HazardType             string              `json:"hazardType"`
	PredictionTime         string              `json:"predictionTime"`
	TargetTime             string              `json:"targetTime"`
	CellID                 string              `json:"cellId"`
	Region                 string              `json:"region"`
	District               string              `json:"district"`
	Community              string              `json:"community"`
	Geometry               polygonGeometry     `json:"geometry"`
	Probability            float64             `json:"probability"`
	Severity               string              `json:"severity"`
	ExpectedOnset          string              `json:"expectedOnset"`
	Confidence             string              `json:"confidence"`
	ExplanationFactors     []explanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string              `json:"inputFeatureSetVersion"`
	SourceFeatureRow       string              `json:"sourceFeatureRow"`
}

type predictionResponse struct {
	Prediction predictionSummary `json:"prediction"`
	Log        predictionLog     `json:"log"`
	Safety     safetyPolicy      `json:"safety"`
}

type predictionSummary struct {
	ID                     string              `json:"id"`
	ModelVersion           string              `json:"modelVersion"`
	HazardType             string              `json:"hazardType"`
	PredictionTime         string              `json:"predictionTime"`
	TargetTime             string              `json:"targetTime"`
	CellID                 string              `json:"cellId"`
	Region                 string              `json:"region"`
	District               string              `json:"district"`
	Community              string              `json:"community"`
	Location               coordinates         `json:"location"`
	Geometry               polygonGeometry     `json:"geometry"`
	DistanceMeters         int                 `json:"distanceMeters"`
	Probability            float64             `json:"probability"`
	Severity               string              `json:"severity"`
	ExpectedOnset          string              `json:"expectedOnset"`
	Confidence             string              `json:"confidence"`
	ExplanationFactors     []explanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string              `json:"inputFeatureSetVersion"`
	HumanReviewRequired    bool                `json:"humanReviewRequired"`
	AutoPublishAllowed     bool                `json:"autoPublishAllowed"`
	Source                 string              `json:"source"`
}

type explanationFactor struct {
	Feature      string  `json:"feature"`
	Label        string  `json:"label"`
	Value        any     `json:"value"`
	Contribution float64 `json:"contribution"`
	Direction    string  `json:"direction"`
}

type predictionRequest struct {
	Location      coordinates `json:"location"`
	RequestedBy   string      `json:"requestedBy,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
}

type predictionLog struct {
	ID                     string      `json:"id"`
	PredictionID           string      `json:"predictionId"`
	ModelVersion           string      `json:"modelVersion"`
	InputFeatureSetVersion string      `json:"inputFeatureSetVersion"`
	RequestedBy            string      `json:"requestedBy,omitempty"`
	CorrelationID          string      `json:"correlationId,omitempty"`
	Location               coordinates `json:"location"`
	StorageTarget          string      `json:"storageTarget"`
	HumanReviewRequired    bool        `json:"humanReviewRequired"`
	AutoPublishAllowed     bool        `json:"autoPublishAllowed"`
	CreatedAt              string      `json:"createdAt"`
}

type predictionLogListResponse struct {
	Logs []predictionLog `json:"logs"`
}

type safetyPolicy struct {
	HumanReviewRequired bool   `json:"humanReviewRequired"`
	AutoPublishAllowed  bool   `json:"autoPublishAllowed"`
	Message             string `json:"message"`
}

type polygonGeometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	earthRadiusMeters = 6371000.0
	defaultBindAddr   = ":8094"
)

func main() {
	srv, err := newServer()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/ml/flood/predictions", srv.createFloodPredictionHandler)
	mux.HandleFunc("GET /api/v1/ml/prediction-logs", srv.listPredictionLogsHandler)

	addr := envOrDefault("NADAA_ML_ADDR", defaultBindAddr)
	log.Printf("ml-service listening on %s with model %s", addr, srv.store.model.ModelVersion)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServer() (*server, error) {
	modelDir, err := resolveModelDir()
	if err != nil {
		return nil, err
	}
	return newServerFromModelDir(modelDir)
}

func newServerFromModelDir(modelDir string) (*server, error) {
	store, err := loadPredictionStore(modelDir)
	if err != nil {
		return nil, err
	}
	return &server{store: store}, nil
}

func loadPredictionStore(modelDir string) (*predictionStore, error) {
	modelPath := filepath.Join(modelDir, "baseline-logistic.v1.json")
	predictionPath := filepath.Join(modelDir, "sample-predictions.v1.json")

	var model modelArtifact
	if err := readJSONFile(modelPath, &model); err != nil {
		return nil, fmt.Errorf("load model artifact: %w", err)
	}

	var artifact predictionArtifact
	if err := readJSONFile(predictionPath, &artifact); err != nil {
		return nil, fmt.Errorf("load prediction artifact: %w", err)
	}

	if model.ModelVersion == "" || model.ModelVersion != artifact.ModelVersion {
		return nil, errors.New("model and prediction artifacts must share a modelVersion")
	}
	if artifact.FeatureSetVersion == "" || artifact.PredictionCount != len(artifact.Predictions) {
		return nil, errors.New("prediction artifact has invalid feature set or row count")
	}

	return &predictionStore{model: model, predictions: artifact.Predictions}, nil
}

func resolveModelDir() (string, error) {
	if value := strings.TrimSpace(os.Getenv("NADAA_ML_MODEL_DIR")); value != "" {
		return value, nil
	}

	candidates := []string{
		filepath.Join("data", "flood-risk", "models"),
		filepath.Join("..", "..", "data", "flood-risk", "models"),
		filepath.Join("/app", "data", "flood-risk", "models"),
	}
	for _, candidate := range candidates {
		if fileExists(filepath.Join(candidate, "baseline-logistic.v1.json")) && fileExists(filepath.Join(candidate, "sample-predictions.v1.json")) {
			return candidate, nil
		}
	}
	return "", errors.New("could not find flood-risk model artifacts; set NADAA_ML_MODEL_DIR")
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":       "ok",
		"service":      "ml-service",
		"modelVersion": s.store.model.ModelVersion,
	})
}

func (s *server) createFloodPredictionHandler(w http.ResponseWriter, r *http.Request) {
	var request predictionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	if !validCoordinates(request.Location) {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "location.lat must be between -90 and 90 and location.lng must be between -180 and 180")
		return
	}

	response, err := s.store.predict(request, time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "prediction_unavailable", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *server) listPredictionLogsHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, predictionLogListResponse{Logs: s.store.listLogs()})
}

func (s *predictionStore) predict(request predictionRequest, now time.Time) (predictionResponse, error) {
	prediction, distanceMeters, err := s.nearestPrediction(request.Location)
	if err != nil {
		return predictionResponse{}, err
	}

	summary := predictionSummary{
		ID:                     prediction.ID,
		ModelVersion:           prediction.ModelVersion,
		HazardType:             prediction.HazardType,
		PredictionTime:         prediction.PredictionTime,
		TargetTime:             prediction.TargetTime,
		CellID:                 prediction.CellID,
		Region:                 prediction.Region,
		District:               prediction.District,
		Community:              prediction.Community,
		Location:               request.Location,
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

	logEntry := predictionLog{
		ID:                     fmt.Sprintf("ml_log_%s_%s", now.Format("20060102150405"), sanitizeID(prediction.CellID)),
		PredictionID:           prediction.ID,
		ModelVersion:           prediction.ModelVersion,
		InputFeatureSetVersion: prediction.InputFeatureSetVersion,
		RequestedBy:            strings.TrimSpace(request.RequestedBy),
		CorrelationID:          strings.TrimSpace(request.CorrelationID),
		Location:               request.Location,
		StorageTarget:          "ml_predictions",
		HumanReviewRequired:    true,
		AutoPublishAllowed:     false,
		CreatedAt:              now.Format(time.RFC3339),
	}

	s.mu.Lock()
	s.logs = append(s.logs, logEntry)
	s.mu.Unlock()

	return predictionResponse{
		Prediction: summary,
		Log:        logEntry,
		Safety: safetyPolicy{
			HumanReviewRequired: true,
			AutoPublishAllowed:  false,
			Message:             "Model output is decision support only and cannot publish alerts without authority review and approval.",
		},
	}, nil
}

func (s *predictionStore) nearestPrediction(location coordinates) (storedPrediction, int, error) {
	if len(s.predictions) == 0 {
		return storedPrediction{}, 0, errors.New("no predictions are loaded")
	}

	type candidate struct {
		prediction storedPrediction
		distance   float64
	}
	candidates := make([]candidate, 0, len(s.predictions))
	for _, prediction := range s.predictions {
		centroid, ok := prediction.Geometry.Centroid()
		if !ok {
			continue
		}
		candidates = append(candidates, candidate{
			prediction: prediction,
			distance:   haversineMeters(location, centroid),
		})
	}
	if len(candidates) == 0 {
		return storedPrediction{}, 0, errors.New("loaded predictions do not include usable geometry")
	}

	sort.Slice(candidates, func(i int, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})
	return candidates[0].prediction, int(math.Round(candidates[0].distance)), nil
}

func (s *predictionStore) listLogs() []predictionLog {
	s.mu.Lock()
	defer s.mu.Unlock()

	logs := append([]predictionLog(nil), s.logs...)
	sort.Slice(logs, func(i int, j int) bool {
		return logs[i].CreatedAt > logs[j].CreatedAt
	})
	return logs
}

func (g polygonGeometry) Centroid() (coordinates, bool) {
	if len(g.Coordinates) == 0 || len(g.Coordinates[0]) == 0 {
		return coordinates{}, false
	}

	ring := g.Coordinates[0]
	if len(ring) > 1 && samePoint(ring[0], ring[len(ring)-1]) {
		ring = ring[:len(ring)-1]
	}
	if len(ring) == 0 {
		return coordinates{}, false
	}

	var latSum float64
	var lngSum float64
	for _, point := range ring {
		if len(point) < 2 {
			return coordinates{}, false
		}
		lngSum += point[0]
		latSum += point[1]
	}

	return coordinates{Lat: latSum / float64(len(ring)), Lng: lngSum / float64(len(ring))}, true
}

func samePoint(a []float64, b []float64) bool {
	return len(a) >= 2 && len(b) >= 2 && a[0] == b[0] && a[1] == b[1]
}

func readJSONFile(path string, target any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(target)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

func haversineMeters(a coordinates, b coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)

	sinLat := math.Sin(deltaLat / 2)
	sinLng := math.Sin(deltaLng / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func sanitizeID(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(value) {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			builder.WriteRune(char)
			continue
		}
		builder.WriteRune('_')
	}
	return strings.Trim(builder.String(), "_")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsFromEnv()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Cache-Control", "no-store")
}

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if key == "NADAA_ML_ADDR" && !strings.Contains(value, ":") {
		if _, err := strconv.Atoi(value); err == nil {
			return ":" + value
		}
	}
	return value
}
