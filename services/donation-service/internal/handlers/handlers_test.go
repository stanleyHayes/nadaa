package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/store"
)

const testTokenSecret = "test-donation-token-secret"

func newTestServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		Addr:            ":8100",
		AllowedOrigins:  nil,
		AuthTokenSecret: testTokenSecret,
		AllowMockActors: true,
	}
	return NewServer(store.NewMemoryStore(now), models.SandboxPaymentProvider{CreditPayments: true}, func() time.Time { return now }, cfg)
}

func TestHealthz(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestListAidCatalog(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-catalog", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.AidCatalogResponse
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

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.AidRequestListResponse
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

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.AidRequestListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Requests) != 1 || payload.Requests[0].Category != "medical" {
		t.Fatalf("expected one medical request, got %#v", payload.Requests)
	}
}

func TestCreateDonorPublic(t *testing.T) {
	srv := newTestServer()
	body := models.CreateDonorRequest{
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

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.Donor
	decodeResponse(t, response, &payload)
	if payload.Name != body.Name || payload.Type != "ngo" || payload.Status != "active" || payload.CreatedBy != "public" {
		t.Fatalf("expected created donor, got %#v", payload)
	}
}

func TestCreateAidRequestRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	body := models.CreateAidRequestRequest{
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

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateAndUpdateAidRequestAuthority(t *testing.T) {
	srv := newTestServer()
	body := models.CreateAidRequestRequest{
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

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var created models.AidRequest
	decodeResponse(t, response, &created)
	if created.Status != "open" || created.RequestedBy != "usr_donation_operator" {
		t.Fatalf("expected created aid request, got %#v", created)
	}

	update := models.UpdateAidRequestRequest{Status: "closed"}
	updateResponse := httptest.NewRecorder()
	updateRequest := authorityRequest(http.MethodPatch, "/api/v1/aid-requests/"+created.ID, jsonBody(update))

	srv.Routes().ServeHTTP(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	var updated models.AidRequest
	decodeResponse(t, updateResponse, &updated)
	if updated.Status != "closed" {
		t.Fatalf("expected closed status, got %#v", updated)
	}
}

func TestUpdateDonorRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/donors/donor_001", jsonBody(models.UpdateDonorRequest{Status: "inactive"}))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateAndUpdateDonorAuthority(t *testing.T) {
	srv := newTestServer()
	body := models.CreateDonorRequest{
		Name:         "Govt Relief Fund",
		Type:         "government",
		ContactName:  "Kofi Asante",
		ContactEmail: "kofi@example.gov",
		Region:       "Greater Accra",
		District:     "Accra Metropolitan",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/donors", jsonBody(body))

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var created models.Donor
	decodeResponse(t, response, &created)
	if created.CreatedBy != "usr_donation_operator" {
		t.Fatalf("expected authority createdBy, got %#v", created)
	}

	update := models.UpdateDonorRequest{Status: "inactive", Notes: "Paused for review"}
	updateResponse := httptest.NewRecorder()
	updateRequest := authorityRequest(http.MethodPatch, "/api/v1/donors/"+created.ID, jsonBody(update))

	srv.Routes().ServeHTTP(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}

	var updated models.Donor
	decodeResponse(t, updateResponse, &updated)
	if updated.Status != "inactive" || updated.Notes != "Paused for review" {
		t.Fatalf("expected updated donor, got %#v", updated)
	}
}

func TestPledgeCreatesAndUpdatesRequestStatus(t *testing.T) {
	srv := newTestServer()

	donorBody := models.CreateDonorRequest{
		Name:         "Tema Relief Group",
		Type:         "organization",
		ContactEmail: "tema@example.com",
		Region:       "Greater Accra",
		District:     "Tema Metropolitan",
	}
	donorResponse := httptest.NewRecorder()
	donorRequest := httptest.NewRequest(http.MethodPost, "/api/v1/donors", jsonBody(donorBody))
	donorRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(donorResponse, donorRequest)
	if donorResponse.Code != http.StatusCreated {
		t.Fatalf("expected donor created, got %d: %s", donorResponse.Code, donorResponse.Body.String())
	}
	var donor models.Donor
	decodeResponse(t, donorResponse, &donor)

	pledgeBody := models.CreatePledgeRequest{
		DonorID:         donor.ID,
		QuantityPledged: 300,
		ContactEmail:    "TEMA@example.com", // matches donor email case-insensitively
	}
	pledgeResponse := httptest.NewRecorder()
	pledgeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(pledgeBody))
	pledgeRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(pledgeResponse, pledgeRequest)
	if pledgeResponse.Code != http.StatusCreated {
		t.Fatalf("expected pledge created, got %d: %s", pledgeResponse.Code, pledgeResponse.Body.String())
	}
	var pledge models.Pledge
	decodeResponse(t, pledgeResponse, &pledge)
	if pledge.Status != "pledged" || pledge.QuantityPledged != 300 {
		t.Fatalf("expected pledged record, got %#v", pledge)
	}

	requestResponse := httptest.NewRecorder()
	requestRequest := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/request_001", nil)
	srv.Routes().ServeHTTP(requestResponse, requestRequest)
	if requestResponse.Code != http.StatusOK {
		t.Fatalf("expected request found, got %d: %s", requestResponse.Code, requestResponse.Body.String())
	}
	var req models.AidRequest
	decodeResponse(t, requestResponse, &req)
	if req.Status != "partially_fulfilled" || req.QuantityFulfilled != 300 {
		t.Fatalf("expected partially fulfilled request, got %#v", req)
	}

	pledgeBody2 := models.CreatePledgeRequest{
		DonorID:         donor.ID,
		QuantityPledged: 200,
		ContactEmail:    "tema@example.com",
	}
	pledgeResponse2 := httptest.NewRecorder()
	pledgeRequest2 := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(pledgeBody2))
	pledgeRequest2.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(pledgeResponse2, pledgeRequest2)
	if pledgeResponse2.Code != http.StatusCreated {
		t.Fatalf("expected second pledge created, got %d: %s", pledgeResponse2.Code, pledgeResponse2.Body.String())
	}
	var pledge2 models.Pledge
	decodeResponse(t, pledgeResponse2, &pledge2)

	requestResponse2 := httptest.NewRecorder()
	requestRequest2 := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/request_001", nil)
	srv.Routes().ServeHTTP(requestResponse2, requestRequest2)
	if requestResponse2.Code != http.StatusOK {
		t.Fatalf("expected request found after second pledge, got %d: %s", requestResponse2.Code, requestResponse2.Body.String())
	}
	var reqAfterPledge models.AidRequest
	decodeResponse(t, requestResponse2, &reqAfterPledge)
	if reqAfterPledge.Status != "fulfilled" || reqAfterPledge.QuantityFulfilled != 500 {
		t.Fatalf("expected fulfilled request after second pledge, got %#v", reqAfterPledge)
	}

	allocateResponse := httptest.NewRecorder()
	allocateRequest := authorityRequest(http.MethodPost, "/api/v1/aid-requests/request_001/allocate", jsonBody(models.AllocateRequest{
		PledgeID: pledge.ID,
		Quantity: 300,
	}))
	srv.Routes().ServeHTTP(allocateResponse, allocateRequest)
	if allocateResponse.Code != http.StatusOK {
		t.Fatalf("expected allocate success, got %d: %s", allocateResponse.Code, allocateResponse.Body.String())
	}

	pledgeListRecorder := httptest.NewRecorder()
	pledgeListRequest := authorityRequest(http.MethodGet, "/api/v1/aid-requests/request_001/pledges", nil)
	srv.Routes().ServeHTTP(pledgeListRecorder, pledgeListRequest)
	if pledgeListRecorder.Code != http.StatusOK {
		t.Fatalf("expected pledge list, got %d: %s", pledgeListRecorder.Code, pledgeListRecorder.Body.String())
	}
	var list models.PledgeListResponse
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
	request := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/allocate", jsonBody(models.AllocateRequest{
		PledgeID: "pledge_001",
		Quantity: 10,
	}))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListPledgesRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/pledges", nil)

	srv.Routes().ServeHTTP(response, request)

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
	var reader io.Reader
	if body != nil {
		reader = body
	}
	request := httptest.NewRequest(method, target, reader)
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

