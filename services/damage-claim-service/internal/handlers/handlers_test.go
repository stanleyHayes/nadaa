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

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/store"
)

const testAuthSecret = "test-damage-claim-auth-secret"

func newTestServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	return NewServer(
		store.NewMemoryStore(now),
		func() time.Time { return now },
		&config.Config{
			IncidentServiceURL: "http://127.0.0.1:1",
			AuthTokenSecret:    testAuthSecret,
			AllowMockActors:    true,
		},
	)
}

// newSecureTestServer disables mock-actor headers so only a valid bearer token
// authenticates authority requests.
func newSecureTestServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	return NewServer(
		store.NewMemoryStore(now),
		func() time.Time { return now },
		&config.Config{
			IncidentServiceURL: "http://127.0.0.1:1",
			AuthTokenSecret:    testAuthSecret,
			AllowMockActors:    false,
		},
	)
}

func newTestServerWithIncidentServer(incidentServer *httptest.Server) *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	srv := NewServer(
		store.NewMemoryStore(now),
		func() time.Time { return now },
		&config.Config{
			IncidentServiceURL: incidentServer.URL,
			AuthTokenSecret:    testAuthSecret,
			AllowMockActors:    true,
		},
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
				"lat": 5.6037,
				"lng": -0.187,
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
	if payload.IncidentLocation != "5.6037,-0.187" {
		t.Fatalf("expected incident location enriched from coordinates, got %s", payload.IncidentLocation)
	}
}

func TestCreateClaimEnrichmentToleratesMissingLocation(t *testing.T) {
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		utilsWriteJSON(w, http.StatusOK, map[string]any{
			"id":        "inc_test_003",
			"reference": "NADAA-TEST-20260707-002",
		})
	}))
	defer incidentServer.Close()

	srv := newTestServerWithIncidentServer(incidentServer)
	body := models.CreateClaimRequest{
		IncidentID:          "inc_test_003",
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "fire",
		DamageDescription:   "Fire damage.",
		EstimatedLossAmount: "5000.00",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18},
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
	if payload.IncidentReference != "NADAA-TEST-20260707-002" {
		t.Fatalf("expected incident reference enriched, got %s", payload.IncidentReference)
	}
	if payload.IncidentLocation != "" {
		t.Fatalf("expected empty incident location when absent, got %s", payload.IncidentLocation)
	}
}

func TestCreateClaimForwardsAuthorizationToIncidentService(t *testing.T) {
	var gotAuthorization string
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		utilsWriteJSON(w, http.StatusOK, map[string]any{"id": "inc_test_004", "reference": "NADAA-TEST-20260707-003"})
	}))
	defer incidentServer.Close()

	srv := newTestServerWithIncidentServer(incidentServer)
	body := models.CreateClaimRequest{
		IncidentID:          "inc_test_004",
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "fire",
		DamageDescription:   "Fire damage.",
		EstimatedLossAmount: "5000.00",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18},
		PrivacyConsent:      true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	request.Header.Set("Authorization", "Bearer citizen-token")
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if gotAuthorization != "Bearer citizen-token" {
		t.Fatalf("expected Authorization header forwarded to incident-service, got %q", gotAuthorization)
	}
	if request.Header.Get("X-NADAA-Actor-ID") != "" {
		t.Fatalf("test setup must not fabricate actor headers")
	}
}

func TestCreateClaimSendsServiceTokenWithoutCallerAuth(t *testing.T) {
	var gotAuthorization, gotServiceToken string
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		gotServiceToken = r.Header.Get("X-NADAA-Service-Token")
		utilsWriteJSON(w, http.StatusOK, map[string]any{"id": "inc_test_005", "reference": "NADAA-TEST-20260707-004"})
	}))
	defer incidentServer.Close()

	srv := newTestServerWithIncidentServer(incidentServer)
	srv.config.InternalServiceToken = "test-internal-service-token"
	body := models.CreateClaimRequest{
		IncidentID:          "inc_test_005",
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "flood",
		DamageDescription:   "Flood damage.",
		EstimatedLossAmount: "2500.00",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18},
		PrivacyConsent:      true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if gotAuthorization != "" {
		t.Fatalf("expected no Authorization header without caller auth, got %q", gotAuthorization)
	}
	if gotServiceToken != "test-internal-service-token" {
		t.Fatalf("expected internal service token sent to incident-service, got %q", gotServiceToken)
	}

	var payload models.DamageClaimRecord
	decodeResponse(t, response, &payload)
	if payload.IncidentReference != "NADAA-TEST-20260707-004" {
		t.Fatalf("expected incident reference enriched via service token, got %s", payload.IncidentReference)
	}
}

