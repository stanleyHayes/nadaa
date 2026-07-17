package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/store"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{
		store:  store.NewMemoryStore(now),
		now:    func() time.Time { return now },
		config: &config.Config{AllowMockActors: true},
	}
}

// newTokenTestServer builds a server that only accepts verified bearer tokens
// (mock actor headers disabled), signed with the given secret.
func newTokenTestServer(secret string) *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{
		store:  store.NewMemoryStore(now),
		now:    func() time.Time { return now },
		config: &config.Config{TokenSecret: secret},
	}
}

func TestNearbySheltersReturnsSortedSheltersAndRecoverySupport(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/shelters/nearby?lat=5.5600&lng=-0.2000", nil)

	srv.nearbySheltersHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.NearbyShelterResponse
	decodeResponse(t, response, &payload)
	if len(payload.Shelters) < 2 {
		t.Fatalf("expected nearby shelters, got %#v", payload.Shelters)
	}
	if payload.Shelters[0].ID != "00000000-0000-0000-0000-000000000301" || payload.Shelters[0].DistanceMeters > payload.Shelters[1].DistanceMeters {
		t.Fatalf("expected closest shelter first, got %#v", payload.Shelters)
	}
	if len(payload.RecoverySupport) == 0 || payload.RecoverySupport[0].DistanceMeters <= 0 {
		t.Fatalf("expected nearby recovery support with distances, got %#v", payload.RecoverySupport)
	}
}

func TestNearbySheltersRejectsInvalidCoordinates(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/shelters/nearby?lat=91&lng=-0.2000", nil)

	srv.nearbySheltersHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRecoverySupportNearby(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/recovery-support/nearby?lat=5.5600&lng=-0.2000", nil)

	srv.nearbyRecoverySupportHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.RecoverySupportResponse
	decodeResponse(t, response, &payload)
	if len(payload.RecoverySupport) == 0 {
		t.Fatalf("expected recovery support locations, got %#v", payload)
	}
}

func TestUpdateShelterOccupancyRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{}))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestUpdateShelterOccupancy(t *testing.T) {
	srv := newTestServer()
	occupancy := 450
	body := models.OccupancyUpdateRequest{CurrentOccupancy: &occupancy, Notes: "Shelter reached capacity during flood response."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(body))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ShelterUpdateResponse
	decodeResponse(t, response, &payload)
	if payload.Shelter.CurrentOccupancy != 450 || payload.Shelter.Status != "full" || payload.Shelter.UpdatedBy != "usr_shelter_operator" {
		t.Fatalf("expected full updated shelter, got %#v", payload.Shelter)
	}
}

func TestUpdateShelterOccupancyRejectsOverCapacity(t *testing.T) {
	srv := newTestServer()
	capacity := 10
	occupancy := 11
	body := models.OccupancyUpdateRequest{Capacity: &capacity, CurrentOccupancy: &occupancy}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(body))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestHospitalCapacityListFiltersAndMarksStale(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hospitals/capacity?lat=5.5600&lng=-0.2000&service=emergency&includeStale=true", nil)

	srv.listHospitalCapacityHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.HospitalCapacityResponse
	decodeResponse(t, response, &payload)
	if len(payload.Facilities) < 2 || payload.StaleThresholdMinutes != 30 {
		t.Fatalf("expected hospital capacity records and threshold, got %#v", payload)
	}
	if payload.Facilities[0].DistanceMeters <= 0 {
		t.Fatalf("expected distance-enriched facility, got %#v", payload.Facilities[0])
	}
	foundStale := false
	for _, facility := range payload.Facilities {
		if facility.ID == "hospital_003" {
			foundStale = facility.Stale && facility.StaleReason != ""
		}
	}
	if !foundStale {
		t.Fatalf("expected stale Tema hospital warning, got %#v", payload.Facilities)
	}
}

func TestHospitalCapacityListCanHideStaleAndFilterBeds(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/hospitals/capacity?includeStale=false&minAvailableBeds=20", nil)

	srv.listHospitalCapacityHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.HospitalCapacityResponse
	decodeResponse(t, response, &payload)
	if len(payload.Facilities) != 1 || payload.Facilities[0].ID != "hospital_001" || payload.Facilities[0].Stale {
		t.Fatalf("expected one fresh hospital with enough beds, got %#v", payload.Facilities)
	}
}

func TestUpdateHospitalCapacityRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/hospitals/hospital_001/capacity", jsonBody(models.HospitalCapacityUpdateRequest{}))
	request.SetPathValue("id", "hospital_001")

	srv.updateHospitalCapacityHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestUpdateHospitalCapacity(t *testing.T) {
	srv := newTestServer()
	availableBeds := 18
	icuBeds := 2
	ambulances := 1
	oxygen := true
	body := models.HospitalCapacityUpdateRequest{
		AvailableBeds:       &availableBeds,
		ICUBedsAvailable:    &icuBeds,
		EmergencyCapacity:   "limited",
		EmergencyUnitStatus: "busy",
		AmbulancesAvailable: &ambulances,
		OxygenAvailable:     &oxygen,
		Notes:               "Manual update from emergency desk.",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/hospitals/hospital_001/capacity", jsonBody(body))
	request.SetPathValue("id", "hospital_001")

	srv.updateHospitalCapacityHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.HospitalCapacityUpdateResponse
	decodeResponse(t, response, &payload)
	if payload.Facility.AvailableBeds != 18 ||
		payload.Facility.ICUBedsAvailable != 2 ||
		payload.Facility.EmergencyCapacity != "limited" ||
		payload.Facility.Source != "fixture" ||
		payload.Facility.SourceRef != "hospital-capacity-feed" ||
		payload.Facility.UpdatedBy != "usr_shelter_operator" ||
		payload.Facility.Stale {
		t.Fatalf("expected hospital update with preserved provenance, got %#v", payload.Facility)
	}
}

func TestUpdateHospitalCapacityOverwritesSourceOnlyWhenProvided(t *testing.T) {
	srv := newTestServer()
	availableBeds := 20
	body := models.HospitalCapacityUpdateRequest{
		AvailableBeds: &availableBeds,
		Source:        "manual",
		SourceRef:     "desk-call-123",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/hospitals/hospital_001/capacity", jsonBody(body))
	request.SetPathValue("id", "hospital_001")

	srv.updateHospitalCapacityHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.HospitalCapacityUpdateResponse
	decodeResponse(t, response, &payload)
	if payload.Facility.Source != "manual" || payload.Facility.SourceRef != "desk-call-123" {
		t.Fatalf("expected provided provenance to be stored, got %#v", payload.Facility)
	}
}

func TestHospitalCapacityFixtureImport(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/hospitals/capacity/imports/fixture", jsonBody(models.HospitalCapacityImportRequest{}))

	srv.importHospitalCapacityFixtureHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.HospitalCapacityImportResponse
	decodeResponse(t, response, &payload)
	if payload.Imported != 2 || payload.Source != "fixture_adapter" {
		t.Fatalf("expected two imported fixture updates, got %#v", payload)
	}
	for _, facility := range payload.Facilities {
		if facility.Source != "fixture_adapter" || facility.SourceRef != "hospital-capacity-feed" || facility.Stale {
			t.Fatalf("expected fresh fixture-sourced facility, got %#v", facility)
		}
	}
}

func TestListReliefPointsFiltersByStatus(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/relief-points?status=open", nil)

	srv.listReliefPointsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ReliefPointListResponse
	decodeResponse(t, response, &payload)
	if len(payload.ReliefPoints) < 2 {
		t.Fatalf("expected open relief points, got %#v", payload)
	}
	for _, point := range payload.ReliefPoints {
		if point.Status != "open" {
			t.Fatalf("expected only open points, got %#v", point)
		}
	}
}

func TestNearbyReliefPoints(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/relief-points/nearby?lat=5.5600&lng=-0.2000", nil)

	srv.nearbyReliefPointsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ReliefPointNearbyResponse
	decodeResponse(t, response, &payload)
	if len(payload.ReliefPoints) == 0 {
		t.Fatalf("expected nearby relief points, got %#v", payload)
	}
	if payload.ReliefPoints[0].DistanceMeters < 0 {
		t.Fatalf("expected distance-enriched relief point, got %#v", payload.ReliefPoints[0])
	}
}

func TestCreateReliefPointRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/relief-points", jsonBody(models.CreateReliefPointRequest{
		Name:     "Unauthorized Relief Point",
		Type:     "food",
		Location: models.Coordinates{Lat: 5.55, Lng: -0.19},
	}))

	srv.createReliefPointHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateReliefPoint(t *testing.T) {
	srv := newTestServer()
	body := models.CreateReliefPointRequest{
		Name:     "Test Food Point",
		Type:     "food",
		Location: models.Coordinates{Lat: 5.55, Lng: -0.19},
		StockCategories: []models.ReliefStockCategory{
			{Category: "rice_kg", Quantity: 100, Unit: "kg"},
		},
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/relief-points", jsonBody(body))

	srv.createReliefPointHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.ReliefPoint
	decodeResponse(t, response, &payload)
	if payload.Name != "Test Food Point" || payload.Status != "open" || payload.CreatedBy != "usr_shelter_operator" {
		t.Fatalf("expected created relief point, got %#v", payload)
	}
	if len(payload.StockCategories) != 1 || payload.StockCategories[0].Category != "rice_kg" {
		t.Fatalf("expected stock categories, got %#v", payload.StockCategories)
	}
}

func TestUpdateReliefPointRecordsStockHistory(t *testing.T) {
	srv := newTestServer()
	body := models.UpdateReliefPointRequest{
		Status: "limited",
		StockCategories: []models.ReliefStockCategory{
			{Category: "rice_kg", Quantity: 50, Unit: "kg"},
			{Category: "water_bottles", Quantity: 100, Unit: "bottles"},
		},
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/relief-points/relief_ama_food_001", jsonBody(body))
	request.SetPathValue("id", "relief_ama_food_001")

	srv.updateReliefPointHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ReliefPoint
	decodeResponse(t, response, &payload)
	if payload.Status != "limited" || payload.UpdatedBy != "usr_shelter_operator" {
		t.Fatalf("expected updated relief point, got %#v", payload)
	}

	historyResponse := httptest.NewRecorder()
	historyRequest := httptest.NewRequest(http.MethodGet, "/api/v1/relief-points/relief_ama_food_001/stock-history", nil)
	historyRequest.SetPathValue("id", "relief_ama_food_001")
	srv.listReliefPointStockHistoryHandler(historyResponse, historyRequest)

	if historyResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, historyResponse.Code, historyResponse.Body.String())
	}

	var historyPayload models.ReliefPointStockHistoryResponse
	decodeResponse(t, historyResponse, &historyPayload)
	if len(historyPayload.History) != 1 {
		t.Fatalf("expected one stock history entry, got %#v", historyPayload.History)
	}
}

