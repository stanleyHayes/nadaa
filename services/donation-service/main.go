package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type server struct {
	store *memoryStore
	now   func() time.Time
}

type memoryStore struct {
	mu        sync.RWMutex
	seq       int
	donors    []donorRecord
	catalog   []aidCatalogRecord
	requests  []aidRequestRecord
	pledges   []pledgeRecord
}

type donorRecord struct {
	ID                string   `json:"id"`
	Reference         string   `json:"reference"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ContactName       string   `json:"contactName"`
	ContactEmail      string   `json:"contactEmail"`
	ContactPhone      string   `json:"contactPhone"`
	Region            string   `json:"region"`
	District          string   `json:"district"`
	ItemsOffered      []string `json:"itemsOffered"`
	MonetaryPledgeGhs float64  `json:"monetaryPledgeGhs"`
	Status            string   `json:"status"`
	Notes             string   `json:"notes,omitempty"`
	CreatedBy         string   `json:"createdBy,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type aidCatalogRecord struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	DefaultUnit   string  `json:"defaultUnit"`
	PriorityScore float64 `json:"priorityScore"`
}

type aidRequestRecord struct {
	ID                string    `json:"id"`
	Reference         string    `json:"reference"`
	Title             string    `json:"title"`
	Description       string    `json:"description,omitempty"`
	Category          string    `json:"category"`
	ItemCode          string    `json:"itemCode"`
	QuantityNeeded    int       `json:"quantityNeeded"`
	QuantityFulfilled int       `json:"quantityFulfilled"`
	Unit              string    `json:"unit"`
	Priority          string    `json:"priority"`
	LocationLabel     string    `json:"locationLabel,omitempty"`
	Region            string    `json:"region"`
	District          string    `json:"district"`
	BeneficiaryCount  int       `json:"beneficiaryCount"`
	Status            string    `json:"status"`
	RequestedBy       string    `json:"requestedBy,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type pledgeRecord struct {
	ID                string    `json:"id"`
	Reference         string    `json:"reference"`
	AidRequestID      string    `json:"aidRequestId"`
	DonorID           string    `json:"donorId"`
	DonorName         string    `json:"donorName"`
	QuantityPledged   int       `json:"quantityPledged"`
	QuantityDelivered int       `json:"quantityDelivered"`
	Status            string    `json:"status"`
	DeliveryNote      string    `json:"deliveryNote,omitempty"`
	ContactEmail      string    `json:"contactEmail,omitempty"`
	ContactPhone      string    `json:"contactPhone,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type authorityContext struct {
	ActorUserID   string
	ActorAgencyID string
	ActorRole     string
	MFACompleted  bool
	RequestID     string
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type donorListResponse struct {
	Donors      []donorRecord `json:"donors"`
	GeneratedAt time.Time     `json:"generatedAt"`
}

type aidCatalogResponse struct {
	Items       []aidCatalogRecord `json:"items"`
	GeneratedAt time.Time          `json:"generatedAt"`
}

type aidRequestListResponse struct {
	Requests    []aidRequestRecord `json:"requests"`
	GeneratedAt time.Time          `json:"generatedAt"`
}

type pledgeListResponse struct {
	Pledges     []pledgeRecord `json:"pledges"`
	GeneratedAt time.Time      `json:"generatedAt"`
}

type createDonorRequest struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	ContactName       string   `json:"contactName,omitempty"`
	ContactEmail      string   `json:"contactEmail,omitempty"`
	ContactPhone      string   `json:"contactPhone,omitempty"`
	Region            string   `json:"region,omitempty"`
	District          string   `json:"district,omitempty"`
	ItemsOffered      []string `json:"itemsOffered,omitempty"`
	MonetaryPledgeGhs float64  `json:"monetaryPledgeGhs,omitempty"`
	Notes             string   `json:"notes,omitempty"`
}

type updateDonorRequest struct {
	Status string `json:"status,omitempty"`
	Notes  string `json:"notes,omitempty"`
}

