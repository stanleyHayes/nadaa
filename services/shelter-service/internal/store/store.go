package store

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/utils"
)

// Store is the persistence interface for shelter data.
type Store interface {
	ListShelters() []models.Shelter
	NearbyShelters(location models.Coordinates, limit int) []models.Shelter
	NearbyRecoverySupport(location models.Coordinates, limit int) []models.RecoverySupportLocation
	UpdateShelter(id string, request models.OccupancyUpdateRequest, ctx models.AuthorityContext, now time.Time) (models.Shelter, string, string)
	ListHospitalCapacity(filter models.HospitalCapacityFilter, now time.Time) []models.HospitalCapacity
	UpdateHospitalCapacity(id string, request models.HospitalCapacityUpdateRequest, ctx models.AuthorityContext, now time.Time) (models.HospitalCapacity, string, string)
	ImportHospitalCapacityFixture(request models.HospitalCapacityImportRequest, ctx models.AuthorityContext, now time.Time) ([]models.HospitalCapacity, int)
	ListReliefPoints(filter models.ReliefPointFilter) []models.ReliefPoint
	NearbyReliefPoints(location models.Coordinates, limit int) []models.ReliefPoint
	CreateReliefPoint(request models.CreateReliefPointRequest, ctx models.AuthorityContext, now time.Time) models.ReliefPoint
	UpdateReliefPoint(id string, request models.UpdateReliefPointRequest, ctx models.AuthorityContext, now time.Time) (models.ReliefPoint, string, string)
	ListReliefPointStockHistory(reliefPointID string) []models.ReliefStockHistory
	ListAidRequests(filter models.AidRequestFilter) []models.AidRequest
	CreateAidRequest(request models.CreateAidRequestRequest, ctx models.AuthorityContext, now time.Time) models.AidRequest
	ReviewAidRequest(id string, request models.ReviewAidRequestRequest, ctx models.AuthorityContext, now time.Time) (models.AidRequest, string, string)
	ListAidPledges(aidRequestID string) ([]models.AidPledge, string, string)
	CreateAidPledge(aidRequestID string, request models.CreateAidPledgeRequest, now time.Time) (models.AidPledge, string, string)
	ReviewAidPledge(aidRequestID, pledgeID string, request models.ReviewAidPledgeRequest, ctx models.AuthorityContext, now time.Time) (models.AidPledge, string, string)
	DeleteShelter(id string) bool
	DeleteReliefPoint(id string) bool
	DeleteAidRequest(id string) bool
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu                 sync.RWMutex
	shelters           []models.Shelter
	recovery           []models.RecoverySupportLocation
	hospitals          []models.HospitalCapacity
	reliefPoints       []models.ReliefPoint
	reliefStockHistory []models.ReliefStockHistory
	aidRequests        []models.AidRequest
	aidPledges         []models.AidPledge
	// Monotonic ID sequences, independent of slice length, so deletes can
	// never cause a newly created record to reuse an existing ID.
	reliefPointSeq int
	aidRequestSeq  int
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
func NewMemoryStore(now time.Time) Store {
	reliefPoints := seedReliefPoints(now)
	aidRequests := seedAidRequests(now)
	return &MemoryStore{
		shelters:     seedShelters(now),
		recovery:     seedRecovery(now),
		hospitals:    seedHospitals(now),
		reliefPoints: reliefPoints,
		aidRequests:  aidRequests,
		aidPledges:   seedAidPledges(now),
		// Seeded fixture IDs do not use the sequential relief_NNN/aid_NNN
		// format, so starting past the seed count keeps generated IDs unique.
		reliefPointSeq: len(reliefPoints) + 1,
		aidRequestSeq:  len(aidRequests) + 1,
	}
}

// ListShelters returns all shelters sorted by status and name.
func (m *MemoryStore) ListShelters() []models.Shelter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shelters := copyShelters(m.shelters)
	sort.Slice(shelters, func(i, j int) bool {
		if shelters[i].Status == shelters[j].Status {
			return shelters[i].Name < shelters[j].Name
		}
		return shelterStatusRank(shelters[i].Status) < shelterStatusRank(shelters[j].Status)
	})
	return shelters
}

