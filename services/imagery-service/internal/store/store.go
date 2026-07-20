// Package store provides persistence for imagery records.
package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/utils"
)

// Store is the persistence interface for imagery data.
type Store interface {
	List(filter models.ImageryListFilter) []models.ImageryRecord
	ListActive() []models.ImageryRecord
	Create(input models.ImageryUploadInput, fileName, contentType, storagePath, uploadedBy string, sizeBytes int64, now time.Time) models.ImageryRecord
	SetStoragePath(id, path string) bool
	GetByID(id string) (models.ImageryRecord, bool)
	Delete(id string) bool
	Expire(id string, now time.Time) (models.ImageryRecord, bool)
	RunLifecycle(now time.Time) int
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu            sync.RWMutex
	records       []models.ImageryRecord
	retentionDays int
	nextID        int
}

// NewMemoryStore creates an in-memory store seeded with fixtures.
func NewMemoryStore(now time.Time, retentionDays int) Store {
	activeGeometry := json.RawMessage(`{"type":"Polygon","coordinates":[[[-0.21,5.55],[-0.18,5.55],[-0.18,5.58],[-0.21,5.58],[-0.21,5.55]]]}`)
	expiringGeometry := json.RawMessage(`{"type":"Polygon","coordinates":[[[-0.25,5.60],[-0.22,5.60],[-0.22,5.63],[-0.25,5.63],[-0.25,5.60]]]}`)

	store := &MemoryStore{
		retentionDays: retentionDays,
		records: []models.ImageryRecord{
			{
				ID:               "img_seed_active",
				Reference:        "NADAA-IMG-img_seed_active",
				Source:           "satellite",
				CaptureTime:      now.Add(-24 * time.Hour),
				Geometry:         activeGeometry,
				CoverageAreaKm2:  12.5,
				ResolutionMeters: 10.0,
				License:          "CC-BY-4.0 NADMO",
				FileName:         "seed-active.tif",
				ContentType:      "image/tiff",
				SizeBytes:        1024,
				StoragePath:      "uploads/seed-active.tif",
				Status:           "active",
				UploadedBy:       "usr_imagery_admin",
				CreatedAt:        now.Add(-30 * 24 * time.Hour),
				ExpiresAt:        now.Add(60 * 24 * time.Hour),
			},
			{
				ID:                "img_seed_expiring",
				Reference:         "NADAA-IMG-img_seed_expiring",
				Source:            "drone",
				CaptureTime:       now.Add(-48 * time.Hour),
				Geometry:          expiringGeometry,
				CoverageAreaKm2:   2.1,
				ResolutionMeters:  0.5,
				License:           "Internal",
				RelatedIncidentID: "incident_001",
				FileName:          "seed-expiring.jpg",
				ContentType:       "image/jpeg",
				SizeBytes:         2048,
				StoragePath:       "uploads/seed-expiring.jpg",
				Status:            "active",
				UploadedBy:        "usr_drone_operator",
				CreatedAt:         now.Add(-100 * 24 * time.Hour),
				ExpiresAt:         now.Add(-1 * time.Hour),
			},
		},
	}
	// Seed the monotonic ID counter above every existing ID so Create never
	// reuses an ID, even after records are deleted.
	store.nextID = len(store.records)
	for _, record := range store.records {
		if n, err := strconv.Atoi(strings.TrimPrefix(record.ID, "img_")); err == nil && n > store.nextID {
			store.nextID = n
		}
	}
	return store
}

// Create inserts a new imagery record.
func (m *MemoryStore) Create(input models.ImageryUploadInput, fileName, contentType, storagePath, uploadedBy string, sizeBytes int64, now time.Time) models.ImageryRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	id := fmt.Sprintf("img_%03d", m.nextID)
	record := models.ImageryRecord{
		ID:                id,
		Reference:         "NADAA-IMG-" + id,
		Source:            input.Source,
		CaptureTime:       input.CaptureTime,
		Geometry:          input.Geometry,
		CoverageAreaKm2:   input.CoverageAreaKm2,
		ResolutionMeters:  input.ResolutionMeters,
		License:           input.License,
		RelatedIncidentID: input.RelatedIncidentID,
		RelatedRiskZoneID: input.RelatedRiskZoneID,
		MlWorkflowID:      input.MlWorkflowID,
		FileName:          fileName,
		ContentType:       contentType,
		SizeBytes:         sizeBytes,
		StoragePath:       storagePath,
		Status:            "active",
		UploadedBy:        uploadedBy,
		CreatedAt:         now,
		ExpiresAt:         now.Add(time.Duration(m.retentionDays) * 24 * time.Hour),
	}
	m.records = append(m.records, record)
	return record
}

