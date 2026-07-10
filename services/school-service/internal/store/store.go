package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/school-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/utils"
)

// Store is the persistence interface for school emergency preparedness data.
type Store interface {
	ListSchools(filter models.SchoolFilter, scopedDistrict string, systemAdmin bool) []models.SchoolSummary
	GetSchool(id string) (models.SchoolProfile, bool)
	CreateSchool(request models.CreateSchoolRequest, ctx models.AuthorityContext, now time.Time) models.SchoolProfile
	UpdateSchool(id string, request models.UpdateSchoolRequest, ctx models.AuthorityContext, now time.Time) (models.SchoolProfile, string, string)
	ListDrills(schoolID string) []models.DrillRecord
	CreateDrill(schoolID string, request models.CreateDrillRequest, ctx models.AuthorityContext, now time.Time) (models.DrillRecord, string, string)
	GetLatestReadiness(schoolID string) (*models.ReadinessCheck, bool)
	CreateReadinessCheck(schoolID string, request models.CreateReadinessRequest, ctx models.AuthorityContext, now time.Time) (models.ReadinessCheck, string, string)
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu               sync.RWMutex
	schoolCounter    int
	drillCounter     int
	readinessCounter int
	schools          []models.SchoolProfile
	drills           []models.DrillRecord
	readinessChecks  []models.ReadinessCheck
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	s := &MemoryStore{
		schoolCounter:    3,
		drillCounter:     3,
		readinessCounter: 2,
	}
	s.schools = seedSchools(now)
	s.drills = seedDrills(now)
	s.readinessChecks = seedReadinessChecks(now)
	return s
}