// NearbyShelters returns shelters near the given location.
func (m *MemoryStore) NearbyShelters(location models.Coordinates, limit int) []models.Shelter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shelters := make([]models.Shelter, 0, len(m.shelters))
	for _, shelter := range m.shelters {
		shelter.DistanceMeters = int(math.Round(utils.DistanceMeters(location, shelter.Location)))
		if float64(shelter.DistanceMeters) <= utils.NearbySearchMeters {
			shelters = append(shelters, shelter)
		}
	}

	sort.Slice(shelters, func(i, j int) bool {
		if shelters[i].DistanceMeters == shelters[j].DistanceMeters {
			return shelterStatusRank(shelters[i].Status) < shelterStatusRank(shelters[j].Status)
		}
		return shelters[i].DistanceMeters < shelters[j].DistanceMeters
	})
	if limit > 0 && len(shelters) > limit {
		shelters = shelters[:limit]
	}
	return copyShelters(shelters)
}

// NearbyRecoverySupport returns recovery support locations near the given location.
func (m *MemoryStore) NearbyRecoverySupport(location models.Coordinates, limit int) []models.RecoverySupportLocation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	locations := make([]models.RecoverySupportLocation, 0, len(m.recovery))
	for _, item := range m.recovery {
		item.DistanceMeters = int(math.Round(utils.DistanceMeters(location, item.Location)))
		if float64(item.DistanceMeters) <= utils.NearbySearchMeters {
			locations = append(locations, item)
		}
	}

	sort.Slice(locations, func(i, j int) bool {
		return locations[i].DistanceMeters < locations[j].DistanceMeters
	})
	if limit > 0 && len(locations) > limit {
		locations = locations[:limit]
	}
	return copyRecovery(locations)
}

// UpdateShelter updates a shelter's occupancy and status.
func (m *MemoryStore) UpdateShelter(id string, request models.OccupancyUpdateRequest, ctx models.AuthorityContext, now time.Time) (models.Shelter, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.shelters {
		if m.shelters[index].ID != id {
			continue
		}

		next := m.shelters[index]
		if request.Capacity != nil {
			next.Capacity = *request.Capacity
		}
		if request.CurrentOccupancy != nil {
			next.CurrentOccupancy = *request.CurrentOccupancy
		}
		// Validate the merged record so capacity-only or occupancy-only
		// updates cannot break the occupancy <= capacity invariant.
		if next.CurrentOccupancy > next.Capacity {
			return models.Shelter{}, "invalid_occupancy", "currentOccupancy cannot exceed capacity"
		}
		if request.Status != "" {
			next.Status = request.Status
		} else {
			next.Status = statusForOccupancy(next.Capacity, next.CurrentOccupancy, next.Status)
		}
		if request.Notes != "" {
			next.Notes = request.Notes
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		m.shelters[index] = next
		return next, "", ""
	}

	return models.Shelter{}, "not_found", "shelter was not found"
}

// DeleteShelter removes a shelter and reports whether it existed.
func (m *MemoryStore) DeleteShelter(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.shelters {
		if m.shelters[index].ID == id {
			m.shelters = append(m.shelters[:index], m.shelters[index+1:]...)
			return true
		}
	}
	return false
}

// ListReliefPoints returns relief points matching the filter.
func (m *MemoryStore) ListReliefPoints(filter models.ReliefPointFilter) []models.ReliefPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reliefPoints := make([]models.ReliefPoint, 0, len(m.reliefPoints))
	for _, point := range m.reliefPoints {
		if filter.Status != "" && point.Status != filter.Status {
			continue
		}
		if filter.Type != "" && point.Type != filter.Type {
			continue
		}
		if filter.BBox != nil && !pointInBBox(point.Location, *filter.BBox) {
			continue
		}
		if filter.Location != nil {
			distance := int(math.Round(utils.DistanceMeters(*filter.Location, point.Location)))
			if float64(distance) > filter.RadiusMeters {
				continue
			}
			point.DistanceMeters = distance
		}
		reliefPoints = append(reliefPoints, point)
	}

	sort.Slice(reliefPoints, func(i, j int) bool {
		if reliefPoints[i].Status != reliefPoints[j].Status {
			return reliefPointStatusRank(reliefPoints[i].Status) < reliefPointStatusRank(reliefPoints[j].Status)
		}
		if reliefPoints[i].DistanceMeters != reliefPoints[j].DistanceMeters {
			return reliefPoints[i].DistanceMeters < reliefPoints[j].DistanceMeters
		}
		return reliefPoints[i].Name < reliefPoints[j].Name
	})

	if filter.Limit > 0 && len(reliefPoints) > filter.Limit {
		reliefPoints = reliefPoints[:filter.Limit]
	}
	return copyReliefPoints(reliefPoints)
}

