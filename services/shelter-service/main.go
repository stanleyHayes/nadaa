package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *memoryStore
	now   func() time.Time
}

type memoryStore struct {
	mu                 sync.RWMutex
	shelters           []shelterRecord
	recovery           []recoverySupportLocation
	hospitals          []hospitalCapacityRecord
	reliefPoints       []reliefPointRecord
	reliefStockHistory []reliefStockHistoryRecord
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type shelterRecord struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Type             string      `json:"type"`
	Region           string      `json:"region"`
	District         string      `json:"district"`
	Address          string      `json:"address"`
	Location         coordinates `json:"location"`
	Capacity         int         `json:"capacity"`
	CurrentOccupancy int         `json:"currentOccupancy"`
	Status           string      `json:"status"`
	Contact          string      `json:"contact"`
	Facilities       []string    `json:"facilities"`
	Notes            string      `json:"notes,omitempty"`
	DistanceMeters   int         `json:"distanceMeters,omitempty"`
	UpdatedBy        string      `json:"updatedBy,omitempty"`
	UpdatedAt        time.Time   `json:"updatedAt"`
}

type recoverySupportLocation struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Region         string      `json:"region"`
	District       string      `json:"district"`
	Address        string      `json:"address"`
	Location       coordinates `json:"location"`
	Contact        string      `json:"contact"`
	Services       []string    `json:"services"`
	Hours          string      `json:"hours"`
	Status         string      `json:"status"`
	DistanceMeters int         `json:"distanceMeters,omitempty"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type reliefPointRecord struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	Region          string                `json:"region"`
	District        string                `json:"district"`
	Address         string                `json:"address"`
	Location        coordinates           `json:"location"`
	Contact         string                `json:"contact"`
	OperatingHours  string                `json:"operatingHours"`
	Eligibility     string                `json:"eligibility"`
	Schedule        string                `json:"schedule"`
	StockCategories []reliefStockCategory `json:"stockCategories"`
	Status          string                `json:"status"`
	Source          string                `json:"source"`
	SourceRef       string                `json:"sourceRef,omitempty"`
	DistanceMeters  int                   `json:"distanceMeters,omitempty"`
	CreatedBy       string                `json:"createdBy,omitempty"`
	UpdatedBy       string                `json:"updatedBy,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}

type reliefStockCategory struct {
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	Unit        string    `json:"unit"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type reliefStockHistoryRecord struct {
	ID              string                `json:"id"`
	ReliefPointID   string                `json:"reliefPointId"`
	ChangedBy       string                `json:"changedBy"`
	ChangedAt       time.Time             `json:"changedAt"`
	Note            string                `json:"note,omitempty"`
	StockCategories []reliefStockCategory `json:"stockCategories"`
}

type hospitalCapacityRecord struct {
	ID                     string      `json:"id"`
	Name                   string      `json:"name"`
	Type                   string      `json:"type"`
	Region                 string      `json:"region"`
	District               string      `json:"district"`
	Address                string      `json:"address"`
	Location               coordinates `json:"location"`
	Contact                string      `json:"contact"`
	Services               []string    `json:"services"`
	TotalBeds              int         `json:"totalBeds"`
	AvailableBeds          int         `json:"availableBeds"`
	ICUBedsAvailable       int         `json:"icuBedsAvailable"`
	MaternityBedsAvailable int         `json:"maternityBedsAvailable"`
	PediatricBedsAvailable int         `json:"pediatricBedsAvailable"`
	IsolationBedsAvailable int         `json:"isolationBedsAvailable"`
	EmergencyCapacity      string      `json:"emergencyCapacity"`
	EmergencyUnitStatus    string      `json:"emergencyUnitStatus"`
	AmbulancesAvailable    int         `json:"ambulancesAvailable"`
	OxygenAvailable        bool        `json:"oxygenAvailable"`
	Notes                  string      `json:"notes,omitempty"`
	Source                 string      `json:"source"`
	SourceRef              string      `json:"sourceRef,omitempty"`
	UpdatedBy              string      `json:"updatedBy,omitempty"`
	UpdatedAt              time.Time   `json:"updatedAt"`
	DistanceMeters         int         `json:"distanceMeters,omitempty"`
	Stale                  bool        `json:"stale"`
	StaleReason            string      `json:"staleReason,omitempty"`
}

type shelterListResponse struct {
	Shelters    []shelterRecord `json:"shelters"`
	GeneratedAt time.Time       `json:"generatedAt"`
}

type nearbyShelterResponse struct {
	Shelters        []shelterRecord           `json:"shelters"`
	RecoverySupport []recoverySupportLocation `json:"recoverySupport"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
}

type recoverySupportResponse struct {
	RecoverySupport []recoverySupportLocation `json:"recoverySupport"`
	GeneratedAt     time.Time                 `json:"generatedAt"`
}

type reliefPointListResponse struct {
	ReliefPoints []reliefPointRecord `json:"reliefPoints"`
	GeneratedAt  time.Time           `json:"generatedAt"`
}

type reliefPointNearbyResponse struct {
	ReliefPoints []reliefPointRecord `json:"reliefPoints"`
	GeneratedAt  time.Time           `json:"generatedAt"`
}

type reliefPointStockHistoryResponse struct {
	ReliefPointID string                     `json:"reliefPointId"`
	History       []reliefStockHistoryRecord `json:"history"`
	GeneratedAt   time.Time                  `json:"generatedAt"`
}

type createReliefPointRequest struct {
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	Region          string                `json:"region,omitempty"`
	District        string                `json:"district,omitempty"`
	Address         string                `json:"address,omitempty"`
	Location        coordinates           `json:"location"`
	Contact         string                `json:"contact,omitempty"`
	OperatingHours  string                `json:"operatingHours,omitempty"`
	Eligibility     string                `json:"eligibility,omitempty"`
	Schedule        string                `json:"schedule,omitempty"`
	StockCategories []reliefStockCategory `json:"stockCategories,omitempty"`
	Status          string                `json:"status,omitempty"`
	Source          string                `json:"source,omitempty"`
	SourceRef       string                `json:"sourceRef,omitempty"`
}

type updateReliefPointRequest struct {
	Name            string                `json:"name,omitempty"`
	Type            string                `json:"type,omitempty"`
	Region          string                `json:"region,omitempty"`
	District        string                `json:"district,omitempty"`
	Address         string                `json:"address,omitempty"`
	Location        *coordinates          `json:"location,omitempty"`
	Contact         string                `json:"contact,omitempty"`
	OperatingHours  string                `json:"operatingHours,omitempty"`
	Eligibility     string                `json:"eligibility,omitempty"`
	Schedule        string                `json:"schedule,omitempty"`
	StockCategories []reliefStockCategory `json:"stockCategories,omitempty"`
	Status          string                `json:"status,omitempty"`
	SourceRef       string                `json:"sourceRef,omitempty"`
}

