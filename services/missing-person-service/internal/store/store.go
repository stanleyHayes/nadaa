package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/models"
)

// Store is the persistence interface for missing-person records.
type Store interface {
	ListMissingPersons(filter models.MissingPersonFilter) []models.MissingPerson
	ListPublicMissingPersons(filter models.MissingPersonFilter) []models.PublicMissingPerson
	CreateMissingPerson(request models.CreateMissingPersonRequest, createdBy string, now time.Time) models.MissingPerson
	GetMissingPerson(id string) (models.MissingPerson, bool)
	GetPublicMissingPerson(id string) (models.PublicMissingPerson, bool)
	ReviewMissingPerson(id string, request models.ReviewMissingPersonRequest, ctx models.AuthorityContext, now time.Time) (models.MissingPerson, string, string)
	CloseMissingPerson(id string, request models.CloseMissingPersonRequest, ctx models.AuthorityContext, now time.Time) (models.MissingPerson, string, string)
	ListAudit(id string) []models.MissingPersonAuditEntry
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu       sync.RWMutex
	seq      int
	auditSeq int
	records  []models.MissingPerson
	audits   []models.MissingPersonAuditEntry
}

// NewMemoryStore creates an in-memory store seeded with fixtures.
func NewMemoryStore(now time.Time) Store {
	s := &MemoryStore{seq: 2, auditSeq: 2}
	publicReviewedAt := now.Add(-3 * time.Hour)
	privateReviewedAt := now.Add(-2 * time.Hour)
	lat := 5.56
	lng := -0.2
	ageChild := 12
	ageSenior := 68
	s.records = []models.MissingPerson{
		{
			ID:          "missing_001",
			Reference:   "MP-20260707-001",
			PersonName:  "Kojo Mensah",
			Age:         &ageChild,
			Gender:      "male",
			Description: "Last seen wearing a blue school shirt and black shorts near the shelter registration desk.",
			PhotoURL:    "https://example.test/photos/kojo-mensah.jpg",
			LastSeenAt:  now.Add(-6 * time.Hour),
			LastSeenLocation: models.LastSeenLocation{
				Label:    "Accra Metro Assembly Shelter",
				Region:   "Greater Accra",
				District: "Accra Metropolitan",
				Lat:      &lat,
				Lng:      &lng,
			},
			RelatedIncidentID: "inc_accra_flood_0241",
			Reporter: models.ReporterContact{
				Name:                 "Ama Mensah",
				Phone:                "+233200000111",
				Relationship:         "mother",
				ConsentToContact:     true,
				ConsentToPublicShare: true,
			},
			Status:           "active",
			ReviewStatus:     "approved",
			PublicVisibility: "public",
			PublicSummary:    "Child separated during shelter registration. Please contact authorities through 112 with credible sightings.",
			ReviewNotes:      "Approved by fixture reviewer with guardian consent.",
			CreatedBy:        "public",
			CreatedAt:        now.Add(-5 * time.Hour),
			UpdatedAt:        publicReviewedAt,
			ReviewedBy:       "usr_seed_reviewer",
			ReviewedAt:       &publicReviewedAt,
		},
		{
			ID:          "missing_002",
			Reference:   "MP-20260707-002",
			PersonName:  "Efua Boateng",
			Age:         &ageSenior,
			Gender:      "female",
			Description: "Older adult reported missing after evacuation from a low-lying household.",
			LastSeenAt:  now.Add(-8 * time.Hour),
			LastSeenLocation: models.LastSeenLocation{
				Label:    "Osu Community Hall",
				Region:   "Greater Accra",
				District: "Korle Klottey",
			},
			Reporter: models.ReporterContact{
				Name:             "Kweku Boateng",
				Phone:            "+233200000222",
				Relationship:     "son",
				ConsentToContact: true,
			},
			Status:           "pending_review",
			ReviewStatus:     "pending",
			PublicVisibility: "private",
			ReviewNotes:      "Awaiting authority consent verification.",
			CreatedBy:        "public",
			CreatedAt:        now.Add(-7 * time.Hour),
			UpdatedAt:        privateReviewedAt,
			ReviewedAt:       &privateReviewedAt,
		},
	}
	s.audits = []models.MissingPersonAuditEntry{
		{
			ID:          "audit_001",
			RecordID:    "missing_001",
			Action:      "missing_person.created",
			ActorUserID: "public",
			Notes:       "Public intake fixture.",
			CreatedAt:   now.Add(-5 * time.Hour),
		},
		{
			ID:            "audit_002",
			RecordID:      "missing_001",
			Action:        "missing_person.reviewed",
			ActorUserID:   "usr_seed_reviewer",
			ActorAgencyID: "00000000-0000-0000-0000-000000000204",
			ActorRole:     "district_officer",
			Notes:         "Approved for public visibility.",
			CreatedAt:     publicReviewedAt,
		},
	}
	return s
}