func TestPublicAidRequestListExcludesPendingPrivateNeeds(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests", nil)

	srv.listAidRequestsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.AidRequestListResponse
	decodeResponse(t, response, &payload)
	if len(payload.AidRequests) != 1 {
		t.Fatalf("expected one public aid request, got %#v", payload.AidRequests)
	}
	if payload.AidRequests[0].ID != "aid_ama_hygiene_001" || payload.AidRequests[0].QuantityPledged != 80 {
		t.Fatalf("expected public hygiene request with pledged summary, got %#v", payload.AidRequests[0])
	}
}

func TestCreateReviewAndPledgeAidRequest(t *testing.T) {
	srv := newTestServer()
	createBody := models.CreateAidRequestRequest{
		Title:                 "Baby food for displaced families",
		Category:              "food",
		Priority:              "urgent",
		Region:                "Greater Accra",
		District:              "Accra Metropolitan",
		Location:              models.Coordinates{Lat: 5.560, Lng: -0.200},
		ReceivingOrganization: "AMA Central Food Distribution",
		Contact:               "0302112233",
		QuantityNeeded:        120,
		QuantityUnit:          "packs",
		Description:           "Baby food packs for infants staying near the flood relief point.",
		NeededBy:              srv.now().Add(24 * time.Hour),
		Visibility:            "public",
		SourceReliefPointID:   "relief_ama_food_001",
	}

	createResponse := httptest.NewRecorder()
	createRequest := authorityRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(createBody))
	srv.createAidRequestHandler(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var created models.AidRequest
	decodeResponse(t, createResponse, &created)
	if created.Status != "pending_review" || created.CreatedBy != "usr_shelter_operator" {
		t.Fatalf("expected pending review aid request, got %#v", created)
	}

	reviewResponse := httptest.NewRecorder()
	reviewRequest := authorityRequest(http.MethodPatch, "/api/v1/aid-requests/"+created.ID+"/review", jsonBody(models.ReviewAidRequestRequest{
		Status:         "approved",
		ApprovalNotes:  "Verified receiving desk and category need.",
		AntiFraudNotes: "Contact checked against relief point operator.",
	}))
	reviewRequest.SetPathValue("id", created.ID)
	srv.reviewAidRequestHandler(reviewResponse, reviewRequest)

	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
	}

	pledgeResponse := httptest.NewRecorder()
	pledgeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/"+created.ID+"/pledges", jsonBody(models.CreateAidPledgeRequest{
		DonorName: "Neighborhood Grocers Association",
		DonorType: "business",
		Contact:   "donations@example.org",
		Quantity:  60,
		Unit:      "packs",
		Note:      "Can deliver tomorrow morning.",
	}))
	pledgeRequest.SetPathValue("id", created.ID)
	srv.createAidPledgeHandler(pledgeResponse, pledgeRequest)

	if pledgeResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, pledgeResponse.Code, pledgeResponse.Body.String())
	}

	var pledge models.AidPledge
	decodeResponse(t, pledgeResponse, &pledge)
	if pledge.Status != "pledged" || pledge.ReviewStatus != "pending_review" || pledge.AidRequestID != created.ID {
		t.Fatalf("expected created pledge, got %#v", pledge)
	}

	pledgesResponse := httptest.NewRecorder()
	pledgesRequest := authorityRequest(http.MethodGet, "/api/v1/aid-requests/"+created.ID+"/pledges", nil)
	pledgesRequest.SetPathValue("id", created.ID)
	srv.listAidPledgesHandler(pledgesResponse, pledgesRequest)

	if pledgesResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, pledgesResponse.Code, pledgesResponse.Body.String())
	}

	var pledgesPayload models.AidPledgeListResponse
	decodeResponse(t, pledgesResponse, &pledgesPayload)
	if len(pledgesPayload.Pledges) != 1 {
		t.Fatalf("expected one pledge, got %#v", pledgesPayload.Pledges)
	}
}