type occupancyUpdateRequest struct {
	Capacity         *int   `json:"capacity,omitempty"`
	CurrentOccupancy *int   `json:"currentOccupancy,omitempty"`
	Status           string `json:"status,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

type shelterUpdateResponse struct {
	Shelter shelterRecord `json:"shelter"`
}

type hospitalCapacityResponse struct {
	Facilities            []hospitalCapacityRecord `json:"facilities"`
	GeneratedAt           time.Time                `json:"generatedAt"`
	StaleThresholdMinutes int                      `json:"staleThresholdMinutes"`
}

type hospitalCapacityUpdateRequest struct {
	TotalBeds              *int   `json:"totalBeds,omitempty"`
	AvailableBeds          *int   `json:"availableBeds,omitempty"`
	ICUBedsAvailable       *int   `json:"icuBedsAvailable,omitempty"`
	MaternityBedsAvailable *int   `json:"maternityBedsAvailable,omitempty"`
	PediatricBedsAvailable *int   `json:"pediatricBedsAvailable,omitempty"`
	IsolationBedsAvailable *int   `json:"isolationBedsAvailable,omitempty"`
	EmergencyCapacity      string `json:"emergencyCapacity,omitempty"`
	EmergencyUnitStatus    string `json:"emergencyUnitStatus,omitempty"`
	AmbulancesAvailable    *int   `json:"ambulancesAvailable,omitempty"`
	OxygenAvailable        *bool  `json:"oxygenAvailable,omitempty"`
	Notes                  string `json:"notes,omitempty"`
	Source                 string `json:"source,omitempty"`
	SourceRef              string `json:"sourceRef,omitempty"`
}

type hospitalCapacityUpdateResponse struct {
	Facility hospitalCapacityRecord `json:"facility"`
}

type hospitalCapacityImportRequest struct {
	Source    string                          `json:"source,omitempty"`
	SourceRef string                          `json:"sourceRef,omitempty"`
	Records   []hospitalCapacityFixtureRecord `json:"records,omitempty"`
}

type hospitalCapacityFixtureRecord struct {
	FacilityID             string `json:"facilityId"`
	AvailableBeds          int    `json:"availableBeds"`
	ICUBedsAvailable       int    `json:"icuBedsAvailable,omitempty"`
	MaternityBedsAvailable int    `json:"maternityBedsAvailable,omitempty"`
	PediatricBedsAvailable int    `json:"pediatricBedsAvailable,omitempty"`
	IsolationBedsAvailable int    `json:"isolationBedsAvailable,omitempty"`
	EmergencyCapacity      string `json:"emergencyCapacity"`
	EmergencyUnitStatus    string `json:"emergencyUnitStatus,omitempty"`
	AmbulancesAvailable    int    `json:"ambulancesAvailable,omitempty"`
	OxygenAvailable        *bool  `json:"oxygenAvailable,omitempty"`
	Notes                  string `json:"notes,omitempty"`
}

type hospitalCapacityImportResponse struct {
	Imported    int                      `json:"imported"`
	Facilities  []hospitalCapacityRecord `json:"facilities"`
	GeneratedAt time.Time                `json:"generatedAt"`
	Source      string                   `json:"source"`
}

type authorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var shelterUpdateRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

var allowedShelterStatuses = map[string]bool{
	"open":    true,
	"full":    true,
	"closed":  true,
	"unknown": true,
}

var allowedEmergencyCapacityStatuses = map[string]bool{
	"available": true,
	"limited":   true,
	"full":      true,
	"offline":   true,
	"unknown":   true,
}

var allowedEmergencyUnitStatuses = map[string]bool{
	"open":    true,
	"busy":    true,
	"divert":  true,
	"closed":  true,
	"unknown": true,
}

var allowedReliefPointStatuses = map[string]bool{
	"open":    true,
	"limited": true,
	"closed":  true,
	"paused":  true,
}

var allowedReliefPointTypes = map[string]bool{
	"food":     true,
	"water":    true,
	"medical":  true,
	"hygiene":  true,
	"blankets": true,
	"cash":     true,
	"mixed":    true,
}

const (
	earthRadiusMeters                 = 6371000.0
	nearbySearchMeters                = 30000.0
	defaultNearbyLimit                = 6
	hospitalCapacityStaleAfter        = 30 * time.Minute
	hospitalCapacityStaleAfterMinutes = int(hospitalCapacityStaleAfter / time.Minute)
)

func main() {
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/shelters", srv.listSheltersHandler)
	mux.HandleFunc("GET /api/v1/shelters/nearby", srv.nearbySheltersHandler)
	mux.HandleFunc("GET /api/v1/recovery-support/nearby", srv.nearbyRecoverySupportHandler)
	mux.HandleFunc("PATCH /api/v1/shelters/{id}/occupancy", srv.updateShelterOccupancyHandler)
	mux.HandleFunc("GET /api/v1/hospitals/capacity", srv.listHospitalCapacityHandler)
	mux.HandleFunc("PATCH /api/v1/hospitals/{id}/capacity", srv.updateHospitalCapacityHandler)
	mux.HandleFunc("POST /api/v1/hospitals/capacity/imports/fixture", srv.importHospitalCapacityFixtureHandler)
	mux.HandleFunc("GET /api/v1/relief-points", srv.listReliefPointsHandler)
	mux.HandleFunc("GET /api/v1/relief-points/nearby", srv.nearbyReliefPointsHandler)
	mux.HandleFunc("POST /api/v1/relief-points", srv.createReliefPointHandler)
	mux.HandleFunc("PATCH /api/v1/relief-points/{id}", srv.updateReliefPointHandler)
	mux.HandleFunc("GET /api/v1/relief-points/{id}/stock-history", srv.listReliefPointStockHistoryHandler)

	addr := envOrDefault("NADAA_SHELTER_ADDR", ":8093")
	log.Printf("shelter-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServer() *server {
	now := time.Now
	return &server{store: newMemoryStore(now().UTC()), now: now}
}

func newMemoryStore(now time.Time) *memoryStore {
	return &memoryStore{
		shelters: []shelterRecord{
			{
				ID:               "00000000-0000-0000-0000-000000000301",
				Name:             "Accra Metro Assembly Shelter",
				Type:             "evacuation_shelter",
				Region:           "Greater Accra",
				District:         "Accra Metropolitan",
				Address:          "Accra Metropolitan Assembly Hall",
				Location:         coordinates{Lat: 5.560, Lng: -0.200},
				Capacity:         450,
				CurrentOccupancy: 116,
				Status:           "open",
				Contact:          "112",
				Facilities:       []string{"water", "first_aid", "accessible_entry", "family_area"},
				Notes:            "Primary flood evacuation shelter for central Accra.",
				UpdatedAt:        now,
			},
			{
				ID:               "00000000-0000-0000-0000-000000000302",
				Name:             "Osu Community Hall",
				Type:             "temporary_shelter",
				Region:           "Greater Accra",
				District:         "Korle Klottey",
				Address:          "Osu Community Hall",
				Location:         coordinates{Lat: 5.550, Lng: -0.180},
				Capacity:         220,
				CurrentOccupancy: 34,
				Status:           "open",
				Contact:          "112",
				Facilities:       []string{"water", "first_aid", "family_area"},
				Notes:            "Suitable for short-term shelter and reunification.",
				UpdatedAt:        now,
			},
			{
				ID:               "00000000-0000-0000-0000-000000000303",
				Name:             "Kaneshie Social Centre",
				Type:             "relief_shelter",
				Region:           "Greater Accra",
				District:         "Okaikwei South",
				Address:          "Kaneshie Market Road",
				Location:         coordinates{Lat: 5.566, Lng: -0.242},
				Capacity:         180,
				CurrentOccupancy: 180,
				Status:           "full",
				Contact:          "112",
				Facilities:       []string{"water", "food_distribution"},
				Notes:            "At capacity; redirect new arrivals unless occupancy changes.",
				UpdatedAt:        now,
			},
		},
		recovery: []recoverySupportLocation{
			{
				ID:        "recovery_ama_relief_001",
				Name:      "AMA Relief Distribution Point",
				Type:      "relief_point",
				Region:    "Greater Accra",
				District:  "Accra Metropolitan",
				Address:   "Independence Avenue recovery desk",
				Location:  coordinates{Lat: 5.558, Lng: -0.197},
				Contact:   "112",
				Services:  []string{"food", "water", "blankets", "family_reunification"},
				Hours:     "08:00-20:00",
				Status:    "open",
				UpdatedAt: now,
			},
			{
				ID:        "recovery_korle_bu_medical_001",
				Name:      "Korle Bu Emergency Stabilization Desk",
				Type:      "medical_support",
				Region:    "Greater Accra",
				District:  "Accra Metropolitan",
				Address:   "Korle Bu emergency entrance",
				Location:  coordinates{Lat: 5.536, Lng: -0.227},
				Contact:   "112",
				Services:  []string{"first_aid", "triage", "medical_referral"},
				Hours:     "24 hours",
				Status:    "open",
				UpdatedAt: now,
			},
			{
				ID:        "recovery_osu_registration_001",
				Name:      "Osu Recovery Registration Desk",
				Type:      "recovery_registration",
				Region:    "Greater Accra",
				District:  "Korle Klottey",
				Address:   "Osu Community Hall annex",
				Location:  coordinates{Lat: 5.551, Lng: -0.181},
				Contact:   "112",
				Services:  []string{"needs_registration", "damage_reporting", "case_follow_up"},
				Hours:     "08:00-18:00",
				Status:    "open",
				UpdatedAt: now,
			},
		},
		hospitals: []hospitalCapacityRecord{
			{
				ID:                     "hospital_001",
				Name:                   "Korle Bu Teaching Hospital",
				Type:                   "teaching_hospital",
				Region:                 "Greater Accra",
				District:               "Accra Metropolitan",
				Address:                "Korle Bu emergency entrance",
				Location:               coordinates{Lat: 5.536, Lng: -0.227},
				Contact:                "0302665401",
				Services:               []string{"emergency", "trauma", "icu", "maternity", "pediatric", "oxygen"},
				TotalBeds:              820,
				AvailableBeds:          46,
				ICUBedsAvailable:       4,
				MaternityBedsAvailable: 9,
				PediatricBedsAvailable: 5,
				IsolationBedsAvailable: 3,
				EmergencyCapacity:      "available",
				EmergencyUnitStatus:    "open",
				AmbulancesAvailable:    3,
				OxygenAvailable:        true,
				Notes:                  "Major referral facility for Accra emergency transfers.",
				Source:                 "fixture",
				SourceRef:              "hospital-capacity-feed",
				UpdatedAt:              now,
			},
			{
				ID:                     "hospital_002",
				Name:                   "Greater Accra Regional Hospital",
				Type:                   "regional_hospital",
				Region:                 "Greater Accra",
				District:               "Accra Metropolitan",
				Address:                "Ridge Hospital emergency unit",
				Location:               coordinates{Lat: 5.563, Lng: -0.191},
				Contact:                "0302425201",
				Services:               []string{"emergency", "trauma", "oxygen", "ambulance"},
				TotalBeds:              420,
				AvailableBeds:          12,
				ICUBedsAvailable:       1,
				MaternityBedsAvailable: 3,
				PediatricBedsAvailable: 2,
				IsolationBedsAvailable: 0,
				EmergencyCapacity:      "limited",
				EmergencyUnitStatus:    "busy",
				AmbulancesAvailable:    1,
				OxygenAvailable:        true,
				Notes:                  "Emergency unit busy; use for critical stabilization.",
				Source:                 "fixture",
				SourceRef:              "hospital-capacity-feed",
				UpdatedAt:              now.Add(-12 * time.Minute),
			},
			{
				ID:                     "hospital_003",
				Name:                   "Tema General Hospital",
				Type:                   "general_hospital",
				Region:                 "Greater Accra",
				District:               "Tema Metropolitan",
				Address:                "Tema Community 12",
				Location:               coordinates{Lat: 5.669, Lng: -0.016},
				Contact:                "0303202231",
				Services:               []string{"emergency", "maternity", "pediatric", "ambulance"},
				TotalBeds:              310,
				AvailableBeds:          0,
				ICUBedsAvailable:       0,
				MaternityBedsAvailable: 1,
				PediatricBedsAvailable: 0,
				IsolationBedsAvailable: 0,
				EmergencyCapacity:      "full",
				EmergencyUnitStatus:    "divert",
				AmbulancesAvailable:    0,
				OxygenAvailable:        false,
				Notes:                  "Capacity stale; confirm before transfer.",
				Source:                 "fixture",
				SourceRef:              "hospital-capacity-feed",
				UpdatedAt:              now.Add(-45 * time.Minute),
			},
		},
		reliefPoints: []reliefPointRecord{
			{
				ID:             "relief_ama_food_001",
				Name:           "AMA Central Food Distribution",
				Type:           "food",
				Region:         "Greater Accra",
				District:       "Accra Metropolitan",
				Address:        "Independence Avenue, Accra",
				Location:       coordinates{Lat: 5.560, Lng: -0.200},
				Contact:        "0302112233",
				OperatingHours: "08:00-18:00",
				Eligibility:    "Open to households affected by flooding.",
				Schedule:       "Daily",
				StockCategories: []reliefStockCategory{
					{Category: "rice_kg", Quantity: 500, Unit: "kg", LastUpdated: now},
					{Category: "water_bottles", Quantity: 1200, Unit: "bottles", LastUpdated: now},
				},
				Status:    "open",
				Source:    "manual",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:             "relief_kaneshie_water_001",
				Name:           "Kaneshie Water Relief Point",
				Type:           "water",
				Region:         "Greater Accra",
				District:       "Okaikwei South",
				Address:        "Kaneshie Market Road",
				Location:       coordinates{Lat: 5.566, Lng: -0.242},
				Contact:        "112",
				OperatingHours: "06:00-22:00",
				Eligibility:    "Residents in flood-affected zones.",
				Schedule:       "Daily",
				StockCategories: []reliefStockCategory{
					{Category: "water_bottles", Quantity: 800, Unit: "bottles", LastUpdated: now},
					{Category: "water_sachets", Quantity: 2000, Unit: "sachets", LastUpdated: now},
				},
				Status:    "limited",
				Source:    "manual",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:             "relief_madina_mixed_001",
				Name:           "Madina Emergency Relief Hub",
				Type:           "mixed",
				Region:         "Greater Accra",
				District:       "La Nkwantanang Madina",
				Address:        "Madina Zongo Junction",
				Location:       coordinates{Lat: 5.680, Lng: -0.160},
				Contact:        "0302445566",
				OperatingHours: "07:00-19:00",
				Eligibility:    "Displaced families and vulnerable groups.",
				Schedule:       "Daily",
				StockCategories: []reliefStockCategory{
					{Category: "blankets", Quantity: 150, Unit: "units", LastUpdated: now},
					{Category: "hygiene_kits", Quantity: 80, Unit: "kits", LastUpdated: now},
					{Category: "water_bottles", Quantity: 300, Unit: "bottles", LastUpdated: now},
				},
				Status:    "open",
				Source:    "manual",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "shelter-service"})
}

func (s *server) listSheltersHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, shelterListResponse{Shelters: s.store.listShelters(), GeneratedAt: s.now().UTC()})
}

func (s *server) nearbySheltersHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, nearbyShelterResponse{
		Shelters:        s.store.nearbyShelters(location, defaultNearbyLimit),
		RecoverySupport: s.store.nearbyRecoverySupport(location, defaultNearbyLimit),
		GeneratedAt:     s.now().UTC(),
	})
}

func (s *server) nearbyRecoverySupportHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, recoverySupportResponse{
		RecoverySupport: s.store.nearbyRecoverySupport(location, defaultNearbyLimit),
		GeneratedAt:     s.now().UTC(),
	})
}

func (s *server) updateShelterOccupancyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, shelterUpdateRoles)
	if !ok {
		return
	}

	var request occupancyUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeOccupancyUpdate(request)
	if code != "" {
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	shelter, code, message := s.store.updateShelter(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		writeError(w, statusForCode(code), code, message)
		return
	}
	writeJSON(w, http.StatusOK, shelterUpdateResponse{Shelter: shelter})
}

func (s *server) listHospitalCapacityHandler(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseHospitalCapacityFilter(w, r)
	if !ok {
		return
	}
	facilities := s.store.listHospitalCapacity(filter, s.now().UTC())
	log.Printf("INFO shelter-service hospital_capacity_list count=%d service=%s emergencyCapacity=%s minAvailableBeds=%d includeStale=%t", len(facilities), filter.Service, filter.EmergencyCapacity, filter.MinAvailableBeds, filter.IncludeStale)
	writeJSON(w, http.StatusOK, hospitalCapacityResponse{
		Facilities:            facilities,
		GeneratedAt:           s.now().UTC(),
		StaleThresholdMinutes: hospitalCapacityStaleAfterMinutes,
	})
}

func (s *server) updateHospitalCapacityHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, shelterUpdateRoles)
	if !ok {
		return
	}

	var request hospitalCapacityUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service hospital_capacity_update invalid_json facilityId=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	log.Printf("INFO shelter-service hospital_capacity_update received facilityId=%s actor=%s source=%s", r.PathValue("id"), ctx.ActorUserID, request.Source)

	normalized, code, message := normalizeHospitalCapacityUpdate(request)
	if code != "" {
		log.Printf("WARN shelter-service hospital_capacity_update validation_failed facilityId=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	facility, code, message := s.store.updateHospitalCapacity(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN shelter-service hospital_capacity_update failed facilityId=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service hospital_capacity_update completed facilityId=%s availableBeds=%d emergencyCapacity=%s source=%s", facility.ID, facility.AvailableBeds, facility.EmergencyCapacity, facility.Source)
	writeJSON(w, http.StatusOK, hospitalCapacityUpdateResponse{Facility: facility})
}

func (s *server) importHospitalCapacityFixtureHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, shelterUpdateRoles)
	if !ok {
		return
	}

	var request hospitalCapacityImportRequest
	if err := optionalDecodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service hospital_capacity_fixture_import invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	normalized, code, message := normalizeHospitalCapacityImport(request)
	if code != "" {
		log.Printf("WARN shelter-service hospital_capacity_fixture_import validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	facilities, imported := s.store.importHospitalCapacityFixture(normalized, ctx, s.now().UTC())
	log.Printf("INFO shelter-service hospital_capacity_fixture_import completed actor=%s source=%s imported=%d", ctx.ActorUserID, normalized.Source, imported)
	writeJSON(w, http.StatusOK, hospitalCapacityImportResponse{
		Imported:    imported,
		Facilities:  facilities,
		GeneratedAt: s.now().UTC(),
		Source:      normalized.Source,
	})
}

func (s *server) listReliefPointsHandler(w http.ResponseWriter, r *http.Request) {
	filter := parseReliefPointFilter(r)
	reliefPoints := s.store.listReliefPoints(filter)
	log.Printf("INFO shelter-service relief_point_list count=%d status=%s type=%s hasLocation=%t bbox=%t", len(reliefPoints), filter.Status, filter.Type, filter.Location != nil, filter.BBox != nil)
	writeJSON(w, http.StatusOK, reliefPointListResponse{ReliefPoints: reliefPoints, GeneratedAt: s.now().UTC()})
}

func (s *server) nearbyReliefPointsHandler(w http.ResponseWriter, r *http.Request) {
	location, ok := parseLocation(w, r)
	if !ok {
		return
	}
	reliefPoints := s.store.nearbyReliefPoints(location, defaultNearbyLimit)
	log.Printf("INFO shelter-service relief_point_nearby count=%d lat=%.3f lng=%.3f", len(reliefPoints), location.Lat, location.Lng)
	writeJSON(w, http.StatusOK, reliefPointNearbyResponse{ReliefPoints: reliefPoints, GeneratedAt: s.now().UTC()})
}

func (s *server) createReliefPointHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, shelterUpdateRoles)
	if !ok {
		return
	}

	var request createReliefPointRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service relief_point_create invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreateReliefPoint(request)
	if code != "" {
		log.Printf("WARN shelter-service relief_point_create validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	reliefPoint := s.store.createReliefPoint(normalized, ctx, s.now().UTC())
	log.Printf("INFO shelter-service relief_point_create completed id=%s actor=%s source=%s", reliefPoint.ID, ctx.ActorUserID, reliefPoint.Source)
	writeJSON(w, http.StatusCreated, reliefPoint)
}

func (s *server) updateReliefPointHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r, shelterUpdateRoles)
	if !ok {
		return
	}

	var request updateReliefPointRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN shelter-service relief_point_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateReliefPoint(request)
	if code != "" {
		log.Printf("WARN shelter-service relief_point_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	reliefPoint, code, message := s.store.updateReliefPoint(r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN shelter-service relief_point_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO shelter-service relief_point_update completed id=%s actor=%s status=%s", reliefPoint.ID, ctx.ActorUserID, reliefPoint.Status)
	writeJSON(w, http.StatusOK, reliefPoint)
}

func (s *server) listReliefPointStockHistoryHandler(w http.ResponseWriter, r *http.Request) {
	history := s.store.listReliefPointStockHistory(r.PathValue("id"))
	log.Printf("INFO shelter-service relief_point_stock_history reliefPointId=%s count=%d", r.PathValue("id"), len(history))
	writeJSON(w, http.StatusOK, reliefPointStockHistoryResponse{ReliefPointID: r.PathValue("id"), History: history, GeneratedAt: s.now().UTC()})
}

func (m *memoryStore) listShelters() []shelterRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shelters := copyShelters(m.shelters)
	sort.Slice(shelters, func(i, j int) bool {
		if shelters[i].Status == shelters[j].Status {
			return shelters[i].Name < shelters[j].Name
		}
		return shelterStatusRank(shelters[i].Status) < shelterStatusRank(shelters[j].Status)
	})
	return shelters
}

func (m *memoryStore) nearbyShelters(location coordinates, limit int) []shelterRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shelters := make([]shelterRecord, 0, len(m.shelters))
	for _, shelter := range m.shelters {
		shelter.DistanceMeters = int(math.Round(distanceMeters(location, shelter.Location)))
		if float64(shelter.DistanceMeters) <= nearbySearchMeters {
			shelters = append(shelters, shelter)
		}
	}

	sort.Slice(shelters, func(i, j int) bool {
		if shelters[i].DistanceMeters == shelters[j].DistanceMeters {
			return shelterStatusRank(shelters[i].Status) < shelterStatusRank(shelters[j].Status)
		}
		return shelters[i].DistanceMeters < shelters[j].DistanceMeters
	})
	if limit > 0 && len(shelters) > limit {
		shelters = shelters[:limit]
	}
	return copyShelters(shelters)
}

func (m *memoryStore) nearbyRecoverySupport(location coordinates, limit int) []recoverySupportLocation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locations := make([]recoverySupportLocation, 0, len(m.recovery))
	for _, item := range m.recovery {
		item.DistanceMeters = int(math.Round(distanceMeters(location, item.Location)))
		if float64(item.DistanceMeters) <= nearbySearchMeters {
			locations = append(locations, item)
		}
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].DistanceMeters < locations[j].DistanceMeters
	})
	if limit > 0 && len(locations) > limit {
		locations = locations[:limit]
	}
	return copyRecovery(locations)
}

func (m *memoryStore) updateShelter(id string, request occupancyUpdateRequest, ctx authorityContext, now time.Time) (shelterRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.shelters {
		if m.shelters[index].ID != id {
			continue
		}

		next := m.shelters[index]
		if request.Capacity != nil {
			next.Capacity = *request.Capacity
		}
		if request.CurrentOccupancy != nil {
			next.CurrentOccupancy = *request.CurrentOccupancy
		}
		if request.Status != "" {
			next.Status = request.Status
		} else {
			next.Status = statusForOccupancy(next.Capacity, next.CurrentOccupancy, next.Status)
		}
		if request.Notes != "" {
			next.Notes = request.Notes
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		m.shelters[index] = next
		return next, "", ""
	}

	return shelterRecord{}, "not_found", "shelter was not found"
}

func (m *memoryStore) listReliefPoints(filter reliefPointFilter) []reliefPointRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reliefPoints := make([]reliefPointRecord, 0, len(m.reliefPoints))
	for _, point := range m.reliefPoints {
		if filter.Status != "" && point.Status != filter.Status {
			continue
		}
		if filter.Type != "" && point.Type != filter.Type {
			continue
		}
		if filter.BBox != nil && !pointInBBox(point.Location, *filter.BBox) {
			continue
		}
		if filter.Location != nil {
			distance := int(math.Round(distanceMeters(*filter.Location, point.Location)))
			if float64(distance) > filter.RadiusMeters {
				continue
			}
			point.DistanceMeters = distance
		}
		reliefPoints = append(reliefPoints, point)
	}

	sort.Slice(reliefPoints, func(i, j int) bool {
		if reliefPoints[i].Status != reliefPoints[j].Status {
			return reliefPointStatusRank(reliefPoints[i].Status) < reliefPointStatusRank(reliefPoints[j].Status)
		}
		if reliefPoints[i].DistanceMeters != reliefPoints[j].DistanceMeters {
			return reliefPoints[i].DistanceMeters < reliefPoints[j].DistanceMeters
		}
		return reliefPoints[i].Name < reliefPoints[j].Name
	})

	if filter.Limit > 0 && len(reliefPoints) > filter.Limit {
		reliefPoints = reliefPoints[:filter.Limit]
	}
	return copyReliefPoints(reliefPoints)
}

func (m *memoryStore) nearbyReliefPoints(location coordinates, limit int) []reliefPointRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reliefPoints := make([]reliefPointRecord, 0, len(m.reliefPoints))
	for _, point := range m.reliefPoints {
		point.DistanceMeters = int(math.Round(distanceMeters(location, point.Location)))
		if float64(point.DistanceMeters) <= nearbySearchMeters {
			reliefPoints = append(reliefPoints, point)
		}
	}

	sort.Slice(reliefPoints, func(i, j int) bool {
		if reliefPoints[i].DistanceMeters == reliefPoints[j].DistanceMeters {
			return reliefPointStatusRank(reliefPoints[i].Status) < reliefPointStatusRank(reliefPoints[j].Status)
		}
		return reliefPoints[i].DistanceMeters < reliefPoints[j].DistanceMeters
	})

	if limit > 0 && len(reliefPoints) > limit {
		reliefPoints = reliefPoints[:limit]
	}
	return copyReliefPoints(reliefPoints)
}

func (m *memoryStore) createReliefPoint(request createReliefPointRequest, ctx authorityContext, now time.Time) reliefPointRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	point := reliefPointRecord{
		ID:              fmt.Sprintf("relief_%03d", len(m.reliefPoints)+1),
		Name:            request.Name,
		Type:            request.Type,
		Region:          request.Region,
		District:        request.District,
		Address:         request.Address,
		Location:        request.Location,
		Contact:         request.Contact,
		OperatingHours:  request.OperatingHours,
		Eligibility:     request.Eligibility,
		Schedule:        request.Schedule,
		StockCategories: copyStockCategories(request.StockCategories),
		Status:          request.Status,
		Source:          request.Source,
		SourceRef:       request.SourceRef,
		CreatedBy:       ctx.ActorUserID,
		UpdatedBy:       ctx.ActorUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if point.Status == "" {
		point.Status = "open"
	}
	for i := range point.StockCategories {
		point.StockCategories[i].LastUpdated = now
	}
	m.reliefPoints = append(m.reliefPoints, point)
	return point
}

func (m *memoryStore) updateReliefPoint(id string, request updateReliefPointRequest, ctx authorityContext, now time.Time) (reliefPointRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.reliefPoints {
		if m.reliefPoints[index].ID != id {
			continue
		}

		next := m.reliefPoints[index]
		stockChanged := false
		if request.Name != "" {
			next.Name = request.Name
		}
		if request.Type != "" {
			next.Type = request.Type
		}
		if request.Region != "" {
			next.Region = request.Region
		}
		if request.District != "" {
			next.District = request.District
		}
		if request.Address != "" {
			next.Address = request.Address
		}
		if request.Location != nil {
			next.Location = *request.Location
		}
		if request.Contact != "" {
			next.Contact = request.Contact
		}
		if request.OperatingHours != "" {
			next.OperatingHours = request.OperatingHours
		}
		if request.Eligibility != "" {
			next.Eligibility = request.Eligibility
		}
		if request.Schedule != "" {
			next.Schedule = request.Schedule
		}
		if request.Status != "" {
			next.Status = request.Status
		}
		if request.SourceRef != "" {
			next.SourceRef = request.SourceRef
		}
		if request.StockCategories != nil {
			for i := range request.StockCategories {
				request.StockCategories[i].LastUpdated = now
			}
			if !stockCategoriesEqual(next.StockCategories, request.StockCategories) {
				stockChanged = true
			}
			next.StockCategories = copyStockCategories(request.StockCategories)
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now

		if stockChanged {
			history := reliefStockHistoryRecord{
				ID:              fmt.Sprintf("rsh_%03d", len(m.reliefStockHistory)+1),
				ReliefPointID:   next.ID,
				ChangedBy:       ctx.ActorUserID,
				ChangedAt:       now,
				StockCategories: copyStockCategories(next.StockCategories),
			}
			m.reliefStockHistory = append(m.reliefStockHistory, history)
		}

		m.reliefPoints[index] = next
		return next, "", ""
	}

	return reliefPointRecord{}, "not_found", "relief point was not found"
}

func (m *memoryStore) listReliefPointStockHistory(reliefPointID string) []reliefStockHistoryRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reliefPointID = strings.TrimSpace(reliefPointID)
	history := make([]reliefStockHistoryRecord, 0)
	for _, record := range m.reliefStockHistory {
		if record.ReliefPointID == reliefPointID {
			history = append(history, record)
		}
	}
	sort.Slice(history, func(i, j int) bool {
		return history[i].ChangedAt.After(history[j].ChangedAt)
	})
	return history
}

type hospitalCapacityFilter struct {
	Location          *coordinates
	Service           string
	EmergencyCapacity string
	MinAvailableBeds  int
	IncludeStale      bool
	Limit             int
}

func (m *memoryStore) listHospitalCapacity(filter hospitalCapacityFilter, now time.Time) []hospitalCapacityRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	facilities := make([]hospitalCapacityRecord, 0, len(m.hospitals))
	for _, facility := range m.hospitals {
		facility = withHospitalStaleness(facility, now)
		if filter.Location != nil {
			facility.DistanceMeters = int(math.Round(distanceMeters(*filter.Location, facility.Location)))
			if float64(facility.DistanceMeters) > nearbySearchMeters {
				continue
			}
		}
		if filter.Service != "" && !containsNormalized(facility.Services, filter.Service) {
			continue
		}
		if filter.EmergencyCapacity != "" && facility.EmergencyCapacity != filter.EmergencyCapacity {
			continue
		}
		if filter.MinAvailableBeds > 0 && facility.AvailableBeds < filter.MinAvailableBeds {
			continue
		}
		if !filter.IncludeStale && facility.Stale {
			continue
		}
		facilities = append(facilities, facility)
	}
	sort.Slice(facilities, func(i, j int) bool {
		if facilities[i].Stale != facilities[j].Stale {
			return !facilities[i].Stale
		}
		if filter.Location != nil && facilities[i].DistanceMeters != facilities[j].DistanceMeters {
			return facilities[i].DistanceMeters < facilities[j].DistanceMeters
		}
		if facilities[i].EmergencyCapacity != facilities[j].EmergencyCapacity {
			return hospitalCapacityRank(facilities[i].EmergencyCapacity) < hospitalCapacityRank(facilities[j].EmergencyCapacity)
		}
		if facilities[i].AvailableBeds != facilities[j].AvailableBeds {
			return facilities[i].AvailableBeds > facilities[j].AvailableBeds
		}
		return facilities[i].Name < facilities[j].Name
	})
	if filter.Limit > 0 && len(facilities) > filter.Limit {
		facilities = facilities[:filter.Limit]
	}
	return copyHospitals(facilities)
}

func (m *memoryStore) updateHospitalCapacity(id string, request hospitalCapacityUpdateRequest, ctx authorityContext, now time.Time) (hospitalCapacityRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.hospitals {
		if m.hospitals[index].ID != id {
			continue
		}

		next := m.hospitals[index]
		if request.TotalBeds != nil {
			next.TotalBeds = *request.TotalBeds
		}
		if request.AvailableBeds != nil {
			next.AvailableBeds = *request.AvailableBeds
		}
		if next.AvailableBeds > next.TotalBeds {
			return hospitalCapacityRecord{}, "invalid_available_beds", "availableBeds cannot exceed totalBeds"
		}
		if request.ICUBedsAvailable != nil {
			next.ICUBedsAvailable = *request.ICUBedsAvailable
		}
		if request.MaternityBedsAvailable != nil {
			next.MaternityBedsAvailable = *request.MaternityBedsAvailable
		}
		if request.PediatricBedsAvailable != nil {
			next.PediatricBedsAvailable = *request.PediatricBedsAvailable
		}
		if request.IsolationBedsAvailable != nil {
			next.IsolationBedsAvailable = *request.IsolationBedsAvailable
		}
		if request.EmergencyCapacity != "" {
			next.EmergencyCapacity = request.EmergencyCapacity
		} else {
			next.EmergencyCapacity = hospitalCapacityFromBeds(next.TotalBeds, next.AvailableBeds, next.EmergencyCapacity)
		}
		if request.EmergencyUnitStatus != "" {
			next.EmergencyUnitStatus = request.EmergencyUnitStatus
		}
		if request.AmbulancesAvailable != nil {
			next.AmbulancesAvailable = *request.AmbulancesAvailable
		}
		if request.OxygenAvailable != nil {
			next.OxygenAvailable = *request.OxygenAvailable
		}
		if request.Notes != "" {
			next.Notes = request.Notes
		}
		next.Source = request.Source
		next.SourceRef = request.SourceRef
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		next = withHospitalStaleness(next, now)
		m.hospitals[index] = next
		return copyHospital(next), "", ""
	}

	return hospitalCapacityRecord{}, "not_found", "hospital facility was not found"
}

func (m *memoryStore) importHospitalCapacityFixture(request hospitalCapacityImportRequest, ctx authorityContext, now time.Time) ([]hospitalCapacityRecord, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	updates := request.Records
	if len(updates) == 0 {
		updates = defaultHospitalCapacityFixture()
	}
	byID := map[string]hospitalCapacityFixtureRecord{}
	for _, record := range updates {
		byID[record.FacilityID] = record
	}

	imported := 0
	facilities := make([]hospitalCapacityRecord, 0, len(m.hospitals))
	for index := range m.hospitals {
		update, ok := byID[m.hospitals[index].ID]
		if !ok {
			continue
		}
		next := m.hospitals[index]
		next.AvailableBeds = update.AvailableBeds
		next.ICUBedsAvailable = update.ICUBedsAvailable
		next.MaternityBedsAvailable = update.MaternityBedsAvailable
		next.PediatricBedsAvailable = update.PediatricBedsAvailable
		next.IsolationBedsAvailable = update.IsolationBedsAvailable
		next.EmergencyCapacity = update.EmergencyCapacity
		if update.EmergencyUnitStatus != "" {
			next.EmergencyUnitStatus = update.EmergencyUnitStatus
		}
		next.AmbulancesAvailable = update.AmbulancesAvailable
		if update.OxygenAvailable != nil {
			next.OxygenAvailable = *update.OxygenAvailable
		}
		if update.Notes != "" {
			next.Notes = update.Notes
		}
		next.Source = request.Source
		next.SourceRef = request.SourceRef
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		next = withHospitalStaleness(next, now)
		m.hospitals[index] = next
		facilities = append(facilities, copyHospital(next))
		imported++
	}
	return facilities, imported
}

func normalizeOccupancyUpdate(request occupancyUpdateRequest) (occupancyUpdateRequest, string, string) {
	request.Status = strings.TrimSpace(strings.ToLower(request.Status))
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
	if request.Status != "" && !allowedShelterStatuses[request.Status] {
		return request, "invalid_status", "status must be open, full, closed, or unknown"
	}
	if len(request.Notes) > 500 || unsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 500 safe characters or fewer"
	}
	return request, "", ""
}

func parseHospitalCapacityFilter(w http.ResponseWriter, r *http.Request) (hospitalCapacityFilter, bool) {
	filter := hospitalCapacityFilter{
		EmergencyCapacity: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("emergencyCapacity"))),
		IncludeStale:      strings.TrimSpace(strings.ToLower(r.URL.Query().Get("includeStale"))) != "false",
		Limit:             defaultNearbyLimit,
		Service:           normalizeToken(r.URL.Query().Get("service")),
	}
	if filter.EmergencyCapacity != "" && !allowedEmergencyCapacityStatuses[filter.EmergencyCapacity] {
		writeError(w, http.StatusBadRequest, "invalid_emergency_capacity", "emergencyCapacity must be available, limited, full, offline, or unknown")
		return filter, false
	}
	if value := strings.TrimSpace(r.URL.Query().Get("minAvailableBeds")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			writeError(w, http.StatusBadRequest, "invalid_min_available_beds", "minAvailableBeds must be zero or greater")
			return filter, false
		}
		filter.MinAvailableBeds = parsed
	}
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 50 {
			writeError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 50")
			return filter, false
		}
		filter.Limit = parsed
	}

	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" && lngText == "" {
		return filter, true
	}
	if latText == "" || lngText == "" {
		writeError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng must be supplied together")
		return filter, false
	}
	location, ok := parseLocation(w, r)
	if !ok {
		return filter, false
	}
	filter.Location = &location
	return filter, true
}

func normalizeHospitalCapacityUpdate(request hospitalCapacityUpdateRequest) (hospitalCapacityUpdateRequest, string, string) {
	request.EmergencyCapacity = normalizeToken(request.EmergencyCapacity)
	request.EmergencyUnitStatus = normalizeToken(request.EmergencyUnitStatus)
	request.Notes = strings.TrimSpace(request.Notes)
	request.Source = normalizeToken(request.Source)
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
			return request, "invalid_" + normalizeToken(item.name), item.name + " must be zero or greater"
		}
	}
	if request.TotalBeds != nil && request.AvailableBeds != nil && *request.AvailableBeds > *request.TotalBeds {
		return request, "invalid_available_beds", "availableBeds cannot exceed totalBeds"
	}
	if request.EmergencyCapacity != "" && !allowedEmergencyCapacityStatuses[request.EmergencyCapacity] {
		return request, "invalid_emergency_capacity", "emergencyCapacity must be available, limited, full, offline, or unknown"
	}
	if request.EmergencyUnitStatus != "" && !allowedEmergencyUnitStatuses[request.EmergencyUnitStatus] {
		return request, "invalid_emergency_unit_status", "emergencyUnitStatus must be open, busy, divert, closed, or unknown"
	}
	if len(request.Notes) > 700 || unsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 700 safe characters or fewer"
	}
	if request.Source == "" {
		request.Source = "manual"
	}
	if len(request.Source) > 80 || unsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || unsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeHospitalCapacityImport(request hospitalCapacityImportRequest) (hospitalCapacityImportRequest, string, string) {
	request.Source = normalizeToken(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)
	if request.Source == "" {
		request.Source = "fixture_adapter"
	}
	if request.SourceRef == "" {
		request.SourceRef = "hospital-capacity-feed"
	}
	if len(request.Source) > 80 || unsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || unsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	for index := range request.Records {
		record := &request.Records[index]
		record.FacilityID = strings.TrimSpace(record.FacilityID)
		record.EmergencyCapacity = normalizeToken(record.EmergencyCapacity)
		record.EmergencyUnitStatus = normalizeToken(record.EmergencyUnitStatus)
		record.Notes = strings.TrimSpace(record.Notes)
		if record.FacilityID == "" || len(record.FacilityID) > 128 || unsafeText(record.FacilityID) {
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
		if !allowedEmergencyCapacityStatuses[record.EmergencyCapacity] {
			return request, "invalid_emergency_capacity", "emergencyCapacity must be available, limited, full, offline, or unknown"
		}
		if record.EmergencyUnitStatus != "" && !allowedEmergencyUnitStatuses[record.EmergencyUnitStatus] {
			return request, "invalid_emergency_unit_status", "emergencyUnitStatus must be open, busy, divert, closed, or unknown"
		}
		if len(record.Notes) > 700 || unsafeText(record.Notes) {
			return request, "invalid_notes", "notes must be 700 safe characters or fewer"
		}
	}
	return request, "", ""
}

func parseLocation(w http.ResponseWriter, r *http.Request) (coordinates, bool) {
	latText := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngText := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latText == "" || lngText == "" {
		writeError(w, http.StatusBadRequest, "missing_coordinates", "lat and lng query parameters are required")
		return coordinates{}, false
	}

	lat, latErr := strconv.ParseFloat(latText, 64)
	lng, lngErr := strconv.ParseFloat(lngText, 64)
	if latErr != nil || lngErr != nil {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat and lng must be valid decimal coordinates")
		return coordinates{}, false
	}

	location := coordinates{Lat: lat, Lng: lng}
	if !validCoordinates(location) {
		writeError(w, http.StatusBadRequest, "invalid_coordinates", "lat must be between -90 and 90 and lng must be between -180 and 180")
		return coordinates{}, false
	}
	return location, true
}

func requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (authorityContext, bool) {
	ctx := authorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		log.Printf("WARN shelter-service authority_context_missing requestId=%s path=%s", ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return authorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN shelter-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for shelter updates")
		return authorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		log.Printf("WARN shelter-service authority_forbidden actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to update shelter capacity")
		return authorityContext{}, false
	}
	return ctx, true
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func optionalDecodeJSON(r *http.Request, target any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	return decodeJSON(r, target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR shelter-service write_json_response_failed error=%v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsFromEnv()
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

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func statusForCode(code string) int {
	if code == "not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func copyShelters(source []shelterRecord) []shelterRecord {
	shelters := make([]shelterRecord, 0, len(source))
	for _, shelter := range source {
		shelter.Facilities = append([]string{}, shelter.Facilities...)
		shelters = append(shelters, shelter)
	}
	return shelters
}

func copyRecovery(source []recoverySupportLocation) []recoverySupportLocation {
	locations := make([]recoverySupportLocation, 0, len(source))
	for _, item := range source {
		item.Services = append([]string{}, item.Services...)
		locations = append(locations, item)
	}
	return locations
}

func copyHospitals(source []hospitalCapacityRecord) []hospitalCapacityRecord {
	facilities := make([]hospitalCapacityRecord, 0, len(source))
	for _, facility := range source {
		facilities = append(facilities, copyHospital(facility))
	}
	return facilities
}

func copyHospital(facility hospitalCapacityRecord) hospitalCapacityRecord {
	facility.Services = append([]string{}, facility.Services...)
	return facility
}

func shelterStatusRank(status string) int {
	switch status {
	case "open":
		return 0
	case "unknown":
		return 1
	case "full":
		return 2
	case "closed":
		return 3
	default:
		return 4
	}
}

func hospitalCapacityRank(status string) int {
	switch status {
	case "available":
		return 0
	case "limited":
		return 1
	case "unknown":
		return 2
	case "full":
		return 3
	case "offline":
		return 4
	default:
		return 5
	}
}

func statusForOccupancy(capacity int, occupancy int, fallback string) string {
	if capacity > 0 && occupancy >= capacity {
		return "full"
	}
	if fallback == "full" && occupancy < capacity {
		return "open"
	}
	if fallback == "" {
		return "open"
	}
	return fallback
}

func hospitalCapacityFromBeds(totalBeds int, availableBeds int, fallback string) string {
	if totalBeds <= 0 {
		if fallback == "" {
			return "unknown"
		}
		return fallback
	}
	if availableBeds <= 0 {
		return "full"
	}
	if float64(availableBeds)/float64(totalBeds) <= 0.1 {
		return "limited"
	}
	return "available"
}

func withHospitalStaleness(facility hospitalCapacityRecord, now time.Time) hospitalCapacityRecord {
	facility.Stale = false
	facility.StaleReason = ""
	if facility.UpdatedAt.IsZero() {
		facility.Stale = true
		facility.StaleReason = "capacity timestamp missing"
		return facility
	}
	if now.Sub(facility.UpdatedAt) > hospitalCapacityStaleAfter {
		facility.Stale = true
		facility.StaleReason = "capacity update older than 30 minutes"
	}
	return facility
}

func defaultHospitalCapacityFixture() []hospitalCapacityFixtureRecord {
	return []hospitalCapacityFixtureRecord{
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
			OxygenAvailable:        boolPtr(true),
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
			OxygenAvailable:        boolPtr(true),
			Notes:                  "Fixture adapter reports heavy emergency load.",
		},
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

func distanceMeters(a coordinates, b coordinates) float64 {
	lat1 := degreesToRadians(a.Lat)
	lat2 := degreesToRadians(b.Lat)
	deltaLat := degreesToRadians(b.Lat - a.Lat)
	deltaLng := degreesToRadians(b.Lng - a.Lng)

	h := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func unsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

func normalizeToken(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(value)), "-", "_"), " ", "_")
}

func containsNormalized(values []string, needle string) bool {
	needle = normalizeToken(needle)
	for _, value := range values {
		if normalizeToken(value) == needle {
			return true
		}
	}
	return false
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

type boundingBox struct {
	MinLat float64
	MinLng float64
	MaxLat float64
	MaxLng float64
}

func parseBBox(value string) (*boundingBox, bool) {
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
	return &boundingBox{MinLat: minLat, MinLng: minLng, MaxLat: maxLat, MaxLng: maxLng}, true
}

func pointInBBox(location coordinates, box boundingBox) bool {
	return location.Lat >= box.MinLat && location.Lat <= box.MaxLat &&
		location.Lng >= box.MinLng && location.Lng <= box.MaxLng
}

type reliefPointFilter struct {
	Status       string
	Type         string
	Location     *coordinates
	RadiusMeters float64
	BBox         *boundingBox
	Limit        int
}

func parseReliefPointFilter(r *http.Request) reliefPointFilter {
	query := r.URL.Query()
	filter := reliefPointFilter{
		Status:       normalizeToken(query.Get("status")),
		Type:         normalizeToken(query.Get("type")),
		RadiusMeters: nearbySearchMeters,
	}
	if latText := strings.TrimSpace(query.Get("lat")); latText != "" {
		if lngText := strings.TrimSpace(query.Get("lng")); lngText != "" {
			lat, latErr := strconv.ParseFloat(latText, 64)
			lng, lngErr := strconv.ParseFloat(lngText, 64)
			if latErr == nil && lngErr == nil {
				location := coordinates{Lat: lat, Lng: lng}
				if validCoordinates(location) {
					filter.Location = &location
				}
			}
		}
	}
	if radiusText := strings.TrimSpace(query.Get("radius")); radiusText != "" {
		if radius, err := strconv.ParseFloat(radiusText, 64); err == nil && radius > 0 {
			filter.RadiusMeters = radius
		}
	}
	if limitText := strings.TrimSpace(query.Get("limit")); limitText != "" {
		if limit, err := strconv.Atoi(limitText); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if bboxValue := strings.TrimSpace(query.Get("bbox")); bboxValue != "" {
		if box, ok := parseBBox(bboxValue); ok {
			filter.BBox = box
		}
	}
	return filter
}

func normalizeCreateReliefPoint(request createReliefPointRequest) (createReliefPointRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = normalizeToken(request.Type)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Address = strings.TrimSpace(request.Address)
	request.Contact = strings.TrimSpace(request.Contact)
	request.OperatingHours = strings.TrimSpace(request.OperatingHours)
	request.Eligibility = strings.TrimSpace(request.Eligibility)
	request.Schedule = strings.TrimSpace(request.Schedule)
	request.Status = normalizeToken(request.Status)
	request.Source = normalizeToken(request.Source)
	request.SourceRef = strings.TrimSpace(request.SourceRef)

	if request.Name == "" || len(request.Name) > 200 || unsafeText(request.Name) {
		return request, "invalid_name", "name is required and must be 200 safe characters or fewer"
	}
	if !allowedReliefPointTypes[request.Type] {
		return request, "invalid_type", "type must be food, water, medical, hygiene, blankets, cash, or mixed"
	}
	if !validCoordinates(request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	if request.Status == "" {
		request.Status = "open"
	}
	if !allowedReliefPointStatuses[request.Status] {
		return request, "invalid_status", "status must be open, limited, closed, or paused"
	}
	if len(request.Region) > 100 || unsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || unsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if len(request.Address) > 300 || unsafeText(request.Address) {
		return request, "invalid_address", "address must be 300 safe characters or fewer"
	}
	if len(request.Contact) > 100 || unsafeText(request.Contact) {
		return request, "invalid_contact", "contact must be 100 safe characters or fewer"
	}
	if len(request.OperatingHours) > 100 || unsafeText(request.OperatingHours) {
		return request, "invalid_operating_hours", "operatingHours must be 100 safe characters or fewer"
	}
	if len(request.Eligibility) > 700 || unsafeText(request.Eligibility) {
		return request, "invalid_eligibility", "eligibility must be 700 safe characters or fewer"
	}
	if len(request.Schedule) > 200 || unsafeText(request.Schedule) {
		return request, "invalid_schedule", "schedule must be 200 safe characters or fewer"
	}
	if request.Source == "" {
		request.Source = "manual"
	}
	if len(request.Source) > 80 || unsafeText(request.Source) {
		return request, "invalid_source", "source must be 80 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || unsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	request.StockCategories = normalizeStockCategories(request.StockCategories)
	return request, "", ""
}

func normalizeUpdateReliefPoint(request updateReliefPointRequest) (updateReliefPointRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = normalizeToken(request.Type)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Address = strings.TrimSpace(request.Address)
	request.Contact = strings.TrimSpace(request.Contact)
	request.OperatingHours = strings.TrimSpace(request.OperatingHours)
	request.Eligibility = strings.TrimSpace(request.Eligibility)
	request.Schedule = strings.TrimSpace(request.Schedule)
	request.Status = normalizeToken(request.Status)
	request.SourceRef = strings.TrimSpace(request.SourceRef)

	if request.Name != "" && (len(request.Name) > 200 || unsafeText(request.Name)) {
		return request, "invalid_name", "name must be 200 safe characters or fewer"
	}
	if request.Type != "" && !allowedReliefPointTypes[request.Type] {
		return request, "invalid_type", "type must be food, water, medical, hygiene, blankets, cash, or mixed"
	}
	if request.Location != nil && !validCoordinates(*request.Location) {
		return request, "invalid_location", "location must contain valid lat and lng values"
	}
	if request.Status != "" && !allowedReliefPointStatuses[request.Status] {
		return request, "invalid_status", "status must be open, limited, closed, or paused"
	}
	if len(request.Region) > 100 || unsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || unsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if len(request.Address) > 300 || unsafeText(request.Address) {
		return request, "invalid_address", "address must be 300 safe characters or fewer"
	}
	if len(request.Contact) > 100 || unsafeText(request.Contact) {
		return request, "invalid_contact", "contact must be 100 safe characters or fewer"
	}
	if len(request.OperatingHours) > 100 || unsafeText(request.OperatingHours) {
		return request, "invalid_operating_hours", "operatingHours must be 100 safe characters or fewer"
	}
	if len(request.Eligibility) > 700 || unsafeText(request.Eligibility) {
		return request, "invalid_eligibility", "eligibility must be 700 safe characters or fewer"
	}
	if len(request.Schedule) > 200 || unsafeText(request.Schedule) {
		return request, "invalid_schedule", "schedule must be 200 safe characters or fewer"
	}
	if len(request.SourceRef) > 120 || unsafeText(request.SourceRef) {
		return request, "invalid_source_ref", "sourceRef must be 120 safe characters or fewer"
	}
	request.StockCategories = normalizeStockCategories(request.StockCategories)
	return request, "", ""
}

func normalizeStockCategories(categories []reliefStockCategory) []reliefStockCategory {
	result := make([]reliefStockCategory, 0, len(categories))
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

func copyReliefPoints(source []reliefPointRecord) []reliefPointRecord {
	reliefPoints := make([]reliefPointRecord, 0, len(source))
	for _, point := range source {
		point.StockCategories = copyStockCategories(point.StockCategories)
		reliefPoints = append(reliefPoints, point)
	}
	return reliefPoints
}

func copyStockCategories(source []reliefStockCategory) []reliefStockCategory {
	categories := make([]reliefStockCategory, len(source))
	copy(categories, source)
	return categories
}

func stockCategoriesEqual(a, b []reliefStockCategory) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Category != b[i].Category || a[i].Quantity != b[i].Quantity || a[i].Unit != b[i].Unit {
			return false
		}
	}
	return true
}

func reliefPointStatusRank(status string) int {
	switch status {
	case "open":
		return 0
	case "limited":
		return 1
	case "paused":
		return 2
	case "closed":
		return 3
	default:
		return 4
	}
}
