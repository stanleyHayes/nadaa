package store

import "github.com/stanleyHayes/nadaa/services/risk-service/internal/models"

func seedRiskZones() []models.RiskZone {
	return []models.RiskZone{
		{
			ID:          "00000000-0000-0000-0000-000000000401",
			HazardType:  "flood",
			RiskLevel:   "severe",
			Bounds:      models.BoundingBox{MinLat: 5.530, MaxLat: 5.590, MinLng: -0.230, MaxLng: -0.160},
			Probability: 0.86,
			Explanation: "Low-lying Accra sample zone with historical flood reports and rainfall sensitivity.",
		},
		{
			ID:          "00000000-0000-0000-0000-000000000402",
			HazardType:  "fire",
			RiskLevel:   "moderate",
			Bounds:      models.BoundingBox{MinLat: 5.540, MaxLat: 5.610, MinLng: -0.210, MaxLng: -0.140},
			Probability: 0.38,
			Explanation: "Dense commercial area sample zone.",
		},
	}
}

func seedShelters() []models.ShelterSummary {
	return []models.ShelterSummary{
		{
			ID:               "00000000-0000-0000-0000-000000000301",
			Name:             "Accra Metro Assembly Shelter",
			Location:         models.Coordinates{Lat: 5.560, Lng: -0.200},
			Capacity:         450,
			CurrentOccupancy: 116,
			Contact:          "112",
			Status:           "open",
			Facilities:       []string{"water", "first_aid", "accessible_entry", "family_area"},
		},
		{
			ID:               "00000000-0000-0000-0000-000000000302",
			Name:             "Osu Community Hall",
			Location:         models.Coordinates{Lat: 5.550, Lng: -0.180},
			Capacity:         220,
			CurrentOccupancy: 34,
			Contact:          "112",
			Status:           "open",
			Facilities:       []string{"water", "first_aid"},
		},
	}
}

func seedFacilities() []models.FacilitySummary {
	return []models.FacilitySummary{
		{
			ID:       "00000000-0000-0000-0000-000000000101",
			Name:     "NADMO Accra Metro",
			Type:     "nadmo",
			Location: models.Coordinates{Lat: 5.560, Lng: -0.200},
			Region:   "Greater Accra",
			District: "Accra Metropolitan",
			Contact:  "112",
		},
		{
			ID:       "00000000-0000-0000-0000-000000000102",
			Name:     "Ghana National Fire Service Accra",
			Type:     "fire",
			Location: models.Coordinates{Lat: 5.565, Lng: -0.185},
			Region:   "Greater Accra",
			District: "Accra Metropolitan",
			Contact:  "112",
		},
		{
			ID:       "00000000-0000-0000-0000-000000000103",
			Name:     "National Ambulance Service Accra",
			Type:     "ambulance",
			Location: models.Coordinates{Lat: 5.555, Lng: -0.190},
			Region:   "Greater Accra",
			District: "Accra Metropolitan",
			Contact:  "112",
		},
	}
}

func seedHistoricalReports() []models.HistoricalReport {
	return []models.HistoricalReport{
		{
			HazardType: "flood",
			Location:   models.Coordinates{Lat: 5.6037, Lng: -0.1870},
			Severity:   "high",
		},
	}
}
