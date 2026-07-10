package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
)

const (
	// EarthRadiusMeters is the approximate radius of the Earth in meters.
	EarthRadiusMeters = 6371000.0
	// NearbySearchMeters is the default search radius for nearby closure queries.
	NearbySearchMeters = 30000.0
	// DefaultLimit is the default page size for listing closures.
	DefaultLimit = 50
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
		log.Printf("ERROR road-closure-service write_json_response_failed error=%v", err)
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

// NormalizeQueryValue trims and lowercases a query value.
func NormalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

// UnsafeText checks for common script injection markers.
func UnsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

// ValidCoordinates returns true if location is within WGS84 bounds.
func ValidCoordinates(location models.Coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

// DistanceMeters returns the haversine distance between two coordinates.
func DistanceMeters(a, b models.Coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)
	h := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return EarthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

// ParseWKTLineString parses a LINESTRING WKT string into a geometry.
func ParseWKTLineString(value string) (models.LineStringGeometry, error) {
	value = strings.TrimSpace(value)
	prefix := "LINESTRING("
	suffix := ")"
	if !strings.HasPrefix(strings.ToUpper(value), prefix) || !strings.HasSuffix(value, suffix) {
		return models.LineStringGeometry{}, errors.New("geometry must be a LINESTRING(...)")
	}
	inner := value[len(prefix) : len(value)-len(suffix)]
	parts := strings.Split(inner, ",")
	coords := make([][]float64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		nums := strings.Fields(part)
		if len(nums) != 2 {
			return models.LineStringGeometry{}, errors.New("each LINESTRING coordinate must have two numbers")
		}
		lng, err1 := strconv.ParseFloat(nums[0], 64)
		lat, err2 := strconv.ParseFloat(nums[1], 64)
		if err1 != nil || err2 != nil {
			return models.LineStringGeometry{}, errors.New("LINESTRING coordinates must be valid decimal numbers")
		}
		coords = append(coords, []float64{lng, lat})
	}
	return models.LineStringGeometry{Type: "LineString", Coordinates: coords}, nil
}

// FormatWKTLineString formats a geometry as a WKT LINESTRING.
func FormatWKTLineString(geometry models.LineStringGeometry) string {
	parts := make([]string, 0, len(geometry.Coordinates))
	for _, c := range geometry.Coordinates {
		parts = append(parts, fmt.Sprintf("%f %f", c[0], c[1]))
	}
	return "LINESTRING(" + strings.Join(parts, ", ") + ")"
}

// ValidateGeometry validates a LineString geometry.
func ValidateGeometry(geometry models.LineStringGeometry) (string, string) {
	if geometry.Type != "LineString" {
		return "invalid_geometry_type", "geometry type must be LineString"
	}
	if len(geometry.Coordinates) < 2 {
		return "invalid_geometry", "LineString must contain at least two coordinates"
	}
	for _, point := range geometry.Coordinates {
		if len(point) != 2 {
			return "invalid_geometry", "each LineString coordinate must be [lng, lat]"
		}
		if point[0] < -180 || point[0] > 180 || point[1] < -90 || point[1] > 90 {
			return "invalid_geometry", "coordinates must be valid WGS84 longitude and latitude"
		}
	}
	return "", ""
}

// SeverityFromStatus maps a closure status to a default severity.
func SeverityFromStatus(status string) string {
	switch status {
	case "active":
		return "high"
	case "scheduled":
		return "low"
	case "lifted", "cancelled":
		return "low"
	default:
		return "moderate"
	}
}

// StatusForCode maps error codes to HTTP status codes.
func StatusForCode(code string) int {
	if code == "not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// IsClosureEffective returns true if the closure is currently effective.
func IsClosureEffective(closure models.RoadClosureRecord, now time.Time) bool {
	if closure.Status == "lifted" || closure.Status == "cancelled" {
		return false
	}
	if closure.ValidFrom.After(now) {
		return false
	}
	if closure.ValidTo != nil && closure.ValidTo.Before(now) {
		return false
	}
	return true
}

// ClosureIntersectsBBox returns true if any geometry coordinate is inside the box.
func ClosureIntersectsBBox(geometry models.LineStringGeometry, box models.BBox) bool {
	for _, point := range geometry.Coordinates {
		lng, lat := point[0], point[1]
		if lng >= box.MinLng && lng <= box.MaxLng && lat >= box.MinLat && lat <= box.MaxLat {
			return true
		}
	}
	return false
}

// MinDistanceToLineString returns the minimum distance from a location to any geometry coordinate.
func MinDistanceToLineString(location models.Coordinates, geometry models.LineStringGeometry) float64 {
	minDistance := math.MaxFloat64
	for _, point := range geometry.Coordinates {
		d := DistanceMeters(location, models.Coordinates{Lat: point[1], Lng: point[0]})
		if d < minDistance {
			minDistance = d
		}
	}
	return minDistance
}

// StatusRank returns a sort rank for status.
func StatusRank(status string) int {
	switch status {
	case "active":
		return 0
	case "scheduled":
		return 1
	case "lifted":
		return 2
	case "cancelled":
		return 3
	default:
		return 4
	}
}

// SeverityRank returns a sort rank for severity.
func SeverityRank(severity string) int {
	switch severity {
	case "emergency":
		return 0
	case "severe":
		return 1
	case "high":
		return 2
	case "moderate":
		return 3
	case "low":
		return 4
	default:
		return 5
	}
}

// CopyClosures returns a shallow copy of the closure slice with fresh coordinate slices.
func CopyClosures(source []models.RoadClosureRecord) []models.RoadClosureRecord {
	closures := make([]models.RoadClosureRecord, 0, len(source))
	for _, closure := range source {
		closure.Geometry.Coordinates = append([][]float64{}, closure.Geometry.Coordinates...)
		closures = append(closures, closure)
	}
	return closures
}