type createAidRequestRequest struct {
	Title            string `json:"title"`
	Description      string `json:"description,omitempty"`
	Category         string `json:"category"`
	ItemCode         string `json:"itemCode"`
	QuantityNeeded   int    `json:"quantityNeeded"`
	Unit             string `json:"unit"`
	Priority         string `json:"priority"`
	LocationLabel    string `json:"locationLabel,omitempty"`
	Region           string `json:"region"`
	District         string `json:"district"`
	BeneficiaryCount int    `json:"beneficiaryCount,omitempty"`
}

type updateAidRequestRequest struct {
	Status         string `json:"status,omitempty"`
	QuantityNeeded int    `json:"quantityNeeded,omitempty"`
}

type createPledgeRequest struct {
	DonorID         string `json:"donorId"`
	DonorName       string `json:"donorName,omitempty"`
	QuantityPledged int    `json:"quantityPledged"`
	ContactEmail    string `json:"contactEmail,omitempty"`
	ContactPhone    string `json:"contactPhone,omitempty"`
	DeliveryNote    string `json:"deliveryNote,omitempty"`
}

type updatePledgeRequest struct {
	Status            string `json:"status,omitempty"`
	QuantityDelivered int    `json:"quantityDelivered,omitempty"`
	DeliveryNote      string `json:"deliveryNote,omitempty"`
}

type allocateRequest struct {
	PledgeID string `json:"pledgeId"`
	Quantity int    `json:"quantity"`
}

var authorityRoles = map[string]bool{
	"system_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
	"ngo":              true,
	"agency_admin":     true,
	"agency_viewer":    true,
}

var allowedDonorTypes = map[string]bool{
	"individual":   true,
	"organization": true,
	"ngo":          true,
	"government":   true,
	"other":        true,
}

var allowedDonorStatuses = map[string]bool{
	"active":   true,
	"inactive": true,
}

var allowedRequestStatuses = map[string]bool{
	"open":               true,
	"partially_fulfilled": true,
	"fulfilled":          true,
	"closed":             true,
}

var allowedPriorities = map[string]bool{
	"low":      true,
	"medium":   true,
	"high":     true,
	"critical": true,
}

var allowedPledgeStatuses = map[string]bool{
	"pledged":   true,
	"delivered": true,
	"cancelled": true,
}

var allowedCatalogCategories = map[string]bool{
	"food":       true,
	"water":      true,
	"medical":    true,
	"shelter":    true,
	"sanitation": true,
}

const defaultDonationAddr = ":8100"

