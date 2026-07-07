package store

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
)

func seedClosures(now time.Time) []models.RoadClosureRecord {
	return []models.RoadClosureRecord{
		{
			ID:        "road_closure_001",
			RoadName:  "Accra New Town Road",
			Reason:    "Flooding",
			Status:    "active",
			Severity:  "high",
			Source:    "manual",
			SourceRef: "nadmo-accra-ops",
			Geometry: models.LineStringGeometry{
				Type: "LineString",
				Coordinates: [][]float64{
					{-0.205, 5.570},
					{-0.190, 5.580},
				},
			},
			ValidFrom:  now.Add(-2 * time.Hour),
			ValidTo:    timePtr(now.Add(6 * time.Hour)),
			DetourNote: "Use Kanda Highway",
			CreatedBy:  "usr_dispatcher_001",
			UpdatedBy:  "usr_dispatcher_001",
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-2 * time.Hour),
		},
		{
			ID:        "road_closure_002",
			RoadName:  "Kaneshie Market Road",
			Reason:    "Debris and flood water",
			Status:    "active",
			Severity:  "severe",
			Source:    "manual",
			SourceRef: "ghana-police",
			Geometry: models.LineStringGeometry{
				Type: "LineString",
				Coordinates: [][]float64{
					{-0.248, 5.566},
					{-0.240, 5.568},
					{-0.235, 5.564},
				},
			},
			ValidFrom:  now.Add(-1 * time.Hour),
			ValidTo:    timePtr(now.Add(12 * time.Hour)),
			DetourNote: "Use Mallam Road",
			CreatedBy:  "usr_dispatcher_002",
			UpdatedBy:  "usr_dispatcher_002",
			CreatedAt:  now.Add(-1 * time.Hour),
			UpdatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:        "road_closure_003",
			RoadName:  "Labone Street",
			Reason:    "Scheduled drainage maintenance",
			Status:    "scheduled",
			Severity:  "low",
			Source:    "manual",
			SourceRef: "district-accra",
			Geometry: models.LineStringGeometry{
				Type: "LineString",
				Coordinates: [][]float64{
					{-0.183, 5.553},
					{-0.178, 5.555},
				},
			},
			ValidFrom:  now.Add(24 * time.Hour),
			ValidTo:    timePtr(now.Add(48 * time.Hour)),
			DetourNote: "Use Cantonments Road",
			CreatedBy:  "usr_district_officer_001",
			UpdatedBy:  "usr_district_officer_001",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