// ListMissingPersons returns full sensitive records matching a filter.
func (m *MemoryStore) ListMissingPersons(filter models.MissingPersonFilter) []models.MissingPerson {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]models.MissingPerson, 0)
	for _, record := range m.records {
		if !matchesFilter(record, filter) {
			continue
		}
		records = append(records, record)
	}
	sortRecords(records)
	return copyMissingPersons(records)
}

// ListPublicMissingPersons returns approved public-safe records.
func (m *MemoryStore) ListPublicMissingPersons(filter models.MissingPersonFilter) []models.PublicMissingPerson {
	m.mu.RLock()
	defer m.mu.RUnlock()

	records := make([]models.PublicMissingPerson, 0)
	for _, record := range m.records {
		if record.PublicVisibility != "public" || record.ReviewStatus != "approved" {
			continue
		}
		if record.Status != "active" && record.Status != "located" {
			continue
		}
		if !matchesFilter(record, filter) {
			continue
		}
		records = append(records, toPublic(record))
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].UpdatedAt.After(records[j].UpdatedAt)
	})
	return records
}

// CreateMissingPerson creates a private pending-review record.
func (m *MemoryStore) CreateMissingPerson(request models.CreateMissingPersonRequest, createdBy string, now time.Time) models.MissingPerson {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seq++
	record := models.MissingPerson{
		ID:                fmt.Sprintf("missing_%03d", m.seq),
		Reference:         fmt.Sprintf("MP-%s-%03d", now.Format("20060102"), m.seq),
		PersonName:        request.PersonName,
		Age:               request.Age,
		Gender:            request.Gender,
		Description:       request.Description,
		PhotoURL:          request.PhotoURL,
		LastSeenAt:        request.LastSeenAt,
		LastSeenLocation:  request.LastSeenLocation,
		RelatedIncidentID: request.RelatedIncidentID,
		Reporter:          request.Reporter,
		Status:            "pending_review",
		ReviewStatus:      "pending",
		PublicVisibility:  "private",
		CreatedBy:         createdBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	m.records = append(m.records, record)
	m.appendAuditLocked(record.ID, "missing_person.created", models.AuthorityContext{ActorUserID: createdBy}, "Public intake received.", now)
	return record
}

// GetMissingPerson returns a full record by ID.
func (m *MemoryStore) GetMissingPerson(id string) (models.MissingPerson, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, record := range m.records {
		if record.ID == strings.TrimSpace(id) {
			return record, true
		}
	}
	return models.MissingPerson{}, false
}

// GetPublicMissingPerson returns a public-safe record by ID.
func (m *MemoryStore) GetPublicMissingPerson(id string) (models.PublicMissingPerson, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, record := range m.records {
		if record.ID != strings.TrimSpace(id) || record.PublicVisibility != "public" || record.ReviewStatus != "approved" {
			continue
		}
		if record.Status != "active" && record.Status != "located" {
			continue
		}
		return toPublic(record), true
	}
	return models.PublicMissingPerson{}, false
}

// ReviewMissingPerson applies authority review and public visibility controls.
func (m *MemoryStore) ReviewMissingPerson(id string, request models.ReviewMissingPersonRequest, ctx models.AuthorityContext, now time.Time) (models.MissingPerson, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index := range m.records {
		if m.records[index].ID != strings.TrimSpace(id) {
			continue
		}
		next := m.records[index]
		switch request.Decision {
		case "approve_public":
			if !next.Reporter.ConsentToPublicShare && !request.ConsentOverride {
				return models.MissingPerson{}, "public_consent_required", "reporter declined consent to public sharing; approve_public requires an explicit consentOverride"
			}
			next.ReviewStatus = "approved"
			next.PublicVisibility = "public"
			next.Status = statusOrDefault(request.Status, "active")
		case "approve_private":
			next.ReviewStatus = "approved"
			next.PublicVisibility = "private"
			next.Status = statusOrDefault(request.Status, "active")
		case "reject":
			next.ReviewStatus = "rejected"
			next.PublicVisibility = "private"
			next.Status = "rejected"
		default:
			return models.MissingPerson{}, "invalid_decision", "decision must be approve_public, approve_private, or reject"
		}
		next.PublicSummary = request.PublicSummary
		next.ReviewNotes = request.ReviewNotes
		next.ReviewedBy = ctx.ActorUserID
		next.ReviewedAt = &now
		next.ClosureType = ""
		next.ClosureNotes = ""
		next.ClosedBy = ""
		next.ClosedAt = nil
		next.UpdatedAt = now
		m.records[index] = next
		auditNotes := request.ReviewNotes
		if request.Decision == "approve_public" && request.ConsentOverride {
			auditNotes = strings.TrimSpace(auditNotes + " [consentOverride=true]")
		}
		m.appendAuditLocked(next.ID, "missing_person.reviewed", ctx, auditNotes, now)
		return next, "", ""
	}
	return models.MissingPerson{}, "not_found", "missing person record was not found"
}

// CloseMissingPerson applies closure or reunification status and writes audit.
func (m *MemoryStore) CloseMissingPerson(id string, request models.CloseMissingPersonRequest, ctx models.AuthorityContext, now time.Time) (models.MissingPerson, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index := range m.records {
		if m.records[index].ID != strings.TrimSpace(id) {
			continue
		}
		next := m.records[index]
		next.ClosureType = request.ClosureType
		next.ClosureNotes = request.ClosureNotes
		switch {
		case request.ReunitedWithFamily || request.ClosureType == "reunited":
			next.Status = "reunited"
		case request.ClosureType == "located_safe":
			next.Status = "located"
		default:
			next.Status = "closed"
		}
		next.PublicVisibility = "private"
		next.ClosedBy = ctx.ActorUserID
		next.ClosedAt = &now
		next.UpdatedAt = now
		m.records[index] = next
		m.appendAuditLocked(next.ID, "missing_person.closed", ctx, request.ClosureNotes, now)
		return next, "", ""
	}
	return models.MissingPerson{}, "not_found", "missing person record was not found"
}

// ListAudit returns audit entries for a record.
func (m *MemoryStore) ListAudit(id string) []models.MissingPersonAuditEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := make([]models.MissingPersonAuditEntry, 0)
	for _, entry := range m.audits {
		if entry.RecordID == strings.TrimSpace(id) {
			entries = append(entries, entry)
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].CreatedAt.Before(entries[j].CreatedAt)
	})
	return append([]models.MissingPersonAuditEntry{}, entries...)
}

