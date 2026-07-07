package store

import (
	"sort"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/guide-service/internal/models"
)

// Store is the persistence interface for guide data.
type Store interface {
	ListGuides(filters models.GuideFilters) []models.EmergencyGuide
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu     sync.RWMutex
	guides []models.EmergencyGuide
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	return &MemoryStore{guides: seedGuides(now)}
}

// ListGuides returns guides matching the provided filters.
func (m *MemoryStore) ListGuides(filters models.GuideFilters) []models.EmergencyGuide {
	m.mu.RLock()
	defer m.mu.RUnlock()

	guides := m.listGuidesLocked(filters)

	if filters.Language != "" && filters.Language != "en" && len(guides) == 0 {
		fallbackFilters := filters
		fallbackFilters.Language = "en"
		guides = m.listGuidesLocked(fallbackFilters)
	}

	sortGuides(guides)
	return guides
}

func (m *MemoryStore) listGuidesLocked(filters models.GuideFilters) []models.EmergencyGuide {
	guides := make([]models.EmergencyGuide, 0, len(m.guides))
	for _, guide := range m.guides {
		if filters.HazardType != "" && guide.HazardType != filters.HazardType {
			continue
		}
		if filters.Stage != "" && guide.Stage != filters.Stage {
			continue
		}
		if filters.Language != "" && guide.Language != filters.Language {
			continue
		}
		if filters.Offline != nil && guide.OfflineAvailable != *filters.Offline {
			continue
		}
		guides = append(guides, guide)
	}
	return guides
}

func sortGuides(guides []models.EmergencyGuide) {
	sort.Slice(guides, func(i, j int) bool {
		if guides[i].SortOrder == guides[j].SortOrder {
			return guides[i].Title < guides[j].Title
		}
		return guides[i].SortOrder < guides[j].SortOrder
	})
}
