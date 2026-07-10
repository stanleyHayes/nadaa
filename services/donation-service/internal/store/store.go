package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
)

// Store is the persistence interface for donation data.
type Store interface {
	ListDonors(filter models.DonorFilter) []models.Donor
	CreateDonor(request models.CreateDonorRequest, createdBy string, now time.Time) models.Donor
	GetDonor(id string) (models.Donor, bool)
	UpdateDonor(id string, request models.UpdateDonorRequest, updatedBy string, now time.Time) (models.Donor, string, string)
	ListCatalog() []models.AidCatalogItem
	ListAidRequests(filter models.AidRequestFilter) []models.AidRequest
	CreateAidRequest(request models.CreateAidRequestRequest, requestedBy string, now time.Time) models.AidRequest
	GetAidRequest(id string) (models.AidRequest, bool)
	UpdateAidRequest(id string, request models.UpdateAidRequestRequest, updatedBy string, now time.Time) (models.AidRequest, string, string)
	CreatePledge(aidRequestID string, request models.CreatePledgeRequest, now time.Time) (models.Pledge, string, string)
	ListPledgesForRequest(aidRequestID string) []models.Pledge
	ListPledges(filter models.PledgeFilter) []models.Pledge
	UpdatePledge(id string, request models.UpdatePledgeRequest, updatedBy string, now time.Time) (models.Pledge, string, string)
	AllocatePledge(aidRequestID string, request models.AllocateRequest, updatedBy string, now time.Time) (models.Pledge, string, string)
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu       sync.RWMutex
	seq      int
	donors   []models.Donor
	catalog  []models.AidCatalogItem
	requests []models.AidRequest
	pledges  []models.Pledge
}

type seqHolder struct {
	value int
}

// NewMemoryStore creates an in-memory store seeded with fixtures.
func NewMemoryStore(now time.Time) Store {
	store := &MemoryStore{}
	store.catalog = seedCatalog(now)
	store.requests = seedAidRequests(now)
	store.seq = len(store.catalog)
	return store
}

func seedCatalog(_ time.Time) []models.AidCatalogItem {
	return []models.AidCatalogItem{
		{
			ID:            nextIDFor(&seqHolder{value: 1}, "catalog"),
			Code:          "food_parcel",
			Name:          "Ready-to-eat food parcels",
			Category:      "food",
			DefaultUnit:   "parcels",
			PriorityScore: 90,
		},
		{
			ID:            nextIDFor(&seqHolder{value: 2}, "catalog"),
			Code:          "water_liter",
			Name:          "Clean drinking water",
			Category:      "water",
			DefaultUnit:   "liters",
			PriorityScore: 95,
		},
		{
			ID:            nextIDFor(&seqHolder{value: 3}, "catalog"),
			Code:          "medical_kit",
			Name:          "Emergency medical kit",
			Category:      "medical",
			DefaultUnit:   "kits",
			PriorityScore: 100,
		},
		{
			ID:            nextIDFor(&seqHolder{value: 4}, "catalog"),
			Code:          "shelter_kit",
			Name:          "Family shelter kit",
			Category:      "shelter",
			DefaultUnit:   "kits",
			PriorityScore: 85,
		},
		{
			ID:            nextIDFor(&seqHolder{value: 5}, "catalog"),
			Code:          "hygiene_kit",
			Name:          "Hygiene and sanitation kit",
			Category:      "sanitation",
			DefaultUnit:   "kits",
			PriorityScore: 80,
		},
	}
}

func nextIDFor(holder *seqHolder, prefix string) string {
	id := fmt.Sprintf("%s_%03d", prefix, holder.value)
	holder.value++
	return id
}