// List returns imagery records matching the provided filter.
func (m *MemoryStore) List(filter models.ImageryListFilter) []models.ImageryRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]models.ImageryRecord, 0)
	for _, record := range m.records {
		if filter.Source != "" && record.Source != filter.Source {
			continue
		}
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if filter.RelatedIncidentID != "" && record.RelatedIncidentID != filter.RelatedIncidentID {
			continue
		}
		if filter.RelatedRiskZoneID != "" && record.RelatedRiskZoneID != filter.RelatedRiskZoneID {
			continue
		}
		if filter.Query != "" && !recordMatchesQuery(record, strings.ToLower(filter.Query)) {
			continue
		}
		records = append(records, copyRecord(record))
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records
}

// ListActive returns all active imagery records.
func (m *MemoryStore) ListActive() []models.ImageryRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]models.ImageryRecord, 0)
	for _, record := range m.records {
		if record.Status == "active" {
			records = append(records, copyRecord(record))
		}
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records
}

func recordMatchesQuery(record models.ImageryRecord, query string) bool {
	return strings.Contains(strings.ToLower(record.Reference), query) ||
		strings.Contains(strings.ToLower(record.FileName), query) ||
		strings.Contains(strings.ToLower(record.License), query) ||
		strings.Contains(strings.ToLower(record.RelatedIncidentID), query) ||
		strings.Contains(strings.ToLower(record.RelatedRiskZoneID), query) ||
		strings.Contains(strings.ToLower(record.MlWorkflowID), query)
}

// SetStoragePath updates the stored file path for an imagery record after the file is written.
func (m *MemoryStore) SetStoragePath(id, path string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.records {
		if m.records[index].ID == id {
			m.records[index].StoragePath = path
			return true
		}
	}
	return false
}

// GetByID returns an imagery record by id.
func (m *MemoryStore) GetByID(id string) (models.ImageryRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, record := range m.records {
		if record.ID == id {
			return copyRecord(record), true
		}
	}
	return models.ImageryRecord{}, false
}

// Delete removes an imagery record by id.
func (m *MemoryStore) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index, record := range m.records {
		if record.ID == id {
			m.records = append(m.records[:index], m.records[index+1:]...)
			return true
		}
	}
	return false
}

// Expire marks an imagery record as expired.
func (m *MemoryStore) Expire(id string, now time.Time) (models.ImageryRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.records {
		if m.records[index].ID != id {
			continue
		}
		m.records[index].Status = "expired"
		return copyRecord(m.records[index]), true
	}
	return models.ImageryRecord{}, false
}

// RunLifecycle expires records whose ExpiresAt has passed and deletes their
// stored files, so expired imagery is neither retained on disk nor
// downloadable. File deletion is best-effort: a missing file is fine (nothing
// to retain), other failures are logged and the record is still marked
// expired so the download path refuses it.
func (m *MemoryStore) RunLifecycle(now time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for index := range m.records {
		if m.records[index].Status != "active" {
			continue
		}
		if now.After(m.records[index].ExpiresAt) {
			m.records[index].Status = "expired"
			count++
			if path := m.records[index].StoragePath; path != "" {
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					// #nosec G706 -- record id and path are sanitized with utils.SafeLogValue.
					log.Printf("ERROR imagery-service lifecycle_file_delete_failed id=%s path=%s error=%v", utils.SafeLogValue(m.records[index].ID), utils.SafeLogValue(path), err)
				}
			}
		}
	}
	return count
}

func copyRecord(record models.ImageryRecord) models.ImageryRecord {
	record.Geometry = append(json.RawMessage(nil), record.Geometry...)
	return record
}
