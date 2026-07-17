package store

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/models"
)

func seedDatasets(now time.Time) map[string]models.Dataset {
	return map[string]models.Dataset{
		"dataset_flood_reports_2026": {
			ID:                  "dataset_flood_reports_2026",
			Title:               "Flood incident reports — Greater Accra 2026",
			Description:         "Anonymized and aggregated flood incident reports for public awareness, research, and planning in Greater Accra.",
			Category:            models.OpenDataCategoryFlood,
			License:             models.LicenseCreativeCommonsBY40,
			UpdateFrequency:     models.UpdateFrequencyDaily,
			PrivacyReviewStatus: models.PrivacyReviewApproved,
			AnonymizationLevel:  models.AnonymizationAggregated,
			Metadata: models.DatasetMetadata{
				Publisher:          "NADAA Ghana / NADMO",
				ContactEmail:       "opendata@nadaa.gov.gh",
				RegionCoverage:     []string{"Greater Accra"},
				TemporalCoverage:   "2026-01-01/2026-12-31",
				SpatialResolution:  "district",
				Keywords:           []string{"flood", "incident", "accra", "risk"},
				SourceSystems:      []string{"incident-service"},
				AnonymizationNotes: "Reporter identity removed; locations rounded to district centroids; counts aggregated by day and district.",
			},
			SampleRows: []map[string]any{
				{"date": "2026-07-01", "district": "Accra Metropolitan", "reportCount": 12, "maxUrgency": "high", "injuriesReported": false},
				{"date": "2026-07-02", "district": "Tema Metropolitan", "reportCount": 7, "maxUrgency": "moderate", "injuriesReported": true},
			},
			Columns: []models.DatasetColumn{
				{Name: "date", Type: "date", Description: "Aggregation date", Nullable: false},
				{Name: "district", Type: "string", Description: "District name", Nullable: false},
				{Name: "reportCount", Type: "integer", Description: "Number of anonymized reports", Nullable: false},
				{Name: "maxUrgency", Type: "string", Description: "Highest urgency in bucket", Nullable: false},
				{Name: "injuriesReported", Type: "boolean", Description: "Whether any bucketed report indicated injuries", Nullable: false},
			},
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		},
		"dataset_road_closures_active": {
			ID:                  "dataset_road_closures_active",
			Title:               "Active road closures",
			Description:         "Currently active road closures, detours, and severity classifications for public navigation support.",
			Category:            models.OpenDataCategoryRoadClosure,
			License:             models.LicenseOpenDataCommonsODCOpen,
			UpdateFrequency:     models.UpdateFrequencyHourly,
			PrivacyReviewStatus: models.PrivacyReviewApproved,
			AnonymizationLevel:  models.AnonymizationNone,
			Metadata: models.DatasetMetadata{
				Publisher:         "NADAA Ghana / Ghana Police Service / Department of Urban Roads",
				ContactEmail:      "opendata@nadaa.gov.gh",
				RegionCoverage:    []string{"Greater Accra", "Ashanti", "Western"},
				TemporalCoverage:  "ongoing",
				SpatialResolution: "road segment",
				Keywords:          []string{"road", "closure", "detour", "traffic"},
				SourceSystems:     []string{"road-closure-service"},
			},
			SampleRows: []map[string]any{
				{"roadName": "Kaneshie Market Road", "status": "active", "severity": "high", "detourNote": "Use Graphic Road", "district": "Accra Metropolitan"},
			},
			Columns: []models.DatasetColumn{
				{Name: "roadName", Type: "string", Nullable: false},
				{Name: "status", Type: "string", Nullable: false},
				{Name: "severity", Type: "string", Nullable: false},
				{Name: "detourNote", Type: "string", Nullable: true},
				{Name: "district", Type: "string", Nullable: false},
			},
			CreatedAt: now.Add(-60 * 24 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		"dataset_shelter_occupancy": {
			ID:                  "dataset_shelter_occupancy",
			Title:               "Shelter occupancy summary",
			Description:         "Aggregated shelter occupancy, capacity, and status by district. No personal information is included.",
			Category:            models.OpenDataCategoryShelter,
			License:             models.LicenseCreativeCommonsBY40,
			UpdateFrequency:     models.UpdateFrequencyHourly,
			PrivacyReviewStatus: models.PrivacyReviewApproved,
			AnonymizationLevel:  models.AnonymizationAggregated,
			Metadata: models.DatasetMetadata{
				Publisher:          "NADAA Ghana / NADMO",
				ContactEmail:       "opendata@nadaa.gov.gh",
				RegionCoverage:     []string{"Greater Accra"},
				TemporalCoverage:   "ongoing",
				SpatialResolution:  "facility",
				Keywords:           []string{"shelter", "occupancy", "capacity", "evacuation"},
				SourceSystems:      []string{"shelter-service"},
				AnonymizationNotes: "Individual shelter registrations not included; only aggregate occupancy counts.",
			},
			SampleRows: []map[string]any{
				{"shelterName": "Accra Metro Assembly Shelter", "district": "Accra Metropolitan", "capacity": 450, "currentOccupancy": 116, "status": "open"},
			},
			Columns: []models.DatasetColumn{
				{Name: "shelterName", Type: "string", Nullable: false},
				{Name: "district", Type: "string", Nullable: false},
				{Name: "capacity", Type: "integer", Nullable: false},
				{Name: "currentOccupancy", Type: "integer", Nullable: false},
				{Name: "status", Type: "string", Nullable: false},
			},
			CreatedAt: now.Add(-45 * 24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
		},
		"dataset_flood_simulation_cells": {
			ID:                  "dataset_flood_simulation_cells",
			Title:               "Flood simulation cell outputs",
			Description:         "Aggregated flood simulation cell probabilities and severity bands for research and planning. Decision-support outputs only.",
			Category:            models.OpenDataCategoryRisk,
			License:             models.LicenseCreativeCommonsBY40,
			UpdateFrequency:     models.UpdateFrequencyDaily,
			PrivacyReviewStatus: models.PrivacyReviewApproved,
			AnonymizationLevel:  models.AnonymizationAggregated,
			Metadata: models.DatasetMetadata{
				Publisher:          "NADAA Ghana / Hydrological Services",
				ContactEmail:       "opendata@nadaa.gov.gh",
				RegionCoverage:     []string{"Greater Accra"},
				TemporalCoverage:   "2026-07-01/2026-12-31",
				SpatialResolution:  "500m grid",
				Keywords:           []string{"flood", "simulation", "risk", "grid"},
				SourceSystems:      []string{"ml-service"},
				AnonymizationNotes: "Grid-level aggregates; no structure or household identifiers.",
			},
			SampleRows: []map[string]any{
				{"cellId": "cell_001", "region": "Greater Accra", "district": "Accra Metropolitan", "probability": 0.82, "severity": "severe", "depthBand": "0.5-1.0m"},
			},
			Columns: []models.DatasetColumn{
				{Name: "cellId", Type: "string", Nullable: false},
				{Name: "region", Type: "string", Nullable: false},
				{Name: "district", Type: "string", Nullable: false},
				{Name: "probability", Type: "number", Nullable: false},
				{Name: "severity", Type: "string", Nullable: false},
				{Name: "depthBand", Type: "string", Nullable: true},
			},
			CreatedAt: now.Add(-10 * 24 * time.Hour),
			UpdatedAt: now.Add(-6 * time.Hour),
		},
		"dataset_raw_incident_feed": {
			ID:                  "dataset_raw_incident_feed",
			Title:               "Raw incident feed — restricted access",
			Description:         "Detailed incident records for approved research partners. Requires access agreement and privacy review.",
			Category:            models.OpenDataCategoryIncident,
			License:             models.LicenseGhanaOpenGovernment,
			UpdateFrequency:     models.UpdateFrequencyDaily,
			PrivacyReviewStatus: models.PrivacyReviewPending,
			AnonymizationLevel:  models.AnonymizationAnonymized,
			AccessRestriction:   "Requires approved access request and data-use agreement.",
			Metadata: models.DatasetMetadata{
				Publisher:          "NADAA Ghana / NADMO",
				ContactEmail:       "opendata@nadaa.gov.gh",
				RegionCoverage:     []string{"Greater Accra", "Ashanti"},
				TemporalCoverage:   "2026-01-01/2026-12-31",
				SpatialResolution:  "exact coordinates (anonymized)",
				Keywords:           []string{"incident", "response", "research"},
				SourceSystems:      []string{"incident-service"},
				AnonymizationNotes: "Reporter identity anonymized; exact locations jittered; review pending before publication.",
			},
			SampleRows: []map[string]any{
				{"reference": "NADAA-20260701-001", "type": "flood", "district": "Accra Metropolitan", "status": "verified", "peopleAffected": 12},
			},
			Columns: []models.DatasetColumn{
				{Name: "reference", Type: "string", Nullable: false},
				{Name: "type", Type: "string", Nullable: false},
				{Name: "district", Type: "string", Nullable: false},
				{Name: "status", Type: "string", Nullable: false},
				{Name: "peopleAffected", Type: "integer", Nullable: false},
			},
			CreatedAt: now.Add(-20 * 24 * time.Hour),
			UpdatedAt: now,
		},
	}
}

func generateID(prefix string, now time.Time) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s-%d", prefix, now.UnixNano()))
	return fmt.Sprintf("%s_%x", prefix, sum[:8])
}

func estimateSize(datasetID, format string) int64 {
	base := int64(len(datasetID) * 1024)
	switch format {
	case "csv":
		return base
	case "json":
		return base * 2
	case "parquet":
		return base / 2
	default:
		return base
	}
}
