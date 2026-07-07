package store

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
)

func seedContracts(now time.Time) []models.IntegrationContract {
	return []models.IntegrationContract{
		newContract("gmet-rainfall-nowcast", "Ghana Meteorological Agency", "meteorological", "weather", "inbound", "GMet", "Every 15 minutes during watch/warning periods", "api_key", []string{"X-NADAA-Source", "X-NADAA-Signature"}, []models.PayloadContract{weatherPayload()}, "Imported observations keep source, observedAt, validFrom, validTo, stationId, and location.", now),
		newContract("hydro-water-level-feed", "Ghana Hydrological Authority", "hydrological", "hydrology", "inbound", "Ghana Hydrological Authority", "Every 15 minutes during rainy season, hourly otherwise", "mtls", []string{"X-NADAA-Source"}, []models.PayloadContract{hydrologyPayload()}, "Water-level records remain owned by the originating hydrological authority.", now),
		newContract("nadmo-incident-sync", "NADMO National Operations", "nadmo", "incident_sync", "outbound", "NADAA platform operator", "Near real time on verification, assignment, and closure", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []models.PayloadContract{incidentSyncPayload()}, "Manual dispatcher call and dashboard export remain the fallback if sync fails.", now),
		newContract("nadmo-alert-sync", "NADMO National Operations", "nadmo", "alert_sync", "outbound", "NADAA platform operator", "Near real time after human approval", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []models.PayloadContract{alertSyncPayload()}, "Public alert publication must continue through approved NADAA workflow if partner sync fails.", now),
		newContract("police-road-closure-feed", "Ghana Police Service", "police", "road_closure", "bidirectional", "Ghana Police Service", "On change, with hourly reconciliation", "signed_webhook", []string{"X-NADAA-Source", "X-NADAA-Signature"}, []models.PayloadContract{roadClosurePayload(), incidentSyncPayload()}, "Road closures imported from police remain source-attributed and reviewable before route use.", now),
		newContract("fire-incident-dispatch", "Ghana National Fire Service", "fire", "incident_sync", "outbound", "NADAA platform operator", "Near real time for fire, flood rescue, and electrical hazard assignments", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []models.PayloadContract{incidentSyncPayload()}, "If webhook delivery fails, dispatcher contacts fire service through 112 and records the manual handoff.", now),
		newContract("ambulance-medical-dispatch", "National Ambulance Service", "ambulance", "incident_sync", "outbound", "NADAA platform operator", "Near real time for injury and medical emergency assignments", "signed_webhook", []string{"X-NADAA-Request-ID", "X-NADAA-Signature"}, []models.PayloadContract{incidentSyncPayload()}, "If partner endpoint is down, dispatcher uses 112 and keeps incident status manual.", now),
		newContract("district-shelter-status", "District Assemblies", "district_assembly", "shelter_status", "bidirectional", "District Assembly or shelter operator", "Every 30 minutes during response, daily otherwise", "api_key", []string{"X-NADAA-Source"}, []models.PayloadContract{shelterStatusPayload()}, "Shelter updates are advisory until confirmed by authorized district or NADMO users.", now),
		newContract("hospital-capacity-feed", "Hospitals And Health Facilities", "hospital", "hospital_capacity", "inbound", "Participating health facility", "Every 30 minutes during active incidents", "api_key", []string{"X-NADAA-Source"}, []models.PayloadContract{hospitalCapacityPayload()}, "Capacity data is operationally sensitive and should be visible only to authorized responders.", now),
		newContract("utility-outage-feed", "Utilities And Power Providers", "utility", "utility_outage", "inbound", "Originating utility", "On change, with hourly reconciliation", "signed_webhook", []string{"X-NADAA-Source", "X-NADAA-Signature"}, []models.PayloadContract{utilityOutagePayload()}, "Electrical hazard and outage imports never suppress citizen reports; they enrich dispatcher context.", now),
	}
}

func newContract(id string, partner string, partnerType string, domain string, direction string, dataOwner string, cadence string, authMode string, headers []string, payloads []models.PayloadContract, notes string, now time.Time) models.IntegrationContract {
	return models.IntegrationContract{
		ID:                     id,
		Partner:                partner,
		PartnerType:            partnerType,
		Domain:                 domain,
		Direction:              direction,
		DataOwner:              dataOwner,
		Cadence:                cadence,
		Authentication:         models.Authentication{Mode: authMode, RequiredHeaders: headers, SecretScope: "environment_secret_manager"},
		Payloads:               payloads,
		FailureBehavior:        standardFailureBehavior(domain),
		SourceOfTruth:          sourceOfTruth(direction),
		FreshnessWindowMinutes: freshnessWindow(domain),
		ContactPoint:           "integration-owner@nadaa.local",
		Status:                 "mock_contract",
		Notes:                  notes,
		UpdatedAt:              now,
	}
}

func standardFailureBehavior(domain string) models.FailureBehavior {
	queue := "integration.dead_letter." + domain
	return models.FailureBehavior{
		Retryable:       true,
		MaxAttempts:     5,
		BackoffSeconds:  []int{30, 120, 300, 900, 1800},
		DeadLetterQueue: queue,
		ManualFallback:  "Record failed exchange in import job logs and continue manual incident response or alert approval.",
	}
}

func sourceOfTruth(direction string) string {
	if direction == "inbound" {
		return "originating_partner"
	}
	if direction == "outbound" {
		return "nadaa"
	}
	return "field_specific"
}

