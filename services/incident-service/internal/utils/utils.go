package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
)

var (
	// MediaRefPattern validates safe media/user/agency identifiers.
	MediaRefPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,128}$`)
	// WordPattern extracts lowercase alphanumeric tokens.
	WordPattern = regexp.MustCompile(`[a-z0-9]+`)
	// AllowedMediaTypes maps supported media types to their max size in bytes.
	AllowedMediaTypes = map[string]int64{
		"image/jpeg":      10 * 1024 * 1024,
		"image/png":       10 * 1024 * 1024,
		"image/webp":      10 * 1024 * 1024,
		"video/mp4":       100 * 1024 * 1024,
		"video/quicktime": 100 * 1024 * 1024,
		"audio/mpeg":      25 * 1024 * 1024,
		"audio/mp4":       25 * 1024 * 1024,
		"audio/wav":       25 * 1024 * 1024,
	}
)

// DecodeJSON decodes a JSON request body into target.
func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// OptionalDecodeJSON decodes a JSON body when content is present.
func OptionalDecodeJSON(r *http.Request, target any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	return DecodeJSON(r, target)
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR incident-service write_json_response_failed error=%v", err)
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NADAA-Actor-ID, X-NADAA-Actor-Role, X-NADAA-Agency-ID, X-NADAA-MFA-Completed, X-NADAA-Request-ID")
}

// EnvOrDefault returns the value of key or fallback if unset.
func EnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// EnvIntOrDefault returns the integer value of key or fallback if unset/invalid.
func EnvIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// AllowedOriginsFromEnv parses the NADAA_ALLOWED_ORIGINS allowlist.
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

// ValidCoordinates validates latitude and longitude ranges.
func ValidCoordinates(location models.Coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

// ValidFileName validates a safe upload file name.
func ValidFileName(fileName string) bool {
	if fileName == "" || len(fileName) > 180 || UnsafeText(fileName) {
		return false
	}
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") || strings.Contains(fileName, "..") {
		return false
	}
	return true
}

// UnsafeText rejects strings containing common script injection markers.
func UnsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

// ClientIdentifier extracts the best-effort client IP from a request.
func ClientIdentifier(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil || host == "" {
		return r.RemoteAddr
	}
	return host
}

// NewID generates a random prefixed identifier.
func NewID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, bytes)
}

// NormalizeQueryValue trims and lowercases a query value.
func NormalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

// IncidentStatusSlug normalizes an incident status string.
func IncidentStatusSlug(status string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(status)), "-", "_"), " ", "_")
}

// RequiresResolutionNotes returns true for terminal statuses that need notes.
func RequiresResolutionNotes(status string) bool {
	return status == "closed" || status == "false_report"
}