// NearbyReliefPoints returns relief points near the given location.
func (m *MemoryStore) NearbyReliefPoints(location models.Coordinates, limit int) []models.ReliefPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reliefPoints := make([]models.ReliefPoint, 0, len(m.reliefPoints))
	for _, point := range m.reliefPoints {
		point.DistanceMeters = int(math.Round(utils.DistanceMeters(location, point.Location)))
		if float64(point.DistanceMeters) <= utils.NearbySearchMeters {
			reliefPoints = append(reliefPoints, point)
		}
	}

	sort.Slice(reliefPoints, func(i, j int) bool {
		if reliefPoints[i].DistanceMeters == reliefPoints[j].DistanceMeters {
			return reliefPointStatusRank(reliefPoints[i].Status) < reliefPointStatusRank(reliefPoints[j].Status)
		}
		return reliefPoints[i].DistanceMeters < reliefPoints[j].DistanceMeters
	})

	if limit > 0 && len(reliefPoints) > limit {
		reliefPoints = reliefPoints[:limit]
	}
	return copyReliefPoints(reliefPoints)
}

// CreateReliefPoint creates a new relief point.
func (m *MemoryStore) CreateReliefPoint(request models.CreateReliefPointRequest, ctx models.AuthorityContext, now time.Time) models.ReliefPoint {
	m.mu.Lock()
	defer m.mu.Unlock()

	point := models.ReliefPoint{
		ID:              fmt.Sprintf("relief_%03d", m.reliefPointSeq),
		Name:            request.Name,
		Type:            request.Type,
		Region:          request.Region,
		District:        request.District,
		Address:         request.Address,
		Location:        request.Location,
		Contact:         request.Contact,
		OperatingHours:  request.OperatingHours,
		Eligibility:     request.Eligibility,
		Schedule:        request.Schedule,
		StockCategories: copyStockCategories(request.StockCategories),
		Status:          request.Status,
		Source:          request.Source,
		SourceRef:       request.SourceRef,
		CreatedBy:       ctx.ActorUserID,
		UpdatedBy:       ctx.ActorUserID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if point.Status == "" {
		point.Status = "open"
	}
	for i := range point.StockCategories {
		point.StockCategories[i].LastUpdated = now
	}
	m.reliefPointSeq++
	m.reliefPoints = append(m.reliefPoints, point)
	return point
}

// UpdateReliefPoint updates an existing relief point.
//
//nolint:gocognit // legacy complex function; refactor into validation/execution helpers in a future pass.
func (m *MemoryStore) UpdateReliefPoint(id string, request models.UpdateReliefPointRequest, ctx models.AuthorityContext, now time.Time) (models.ReliefPoint, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.reliefPoints {
		if m.reliefPoints[index].ID != id {
			continue
		}

		next := m.reliefPoints[index]
		stockChanged := false
		if request.Name != "" {
			next.Name = request.Name
		}
		if request.Type != "" {
			next.Type = request.Type
		}
		if request.Region != "" {
			next.Region = request.Region
		}
		if request.District != "" {
			next.District = request.District
		}
		if request.Address != "" {
			next.Address = request.Address
		}
		if request.Location != nil {
			next.Location = *request.Location
		}
		if request.Contact != "" {
			next.Contact = request.Contact
		}
		if request.OperatingHours != "" {
			next.OperatingHours = request.OperatingHours
		}
		if request.Eligibility != "" {
			next.Eligibility = request.Eligibility
		}
		if request.Schedule != "" {
			next.Schedule = request.Schedule
		}
		if request.Status != "" {
			next.Status = request.Status
		}
		if request.SourceRef != "" {
			next.SourceRef = request.SourceRef
		}
		if request.StockCategories != nil {
			for i := range request.StockCategories {
				request.StockCategories[i].LastUpdated = now
			}
			if !stockCategoriesEqual(next.StockCategories, request.StockCategories) {
				stockChanged = true
			}
			next.StockCategories = copyStockCategories(request.StockCategories)
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now

		if stockChanged {
			history := models.ReliefStockHistory{
				ID:              fmt.Sprintf("rsh_%03d", len(m.reliefStockHistory)+1),
				ReliefPointID:   next.ID,
				ChangedBy:       ctx.ActorUserID,
				ChangedAt:       now,
				StockCategories: copyStockCategories(next.StockCategories),
			}
			m.reliefStockHistory = append(m.reliefStockHistory, history)
		}

		m.reliefPoints[index] = next
		return next, "", ""
	}

	return models.ReliefPoint{}, "not_found", "relief point was not found"
}

// DeleteReliefPoint removes a relief point and reports whether it existed.
func (m *MemoryStore) DeleteReliefPoint(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.reliefPoints {
		if m.reliefPoints[index].ID == id {
			m.reliefPoints = append(m.reliefPoints[:index], m.reliefPoints[index+1:]...)
			return true
		}
	}
	return false
}

// ListReliefPointStockHistory returns stock history for a relief point.
func (m *MemoryStore) ListReliefPointStockHistory(reliefPointID string) []models.ReliefStockHistory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reliefPointID = strings.TrimSpace(reliefPointID)
	history := make([]models.ReliefStockHistory, 0)
	for _, record := range m.reliefStockHistory {
		if record.ReliefPointID == reliefPointID {
			history = append(history, record)
		}
	}
	sort.Slice(history, func(i, j int) bool {
		return history[i].ChangedAt.After(history[j].ChangedAt)
	})
	return history
}

// ListAidRequests returns aid requests matching the filter.
//
//nolint:gocognit // legacy complex function; refactor into smaller helpers in a future pass.
func (m *MemoryStore) ListAidRequests(filter models.AidRequestFilter) []models.AidRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	aidRequests := make([]models.AidRequest, 0, len(m.aidRequests))
	for _, request := range m.aidRequests {
		publicVisible := request.Visibility == "public" && utils.PublicAidRequestStatuses[request.Status]
		if !filter.IncludePrivate && !publicVisible {
			continue
		}
		if filter.IncludePrivate && !publicVisible && !canViewPrivateAidRequest(filter, request) {
			continue
		}
		if filter.Category != "" && request.Category != filter.Category {
			continue
		}
		if filter.Priority != "" && request.Priority != filter.Priority {
			continue
		}
		if filter.Status != "" && request.Status != filter.Status {
			continue
		}
		if filter.Region != "" && utils.NormalizeToken(request.Region) != filter.Region {
			continue
		}
		if filter.District != "" && utils.NormalizeToken(request.District) != filter.District {
			continue
		}
		if filter.Location != nil {
			distance := utils.DistanceMeters(*filter.Location, request.Location)
			if distance > filter.RadiusMeters {
				continue
			}
		}
		aidRequests = append(aidRequests, m.copyAidRequestWithPledgesLocked(request))
	}

	sort.Slice(aidRequests, func(i, j int) bool {
		if aidRequests[i].Status != aidRequests[j].Status {
			return aidRequestStatusRank(aidRequests[i].Status) < aidRequestStatusRank(aidRequests[j].Status)
		}
		if aidRequests[i].Priority != aidRequests[j].Priority {
			return aidRequestPriorityRank(aidRequests[i].Priority) < aidRequestPriorityRank(aidRequests[j].Priority)
		}
		if !aidRequests[i].NeededBy.Equal(aidRequests[j].NeededBy) {
			return aidRequests[i].NeededBy.Before(aidRequests[j].NeededBy)
		}
		return aidRequests[i].Title < aidRequests[j].Title
	})

	if filter.Limit > 0 && len(aidRequests) > filter.Limit {
		aidRequests = aidRequests[:filter.Limit]
	}
	return aidRequests
}

