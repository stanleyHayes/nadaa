package store

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"slices"
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
	ListCVResults(limit, offset int) ([]models.CVAnalysisResult, int)
	ReviewCVResult(id string, req models.CVReviewRequest, reviewedBy string, now time.Time) (models.CVAnalysisResult, bool)
}

// maxCVResults bounds the in-memory CV result cache; the oldest results are
// evicted FIFO once the cap is reached.
const maxCVResults = 500

// cvResultCache holds cached CV analysis results in memory.
type cvResultCache struct {
	results map[string]models.CVAnalysisResult
	// order tracks insertion order (oldest first) for FIFO eviction and a
	// stable newest-first listing.
	order []string
	mu    sync.RWMutex
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
	if _, exists := c.results[result.ImageID]; !exists {
		c.order = append(c.order, result.ImageID)
	}
	c.results[result.ImageID] = result
	// FIFO eviction keeps the cache bounded.
	for len(c.order) > maxCVResults {
		delete(c.results, c.order[0])
		c.order = c.order[1:]
	}
}

func (c *cvResultCache) list() []models.CVAnalysisResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]models.CVAnalysisResult, 0, len(c.order))
	for _, imageID := range slices.Backward(c.order) {
		out = append(out, c.results[imageID])
	}
	return out
}

// review records a human review decision on the result identified by result
// ID or image ID.
func (c *cvResultCache) review(id string, req models.CVReviewRequest, reviewedBy, reviewedAt string) (models.CVAnalysisResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result, ok := c.results[id]
	if !ok {
		for _, candidate := range c.results {
			if candidate.ID == id {
				result = candidate
				ok = true
				break
			}
		}
		if !ok {
			return models.CVAnalysisResult{}, false
		}
	}
	result.ReviewStatus = req.Decision
	result.ReviewNote = strings.TrimSpace(req.Note)
	result.ReviewedBy = reviewedBy
	result.ReviewedAt = reviewedAt
	c.results[result.ImageID] = result
	return result, true
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

// ListCVResults returns a page of cached CV results, newest first, with the
// total number of retained results.
func (m *MemoryStore) ListCVResults(limit, offset int) ([]models.CVAnalysisResult, int) {
	return paginate(m.cvCache.list(), limit, offset)
}

// ReviewCVResult records a human review decision for a cached CV result,
// looked up by result ID or image ID.
func (m *MemoryStore) ReviewCVResult(id string, req models.CVReviewRequest, reviewedBy string, now time.Time) (models.CVAnalysisResult, bool) {
	return m.cvCache.review(id, req, reviewedBy, now.Format(time.RFC3339))
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