func freshnessWindow(domain string) int {
	switch domain {
	case "weather", "hydrology":
		return 30
	case "incident_sync", "alert_sync", "utility_outage", "road_closure":
		return 5
	default:
		return 60
	}
}

func seedObservations(now time.Time) []models.WeatherHydrologyObservation {
	return []models.WeatherHydrologyObservation{
		newObservation("obs_gmet_accra_001", "gmet-accra-nowcast", "rainfall_mm", 34.2, "mm", "GHA-ACC-RAIN-001", models.Coordinates{Lat: 5.6037, Lng: -0.1870}, now.Add(-15*time.Minute), now),
		newObservation("obs_gmet_accra_002", "gmet-accra-nowcast", "rainfall_mm", 42.8, "mm", "GHA-ACC-RAIN-002", models.Coordinates{Lat: 5.5600, Lng: -0.2000}, now.Add(-10*time.Minute), now),
		newObservation("obs_hydro_odaw_001", "hydro-odaw-level", "water_level_m", 1.76, "m", "GHA-ODAW-LVL-001", models.Coordinates{Lat: 5.5750, Lng: -0.2050}, now.Add(-12*time.Minute), now),
		newObservation("obs_hydro_korle_001", "hydro-korle-level", "water_level_m", 1.34, "m", "GHA-KORLE-LVL-001", models.Coordinates{Lat: 5.5400, Lng: -0.2150}, now.Add(-9*time.Minute), now),
	}
}

func newObservation(id string, source string, metric string, value float64, unit string, stationID string, location models.Coordinates, observedAt time.Time, now time.Time) models.WeatherHydrologyObservation {
	return models.WeatherHydrologyObservation{
		ID:          id,
		Source:      source,
		Metric:      metric,
		Value:       value,
		Unit:        unit,
		StationID:   stationID,
		Location:    location,
		ObservedAt:  observedAt,
		ValidFrom:   observedAt,
		ValidTo:     observedAt.Add(30 * time.Minute),
		Quality:     "fixture",
		GeneratedBy: "mock_adapter",
	}
}

func weatherPayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "WeatherObservation",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "observedAt", "validFrom", "validTo", "location.lat", "location.lng", "rainfallMm"},
		OptionalFields: []string{"stationId", "forecastWindowMinutes", "confidence"},
		PII:            "none",
		Geometry:       "Point WGS84",
		ExampleRef:     "docs/integrations.md#weather-observation",
	}
}

func hydrologyPayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "HydrologyObservation",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "observedAt", "validFrom", "validTo", "location.lat", "location.lng", "waterLevelM"},
		OptionalFields: []string{"stationId", "riverBasin", "thresholdLevelM"},
		PII:            "none",
		Geometry:       "Point WGS84",
		ExampleRef:     "docs/integrations.md#hydrology-observation",
	}
}

func incidentSyncPayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "IncidentSync",
		ContentType:    "application/json",
		RequiredFields: []string{"type", "sourceId", "reference", "hazardType", "status", "severity", "location", "targetAgencyIds", "correlationId"},
		OptionalFields: []string{"summary", "occurredAt", "mediaCount", "accessibilityNeeds"},
		PII:            "minimal_operational",
		Geometry:       "Point WGS84",
		ExampleRef:     "docs/integrations.md#incident-sync",
	}
}

func alertSyncPayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "AlertSync",
		ContentType:    "application/json",
		RequiredFields: []string{"type", "sourceId", "reference", "hazardType", "severity", "title", "message", "targetLabel", "targetAgencyIds", "correlationId"},
		OptionalFields: []string{"startsAt", "expiresAt", "recommendedAction", "evacuationRequired"},
		PII:            "none",
		Geometry:       "Target geometry summary or reference",
		ExampleRef:     "docs/integrations.md#alert-sync",
	}
}

func roadClosurePayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "RoadClosure",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "roadName", "status", "geometry", "validFrom"},
		OptionalFields: []string{"validTo", "reason", "detour"},
		PII:            "none",
		Geometry:       "LineString WGS84",
		ExampleRef:     "docs/integrations.md#road-closure",
	}
}

func shelterStatusPayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "ShelterStatus",
		ContentType:    "application/json",
		RequiredFields: []string{"shelterId", "status", "capacity", "currentOccupancy", "updatedAt"},
		OptionalFields: []string{"facilities", "contact", "needs"},
		PII:            "aggregate_only",
		Geometry:       "Point WGS84 or shelter reference",
		ExampleRef:     "docs/integrations.md#shelter-status",
	}
}

func hospitalCapacityPayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "HospitalCapacity",
		ContentType:    "application/json",
		RequiredFields: []string{"facilityId", "availableBeds", "emergencyCapacity", "updatedAt"},
		OptionalFields: []string{"traumaCapacity", "ambulanceBayStatus", "contact"},
		PII:            "aggregate_only",
		Geometry:       "Point WGS84 or facility reference",
		ExampleRef:     "docs/integrations.md#hospital-capacity",
	}
}

func utilityOutagePayload() models.PayloadContract {
	return models.PayloadContract{
		Name:           "UtilityOutage",
		ContentType:    "application/json",
		RequiredFields: []string{"source", "utilityType", "status", "area", "validFrom"},
		OptionalFields: []string{"validTo", "hazardType", "customerImpactEstimate"},
		PII:            "none",
		Geometry:       "Polygon or MultiPolygon WGS84",
		ExampleRef:     "docs/integrations.md#utility-outage",
	}
}