func TestCreateClaimPrefersAuthorizationOverServiceToken(t *testing.T) {
	var gotAuthorization, gotServiceToken string
	incidentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthorization = r.Header.Get("Authorization")
		gotServiceToken = r.Header.Get("X-NADAA-Service-Token")
		utilsWriteJSON(w, http.StatusOK, map[string]any{"id": "inc_test_006", "reference": "NADAA-TEST-20260707-005"})
	}))
	defer incidentServer.Close()

	srv := newTestServerWithIncidentServer(incidentServer)
	srv.config.InternalServiceToken = "test-internal-service-token"
	body := models.CreateClaimRequest{
		IncidentID:          "inc_test_006",
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "flood",
		DamageDescription:   "Flood damage.",
		EstimatedLossAmount: "2500.00",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18},
		PrivacyConsent:      true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	request.Header.Set("Authorization", "Bearer dispatcher-token")
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if gotAuthorization != "Bearer dispatcher-token" {
		t.Fatalf("expected caller Authorization forwarded to incident-service, got %q", gotAuthorization)
	}
	if gotServiceToken != "" {
		t.Fatalf("expected service token omitted when caller auth is present, got %q", gotServiceToken)
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

func TestVerifyClaimRejectsPendingStatus(t *testing.T) {
	srv := newTestServer()
	body := models.VerifyClaimRequest{VerificationStatus: "pending", Notes: "Still pending."}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/claims/claim_001/verify", jsonBody(body))
	request.SetPathValue("id", "claim_001")
	srv.verifyClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	claim, ok := srv.store.Get("claim_001")
	if !ok {
		t.Fatalf("expected claim_001 to exist")
	}
	if claim.VerifiedBy != "" || claim.VerifiedAt != nil {
		t.Fatalf("expected no verification metadata stamped for pending status, got %#v", claim)
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

func TestCloseClaimAlreadyClosed(t *testing.T) {
	srv := newTestServer()
	body := models.CloseClaimRequest{Reason: "Duplicate submission."}

	first := httptest.NewRecorder()
	firstRequest := authorityRequest(http.MethodPost, "/claims/claim_001/close", jsonBody(body))
	firstRequest.SetPathValue("id", "claim_001")
	srv.closeClaimHandler(first, firstRequest)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first close status %d, got %d: %s", http.StatusOK, first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	secondRequest := authorityRequest(http.MethodPost, "/claims/claim_001/close", jsonBody(body))
	secondRequest.SetPathValue("id", "claim_001")
	srv.closeClaimHandler(second, secondRequest)

	if second.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, second.Code, second.Body.String())
	}

	claim, ok := srv.store.Get("claim_001")
	if !ok {
		t.Fatalf("expected claim_001 to exist")
	}
	if strings.Count(claim.VerificationNotes, "Closed by") != 1 {
		t.Fatalf("expected exactly one closure note, got %q", claim.VerificationNotes)
	}
}

func TestWriteClaimCSVEscapesFormulaCells(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	claim := models.DamageClaimRecord{
		ID:                  "claim_xss",
		Reference:           "DC-2026-00009",
		Reporter:            models.ReporterInfo{Name: "=HYPERLINK(\"http://evil\")", Phone: "+233200000000", Email: "@evil.example.com"},
		DamageType:          "flood",
		DamageDescription:   "-10+cmd|' /C calc'!A0",
		EstimatedLossAmount: "100.00",
		VerificationStatus:  "pending",
		Status:              "submitted",
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18, Address: "=1+1"},
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	response := httptest.NewRecorder()
	writeClaimCSV(response, claim)

	reader := csv.NewReader(response.Body)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}

	values := map[string]string{}
	for _, record := range records {
		if len(record) == 2 {
			values[record[0]] = record[1]
		}
	}

	escaped := map[string]string{
		"ReporterName":      "'=HYPERLINK(\"http://evil\")",
		"ReporterEmail":     "'@evil.example.com",
		"DamageDescription": "'-10+cmd|' /C calc'!A0",
		"LocationAddress":   "'=1+1",
	}
	for field, want := range escaped {
		if values[field] != want {
			t.Fatalf("expected %s cell %q, got %q", field, want, values[field])
		}
	}
	if values["EstimatedLossAmount"] != "100.00" {
		t.Fatalf("expected plain value untouched, got %q", values["EstimatedLossAmount"])
	}
}

