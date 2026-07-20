package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
)

// AllowedHazards is the set of supported hazard types.
var AllowedHazards = map[string]bool{
	"flood":             true,
	"fire":              true,
	"road_crash":        true,
	"building_collapse": true,
	"medical_emergency": true,
	"security_incident": true,
	"disease_outbreak":  true,
	"electrical_hazard": true,
	"blocked_drain":     true,
	"landslide":         true,
	"marine_accident":   true,
	"storm":             true,
	"tidal_wave":        true,
	"other":             true,
}

// AllowedSeverities is the set of supported alert severities.
var AllowedSeverities = map[string]bool{
	"advisory":       true,
	"watch":          true,
	"warning":        true,
	"severe_warning": true,
	"emergency":      true,
}

// AllowedRiskLevels is the set of supported ML risk levels.
var AllowedRiskLevels = map[string]bool{
	"low":       true,
	"moderate":  true,
	"high":      true,
	"severe":    true,
	"emergency": true,
}

// AllowedTargetTypes is the set of supported alert target types.
var AllowedTargetTypes = map[string]bool{
	"national":  true,
	"region":    true,
	"district":  true,
	"radius":    true,
	"community": true,
	"custom":    true,
}

// DraftRoles are roles allowed to create and edit draft alerts.
var DraftRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

// ApprovalRoles are roles allowed to approve or reject submitted alerts.
var ApprovalRoles = map[string]bool{
	"system_admin":  true,
	"agency_admin":  true,
	"nadmo_officer": true,
}

// OverrideRoles are roles allowed to perform an emergency override.
var OverrideRoles = map[string]bool{
	"system_admin":  true,
	"nadmo_officer": true,
}

// TargetCatalog maps known target identifiers to their metadata.
var TargetCatalog = map[string]models.TargetCatalogRecord{
	"region:greater-accra": {
		ID:                  "greater-accra",
		Type:                "region",
		Label:               "Greater Accra Region",
		Center:              models.Coordinates{Lat: 5.75, Lng: -0.11},
		RadiusMeters:        52000,
		AreaSqKm:            3245,
		EstimatedPopulation: 5455000,
	},
	"district:accra-metropolitan": {
		ID:                  "accra-metropolitan",
		Type:                "district",
		Label:               "Accra Metropolitan",
		Center:              models.Coordinates{Lat: 5.56, Lng: -0.2},
		RadiusMeters:        9000,
		AreaSqKm:            61,
		EstimatedPopulation: 284000,
	},
	"district:tema-metropolitan": {
		ID:                  "tema-metropolitan",
		Type:                "district",
		Label:               "Tema Metropolitan",
		Center:              models.Coordinates{Lat: 5.642, Lng: -0.028},
		RadiusMeters:        12000,
		AreaSqKm:            565,
		EstimatedPopulation: 402000,
	},
	"district:ablekuma-west": {
		ID:                  "ablekuma-west",
		Type:                "district",
		Label:               "Ablekuma West",
		Center:              models.Coordinates{Lat: 5.601, Lng: -0.286},
		RadiusMeters:        7000,
		AreaSqKm:            15,
		EstimatedPopulation: 220000,
	},
	"community:accra-central": {
		ID:                  "accra-central",
		Type:                "community",
		Label:               "Accra Central",
		Center:              models.Coordinates{Lat: 5.556, Lng: -0.202},
		RadiusMeters:        3000,
		AreaSqKm:            8,
		EstimatedPopulation: 75000,
	},
}

// DecodeJSON decodes a JSON request body into target.
func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// OptionalDecodeJSON decodes a JSON request body when one is present.
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