func TestCreateAidPledgeRejectsPendingRequest(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/aid_madina_water_001/pledges", jsonBody(models.CreateAidPledgeRequest{
		DonorName: "Early Donor",
		DonorType: "individual",
		Contact:   "donor@example.org",
		Quantity:  10,
		Unit:      "boxes",
	}))
	request.SetPathValue("id", "aid_madina_water_001")

	srv.createAidPledgeHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestAidRequestExportRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/report.csv", nil)

	srv.exportAidRequestsHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestDeleteShelter(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := adminRequest(http.MethodDelete, "/api/v1/shelters/00000000-0000-0000-0000-000000000301", nil)
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.deleteShelterHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, response.Code, response.Body.String())
	}
}

func TestDeleteShelterNotFound(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := adminRequest(http.MethodDelete, "/api/v1/shelters/missing", nil)
	request.SetPathValue("id", "missing")

	srv.deleteShelterHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestDeleteReliefPoint(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := adminRequest(http.MethodDelete, "/api/v1/relief-points/relief_ama_food_001", nil)
	request.SetPathValue("id", "relief_ama_food_001")

	srv.deleteReliefPointHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, response.Code, response.Body.String())
	}
}

func TestDeleteReliefPointNotFound(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := adminRequest(http.MethodDelete, "/api/v1/relief-points/missing", nil)
	request.SetPathValue("id", "missing")

	srv.deleteReliefPointHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestDeleteAidRequest(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := adminRequest(http.MethodDelete, "/api/v1/aid-requests/aid_ama_hygiene_001", nil)
	request.SetPathValue("id", "aid_ama_hygiene_001")

	srv.deleteAidRequestHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, response.Code, response.Body.String())
	}
}