// canViewPrivateAidRequest reports whether the viewer described by the filter
// may see a private (non-public) aid request: privileged roles see all, agency
// roles only see requests owned by their own agency.
func canViewPrivateAidRequest(filter models.AidRequestFilter, request models.AidRequest) bool {
	if utils.AidRequestPrivateViewAllRoles[filter.ViewerRole] {
		return true
	}
	return request.AgencyID != "" && request.AgencyID == filter.ViewerAgencyID
}

// CreateAidRequest creates a new aid request.
func (m *MemoryStore) CreateAidRequest(request models.CreateAidRequestRequest, ctx models.AuthorityContext, now time.Time) models.AidRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	aidRequest := models.AidRequest{
		ID:                    fmt.Sprintf("aid_%03d", m.aidRequestSeq),
		Title:                 request.Title,
		Category:              request.Category,
		Priority:              request.Priority,
		Status:                "pending_review",
		Region:                request.Region,
		District:              request.District,
		Location:              request.Location,
		ReceivingOrganization: request.ReceivingOrganization,
		Contact:               request.Contact,
		QuantityNeeded:        request.QuantityNeeded,
		QuantityUnit:          request.QuantityUnit,
		Description:           request.Description,
		NeededBy:              request.NeededBy,
		Visibility:            request.Visibility,
		SourceReliefPointID:   request.SourceReliefPointID,
		AgencyID:              ctx.ActorAgencyID,
		CreatedBy:             ctx.ActorUserID,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	m.aidRequestSeq++
	m.aidRequests = append(m.aidRequests, aidRequest)
	return m.copyAidRequestWithPledgesLocked(aidRequest)
}

