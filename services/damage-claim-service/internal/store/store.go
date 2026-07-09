package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
)

// Store is the persistence interface for damage claims.
type Store interface {
	Health() string
	Create(req models.CreateClaimRequest, incidentRef, incidentLocation string, now time.Time) models.DamageClaimRecord
	List(filter models.ListClaimsFilter) []models.DamageClaimRecord
	Get(id string) (models.DamageClaimRecord, bool)
	Update(id string, req models.UpdateClaimRequest, now time.Time) (models.DamageClaimRecord, bool)
	Verify(id string, req models.VerifyClaimRequest, verifiedBy string, now time.Time) (models.DamageClaimRecord, string)
	Close(id string, reason string, closedBy string, now time.Time) (models.DamageClaimRecord, bool)
}

// MemoryStore is an in-memory implementation of Store with fixture seed data.
type MemoryStore struct {
	mu       sync.RWMutex
	claims   map[string]models.DamageClaimRecord
	sequence int
}

// NewMemoryStore creates an in-memory store seeded with fixture claims.
func NewMemoryStore(now time.Time) Store {
	s := &MemoryStore{
		claims:   make(map[string]models.DamageClaimRecord),
		sequence: 3,
	}
	s.claims["claim_001"] = models.DamageClaimRecord{
		ID:                  "claim_001",
		Reference:           fmt.Sprintf("DC-%s-%05d", now.Format("2006"), 1),
		IncidentID:          "inc_001",
		IncidentReference:   "NADAA-ACC-20260706-001",
		IncidentLocation:    "Accra Central",
		Reporter:            models.ReporterInfo{Name: "Ama Mensah", Phone: "+233241234567", Email: "ama.mensah@example.com", UserID: "usr_citizen_001"},
		DamageType:          "flood",
		DamageDescription:   "Ground floor flooded; furniture and appliances damaged.",
		EstimatedLossAmount: "12500.00",
		DamagePhotos:        []string{"https://media.nadaa.local/claims/claim_001/photo1.jpg"},
		Location:            models.ClaimLocation{Lat: 5.6037, Lng: -0.1870, Address: "Accra Central"},
		VerificationStatus:  "pending",
		Status:              "submitted",
		PrivacyConsent:      true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	s.claims["claim_002"] = models.DamageClaimRecord{
		ID:                  "claim_002",
		Reference:           fmt.Sprintf("DC-%s-%05d", now.Format("2006"), 2),
		IncidentID:          "inc_002",
		IncidentReference:   "NADAA-TEMA-20260706-002",
		IncidentLocation:    "Tema Community 12",
		Reporter:            models.ReporterInfo{Name: "Kwame Asare", Phone: "+233209876543", UserID: "usr_citizen_002"},
		DamageType:          "structural",
		DamageDescription:   "Roof partially collapsed after heavy rains.",
		EstimatedLossAmount: "8700.00",
		DamagePhotos:        []string{"https://media.nadaa.local/claims/claim_002/photo1.jpg", "https://media.nadaa.local/claims/claim_002/photo2.jpg"},
		Location:            models.ClaimLocation{Lat: 5.6690, Lng: -0.0160, Address: "Tema Community 12"},
		VerificationStatus:  "verified",
		VerifiedBy:          "usr_insurance_officer",
		VerifiedAt:          timePtr(now.Add(-2 * time.Hour)),
		VerificationNotes:   "Photos consistent with weather report; engineer report attached.",
		Status:              "submitted",
		PrivacyConsent:      true,
		CreatedAt:           now.Add(-4 * time.Hour),
		UpdatedAt:           now.Add(-2 * time.Hour),
	}
	s.claims["claim_003"] = models.DamageClaimRecord{
		ID:                  "claim_003",
		Reference:           fmt.Sprintf("DC-%s-%05d", now.Format("2006"), 3),
		IncidentID:          "",
		Reporter:            models.ReporterInfo{Name: "Yaa Boakye", Phone: "+233244445555", UserID: "usr_citizen_003"},
		DamageType:          "vehicle",
		DamageDescription:   "Vehicle swept away by flood water; total loss suspected.",
		EstimatedLossAmount: "45000.00",
		DamagePhotos:        []string{},
		Location:            models.ClaimLocation{Lat: 5.5800, Lng: -0.2100, Address: "Kaneshie Market Road"},
		VerificationStatus:  "pending",
		Status:              "draft",
		PrivacyConsent:      true,
		CreatedAt:           now.Add(-1 * time.Hour),
		UpdatedAt:           now.Add(-1 * time.Hour),
	}
	return s
}

// Health returns a simple health indicator.
func (m *MemoryStore) Health() string {
	return "ok"
}

// Create stores a new damage claim.
func (m *MemoryStore) Create(req models.CreateClaimRequest, incidentRef, incidentLocation string, now time.Time) models.DamageClaimRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sequence++
	claim := models.DamageClaimRecord{
		ID:                  fmt.Sprintf("claim_%03d", m.sequence),
		Reference:           fmt.Sprintf("DC-%s-%05d", now.Format("2006"), m.sequence),
		IncidentID:          req.IncidentID,
		IncidentReference:   incidentRef,
		IncidentLocation:    incidentLocation,
		Reporter:            req.Reporter,
		DamageType:          req.DamageType,
		DamageDescription:   req.DamageDescription,
		EstimatedLossAmount: req.EstimatedLossAmount,
		DamagePhotos:        append([]string{}, req.DamagePhotos...),
		Location:            req.Location,
		VerificationStatus:  "pending",
		Status:              "submitted",
		PrivacyConsent:      req.PrivacyConsent,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	m.claims[claim.ID] = claim
	return copyClaim(claim)
}

// List returns filtered and sorted claims.
func (m *MemoryStore) List(filter models.ListClaimsFilter) []models.DamageClaimRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query := strings.ToLower(strings.TrimSpace(filter.Query))
	claims := make([]models.DamageClaimRecord, 0, len(m.claims))
	for _, claim := range m.claims {
		if filter.Status != "" && claim.Status != filter.Status {
			continue
		}
		if filter.VerificationStatus != "" && claim.VerificationStatus != filter.VerificationStatus {
			continue
		}
		if filter.IncidentID != "" && claim.IncidentID != filter.IncidentID {
			continue
		}
		if query != "" {
			match := strings.Contains(strings.ToLower(claim.Reference), query) ||
				strings.Contains(strings.ToLower(claim.Reporter.Name), query) ||
				strings.Contains(strings.ToLower(claim.DamageType), query) ||
				strings.Contains(strings.ToLower(claim.DamageDescription), query) ||
				strings.Contains(strings.ToLower(claim.Location.Address), query)
			if !match {
				continue
			}
		}
		claims = append(claims, copyClaim(claim))
	}

	sort.Slice(claims, func(i, j int) bool {
		return claims[i].CreatedAt.After(claims[j].CreatedAt)
	})
	return claims
}

