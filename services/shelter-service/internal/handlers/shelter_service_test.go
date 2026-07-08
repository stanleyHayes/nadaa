package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/shelter-service/internal/store"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{store: store.NewMemoryStore(now), now: func() time.Time { return now }}
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
		payload.Facility.Source != "manual" ||
		payload.Facility.UpdatedBy != "usr_shelter_operator" ||
		payload.Facility.Stale {
		t.Fatalf("expected manual hospital update metadata, got %#v", payload.Facility)
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
		Name: "Test Food Point",
		Type: "food",
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

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
