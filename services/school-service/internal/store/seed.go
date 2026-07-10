package store

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/school-service/internal/models"
)

func seedSchools(now time.Time) []models.SchoolProfile {
	return []models.SchoolProfile{
		{
			ID:                "school_001",
			Name:              "Accra Methodist Primary School",
			Location:          models.Coordinates{Lat: 5.560, Lng: -0.205},
			Region:            "Greater Accra",
			District:          "Accra Metropolitan",
			Address:           "Methodist Church premises, Accra New Town",
			StudentPopulation: 650,
			EmergencyContacts: []models.EmergencyContact{
				{Name: "Mrs. Grace Osei", Role: "headteacher", Phone: "+233200000101", IsPrimary: true},
				{Name: "Mr. Kofi Asante", Role: "safety_officer", Phone: "+233200000102", IsPrimary: false},
			},
			Hazards: []string{"flood", "fire"},
			EvacuationPoints: []models.EvacuationPoint{
				{Label: "Front assembly ground", Location: models.Coordinates{Lat: 5.561, Lng: -0.206}, Capacity: 700},
				{Label: "Church hall", Location: models.Coordinates{Lat: 5.559, Lng: -0.204}, Capacity: 300},
			},
			CreatedBy: "usr_district_officer_001",
			UpdatedBy: "usr_district_officer_001",
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-2 * 24 * time.Hour),
		},
		{
			ID:                "school_002",
			Name:              "Tema Community 2 JHS",
			Location:          models.Coordinates{Lat: 5.642, Lng: -0.028},
			Region:            "Greater Accra",
			District:          "Tema Metropolitan",
			Address:           "Community 2, Tema",
			StudentPopulation: 420,
			EmergencyContacts: []models.EmergencyContact{
				{Name: "Mr. Daniel Mensah", Role: "headteacher", Phone: "+233200000201", IsPrimary: true},
			},
			Hazards: []string{"flood", "storm"},
			EvacuationPoints: []models.EvacuationPoint{
				{Label: "Main assembly ground", Location: models.Coordinates{Lat: 5.643, Lng: -0.027}, Capacity: 500},
			},
			CreatedBy: "usr_district_officer_002",
			UpdatedBy: "usr_district_officer_002",
			CreatedAt: now.Add(-45 * 24 * time.Hour),
			UpdatedAt: now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:                "school_003",
			Name:              "Korle Bu Basic School",
			Location:          models.Coordinates{Lat: 5.540, Lng: -0.220},
			Region:            "Greater Accra",
			District:          "Accra Metropolitan",
			Address:           "Korle Bu, Accra",
			StudentPopulation: 380,
			EmergencyContacts: []models.EmergencyContact{
				{Name: "Ms. Abena Frempong", Role: "headteacher", Phone: "+233200000301", IsPrimary: true},
			},
			Hazards: []string{"flood"},
			EvacuationPoints: []models.EvacuationPoint{
				{Label: "School park", Location: models.Coordinates{Lat: 5.541, Lng: -0.221}, Capacity: 400},
			},
			CreatedBy: "usr_district_officer_001",
			UpdatedBy: "usr_district_officer_001",
			CreatedAt: now.Add(-20 * 24 * time.Hour),
			UpdatedAt: now.Add(-1 * 24 * time.Hour),
		},
	}
}

func seedDrills(now time.Time) []models.DrillRecord {
	return []models.DrillRecord{
		{
			ID:           "drill_001",
			SchoolID:     "school_001",
			Date:         now.Add(-14 * 24 * time.Hour),
			Type:         "fire",
			Participants: 620,
			Notes:        "All classes evacuated in under three minutes. Improve signage for P3 block.",
			Completed:    true,
			CreatedBy:    "usr_district_officer_001",
			CreatedAt:    now.Add(-14 * 24 * time.Hour),
		},
		{
			ID:           "drill_002",
			SchoolID:     "school_001",
			Date:         now.Add(-60 * 24 * time.Hour),
			Type:         "flood",
			Participants: 580,
			Notes:        "Wet-season flood simulation; relocated youngest pupils first.",
			Completed:    true,
			CreatedBy:    "usr_district_officer_001",
			CreatedAt:    now.Add(-60 * 24 * time.Hour),
		},
		{
			ID:           "drill_003",
			SchoolID:     "school_002",
			Date:         now.Add(-21 * 24 * time.Hour),
			Type:         "storm",
			Participants: 410,
			Notes:        "Shelter-in-place drill during harmattan storm rehearsal.",
			Completed:    true,
			CreatedBy:    "usr_district_officer_002",
			CreatedAt:    now.Add(-21 * 24 * time.Hour),
		},
	}
}

func seedReadinessChecks(now time.Time) []models.ReadinessCheck {
	return []models.ReadinessCheck{
		{
			ID:          "readiness_001",
			SchoolID:    "school_001",
			CheckDate:   now.Add(-7 * 24 * time.Hour),
			RiskLevel:   "high",
			AreaRiskRef: "risk_accra_north_001",
			ChecklistItems: []models.ChecklistItem{
				{Label: "Emergency contacts updated", Checked: true, Category: "admin"},
				{Label: "Evacuation routes marked", Checked: true, Category: "planning"},
				{Label: "First aid kits stocked", Checked: false, Category: "equipment"},
				{Label: "Teachers trained on alarm protocol", Checked: true, Category: "training"},
			},
			OverallStatus: "needs_improvement",
			Notes:         "First aid kits need restocking before rainy season.",
			CheckedBy:     "usr_district_officer_001",
			CreatedAt:     now.Add(-7 * 24 * time.Hour),
			UpdatedAt:     now.Add(-7 * 24 * time.Hour),
		},
		{
			ID:          "readiness_002",
			SchoolID:    "school_002",
			CheckDate:   now.Add(-10 * 24 * time.Hour),
			RiskLevel:   "moderate",
			AreaRiskRef: "risk_tema_central_001",
			ChecklistItems: []models.ChecklistItem{
				{Label: "Emergency contacts updated", Checked: true, Category: "admin"},
				{Label: "Evacuation routes marked", Checked: true, Category: "planning"},
				{Label: "First aid kits stocked", Checked: true, Category: "equipment"},
			},
			OverallStatus: "ready",
			Notes:         "School is well prepared for the current season.",
			CheckedBy:     "usr_district_officer_002",
			CreatedAt:     now.Add(-10 * 24 * time.Hour),
			UpdatedAt:     now.Add(-10 * 24 * time.Hour),
		},
	}
}
