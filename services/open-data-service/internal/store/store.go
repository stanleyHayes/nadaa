package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/models"
)

// Store is the persistence interface for open data.
type Store interface {
	Health() string
	ListDatasets(category string, status models.PrivacyReviewStatus) []models.Dataset
	GetDataset(id string) (models.Dataset, bool)
	GetDatasetDownloads(datasetID string) []models.DatasetDownload
	CreateRequest(req models.OpenDataRequest) models.OpenDataRequest
	ListRequests() []models.OpenDataRequest
	GetRequest(id string) (models.OpenDataRequest, bool)
	UpdateRequest(req models.OpenDataRequest) models.OpenDataRequest
	RecordDownload(datasetID, format, ip string, now time.Time) models.DatasetDownload
	GetDownload(id string) (models.DatasetDownload, bool)
	RecordAuditEvent(event models.AuditEvent) models.AuditEvent
	ListAuditEvents() []models.AuditEvent
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu             sync.RWMutex
	createdAt      time.Time
	datasets       map[string]models.Dataset
	downloads      map[string]models.DatasetDownload
	requests       map[string]models.OpenDataRequest
	requestCounter int
	auditEvents    []models.AuditEvent
	auditCounter   int
}

// NewMemoryStore creates an in-memory store seeded with sample datasets.
func NewMemoryStore(now time.Time) Store {
	s := &MemoryStore{
		createdAt: now,
		datasets:  seedDatasets(now),
		downloads: map[string]models.DatasetDownload{},
		requests:  map[string]models.OpenDataRequest{},
	}
	return s
}

// Health returns a simple health indicator.
func (m *MemoryStore) Health() string {
	return "ok"
}

// ListDatasets returns datasets filtered by category and privacy review status.
func (m *MemoryStore) ListDatasets(category string, status models.PrivacyReviewStatus) []models.Dataset {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]models.Dataset, 0, len(m.datasets))
	for _, dataset := range m.datasets {
		if category != "" && string(dataset.Category) != category {
			continue
		}
		if status != "" && dataset.PrivacyReviewStatus != status {
			continue
		}
		out = append(out, dataset)
	}
	return out
}

// GetDataset returns a dataset by ID.
func (m *MemoryStore) GetDataset(id string) (models.Dataset, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dataset, ok := m.datasets[id]
	return dataset, ok
}

// GetDatasetDownloads returns downloads for a dataset.
func (m *MemoryStore) GetDatasetDownloads(datasetID string) []models.DatasetDownload {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]models.DatasetDownload, 0)
	for _, download := range m.downloads {
		if download.DatasetID == datasetID {
			out = append(out, download)
		}
	}
	return out
}

// CreateRequest stores a new access request.
func (m *MemoryStore) CreateRequest(req models.OpenDataRequest) models.OpenDataRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCounter++
	// Assign a unique id under the lock so concurrent or same-timestamp requests
	// never collide and silently overwrite one another in the map.
	req.ID = fmt.Sprintf("odr_%06d", m.requestCounter)
	m.requests[req.ID] = req
	return req
}

// ListRequests returns all access requests.
func (m *MemoryStore) ListRequests() []models.OpenDataRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]models.OpenDataRequest, 0, len(m.requests))
	for _, req := range m.requests {
		out = append(out, req)
	}
	return out
}

// GetRequest returns a request by ID.
func (m *MemoryStore) GetRequest(id string) (models.OpenDataRequest, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	req, ok := m.requests[id]
	return req, ok
}

// UpdateRequest updates an existing request.
func (m *MemoryStore) UpdateRequest(req models.OpenDataRequest) models.OpenDataRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[req.ID] = req
	return req
}

// RecordDownload creates a download record and returns it. The URL points at
// the real download endpoint; no checksum is fabricated for an artifact whose
// bytes this service does not hold.
func (m *MemoryStore) RecordDownload(datasetID, format, ip string, now time.Time) models.DatasetDownload {
	m.mu.Lock()
	defer m.mu.Unlock()

	download := models.DatasetDownload{
		ID:        generateID("dl", now),
		DatasetID: datasetID,
		Format:    format,
		URL:       "/api/v1/open-data/datasets/" + datasetID + "/download?format=" + format,
		Size:      estimateSize(datasetID, format),
		CreatedAt: now,
	}
	m.downloads[download.ID] = download
	return download
}

// GetDownload returns a download by ID.
func (m *MemoryStore) GetDownload(id string) (models.DatasetDownload, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	download, ok := m.downloads[id]
	return download, ok
}

// RecordAuditEvent appends an audit event to the local audit list and returns
// the persisted record. The ID is assigned under the lock.
func (m *MemoryStore) RecordAuditEvent(event models.AuditEvent) models.AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.auditCounter++
	event.ID = fmt.Sprintf("audit_%06d", m.auditCounter)
	m.auditEvents = append(m.auditEvents, event)
	return event
}

// ListAuditEvents returns the locally persisted audit events in record order.
func (m *MemoryStore) ListAuditEvents() []models.AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]models.AuditEvent, len(m.auditEvents))
	copy(out, m.auditEvents)
	return out
}