func TestDeleteAidRequestNotFound(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := adminRequest(http.MethodDelete, "/api/v1/aid-requests/missing", nil)
	request.SetPathValue("id", "missing")

	srv.deleteAidRequestHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestAidRequestIDsDoNotCollideAfterDelete(t *testing.T) {
	srv := newTestServer()
	createBody := models.CreateAidRequestRequest{
		Title:                 "Water for displaced households",
		Category:              "water",
		Priority:              "high",
		Location:              models.Coordinates{Lat: 5.560, Lng: -0.200},
		ReceivingOrganization: "AMA Central Food Distribution",
		QuantityNeeded:        100,
		QuantityUnit:          "bottles",
		Description:           "Bottled water for households displaced by flooding.",
		NeededBy:              srv.now().Add(24 * time.Hour),
	}

	firstResponse := httptest.NewRecorder()
	srv.createAidRequestHandler(firstResponse, authorityRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(createBody)))
	if firstResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, firstResponse.Code, firstResponse.Body.String())
	}
	var first models.AidRequest
	decodeResponse(t, firstResponse, &first)

	deleteResponse := httptest.NewRecorder()
	deleteRequest := adminRequest(http.MethodDelete, "/api/v1/aid-requests/"+first.ID, nil)
	deleteRequest.SetPathValue("id", first.ID)
	srv.deleteAidRequestHandler(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResponse.Code)
	}

	secondResponse := httptest.NewRecorder()
	srv.createAidRequestHandler(secondResponse, authorityRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(createBody)))
	if secondResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, secondResponse.Code, secondResponse.Body.String())
	}
	var second models.AidRequest
	decodeResponse(t, secondResponse, &second)
	if first.ID == "" || second.ID == first.ID {
		t.Fatalf("expected a fresh aid request ID after delete, got first=%s second=%s", first.ID, second.ID)
	}
}

func TestReliefPointIDsDoNotCollideAfterDelete(t *testing.T) {
	srv := newTestServer()
	createBody := models.CreateReliefPointRequest{
		Name:     "Collision Test Point",
		Type:     "food",
		Location: models.Coordinates{Lat: 5.55, Lng: -0.19},
	}

	firstResponse := httptest.NewRecorder()
	srv.createReliefPointHandler(firstResponse, authorityRequest(http.MethodPost, "/api/v1/relief-points", jsonBody(createBody)))
	if firstResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, firstResponse.Code, firstResponse.Body.String())
	}
	var first models.ReliefPoint
	decodeResponse(t, firstResponse, &first)

	deleteResponse := httptest.NewRecorder()
	deleteRequest := adminRequest(http.MethodDelete, "/api/v1/relief-points/"+first.ID, nil)
	deleteRequest.SetPathValue("id", first.ID)
	srv.deleteReliefPointHandler(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResponse.Code)
	}

	secondResponse := httptest.NewRecorder()
	srv.createReliefPointHandler(secondResponse, authorityRequest(http.MethodPost, "/api/v1/relief-points", jsonBody(createBody)))
	if secondResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, secondResponse.Code, secondResponse.Body.String())
	}
	var second models.ReliefPoint
	decodeResponse(t, secondResponse, &second)
	if first.ID == "" || second.ID == first.ID {
		t.Fatalf("expected a fresh relief point ID after delete, got first=%s second=%s", first.ID, second.ID)
	}
}

func TestPrivateAidListRejectsResponderRole(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := agencyRoleRequest(http.MethodGet, "/api/v1/aid-requests?includePrivate=true", nil, "responder", "00000000-0000-0000-0000-000000000204")

	srv.listAidRequestsHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
}

func TestPrivateAidListScopesAgencyRolesToOwnAgency(t *testing.T) {
	srv := newTestServer()

	ownResponse := httptest.NewRecorder()
	srv.listAidRequestsHandler(ownResponse, agencyRoleRequest(http.MethodGet, "/api/v1/aid-requests?includePrivate=true", nil, "agency_admin", "00000000-0000-0000-0000-000000000204"))
	if ownResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, ownResponse.Code, ownResponse.Body.String())
	}
	var ownPayload models.AidRequestListResponse
	decodeResponse(t, ownResponse, &ownPayload)
	if len(ownPayload.AidRequests) != 2 {
		t.Fatalf("expected own-agency private request plus the public one, got %#v", ownPayload.AidRequests)
	}

	otherResponse := httptest.NewRecorder()
	srv.listAidRequestsHandler(otherResponse, agencyRoleRequest(http.MethodGet, "/api/v1/aid-requests?includePrivate=true", nil, "agency_admin", "00000000-0000-0000-0000-000000000999"))
	if otherResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, otherResponse.Code, otherResponse.Body.String())
	}
	var otherPayload models.AidRequestListResponse
	decodeResponse(t, otherResponse, &otherPayload)
	if len(otherPayload.AidRequests) != 1 || otherPayload.AidRequests[0].Visibility != "public" {
		t.Fatalf("expected only the public aid request for another agency, got %#v", otherPayload.AidRequests)
	}
}