func main() {
	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("GET /api/v1/donors", srv.listDonorsHandler)
	mux.HandleFunc("POST /api/v1/donors", srv.createDonorHandler)
	mux.HandleFunc("GET /api/v1/donors/{id}", srv.getDonorHandler)
	mux.HandleFunc("PATCH /api/v1/donors/{id}", srv.updateDonorHandler)
	mux.HandleFunc("GET /api/v1/aid-catalog", srv.listCatalogHandler)
	mux.HandleFunc("GET /api/v1/aid-requests", srv.listAidRequestsHandler)
	mux.HandleFunc("POST /api/v1/aid-requests", srv.createAidRequestHandler)
	mux.HandleFunc("GET /api/v1/aid-requests/{id}", srv.getAidRequestHandler)
	mux.HandleFunc("PATCH /api/v1/aid-requests/{id}", srv.updateAidRequestHandler)
	mux.HandleFunc("GET /api/v1/aid-requests/{id}/pledges", srv.listRequestPledgesHandler)
	mux.HandleFunc("POST /api/v1/aid-requests/{id}/pledges", srv.createPledgeHandler)
	mux.HandleFunc("GET /api/v1/pledges", srv.listPledgesHandler)
	mux.HandleFunc("PATCH /api/v1/pledges/{id}", srv.updatePledgeHandler)
	mux.HandleFunc("POST /api/v1/aid-requests/{id}/allocate", srv.allocatePledgeHandler)

	addr := envOrDefault("PORT", defaultDonationAddr)
	log.Printf("INFO donation-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServer() *server {
	now := time.Now
	return &server{store: newMemoryStore(now().UTC()), now: now}
}

func newMemoryStore(now time.Time) *memoryStore {
	store := &memoryStore{}
	store.catalog = seedCatalog(now)
	store.requests = seedAidRequests(now)
	store.seq = len(store.catalog)
	return store
}

func seedCatalog(now time.Time) []aidCatalogRecord {
	return []aidCatalogRecord{
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

type seqHolder struct {
	value int
}

func nextIDFor(holder *seqHolder, prefix string) string {
	id := fmt.Sprintf("%s_%03d", prefix, holder.value)
	holder.value++
	return id
}

func seedAidRequests(now time.Time) []aidRequestRecord {
	return []aidRequestRecord{
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

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "donation-service"})
}

func (s *server) listDonorsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	filter := donorFilter{
		Type: normalizeToken(r.URL.Query().Get("type")),
		Query: strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q"))),
	}
	if filter.Type != "" && !allowedDonorTypes[filter.Type] {
		log.Printf("WARN donation-service donor_list invalid_type actor=%s type=%s", ctx.ActorUserID, filter.Type)
		writeError(w, http.StatusBadRequest, "invalid_type", "type must be individual, organization, ngo, government, or other")
		return
	}

	donors := s.store.listDonors(filter)
	log.Printf("INFO donation-service donor_list count=%d actor=%s type=%s q=%t", len(donors), ctx.ActorUserID, filter.Type, filter.Query != "")
	writeJSON(w, http.StatusOK, donorListResponse{Donors: donors, GeneratedAt: s.now().UTC()})
}

func (s *server) createDonorHandler(w http.ResponseWriter, r *http.Request) {
	var request createDonorRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service donor_create invalid_json error=%v", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreateDonor(request)
	if code != "" {
		log.Printf("WARN donation-service donor_create validation_failed code=%s", code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	createdBy := "public"
	if ctx, ok := authorityContextFromRequest(r); ok {
		createdBy = ctx.ActorUserID
	}

	donor := s.store.createDonor(normalized, createdBy, s.now().UTC())
	log.Printf("INFO donation-service donor_create completed id=%s reference=%s createdBy=%s", donor.ID, donor.Reference, donor.CreatedBy)
	writeJSON(w, http.StatusCreated, donor)
}

func (s *server) getDonorHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAuthority(w, r); !ok {
		return
	}

	donor, ok := s.store.getDonor(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "donor was not found")
		return
	}
	writeJSON(w, http.StatusOK, donor)
}

func (s *server) updateDonorHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request updateDonorRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service donor_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateDonor(request)
	if code != "" {
		log.Printf("WARN donation-service donor_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	donor, code, message := s.store.updateDonor(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service donor_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service donor_update completed id=%s actor=%s status=%s", donor.ID, ctx.ActorUserID, donor.Status)
	writeJSON(w, http.StatusOK, donor)
}

func (s *server) listCatalogHandler(w http.ResponseWriter, _ *http.Request) {
	items := s.store.listCatalog()
	log.Printf("INFO donation-service aid_catalog_list count=%d", len(items))
	writeJSON(w, http.StatusOK, aidCatalogResponse{Items: items, GeneratedAt: s.now().UTC()})
}

func (s *server) listAidRequestsHandler(w http.ResponseWriter, r *http.Request) {
	filter := aidRequestFilter{
		Status:   normalizeToken(r.URL.Query().Get("status")),
		Category: normalizeToken(r.URL.Query().Get("category")),
		Region:   strings.TrimSpace(strings.ToLower(r.URL.Query().Get("region"))),
		Priority: normalizeToken(r.URL.Query().Get("priority")),
	}
	if filter.Status != "" && !allowedRequestStatuses[filter.Status] {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be open, partially_fulfilled, fulfilled, or closed")
		return
	}
	if filter.Priority != "" && !allowedPriorities[filter.Priority] {
		writeError(w, http.StatusBadRequest, "invalid_priority", "priority must be low, medium, high, or critical")
		return
	}

	requests := s.store.listAidRequests(filter)
	log.Printf("INFO donation-service aid_request_list count=%d status=%s category=%s region=%s priority=%s", len(requests), filter.Status, filter.Category, filter.Region, filter.Priority)
	writeJSON(w, http.StatusOK, aidRequestListResponse{Requests: requests, GeneratedAt: s.now().UTC()})
}

func (s *server) createAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request createAidRequestRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service aid_request_create invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreateAidRequest(request)
	if code != "" {
		log.Printf("WARN donation-service aid_request_create validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	req := s.store.createAidRequest(normalized, ctx.ActorUserID, s.now().UTC())
	log.Printf("INFO donation-service aid_request_create completed id=%s reference=%s actor=%s", req.ID, req.Reference, ctx.ActorUserID)
	writeJSON(w, http.StatusCreated, req)
}

func (s *server) getAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	req, ok := s.store.getAidRequest(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "aid request was not found")
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (s *server) updateAidRequestHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request updateAidRequestRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service aid_request_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdateAidRequest(request)
	if code != "" {
		log.Printf("WARN donation-service aid_request_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	req, code, message := s.store.updateAidRequest(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service aid_request_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service aid_request_update completed id=%s actor=%s status=%s fulfilled=%d/%d", req.ID, ctx.ActorUserID, req.Status, req.QuantityFulfilled, req.QuantityNeeded)
	writeJSON(w, http.StatusOK, req)
}

func (s *server) listRequestPledgesHandler(w http.ResponseWriter, r *http.Request) {
	pledges := s.store.listPledgesForRequest(r.PathValue("id"))
	log.Printf("INFO donation-service pledge_list_for_request aidRequestId=%s count=%d", r.PathValue("id"), len(pledges))
	writeJSON(w, http.StatusOK, pledgeListResponse{Pledges: pledges, GeneratedAt: s.now().UTC()})
}

func (s *server) createPledgeHandler(w http.ResponseWriter, r *http.Request) {
	var request createPledgeRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service pledge_create invalid_json error=%v", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreatePledge(request)
	if code != "" {
		log.Printf("WARN donation-service pledge_create validation_failed code=%s", code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.createPledge(r.PathValue("id"), normalized, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service pledge_create failed aidRequestId=%s code=%s", r.PathValue("id"), code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service pledge_create completed id=%s reference=%s aidRequestId=%s donorId=%s quantity=%d", pledge.ID, pledge.Reference, pledge.AidRequestID, pledge.DonorID, pledge.QuantityPledged)
	writeJSON(w, http.StatusCreated, pledge)
}

func (s *server) listPledgesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	filter := pledgeFilter{Status: normalizeToken(r.URL.Query().Get("status"))}
	if filter.Status != "" && !allowedPledgeStatuses[filter.Status] {
		log.Printf("WARN donation-service pledge_list invalid_status actor=%s status=%s", ctx.ActorUserID, filter.Status)
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be pledged, delivered, or cancelled")
		return
	}

	pledges := s.store.listPledges(filter)
	log.Printf("INFO donation-service pledge_list count=%d actor=%s status=%s", len(pledges), ctx.ActorUserID, filter.Status)
	writeJSON(w, http.StatusOK, pledgeListResponse{Pledges: pledges, GeneratedAt: s.now().UTC()})
}

func (s *server) updatePledgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request updatePledgeRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service pledge_update invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdatePledge(request)
	if code != "" {
		log.Printf("WARN donation-service pledge_update validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.updatePledge(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service pledge_update failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service pledge_update completed id=%s actor=%s status=%s delivered=%d", pledge.ID, ctx.ActorUserID, pledge.Status, pledge.QuantityDelivered)
	writeJSON(w, http.StatusOK, pledge)
}

func (s *server) allocatePledgeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request allocateRequest
	if err := decodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service pledge_allocate invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeAllocate(request)
	if code != "" {
		log.Printf("WARN donation-service pledge_allocate validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		writeError(w, http.StatusBadRequest, code, message)
		return
	}

	pledge, code, message := s.store.allocatePledge(r.PathValue("id"), normalized, ctx.ActorUserID, s.now().UTC())
	if code != "" {
		log.Printf("WARN donation-service pledge_allocate failed aidRequestId=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		writeError(w, statusForCode(code), code, message)
		return
	}
	log.Printf("INFO donation-service pledge_allocate completed id=%s actor=%s quantity=%d", pledge.ID, ctx.ActorUserID, pledge.QuantityDelivered)
	writeJSON(w, http.StatusOK, pledge)
}

type donorFilter struct {
	Type  string
	Query string
}

type aidRequestFilter struct {
	Status   string
	Category string
	Region   string
	Priority string
}

type pledgeFilter struct {
	Status string
}

func (m *memoryStore) nextSeq() int {
	m.seq++
	return m.seq
}

func (m *memoryStore) generateRef(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%s-%03d", prefix, now.Format("20060102"), m.nextSeq())
}

func (m *memoryStore) listDonors(filter donorFilter) []donorRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	donors := make([]donorRecord, 0)
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

func donorMatchesQuery(donor donorRecord, query string) bool {
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

func (m *memoryStore) createDonor(request createDonorRequest, createdBy string, now time.Time) donorRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	donor := donorRecord{
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

func (m *memoryStore) getDonor(id string) (donorRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, donor := range m.donors {
		if donor.ID == id {
			return donor, true
		}
	}
	return donorRecord{}, false
}

func (m *memoryStore) updateDonor(id string, request updateDonorRequest, updatedBy string, now time.Time) (donorRecord, string, string) {
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

	return donorRecord{}, "not_found", "donor was not found"
}

func (m *memoryStore) listCatalog() []aidCatalogRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]aidCatalogRecord, len(m.catalog))
	copy(items, m.catalog)
	sort.Slice(items, func(i, j int) bool {
		if items[i].PriorityScore != items[j].PriorityScore {
			return items[i].PriorityScore > items[j].PriorityScore
		}
		return items[i].Name < items[j].Name
	})
	return items
}

func (m *memoryStore) listAidRequests(filter aidRequestFilter) []aidRequestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requests := make([]aidRequestRecord, 0)
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

func (m *memoryStore) createAidRequest(request createAidRequestRequest, requestedBy string, now time.Time) aidRequestRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	req := aidRequestRecord{
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

func (m *memoryStore) getAidRequest(id string) (aidRequestRecord, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	id = strings.TrimSpace(id)
	for _, req := range m.requests {
		if req.ID == id {
			return req, true
		}
	}
	return aidRequestRecord{}, false
}

func (m *memoryStore) updateAidRequest(id string, request updateAidRequestRequest, updatedBy string, now time.Time) (aidRequestRecord, string, string) {
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

	return aidRequestRecord{}, "not_found", "aid request was not found"
}

func (m *memoryStore) createPledge(aidRequestID string, request createPledgeRequest, now time.Time) (pledgeRecord, string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	request.DonorID = strings.TrimSpace(request.DonorID)

	var donor donorRecord
	found := false
	for _, d := range m.donors {
		if d.ID == request.DonorID {
			donor = d
			found = true
			break
		}
	}
	if !found {
		return pledgeRecord{}, "donor_not_found", "donor was not found"
	}

	requestIndex := -1
	for index := range m.requests {
		if m.requests[index].ID == aidRequestID {
			requestIndex = index
			break
		}
	}
	if requestIndex == -1 {
		return pledgeRecord{}, "not_found", "aid request was not found"
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

	pledge := pledgeRecord{
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

func (m *memoryStore) listPledgesForRequest(aidRequestID string) []pledgeRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	aidRequestID = strings.TrimSpace(aidRequestID)
	pledges := make([]pledgeRecord, 0)
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

func (m *memoryStore) listPledges(filter pledgeFilter) []pledgeRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pledges := make([]pledgeRecord, 0)
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

func (m *memoryStore) updatePledge(id string, request updatePledgeRequest, updatedBy string, now time.Time) (pledgeRecord, string, string) {
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
		return pledgeRecord{}, "not_found", "pledge was not found"
	}

	next := m.pledges[pledgeIndex]
	if request.Status != "" {
		next.Status = request.Status
	}
	if request.QuantityDelivered > 0 {
		if request.QuantityDelivered > next.QuantityPledged {
			return pledgeRecord{}, "invalid_quantity", "quantityDelivered cannot exceed quantityPledged"
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

func (m *memoryStore) allocatePledge(aidRequestID string, request allocateRequest, updatedBy string, now time.Time) (pledgeRecord, string, string) {
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
		return pledgeRecord{}, "not_found", "pledge was not found for this aid request"
	}

	if request.Quantity > m.pledges[pledgeIndex].QuantityPledged {
		return pledgeRecord{}, "invalid_quantity", "quantity cannot exceed quantityPledged"
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

func (m *memoryStore) recalcSingleRequest(req aidRequestRecord, now time.Time) aidRequestRecord {
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
		if fulfilled >= req.QuantityNeeded {
			req.Status = "fulfilled"
		} else if fulfilled > 0 {
			req.Status = "partially_fulfilled"
		} else {
			req.Status = "open"
		}
	}
	req.UpdatedAt = now
	return req
}

func normalizeCreateDonor(request createDonorRequest) (createDonorRequest, string, string) {
	request.Name = strings.TrimSpace(request.Name)
	request.Type = normalizeToken(request.Type)
	request.ContactName = strings.TrimSpace(request.ContactName)
	request.ContactEmail = strings.TrimSpace(strings.ToLower(request.ContactEmail))
	request.ContactPhone = strings.TrimSpace(request.ContactPhone)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)
	request.Notes = strings.TrimSpace(request.Notes)

	if request.Name == "" || len(request.Name) > 200 || unsafeText(request.Name) {
		return request, "invalid_name", "name is required and must be 200 safe characters or fewer"
	}
	if !allowedDonorTypes[request.Type] {
		return request, "invalid_type", "type must be individual, organization, ngo, government, or other"
	}
	if len(request.ContactName) > 200 || unsafeText(request.ContactName) {
		return request, "invalid_contact_name", "contactName must be 200 safe characters or fewer"
	}
	if request.ContactEmail != "" && (len(request.ContactEmail) > 200 || !strings.Contains(request.ContactEmail, "@") || unsafeText(request.ContactEmail)) {
		return request, "invalid_contact_email", "contactEmail must be a valid email address"
	}
	if len(request.ContactPhone) > 50 || unsafeText(request.ContactPhone) {
		return request, "invalid_contact_phone", "contactPhone must be 50 safe characters or fewer"
	}
	if len(request.Region) > 100 || unsafeText(request.Region) {
		return request, "invalid_region", "region must be 100 safe characters or fewer"
	}
	if len(request.District) > 100 || unsafeText(request.District) {
		return request, "invalid_district", "district must be 100 safe characters or fewer"
	}
	if request.MonetaryPledgeGhs < 0 {
		return request, "invalid_monetary_pledge", "monetaryPledgeGhs must be zero or greater"
	}
	if len(request.Notes) > 700 || unsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 700 safe characters or fewer"
	}

	cleaned := make([]string, 0, len(request.ItemsOffered))
	for _, item := range request.ItemsOffered {
		item = strings.TrimSpace(item)
		if item != "" && !unsafeText(item) {
			cleaned = append(cleaned, item)
		}
	}
	request.ItemsOffered = cleaned
	return request, "", ""
}

func normalizeUpdateDonor(request updateDonorRequest) (updateDonorRequest, string, string) {
	request.Status = normalizeToken(request.Status)
	request.Notes = strings.TrimSpace(request.Notes)

	if request.Status == "" && request.Notes == "" {
		return request, "no_changes", "at least one of status or notes must be supplied"
	}
	if request.Status != "" && !allowedDonorStatuses[request.Status] {
		return request, "invalid_status", "status must be active or inactive"
	}
	if len(request.Notes) > 700 || unsafeText(request.Notes) {
		return request, "invalid_notes", "notes must be 700 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeCreateAidRequest(request createAidRequestRequest) (createAidRequestRequest, string, string) {
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	request.Category = normalizeToken(request.Category)
	request.ItemCode = strings.TrimSpace(request.ItemCode)
	request.Unit = strings.TrimSpace(request.Unit)
	request.Priority = normalizeToken(request.Priority)
	request.LocationLabel = strings.TrimSpace(request.LocationLabel)
	request.Region = strings.TrimSpace(request.Region)
	request.District = strings.TrimSpace(request.District)

	if request.Title == "" || len(request.Title) > 200 || unsafeText(request.Title) {
		return request, "invalid_title", "title is required and must be 200 safe characters or fewer"
	}
	if len(request.Description) > 1000 || unsafeText(request.Description) {
		return request, "invalid_description", "description must be 1000 safe characters or fewer"
	}
	if request.Category == "" || !allowedCatalogCategories[request.Category] {
		return request, "invalid_category", "category must be food, water, medical, shelter, or sanitation"
	}
	if request.ItemCode == "" || len(request.ItemCode) > 100 || unsafeText(request.ItemCode) {
		return request, "invalid_item_code", "itemCode is required and must be 100 safe characters or fewer"
	}
	if request.QuantityNeeded <= 0 {
		return request, "invalid_quantity_needed", "quantityNeeded must be greater than zero"
	}
	if request.Unit == "" || len(request.Unit) > 50 || unsafeText(request.Unit) {
		return request, "invalid_unit", "unit is required and must be 50 safe characters or fewer"
	}
	if !allowedPriorities[request.Priority] {
		return request, "invalid_priority", "priority must be low, medium, high, or critical"
	}
	if len(request.LocationLabel) > 200 || unsafeText(request.LocationLabel) {
		return request, "invalid_location_label", "locationLabel must be 200 safe characters or fewer"
	}
	if request.Region == "" || len(request.Region) > 100 || unsafeText(request.Region) {
		return request, "invalid_region", "region is required and must be 100 safe characters or fewer"
	}
	if request.District == "" || len(request.District) > 100 || unsafeText(request.District) {
		return request, "invalid_district", "district is required and must be 100 safe characters or fewer"
	}
	if request.BeneficiaryCount < 0 {
		return request, "invalid_beneficiary_count", "beneficiaryCount must be zero or greater"
	}
	return request, "", ""
}

func normalizeUpdateAidRequest(request updateAidRequestRequest) (updateAidRequestRequest, string, string) {
	request.Status = normalizeToken(request.Status)

	if request.Status == "" && request.QuantityNeeded == 0 {
		return request, "no_changes", "at least one of status or quantityNeeded must be supplied"
	}
	if request.Status != "" && !allowedRequestStatuses[request.Status] {
		return request, "invalid_status", "status must be open, partially_fulfilled, fulfilled, or closed"
	}
	if request.QuantityNeeded < 0 {
		return request, "invalid_quantity_needed", "quantityNeeded must be zero or greater"
	}
	return request, "", ""
}

func normalizeCreatePledge(request createPledgeRequest) (createPledgeRequest, string, string) {
	request.DonorID = strings.TrimSpace(request.DonorID)
	request.DonorName = strings.TrimSpace(request.DonorName)
	request.ContactEmail = strings.TrimSpace(strings.ToLower(request.ContactEmail))
	request.ContactPhone = strings.TrimSpace(request.ContactPhone)
	request.DeliveryNote = strings.TrimSpace(request.DeliveryNote)

	if request.DonorID == "" {
		return request, "invalid_donor_id", "donorId is required"
	}
	if request.QuantityPledged <= 0 {
		return request, "invalid_quantity_pledged", "quantityPledged must be greater than zero"
	}
	if request.ContactEmail != "" && (len(request.ContactEmail) > 200 || !strings.Contains(request.ContactEmail, "@") || unsafeText(request.ContactEmail)) {
		return request, "invalid_contact_email", "contactEmail must be a valid email address"
	}
	if len(request.ContactPhone) > 50 || unsafeText(request.ContactPhone) {
		return request, "invalid_contact_phone", "contactPhone must be 50 safe characters or fewer"
	}
	if len(request.DeliveryNote) > 500 || unsafeText(request.DeliveryNote) {
		return request, "invalid_delivery_note", "deliveryNote must be 500 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeUpdatePledge(request updatePledgeRequest) (updatePledgeRequest, string, string) {
	request.Status = normalizeToken(request.Status)
	request.DeliveryNote = strings.TrimSpace(request.DeliveryNote)

	if request.Status == "" && request.QuantityDelivered == 0 && request.DeliveryNote == "" {
		return request, "no_changes", "at least one field must be supplied"
	}
	if request.Status != "" && !allowedPledgeStatuses[request.Status] {
		return request, "invalid_status", "status must be pledged, delivered, or cancelled"
	}
	if request.QuantityDelivered < 0 {
		return request, "invalid_quantity_delivered", "quantityDelivered must be zero or greater"
	}
	if len(request.DeliveryNote) > 500 || unsafeText(request.DeliveryNote) {
		return request, "invalid_delivery_note", "deliveryNote must be 500 safe characters or fewer"
	}
	return request, "", ""
}

func normalizeAllocate(request allocateRequest) (allocateRequest, string, string) {
	request.PledgeID = strings.TrimSpace(request.PledgeID)

	if request.PledgeID == "" {
		return request, "invalid_pledge_id", "pledgeId is required"
	}
	if request.Quantity <= 0 {
		return request, "invalid_quantity", "quantity must be greater than zero"
	}
	return request, "", ""
}

func requireAuthority(w http.ResponseWriter, r *http.Request) (authorityContext, bool) {
	ctx, ok := authorityContextFromRequest(r)
	if !ok {
		log.Printf("WARN donation-service authority_context_missing requestId=%s path=%s", ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return authorityContext{}, false
	}
	if !ctx.MFACompleted {
		log.Printf("WARN donation-service authority_mfa_required actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for authority actions")
		return authorityContext{}, false
	}
	if !authorityRoles[ctx.ActorRole] {
		log.Printf("WARN donation-service authority_forbidden actor=%s role=%s requestId=%s path=%s", ctx.ActorUserID, ctx.ActorRole, ctx.RequestID, r.URL.Path)
		writeError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this operation")
		return authorityContext{}, false
	}
	return ctx, true
}

func authorityContextFromRequest(r *http.Request) (authorityContext, bool) {
	ctx := authorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}
	return ctx, ctx.ActorUserID != "" && ctx.ActorAgencyID != "" && ctx.ActorRole != ""
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("ERROR donation-service write_json_response_failed error=%v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsFromEnv()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Cache-Control", "no-store")
}

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-NADAA-Actor-ID, X-NADAA-Actor-Role, X-NADAA-Agency-ID, X-NADAA-MFA-Completed, X-NADAA-Request-ID")
}

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func statusForCode(code string) int {
	if code == "not_found" || code == "donor_not_found" {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func unsafeText(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:")
}

func normalizeToken(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(strings.ToLower(value)), "-", "_"), " ", "_")
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

func copyDonors(source []donorRecord) []donorRecord {
	donors := make([]donorRecord, len(source))
	copy(donors, source)
	for i := range donors {
		donors[i].ItemsOffered = append([]string{}, donors[i].ItemsOffered...)
	}
	return donors
}

func copyAidRequests(source []aidRequestRecord) []aidRequestRecord {
	requests := make([]aidRequestRecord, len(source))
	copy(requests, source)
	return requests
}

func copyPledges(source []pledgeRecord) []pledgeRecord {
	pledges := make([]pledgeRecord, len(source))
	copy(pledges, source)
	return pledges
}