func seedAidRequests(now time.Time) []models.AidRequest {
	return []models.AidRequest{
		{
			ID:                "request_001",
			Reference:         "AR-20260707-001",
			Title:             "Flood relief food parcels for Accra Metropolitan",
			Description:       "Ready-to-eat food parcels for households displaced by flooding in central Accra.",
			Category:          "food",
			ItemCode:          "food_parcel",
			QuantityNeeded:    500,
			QuantityFulfilled: 0,
			Unit:              "parcels",
			Priority:          "high",
			LocationLabel:     "Accra Metropolitan Assembly Hall",
			Region:            "Greater Accra",
			District:          "Accra Metropolitan",
			BeneficiaryCount:  2500,
			Status:            "open",
			RequestedBy:       "seed",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		{
			ID:                "request_002",
			Reference:         "AR-20260707-002",
			Title:             "Emergency medical supplies for Tema",
			Description:       "First-aid and emergency medical kits for flood-affected communities in Tema.",
			Category:          "medical",
			ItemCode:          "medical_kit",
			QuantityNeeded:    200,
			QuantityFulfilled: 0,
			Unit:              "kits",
			Priority:          "critical",
			LocationLabel:     "Tema General Hospital",
			Region:            "Greater Accra",
			District:          "Tema Metropolitan",
			BeneficiaryCount:  800,
			Status:            "open",
			RequestedBy:       "seed",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}
}

func (m *MemoryStore) nextSeq() int {
	m.seq++
	return m.seq
}

func (m *MemoryStore) generateRef(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%s-%03d", prefix, now.Format("20060102"), m.nextSeq())
}

// ListDonors returns donors matching the provided filter.
func (m *MemoryStore) ListDonors(filter models.DonorFilter) []models.Donor {
	m.mu.RLock()
	defer m.mu.RUnlock()

	donors := make([]models.Donor, 0)
	for _, donor := range m.donors {
		if filter.Type != "" && donor.Type != filter.Type {
			continue
		}
		if filter.Query != "" && !donorMatchesQuery(donor, filter.Query) {
			continue
		}
		donors = append(donors, donor)
	}
	sort.Slice(donors, func(i, j int) bool {
		if donors[i].Status != donors[j].Status {
			return donorStatusRank(donors[i].Status) < donorStatusRank(donors[j].Status)
		}
		return donors[i].Name < donors[j].Name
	})
	return copyDonors(donors)
}

func donorMatchesQuery(donor models.Donor, query string) bool {
	fields := []string{
		strings.ToLower(donor.Name),
		strings.ToLower(donor.ContactName),
		strings.ToLower(donor.ContactEmail),
		strings.ToLower(donor.Region),
		strings.ToLower(donor.District),
	}
	for _, field := range fields {
		if strings.Contains(field, query) {
			return true
		}
	}
	return false
}

// CreateDonor creates a new donor record.
func (m *MemoryStore) CreateDonor(request models.CreateDonorRequest, createdBy string, now time.Time) models.Donor {
	m.mu.Lock()
	defer m.mu.Unlock()

	donor := models.Donor{
		ID:                fmt.Sprintf("donor_%03d", m.nextSeq()),
		Reference:         m.generateRef("DON", now),
		Name:              request.Name,
		Type:              request.Type,
		ContactName:       request.ContactName,
		ContactEmail:      request.ContactEmail,
		ContactPhone:      request.ContactPhone,
		Region:            request.Region,
		District:          request.District,
		ItemsOffered:      append([]string{}, request.ItemsOffered...),
		MonetaryPledgeGhs: request.MonetaryPledgeGhs,
		Status:            "active",
		Notes:             request.Notes,
		CreatedBy:         createdBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	m.donors = append(m.donors, donor)
	return donor
}

// GetDonor returns a donor by id.
func (m *MemoryStore) GetDonor(id string) (models.Donor, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, donor := range m.donors {
		if donor.ID == id {
			return donor, true
		}
	}
	return models.Donor{}, false
}

// UpdateDonor updates a donor and returns the updated record or an error code.
func (m *MemoryStore) UpdateDonor(id string, request models.UpdateDonorRequest, _ string, now time.Time) (models.Donor, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.donors {
		if m.donors[index].ID != id {
			continue
		}

		next := m.donors[index]
		if request.Status != "" {
			next.Status = request.Status
		}
		next.Notes = request.Notes
		next.UpdatedAt = now
		m.donors[index] = next
		return next, "", ""
	}

	return models.Donor{}, "not_found", "donor was not found"
}

// ListCatalog returns all aid catalog items sorted by priority descending.
func (m *MemoryStore) ListCatalog() []models.AidCatalogItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]models.AidCatalogItem, len(m.catalog))
	copy(items, m.catalog)
	sort.Slice(items, func(i, j int) bool {
		if items[i].PriorityScore != items[j].PriorityScore {
			return items[i].PriorityScore > items[j].PriorityScore
		}
		return items[i].Name < items[j].Name
	})
	return items
}

// ListAidRequests returns aid requests matching the provided filter.
func (m *MemoryStore) ListAidRequests(filter models.AidRequestFilter) []models.AidRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requests := make([]models.AidRequest, 0)
	for _, req := range m.requests {
		if filter.Status != "" && req.Status != filter.Status {
			continue
		}
		if filter.Category != "" && req.Category != filter.Category {
			continue
		}
		if filter.Region != "" && !strings.Contains(strings.ToLower(req.Region), filter.Region) {
			continue
		}
		if filter.Priority != "" && req.Priority != filter.Priority {
			continue
		}
		requests = append(requests, req)
	}
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].Priority != requests[j].Priority {
			return priorityRank(requests[i].Priority) < priorityRank(requests[j].Priority)
		}
		return requests[i].CreatedAt.After(requests[j].CreatedAt)
	})
	return copyAidRequests(requests)
}

