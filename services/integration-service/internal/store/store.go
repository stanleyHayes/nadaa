package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/integration-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/integration-service/internal/utils"
)

// Store is the persistence interface for integration data.
type Store interface {
	ListContracts(domain, direction, partner string) []models.IntegrationContract
	ListObservations(source, metric string) []models.WeatherHydrologyObservation
	ListImportedObservations(source, metric string) []models.ImportedWeatherHydrologyObservation
	CreateObservationImportJob(request models.ObservationImportRequest, trigger string, now time.Time, attempt int) models.ObservationImportJob
	ListObservationImportJobs(status string) []models.ObservationImportJob
	RetryObservationImportJob(jobID string, now time.Time) (models.ObservationImportJob, bool, string)
	CreateSyncEvent(request models.SyncRequest, now time.Time) (models.SyncEvent, bool)
	ListSyncEvents(eventType string) []models.SyncEvent
	ImportRoadClosure(request models.RoadClosureImportRequest) models.RoadClosureImportRecord
	ListRoadClosureImports(source string) []models.RoadClosureImportRecord
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu                   sync.RWMutex
	contracts            []models.IntegrationContract
	observations         []models.WeatherHydrologyObservation
	importedObservations []models.ImportedWeatherHydrologyObservation
	importJobs           []models.ObservationImportJob
	syncEvents           []models.SyncEvent
	syncEventSeq         int
	roadClosureImports   []models.RoadClosureImportRecord
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	return &MemoryStore{
		contracts:    seedContracts(now),
		observations: seedObservations(now),
	}
}

// ListContracts returns contracts matching the provided filters.
func (m *MemoryStore) ListContracts(domain string, direction string, partner string) []models.IntegrationContract {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contracts := make([]models.IntegrationContract, 0, len(m.contracts))
	for _, contract := range m.contracts {
		if domain != "" && contract.Domain != domain {
			continue
		}
		if direction != "" && contract.Direction != direction {
			continue
		}
		if partner != "" && !strings.Contains(utils.NormalizeQueryValue(contract.Partner), partner) {
			continue
		}
		contracts = append(contracts, contract)
	}

	sort.Slice(contracts, func(i, j int) bool {
		if contracts[i].Domain == contracts[j].Domain {
			return contracts[i].Partner < contracts[j].Partner
		}
		return contracts[i].Domain < contracts[j].Domain
	})
	return contracts
}

// ListObservations returns mock observations matching the filters.
func (m *MemoryStore) ListObservations(source string, metric string) []models.WeatherHydrologyObservation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	observations := make([]models.WeatherHydrologyObservation, 0, len(m.observations))
	for _, observation := range m.observations {
		if source != "" && observation.Source != source {
			continue
		}
		if metric != "" && observation.Metric != metric {
			continue
		}
		observations = append(observations, observation)
	}

	sort.Slice(observations, func(i, j int) bool {
		return observations[i].ObservedAt.Before(observations[j].ObservedAt)
	})
	return observations
}

// ListImportedObservations returns imported observations matching the filters.
func (m *MemoryStore) ListImportedObservations(source string, metric string) []models.ImportedWeatherHydrologyObservation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	observations := make([]models.ImportedWeatherHydrologyObservation, 0, len(m.importedObservations))
	for _, observation := range m.importedObservations {
		if source != "" && observation.Source != source {
			continue
		}
		if metric != "" && observation.Metric != metric {
			continue
		}
		observations = append(observations, observation)
	}

	sort.Slice(observations, func(i, j int) bool {
		return observations[i].ObservedAt.Before(observations[j].ObservedAt)
	})
	return observations
}

// CreateObservationImportJob creates and runs an import job.
func (m *MemoryStore) CreateObservationImportJob(request models.ObservationImportRequest, trigger string, now time.Time, attempt int) models.ObservationImportJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createObservationImportJobLocked(request, trigger, now, attempt)
}

