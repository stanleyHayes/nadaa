package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
)

const (
	// EarthRadiusMeters is the approximate radius of the Earth in meters.
	EarthRadiusMeters = 6371000.0
	// NearbySearchMeters is the default radius for nearby shelter searches.
	NearbySearchMeters = 30000.0
	// DefaultNearbyLimit is the default number of nearby results returned.
	DefaultNearbyLimit = 6
	// HospitalCapacityStaleAfter is the duration after which hospital capacity is considered stale.
	HospitalCapacityStaleAfter = 30 * time.Minute
)

// HospitalCapacityStaleAfterMinutes is the staleness threshold in minutes.
var HospitalCapacityStaleAfterMinutes = int(HospitalCapacityStaleAfter / time.Minute)

// ShelterUpdateRoles lists roles allowed to update shelter capacity.
var ShelterUpdateRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

// AllowedShelterStatuses lists supported shelter statuses.
var AllowedShelterStatuses = map[string]bool{
	"open":    true,
	"full":    true,
	"closed":  true,
	"unknown": true,
}

// AllowedEmergencyCapacityStatuses lists supported emergency capacity values.
var AllowedEmergencyCapacityStatuses = map[string]bool{
	"available": true,
	"limited":   true,
	"full":      true,
	"offline":   true,
	"unknown":   true,
}

// AllowedEmergencyUnitStatuses lists supported emergency unit statuses.
var AllowedEmergencyUnitStatuses = map[string]bool{
	"open":    true,
	"busy":    true,
	"divert":  true,
	"closed":  true,
	"unknown": true,
}

// AllowedReliefPointStatuses lists supported relief point statuses.
var AllowedReliefPointStatuses = map[string]bool{
	"open":    true,
	"limited": true,
	"closed":  true,
	"paused":  true,
}

// AllowedReliefPointTypes lists supported relief point types.
var AllowedReliefPointTypes = map[string]bool{
	"food":     true,
	"water":    true,
	"medical":  true,
	"hygiene":  true,
	"blankets": true,
	"cash":     true,
	"mixed":    true,
}

// AllowedAidRequestCategories lists supported aid request categories.
var AllowedAidRequestCategories = map[string]bool{
	"food":       true,
	"water":      true,
	"medical":    true,
	"hygiene":    true,
	"shelter":    true,
	"logistics":  true,
	"cash":       true,
	"equipment":  true,
	"volunteers": true,
	"other":      true,
}

// AllowedAidRequestPriorities lists supported aid priorities.
var AllowedAidRequestPriorities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"urgent": true,
}

// AllowedAidRequestStatuses lists supported aid request statuses.
var AllowedAidRequestStatuses = map[string]bool{
	"pending_review":    true,
	"approved":          true,
	"open":              true,
	"partially_matched": true,
	"fulfilled":         true,
	"paused":            true,
	"closed":            true,
	"rejected":          true,
}

// AllowedAidRequestReviewStatuses lists statuses an aid request can be reviewed into.
var AllowedAidRequestReviewStatuses = map[string]bool{
	"approved": true,
	"open":     true,
	"paused":   true,
	"closed":   true,
	"rejected": true,
}

// PublicAidRequestStatuses are statuses visible to the public.
var PublicAidRequestStatuses = map[string]bool{
	"approved":          true,
	"open":              true,
	"partially_matched": true,
}

// AllowedAidRequestVisibility lists supported visibility values.
var AllowedAidRequestVisibility = map[string]bool{
	"public":        true,
	"partners_only": true,
}

// AllowedAidDonorTypes lists supported donor types.
var AllowedAidDonorTypes = map[string]bool{
	"individual":  true,
	"business":    true,
	"ngo":         true,
	"faith_group": true,
	"diaspora":    true,
	"government":  true,
	"other":       true,
}

// AllowedAidPledgeStatuses lists supported pledge statuses.
var AllowedAidPledgeStatuses = map[string]bool{
	"pledged":   true,
	"accepted":  true,
	"received":  true,
	"cancelled": true,
	"flagged":   true,
}

// AllowedAidPledgeReviewStatuses lists supported pledge review statuses.
var AllowedAidPledgeReviewStatuses = map[string]bool{
	"pending_review": true,
	"cleared":        true,
	"flagged":        true,
}