// CreateAidRequest creates a new aid request record.
func (m *MemoryStore) CreateAidRequest(request models.CreateAidRequestRequest, requestedBy string, now time.Time) models.AidRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	req := models.AidRequest{
		ID:                fmt.Sprintf("request_%03d", m.nextSeq()),
		Reference:         m.generateRef("AR", now),
		Title:             request.Title,
		Description:       request.Description,
		Category:          request.Category,
		ItemCode:          request.ItemCode,
		QuantityNeeded:    request.QuantityNeeded,
		QuantityFulfilled: 0,
		Unit:              request.Unit,
		Priority:          request.Priority,
		LocationLabel:     request.LocationLabel,
		Region:            request.Region,
		District:          request.District,
		BeneficiaryCount:  request.BeneficiaryCount,
		Status:            "open",
		RequestedBy:       requestedBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	m.requests = append(m.requests, req)
	return req
}

// GetAidRequest returns an aid request by id.
func (m *MemoryStore) GetAidRequest(id string) (models.AidRequest, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, req := range m.requests {
		if req.ID == id {
			return req, true
		}
	}
	return models.AidRequest{}, false
}

// UpdateAidRequest updates an aid request and returns the updated record or an error code.
func (m *MemoryStore) UpdateAidRequest(id string, request models.UpdateAidRequestRequest, _ string, now time.Time) (models.AidRequest, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.requests {
		if m.requests[index].ID != id {
			continue
		}

		next := m.requests[index]
		if request.Status != "" {
			next.Status = request.Status
		}
		if request.QuantityNeeded > 0 {
			next.QuantityNeeded = request.QuantityNeeded
		}
		next.UpdatedAt = now
		next = m.recalcSingleRequest(next, now)
		m.requests[index] = next
		return next, "", ""
	}

	return models.AidRequest{}, "not_found", "aid request was not found"
}

// CreatePledge creates a pledge against an aid request.
func (m *MemoryStore) CreatePledge(aidRequestID string, request models.CreatePledgeRequest, now time.Time) (models.Pledge, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	request.DonorID = strings.TrimSpace(request.DonorID)

	var donor models.Donor
	found := false
	for _, d := range m.donors {
		if d.ID == request.DonorID {
			donor = d
			found = true
			break
		}
	}
	if !found {
		return models.Pledge{}, "donor_not_found", "donor was not found"
	}

	requestIndex := -1
	for index := range m.requests {
		if m.requests[index].ID == aidRequestID {
			requestIndex = index
			break
		}
	}
	if requestIndex == -1 {
		return models.Pledge{}, "not_found", "aid request was not found"
	}

	donorName := strings.TrimSpace(request.DonorName)
	if donorName == "" {
		donorName = donor.Name
	}
	contactEmail := strings.TrimSpace(request.ContactEmail)
	if contactEmail == "" {
		contactEmail = donor.ContactEmail
	}
	contactPhone := strings.TrimSpace(request.ContactPhone)
	if contactPhone == "" {
		contactPhone = donor.ContactPhone
	}

	pledge := models.Pledge{
		ID:              fmt.Sprintf("pledge_%03d", m.nextSeq()),
		Reference:       m.generateRef("PL", now),
		AidRequestID:    aidRequestID,
		DonorID:         donor.ID,
		DonorName:       donorName,
		QuantityPledged: request.QuantityPledged,
		Status:          "pledged",
		DeliveryNote:    request.DeliveryNote,
		ContactEmail:    contactEmail,
		ContactPhone:    contactPhone,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	m.pledges = append(m.pledges, pledge)

	m.requests[requestIndex] = m.recalcSingleRequest(m.requests[requestIndex], now)
	return pledge, "", ""
}

// ListPledgesForRequest returns pledges for a specific aid request.
func (m *MemoryStore) ListPledgesForRequest(aidRequestID string) []models.Pledge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	pledges := make([]models.Pledge, 0)
	for _, pledge := range m.pledges {
		if pledge.AidRequestID == aidRequestID {
			pledges = append(pledges, pledge)
		}
	}
	sort.Slice(pledges, func(i, j int) bool {
		return pledges[i].CreatedAt.After(pledges[j].CreatedAt)
	})
	return copyPledges(pledges)
}

