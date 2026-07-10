package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/store"
)

func newTestServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	return NewServer(
		store.NewMemoryStore(now),
		func() time.Time { return now },
		&config.Config{IncidentServiceURL: "http://127.0.0.1:1"},
	)
}

func newTestServerWithIncidentServer(incidentServer *httptest.Server) *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	srv := NewServer(
		store.NewMemoryStore(now),
		func() time.Time { return now },
		&config.Config{IncidentServiceURL: incidentServer.URL},
	)
	srv.httpClient = incidentServer.Client()
	srv.incidentServiceURL = incidentServer.URL
	return srv
}

func TestCreateClaimRequiresPrivacyConsent(t *testing.T) {
	srv := newTestServer()
	body := models.CreateClaimRequest{
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "flood",
		DamageDescription:   "Test damage description.",
		EstimatedLossAmount: "1000.00",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18, Address: "Test Address"},
		PrivacyConsent:      false,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestCreateClaim(t *testing.T) {
	srv := newTestServer()
	body := models.CreateClaimRequest{
		IncidentID:          "inc_test_001",
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000", Email: "test@example.com", UserID: "usr_test"},
		DamageType:          "flood",
		DamageDescription:   "Test damage description.",
		EstimatedLossAmount: "1000.00",
		DamagePhotos:        []string{"https://media.nadaa.local/photo1.jpg"},
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18, Address: "Test Address"},
		PrivacyConsent:      true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.DamageClaimRecord
	decodeResponse(t, response, &payload)
	if payload.Reference == "" || payload.VerificationStatus != "pending" || payload.Status != "submitted" {
		t.Fatalf("expected created claim with pending verification and submitted status, got %#v", payload)
	}
	if payload.Reporter.Name != "Test Reporter" || payload.Reporter.Phone != "+233200000000" {
		t.Fatalf("expected reporter details, got %#v", payload.Reporter)
	}
	if payload.IncidentID != "inc_test_001" {
		t.Fatalf("expected incident id preserved, got %s", payload.IncidentID)
	}
}

func TestCreateClaimEnrichesIncidentReference(t *testing.T) {
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/incidents/inc_test_002" {
			http.NotFound(w, r)
			return
		}
		utilsWriteJSON(w, http.StatusOK, map[string]any{
			"id":        "inc_test_002",
			"reference": "NADAA-TEST-20260707-001",
			"location": map[string]any{
				"address": "Incident Address",
			},
		})
	}))
	defer incidentServer.Close()

	srv := newTestServerWithIncidentServer(incidentServer)
	body := models.CreateClaimRequest{
		IncidentID:          "inc_test_002",
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "fire",
		DamageDescription:   "Fire damage.",
		EstimatedLossAmount: "5000.00",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18, Address: "Test Address"},
		PrivacyConsent:      true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.DamageClaimRecord
	decodeResponse(t, response, &payload)
	if payload.IncidentReference != "NADAA-TEST-20260707-001" {
		t.Fatalf("expected incident reference enriched, got %s", payload.IncidentReference)
	}
	if payload.IncidentLocation != "Incident Address" {
		t.Fatalf("expected incident location enriched, got %s", payload.IncidentLocation)
	}
}

func TestListClaimsRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/claims", nil)
	srv.listClaimsHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListClaimsFilters(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/claims?status=submitted&verificationStatus=pending", nil)
	srv.listClaimsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ClaimListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Claims) == 0 {
		t.Fatalf("expected filtered claims, got %#v", payload)
	}
	for _, claim := range payload.Claims {
		if claim.Status != "submitted" || claim.VerificationStatus != "pending" {
			t.Fatalf("expected only submitted and pending claims, got %#v", claim)
		}
	}
}

func TestListClaimsQuery(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/claims?q=kwame", nil)
	srv.listClaimsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ClaimListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Claims) != 1 || payload.Claims[0].Reporter.Name != "Kwame Asare" {
		t.Fatalf("expected query to match Kwame Asare, got %#v", payload.Claims)
	}
}

func TestVerifyClaim(t *testing.T) {
	srv := newTestServer()
	body := models.VerifyClaimRequest{VerificationStatus: "verified", Notes: "Photos verified."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/claims/claim_001/verify", jsonBody(body))
	request.SetPathValue("id", "claim_001")
	srv.verifyClaimHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DamageClaimRecord
	decodeResponse(t, response, &payload)
	if payload.VerificationStatus != "verified" || payload.VerifiedBy != "usr_insurance_officer" || payload.VerificationNotes != "Photos verified." {
		t.Fatalf("expected verified claim, got %#v", payload)
	}
}

func TestVerifyClaimInvalidTransition(t *testing.T) {
	srv := newTestServer()
	body := models.VerifyClaimRequest{VerificationStatus: "verified", Notes: "Photos verified."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/claims/claim_002/verify", jsonBody(body))
	request.SetPathValue("id", "claim_002")
	srv.verifyClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestInvalidVerificationStatus(t *testing.T) {
	srv := newTestServer()
	body := models.VerifyClaimRequest{VerificationStatus: "approved", Notes: "Invalid status."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/claims/claim_001/verify", jsonBody(body))
	request.SetPathValue("id", "claim_001")
	srv.verifyClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestExportClaimCSV(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/claims/claim_001/export?format=csv", nil)
	request.SetPathValue("id", "claim_001")
	srv.exportClaimHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	contentType := response.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Fatalf("expected text/csv content type, got %s", contentType)
	}

	reader := csv.NewReader(response.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}
	if len(records) == 0 || records[0][0] != "Field" {
		t.Fatalf("expected CSV header starting with Field, got %#v", records)
	}

	foundReference := false
	for _, record := range records {
		if len(record) >= 2 && record[0] == "Reference" && strings.Contains(record[1], "DC-") {
			foundReference = true
			break
		}
	}
	if !foundReference {
		t.Fatalf("expected CSV to contain reference row, got %#v", records)
	}
}

func TestExportClaimPDF(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/claims/claim_001/export?format=pdf", nil)
	request.SetPathValue("id", "claim_001")
	srv.exportClaimHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	contentType := response.Header().Get("Content-Type")
	if contentType != "application/pdf" {
		t.Fatalf("expected application/pdf content type, got %s", contentType)
	}

	body := response.Body.String()
	if !strings.HasPrefix(body, "%PDF-1.4") {
		t.Fatalf("expected PDF header, got %s", body[:min(len(body), 20)])
	}
	if !strings.Contains(body, "Damage Claim Report") {
		t.Fatalf("expected PDF body to contain report title")
	}
}

func TestExportClaimInvalidFormat(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/claims/claim_001/export?format=xml", nil)
	request.SetPathValue("id", "claim_001")
	srv.exportClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCloseClaim(t *testing.T) {
	srv := newTestServer()
	body := models.CloseClaimRequest{Reason: "Duplicate submission."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/claims/claim_001/close", jsonBody(body))
	request.SetPathValue("id", "claim_001")
	srv.closeClaimHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DamageClaimRecord
	decodeResponse(t, response, &payload)
	if payload.Status != "closed" {
		t.Fatalf("expected closed status, got %s", payload.Status)
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
		body = bytes.NewReader([]byte{})
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_insurance_officer")
	request.Header.Set("X-NADAA-Actor-Role", "insurance_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-damage-claim")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func utilsWriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