func TestCreateClaimRejectsTooManyPhotos(t *testing.T) {
	srv := newTestServer()
	photos := make([]string, 21)
	for i := range photos {
		photos[i] = "https://media.nadaa.local/photo.jpg"
	}
	body := models.CreateClaimRequest{
		Reporter:            models.ReporterInfo{Name: "Test Reporter", Phone: "+233200000000"},
		DamageType:          "flood",
		DamageDescription:   "Test damage description.",
		EstimatedLossAmount: "1000.00",
		DamagePhotos:        photos,
		Location:            models.ClaimLocation{Lat: 5.6, Lng: -0.18},
		PrivacyConsent:      true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", jsonBody(body))
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "too_many_damage_photos") {
		t.Fatalf("expected too_many_damage_photos error, got %s", response.Body.String())
	}
}

func TestCreateClaimRejectsOversizedBody(t *testing.T) {
	srv := newTestServer()
	// A syntactically valid JSON body larger than the 1 MiB cap.
	body := bytes.NewReader([]byte(`{"reporter":{"name":"` + strings.Repeat("a", 1<<20) + `"}}`))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/claims", body)
	srv.createClaimHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "invalid_json") {
		t.Fatalf("expected invalid_json error for oversized body, got %s", response.Body.String())
	}
}

func TestListClaimsWithBearerToken(t *testing.T) {
	srv := newSecureTestServer()
	response := httptest.NewRecorder()
	request := tokenAuthorityRequest(t, http.MethodGet, "/claims", nil)
	srv.listClaimsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestVerifyClaimWithBearerTokenUsesClaimsActor(t *testing.T) {
	srv := newSecureTestServer()
	body := models.VerifyClaimRequest{VerificationStatus: "verified", Notes: "Verified via token."}

	response := httptest.NewRecorder()
	request := tokenAuthorityRequest(t, http.MethodPost, "/claims/claim_001/verify", jsonBody(body))
	request.SetPathValue("id", "claim_001")
	srv.verifyClaimHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DamageClaimRecord
	decodeResponse(t, response, &payload)
	if payload.VerifiedBy != "usr_token_officer" {
		t.Fatalf("expected verifiedBy from token claims, got %s", payload.VerifiedBy)
	}
}

func TestAuthorityRejectsLegacyHeadersWhenMockActorsDisabled(t *testing.T) {
	srv := newSecureTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/claims", nil)
	srv.listClaimsHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityRejectsForgedToken(t *testing.T) {
	srv := newSecureTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/claims", nil)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, "wrong-secret", testAuthorityClaims()))
	srv.listClaimsHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityRejectsExpiredToken(t *testing.T) {
	srv := newSecureTestServer()
	claims := testAuthorityClaims()
	claims.ExpiresAt = time.Date(2026, 7, 7, 11, 0, 0, 0, time.UTC).Unix()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/claims", nil)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, testAuthSecret, claims))
	srv.listClaimsHandler(response, request)

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

// testAuthorityClaims mirrors the claims auth-service signs for an agency
// authority user. Expiry is anchored to the test server's fixed clock.
func testAuthorityClaims() models.TokenClaims {
	return models.TokenClaims{
		UserID:    "usr_token_officer",
		UserType:  "agency",
		Role:      "insurance_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000204",
		MFA:       true,
		ExpiresAt: time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC).Unix(),
	}
}

// signTestToken signs claims with the same nadaa.<payload>.<sig> scheme as
// auth-service; tests may use any secret since they construct the server.
func signTestToken(t *testing.T, secret string, claims models.TokenClaims) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func tokenAuthorityRequest(t *testing.T, method string, target string, body *bytes.Reader) *http.Request {
	t.Helper()
	if body == nil {
		body = bytes.NewReader([]byte{})
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, testAuthSecret, testAuthorityClaims()))
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func utilsWriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