// ListPledges returns pledges matching the provided filter.
func (m *MemoryStore) ListPledges(filter models.PledgeFilter) []models.Pledge {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pledges := make([]models.Pledge, 0)
	for _, pledge := range m.pledges {
		if filter.Status != "" && pledge.Status != filter.Status {
			continue
		}
		pledges = append(pledges, pledge)
	}
	sort.Slice(pledges, func(i, j int) bool {
		return pledges[i].CreatedAt.After(pledges[j].CreatedAt)
	})
	return copyPledges(pledges)
}

// UpdatePledge updates a pledge and recalculates the related aid request.
func (m *MemoryStore) UpdatePledge(id string, request models.UpdatePledgeRequest, _ string, now time.Time) (models.Pledge, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	pledgeIndex := -1
	for index := range m.pledges {
		if m.pledges[index].ID == id {
			pledgeIndex = index
			break
		}
	}
	if pledgeIndex == -1 {
		return models.Pledge{}, "not_found", "pledge was not found"
	}

	next := m.pledges[pledgeIndex]
	if request.Status != "" {
		next.Status = request.Status
	}
	if request.QuantityDelivered > 0 {
		if request.QuantityDelivered > next.QuantityPledged {
			return models.Pledge{}, "invalid_quantity", "quantityDelivered cannot exceed quantityPledged"
		}
		next.QuantityDelivered = request.QuantityDelivered
	}
	if request.DeliveryNote != "" {
		next.DeliveryNote = request.DeliveryNote
	}
	next.UpdatedAt = now
	m.pledges[pledgeIndex] = next

	for index := range m.requests {
		if m.requests[index].ID == next.AidRequestID {
			m.requests[index] = m.recalcSingleRequest(m.requests[index], now)
			break
		}
	}
	return next, "", ""
}

// AllocatePledge marks a pledged quantity as delivered for an aid request.
func (m *MemoryStore) AllocatePledge(aidRequestID string, request models.AllocateRequest, _ string, now time.Time) (models.Pledge, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	request.PledgeID = strings.TrimSpace(request.PledgeID)

	pledgeIndex := -1
	for index := range m.pledges {
		if m.pledges[index].ID == request.PledgeID && m.pledges[index].AidRequestID == aidRequestID {
			pledgeIndex = index
			break
		}
	}
	if pledgeIndex == -1 {
		return models.Pledge{}, "not_found", "pledge was not found for this aid request"
	}

	if request.Quantity > m.pledges[pledgeIndex].QuantityPledged {
		return models.Pledge{}, "invalid_quantity", "quantity cannot exceed quantityPledged"
	}

	next := m.pledges[pledgeIndex]
	next.Status = "delivered"
	next.QuantityDelivered = request.Quantity
	next.UpdatedAt = now
	m.pledges[pledgeIndex] = next

	for index := range m.requests {
		if m.requests[index].ID == aidRequestID {
			m.requests[index] = m.recalcSingleRequest(m.requests[index], now)
			break
		}
	}
	return next, "", ""
}

func (m *MemoryStore) recalcSingleRequest(req models.AidRequest, now time.Time) models.AidRequest {
	fulfilled := 0
	for _, pledge := range m.pledges {
		if pledge.AidRequestID != req.ID {
			continue
		}
		if pledge.Status == "cancelled" {
			continue
		}
		if pledge.Status == "delivered" {
			fulfilled += pledge.QuantityDelivered
		} else {
			fulfilled += pledge.QuantityPledged
		}
	}
	req.QuantityFulfilled = fulfilled
	if req.Status != "closed" {
		switch {
		case fulfilled >= req.QuantityNeeded:
			req.Status = "fulfilled"
		case fulfilled > 0:
			req.Status = "partially_fulfilled"
		default:
			req.Status = "open"
		}
	}
	req.UpdatedAt = now
	return req
}

func donorStatusRank(status string) int {
	switch status {
	case "active":
		return 0
	case "inactive":
		return 1
	default:
		return 2
	}
}

func priorityRank(priority string) int {
	switch priority {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

func copyDonors(source []models.Donor) []models.Donor {
	donors := make([]models.Donor, len(source))
	copy(donors, source)
	for i := range donors {
		donors[i].ItemsOffered = append([]string{}, donors[i].ItemsOffered...)
	}
	return donors
}

func copyAidRequests(source []models.AidRequest) []models.AidRequest {
	requests := make([]models.AidRequest, len(source))
	copy(requests, source)
	return requests
}

func copyPledges(source []models.Pledge) []models.Pledge {
	pledges := make([]models.Pledge, len(source))
	copy(pledges, source)
	return pledges
}
