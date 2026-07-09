package store

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// CVStore extends Store with computer vision operations.
type CVStore interface {
	AnalyzeImage(req models.CVAnalysisRequest, now time.Time) (models.CVAnalysisResult, error)
	GetCVResult(imageID string) (models.CVAnalysisResult, bool)
	ListCVResults() []models.CVAnalysisResult
}

// cvResultCache holds cached CV analysis results in memory.
type cvResultCache struct {
	results map[string]models.CVAnalysisResult
	mu      sync.RWMutex
}

func newCVResultCache() *cvResultCache {
	return &cvResultCache{results: make(map[string]models.CVAnalysisResult)}
}

func (c *cvResultCache) get(imageID string) (models.CVAnalysisResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result, ok := c.results[imageID]
	return result, ok
}

func (c *cvResultCache) set(result models.CVAnalysisResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[result.ImageID] = result
}

func (c *cvResultCache) list() []models.CVAnalysisResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]models.CVAnalysisResult, 0, len(c.results))
	for _, result := range c.results {
		out = append(out, result)
	}
	return out
}

// AnalyzeImage performs a deterministic rule-based mock CV analysis.
func (m *MemoryStore) AnalyzeImage(req models.CVAnalysisRequest, now time.Time) (models.CVAnalysisResult, error) {
	imageID := strings.TrimSpace(req.ImageID)
	if imageID == "" {
		return models.CVAnalysisResult{}, errors.New("imageId is required")
	}

	if cached, ok := m.cvCache.get(imageID); ok {
		return cached, nil
	}

	labels, humanReviewRequired := m.mockCVLabels(req)

	result := models.CVAnalysisResult{
		ID:                  fmt.Sprintf("cv_%s_%s", now.Format("20060102150405"), utils.SanitizeID(imageID)),
		ImageID:             imageID,
		Labels:              labels,
		ModelVersion:        "cv-mock-rule-engine-0.1.0",
		Limitations:         "This is a deterministic rule-based mock engine. It does not perform real image inference. Results are for contract testing and UI integration only. Always verify with human review before operational decisions.",
		HumanReviewRequired: humanReviewRequired,
		CreatedAt:           now.Format(time.RFC3339),
		ReviewStatus:        "pending",
	}

	m.cvCache.set(result)
	return result, nil
}

// GetCVResult returns a cached CV result by image ID.
func (m *MemoryStore) GetCVResult(imageID string) (models.CVAnalysisResult, bool) {
	return m.cvCache.get(imageID)
}

// ListCVResults returns all cached CV results.
func (m *MemoryStore) ListCVResults() []models.CVAnalysisResult {
	return m.cvCache.list()
}

func randomFloat64(minVal, maxVal float64) float64 {
	delta := maxVal - minVal
	n, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return minVal + delta/2
	}
	return minVal + (float64(n.Int64())/1000.0)*delta
}

// mockCVLabels simulates CV inference based on filename/metadata hints.
func (m *MemoryStore) mockCVLabels(req models.CVAnalysisRequest) ([]models.CVLabel, bool) {
	hint := strings.ToLower(req.ImageName + " " + req.ImageURL)

	if strings.Contains(hint, "flood") || strings.Contains(hint, "water") || strings.Contains(hint, "submerged") {
		return []models.CVLabel{
			{Label: "flood_evidence", Confidence: 0.92},
			{Label: "water_surface", Confidence: 0.88},
			{Label: "submerged_road", Confidence: 0.76},
		}, false
	}

	if strings.Contains(hint, "fire") || strings.Contains(hint, "flame") || strings.Contains(hint, "burn") || strings.Contains(hint, "smoke") {
		labels := []models.CVLabel{
			{Label: "fire_evidence", Confidence: 0.89},
		}
		if strings.Contains(hint, "smoke") {
			labels = append(labels, models.CVLabel{Label: "smoke_evidence", Confidence: 0.85})
		}
		return labels, false
	}

	if strings.Contains(hint, "injured") || strings.Contains(hint, "casualty") || strings.Contains(hint, "distress") || strings.Contains(hint, "victim") {
		return []models.CVLabel{
			{Label: "sensitive", Confidence: 0.95},
			{Label: "person_in_distress", Confidence: 0.82},
		}, true
	}

	if strings.Contains(hint, "random") || strings.Contains(hint, "test") || strings.Contains(hint, "blank") {
		return []models.CVLabel{
			{Label: "unclear", Confidence: 0.45 + randomFloat64(0, 0.15)},
		}, true
	}

	return []models.CVLabel{
		{Label: "no_evidence", Confidence: 0.62 + randomFloat64(0, 0.15)},
	}, true
}