// ReviewAidRequest reviews and updates an aid request.
func (m *MemoryStore) ReviewAidRequest(id string, request models.ReviewAidRequestRequest, ctx models.AuthorityContext, now time.Time) (models.AidRequest, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.aidRequests {
		if m.aidRequests[index].ID != id {
			continue
		}
		next := m.aidRequests[index]
		next.Status = request.Status
		next.ApprovedBy = ctx.ActorUserID
		next.ApprovalNotes = request.ApprovalNotes
		next.AntiFraudNotes = request.AntiFraudNotes
		next.UpdatedAt = now
		m.aidRequests[index] = next
		return m.copyAidRequestWithPledgesLocked(next), "", ""
	}
	return models.AidRequest{}, "not_found", "aid request was not found"
}

// DeleteAidRequest removes an aid request and reports whether it existed.
func (m *MemoryStore) DeleteAidRequest(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.aidRequests {
		if m.aidRequests[index].ID == id {
			m.aidRequests = append(m.aidRequests[:index], m.aidRequests[index+1:]...)
			return true
		}
	}
	return false
}

// ListAidPledges returns pledges for an aid request.
func (m *MemoryStore) ListAidPledges(aidRequestID string) ([]models.AidPledge, string, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.aidRequestExistsLocked(aidRequestID) {
		return nil, "not_found", "aid request was not found"
	}
	return m.pledgesForAidRequestLocked(aidRequestID), "", ""
}

// CreateAidPledge creates a pledge against an aid request.
func (m *MemoryStore) CreateAidPledge(aidRequestID string, request models.CreateAidPledgeRequest, now time.Time) (models.AidPledge, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	requestIndex := -1
	for index := range m.aidRequests {
		if m.aidRequests[index].ID == aidRequestID {
			requestIndex = index
			break
		}
	}
	if requestIndex == -1 {
		return models.AidPledge{}, "not_found", "aid request was not found"
	}
	if !utils.PublicAidRequestStatuses[m.aidRequests[requestIndex].Status] {
		return models.AidPledge{}, "aid_request_not_open", "pledges are only accepted for approved or open aid requests"
	}

	pledge := models.AidPledge{
		ID:           fmt.Sprintf("pledge_%03d", len(m.aidPledges)+1),
		AidRequestID: aidRequestID,
		DonorName:    request.DonorName,
		DonorType:    request.DonorType,
		Contact:      request.Contact,
		Quantity:     request.Quantity,
		Unit:         request.Unit,
		Note:         request.Note,
		Status:       "pledged",
		ReviewStatus: "pending_review",
		PledgedAt:    now,
		UpdatedAt:    now,
	}
	m.aidPledges = append(m.aidPledges, pledge)
	m.refreshAidRequestPledgeSummaryLocked(requestIndex, now)
	return pledge, "", ""
}

// ReviewAidPledge reviews and updates an aid pledge.
func (m *MemoryStore) ReviewAidPledge(aidRequestID, pledgeID string, request models.ReviewAidPledgeRequest, ctx models.AuthorityContext, now time.Time) (models.AidPledge, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	pledgeID = strings.TrimSpace(pledgeID)
	requestIndex := -1
	for index := range m.aidRequests {
		if m.aidRequests[index].ID == aidRequestID {
			requestIndex = index
			break
		}
	}
	if requestIndex == -1 {
		return models.AidPledge{}, "not_found", "aid request was not found"
	}
	for index := range m.aidPledges {
		if m.aidPledges[index].ID != pledgeID || m.aidPledges[index].AidRequestID != aidRequestID {
			continue
		}
		if request.Status != "" {
			m.aidPledges[index].Status = request.Status
		}
		if request.ReviewStatus != "" {
			m.aidPledges[index].ReviewStatus = request.ReviewStatus
		}
		if request.FraudReviewNotes != "" {
			m.aidPledges[index].FraudReviewNotes = request.FraudReviewNotes
		}
		m.aidPledges[index].ReviewedBy = ctx.ActorUserID
		m.aidPledges[index].UpdatedAt = now
		m.refreshAidRequestPledgeSummaryLocked(requestIndex, now)
		return m.aidPledges[index], "", ""
	}
	return models.AidPledge{}, "not_found", "aid pledge was not found"
}