func TestPrivateAidListPrivilegedRoleSeesAll(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/aid-requests?includePrivate=true", nil)

	srv.listAidRequestsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.AidRequestListResponse
	decodeResponse(t, response, &payload)
	if len(payload.AidRequests) != 2 {
		t.Fatalf("expected district_officer to see every aid request, got %#v", payload.AidRequests)
	}
}

func TestFulfilledAidRequestReopensWhenPledgeCancelled(t *testing.T) {
	srv := newTestServer()
	createBody := models.CreateAidRequestRequest{
		Title:                 "Blankets for flood shelters",
		Category:              "shelter",
		Priority:              "high",
		Location:              models.Coordinates{Lat: 5.560, Lng: -0.200},
		ReceivingOrganization: "AMA Central Food Distribution",
		QuantityNeeded:        60,
		QuantityUnit:          "blankets",
		Description:           "Blankets for families staying at the evacuation shelter.",
		NeededBy:              srv.now().Add(24 * time.Hour),
		Visibility:            "public",
	}

	createResponse := httptest.NewRecorder()
	srv.createAidRequestHandler(createResponse, authorityRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(createBody)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.AidRequest
	decodeResponse(t, createResponse, &created)

	reviewResponse := httptest.NewRecorder()
	reviewRequest := authorityRequest(http.MethodPatch, "/api/v1/aid-requests/"+created.ID+"/review", jsonBody(models.ReviewAidRequestRequest{
		Status:        "approved",
		ApprovalNotes: "Verified receiving desk.",
	}))
	reviewRequest.SetPathValue("id", created.ID)
	srv.reviewAidRequestHandler(reviewResponse, reviewRequest)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
	}

	pledgeResponse := httptest.NewRecorder()
	pledgeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/"+created.ID+"/pledges", jsonBody(models.CreateAidPledgeRequest{
		DonorName: "Community Textiles",
		DonorType: "business",
		Contact:   "donations@example.org",
		Quantity:  60,
		Unit:      "blankets",
	}))
	pledgeRequest.SetPathValue("id", created.ID)
	srv.createAidPledgeHandler(pledgeResponse, pledgeRequest)
	if pledgeResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, pledgeResponse.Code, pledgeResponse.Body.String())
	}
	var pledge models.AidPledge
	decodeResponse(t, pledgeResponse, &pledge)

	if status := aidRequestStatus(t, srv, created.ID); status != "fulfilled" {
		t.Fatalf("expected fulfilled aid request after full pledge, got %s", status)
	}

	cancelResponse := httptest.NewRecorder()
	cancelRequest := authorityRequest(http.MethodPatch, "/api/v1/aid-requests/"+created.ID+"/pledges/"+pledge.ID+"/review", jsonBody(models.ReviewAidPledgeRequest{
		Status: "cancelled",
	}))
	cancelRequest.SetPathValue("id", created.ID)
	cancelRequest.SetPathValue("pledgeId", pledge.ID)
	srv.reviewAidPledgeHandler(cancelResponse, cancelRequest)
	if cancelResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, cancelResponse.Code, cancelResponse.Body.String())
	}

	if status := aidRequestStatus(t, srv, created.ID); status != "open" {
		t.Fatalf("expected aid request to reopen after pledge cancellation, got %s", status)
	}
}

func aidRequestStatus(t *testing.T, srv *server, id string) string {
	t.Helper()
	response := httptest.NewRecorder()
	srv.listAidRequestsHandler(response, authorityRequest(http.MethodGet, "/api/v1/aid-requests?includePrivate=true", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.AidRequestListResponse
	decodeResponse(t, response, &payload)
	for _, request := range payload.AidRequests {
		if request.ID == id {
			return request.Status
		}
	}
	t.Fatalf("aid request %s not found in list", id)
	return ""
}

func TestUpdateShelterRejectsCapacityBelowExistingOccupancy(t *testing.T) {
	srv := newTestServer()
	capacity := 50
	body := models.OccupancyUpdateRequest{Capacity: &capacity}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(body))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestOccupancyOnlyUpdatePreservesClosedStatus(t *testing.T) {
	srv := newTestServer()

	closeResponse := httptest.NewRecorder()
	closeRequest := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{Status: "closed"}))
	closeRequest.SetPathValue("id", "00000000-0000-0000-0000-000000000301")
	srv.updateShelterOccupancyHandler(closeResponse, closeRequest)
	if closeResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, closeResponse.Code, closeResponse.Body.String())
	}

	occupancy := 450
	occupancyResponse := httptest.NewRecorder()
	occupancyRequest := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{CurrentOccupancy: &occupancy}))
	occupancyRequest.SetPathValue("id", "00000000-0000-0000-0000-000000000301")
	srv.updateShelterOccupancyHandler(occupancyResponse, occupancyRequest)
	if occupancyResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, occupancyResponse.Code, occupancyResponse.Body.String())
	}

	var payload models.ShelterUpdateResponse
	decodeResponse(t, occupancyResponse, &payload)
	if payload.Shelter.Status != "closed" {
		t.Fatalf("expected closed status to survive an occupancy-only update, got %#v", payload.Shelter)
	}
}

