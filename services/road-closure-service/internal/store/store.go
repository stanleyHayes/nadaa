package store

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/utils"
)

// Store is the persistence interface for road closure data.
type Store interface {
	ListClosures(filter models.ListFilter, now time.Time) []models.RoadClosureRecord
	CreateClosure(request models.CreateRoadClosureRequest, ctx models.AuthorityContext, now time.Time) models.RoadClosureRecord
	UpdateClosure(id string, request models.UpdateRoadClosureRequest, ctx models.AuthorityContext, now time.Time) (models.RoadClosureRecord, string, string)
	ImportAdapter(request models.AdapterImportRequest, ctx models.AuthorityContext, now time.Time) []models.RoadClosureRecord
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu             sync.RWMutex
	closures       []models.RoadClosureRecord
	closureCounter int
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	return &MemoryStore{
		closureCounter: 2,
		closures:       seedClosures(now),
	}
}

// ListClosures returns closures matching the provided filter.
func (m *MemoryStore) ListClosures(filter models.ListFilter, now time.Time) []models.RoadClosureRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]models.RoadClosureRecord, 0, len(m.closures))
	for _, closure := range m.closures {
		if filter.Status != "" && closure.Status != filter.Status {
			continue
		}
		if (filter.Status == "" || filter.Status == "active") && !filter.IncludeExpired && !utils.IsClosureEffective(closure, now) {
			continue
		}
		if filter.BBox != nil && !utils.ClosureIntersectsBBox(closure.Geometry, *filter.BBox) {
			continue
		}
		if filter.Location != nil {
			closure.DistanceMeters = int(math.Round(utils.MinDistanceToLineString(*filter.Location, closure.Geometry)))
			if float64(closure.DistanceMeters) > filter.RadiusMeters {
				continue
			}
		}
		results = append(results, closure)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Status != results[j].Status {
			return utils.StatusRank(results[i].Status) < utils.StatusRank(results[j].Status)
		}
		if results[i].Severity != results[j].Severity {
			return utils.SeverityRank(results[i].Severity) < utils.SeverityRank(results[j].Severity)
		}
		if filter.Location != nil && results[i].DistanceMeters != results[j].DistanceMeters {
			return results[i].DistanceMeters < results[j].DistanceMeters
		}
		return results[i].RoadName < results[j].RoadName
	})

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}
	return utils.CopyClosures(results)
}

// CreateClosure creates a new road closure record.
func (m *MemoryStore) CreateClosure(request models.CreateRoadClosureRequest, ctx models.AuthorityContext, now time.Time) models.RoadClosureRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closureCounter++
	closure := models.RoadClosureRecord{
		ID:         fmt.Sprintf("road_closure_%03d", m.closureCounter),
		RoadName:   request.RoadName,
		Reason:     request.Reason,
		Status:     request.Status,
		Severity:   request.Severity,
		Source:     request.Source,
		SourceRef:  request.SourceRef,
		Geometry:   request.Geometry,
		ValidFrom:  now,
		ValidTo:    request.ValidTo,
		DetourNote: request.DetourNote,
		CreatedBy:  ctx.ActorUserID,
		UpdatedBy:  ctx.ActorUserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if request.ValidFrom != nil {
		closure.ValidFrom = *request.ValidFrom
	}
	m.closures = append(m.closures, closure)
	return closure
}

// UpdateClosure updates an existing road closure record.
func (m *MemoryStore) UpdateClosure(id string, request models.UpdateRoadClosureRequest, ctx models.AuthorityContext, now time.Time) (models.RoadClosureRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.closures {
		if m.closures[index].ID != id {
			continue
		}
		next := m.closures[index]
		if request.RoadName != "" {
			next.RoadName = request.RoadName
		}
		if request.Reason != "" {
			next.Reason = request.Reason
		}
		if request.Status != "" {
			next.Status = request.Status
		}
		if request.Severity != "" {
			next.Severity = request.Severity
		}
		if request.Source != "" {
			next.Source = request.Source
		}
		if request.SourceRef != "" {
			next.SourceRef = request.SourceRef
		}
		if request.Geometry != nil {
			next.Geometry = *request.Geometry
		}
		if request.ValidFrom != nil {
			next.ValidFrom = *request.ValidFrom
		}
		if request.ValidTo != nil {
			next.ValidTo = request.ValidTo
		}
		if request.DetourNote != "" {
			next.DetourNote = request.DetourNote
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		m.closures[index] = next
		return next, "", ""
	}
	return models.RoadClosureRecord{}, "not_found", "road closure was not found"
}

// ImportAdapter imports a closure from an adapter payload.
func (m *MemoryStore) ImportAdapter(request models.AdapterImportRequest, ctx models.AuthorityContext, now time.Time) []models.RoadClosureRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	geometry, err := utils.ParseWKTLineString(request.Geometry)
	if err != nil {
		return nil
	}

	m.closureCounter++
	closure := models.RoadClosureRecord{
		ID:         fmt.Sprintf("road_closure_%03d", m.closureCounter),
		RoadName:   request.RoadName,
		Reason:     request.Reason,
		Status:     request.Status,
		Severity:   utils.SeverityFromStatus(request.Status),
		Source:     request.Source,
		SourceRef:  request.SourceRef,
		Geometry:   geometry,
		ValidFrom:  request.ValidFrom,
		ValidTo:    request.ValidTo,
		DetourNote: request.Detour,
		CreatedBy:  ctx.ActorUserID,
		UpdatedBy:  ctx.ActorUserID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	m.closures = append(m.closures, closure)
	return []models.RoadClosureRecord{closure}
}