// DecodeJSON decodes a JSON request body into target.
func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// OptionalDecodeJSON decodes a JSON body when one is present.
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
		log.Printf("ERROR shelter-service write_json_response_failed error=%v", err)
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
		if allowedOrigins[origin] || strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

// AllowedOriginsFromEnv parses the NADAA_ALLOWED_ORIGINS environment variable.
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
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

// NormalizeQueryValue trims and lowercases a query value.
func NormalizeQueryValue(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

// NormalizeToken normalizes a string token by lowercasing and replacing spaces/dashes with underscores.
func NormalizeToken(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(value)), "-", "_"), " ", "_")
}

// UnsafeText checks for obvious unsafe content patterns.
func UnsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

// ValidCoordinates checks whether coordinates are within valid ranges.
func ValidCoordinates(location models.Coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

// DistanceMeters returns the great-circle distance between two coordinates in meters.
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

// ContainsNormalized checks whether needle is present in values after normalization.
func ContainsNormalized(values []string, needle string) bool {
	needle = NormalizeToken(needle)
	for _, value := range values {
		if NormalizeToken(value) == needle {
			return true
		}
	}
	return false
}

// BoolPtr returns a pointer to a bool value.
func BoolPtr(value bool) *bool {
	return &value
}

// ParseLocation parses lat/lng query parameters and validates them.
func ParseLocation(w http.ResponseWriter, r *http.Request) (models.Coordinates, bool) {
	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" || lngText == "" {
		WriteError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng query parameters are required")
		return models.Coordinates{}, false
	}

	lat, latErr := strconv.ParseFloat(latText, 64)
	lng, lngErr := strconv.ParseFloat(lngText, 64)
	if latErr != nil || lngErr != nil {
		WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
		return models.Coordinates{}, false
	}

	location := models.Coordinates{Lat: lat, Lng: lng}
	if !ValidCoordinates(location) {
		WriteError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
		return models.Coordinates{}, false
	}
	return location, true
}

// ParseBBox parses a comma-separated bounding box string.
func ParseBBox(value string) (*models.BoundingBox, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, true
	}
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return nil, false
	}
	minLng, err1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	minLat, err2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	maxLng, err3 := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	maxLat, err4 := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, false
	}
	return &models.BoundingBox{MinLat: minLat, MinLng: minLng, MaxLat: maxLat, MaxLng: maxLng}, true
}

// StatusForCode maps error codes to HTTP status codes.
func StatusForCode(code string) int {
	if code == "not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// NormalizeOccupancyUpdate validates an occupancy update request.
func NormalizeOccupancyUpdate(request models.OccupancyUpdateRequest) (models.OccupancyUpdateRequest, string, string) {
	request.Status = NormalizeToken(request.Status)
	request.Notes = strings.TrimSpace(request.Notes)

	if request.Capacity == nil && request.CurrentOccupancy == nil && request.Status == "" && request.Notes == "" {
		return request, "no_changes", "at least one occupancy field must be supplied"
	}
	if request.Capacity != nil && *request.Capacity < 0 {
		return request, "invalid_capacity", "capacity must be zero or greater"
	}
	if request.CurrentOccupancy != nil && *request.CurrentOccupancy < 0 {
		return request, "invalid_occupancy", "currentOccupancy must be zero or greater"
	}
	if request.Capacity != nil && request.CurrentOccupancy != nil && *request.CurrentOccupancy > *request.Capacity {
		return request, "invalid_occupancy", "currentOccupancy cannot exceed capacity"
	}
	if request.Status != "" && !AllowedShelterStatuses[request.Status] {
		return request, "invalid_status", "status must be open, full, closed, or unknown"
	}
	if len(request.Notes) > 500 || UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 500 safe characters or fewer"
	}
	return request, "", ""
}