func TestHospitalCapacityFixtureImportSkipsOverbookedRecords(t *testing.T) {
	srv := newTestServer()
	body := models.HospitalCapacityImportRequest{
		Records: []models.HospitalCapacityFixtureRecord{
			{FacilityID: "hospital_001", AvailableBeds: 900, EmergencyCapacity: "available"},
			{FacilityID: "hospital_002", AvailableBeds: 100, EmergencyCapacity: "available"},
		},
	}

	response := httptest.NewRecorder()
	srv.importHospitalCapacityFixtureHandler(response, authorityRequest(http.MethodPost, "/api/v1/hospitals/capacity/imports/fixture", jsonBody(body)))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.HospitalCapacityImportResponse
	decodeResponse(t, response, &payload)
	if payload.Imported != 1 || len(payload.Facilities) != 1 || payload.Facilities[0].ID != "hospital_002" {
		t.Fatalf("expected only the valid record to import, got %#v", payload)
	}

	listResponse := httptest.NewRecorder()
	srv.listHospitalCapacityHandler(listResponse, httptest.NewRequest(http.MethodGet, "/api/v1/hospitals/capacity", nil))
	var listPayload models.HospitalCapacityResponse
	decodeResponse(t, listResponse, &listPayload)
	for _, facility := range listPayload.Facilities {
		if facility.ID == "hospital_001" && facility.AvailableBeds != 46 {
			t.Fatalf("expected overbooked record to be skipped, got %#v", facility)
		}
	}
}

func TestAidRequestExportEscapesFormulaPrefixes(t *testing.T) {
	srv := newTestServer()
	createBody := models.CreateAidRequestRequest{
		Title:                 "=SUM(1,2) emergency packs",
		Category:              "food",
		Priority:              "high",
		Location:              models.Coordinates{Lat: 5.560, Lng: -0.200},
		ReceivingOrganization: "Shelter Org, Inc.",
		QuantityNeeded:        40,
		QuantityUnit:          "packs",
		Description:           "Food packs for families at the evacuation shelter.",
		NeededBy:              srv.now().Add(24 * time.Hour),
		Visibility:            "public",
	}

	createResponse := httptest.NewRecorder()
	srv.createAidRequestHandler(createResponse, authorityRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(createBody)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	exportResponse := httptest.NewRecorder()
	srv.exportAidRequestsHandler(exportResponse, authorityRequest(http.MethodGet, "/api/v1/aid-requests/report.csv", nil))
	if exportResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, exportResponse.Code, exportResponse.Body.String())
	}

	records, err := csv.NewReader(strings.NewReader(exportResponse.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("expected parseable CSV, got %v", err)
	}
	if len(records) != 4 || records[0][0] != "id" || records[0][1] != "title" {
		t.Fatalf("expected header plus three aid request rows, got %#v", records)
	}
	found := false
	for _, record := range records[1:] {
		if record[1] == "'=SUM(1,2) emergency packs" {
			found = true
			if record[6] != "Shelter Org, Inc." {
				t.Fatalf("expected comma-containing organization to round-trip, got %#v", record)
			}
		}
		if strings.HasPrefix(record[1], "=") || strings.HasPrefix(record[6], "=") {
			t.Fatalf("expected formula prefixes to be escaped, got %#v", record)
		}
	}
	if !found {
		t.Fatalf("expected the created aid request row in the export, got %#v", records)
	}
}

