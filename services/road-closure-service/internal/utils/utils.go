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
	// Loopback origins are only echoed back in local development; when an
	// explicit allowlist is configured it is honored as-is otherwise.
	devMode := strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_ENV")), "development")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins, devMode)
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

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool, devMode bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		loopback := strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")
		if allowedOrigins[origin] || (devMode && loopback) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
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

// SafeLogValue strips line-break characters from user-controlled values so
// they cannot forge log entries (gosec G706).
func SafeLogValue(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, value)
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
	// Clamp into [0,1]: floating-point error can push h slightly past 1 near
	// antipodal points, which would make math.Sqrt(1-h) return NaN.
	h = math.Min(1, math.Max(0, h))
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

// ClosureIntersectsBBox returns true if any geometry segment passes through the box.
func ClosureIntersectsBBox(geometry models.LineStringGeometry, box models.BBox) bool {
	for i, point := range geometry.Coordinates {
		lng, lat := point[0], point[1]
		if lng >= box.MinLng && lng <= box.MaxLng && lat >= box.MinLat && lat <= box.MaxLat {
			return true
		}
		if i+1 < len(geometry.Coordinates) {
			next := geometry.Coordinates[i+1]
			if segmentIntersectsBox(lng, lat, next[0], next[1], box) {
				return true
			}
		}
	}
	return false
}

// segmentIntersectsBox reports whether the segment (x1,y1)-(x2,y2) passes
// through the box, using Liang-Barsky clipping so pass-through, touching, and
// fully-inside segments all count as intersections.
func segmentIntersectsBox(x1, y1, x2, y2 float64, box models.BBox) bool {
	t0, t1 := 0.0, 1.0
	dx, dy := x2-x1, y2-y1
	clips := [4][2]float64{
		{-dx, x1 - box.MinLng},
		{dx, box.MaxLng - x1},
		{-dy, y1 - box.MinLat},
		{dy, box.MaxLat - y1},
	}
	for _, clip := range clips {
		p, q := clip[0], clip[1]
		if p == 0 {
			if q < 0 {
				return false
			}
			continue
		}
		r := q / p
		if p < 0 {
			if r > t1 {
				return false
			}
			if r > t0 {
				t0 = r
			}
		} else {
			if r < t0 {
				return false
			}
			if r < t1 {
				t1 = r
			}
		}
	}
	return t0 <= t1
}

// MinDistanceToLineString returns the minimum distance from a location to the
// geometry, measuring to each segment rather than to vertices only.
func MinDistanceToLineString(location models.Coordinates, geometry models.LineStringGeometry) float64 {
	minDistance := math.MaxFloat64
	for i := 0; i+1 < len(geometry.Coordinates); i++ {
		d := distanceToSegmentMeters(location, toCoordinates(geometry.Coordinates[i]), toCoordinates(geometry.Coordinates[i+1]))
		if d < minDistance {
			minDistance = d
		}
	}
	if minDistance == math.MaxFloat64 {
		// Degenerate geometry with fewer than two points: measure to vertices.
		for _, point := range geometry.Coordinates {
			d := DistanceMeters(location, toCoordinates(point))
			if d < minDistance {
				minDistance = d
			}
		}
	}
	return minDistance
}

func toCoordinates(point []float64) models.Coordinates {
	return models.Coordinates{Lat: point[1], Lng: point[0]}
}

// distanceToSegmentMeters approximates the distance from p to segment a-b by
// projecting onto a local equirectangular plane centered at p, which is
// accurate for the short distances involved in closure searches.
func distanceToSegmentMeters(p, a, b models.Coordinates) float64 {
	latRad := degreesToRadians(p.Lat)
	ax := EarthRadiusMeters * degreesToRadians(a.Lng-p.Lng) * math.Cos(latRad)
	ay := EarthRadiusMeters * degreesToRadians(a.Lat-p.Lat)
	bx := EarthRadiusMeters * degreesToRadians(b.Lng-p.Lng) * math.Cos(latRad)
	by := EarthRadiusMeters * degreesToRadians(b.Lat-p.Lat)

	dx, dy := bx-ax, by-ay
	t := 0.0
	if lengthSq := dx*dx + dy*dy; lengthSq > 0 {
		// p maps to the plane origin, so project the origin onto the segment.
		t = math.Min(1, math.Max(0, -(ax*dx+ay*dy)/lengthSq))
	}
	return math.Hypot(ax+t*dx, ay+t*dy)
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