func (m *MemoryStore) createObservationImportJobLocked(request models.ObservationImportRequest, trigger string, now time.Time, attempt int) models.ObservationImportJob {
	request.AdapterID = utils.DefaultImportAdapterID(request.AdapterID)
	request.Metric = utils.NormalizeQueryValue(request.Metric)
	request.RequestedBy = strings.TrimSpace(request.RequestedBy)
	request.CorrelationID = strings.TrimSpace(request.CorrelationID)

	job := models.ObservationImportJob{
		ID:            fmt.Sprintf("import_%s_%s_%03d", utils.SanitizeID(trigger), now.Format("20060102150405"), len(m.importJobs)+1),
		AdapterID:     request.AdapterID,
		Source:        request.AdapterID,
		Metric:        request.Metric,
		Status:        "running",
		Trigger:       trigger,
		Attempts:      attempt,
		Retryable:     true,
		StartedAt:     now,
		RequestedBy:   request.RequestedBy,
		CorrelationID: request.CorrelationID,
	}

	candidates := m.importCandidateObservations(request.Metric)
	if request.SimulateFailure {
		finishedAt := now.Add(250 * time.Millisecond)
		nextRetryAt := now.Add(30 * time.Second)
		job.Status = "failed"
		job.FinishedAt = &finishedAt
		job.NextRetryAt = &nextRetryAt
		job.FailedCount = len(candidates)
		job.Error = strings.TrimSpace(request.FailureMessage)
		if job.Error == "" {
			job.Error = "fixture importer failure requested"
		}
		job.Message = "Import failed before observations were stored; job is retryable."
		m.importJobs = append(m.importJobs, job)
		return job
	}

	imported := 0
	for _, observation := range candidates {
		stored := buildImportedObservation(observation, job.ID, now)
		m.upsertImportedObservation(stored)
		imported++
	}

	finishedAt := now.Add(250 * time.Millisecond)
	job.Status = "succeeded"
	job.FinishedAt = &finishedAt
	job.ImportedCount = imported
	job.Message = fmt.Sprintf("Imported %d weather/hydrology observation%s.", imported, utils.PluralSuffix(imported))
	m.importJobs = append(m.importJobs, job)
	return job
}

func (m *MemoryStore) importCandidateObservations(metric string) []models.WeatherHydrologyObservation {
	candidates := make([]models.WeatherHydrologyObservation, 0, len(m.observations))
	for _, observation := range m.observations {
		if metric != "" && observation.Metric != metric {
			continue
		}
		candidates = append(candidates, observation)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ObservedAt.Before(candidates[j].ObservedAt)
	})
	return candidates
}

func (m *MemoryStore) upsertImportedObservation(next models.ImportedWeatherHydrologyObservation) {
	for index, observation := range m.importedObservations {
		if observation.ID == next.ID {
			m.importedObservations[index] = next
			return
		}
	}
	m.importedObservations = append(m.importedObservations, next)
}

// ListObservationImportJobs returns import jobs matching the status filter.
func (m *MemoryStore) ListObservationImportJobs(status string) []models.ObservationImportJob {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]models.ObservationImportJob, 0, len(m.importJobs))
	for _, job := range m.importJobs {
		if status != "" && job.Status != status {
			continue
		}
		jobs = append(jobs, job)
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartedAt.After(jobs[j].StartedAt)
	})
	return jobs
}

// RetryObservationImportJob retries a failed import job.
func (m *MemoryStore) RetryObservationImportJob(jobID string, now time.Time) (models.ObservationImportJob, bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var previous *models.ObservationImportJob
	for index := range m.importJobs {
		if m.importJobs[index].ID == jobID {
			previous = &m.importJobs[index]
			break
		}
	}
	if previous == nil {
		return models.ObservationImportJob{}, false, ""
	}
	if previous.Status != "failed" || !previous.Retryable {
		return models.ObservationImportJob{}, true, "only failed retryable import jobs can be retried"
	}

	request := models.ObservationImportRequest{
		AdapterID:      previous.AdapterID,
		Metric:         previous.Metric,
		RequestedBy:    previous.RequestedBy,
		CorrelationID:  previous.CorrelationID,
		FailureMessage: previous.Error,
	}
	job := m.createObservationImportJobLocked(request, "retry", now, previous.Attempts+1)
	return job, true, ""
}