func TestAuthorityAcceptsValidBearerToken(t *testing.T) {
	srv := newTokenTestServer("test-secret-key")
	token := signedToken(t, "test-secret-key", map[string]any{
		"sub":      "usr_token_operator",
		"typ":      "agency",
		"role":     "district_officer",
		"agencyId": "00000000-0000-0000-0000-000000000204",
		"mfa":      true,
		"exp":      srv.now().Add(time.Hour).Unix(),
	})
	occupancy := 120

	response := httptest.NewRecorder()
	request := tokenRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{CurrentOccupancy: &occupancy}), token)
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.ShelterUpdateResponse
	decodeResponse(t, response, &payload)
	if payload.Shelter.CurrentOccupancy != 120 || payload.Shelter.UpdatedBy != "usr_token_operator" {
		t.Fatalf("expected token-derived actor on the update, got %#v", payload.Shelter)
	}
}

func TestAuthorityIgnoresMockHeadersWhenDisabled(t *testing.T) {
	srv := newTokenTestServer("test-secret-key")
	occupancy := 120

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{CurrentOccupancy: &occupancy}))
	request.SetPathValue("id", "00000000-0000-0000-0000-000000000301")

	srv.updateShelterOccupancyHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
}

func TestAuthorityRejectsTamperedAndExpiredTokens(t *testing.T) {
	srv := newTokenTestServer("test-secret-key")
	occupancy := 120

	tampered := signedToken(t, "wrong-secret", map[string]any{
		"sub":      "usr_token_operator",
		"typ":      "agency",
		"role":     "district_officer",
		"agencyId": "00000000-0000-0000-0000-000000000204",
		"mfa":      true,
		"exp":      srv.now().Add(time.Hour).Unix(),
	})
	tamperedResponse := httptest.NewRecorder()
	tamperedRequest := tokenRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{CurrentOccupancy: &occupancy}), tampered)
	tamperedRequest.SetPathValue("id", "00000000-0000-0000-0000-000000000301")
	srv.updateShelterOccupancyHandler(tamperedResponse, tamperedRequest)
	if tamperedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for a wrongly-signed token, got %d", http.StatusUnauthorized, tamperedResponse.Code)
	}

	expired := signedToken(t, "test-secret-key", map[string]any{
		"sub":      "usr_token_operator",
		"typ":      "agency",
		"role":     "district_officer",
		"agencyId": "00000000-0000-0000-0000-000000000204",
		"mfa":      true,
		"exp":      srv.now().Add(-time.Hour).Unix(),
	})
	expiredResponse := httptest.NewRecorder()
	expiredRequest := tokenRequest(http.MethodPatch, "/api/v1/shelters/00000000-0000-0000-0000-000000000301/occupancy", jsonBody(models.OccupancyUpdateRequest{CurrentOccupancy: &occupancy}), expired)
	expiredRequest.SetPathValue("id", "00000000-0000-0000-0000-000000000301")
	srv.updateShelterOccupancyHandler(expiredResponse, expiredRequest)
	if expiredResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for an expired token, got %d", http.StatusUnauthorized, expiredResponse.Code)
	}
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func authorityRequest(method string, target string, body *bytes.Reader) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_shelter_operator")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-shelter-update")
	return request
}

// adminRequest authenticates as a system_admin — required for the admin-only
// delete endpoints (ShelterDeleteRoles).
func adminRequest(method string, target string, body *bytes.Reader) *http.Request {
	request := authorityRequest(method, target, body)
	request.Header.Set("X-NADAA-Actor-ID", "usr_shelter_admin")
	request.Header.Set("X-NADAA-Actor-Role", "system_admin")
	return request
}

// agencyRoleRequest authenticates with a specific mock role and agency.
func agencyRoleRequest(method string, target string, body *bytes.Reader, role, agencyID string) *http.Request {
	request := authorityRequest(method, target, body)
	request.Header.Set("X-NADAA-Actor-Role", role)
	request.Header.Set("X-NADAA-Agency-ID", agencyID)
	return request
}

// tokenRequest builds a request carrying a bearer token.
func tokenRequest(method string, target string, body *bytes.Reader, token string) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

// signedToken signs test claims with the same nadaa.<payload>.<sig> scheme as
// auth-service (HMAC-SHA256 over the encoded payload).
func signedToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	return "nadaa." + encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
