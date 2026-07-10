package store

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/models"
)

// seedData returns fixture campaigns and their mocked metrics.
func seedData(now time.Time) ([]models.Campaign, map[string][]models.CampaignMetric) {
	window := models.CampaignPublishWindow{
		StartsAt: now.Add(-30 * 24 * time.Hour),
		EndsAt:   now.Add(60 * 24 * time.Hour),
	}

	campaigns := []models.Campaign{
		{
			ID:            "campaign_001",
			Title:         "Rainy season flood preparedness",
			HazardType:    "flood",
			TargetRegions: []string{"Greater Accra", "Ashanti", "Western"},
			Languages:     []string{"en", "tw"},
			ContentBlocks: []models.CampaignContentBlock{
				{
					Type:  "article",
					Title: "Clear drains before the rains",
					Body:  "Remove loose rubbish from gutters and drains around your home. Report blocked public drains to your district assembly or NADMO.",
				},
				{
					Type:  "checklist",
					Title: "Flood-ready checklist",
					Items: []string{"Know your nearest shelter", "Keep documents in a waterproof bag", "Prepare a family contact plan", "Charge phones and power banks"},
				},
			},
			PublishWindow:  window,
			Status:         "published",
			LinkedGuideIDs: []string{"guide_flood_before_en", "guide_flood_during_en"},
			CreatedBy:      "usr_nadmo_accra",
			UpdatedBy:      "usr_nadmo_accra",
			CreatedAt:      now.Add(-48 * time.Hour),
			UpdatedAt:      now.Add(-24 * time.Hour),
		},
		{
			ID:            "campaign_002",
			Title:         "Harmattan fire safety",
			HazardType:    "fire",
			TargetRegions: []string{"Northern", "Upper East", "Upper West"},
			Languages:     []string{"en", "ha"},
			ContentBlocks: []models.CampaignContentBlock{
				{
					Type:  "article",
					Title: "Dry-season fire prevention",
					Body:  "Keep bush fires away from homes, store fuel safely, and ensure children cannot reach open flames or matches.",
				},
				{
					Type:     "media",
					Title:    "What to do if fire spreads",
					Body:     "Watch the Ghana National Fire Service safety video.",
					MediaURL: "https://nadaa.example/media/harmattan-fire-safety.mp4",
				},
			},
			PublishWindow:  window,
			Status:         "published",
			LinkedGuideIDs: []string{"guide_fire_during_en"},
			CreatedBy:      "usr_nadmo_north",
			UpdatedBy:      "usr_nadmo_north",
			CreatedAt:      now.Add(-72 * time.Hour),
			UpdatedAt:      now.Add(-36 * time.Hour),
		},
	}

	metrics := map[string][]models.CampaignMetric{
		"campaign_001": {
			{ID: "metric_001_1", CampaignID: "campaign_001", Date: now.Add(-2 * 24 * time.Hour), Reach: 12400, Engagement: 1180},
			{ID: "metric_001_2", CampaignID: "campaign_001", Date: now.Add(-1 * 24 * time.Hour), Reach: 15100, Engagement: 1420},
		},
		"campaign_002": {
			{ID: "metric_002_1", CampaignID: "campaign_002", Date: now.Add(-2 * 24 * time.Hour), Reach: 8200, Engagement: 740},
			{ID: "metric_002_2", CampaignID: "campaign_002", Date: now.Add(-1 * 24 * time.Hour), Reach: 9400, Engagement: 890},
		},
	}

	return campaigns, metrics
}

// seedTemplates returns reusable seasonal campaign templates.
func seedTemplates() []models.CampaignTemplate {
	return []models.CampaignTemplate{
		{
			ID:         "template_rainy_flood",
			Name:       "Rainy season flood preparedness",
			HazardType: "flood",
			Season:     "rainy",
			DefaultContent: []models.CampaignContentBlock{
				{Type: "article", Title: "Clear drains before the rains", Body: "Remove loose rubbish from gutters and drains around your home."},
				{Type: "checklist", Title: "Flood-ready checklist", Items: []string{"Know your nearest shelter", "Keep documents dry", "Prepare a family contact plan"}},
			},
		},
		{
			ID:         "template_harmattan_fire",
			Name:       "Harmattan fire safety",
			HazardType: "fire",
			Season:     "harmattan",
			DefaultContent: []models.CampaignContentBlock{
				{Type: "article", Title: "Dry-season fire prevention", Body: "Keep bush fires away from homes and store fuel safely."},
				{Type: "checklist", Title: "Fire safety checklist", Items: []string{"Check electrical wiring", "Keep matches from children", "Plan two ways out"}},
			},
		},
		{
			ID:         "template_dry_disease",
			Name:       "Dry season disease prevention",
			HazardType: "disease_outbreak",
			Season:     "dry",
			DefaultContent: []models.CampaignContentBlock{
				{Type: "article", Title: "Stop the spread", Body: "Wash hands, cover coughs, and seek care early if symptoms appear."},
				{Type: "checklist", Title: "Hygiene checklist", Items: []string{"Wash hands regularly", "Use clean water", "Isolate when sick"}},
			},
		},
	}
}