func signedAuthorityToken(t *testing.T, secret string, expiresAt time.Time) string {
	t.Helper()
	claims := map[string]any{
		"sub":      "usr_donation_operator",
		"typ":      "agency",
		"role":     "district_officer",
		"agencyId": "00000000-0000-0000-0000-000000000204",
		"district": "Accra Metropolitan",
		"mfa":      true,
		"exp":      expiresAt.Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func tokenRequest(t *testing.T, method, target string, body *bytes.Reader) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = body
	}
	request := httptest.NewRequest(method, target, reader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+signedAuthorityToken(t, testTokenSecret, time.Now().Add(time.Hour)))
	return request
}

func TestAuthorityAcceptsValidBearerToken(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := tokenRequest(t, http.MethodGet, "/api/v1/pledges", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d for a valid bearer token, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestAuthorityRejectsInvalidBearerToken(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/pledges", nil)
	request.Header.Set("Authorization", "Bearer "+signedAuthorityToken(t, "wrong-secret", time.Now().Add(time.Hour)))

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for a wrongly-signed token, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityRejectsExpiredBearerToken(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/pledges", nil)
	expired := time.Date(2026, 7, 7, 11, 0, 0, 0, time.UTC) // before the server's fixed now
	request.Header.Set("Authorization", "Bearer "+signedAuthorityToken(t, testTokenSecret, expired))

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for an expired token, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityRejectsForgedHeadersWhenMockActorsDisabled(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8100", AuthTokenSecret: testTokenSecret, AllowMockActors: false}
	srv := NewServer(store.NewMemoryStore(now), models.SandboxPaymentProvider{CreditPayments: true}, func() time.Time { return now }, cfg)

	response := httptest.NewRecorder()
	srv.Routes().ServeHTTP(response, authorityRequest(http.MethodGet, "/api/v1/pledges", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for forged headers with mock actors off, got %d", http.StatusUnauthorized, response.Code)
	}

	tokenResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(tokenResponse, tokenRequest(t, http.MethodGet, "/api/v1/pledges", nil))
	if tokenResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d for a valid token with mock actors off, got %d: %s", http.StatusOK, tokenResponse.Code, tokenResponse.Body.String())
	}
}

func TestGetDonationRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	body, _ := json.Marshal(models.CreateDonationRequest{DonorName: "Ama", Email: "ama@example.com", Amount: 50})
	createResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(createResponse, httptest.NewRequest(http.MethodPost, "/api/v1/donations", bytes.NewReader(body)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201: %s", createResponse.Code, createResponse.Body.String())
	}
	var created models.CreateDonationResponse
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	response := httptest.NewRecorder()
	srv.Routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/donations/"+created.Donation.Reference, nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for an unauthenticated donation lookup, got %d", http.StatusUnauthorized, response.Code)
	}

	authorized := httptest.NewRecorder()
	srv.Routes().ServeHTTP(authorized, tokenRequest(t, http.MethodGet, "/api/v1/donations/"+created.Donation.Reference, nil))
	if authorized.Code != http.StatusOK {
		t.Fatalf("expected status %d for an authority donation lookup, got %d: %s", http.StatusOK, authorized.Code, authorized.Body.String())
	}
}

func TestListRequestPledgesRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/aid-requests/request_001/pledges", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreatePledgeBindsDonorIdentity(t *testing.T) {
	srv := newTestServer()

	donorResponse := httptest.NewRecorder()
	donorRequest := httptest.NewRequest(http.MethodPost, "/api/v1/donors", jsonBody(models.CreateDonorRequest{
		Name:         "Kumasi Aid Circle",
		Type:         "organization",
		ContactEmail: "kumasi@example.com",
		Region:       "Ashanti",
		District:     "Kumasi Metropolitan",
	}))
	donorRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(donorResponse, donorRequest)
	if donorResponse.Code != http.StatusCreated {
		t.Fatalf("expected donor created, got %d: %s", donorResponse.Code, donorResponse.Body.String())
	}
	var donor models.Donor
	decodeResponse(t, donorResponse, &donor)

	// Unknown donorId is rejected.
	unknownResponse := httptest.NewRecorder()
	unknownRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(models.CreatePledgeRequest{
		DonorID:         "donor_999",
		QuantityPledged: 10,
		ContactEmail:    "kumasi@example.com",
	}))
	unknownRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(unknownResponse, unknownRequest)
	if unknownResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for an unknown donor, got %d: %s", http.StatusForbidden, unknownResponse.Code, unknownResponse.Body.String())
	}

	// A contactEmail that does not match the donor's registered email is rejected.
	mismatchResponse := httptest.NewRecorder()
	mismatchRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(models.CreatePledgeRequest{
		DonorID:         donor.ID,
		QuantityPledged: 10,
		ContactEmail:    "someone-else@example.com",
	}))
	mismatchRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(mismatchResponse, mismatchRequest)
	if mismatchResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for a mismatched contact email, got %d: %s", http.StatusForbidden, mismatchResponse.Code, mismatchResponse.Body.String())
	}

	// A case-insensitive email match succeeds.
	matchedResponse := httptest.NewRecorder()
	matchedRequest := httptest.NewRequest(http.MethodPost, "/api/v1/aid-requests/request_001/pledges", jsonBody(models.CreatePledgeRequest{
		DonorID:         donor.ID,
		QuantityPledged: 10,
		ContactEmail:    "KUMASI@example.com",
	}))
	matchedRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(matchedResponse, matchedRequest)
	if matchedResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d for a matching contact email, got %d: %s", http.StatusCreated, matchedResponse.Code, matchedResponse.Body.String())
	}
}

func TestUpdateDonorStatusPreservesNotes(t *testing.T) {
	srv := newTestServer()
	donorResponse := httptest.NewRecorder()
	donorRequest := authorityRequest(http.MethodPost, "/api/v1/donors", jsonBody(models.CreateDonorRequest{
		Name:  "Note Keeper",
		Type:  "individual",
		Notes: "verified via field visit",
	}))
	srv.Routes().ServeHTTP(donorResponse, donorRequest)
	if donorResponse.Code != http.StatusCreated {
		t.Fatalf("expected donor created, got %d: %s", donorResponse.Code, donorResponse.Body.String())
	}
	var donor models.Donor
	decodeResponse(t, donorResponse, &donor)

	updateResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(updateResponse, authorityRequest(http.MethodPatch, "/api/v1/donors/"+donor.ID, jsonBody(models.UpdateDonorRequest{Status: "inactive"})))
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}
	var updated models.Donor
	decodeResponse(t, updateResponse, &updated)
	if updated.Status != "inactive" {
		t.Fatalf("expected status inactive, got %#v", updated)
	}
	if updated.Notes != "verified via field visit" {
		t.Fatalf("expected notes preserved on a status-only update, got %q", updated.Notes)
	}
}