// ListSchools returns school summaries filtered by district scope.
func (m *MemoryStore) ListSchools(filter models.SchoolFilter, scopedDistrict string, systemAdmin bool) []models.SchoolSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]models.SchoolSummary, 0)
	for _, school := range m.schools {
		if !systemAdmin && scopedDistrict != "" && !strings.EqualFold(school.District, scopedDistrict) {
			continue
		}
		if filter.District != "" && !strings.Contains(strings.ToLower(school.District), strings.ToLower(filter.District)) {
			continue
		}
		if filter.Query != "" && !schoolMatchesQuery(school, filter.Query) {
			continue
		}
		results = append(results, toSchoolSummary(school, m.latestDrillDateLocked(school.ID), m.latestReadinessStatusLocked(school.ID)))
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// GetSchool returns a full school profile by ID.
func (m *MemoryStore) GetSchool(id string) (models.SchoolProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, school := range m.schools {
		if school.ID == id {
			return school, true
		}
	}
	return models.SchoolProfile{}, false
}

// CreateSchool creates a new school profile.
func (m *MemoryStore) CreateSchool(request models.CreateSchoolRequest, ctx models.AuthorityContext, now time.Time) models.SchoolProfile {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.schoolCounter++
	school := models.SchoolProfile{
		ID:                fmt.Sprintf("school_%03d", m.schoolCounter),
		Name:              strings.TrimSpace(request.Name),
		Location:          request.Location,
		Region:            strings.TrimSpace(request.Region),
		District:          strings.TrimSpace(request.District),
		Address:           strings.TrimSpace(request.Address),
		StudentPopulation: request.StudentPopulation,
		EmergencyContacts: utils.NormalizeContacts(request.EmergencyContacts),
		Hazards:           utils.NormalizeStrings(request.Hazards),
		EvacuationPoints:  utils.NormalizeEvacuationPoints(request.EvacuationPoints),
		CreatedBy:         ctx.ActorUserID,
		UpdatedBy:         ctx.ActorUserID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	m.schools = append(m.schools, school)
	return school
}

// UpdateSchool updates an existing school profile.
func (m *MemoryStore) UpdateSchool(id string, request models.UpdateSchoolRequest, ctx models.AuthorityContext, now time.Time) (models.SchoolProfile, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.schools {
		if m.schools[index].ID != id {
			continue
		}
		next := m.schools[index]
		if request.Name != "" {
			next.Name = strings.TrimSpace(request.Name)
		}
		if request.Location != nil {
			next.Location = *request.Location
		}
		if request.Region != "" {
			next.Region = strings.TrimSpace(request.Region)
		}
		if request.District != "" {
			next.District = strings.TrimSpace(request.District)
		}
		if request.Address != "" {
			next.Address = strings.TrimSpace(request.Address)
		}
		if request.StudentPopulation != nil {
			next.StudentPopulation = *request.StudentPopulation
		}
		if request.EmergencyContacts != nil {
			next.EmergencyContacts = utils.NormalizeContacts(request.EmergencyContacts)
		}
		if request.Hazards != nil {
			next.Hazards = utils.NormalizeStrings(request.Hazards)
		}
		if request.EvacuationPoints != nil {
			next.EvacuationPoints = utils.NormalizeEvacuationPoints(request.EvacuationPoints)
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		m.schools[index] = next
		return next, "", ""
	}
	return models.SchoolProfile{}, "not_found", "school profile was not found"
}

// ListDrills returns drill records for a school.
func (m *MemoryStore) ListDrills(schoolID string) []models.DrillRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schoolID = strings.TrimSpace(schoolID)
	results := make([]models.DrillRecord, 0)
	for _, drill := range m.drills {
		if drill.SchoolID == schoolID {
			results = append(results, drill)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})
	return results
}

// CreateDrill adds a drill record for a school.
func (m *MemoryStore) CreateDrill(schoolID string, request models.CreateDrillRequest, ctx models.AuthorityContext, now time.Time) (models.DrillRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	schoolID = strings.TrimSpace(schoolID)
	found := false
	for _, school := range m.schools {
		if school.ID == schoolID {
			found = true
			break
		}
	}
	if !found {
		return models.DrillRecord{}, "not_found", "school profile was not found"
	}

	m.drillCounter++
	drill := models.DrillRecord{
		ID:           fmt.Sprintf("drill_%03d", m.drillCounter),
		SchoolID:     schoolID,
		Date:         request.Date,
		Type:         utils.NormalizeString(request.Type),
		Participants: request.Participants,
		Notes:        strings.TrimSpace(request.Notes),
		Completed:    request.Completed,
		CreatedBy:    ctx.ActorUserID,
		CreatedAt:    now,
	}
	m.drills = append(m.drills, drill)
	return drill, "", ""
}

// GetLatestReadiness returns the most recent readiness check for a school.
func (m *MemoryStore) GetLatestReadiness(schoolID string) (*models.ReadinessCheck, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	schoolID = strings.TrimSpace(schoolID)
	var latest *models.ReadinessCheck
	for i := range m.readinessChecks {
		if m.readinessChecks[i].SchoolID != schoolID {
			continue
		}
		if latest == nil || m.readinessChecks[i].CheckDate.After(latest.CheckDate) {
			check := m.readinessChecks[i]
			latest = &check
		}
	}
	return latest, latest != nil
}

// CreateReadinessCheck submits a readiness check for a school.
func (m *MemoryStore) CreateReadinessCheck(schoolID string, request models.CreateReadinessRequest, ctx models.AuthorityContext, now time.Time) (models.ReadinessCheck, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	schoolID = strings.TrimSpace(schoolID)
	found := false
	for _, school := range m.schools {
		if school.ID == schoolID {
			found = true
			break
		}
	}
	if !found {
		return models.ReadinessCheck{}, "not_found", "school profile was not found"
	}

	m.readinessCounter++
	check := models.ReadinessCheck{
		ID:             fmt.Sprintf("readiness_%03d", m.readinessCounter),
		SchoolID:       schoolID,
		CheckDate:      request.CheckDate,
		RiskLevel:      utils.NormalizeString(request.RiskLevel),
		AreaRiskRef:    strings.TrimSpace(request.AreaRiskRef),
		ChecklistItems: utils.NormalizeChecklistItems(request.ChecklistItems),
		OverallStatus:  utils.NormalizeString(request.OverallStatus),
		Notes:          strings.TrimSpace(request.Notes),
		CheckedBy:      ctx.ActorUserID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	m.readinessChecks = append(m.readinessChecks, check)
	return check, "", ""
}

func (m *MemoryStore) latestDrillDateLocked(schoolID string) *time.Time {
	var latest *time.Time
	for _, drill := range m.drills {
		if drill.SchoolID != schoolID || !drill.Completed {
			continue
		}
		if latest == nil || drill.Date.After(*latest) {
			date := drill.Date
			latest = &date
		}
	}
	return latest
}

func (m *MemoryStore) latestReadinessStatusLocked(schoolID string) string {
	latest, _ := m.GetLatestReadiness(schoolID)
	if latest == nil {
		return "not_assessed"
	}
	return latest.OverallStatus
}

func toSchoolSummary(school models.SchoolProfile, lastDrill *time.Time, readinessStatus string) models.SchoolSummary {
	return models.SchoolSummary{
		ID:                school.ID,
		Name:              school.Name,
		Location:          school.Location,
		District:          school.District,
		StudentPopulation: school.StudentPopulation,
		ReadinessStatus:   readinessStatus,
		LastDrillDate:     lastDrill,
		UpdatedAt:         school.UpdatedAt,
	}
}

func schoolMatchesQuery(school models.SchoolProfile, query string) bool {
	query = strings.ToLower(query)
	fields := []string{
		school.Name,
		school.District,
		school.Region,
		school.Address,
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}
