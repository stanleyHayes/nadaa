package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/models"
)

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

// WriteError writes a structured API error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, models.APIError{Error: models.APIErrorBody{Code: code, Message: message}})
}

// WithCORS wraps a handler with security and CORS headers.
func WithCORS(allowedOrigins map[string]bool, next http.Handler) http.Handler {
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// EnvOrDefault returns the value of key or fallback if unset.
func EnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// RiskRank returns a numeric rank for a risk level string.
func RiskRank(level string) int {
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

// RecommendedActions returns guidance based on the overall and flood risk levels.
func RecommendedActions(overall string, floodLevel string) []string {
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

// RisksFloodLevel returns the level of the flood risk entry, or low if absent.
func RisksFloodLevel(risks []models.RiskSummary) string {
	for _, risk := range risks {
		if risk.Type == "flood" {
			return risk.Level
		}
	}
	return "low"
}
