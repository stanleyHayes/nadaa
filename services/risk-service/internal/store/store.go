package store

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/stanleyHayes/nadaa/services/risk-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/risk-service/internal/utils"
)

const (
	nearbyShelterRadius     = 30000.0
	nearbyFacilityRadius    = 30000.0
	nearbyRiskZoneThreshold = 2500.0
	recentReportThreshold   = 3000.0
)

// Store is the persistence interface for risk data.
type Store interface {
	AreaRisk(location models.Coordinates) models.RiskResponse
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu                sync.RWMutex
	riskZones         []models.RiskZone
	shelters          []models.ShelterSummary
	facilities        []models.FacilitySummary
	historicalReports []models.HistoricalReport
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore() Store {
	return &MemoryStore{
		riskZones:         seedRiskZones(),
		shelters:          seedShelters(),
		facilities:        seedFacilities(),
		historicalReports: seedHistoricalReports(),
	}
}

// AreaRisk returns the full risk assessment for a location.
func (m *MemoryStore) AreaRisk(location models.Coordinates) models.RiskResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	risks := []models.RiskSummary{m.floodRisk(location)}
	if fireRisk, ok := m.fireRisk(location); ok {
		risks = append(risks, fireRisk)
	}

	sort.Slice(risks, func(i, j int) bool {
		return utils.RiskRank(risks[i].Level) > utils.RiskRank(risks[j].Level)
	})

	overall := "low"
	for _, risk := range risks {
		if utils.RiskRank(risk.Level) > utils.RiskRank(overall) {
			overall = risk.Level
		}
	}

	return models.RiskResponse{
		Location:           inferLocation(location),
		OverallRisk:        overall,
		Risks:              risks,
		NearestShelters:    m.nearestShelters(location),
		NearbyFacilities:   m.nearbyFacilities(location),
		RecommendedActions: utils.RecommendedActions(overall, risks[0].Level),
	}
}

func (m *MemoryStore) floodRisk(location models.Coordinates) models.RiskSummary {
	for _, zone := range m.riskZones {
		if zone.HazardType != "flood" {
			continue
		}
		if zone.Bounds.Contains(location) {
			return models.RiskSummary{
				Type:        "flood",
				Level:       zone.RiskLevel,
				Probability: zone.Probability,
				Reason:      zone.Explanation,
			}
		}
	}

	nearestZoneDistance := math.MaxFloat64
	for _, zone := range m.riskZones {
		if zone.HazardType != "flood" {
			continue
		}
		nearestZoneDistance = math.Min(nearestZoneDistance, zone.Bounds.DistanceMeters(location))
	}

	nearestReportDistance := math.MaxFloat64
	nearestReportSeverity := ""
	for _, report := range m.historicalReports {
		if report.HazardType != "flood" {
			continue
		}
		distance := models.HaversineMeters(location, report.Location)
		if distance < nearestReportDistance {
			nearestReportDistance = distance
			nearestReportSeverity = report.Severity
		}
	}

	if nearestZoneDistance <= nearbyRiskZoneThreshold && nearestReportDistance <= recentReportThreshold {
		return models.RiskSummary{
			Type:        "flood",
			Level:       "high",
			Probability: 0.72,
			Reason:      fmt.Sprintf("Within %.0fm of a severe flood zone and %.0fm of a recent %s flood report.", nearestZoneDistance, nearestReportDistance, nearestReportSeverity),
		}
	}

	if nearestZoneDistance <= nearbyRiskZoneThreshold {
		return models.RiskSummary{
			Type:        "flood",
			Level:       "high",
			Probability: 0.64,
			Reason:      fmt.Sprintf("Within %.0fm of a severe flood-prone zone.", nearestZoneDistance),
		}
	}

	if nearestReportDistance <= recentReportThreshold {
		return models.RiskSummary{
			Type:        "flood",
			Level:       "moderate",
			Probability: 0.42,
			Reason:      fmt.Sprintf("Within %.0fm of a recent %s flood report.", nearestReportDistance, nearestReportSeverity),
		}
	}

	return models.RiskSummary{
		Type:        "flood",
		Level:       "low",
		Probability: 0.16,
		Reason:      "No active flood risk zone or recent flood report is near these coordinates in the MVP fixture set.",
	}
}

func (m *MemoryStore) fireRisk(location models.Coordinates) (models.RiskSummary, bool) {
	for _, zone := range m.riskZones {
		if zone.HazardType != "fire" {
			continue
		}
		if zone.Bounds.Contains(location) {
			return models.RiskSummary{
				Type:        "fire",
				Level:       zone.RiskLevel,
				Probability: zone.Probability,
				Reason:      zone.Explanation,
			}, true
		}
	}
	return models.RiskSummary{}, false
}

func (m *MemoryStore) nearestShelters(location models.Coordinates) []models.ShelterSummary {
	shelters := make([]models.ShelterSummary, 0, len(m.shelters))
	for _, shelter := range m.shelters {
		distance := models.HaversineMeters(location, shelter.Location)
		if distance > nearbyShelterRadius {
			continue
		}
		shelter.DistanceMeters = int(math.Round(distance))
		shelters = append(shelters, shelter)
	}

	sort.Slice(shelters, func(i, j int) bool {
		return shelters[i].DistanceMeters < shelters[j].DistanceMeters
	})
	return shelters
}

func (m *MemoryStore) nearbyFacilities(location models.Coordinates) []models.FacilitySummary {
	facilities := make([]models.FacilitySummary, 0, len(m.facilities))
	for _, facility := range m.facilities {
		distance := models.HaversineMeters(location, facility.Location)
		if distance > nearbyFacilityRadius {
			continue
		}
		facility.DistanceMeters = int(math.Round(distance))
		facilities = append(facilities, facility)
	}

	sort.Slice(facilities, func(i, j int) bool {
		return facilities[i].DistanceMeters < facilities[j].DistanceMeters
	})
	return facilities
}

func inferLocation(location models.Coordinates) string {
	if location.Lat >= 5.50 && location.Lat <= 5.66 && location.Lng >= -0.28 && location.Lng <= -0.08 {
		return "Accra Metropolitan"
	}
	if location.Lat >= 6.55 && location.Lat <= 6.80 && location.Lng >= -1.75 && location.Lng <= -1.45 {
		return "Kumasi area"
	}
	return "Selected area"
}