// CreateSyncEvent records a new sync event. It dedupes on the validated
// correlation ID: a replay returns the existing event with created=false.
func (m *MemoryStore) CreateSyncEvent(request models.SyncRequest, now time.Time) (models.SyncEvent, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	correlationID := strings.TrimSpace(request.CorrelationID)
	for _, event := range m.syncEvents {
		if event.CorrelationID == correlationID {
			return event, false
		}
	}

	m.syncEventSeq++
	event := models.SyncEvent{
		ID:              fmt.Sprintf("sync_%s_%s_%06d", utils.SanitizeID(request.Type), utils.SanitizeID(request.SourceID), m.syncEventSeq),
		Type:            request.Type,
		SourceID:        request.SourceID,
		Reference:       request.Reference,
		TargetAgencyIDs: request.TargetAgencyIDs,
		CorrelationID:   correlationID,
		Status:          "accepted",
		AdapterID:       adapterIDForSyncType(request.Type),
		QueuedAt:        now,
		Retryable:       true,
	}
	m.syncEvents = append(m.syncEvents, event)
	return event, true
}

// ListSyncEvents returns sync events matching the type filter.
func (m *MemoryStore) ListSyncEvents(eventType string) []models.SyncEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	events := make([]models.SyncEvent, 0, len(m.syncEvents))
	for _, event := range m.syncEvents {
		if eventType != "" && event.Type != eventType {
			continue
		}
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].QueuedAt.After(events[j].QueuedAt)
	})
	return events
}

// ImportRoadClosure persists a road closure import record.
func (m *MemoryStore) ImportRoadClosure(request models.RoadClosureImportRequest) models.RoadClosureImportRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := models.RoadClosureImportRecord{
		ID:         fmt.Sprintf("rci_%03d", len(m.roadClosureImports)+1),
		Source:     request.Source,
		SourceRef:  request.SourceRef,
		RoadName:   request.RoadName,
		Status:     request.Status,
		Reason:     request.Reason,
		Geometry:   request.Geometry,
		ValidFrom:  request.ValidFrom,
		ValidTo:    request.ValidTo,
		Detour:     request.Detour,
		ImportedAt: time.Now().UTC(),
	}
	m.roadClosureImports = append(m.roadClosureImports, record)
	return record
}

// ListRoadClosureImports returns road closure imports matching the source filter.
func (m *MemoryStore) ListRoadClosureImports(source string) []models.RoadClosureImportRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	imports := make([]models.RoadClosureImportRecord, 0, len(m.roadClosureImports))
	for _, record := range m.roadClosureImports {
		if source != "" && record.Source != source {
			continue
		}
		imports = append(imports, record)
	}

	sort.Slice(imports, func(i, j int) bool {
		return imports[i].ImportedAt.After(imports[j].ImportedAt)
	})
	return imports
}

func adapterIDForSyncType(eventType string) string {
	if eventType == "alert" {
		return "mock-alert-sync-adapter"
	}
	return "mock-incident-sync-adapter"
}

func buildImportedObservation(observation models.WeatherHydrologyObservation, jobID string, importedAt time.Time) models.ImportedWeatherHydrologyObservation {
	imported := models.ImportedWeatherHydrologyObservation{
		ID:            "imported_" + observation.ID,
		Source:        observation.Source,
		Metric:        observation.Metric,
		Value:         observation.Value,
		Unit:          observation.Unit,
		StationID:     observation.StationID,
		Location:      observation.Location,
		ObservedAt:    observation.ObservedAt,
		ValidFrom:     observation.ValidFrom,
		ValidTo:       observation.ValidTo,
		ImportJobID:   jobID,
		ImportedAt:    importedAt,
		SourceRecord:  observation.ID,
		StorageTarget: "weather_observations",
		Metadata: map[string]string{
			"quality":     observation.Quality,
			"generatedBy": observation.GeneratedBy,
			"unit":        observation.Unit,
		},
	}
	if observation.Metric == "rainfall_mm" {
		value := observation.Value
		imported.RainfallMM = &value
	}
	if observation.Metric == "water_level_m" {
		value := observation.Value
		imported.WaterLevelM = &value
	}
	return imported
}
