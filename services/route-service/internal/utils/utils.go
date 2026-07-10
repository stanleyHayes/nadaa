package utils

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/stanleyHayes/nadaa/services/route-service/internal/models"
)

const (
	// EarthRadiusMeters is the approximate radius of the Earth in meters.
	EarthRadiusMeters = 6371000.0
	// DefaultClosureBufferMeters is the default distance to keep from a road closure.
	DefaultClosureBufferMeters = 100.0
	// DefaultDetourOffsetMeters is how far a detour pushes away from a blocked sample.
	DefaultDetourOffsetMeters = 300.0
	// DefaultSampleStepMeters is the spacing between sampled points along a route.
	DefaultSampleStepMeters = 200.0
	// WalkingSpeedMetersPerSecond is the assumed evacuation walking speed.
	WalkingSpeedMetersPerSecond = 1.4
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
		log.Printf("ERROR route-service write_json_response_failed error=%v", err)
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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

// NormalizeString trims and lowercases a value.
func NormalizeString(value string) string {
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

// Bearing returns the initial bearing from a to b in degrees clockwise from north.
func Bearing(a, b models.Coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)
	y := math.Sin(deltaLng) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLng)
	bearing := math.Atan2(y, x) * 180 / math.Pi
	if bearing < 0 {
		bearing += 360
	}
	return bearing
}

// DestinationPoint returns the coordinate reached by travelling distanceMeters
// along initial bearing from start.
func DestinationPoint(start models.Coordinates, distanceMeters, bearing float64) models.Coordinates {
	angularDistance := distanceMeters / EarthRadiusMeters
	bearingRad := degreesToRadians(bearing)
	lat1 := degreesToRadians(start.Lat)
	lng1 := degreesToRadians(start.Lng)

	lat2 := math.Asin(math.Sin(lat1)*math.Cos(angularDistance) +
		math.Cos(lat1)*math.Sin(angularDistance)*math.Cos(bearingRad))
	lng2 := lng1 + math.Atan2(math.Sin(bearingRad)*math.Sin(angularDistance)*math.Cos(lat1),
		math.Cos(angularDistance)-math.Sin(lat1)*math.Sin(lat2))

	return models.Coordinates{Lat: lat2 * 180 / math.Pi, Lng: normalizeLng(lng2 * 180 / math.Pi)}
}

func normalizeLng(value float64) float64 {
	for value > 180 {
		value -= 360
	}
	for value < -180 {
		value += 360
	}
	return value
}

// Interpolate returns a point fraction t (0..1) along the great-circle path from a to b.
func Interpolate(a, b models.Coordinates, t float64) models.Coordinates {
	d := DistanceMeters(a, b)
	if d < 1 {
		return a
	}
	bearing := Bearing(a, b)
	return DestinationPoint(a, t*d, bearing)
}

// PointInPolygon reports whether point is inside the polygon using the ray-casting algorithm.
func PointInPolygon(point models.Coordinates, polygon []models.Coordinates) bool {
	inside := false
	j := len(polygon) - 1
	for i := range polygon {
		pi := polygon[i]
		pj := polygon[j]
		if ((pi.Lng > point.Lng) != (pj.Lng > point.Lng)) &&
			(point.Lat < (pj.Lat-pi.Lat)*(point.Lng-pi.Lng)/(pj.Lng-pi.Lng)+pi.Lat) {
			inside = !inside
		}
		j = i
	}
	return inside
}

// MinDistanceToLineString returns the minimum distance from a location to any coordinate in a LineString.
func MinDistanceToLineString(location models.Coordinates, coordinates [][]float64) float64 {
	minDistance := math.MaxFloat64
	for _, point := range coordinates {
		if len(point) < 2 {
			continue
		}
		d := DistanceMeters(location, models.Coordinates{Lat: point[1], Lng: point[0]})
		if d < minDistance {
			minDistance = d
		}
	}
	return minDistance
}

// NormalizeRiskLevels lowercases and trims a slice of risk level strings.
func NormalizeRiskLevels(levels []string) []string {
	out := make([]string, 0, len(levels))
	for _, level := range levels {
		level = NormalizeString(level)
		if level != "" {
			out = append(out, level)
		}
	}
	return out
}

// ShouldAvoidRisk reports whether a risk level should be avoided.
func ShouldAvoidRisk(level string, avoid []string) bool {
	level = NormalizeString(level)
	return slices.Contains(avoid, level)
}