func (m *MemoryStore) aidRequestExistsLocked(aidRequestID string) bool {
	aidRequestID = strings.TrimSpace(aidRequestID)
	for _, request := range m.aidRequests {
		if request.ID == aidRequestID {
			return true
		}
	}
	return false
}

func (m *MemoryStore) copyAidRequestWithPledgesLocked(request models.AidRequest) models.AidRequest {
	pledges := m.pledgesForAidRequestLocked(request.ID)
	request.Pledges = pledges
	request.QuantityPledged = totalActivePledgedQuantity(pledges)
	return request
}

func (m *MemoryStore) pledgesForAidRequestLocked(aidRequestID string) []models.AidPledge {
	pledges := make([]models.AidPledge, 0)
	for _, pledge := range m.aidPledges {
		if pledge.AidRequestID == aidRequestID {
			pledges = append(pledges, pledge)
		}
	}
	sort.Slice(pledges, func(i, j int) bool {
		return pledges[i].PledgedAt.After(pledges[j].PledgedAt)
	})
	return pledges
}

func (m *MemoryStore) refreshAidRequestPledgeSummaryLocked(requestIndex int, now time.Time) {
	request := m.aidRequests[requestIndex]
	pledges := m.pledgesForAidRequestLocked(request.ID)
	request.QuantityPledged = totalActivePledgedQuantity(pledges)
	// fulfilled is included so a request whose pledged total drops below the
	// quantity needed (pledge cancelled/flagged) transitions back out instead
	// of silently hiding the unmet need from donor/coordinator views.
	if utils.PublicAidRequestStatuses[request.Status] || request.Status == "fulfilled" {
		switch {
		case request.QuantityPledged >= request.QuantityNeeded:
			request.Status = "fulfilled"
		case request.QuantityPledged > 0:
			request.Status = "partially_matched"
		case request.Status == "partially_matched" || request.Status == "fulfilled":
			request.Status = "open"
		}
	}
	request.UpdatedAt = now
	m.aidRequests[requestIndex] = request
}

// ListHospitalCapacity returns hospital capacity matching the filter.
func (m *MemoryStore) ListHospitalCapacity(filter models.HospitalCapacityFilter, now time.Time) []models.HospitalCapacity {
	m.mu.RLock()
	defer m.mu.RUnlock()

	facilities := make([]models.HospitalCapacity, 0, len(m.hospitals))
	for _, facility := range m.hospitals {
		facility = withHospitalStaleness(facility, now)
		if filter.Location != nil {
			facility.DistanceMeters = int(math.Round(utils.DistanceMeters(*filter.Location, facility.Location)))
			if float64(facility.DistanceMeters) > utils.NearbySearchMeters {
				continue
			}
		}
		if filter.Service != "" && !utils.ContainsNormalized(facility.Services, filter.Service) {
			continue
		}
		if filter.EmergencyCapacity != "" && facility.EmergencyCapacity != filter.EmergencyCapacity {
			continue
		}
		if filter.MinAvailableBeds > 0 && facility.AvailableBeds < filter.MinAvailableBeds {
			continue
		}
		if !filter.IncludeStale && facility.Stale {
			continue
		}
		facilities = append(facilities, facility)
	}
	sort.Slice(facilities, func(i, j int) bool {
		if facilities[i].Stale != facilities[j].Stale {
			return !facilities[i].Stale
		}
		if filter.Location != nil && facilities[i].DistanceMeters != facilities[j].DistanceMeters {
			return facilities[i].DistanceMeters < facilities[j].DistanceMeters
		}
		if facilities[i].EmergencyCapacity != facilities[j].EmergencyCapacity {
			return hospitalCapacityRank(facilities[i].EmergencyCapacity) < hospitalCapacityRank(facilities[j].EmergencyCapacity)
		}
		if facilities[i].AvailableBeds != facilities[j].AvailableBeds {
			return facilities[i].AvailableBeds > facilities[j].AvailableBeds
		}
		return facilities[i].Name < facilities[j].Name
	})
	if filter.Limit > 0 && len(facilities) > filter.Limit {
		facilities = facilities[:filter.Limit]
	}
	return copyHospitals(facilities)
}