// Get retrieves a claim by id.
func (m *MemoryStore) Get(id string) (models.DamageClaimRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	claim, ok := m.claims[strings.TrimSpace(id)]
	if !ok {
		return models.DamageClaimRecord{}, false
	}
	return copyClaim(claim), true
}

// Update modifies the description, amount, and/or photos of a claim.
func (m *MemoryStore) Update(id string, req models.UpdateClaimRequest, now time.Time) (models.DamageClaimRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	claim, ok := m.claims[strings.TrimSpace(id)]
	if !ok {
		return models.DamageClaimRecord{}, false
	}
	if req.DamageDescription != nil {
		claim.DamageDescription = *req.DamageDescription
	}
	if req.EstimatedLossAmount != nil {
		claim.EstimatedLossAmount = *req.EstimatedLossAmount
	}
	if req.DamagePhotos != nil {
		claim.DamagePhotos = append([]string{}, req.DamagePhotos...)
	}
	claim.UpdatedAt = now
	m.claims[claim.ID] = claim
	return copyClaim(claim), true
}

// Verify transitions a claim from pending to verified or rejected.
func (m *MemoryStore) Verify(id string, req models.VerifyClaimRequest, verifiedBy string, now time.Time) (models.DamageClaimRecord, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	claim, ok := m.claims[strings.TrimSpace(id)]
	if !ok {
		return models.DamageClaimRecord{}, "not_found"
	}
	if claim.VerificationStatus != "pending" {
		return models.DamageClaimRecord{}, "invalid_verification_transition"
	}
	claim.VerificationStatus = req.VerificationStatus
	claim.VerifiedBy = verifiedBy
	claim.VerifiedAt = &now
	claim.VerificationNotes = req.Notes
	claim.UpdatedAt = now
	m.claims[claim.ID] = claim
	return copyClaim(claim), ""
}

// Close marks a claim as closed and appends a closure note.
func (m *MemoryStore) Close(id string, reason string, closedBy string, now time.Time) (models.DamageClaimRecord, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	claim, ok := m.claims[strings.TrimSpace(id)]
	if !ok {
		return models.DamageClaimRecord{}, false
	}
	claim.Status = "closed"
	claim.VerificationNotes = appendNote(claim.VerificationNotes, fmt.Sprintf("Closed by %s: %s", closedBy, reason))
	claim.UpdatedAt = now
	m.claims[claim.ID] = claim
	return copyClaim(claim), true
}

func copyClaim(claim models.DamageClaimRecord) models.DamageClaimRecord {
	claim.DamagePhotos = append([]string{}, claim.DamagePhotos...)
	if claim.VerifiedAt != nil {
		verified := *claim.VerifiedAt
		claim.VerifiedAt = &verified
	}
	return claim
}

func appendNote(existing, note string) string {
	if existing == "" {
		return note
	}
	return existing + "\n" + note
}

func timePtr(t time.Time) *time.Time {
	return &t
}