// NormalizeHospitalCapacityUpdate validates a hospital capacity update request.
func NormalizeHospitalCapacityUpdate(request models.HospitalCapacityUpdateRequest) (models.HospitalCapacityUpdateRequest, string, string) {
	request.EmergencyCapacity = NormalizeToken(request.EmergencyCapacity)
	request.EmergencyUnitStatus = NormalizeToken(request.EmergencyUnitStatus)
	request.Notes = strings.TrimSpace(request.Notes)
	request.Source = NormalizeToken(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)

	if request.TotalBeds == nil &&
		request.AvailableBeds == nil &&
		request.ICUBedsAvailable == nil &&
		request.MaternityBedsAvailable == nil &&
		request.PediatricBedsAvailable == nil &&
		request.IsolationBedsAvailable == nil &&
		request.EmergencyCapacity == "" &&
		request.EmergencyUnitStatus == "" &&
		request.AmbulancesAvailable == nil &&
		request.OxygenAvailable == nil &&
		request.Notes == "" {
		return request, "no_changes", "at least one hospital capacity field must be supplied"
	}
	for _, item := range []struct {
		name  string
		value *int
	}{
		{"totalBeds", request.TotalBeds},
		{"availableBeds", request.AvailableBeds},
		{"icuBedsAvailable", request.ICUBedsAvailable},
		{"maternityBedsAvailable", request.MaternityBedsAvailable},
		{"pediatricBedsAvailable", request.PediatricBedsAvailable},
		{"isolationBedsAvailable", request.IsolationBedsAvailable},
		{"ambulancesAvailable", request.AmbulancesAvailable},
	} {
		if item.value != nil && *item.value < 0 {
			return request, "invalid_" + NormalizeToken(item.name), item.name + " must be zero or greater"
		}
	}
	if request.TotalBeds != nil && request.AvailableBeds != nil && *request.AvailableBeds > *request.TotalBeds {
		return request, "invalid_available_beds", "availableBeds cannot exceed totalBeds"
	}
	if request.EmergencyCapacity != "" && !AllowedEmergencyCapacityStatuses[request.EmergencyCapacity] {
		return request, "invalid_emergency_capacity", "emergencyCapacity must be available, limited, full, offline, or unknown"
	}
	if request.EmergencyUnitStatus != "" && !AllowedEmergencyUnitStatuses[request.EmergencyUnitStatus] {
		return request, "invalid_emergency_unit_status", "emergencyUnitStatus must be open, busy, divert, closed, or unknown"
	}
	if len(request.Notes) > 700 || UnsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 700 safe characters or fewer"
	}
	if request.Source == "" {
		request.Source = "manual"
	}
	if len(request.Source) > 80 || UnsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || UnsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	return request, "", ""
}

