package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type riskResponse struct {
	Location           string           `json:"location"`
	OverallRisk        string           `json:"overallRisk"`
	Risks              []riskSummary    `json:"risks"`
	NearestShelters    []shelterSummary `json:"nearestShelters"`
	RecommendedActions []string         `json:"recommendedActions"`
}

type riskSummary struct {
	Type        string  `json:"type"`
	Level       string  `json:"level"`
	Probability float64 `json:"probability"`
	Reason      string  `json:"reason"`
}

type shelterSummary struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Location         coordinates `json:"location"`
	Capacity         int         `json:"capacity"`
	CurrentOccupancy int         `json:"currentOccupancy"`
	Contact          string      `json:"contact"`
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /api/v1/risk", riskHandler)

	log.Println("risk-service listening on :8081")
	if err := http.ListenAndServe(":8081", withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "risk-service"})
}

func riskHandler(w http.ResponseWriter, r *http.Request) {
	lat, latErr := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, lngErr := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if latErr != nil || lngErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "lat and lng query parameters are required"})
		return
	}

	response := riskResponse{
		Location:    inferLocation(lat, lng),
		OverallRisk: "high",
		Risks: []riskSummary{
			{
				Type:        "flood",
				Level:       "severe",
				Probability: 0.82,
				Reason:      "Heavy rainfall forecast, low elevation, and historical flood reports nearby.",
			},
			{
				Type:        "fire",
				Level:       "moderate",
				Probability: 0.34,
				Reason:      "Dense activity and response access constraints increase localized risk.",
			},
		},
		NearestShelters: []shelterSummary{
			{
				ID:               "shelter-ama-001",
				Name:             "Accra Metro Assembly Shelter",
				Location:         coordinates{Lat: 5.56, Lng: -0.2},
				Capacity:         450,
				CurrentOccupancy: 116,
				Contact:          "112",
			},
		},
		RecommendedActions: []string{
			"Avoid low-lying roads and open drains.",
			"Move valuables above ground level.",
			"Prepare an evacuation route to the nearest safe shelter.",
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func inferLocation(lat float64, lng float64) string {
	if lat >= 5.4 && lat <= 5.8 && lng >= -0.4 && lng <= 0.1 {
		return "Accra Central"
	}

	return "Selected area"
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
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