func (m *MemoryStore) appendAuditLocked(recordID string, action string, ctx models.AuthorityContext, notes string, now time.Time) {
	m.auditSeq++
	m.audits = append(m.audits, models.MissingPersonAuditEntry{
		ID:            fmt.Sprintf("audit_%03d", m.auditSeq),
		RecordID:      recordID,
		Action:        action,
		ActorUserID:   ctx.ActorUserID,
		ActorAgencyID: ctx.ActorAgencyID,
		ActorRole:     ctx.ActorRole,
		Notes:         notes,
		CreatedAt:     now,
	})
}

func statusOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func matchesFilter(record models.MissingPerson, filter models.MissingPersonFilter) bool {
	if filter.Status != "" && record.Status != filter.Status {
		return false
	}
	if filter.District != "" && !strings.Contains(strings.ToLower(record.LastSeenLocation.District), filter.District) {
		return false
	}
	if filter.Query != "" && !recordMatchesQuery(record, filter.Query) {
		return false
	}
	return true
}

func recordMatchesQuery(record models.MissingPerson, query string) bool {
	fields := []string{
		record.Reference,
		record.PersonName,
		record.Description,
		record.LastSeenLocation.Label,
		record.LastSeenLocation.District,
		record.RelatedIncidentID,
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}

func sortRecords(records []models.MissingPerson) {
	sort.Slice(records, func(i, j int) bool {
		if records[i].Status != records[j].Status {
			return statusRank(records[i].Status) < statusRank(records[j].Status)
		}
		return records[i].UpdatedAt.After(records[j].UpdatedAt)
	})
}

func statusRank(status string) int {
	switch status {
	case "pending_review":
		return 0
	case "active":
		return 1
	case "located":
		return 2
	case "reunited":
		return 3
	case "closed":
		return 4
	case "rejected":
		return 5
	default:
		return 9
	}
}

func copyMissingPersons(records []models.MissingPerson) []models.MissingPerson {
	copied := append([]models.MissingPerson{}, records...)
	return copied
}

func toPublic(record models.MissingPerson) models.PublicMissingPerson {
	return models.PublicMissingPerson{
		ID:                record.ID,
		Reference:         record.Reference,
		PersonName:        record.PersonName,
		Age:               record.Age,
		Gender:            record.Gender,
		Description:       record.Description,
		PhotoURL:          record.PhotoURL,
		LastSeenAt:        record.LastSeenAt,
		LastSeenLocation:  record.LastSeenLocation,
		RelatedIncidentID: record.RelatedIncidentID,
		Status:            record.Status,
		PublicSummary:     record.PublicSummary,
		UpdatedAt:         record.UpdatedAt,
	}
}