// isDevelopmentEnv reports whether the service runs in a development
// environment; only then may CORS echo localhost/127.0.0.1 origins outside the
// configured allowlist.
func isDevelopmentEnv() bool {
	return strings.TrimSpace(os.Getenv("NADAA_ENV")) == "development"
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

// NormalizeAlertRequest applies normalization rules to an alert request.
func NormalizeAlertRequest(request models.CreateAlertRequest) models.CreateAlertRequest {
	request.Title = strings.TrimSpace(request.Title)
	request.HazardType = NormalizeQueryValue(request.HazardType)
	request.Severity = NormalizeQueryValue(request.Severity)
	request.Message = strings.TrimSpace(request.Message)
	request.Target = NormalizeTarget(request.Target)
	request.RecommendedAction = strings.TrimSpace(request.RecommendedAction)
	request.ShelterIDs = CompactStrings(request.ShelterIDs)
	request.SourcePrediction = NormalizeSourcePrediction(request.SourcePrediction)
	return request
}

// ValidateAlertRequest validates a normalized create/update alert request.
func ValidateAlertRequest(request models.CreateAlertRequest) (string, string) {
	title := strings.TrimSpace(request.Title)
	message := strings.TrimSpace(request.Message)
	action := strings.TrimSpace(request.RecommendedAction)
	hazard := NormalizeQueryValue(request.HazardType)
	severity := NormalizeQueryValue(request.Severity)

	if len(title) < 4 || len(title) > 140 {
		return "invalid_title", "title must be 4 to 140 characters"
	}
	if !AllowedHazards[hazard] {
		return "invalid_hazard", "hazardType must be a supported NADAA hazard type"
	}
	if !AllowedSeverities[severity] {
		return "invalid_severity", "severity must be advisory, watch, warning, severe_warning, or emergency"
	}
	if len(message) < 10 || len(message) > 1000 {
		return "invalid_message", "message must be 10 to 1000 characters"
	}
	if code, message := ValidateTarget(request.Target); code != "" {
		return code, message
	}
	if request.StartsAt.IsZero() {
		return "missing_starts_at", "startsAt is required"
	}
	if request.ExpiresAt.IsZero() || !request.ExpiresAt.After(request.StartsAt) {
		return "invalid_expiry", "expiresAt must be after startsAt"
	}
	if action == "" {
		return "missing_recommended_action", "recommendedAction is required"
	}
	if code, message := ValidateSourcePrediction(request.SourcePrediction); code != "" {
		return code, message
	}

	return "", ""
}

// NormalizeSourcePrediction applies normalization rules to a source prediction.
func NormalizeSourcePrediction(source *models.AlertSourcePrediction) *models.AlertSourcePrediction {
	if source == nil {
		return nil
	}
	return &models.AlertSourcePrediction{
		PredictionID:           strings.TrimSpace(source.PredictionID),
		PredictionLogID:        strings.TrimSpace(source.PredictionLogID),
		ModelVersion:           strings.TrimSpace(source.ModelVersion),
		InputFeatureSetVersion: strings.TrimSpace(source.InputFeatureSetVersion),
		Probability:            RoundProbability(source.Probability),
		Severity:               NormalizeQueryValue(source.Severity),
		Confidence:             NormalizeQueryValue(source.Confidence),
		HumanReviewRequired:    source.HumanReviewRequired,
		AutoPublishAllowed:     source.AutoPublishAllowed,
		ReviewNote:             strings.TrimSpace(source.ReviewNote),
	}
}

// ValidateSourcePrediction validates a source prediction safety rules.
func ValidateSourcePrediction(source *models.AlertSourcePrediction) (string, string) {
	if source == nil {
		return "", ""
	}
	if source.PredictionID == "" {
		return "missing_prediction_id", "sourcePrediction.predictionId is required"
	}
	if source.ModelVersion == "" || source.InputFeatureSetVersion == "" {
		return "missing_prediction_model", "sourcePrediction model and feature set versions are required"
	}
	if source.Probability < 0 || source.Probability > 1 {
		return "invalid_prediction_probability", "sourcePrediction.probability must be between 0 and 1"
	}
	if !source.HumanReviewRequired || source.AutoPublishAllowed {
		return "invalid_prediction_safety", "sourcePrediction must require human review and disallow auto-publish"
	}
	if !AllowedRiskLevels[source.Severity] {
		return "invalid_prediction_severity", "sourcePrediction.severity must be a supported risk level"
	}
	if source.Confidence != "low" && source.Confidence != "medium" && source.Confidence != "high" {
		return "invalid_prediction_confidence", "sourcePrediction.confidence must be low, medium, or high"
	}
	if len(source.ReviewNote) > 400 {
		return "invalid_prediction_review_note", "sourcePrediction.reviewNote must be 400 characters or fewer"
	}
	return "", ""
}

// NormalizeTarget applies normalization rules to an alert target.
func NormalizeTarget(target models.AlertTarget) models.AlertTarget {
	normalized := models.AlertTarget{
		Type:                NormalizeQueryValue(target.Type),
		IDs:                 NormalizeTargetIDs(target.IDs),
		Label:               strings.TrimSpace(target.Label),
		Center:              NormalizeCenter(target.Center),
		RadiusMeters:        RoundMeters(target.RadiusMeters),
		Geometry:            NormalizeGeometry(target.Geometry),
		AreaSqKm:            RoundArea(target.AreaSqKm),
		EstimatedPopulation: target.EstimatedPopulation,
	}

	switch normalized.Type {
	case "national":
		normalized.IDs = []string{"ghana"}
		if normalized.Label == "" {
			normalized.Label = "Ghana"
		}
		normalized.Center = &models.Coordinates{Lat: 7.9465, Lng: -1.0232}
		normalized.RadiusMeters = 365000
		normalized.AreaSqKm = 238533
		normalized.EstimatedPopulation = 33480000
		normalized.Geometry = GeometryFromBounds(4.54, -3.26, 11.18, 1.2)
	case "region", "district", "community":
		ApplyCatalogTarget(&normalized)
	case "radius":
		if normalized.IDs == nil {
			normalized.IDs = []string{"radius"}
		}
		if normalized.Label == "" {
			normalized.Label = "Radius target"
		}
		if normalized.RadiusMeters > 0 {
			normalized.AreaSqKm = RoundArea(math.Pi * (normalized.RadiusMeters / 1000) * (normalized.RadiusMeters / 1000))
		}
		if normalized.AreaSqKm > 0 && normalized.EstimatedPopulation == 0 {
			normalized.EstimatedPopulation = int(math.Round(normalized.AreaSqKm * 4500))
		}
	case "custom":
		if normalized.IDs == nil {
			normalized.IDs = []string{"custom"}
		}
		if normalized.Label == "" {
			normalized.Label = "Custom geometry"
		}
		if normalized.Geometry != nil {
			normalized.Center = PolygonCenter(normalized.Geometry)
			normalized.AreaSqKm = PolygonAreaSqKm(normalized.Geometry)
			if normalized.AreaSqKm > 0 && normalized.EstimatedPopulation == 0 {
				normalized.EstimatedPopulation = int(math.Round(normalized.AreaSqKm * 5000))
			}
		}
	}

	return normalized
}

// ValidateTarget validates an alert target.
func ValidateTarget(target models.AlertTarget) (string, string) {
	if !AllowedTargetTypes[target.Type] {
		return "invalid_target", "target.type must be national, region, district, radius, community, or custom"
	}
	if target.Type != "national" && len(target.IDs) == 0 {
		return "missing_target_ids", "target.ids are required unless target.type is national"
	}
	if strings.TrimSpace(target.Label) == "" {
		return "missing_target_label", "target.label is required"
	}
	switch target.Type {
	case "region", "district", "community":
		for _, id := range target.IDs {
			if _, ok := TargetCatalog[target.Type+":"+id]; !ok {
				return "unknown_target_id", "target.ids must match a supported region, district, or community"
			}
		}
		if target.Geometry == nil || target.Center == nil {
			return "missing_target_geometry", "target geometry could not be resolved"
		}
	case "radius":
		if target.Center == nil || !ValidCoordinates(*target.Center) {
			return "invalid_target_center", "radius targets require a valid center"
		}
		if target.RadiusMeters < 250 || target.RadiusMeters > 100000 {
			return "invalid_target_radius", "radiusMeters must be between 250 and 100000"
		}
	case "custom":
		if target.Geometry == nil || !ValidPolygonGeometry(*target.Geometry) {
			return "invalid_target_geometry", "custom targets require a closed polygon geometry"
		}
		if target.AreaSqKm <= 0 || target.AreaSqKm > 50000 {
			return "invalid_target_area", "custom target area must be greater than 0 and at most 50000 square kilometers"
		}
	}
	return "", ""
}

// ApplyCatalogTarget enriches a region/district/community target from the catalog.
func ApplyCatalogTarget(target *models.AlertTarget) {
	records := make([]models.TargetCatalogRecord, 0, len(target.IDs))
	for _, id := range target.IDs {
		record, ok := TargetCatalog[target.Type+":"+id]
		if !ok {
			continue
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return
	}

	if target.Label == "" {
		labels := make([]string, 0, len(records))
		for _, record := range records {
			labels = append(labels, record.Label)
		}
		target.Label = strings.Join(labels, ", ")
	}

	lat := 0.0
	lng := 0.0
	area := 0.0
	population := 0
	for _, record := range records {
		lat += record.Center.Lat
		lng += record.Center.Lng
		area += record.AreaSqKm
		population += record.EstimatedPopulation
	}
	target.Center = &models.Coordinates{Lat: RoundCoordinate(lat / float64(len(records))), Lng: RoundCoordinate(lng / float64(len(records)))}
	target.RadiusMeters = MaxCatalogRadius(records)
	target.AreaSqKm = RoundArea(area)
	target.EstimatedPopulation = population
	target.Geometry = GeometryFromCatalogRecords(records)
}

// NormalizeCenter normalizes a coordinates pointer.
func NormalizeCenter(center *models.Coordinates) *models.Coordinates {
	if center == nil {
		return nil
	}
	return &models.Coordinates{Lat: RoundCoordinate(center.Lat), Lng: RoundCoordinate(center.Lng)}
}

// NormalizeGeometry normalizes a geometry pointer.
func NormalizeGeometry(geometry *models.TargetGeometry) *models.TargetGeometry {
	if geometry == nil {
		return nil
	}
	return &models.TargetGeometry{
		Type:        strings.TrimSpace(geometry.Type),
		Coordinates: geometry.Coordinates,
	}
}

// NormalizeTargetIDs normalizes, compacts, and dedupes target ids.
func NormalizeTargetIDs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = NormalizeQueryValue(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// GeometryFromCatalogRecords builds a bounding geometry from catalog records.
func GeometryFromCatalogRecords(records []models.TargetCatalogRecord) *models.TargetGeometry {
	if len(records) == 0 {
		return nil
	}

	minLat := 90.0
	minLng := 180.0
	maxLat := -90.0
	maxLng := -180.0
	for _, record := range records {
		deltaLat, deltaLng := DegreeDeltas(record.Center, record.RadiusMeters)
		minLat = math.Min(minLat, record.Center.Lat-deltaLat)
		maxLat = math.Max(maxLat, record.Center.Lat+deltaLat)
		minLng = math.Min(minLng, record.Center.Lng-deltaLng)
		maxLng = math.Max(maxLng, record.Center.Lng+deltaLng)
	}
	return GeometryFromBounds(minLat, minLng, maxLat, maxLng)
}

// GeometryFromBounds builds a polygon geometry from bounding coordinates.
func GeometryFromBounds(minLat float64, minLng float64, maxLat float64, maxLng float64) *models.TargetGeometry {
	return &models.TargetGeometry{
		Type: "Polygon",
		Coordinates: [][][]float64{{
			{RoundCoordinate(minLng), RoundCoordinate(minLat)},
			{RoundCoordinate(maxLng), RoundCoordinate(minLat)},
			{RoundCoordinate(maxLng), RoundCoordinate(maxLat)},
			{RoundCoordinate(minLng), RoundCoordinate(maxLat)},
			{RoundCoordinate(minLng), RoundCoordinate(minLat)},
		}},
	}
}

// MaxCatalogRadius returns the largest radius among catalog records.
func MaxCatalogRadius(records []models.TargetCatalogRecord) float64 {
	maxRadius := 0.0
	for _, record := range records {
		if record.RadiusMeters > maxRadius {
			maxRadius = record.RadiusMeters
		}
	}
	return RoundMeters(maxRadius)
}

// DegreeDeltas converts a radius in meters to approximate latitude/longitude deltas.
func DegreeDeltas(center models.Coordinates, radiusMeters float64) (float64, float64) {
	latDelta := radiusMeters / 111320
	lngDelta := radiusMeters / (111320 * math.Cos(center.Lat*math.Pi/180))
	if math.IsInf(lngDelta, 0) || math.IsNaN(lngDelta) {
		lngDelta = latDelta
	}
	return latDelta, lngDelta
}

// ValidCoordinates reports whether coordinates are within valid ranges.
func ValidCoordinates(center models.Coordinates) bool {
	return center.Lat >= -90 && center.Lat <= 90 && center.Lng >= -180 && center.Lng <= 180
}

// ValidPolygonGeometry reports whether a geometry is a closed polygon.
func ValidPolygonGeometry(geometry models.TargetGeometry) bool {
	if geometry.Type != "Polygon" || len(geometry.Coordinates) != 1 {
		return false
	}
	ring := geometry.Coordinates[0]
	if len(ring) < 4 {
		return false
	}
	first := ring[0]
	last := ring[len(ring)-1]
	if len(first) != 2 || len(last) != 2 || first[0] != last[0] || first[1] != last[1] {
		return false
	}
	for _, point := range ring {
		if len(point) != 2 {
			return false
		}
		if !ValidCoordinates(models.Coordinates{Lat: point[1], Lng: point[0]}) {
			return false
		}
	}
	return true
}

// PolygonCenter approximates the center of a polygon ring.
func PolygonCenter(geometry *models.TargetGeometry) *models.Coordinates {
	if geometry == nil || len(geometry.Coordinates) == 0 || len(geometry.Coordinates[0]) == 0 {
		return nil
	}
	ring := geometry.Coordinates[0]
	lat := 0.0
	lng := 0.0
	count := 0
	for index, point := range ring {
		if index == len(ring)-1 {
			continue
		}
		if len(point) != 2 {
			return nil
		}
		lat += point[1]
		lng += point[0]
		count++
	}
	if count == 0 {
		return nil
	}
	return &models.Coordinates{Lat: RoundCoordinate(lat / float64(count)), Lng: RoundCoordinate(lng / float64(count))}
}

// PolygonAreaSqKm approximates the area of a polygon in square kilometers.
func PolygonAreaSqKm(geometry *models.TargetGeometry) float64 {
	if geometry == nil || len(geometry.Coordinates) == 0 || len(geometry.Coordinates[0]) < 4 {
		return 0
	}
	center := PolygonCenter(geometry)
	if center == nil {
		return 0
	}

	ring := geometry.Coordinates[0]
	sum := 0.0
	for index := range len(ring) - 1 {
		if len(ring[index]) != 2 || len(ring[index+1]) != 2 {
			return 0
		}
		x1, y1 := LonLatToMeters(ring[index][0], ring[index][1], center.Lat)
		x2, y2 := LonLatToMeters(ring[index+1][0], ring[index+1][1], center.Lat)
		sum += x1*y2 - x2*y1
	}
	return RoundArea(math.Abs(sum) / 2 / 1000000)
}

// LonLatToMeters converts longitude/latitude to local meter offsets.
func LonLatToMeters(lng float64, lat float64, referenceLat float64) (float64, float64) {
	x := lng * 111320 * math.Cos(referenceLat*math.Pi/180)
	y := lat * 110540
	return x, y
}

// TargetSummary returns a human-readable summary of a target.
func TargetSummary(target models.AlertTarget) string {
	switch target.Type {
	case "radius":
		return fmt.Sprintf("%s radius target, approximately %.1f sq km and %d people.", MetersLabel(target.RadiusMeters), target.AreaSqKm, target.EstimatedPopulation)
	case "custom":
		return fmt.Sprintf("Custom polygon target, approximately %.1f sq km and %d people.", target.AreaSqKm, target.EstimatedPopulation)
	default:
		return fmt.Sprintf("%s target covering approximately %.1f sq km and %d people.", target.Label, target.AreaSqKm, target.EstimatedPopulation)
	}
}

// TargetWarnings returns warnings for a target.
func TargetWarnings(target models.AlertTarget) []string {
	warnings := []string{}
	if target.Type == "national" {
		warnings = append(warnings, "National alerts should be reserved for broad life-safety threats.")
	}
	if target.AreaSqKm > 1000 {
		warnings = append(warnings, "Large target area may increase alert fatigue; confirm scope before approval.")
	}
	if target.Type == "custom" {
		warnings = append(warnings, "Custom geometry should be reviewed against official district boundaries before publishing.")
	}
	return warnings
}

// MetersLabel formats a meter value as meters or kilometers.
func MetersLabel(value float64) string {
	if value >= 1000 {
		return fmt.Sprintf("%.1f km", value/1000)
	}
	return fmt.Sprintf("%.0f m", value)
}

// RoundMeters rounds meters to a whole number.
func RoundMeters(value float64) float64 {
	return math.Round(value)
}

// RoundCoordinate rounds a coordinate to six decimal places.
func RoundCoordinate(value float64) float64 {
	return math.Round(value*1000000) / 1000000
}

// RoundArea rounds an area value to one decimal place.
func RoundArea(value float64) float64 {
	return math.Round(value*10) / 10
}

// RoundProbability rounds a probability to four decimal places.
func RoundProbability(value float64) float64 {
	return math.Round(value*10000) / 10000
}

// ContainsString reports whether values contains needle.
func ContainsString(values []string, needle string) bool {
	return slices.Contains(values, needle)
}

// CompactStrings trims and removes empty strings from values.
func CompactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

// SnapshotAlert creates a map snapshot of an alert for audit logging. It must
// carry every citizen-facing field so the audit trail can reconstruct exactly
// what citizens were sent.
func SnapshotAlert(alert models.AuthorityAlert) map[string]any {
	return map[string]any{
		"id":                 alert.ID,
		"title":              alert.Title,
		"hazardType":         alert.HazardType,
		"severity":           alert.Severity,
		"message":            alert.Message,
		"target":             alert.Target,
		"startsAt":           alert.StartsAt,
		"expiresAt":          alert.ExpiresAt,
		"recommendedAction":  alert.RecommendedAction,
		"evacuationRequired": alert.EvacuationRequired,
		"shelterIds":         alert.ShelterIDs,
		"issuingAgencyId":    alert.IssuingAgencyID,
		"issuedBy":           alert.IssuedBy,
		"approvedBy":         alert.ApprovedBy,
		"rejectedBy":         alert.RejectedBy,
		"status":             alert.Status,
		"emergencyOverride":  alert.EmergencyOverride,
		"statusReason":       alert.StatusReason,
		"sourcePrediction":   alert.SourcePrediction,
	}
}

// ValidAlertStatus reports whether status is a recognized alert status.
func ValidAlertStatus(status string) bool {
	switch status {
	case "draft", "submitted", "approved", "rejected", "published", "expired", "cancelled":
		return true
	default:
		return false
	}
}

// StatusForCode maps an error code to an HTTP status code.
func StatusForCode(code string) int {
	switch code {
	case "not_found":
		return http.StatusNotFound
	case "forbidden", "separation_of_duties":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

// TimePtr returns a pointer to a copy of value.
func TimePtr(value time.Time) *time.Time {
	return &value
}
