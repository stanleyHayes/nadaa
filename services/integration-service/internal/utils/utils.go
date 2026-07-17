package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
)

// DecodeJSON decodes a JSON request body into target.
func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

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
	if allowedOrigins[origin] || (isDevelopmentEnv() && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:"))) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

// AllowedOriginsFromEnv returns the parsed NADAA_ALLOWED_ORIGINS value.
func AllowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for origin := range strings.SplitSeq(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

// EnvOrDefault returns the value of key or fallback if unset.
func EnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// NormalizeQueryValue trims and lowercases a query value.
func NormalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

// SanitizeID normalizes a value for use in an identifier.
func SanitizeID(value string) string {
	value = NormalizeQueryValue(value)
	replacer := strings.NewReplacer(" ", "_", "/", "_", ":", "_")
	return replacer.Replace(value)
}

// DefaultImportAdapterID returns the default adapter ID when empty.
func DefaultImportAdapterID(adapterID string) string {
	adapterID = strings.TrimSpace(adapterID)
	if adapterID == "" {
		return "mock-weather-hydrology-adapter"
	}
	return adapterID
}

// SanitizeLogValue strips line breaks so user-controlled values cannot forge log lines.
func SanitizeLogValue(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, value)
}

// ValidateWKTLineString validates a LINESTRING WKT string, returning an error code and message.
// It mirrors the road-closure-service adapter validation so invalid geometry is rejected locally.
func ValidateWKTLineString(value string) (string, string) {
	trimmed := strings.TrimSpace(value)
	prefix := "LINESTRING("
	if !strings.HasPrefix(strings.ToUpper(trimmed), prefix) || !strings.HasSuffix(trimmed, ")") {
		return "invalid_geometry", "geometry must be a LINESTRING(...)"
	}
	parts := strings.Split(trimmed[len(prefix):len(trimmed)-1], ",")
	if len(parts) < 2 {
		return "invalid_geometry", "LineString must contain at least two coordinates"
	}
	for _, part := range parts {
		nums := strings.Fields(strings.TrimSpace(part))
		if len(nums) != 2 {
			return "invalid_geometry", "each LINESTRING coordinate must have two numbers"
		}
		lng, err1 := strconv.ParseFloat(nums[0], 64)
		lat, err2 := strconv.ParseFloat(nums[1], 64)
		if err1 != nil || err2 != nil {
			return "invalid_geometry", "LINESTRING coordinates must be valid decimal numbers"
		}
		if lng < -180 || lng > 180 || lat < -90 || lat > 90 {
			return "invalid_geometry", "coordinates must be valid WGS84 longitude and latitude"
		}
	}
	return "", ""
}

func isDevelopmentEnv() bool {
	return NormalizeQueryValue(os.Getenv("NADAA_ENV")) == "development"
}

// PluralSuffix returns an "s" suffix for counts other than one.
func PluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