// NormalizeHospitalCapacityImport validates a hospital capacity import request.
func NormalizeHospitalCapacityImport(request models.HospitalCapacityImportRequest) (models.HospitalCapacityImportRequest, string, string) {
	request.Source = NormalizeToken(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	if request.Source == "" {
		request.Source = "fixture_adapter"
	}
	if request.SourceRef == "" {
		request.SourceRef = "hospital-capacity-feed"
	}
	if len(request.Source) > 80 || UnsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || UnsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	for index := range request.Records {
		record := &request.Records[index]
		record.FacilityID = strings.TrimSpace(record.FacilityID)
		record.EmergencyCapacity = NormalizeToken(record.EmergencyCapacity)
		record.EmergencyUnitStatus = NormalizeToken(record.EmergencyUnitStatus)
		record.Notes = strings.TrimSpace(record.Notes)
		if record.FacilityID == "" || len(record.FacilityID) > 128 || UnsafeText(record.FacilityID) {
			return request, "invalid_facility_id", "facilityId is required and must be safe"
		}
		if record.AvailableBeds < 0 ||
			record.ICUBedsAvailable < 0 ||
			record.MaternityBedsAvailable < 0 ||
			record.PediatricBedsAvailable < 0 ||
			record.IsolationBedsAvailable < 0 ||
			record.AmbulancesAvailable < 0 {
			return request, "invalid_capacity_value", "capacity values must be zero or greater"
		}
		if !AllowedEmergencyCapacityStatuses[record.EmergencyCapacity] {
			return request, "invalid_emergency_capacity", "emergencyCapacity must be available, limited, full, offline, or unknown"
		}
		if record.EmergencyUnitStatus != "" && !AllowedEmergencyUnitStatuses[record.EmergencyUnitStatus] {
			return request, "invalid_emergency_unit_status", "emergencyUnitStatus must be open, busy, divert, closed, or unknown"
		}
		if len(record.Notes) > 700 || UnsafeText(record.Notes) {
			return request, "invalid_notes", "notes must be 700 safe characters or fewer"
		}
	}
	return request, "", ""
}

// DefaultHospitalCapacityFixture returns the default fixture records for import.
func DefaultHospitalCapacityFixture() []models.HospitalCapacityFixtureRecord {
	return []models.HospitalCapacityFixtureRecord{
		{
			FacilityID:             "hospital_001",
			AvailableBeds:          38,
			ICUBedsAvailable:       3,
			MaternityBedsAvailable: 8,
			PediatricBedsAvailable: 4,
			IsolationBedsAvailable: 2,
			EmergencyCapacity:      "available",
			EmergencyUnitStatus:    "open",
			AmbulancesAvailable:    2,
			OxygenAvailable:        BoolPtr(true),
			Notes:                  "Fixture adapter update from hospital-capacity-feed.",
		},
		{
			FacilityID:             "hospital_002",
			AvailableBeds:          9,
			ICUBedsAvailable:       1,
			MaternityBedsAvailable: 2,
			PediatricBedsAvailable: 1,
			IsolationBedsAvailable: 0,
			EmergencyCapacity:      "limited",
			EmergencyUnitStatus:    "busy",
			AmbulancesAvailable:    1,
			OxygenAvailable:        BoolPtr(true),
			Notes:                  "Fixture adapter reports heavy emergency load.",
		},
	}
}

// NormalizeCreateReliefPoint validates a create relief point request.
func NormalizeCreateReliefPoint(request models.CreateReliefPointRequest) (models.CreateReliefPointRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = NormalizeToken(request.Type)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Address = strings.TrimSpace(request.Address)
	request.Contact = strings.TrimSpace(request.Contact)
	request.OperatingHours = strings.TrimSpace(request.OperatingHours)
	request.Eligibility = strings.TrimSpace(request.Eligibility)
	request.Schedule = strings.TrimSpace(request.Schedule)
	request.Status = NormalizeToken(request.Status)
	request.Source = NormalizeToken(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)

	if request.Name == "" || len(request.Name) > 200 || UnsafeText(request.Name) {
		return request, "invalid_name", "name is required and must be 200 safe characters or fewer"
	}
	if !AllowedReliefPointTypes[request.Type] {
		return request, "invalid_type", "type must be food, water, medical, hygiene, blankets, cash, or mixed"
	}
	if !ValidCoordinates(request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	if request.Status == "" {
		request.Status = "open"
	}
	if !AllowedReliefPointStatuses[request.Status] {
		return request, "invalid_status", "status must be open, limited, closed, or paused"
	}
	if len(request.Region) > 100 || UnsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || UnsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if len(request.Address) > 300 || UnsafeText(request.Address) {
		return request, "invalid_address", "address must be 300 safe characters or fewer"
	}
	if len(request.Contact) > 100 || UnsafeText(request.Contact) {
		return request, "invalid_contact", "contact must be 100 safe characters or fewer"
	}
	if len(request.OperatingHours) > 100 || UnsafeText(request.OperatingHours) {
		return request, "invalid_operating_hours", "operatingHours must be 100 safe characters or fewer"
	}
	if len(request.Eligibility) > 700 || UnsafeText(request.Eligibility) {
		return request, "invalid_eligibility", "eligibility must be 700 safe characters or fewer"
	}
	if len(request.Schedule) > 200 || UnsafeText(request.Schedule) {
		return request, "invalid_schedule", "schedule must be 200 safe characters or fewer"
	}
	if request.Source == "" {
		request.Source = "manual"
	}
	if len(request.Source) > 80 || UnsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || UnsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	request.StockCategories = NormalizeStockCategories(request.StockCategories)
	return request, "", ""
}

// NormalizeUpdateReliefPoint validates an update relief point request.
func NormalizeUpdateReliefPoint(request models.UpdateReliefPointRequest) (models.UpdateReliefPointRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = NormalizeToken(request.Type)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Address = strings.TrimSpace(request.Address)
	request.Contact = strings.TrimSpace(request.Contact)
	request.OperatingHours = strings.TrimSpace(request.OperatingHours)
	request.Eligibility = strings.TrimSpace(request.Eligibility)
	request.Schedule = strings.TrimSpace(request.Schedule)
	request.Status = NormalizeToken(request.Status)
	request.SourceRef = strings.TrimSpace(request.SourceRef)

	if request.Name != "" && (len(request.Name) > 200 || UnsafeText(request.Name)) {
		return request, "invalid_name", "name must be 200 safe characters or fewer"
	}
	if request.Type != "" && !AllowedReliefPointTypes[request.Type] {
		return request, "invalid_type", "type must be food, water, medical, hygiene, blankets, cash, or mixed"
	}
	if request.Location != nil && !ValidCoordinates(*request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	if request.Status != "" && !AllowedReliefPointStatuses[request.Status] {
		return request, "invalid_status", "status must be open, limited, closed, or paused"
	}
	if len(request.Region) > 100 || UnsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || UnsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if len(request.Address) > 300 || UnsafeText(request.Address) {
		return request, "invalid_address", "address must be 300 safe characters or fewer"
	}
	if len(request.Contact) > 100 || UnsafeText(request.Contact) {
		return request, "invalid_contact", "contact must be 100 safe characters or fewer"
	}
	if len(request.OperatingHours) > 100 || UnsafeText(request.OperatingHours) {
		return request, "invalid_operating_hours", "operatingHours must be 100 safe characters or fewer"
	}
	if len(request.Eligibility) > 700 || UnsafeText(request.Eligibility) {
		return request, "invalid_eligibility", "eligibility must be 700 safe characters or fewer"
	}
	if len(request.Schedule) > 200 || UnsafeText(request.Schedule) {
		return request, "invalid_schedule", "schedule must be 200 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || UnsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	request.StockCategories = NormalizeStockCategories(request.StockCategories)
	return request, "", ""
}

// NormalizeStockCategories cleans and defaults stock category values.
func NormalizeStockCategories(categories []models.ReliefStockCategory) []models.ReliefStockCategory {
	result := make([]models.ReliefStockCategory, 0, len(categories))
	for _, category := range categories {
		category.Category = strings.TrimSpace(category.Category)
		category.Unit = strings.TrimSpace(category.Unit)
		if category.Category == "" {
			continue
		}
		if category.Quantity < 0 {
			category.Quantity = 0
		}
		if category.Unit == "" {
			category.Unit = "units"
		}
		result = append(result, category)
	}
	return result
}

// NormalizeCreateAidRequest validates a create aid request.
func NormalizeCreateAidRequest(request models.CreateAidRequestRequest, now time.Time) (models.CreateAidRequestRequest, string, string) {
	request.Title = strings.TrimSpace(request.Title)
	request.Category = NormalizeToken(request.Category)
	request.Priority = NormalizeToken(request.Priority)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.ReceivingOrganization = strings.TrimSpace(request.ReceivingOrganization)
	request.Contact = strings.TrimSpace(request.Contact)
	request.QuantityUnit = strings.TrimSpace(request.QuantityUnit)
	request.Description = strings.TrimSpace(request.Description)
	request.Visibility = NormalizeToken(request.Visibility)
	request.SourceReliefPointID = strings.TrimSpace(request.SourceReliefPointID)

	if request.Title == "" || len(request.Title) > 180 || UnsafeText(request.Title) {
		return request, "invalid_title", "title is required and must be 180 safe characters or fewer"
	}
	if !AllowedAidRequestCategories[request.Category] {
		return request, "invalid_category", "category is not supported"
	}
	if !AllowedAidRequestPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, medium, high, or urgent"
	}
	if !ValidCoordinates(request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	if request.ReceivingOrganization == "" || len(request.ReceivingOrganization) > 180 || UnsafeText(request.ReceivingOrganization) {
		return request, "invalid_receiving_organization", "receivingOrganization is required and must be 180 safe characters or fewer"
	}
	if request.QuantityNeeded <= 0 {
		return request, "invalid_quantity_needed", "quantityNeeded must be greater than zero"
	}
	if request.QuantityUnit == "" || len(request.QuantityUnit) > 40 || UnsafeText(request.QuantityUnit) {
		return request, "invalid_quantity_unit", "quantityUnit is required and must be 40 safe characters or fewer"
	}
	if request.Description == "" || len(request.Description) > 900 || UnsafeText(request.Description) {
		return request, "invalid_description", "description is required and must be 900 safe characters or fewer"
	}
	if request.NeededBy.IsZero() || request.NeededBy.Before(now.Add(-time.Minute)) {
		return request, "invalid_needed_by", "neededBy must be a future timestamp"
	}
	if request.Visibility == "" {
		request.Visibility = "public"
	}
	if !AllowedAidRequestVisibility[request.Visibility] {
		return request, "invalid_visibility", "visibility must be public or partners_only"
	}
	if len(request.Region) > 100 || UnsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || UnsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if len(request.Contact) > 100 || UnsafeText(request.Contact) {
		return request, "invalid_contact", "contact must be 100 safe characters or fewer"
	}
	if len(request.SourceReliefPointID) > 128 || UnsafeText(request.SourceReliefPointID) {
		return request, "invalid_source_relief_point", "sourceReliefPointId must be 128 safe characters or fewer"
	}
	return request, "", ""
}

// NormalizeReviewAidRequest validates an aid request review.
func NormalizeReviewAidRequest(request models.ReviewAidRequestRequest) (models.ReviewAidRequestRequest, string, string) {
	request.Status = NormalizeToken(request.Status)
	request.ApprovalNotes = strings.TrimSpace(request.ApprovalNotes)
	request.AntiFraudNotes = strings.TrimSpace(request.AntiFraudNotes)
	if !AllowedAidRequestReviewStatuses[request.Status] {
		return request, "invalid_status", "status must be approved, open, paused, closed, or rejected"
	}
	if request.Status == "approved" && request.ApprovalNotes == "" {
		return request, "approval_notes_required", "approvalNotes are required when approving an aid request"
	}
	if len(request.ApprovalNotes) > 700 || UnsafeText(request.ApprovalNotes) {
		return request, "invalid_approval_notes", "approvalNotes must be 700 safe characters or fewer"
	}
	if len(request.AntiFraudNotes) > 700 || UnsafeText(request.AntiFraudNotes) {
		return request, "invalid_anti_fraud_notes", "antiFraudNotes must be 700 safe characters or fewer"
	}
	return request, "", ""
}

// NormalizeCreateAidPledge validates a create pledge request.
func NormalizeCreateAidPledge(request models.CreateAidPledgeRequest) (models.CreateAidPledgeRequest, string, string) {
	request.DonorName = strings.TrimSpace(request.DonorName)
	request.DonorType = NormalizeToken(request.DonorType)
	request.Contact = strings.TrimSpace(request.Contact)
	request.Unit = strings.TrimSpace(request.Unit)
	request.Note = strings.TrimSpace(request.Note)
	if request.DonorName == "" || len(request.DonorName) > 160 || UnsafeText(request.DonorName) {
		return request, "invalid_donor_name", "donorName is required and must be 160 safe characters or fewer"
	}
	if !AllowedAidDonorTypes[request.DonorType] {
		return request, "invalid_donor_type", "donorType is not supported"
	}
	if request.Contact == "" || len(request.Contact) > 160 || UnsafeText(request.Contact) {
		return request, "invalid_contact", "contact is required and must be 160 safe characters or fewer"
	}
	if request.Quantity <= 0 {
		return request, "invalid_quantity", "quantity must be greater than zero"
	}
	if request.Unit == "" || len(request.Unit) > 40 || UnsafeText(request.Unit) {
		return request, "invalid_unit", "unit is required and must be 40 safe characters or fewer"
	}
	if len(request.Note) > 700 || UnsafeText(request.Note) {
		return request, "invalid_note", "note must be 700 safe characters or fewer"
	}
	return request, "", ""
}

// NormalizeReviewAidPledge validates a pledge review request.
func NormalizeReviewAidPledge(request models.ReviewAidPledgeRequest) (models.ReviewAidPledgeRequest, string, string) {
	request.Status = NormalizeToken(request.Status)
	request.ReviewStatus = NormalizeToken(request.ReviewStatus)
	request.FraudReviewNotes = strings.TrimSpace(request.FraudReviewNotes)
	if request.Status == "" && request.ReviewStatus == "" && request.FraudReviewNotes == "" {
		return request, "no_changes", "at least one pledge review field must be supplied"
	}
	if request.Status != "" && !AllowedAidPledgeStatuses[request.Status] {
		return request, "invalid_status", "status is not supported"
	}
	if request.ReviewStatus != "" && !AllowedAidPledgeReviewStatuses[request.ReviewStatus] {
		return request, "invalid_review_status", "reviewStatus is not supported"
	}
	if len(request.FraudReviewNotes) > 700 || UnsafeText(request.FraudReviewNotes) {
		return request, "invalid_fraud_review_notes", "fraudReviewNotes must be 700 safe characters or fewer"
	}
	return request, "", ""
}

// FormatNeededByCSV formats a timestamp for CSV export.
func FormatNeededByCSV(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Fprintf is a thin wrapper around fmt.Fprintf to satisfy linters.
func Fprintf(w http.ResponseWriter, format string, a ...any) (int, error) {
	return fmt.Fprintf(w, format, a...)
}
