package store

import (
	"context"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/utils"
)

// Store is the persistence interface for campaign data.
type Store interface {
	ListCampaigns(ctx context.Context, filters models.CampaignFilters, now time.Time) []models.Campaign
	GetCampaign(ctx context.Context, id string) (models.Campaign, bool)
	CreateCampaign(ctx context.Context, request models.CreateCampaignRequest, authority models.AuthorityContext, now time.Time) models.Campaign
	UpdateCampaign(ctx context.Context, id string, request models.UpdateCampaignRequest, authority models.AuthorityContext, now time.Time) (models.Campaign, string, string)
	ListMetrics(ctx context.Context, campaignID string, now time.Time) []models.CampaignMetric
	ListTemplates(ctx context.Context) []models.CampaignTemplate
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu              sync.RWMutex
	campaigns       []models.Campaign
	metrics         map[string][]models.CampaignMetric
	templates       []models.CampaignTemplate
	campaignCounter int
	metricCounter   int
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	seedCampaigns, seedMetrics := seedData(now)
	return &MemoryStore{
		campaigns:       seedCampaigns,
		metrics:         seedMetrics,
		templates:       seedTemplates(),
		campaignCounter: len(seedCampaigns),
		metricCounter:   countMetrics(seedMetrics),
	}
}

// ListCampaigns returns campaigns matching the provided filters.
func (m *MemoryStore) ListCampaigns(ctx context.Context, filters models.CampaignFilters, now time.Time) []models.Campaign {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]models.Campaign, 0, len(m.campaigns))
	for _, campaign := range m.campaigns {
		if filters.Status != "" && campaign.Status != filters.Status {
			continue
		}
		if !filters.IncludeAll && filters.Status == "" && campaign.Status != "published" {
			continue
		}
		if !filters.IncludeAll && campaign.Status == "published" && !isWithinWindow(campaign.PublishWindow, now) {
			continue
		}
		if filters.HazardType != "" && campaign.HazardType != filters.HazardType {
			continue
		}
		if filters.Region != "" && !containsNormalized(campaign.TargetRegions, filters.Region) {
			continue
		}
		if filters.Language != "" && !containsNormalized(campaign.Languages, filters.Language) {
			continue
		}
		results = append(results, campaign)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt.After(results[j].UpdatedAt)
	})
	return results
}

// GetCampaign returns a campaign by ID.
func (m *MemoryStore) GetCampaign(ctx context.Context, id string) (models.Campaign, bool) {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, campaign := range m.campaigns {
		if campaign.ID == id {
			return campaign, true
		}
	}
	return models.Campaign{}, false
}

// CreateCampaign creates a new campaign.
func (m *MemoryStore) CreateCampaign(ctx context.Context, request models.CreateCampaignRequest, authority models.AuthorityContext, now time.Time) models.Campaign {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	m.campaignCounter++
	campaign := models.Campaign{
		ID:             fmt.Sprintf("campaign_%03d", m.campaignCounter),
		Title:          request.Title,
		HazardType:     request.HazardType,
		TargetRegions:  request.TargetRegions,
		Languages:      request.Languages,
		ContentBlocks:  request.ContentBlocks,
		PublishWindow:  request.PublishWindow,
		Status:         request.Status,
		LinkedGuideIDs: request.LinkedGuideIDs,
		LinkedAlertIDs: request.LinkedAlertIDs,
		CreatedBy:      authority.ActorUserID,
		UpdatedBy:      authority.ActorUserID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	m.campaigns = append(m.campaigns, campaign)
	return campaign
}

// UpdateCampaign updates an existing campaign.
func (m *MemoryStore) UpdateCampaign(ctx context.Context, id string, request models.UpdateCampaignRequest, authority models.AuthorityContext, now time.Time) (models.Campaign, string, string) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.campaigns {
		if m.campaigns[index].ID != id {
			continue
		}
		next := m.campaigns[index]
		if request.Title != "" {
			next.Title = request.Title
		}
		if request.HazardType != "" {
			next.HazardType = request.HazardType
		}
		if request.TargetRegions != nil {
			next.TargetRegions = request.TargetRegions
		}
		if request.Languages != nil {
			next.Languages = request.Languages
		}
		if request.ContentBlocks != nil {
			next.ContentBlocks = request.ContentBlocks
		}
		if request.PublishWindow != nil {
			next.PublishWindow = *request.PublishWindow
		}
		if request.Status != "" {
			next.Status = request.Status
		}
		if request.LinkedGuideIDs != nil {
			next.LinkedGuideIDs = request.LinkedGuideIDs
		}
		if request.LinkedAlertIDs != nil {
			next.LinkedAlertIDs = request.LinkedAlertIDs
		}
		// Validate the effective (merged) status + window, so a status-only publish
		// re-checks the window and a future-dated draft window is not treated as
		// published.
		if code, message := validatePublishedWindow(next, now); code != "" {
			return models.Campaign{}, code, message
		}
		next.UpdatedBy = authority.ActorUserID
		next.UpdatedAt = now
		m.campaigns[index] = next
		return next, "", ""
	}
	return models.Campaign{}, "not_found", "campaign was not found"
}

