package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
)

// DecodeJSON decodes a JSON request body into target, rejecting unknown fields.
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
		log.Printf("ERROR damage-claim-service write_json_response_failed error=%v", err)
	}
}

// WriteError writes a structured API error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, models.APIError{Error: models.APIErrorBody{Code: code, Message: message}})
}

// WithCORS wraps a handler with security and CORS headers.
func WithCORS(allowedOrigins map[string]bool, next http.Handler) http.Handler {
	// Localhost/127.0.0.1 origins bypass the configured allowlist only in
	// development; in every other environment the allowlist is authoritative.
	allowLocalhost := strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_ENV")), "development")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins, allowLocalhost)
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

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool, allowLocalhost bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		localhostOrigin := strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
		if allowedOrigins[origin] || (allowLocalhost && localhostOrigin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

// EnvOrDefault returns the value of key or fallback if unset.
func EnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// AllowedOriginsFromEnv parses NADAA_ALLOWED_ORIGINS into a set.
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

// TrimTrailingSlash removes trailing slashes from a URL string.
func TrimTrailingSlash(value string) string {
	return strings.TrimRight(value, "/")
}

// NormalizeString trims and lowercases a value.
func NormalizeString(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

// NormalizeToken lowercases, trims, and replaces spaces/hyphens with underscores.
func NormalizeToken(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(value)), "-", "_"), " ", "_")
}

// UnsafeText checks for common script injection markers.
func UnsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

// ValidCoordinates returns true if location is within WGS84 bounds.
func ValidCoordinates(location models.ClaimLocation) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

// ValidEmail reports whether value looks like a valid email address.
func ValidEmail(value string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(value)
}

// decimalPattern matches a plain base-10 decimal: digits with an optional
// fractional part. It rejects NaN, Inf, hex, and scientific notation, all of
// which strconv.ParseFloat would otherwise accept.
var decimalPattern = regexp.MustCompile(`^\d+(\.\d+)?$`)

// ValidDecimal reports whether value is a valid base-10 decimal.
func ValidDecimal(value string) bool {
	return decimalPattern.MatchString(value)
}

// SafeLogValue strips CR/LF from user-controlled values before they are
// written to logs, preventing forged log lines.
func SafeLogValue(value string) string {
	return strings.NewReplacer("\n", " ", "\r", " ").Replace(value)
}

// StatusForCode maps internal error codes to HTTP status codes.
func StatusForCode(code string) int {
	if code == "not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
