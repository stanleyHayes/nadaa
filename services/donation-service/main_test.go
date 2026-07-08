package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	return &server{store: newMemoryStore(now), now: func() time.Time { return now }}
}

func TestHealthz(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestListAidCatalog(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-catalog", nil)

	srv.listCatalogHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload aidCatalogResponse
	decodeResponse(t, response, &payload)
	if len(payload.Items) < 5 {
		t.Fatalf("expected at least 5 catalog items, got %#v", payload.Items)
	}
	if payload.Items[0].PriorityScore < payload.Items[1].PriorityScore {
		t.Fatalf("expected catalog sorted by priority descending, got %#v", payload.Items)
	}
}

func TestListAidRequests(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests", nil)

	srv.listAidRequestsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload aidRequestListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Requests) != 2 {
		t.Fatalf("expected 2 seeded aid requests, got %#v", payload.Requests)
	}
	if payload.Requests[0].Priority != "critical" {
		t.Fatalf("expected critical request first, got %#v", payload.Requests)
	}
}

func TestListAidRequestsFiltersByCategory(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests?category=medical", nil)

	srv.listAidRequestsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload aidRequestListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Requests) != 1 || payload.Requests[0].Category != "medical" {
		t.Fatalf("expected one medical request, got %#v", payload.Requests)
	}
}