// UpdateHospitalCapacity updates a hospital's capacity record.
//
//nolint:gocognit // legacy complex function; refactor into validation/execution helpers in a future pass.
func (m *MemoryStore) UpdateHospitalCapacity(id string, request models.HospitalCapacityUpdateRequest, ctx models.AuthorityContext, now time.Time) (models.HospitalCapacity, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id = strings.TrimSpace(id)
	for index := range m.hospitals {
		if m.hospitals[index].ID != id {
			continue
		}

		next := m.hospitals[index]
		if request.TotalBeds != nil {
			next.TotalBeds = *request.TotalBeds
		}
		if request.AvailableBeds != nil {
			next.AvailableBeds = *request.AvailableBeds
		}
		if next.AvailableBeds > next.TotalBeds {
			return models.HospitalCapacity{}, "invalid_available_beds", "availableBeds cannot exceed totalBeds"
		}
		if request.ICUBedsAvailable != nil {
			next.ICUBedsAvailable = *request.ICUBedsAvailable
		}
		if request.MaternityBedsAvailable != nil {
			next.MaternityBedsAvailable = *request.MaternityBedsAvailable
		}
		if request.PediatricBedsAvailable != nil {
			next.PediatricBedsAvailable = *request.PediatricBedsAvailable
		}
		if request.IsolationBedsAvailable != nil {
			next.IsolationBedsAvailable = *request.IsolationBedsAvailable
		}
		if request.EmergencyCapacity != "" {
			next.EmergencyCapacity = request.EmergencyCapacity
		} else {
			next.EmergencyCapacity = hospitalCapacityFromBeds(next.TotalBeds, next.AvailableBeds, next.EmergencyCapacity)
		}
		if request.EmergencyUnitStatus != "" {
			next.EmergencyUnitStatus = request.EmergencyUnitStatus
		}
		if request.AmbulancesAvailable != nil {
			next.AmbulancesAvailable = *request.AmbulancesAvailable
		}
		if request.OxygenAvailable != nil {
			next.OxygenAvailable = *request.OxygenAvailable
		}
		if request.Notes != "" {
			next.Notes = request.Notes
		}
		// Provenance is only overwritten when the caller supplies it; a
		// partial PATCH must not clobber the stored Source/SourceRef.
		if request.Source != "" {
			next.Source = request.Source
		}
		if request.SourceRef != "" {
			next.SourceRef = request.SourceRef
		}
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		next = withHospitalStaleness(next, now)
		m.hospitals[index] = next
		return copyHospital(next), "", ""
	}

	return models.HospitalCapacity{}, "not_found", "hospital facility was not found"
}

// ImportHospitalCapacityFixture imports fixture capacity updates.
func (m *MemoryStore) ImportHospitalCapacityFixture(request models.HospitalCapacityImportRequest, ctx models.AuthorityContext, now time.Time) ([]models.HospitalCapacity, int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	updates := request.Records
	if len(updates) == 0 {
		updates = utils.DefaultHospitalCapacityFixture()
	}
	byID := map[string]models.HospitalCapacityFixtureRecord{}
	for _, record := range updates {
		byID[record.FacilityID] = record
	}

	imported := 0
	facilities := make([]models.HospitalCapacity, 0, len(m.hospitals))
	for index := range m.hospitals {
		update, ok := byID[m.hospitals[index].ID]
		if !ok {
			continue
		}
		next := m.hospitals[index]
		// Skip feed records that would break the availableBeds <= totalBeds
		// invariant (the manual update path rejects the same condition).
		if update.AvailableBeds > next.TotalBeds {
			continue
		}
		next.AvailableBeds = update.AvailableBeds
		next.ICUBedsAvailable = update.ICUBedsAvailable
		next.MaternityBedsAvailable = update.MaternityBedsAvailable
		next.PediatricBedsAvailable = update.PediatricBedsAvailable
		next.IsolationBedsAvailable = update.IsolationBedsAvailable
		next.EmergencyCapacity = update.EmergencyCapacity
		if update.EmergencyUnitStatus != "" {
			next.EmergencyUnitStatus = update.EmergencyUnitStatus
		}
		next.AmbulancesAvailable = update.AmbulancesAvailable
		if update.OxygenAvailable != nil {
			next.OxygenAvailable = *update.OxygenAvailable
		}
		if update.Notes != "" {
			next.Notes = update.Notes
		}
		next.Source = request.Source
		next.SourceRef = request.SourceRef
		next.UpdatedBy = ctx.ActorUserID
		next.UpdatedAt = now
		next = withHospitalStaleness(next, now)
		m.hospitals[index] = next
		facilities = append(facilities, copyHospital(next))
		imported++
	}
	return facilities, imported
}

func copyShelters(source []models.Shelter) []models.Shelter {
	shelters := make([]models.Shelter, 0, len(source))
	for _, shelter := range source {
		shelter.Facilities = append([]string{}, shelter.Facilities...)
		shelters = append(shelters, shelter)
	}
	return shelters
}

