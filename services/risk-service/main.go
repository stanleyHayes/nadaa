package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type server struct {
	store    *riskStore
	mlClient *mlClient
}

type riskStore struct {
	riskZones         []riskZone
	shelters          []shelterSummary
	facilities        []facilitySummary
	historicalReports []historicalReport
}

type riskResponse struct {
	Location           string            `json:"location"`
	OverallRisk        string            `json:"overallRisk"`
	Risks              []riskSummary     `json:"risks"`
	MLPrediction       *mlPrediction     `json:"mlPrediction,omitempty"`
	NearestShelters    []shelterSummary  `json:"nearestShelters"`
	NearbyFacilities   []facilitySummary `json:"nearbyFacilities"`
	RecommendedActions []string          `json:"recommendedActions"`
}

type riskSummary struct {
	Type        string  `json:"type"`
	Level       string  `json:"level"`
	Probability float64 `json:"probability"`
	Reason      string  `json:"reason"`
}

type mlPrediction struct {
	ID                     string                `json:"id"`
	ModelVersion           string                `json:"modelVersion"`
	HazardType             string                `json:"hazardType"`
	PredictionTime         string                `json:"predictionTime"`
	TargetTime             string                `json:"targetTime"`
	CellID                 string                `json:"cellId"`
	Region                 string                `json:"region"`
	District               string                `json:"district"`
	Community              string                `json:"community"`
	Probability            float64               `json:"probability"`
	Severity               string                `json:"severity"`
	ExpectedOnset          string                `json:"expectedOnset"`
	Confidence             string                `json:"confidence"`
	ExplanationFactors     []mlExplanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string                `json:"inputFeatureSetVersion"`
	PredictionLogID        string                `json:"predictionLogId"`
	HumanReviewRequired    bool                  `json:"humanReviewRequired"`
	AutoPublishAllowed     bool                  `json:"autoPublishAllowed"`
	Source                 string                `json:"source"`
}

type mlExplanationFactor struct {
	Feature      string  `json:"feature"`
	Label        string  `json:"label"`
	Value        any     `json:"value"`
	Contribution float64 `json:"contribution"`
	Direction    string  `json:"direction"`
}

type mlPredictionResponse struct {
	Prediction mlPredictionPayload `json:"prediction"`
	Log        mlPredictionLog     `json:"log"`
}

type mlPredictionPayload struct {
	ID                     string                `json:"id"`
	ModelVersion           string                `json:"modelVersion"`
	HazardType             string                `json:"hazardType"`
	PredictionTime         string                `json:"predictionTime"`
	TargetTime             string                `json:"targetTime"`
	CellID                 string                `json:"cellId"`
	Region                 string                `json:"region"`
	District               string                `json:"district"`
	Community              string                `json:"community"`
	Probability            float64               `json:"probability"`
	Severity               string                `json:"severity"`
	ExpectedOnset          string                `json:"expectedOnset"`
	Confidence             string                `json:"confidence"`
	ExplanationFactors     []mlExplanationFactor `json:"explanationFactors"`
	InputFeatureSetVersion string                `json:"inputFeatureSetVersion"`
	HumanReviewRequired    bool                  `json:"humanReviewRequired"`
	AutoPublishAllowed     bool                  `json:"autoPublishAllowed"`
	Source                 string                `json:"source"`
}

type mlPredictionLog struct {
	ID                     string `json:"id"`
	ModelVersion           string `json:"modelVersion"`
	InputFeatureSetVersion string `json:"inputFeatureSetVersion"`
}

type mlPredictionRequest struct {
	Location      coordinates `json:"location"`
	RequestedBy   string      `json:"requestedBy"`
	CorrelationID string      `json:"correlationId"`
}

type mlClient struct {
	baseURL    string
	httpClient *http.Client
}

type shelterSummary struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Location         coordinates `json:"location"`
	Capacity         int         `json:"capacity,omitempty"`
	CurrentOccupancy int         `json:"currentOccupancy,omitempty"`
	Contact          string      `json:"contact,omitempty"`
	DistanceMeters   int         `json:"distanceMeters,omitempty"`
	Status           string      `json:"status,omitempty"`
	Facilities       []string    `json:"facilities,omitempty"`
}