func TestCreateDonorPublic(t *testing.T) {
	srv := newTestServer()
	body := createDonorRequest{
		Name:         "Accra Community Relief",
		Type:         "ngo",
		ContactName:  "Ama Mensah",
		ContactEmail: "relief@example.com",
		ContactPhone: "+233201234567",
		Region:       "Greater Accra",
		District:     "Accra Metropolitan",
		ItemsOffered: []string{"food parcels", "water"},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/donors", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.createDonorHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload donorRecord
	decodeResponse(t, response, &payload)
	if payload.Name != body.Name || payload.Type != "ngo" || payload.Status != "active" || payload.CreatedBy != "public" {
		t.Fatalf("expected created donor, got %#v", payload)
	}
}

func TestCreateAidRequestRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	body := createAidRequestRequest{
		Title:          "Test request",
		Category:       "water",
		ItemCode:       "water_liter",
		QuantityNeeded: 100,
		Unit:           "liters",
		Priority:       "high",
		Region:         "Greater Accra",
		District:       "Tema Metropolitan",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.createAidRequestHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateAndUpdateAidRequestAuthority(t *testing.T) {
	srv := newTestServer()
	body := createAidRequestRequest{
		Title:          "Water for Madina",
		Category:       "water",
		ItemCode:       "water_liter",
		QuantityNeeded: 1000,
		Unit:           "liters",
		Priority:       "high",
		Region:         "Greater Accra",
		District:       "La Nkwantanang Madina",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/aid-requests", jsonBody(body))

	srv.createAidRequestHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var created aidRequestRecord
	decodeResponse(t, response, &created)
	if created.Status != "open" || created.RequestedBy != "usr_donation_operator" {
		t.Fatalf("expected created aid request, got %#v", created)
	}

	update := updateAidRequestRequest{Status: "closed"}
	updateResponse := httptest.NewRecorder()
	updateRequest := authorityRequest(http.MethodPatch, "/api/v1/aid-requests/"+created.ID, jsonBody(update))
	updateRequest.SetPathValue("id", created.ID)

	srv.updateAidRequestHandler(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	var updated aidRequestRecord
	decodeResponse(t, updateResponse, &updated)
	if updated.Status != "closed" {
		t.Fatalf("expected closed status, got %#v", updated)
	}
}

func TestUpdateDonorRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/donors/donor_001", jsonBody(updateDonorRequest{Status: "inactive"}))
	request.SetPathValue("id", "donor_001")

	srv.updateDonorHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateAndUpdateDonorAuthority(t *testing.T) {
	srv := newTestServer()
	body := createDonorRequest{
		Name:         "Govt Relief Fund",
		Type:         "government",
		ContactName:  "Kofi Asante",
		ContactEmail: "kofi@example.gov",
		Region:       "Greater Accra",
		District:     "Accra Metropolitan",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/donors", jsonBody(body))

	srv.createDonorHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var created donorRecord
	decodeResponse(t, response, &created)
	if created.CreatedBy != "usr_donation_operator" {
		t.Fatalf("expected authority createdBy, got %#v", created)
	}

	update := updateDonorRequest{Status: "inactive", Notes: "Paused for review"}
	updateResponse := httptest.NewRecorder()
	updateRequest := authorityRequest(http.MethodPatch, "/api/v1/donors/"+created.ID, jsonBody(update))
	updateRequest.SetPathValue("id", created.ID)

	srv.updateDonorHandler(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	var updated donorRecord
	decodeResponse(t, updateResponse, &updated)
	if updated.Status != "inactive" || updated.Notes != "Paused for review" {
		t.Fatalf("expected updated donor, got %#v", updated)
	}
}

func TestPledgeCreatesAndUpdatesRequestStatus(t *testing.T) {
	srv := newTestServer()

	donorBody := createDonorRequest{
		Name:         "Tema Relief Group",
		Type:         "organization",
		ContactEmail: "tema@example.com",
		Region:       "Greater Accra",
		District:     "Tema Metropolitan",
	}
	donorResponse := httptest.NewRecorder()
	donorRequest := httptest.NewRequest(http.MethodPost, "/api/v1/donors", jsonBody(donorBody))
	donorRequest.Header.Set("Content-Type", "application/json")
	srv.createDonorHandler(donorResponse, donorRequest)
	if donorResponse.Code != http.StatusCreated {
		t.Fatalf("expected donor created, got %d: %s", donorResponse.Code, donorResponse.Body.String())
	}
	var donor donorRecord
	decodeResponse(t, donorResponse, &donor)

	pledgeBody := createPledgeRequest{
		DonorID:         donor.ID,
		QuantityPledged: 300,
	}
	pledgeResponse := httptest.NewRecorder()
	pledgeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(pledgeBody))
	pledgeRequest.Header.Set("Content-Type", "application/json")
	pledgeRequest.SetPathValue("id", "request_001")
	srv.createPledgeHandler(pledgeResponse, pledgeRequest)
	if pledgeResponse.Code != http.StatusCreated {
		t.Fatalf("expected pledge created, got %d: %s", pledgeResponse.Code, pledgeResponse.Body.String())
	}
	var pledge pledgeRecord
	decodeResponse(t, pledgeResponse, &pledge)
	if pledge.Status != "pledged" || pledge.QuantityPledged != 300 {
		t.Fatalf("expected pledged record, got %#v", pledge)
	}

	requestResponse := httptest.NewRecorder()
	requestRequest := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/request_001", nil)
	requestRequest.SetPathValue("id", "request_001")
	srv.getAidRequestHandler(requestResponse, requestRequest)
	if requestResponse.Code != http.StatusOK {
		t.Fatalf("expected request found, got %d: %s", requestResponse.Code, requestResponse.Body.String())
	}
	var req aidRequestRecord
	decodeResponse(t, requestResponse, &req)
	if req.Status != "partially_fulfilled" || req.QuantityFulfilled != 300 {
		t.Fatalf("expected partially fulfilled request, got %#v", req)
	}

	pledgeBody2 := createPledgeRequest{
		DonorID:         donor.ID,
		QuantityPledged: 200,
	}
	pledgeResponse2 := httptest.NewRecorder()
	pledgeRequest2 := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(pledgeBody2))
	pledgeRequest2.Header.Set("Content-Type", "application/json")
	pledgeRequest2.SetPathValue("id", "request_001")
	srv.createPledgeHandler(pledgeResponse2, pledgeRequest2)
	if pledgeResponse2.Code != http.StatusCreated {
		t.Fatalf("expected second pledge created, got %d: %s", pledgeResponse2.Code, pledgeResponse2.Body.String())
	}
	var pledge2 pledgeRecord
	decodeResponse(t, pledgeResponse2, &pledge2)

	requestResponse2 := httptest.NewRecorder()
	requestRequest2 := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/request_001", nil)
	requestRequest2.SetPathValue("id", "request_001")
	srv.getAidRequestHandler(requestResponse2, requestRequest2)
	if requestResponse2.Code != http.StatusOK {
		t.Fatalf("expected request found after second pledge, got %d: %s", requestResponse2.Code, requestResponse2.Body.String())
	}
	var reqAfterPledge aidRequestRecord
	decodeResponse(t, requestResponse2, &reqAfterPledge)
	if reqAfterPledge.Status != "fulfilled" || reqAfterPledge.QuantityFulfilled != 500 {
		t.Fatalf("expected fulfilled request after second pledge, got %#v", reqAfterPledge)
	}

	allocateResponse := httptest.NewRecorder()
	allocateRequest := authorityRequest(http.MethodPost, "/api/v1/aid-requests/request_001/allocate", jsonBody(allocateRequest{
		PledgeID: pledge.ID,
		Quantity: 300,
	}))
	allocateRequest.SetPathValue("id", "request_001")
	srv.allocatePledgeHandler(allocateResponse, allocateRequest)
	if allocateResponse.Code != http.StatusOK {
		t.Fatalf("expected allocate success, got %d: %s", allocateResponse.Code, allocateResponse.Body.String())
	}

	pledgeListRecorder := httptest.NewRecorder()
	pledgeListRequest := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/request_001/pledges", nil)
	pledgeListRequest.SetPathValue("id", "request_001")
	srv.listRequestPledgesHandler(pledgeListRecorder, pledgeListRequest)
	if pledgeListRecorder.Code != http.StatusOK {
		t.Fatalf("expected pledge list, got %d: %s", pledgeListRecorder.Code, pledgeListRecorder.Body.String())
	}
	var list pledgeListResponse
	decodeResponse(t, pledgeListRecorder, &list)
	found := false
	for _, p := range list.Pledges {
		if p.ID == pledge.ID && p.Status == "delivered" && p.QuantityDelivered == 300 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected allocated pledge to be delivered, got %#v", list.Pledges)
	}
}

func TestAllocateRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/allocate", jsonBody(allocateRequest{
		PledgeID: "pledge_001",
		Quantity: 10,
	}))
	request.SetPathValue("id", "request_001")

	srv.allocatePledgeHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListPledgesRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/pledges", nil)

	srv.listPledgesHandler(response, request)

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
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_donation_operator")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-donation-update")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