// validatePublishedWindow enforces publish-window coherence for the effective
// campaign status. Only published campaigns are held to the not-premature and
// not-expired rules; drafts and archived campaigns may hold future or past windows.
func validatePublishedWindow(campaign models.Campaign, now time.Time) (string, string) {
	if campaign.Status != "published" {
		return "", ""
	}
	window := campaign.PublishWindow
	if window.StartsAt.IsZero() || window.EndsAt.IsZero() {
		return "invalid_publish_window", "published campaigns require publishWindow.startsAt and publishWindow.endsAt"
	}
	if window.EndsAt.Before(window.StartsAt) {
		return "invalid_publish_window", "publishWindow.endsAt must be after startsAt"
	}
	if now.After(window.EndsAt) {
		return "stale_campaign", "published campaigns cannot have an ended publish window"
	}
	if now.Before(window.StartsAt) {
		return "premature_campaign", "published campaigns cannot start in the future"
	}
	return "", ""
}

// CampaignPubliclyVisible reports whether a campaign may be shown to
// unauthenticated callers: it must be published and within its publish window.
func CampaignPubliclyVisible(campaign models.Campaign, now time.Time) bool {
	return campaign.Status == "published" && isWithinWindow(campaign.PublishWindow, now)
}

// ListMetrics returns mocked metrics for a campaign.
func (m *MemoryStore) ListMetrics(ctx context.Context, campaignID string, now time.Time) []models.CampaignMetric {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()

	if existing, ok := m.metrics[campaignID]; ok && len(existing) > 0 {
		return copyMetrics(existing)
	}
	return generateMockMetrics(campaignID, now)
}

// ListTemplates returns available campaign templates.
func (m *MemoryStore) ListTemplates(ctx context.Context) []models.CampaignTemplate {
	_ = ctx
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]models.CampaignTemplate(nil), m.templates...)
}

func isWithinWindow(window models.CampaignPublishWindow, now time.Time) bool {
	return !now.Before(window.StartsAt) && !now.After(window.EndsAt)
}

func containsNormalized(values []string, target string) bool {
	target = utils.NormalizeQueryValue(target)
	for _, value := range values {
		if utils.NormalizeQueryValue(value) == target {
			return true
		}
	}
	return false
}

func countMetrics(metrics map[string][]models.CampaignMetric) int {
	total := 0
	for _, list := range metrics {
		total += len(list)
	}
	return total
}

func copyMetrics(metrics []models.CampaignMetric) []models.CampaignMetric {
	copied := make([]models.CampaignMetric, len(metrics))
	copy(copied, metrics)
	return copied
}

func generateMockMetrics(campaignID string, now time.Time) []models.CampaignMetric {
	h := fnv.New32a()
	_, _ = h.Write([]byte(campaignID))
	base := int(h.Sum32()%5000) + 1000

	metrics := make([]models.CampaignMetric, 7)
	for i := 0; i < 7; i++ {
		day := now.Add(-time.Duration(6-i) * 24 * time.Hour).Truncate(24 * time.Hour)
		reach := base + i*120 + int(h.Sum32())%(i*50+100)
		engagement := reach / 10
		metrics[i] = models.CampaignMetric{
			ID:         fmt.Sprintf("metric_%s_%d", campaignID, i),
			CampaignID: campaignID,
			Date:       day,
			Reach:      reach,
			Engagement: engagement,
		}
	}
	return metrics
}