type facilitySummary struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Location       coordinates `json:"location"`
	Region         string      `json:"region,omitempty"`
	District       string      `json:"district,omitempty"`
	Contact        string      `json:"contact,omitempty"`
	DistanceMeters int         `json:"distanceMeters,omitempty"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type riskZone struct {
	ID          string
	HazardType  string
	RiskLevel   string
	Bounds      boundingBox
	Probability float64
	Explanation string
}

type historicalReport struct {
	HazardType string
	Location   coordinates
	Severity   string
}

type boundingBox struct {
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	earthRadiusMeters       = 6371000.0
	nearbyShelterRadius     = 30000.0
	nearbyFacilityRadius    = 30000.0
	nearbyRiskZoneThreshold = 2500.0
	recentReportThreshold   = 3000.0
)

func main() {
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/risk", srv.riskHandler)

	addr := envOrDefault("NADAA_RISK_ADDR", ":8081")
	log.Printf("risk-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServer() *server {
	return &server{store: newFixtureRiskStore(), mlClient: newMLClientFromEnv()}
}

func newFixtureRiskStore() *riskStore {
	return &riskStore{
		riskZones: []riskZone{
			{
				ID:          "00000000-0000-0000-0000-000000000401",
				HazardType:  "flood",
				RiskLevel:   "severe",
				Bounds:      boundingBox{MinLat: 5.530, MaxLat: 5.590, MinLng: -0.230, MaxLng: -0.160},
				Probability: 0.86,
				Explanation: "Low-lying Accra sample zone with historical flood reports and rainfall sensitivity.",
			},
			{
				ID:          "00000000-0000-0000-0000-000000000402",
				HazardType:  "fire",
				RiskLevel:   "moderate",
				Bounds:      boundingBox{MinLat: 5.540, MaxLat: 5.610, MinLng: -0.210, MaxLng: -0.140},
				Probability: 0.38,
				Explanation: "Dense commercial area sample zone.",
			},
		},
		shelters: []shelterSummary{
			{
				ID:               "00000000-0000-0000-0000-000000000301",
				Name:             "Accra Metro Assembly Shelter",
				Location:         coordinates{Lat: 5.560, Lng: -0.200},
				Capacity:         450,
				CurrentOccupancy: 116,
				Contact:          "112",
				Status:           "open",
				Facilities:       []string{"water", "first_aid", "accessible_entry", "family_area"},
			},
			{
				ID:               "00000000-0000-0000-0000-000000000302",
				Name:             "Osu Community Hall",
				Location:         coordinates{Lat: 5.550, Lng: -0.180},
				Capacity:         220,
				CurrentOccupancy: 34,
				Contact:          "112",
				Status:           "open",
				Facilities:       []string{"water", "first_aid"},
			},
		},
		facilities: []facilitySummary{
			{
				ID:       "00000000-0000-0000-0000-000000000101",
				Name:     "NADMO Accra Metro",
				Type:     "nadmo",
				Location: coordinates{Lat: 5.560, Lng: -0.200},
				Region:   "Greater Accra",
				District: "Accra Metropolitan",
				Contact:  "112",
			},
			{
				ID:       "00000000-0000-0000-0000-000000000102",
				Name:     "Ghana National Fire Service Accra",
				Type:     "fire",
				Location: coordinates{Lat: 5.565, Lng: -0.185},
				Region:   "Greater Accra",
				District: "Accra Metropolitan",
				Contact:  "112",
			},
			{
				ID:       "00000000-0000-0000-0000-000000000103",
				Name:     "National Ambulance Service Accra",
				Type:     "ambulance",
				Location: coordinates{Lat: 5.555, Lng: -0.190},
				Region:   "Greater Accra",
				District: "Accra Metropolitan",
				Contact:  "112",
			},
		},
		historicalReports: []historicalReport{
			{
				HazardType: "flood",
				Location:   coordinates{Lat: 5.6037, Lng: -0.1870},
				Severity:   "high",
			},
		},
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "risk-service"})
}

func (s *server) riskHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}

	risk := s.store.areaRisk(location)
	if s.mlClient != nil {
		if prediction, err := s.mlClient.predict(r.Context(), location); err != nil {
			log.Printf("ml prediction unavailable: %v", err)
		} else {
			risk.MLPrediction = &prediction
			if riskRank(prediction.Severity) > riskRank(risk.OverallRisk) {
				risk.OverallRisk = prediction.Severity
				risk.RecommendedActions = recommendedActions(risk.OverallRisk, risksFloodLevel(risk.Risks))
			}
		}
	}
	writeJSON(w, http.StatusOK, risk)
}

func parseLocation(w http.ResponseWriter, r *http.Request) (coordinates, bool) {
	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" || lngText == "" {
		writeError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng query parameters are required")
		return coordinates{}, false
	}

	lat, latErr := strconv.ParseFloat(latText, 64)
	lng, lngErr := strconv.ParseFloat(lngText, 64)
	if latErr != nil || lngErr != nil {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
		return coordinates{}, false
	}

	location := coordinates{Lat: lat, Lng: lng}
	if !validCoordinates(location) {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
		return coordinates{}, false
	}

	return location, true
}

func (s *riskStore) areaRisk(location coordinates) riskResponse {
	risks := []riskSummary{s.floodRisk(location)}
	if fireRisk, ok := s.fireRisk(location); ok {
		risks = append(risks, fireRisk)
	}

	sort.Slice(risks, func(i, j int) bool {
		return riskRank(risks[i].Level) > riskRank(risks[j].Level)
	})

	overall := "low"
	for _, risk := range risks {
		if riskRank(risk.Level) > riskRank(overall) {
			overall = risk.Level
		}
	}

	return riskResponse{
		Location:           inferLocation(location),
		OverallRisk:        overall,
		Risks:              risks,
		NearestShelters:    s.nearestShelters(location),
		NearbyFacilities:   s.nearbyFacilities(location),
		RecommendedActions: recommendedActions(overall, risks[0].Level),
	}
}

func (s *riskStore) floodRisk(location coordinates) riskSummary {
	for _, zone := range s.riskZones {
		if zone.HazardType != "flood" {
			continue
		}
		if zone.Bounds.Contains(location) {
			return riskSummary{
				Type:        "flood",
				Level:       zone.RiskLevel,
				Probability: zone.Probability,
				Reason:      zone.Explanation,
			}
		}
	}

	nearestZoneDistance := math.MaxFloat64
	for _, zone := range s.riskZones {
		if zone.HazardType != "flood" {
			continue
		}
		nearestZoneDistance = math.Min(nearestZoneDistance, zone.Bounds.DistanceMeters(location))
	}

	nearestReportDistance := math.MaxFloat64
	nearestReportSeverity := ""
	for _, report := range s.historicalReports {
		if report.HazardType != "flood" {
			continue
		}
		distance := haversineMeters(location, report.Location)
		if distance < nearestReportDistance {
			nearestReportDistance = distance
			nearestReportSeverity = report.Severity
		}
	}

	if nearestZoneDistance <= nearbyRiskZoneThreshold && nearestReportDistance <= recentReportThreshold {
		return riskSummary{
			Type:        "flood",
			Level:       "high",
			Probability: 0.72,
			Reason:      fmt.Sprintf("Within %.0fm of a severe flood zone and %.0fm of a recent %s flood report.", nearestZoneDistance, nearestReportDistance, nearestReportSeverity),
		}
	}

	if nearestZoneDistance <= nearbyRiskZoneThreshold {
		return riskSummary{
			Type:        "flood",
			Level:       "high",
			Probability: 0.64,
			Reason:      fmt.Sprintf("Within %.0fm of a severe flood-prone zone.", nearestZoneDistance),
		}
	}

	if nearestReportDistance <= recentReportThreshold {
		return riskSummary{
			Type:        "flood",
			Level:       "moderate",
			Probability: 0.42,
			Reason:      fmt.Sprintf("Within %.0fm of a recent %s flood report.", nearestReportDistance, nearestReportSeverity),
		}
	}

	return riskSummary{
		Type:        "flood",
		Level:       "low",
		Probability: 0.16,
		Reason:      "No active flood risk zone or recent flood report is near these coordinates in the MVP fixture set.",
	}
}

func (s *riskStore) fireRisk(location coordinates) (riskSummary, bool) {
	for _, zone := range s.riskZones {
		if zone.HazardType != "fire" {
			continue
		}
		if zone.Bounds.Contains(location) {
			return riskSummary{
				Type:        "fire",
				Level:       zone.RiskLevel,
				Probability: zone.Probability,
				Reason:      zone.Explanation,
			}, true
		}
	}
	return riskSummary{}, false
}

func (s *riskStore) nearestShelters(location coordinates) []shelterSummary {
	shelters := make([]shelterSummary, 0, len(s.shelters))
	for _, shelter := range s.shelters {
		distance := haversineMeters(location, shelter.Location)
		if distance > nearbyShelterRadius {
			continue
		}
		shelter.DistanceMeters = int(math.Round(distance))
		shelters = append(shelters, shelter)
	}

	sort.Slice(shelters, func(i, j int) bool {
		return shelters[i].DistanceMeters < shelters[j].DistanceMeters
	})
	return shelters
}

func (s *riskStore) nearbyFacilities(location coordinates) []facilitySummary {
	facilities := make([]facilitySummary, 0, len(s.facilities))
	for _, facility := range s.facilities {
		distance := haversineMeters(location, facility.Location)
		if distance > nearbyFacilityRadius {
			continue
		}
		facility.DistanceMeters = int(math.Round(distance))
		facilities = append(facilities, facility)
	}

	sort.Slice(facilities, func(i, j int) bool {
		return facilities[i].DistanceMeters < facilities[j].DistanceMeters
	})
	return facilities
}

func recommendedActions(overall string, floodLevel string) []string {
	if overall == "severe" || overall == "emergency" {
		return []string{
			"Avoid low-lying roads, open drains, and moving floodwater.",
			"Prepare to move vulnerable people and key documents to higher ground.",
			"Identify the nearest open shelter and keep emergency contacts reachable.",
		}
	}

	if floodLevel == "high" {
		return []string{
			"Avoid known flood-prone roads and monitor official NADMO updates.",
			"Check the route to the nearest open shelter before rainfall intensifies.",
			"Keep phones charged and report blocked drains or rising water early.",
		}
	}

	if floodLevel == "moderate" {
		return []string{
			"Monitor rainfall and local alerts.",
			"Keep drains around your home or workplace clear where safe.",
			"Plan a safe route away from low-lying roads.",
		}
	}

	return []string{
		"Stay aware of official alerts for changing weather conditions.",
		"Know the nearest shelter and emergency contact number.",
		"Report early signs of flooding or blocked drains.",
	}
}

func risksFloodLevel(risks []riskSummary) string {
	for _, risk := range risks {
		if risk.Type == "flood" {
			return risk.Level
		}
	}
	return "low"
}

func inferLocation(location coordinates) string {
	if location.Lat >= 5.50 && location.Lat <= 5.66 && location.Lng >= -0.28 && location.Lng <= -0.08 {
		return "Accra Metropolitan"
	}
	if location.Lat >= 6.55 && location.Lat <= 6.80 && location.Lng >= -1.75 && location.Lng <= -1.45 {
		return "Kumasi area"
	}
	return "Selected area"
}

func (b boundingBox) Contains(location coordinates) bool {
	return location.Lat >= b.MinLat && location.Lat <= b.MaxLat && location.Lng >= b.MinLng && location.Lng <= b.MaxLng
}

func (b boundingBox) DistanceMeters(location coordinates) float64 {
	if b.Contains(location) {
		return 0
	}

	nearest := coordinates{
		Lat: clamp(location.Lat, b.MinLat, b.MaxLat),
		Lng: clamp(location.Lng, b.MinLng, b.MaxLng),
	}
	return haversineMeters(location, nearest)
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

func clamp(value float64, minimum float64, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

func riskRank(level string) int {
	switch level {
	case "emergency":
		return 5
	case "severe":
		return 4
	case "high":
		return 3
	case "moderate":
		return 2
	default:
		return 1
	}
}

func newMLClientFromEnv() *mlClient {
	baseURL := strings.TrimSpace(os.Getenv("NADAA_ML_API_URL"))
	if baseURL == "" {
		return nil
	}
	return newMLClient(baseURL, http.DefaultClient)
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

func (c *mlClient) predict(ctx context.Context, location coordinates) (mlPrediction, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	requestPayload := mlPredictionRequest{
		Location:      location,
		RequestedBy:   "risk-service",
		CorrelationID: fmt.Sprintf("risk_%0.4f_%0.4f", location.Lat, location.Lng),
	}
	body, err := json.Marshal(requestPayload)
	if err != nil {
		return mlPrediction{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/ml/flood/predictions", bytes.NewReader(body))
	if err != nil {
		return mlPrediction{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return mlPrediction{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return mlPrediction{}, fmt.Errorf("ml service returned %d", response.StatusCode)
	}

	var payload mlPredictionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return mlPrediction{}, err
	}
	if payload.Prediction.ModelVersion == "" {
		return mlPrediction{}, fmt.Errorf("ml service returned an empty modelVersion")
	}

	return mlPrediction{
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