func copyRecovery(source []models.RecoverySupportLocation) []models.RecoverySupportLocation {
	locations := make([]models.RecoverySupportLocation, 0, len(source))
	for _, item := range source {
		item.Services = append([]string{}, item.Services...)
		locations = append(locations, item)
	}
	return locations
}

func copyHospitals(source []models.HospitalCapacity) []models.HospitalCapacity {
	facilities := make([]models.HospitalCapacity, 0, len(source))
	for _, facility := range source {
		facilities = append(facilities, copyHospital(facility))
	}
	return facilities
}

func copyHospital(facility models.HospitalCapacity) models.HospitalCapacity {
	facility.Services = append([]string{}, facility.Services...)
	return facility
}

func shelterStatusRank(status string) int {
	switch status {
	case "open":
		return 0
	case "unknown":
		return 1
	case "full":
		return 2
	case "closed":
		return 3
	default:
		return 4
	}
}

func hospitalCapacityRank(status string) int {
	switch status {
	case "available":
		return 0
	case "limited":
		return 1
	case "unknown":
		return 2
	case "full":
		return 3
	case "offline":
		return 4
	default:
		return 5
	}
}

func statusForOccupancy(capacity int, occupancy int, fallback string) string {
	// An explicit "closed" status survives occupancy-only updates; reopening
	// always requires an explicit status change.
	if fallback == "closed" {
		return "closed"
	}
	if capacity > 0 && occupancy >= capacity {
		return "full"
	}
	if fallback == "full" && occupancy < capacity {
		return "open"
	}
	if fallback == "" {
		return "open"
	}
	return fallback
}

func hospitalCapacityFromBeds(totalBeds int, availableBeds int, fallback string) string {
	if totalBeds <= 0 {
		if fallback == "" {
			return "unknown"
		}
		return fallback
	}
	if availableBeds <= 0 {
		return "full"
	}
	if float64(availableBeds)/float64(totalBeds) <= 0.1 {
		return "limited"
	}
	return "available"
}

func withHospitalStaleness(facility models.HospitalCapacity, now time.Time) models.HospitalCapacity {
	facility.Stale = false
	facility.StaleReason = ""
	if facility.UpdatedAt.IsZero() {
		facility.Stale = true
		facility.StaleReason = "capacity timestamp missing"
		return facility
	}
	if now.Sub(facility.UpdatedAt) > utils.HospitalCapacityStaleAfter {
		facility.Stale = true
		facility.StaleReason = "capacity update older than 30 minutes"
	}
	return facility
}

func copyReliefPoints(source []models.ReliefPoint) []models.ReliefPoint {
	reliefPoints := make([]models.ReliefPoint, 0, len(source))
	for _, point := range source {
		point.StockCategories = copyStockCategories(point.StockCategories)
		reliefPoints = append(reliefPoints, point)
	}
	return reliefPoints
}

func copyStockCategories(source []models.ReliefStockCategory) []models.ReliefStockCategory {
	categories := make([]models.ReliefStockCategory, len(source))
	copy(categories, source)
	return categories
}

func stockCategoriesEqual(a, b []models.ReliefStockCategory) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Category != b[i].Category || a[i].Quantity != b[i].Quantity || a[i].Unit != b[i].Unit {
			return false
		}
	}
	return true
}

func pointInBBox(location models.Coordinates, box models.BoundingBox) bool {
	return location.Lat >= box.MinLat && location.Lat <= box.MaxLat &&
		location.Lng >= box.MinLng && location.Lng <= box.MaxLng
}

func totalActivePledgedQuantity(pledges []models.AidPledge) int {
	total := 0
	for _, pledge := range pledges {
		if pledge.ReviewStatus == "flagged" || pledge.Status == "flagged" || pledge.Status == "cancelled" {
			continue
		}
		total += pledge.Quantity
	}
	return total
}

func aidRequestStatusRank(status string) int {
	switch status {
	case "pending_review":
		return 0
	case "approved", "open":
		return 1
	case "partially_matched":
		return 2
	case "fulfilled":
		return 3
	case "paused":
		return 4
	case "rejected", "closed":
		return 5
	default:
		return 6
	}
}

func aidRequestPriorityRank(priority string) int {
	switch priority {
	case "urgent":
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

func reliefPointStatusRank(status string) int {
	switch status {
	case "open":
		return 0
	case "limited":
		return 1
	case "paused":
		return 2
	case "closed":
		return 3
	default:
		return 4
	}
}
