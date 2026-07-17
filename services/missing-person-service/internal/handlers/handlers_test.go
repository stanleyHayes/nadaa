package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/store"
)

const testTokenSecret = "test-missing-person-token-secret"

func newTestServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8101", AllowedOrigins: nil, AuthTokenSecret: testTokenSecret, AllowMockActors: true}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
}

func newTokenOnlyServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8101", AllowedOrigins: nil, AuthTokenSecret: testTokenSecret}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
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

func TestPublicListExcludesPrivatePendingRecord(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/missing-persons", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.PublicMissingPersonListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Records) != 1 {
		t.Fatalf("expected one public record, got %#v", payload.Records)
	}
	if payload.Records[0].ID != "missing_001" {
		t.Fatalf("expected public seed record only, got %#v", payload.Records)
	}
}

func TestAuthorityListRequiresHeaders(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/authority/missing-persons", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateReviewAndCloseMissingPerson(t *testing.T) {
	srv := newTestServer()
	body := createRequest()

	createResponse := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", jsonBody(body))
	createReq.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createReq)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.MissingPerson
	decodeResponse(t, createResponse, &created)
	if created.Status != "pending_review" || created.PublicVisibility != "private" || created.Reporter.Phone == "" {
		t.Fatalf("expected private pending sensitive record, got %#v", created)
	}

	publicResponse := httptest.NewRecorder()
	publicReq := httptest.NewRequest(http.MethodGet, "/api/v1/missing-persons/"+created.ID, nil)
	srv.Routes().ServeHTTP(publicResponse, publicReq)
	if publicResponse.Code != http.StatusNotFound {
		t.Fatalf("expected unapproved public lookup hidden, got %d", publicResponse.Code)
	}

	reviewResponse := httptest.NewRecorder()
	reviewReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
		Decision:      "approve_public",
		PublicSummary: "Public smoke summary with hotline contact through 112.",
		ReviewNotes:   "Guardian consent verified.",
	}))
	srv.Routes().ServeHTTP(reviewResponse, reviewReq)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected review success, got %d: %s", reviewResponse.Code, reviewResponse.Body.String())
	}
	var reviewed models.MissingPerson
	decodeResponse(t, reviewResponse, &reviewed)
	if reviewed.ReviewStatus != "approved" || reviewed.PublicVisibility != "public" || reviewed.Status != "active" {
		t.Fatalf("expected approved public record, got %#v", reviewed)
	}

	closeResponse := httptest.NewRecorder()
	closeReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/close", jsonBody(models.CloseMissingPersonRequest{
		ClosureType:        "reunited",
		ClosureNotes:       "Family reunified at Osu Community Hall.",
		ReunitedWithFamily: true,
	}))
	srv.Routes().ServeHTTP(closeResponse, closeReq)
	if closeResponse.Code != http.StatusOK {
		t.Fatalf("expected close success, got %d: %s", closeResponse.Code, closeResponse.Body.String())
	}
	var closed models.MissingPerson
	decodeResponse(t, closeResponse, &closed)
	if closed.Status != "reunited" || closed.PublicVisibility != "private" || closed.ClosedBy != "usr_missing_operator" {
		t.Fatalf("expected private reunited record, got %#v", closed)
	}

	auditResponse := httptest.NewRecorder()
	auditReq := authorityRequest(http.MethodGet, "/api/v1/authority/missing-persons/"+created.ID+"/audit", nil)
	srv.Routes().ServeHTTP(auditResponse, auditReq)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("expected audit success, got %d: %s", auditResponse.Code, auditResponse.Body.String())
	}
	var audit models.MissingPersonAuditResponse
	decodeResponse(t, auditResponse, &audit)
	if len(audit.Entries) < 3 {
		t.Fatalf("expected create/review/close audit entries, got %#v", audit.Entries)
	}
}

func TestCreateRequiresContactConsent(t *testing.T) {
	srv := newTestServer()
	body := createRequest()
	body.Reporter.ConsentToContact = false

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestAuthorityListAcceptsValidBearerToken(t *testing.T) {
	srv := newTokenOnlyServer()
	token := signTestToken(t, authorityClaims())

	response := httptest.NewRecorder()
	request := bearerRequest(http.MethodGet, "/api/v1/authority/missing-persons", nil, token)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestAuthorityTokenActorComesFromVerifiedClaims(t *testing.T) {
	srv := newTokenOnlyServer()
	token := signTestToken(t, authorityClaims())

	createResponse := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", jsonBody(createRequest()))
	createReq.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createReq)
	var created models.MissingPerson
	decodeResponse(t, createResponse, &created)

	reviewResponse := httptest.NewRecorder()
	reviewReq := bearerRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
		Decision:      "approve_private",
		PublicSummary: "Private approval.",
		ReviewNotes:   "Reviewed with token identity.",
	}), token)
	// Forged headers contradicting the token must be ignored.
	reviewReq.Header.Set("X-NADAA-Actor-ID", "usr_forged")
	reviewReq.Header.Set("X-NADAA-Actor-Role", "system_admin")
	srv.Routes().ServeHTTP(reviewResponse, reviewReq)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected review success, got %d: %s", reviewResponse.Code, reviewResponse.Body.String())
	}
	var reviewed models.MissingPerson
	decodeResponse(t, reviewResponse, &reviewed)
	if reviewed.ReviewedBy != "usr_token_officer" {
		t.Fatalf("expected reviewer from token claims, got %q", reviewed.ReviewedBy)
	}
}

func TestAuthorityRejectsForgedOrExpiredToken(t *testing.T) {
	srv := newTokenOnlyServer()

	forged := signTestTokenWithSecret(t, authorityClaims(), "wrong-secret")
	forgedResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(forgedResponse, bearerRequest(http.MethodGet, "/api/v1/authority/missing-persons", nil, forged))
	if forgedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected forged token rejected with %d, got %d", http.StatusUnauthorized, forgedResponse.Code)
	}

	expiredClaims := authorityClaims()
	expiredClaims.ExpiresAt = time.Date(2026, 7, 7, 11, 0, 0, 0, time.UTC).Unix()
	expired := signTestToken(t, expiredClaims)
	expiredResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(expiredResponse, bearerRequest(http.MethodGet, "/api/v1/authority/missing-persons", nil, expired))
	if expiredResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected expired token rejected with %d, got %d", http.StatusUnauthorized, expiredResponse.Code)
	}
}

func TestAuthorityHeadersRejectedWhenMockActorsDisabled(t *testing.T) {
	srv := newTokenOnlyServer()

	response := httptest.NewRecorder()
	srv.Routes().ServeHTTP(response, authorityRequest(http.MethodGet, "/api/v1/authority/missing-persons", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestPublicDetailHidesClosedRecord(t *testing.T) {
	srv := newTestServer()
	created := createRecord(t, srv)

	publicResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(publicResponse, httptest.NewRequest(http.MethodGet, "/api/v1/missing-persons/"+created.ID, nil))
	if publicResponse.Code != http.StatusOK {
		t.Fatalf("expected approved record publicly visible, got %d", publicResponse.Code)
	}

	closeResponse := httptest.NewRecorder()
	closeReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/close", jsonBody(models.CloseMissingPersonRequest{
		ClosureType:  "withdrawn",
		ClosureNotes: "Family withdrew the report.",
	}))
	srv.Routes().ServeHTTP(closeResponse, closeReq)
	if closeResponse.Code != http.StatusOK {
		t.Fatalf("expected close success, got %d: %s", closeResponse.Code, closeResponse.Body.String())
	}

	hiddenResponse := httptest.NewRecorder()
	srv.Routes().ServeHTTP(hiddenResponse, httptest.NewRequest(http.MethodGet, "/api/v1/missing-persons/"+created.ID, nil))
	if hiddenResponse.Code != http.StatusNotFound {
		t.Fatalf("expected closed record hidden from public detail, got %d", hiddenResponse.Code)
	}
}

func TestReviewRejectsContradictoryStatus(t *testing.T) {
	srv := newTestServer()
	created := createRecord(t, srv)

	for _, decision := range []string{"approve_public", "approve_private"} {
		response := httptest.NewRecorder()
		request := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
			Decision:      decision,
			PublicSummary: "Public smoke summary with hotline contact through 112.",
			ReviewNotes:   "Attempting contradictory status.",
			Status:        "rejected",
		}))
		srv.Routes().ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected %s with status=rejected rejected with %d, got %d", decision, http.StatusBadRequest, response.Code)
		}
	}
}

func TestApprovePublicRequiresConsentOrOverride(t *testing.T) {
	srv := newTestServer()
	body := createRequest()
	body.Reporter.ConsentToPublicShare = false

	createResponse := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", jsonBody(body))
	createReq.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createReq)
	var created models.MissingPerson
	decodeResponse(t, createResponse, &created)

	deniedResponse := httptest.NewRecorder()
	deniedReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
		Decision:      "approve_public",
		PublicSummary: "Public smoke summary with hotline contact through 112.",
		ReviewNotes:   "Attempting publication without consent.",
	}))
	srv.Routes().ServeHTTP(deniedResponse, deniedReq)
	if deniedResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected approve_public without consent rejected with %d, got %d", http.StatusBadRequest, deniedResponse.Code)
	}

	overrideResponse := httptest.NewRecorder()
	overrideReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
		Decision:        "approve_public",
		PublicSummary:   "Public smoke summary with hotline contact through 112.",
		ReviewNotes:     "Police request publication for immediate safety.",
		ConsentOverride: true,
	}))
	srv.Routes().ServeHTTP(overrideResponse, overrideReq)
	if overrideResponse.Code != http.StatusOK {
		t.Fatalf("expected consentOverride approval success, got %d: %s", overrideResponse.Code, overrideResponse.Body.String())
	}

	auditResponse := httptest.NewRecorder()
	auditReq := authorityRequest(http.MethodGet, "/api/v1/authority/missing-persons/"+created.ID+"/audit", nil)
	srv.Routes().ServeHTTP(auditResponse, auditReq)
	var audit models.MissingPersonAuditResponse
	decodeResponse(t, auditResponse, &audit)
	overrideRecorded := false
	for _, entry := range audit.Entries {
		if entry.Action == "missing_person.reviewed" && strings.Contains(entry.Notes, "consentOverride=true") {
			overrideRecorded = true
		}
	}
	if !overrideRecorded {
		t.Fatalf("expected consentOverride recorded in audit trail, got %#v", audit.Entries)
	}
}

func TestReReviewClearsClosureMetadata(t *testing.T) {
	srv := newTestServer()
	created := createRecord(t, srv)

	closeResponse := httptest.NewRecorder()
	closeReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/close", jsonBody(models.CloseMissingPersonRequest{
		ClosureType:  "duplicate",
		ClosureNotes: "Duplicate of missing_001.",
	}))
	srv.Routes().ServeHTTP(closeResponse, closeReq)
	if closeResponse.Code != http.StatusOK {
		t.Fatalf("expected close success, got %d: %s", closeResponse.Code, closeResponse.Body.String())
	}

	reviewResponse := httptest.NewRecorder()
	reviewReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
		Decision:      "approve_public",
		PublicSummary: "Reopened after duplicate determination was reversed.",
		ReviewNotes:   "Reopened the case.",
	}))
	srv.Routes().ServeHTTP(reviewResponse, reviewReq)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected re-review success, got %d: %s", reviewResponse.Code, reviewResponse.Body.String())
	}
	var reviewed models.MissingPerson
	decodeResponse(t, reviewResponse, &reviewed)
	if reviewed.Status != "active" || reviewed.ClosureType != "" || reviewed.ClosureNotes != "" || reviewed.ClosedBy != "" || reviewed.ClosedAt != nil {
		t.Fatalf("expected re-reviewed record free of closure metadata, got %#v", reviewed)
	}
}

func TestAgeZeroSerializes(t *testing.T) {
	srv := newTestServer()
	body := createRequest()
	infant := 0
	body.Age = &infant

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"age":0`) {
		t.Fatalf("expected age 0 serialized for infant, got %s", response.Body.String())
	}
}

func TestPublicIntakeRejectsOversizedBody(t *testing.T) {
	srv := newTestServer()
	oversized := bytes.NewReader(append([]byte(`{"personName":"`), bytes.Repeat([]byte("a"), 2<<20)...))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", oversized)
	request.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCORSLocalhostEchoRequiresDevelopment(t *testing.T) {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		Addr:            ":8101",
		AllowedOrigins:  map[string]bool{"https://nadaa.gov.gh": true},
		AuthTokenSecret: testTokenSecret,
		AllowMockActors: true,
	}
	srv := NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)

	t.Setenv("NADAA_ENV", "production")
	prodResponse := httptest.NewRecorder()
	prodRequest := httptest.NewRequest(http.MethodGet, "/api/v1/missing-persons", nil)
	prodRequest.Header.Set("Origin", "http://localhost:3000")
	srv.Routes().ServeHTTP(prodResponse, prodRequest)
	if got := prodResponse.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected localhost origin not echoed outside development, got %q", got)
	}

	t.Setenv("NADAA_ENV", "development")
	devResponse := httptest.NewRecorder()
	devRequest := httptest.NewRequest(http.MethodGet, "/api/v1/missing-persons", nil)
	devRequest.Header.Set("Origin", "http://localhost:3000")
	srv.Routes().ServeHTTP(devResponse, devRequest)
	if got := devResponse.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected localhost origin echoed in development, got %q", got)
	}
}

func createRequest() models.CreateMissingPersonRequest {
	lat := 5.55
	lng := -0.18
	age := 17
	return models.CreateMissingPersonRequest{
		PersonName:  "Smoke Test Person",
		Age:         &age,
		Gender:      "unknown",
		Description: "Last seen near the community hall after evacuation.",
		PhotoURL:    "https://example.test/smoke.jpg",
		LastSeenAt:  time.Date(2026, 7, 7, 8, 30, 0, 0, time.UTC),
		LastSeenLocation: models.LastSeenLocation{
			Label:    "Osu Community Hall",
			Region:   "Greater Accra",
			District: "Korle Klottey",
			Lat:      &lat,
			Lng:      &lng,
		},
		RelatedIncidentID: "inc_accra_flood_0241",
		Reporter: models.ReporterContact{
			Name:                 "Smoke Reporter",
			Phone:                "+233200000333",
			Email:                "reporter@example.com",
			Relationship:         "guardian",
			ConsentToContact:     true,
			ConsentToPublicShare: true,
		},
	}
}

func authorityRequest(method string, target string, body *bytes.Reader) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_missing_operator")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-missing-person")
	return request
}

func bearerRequest(method string, target string, body *bytes.Reader, token string) *http.Request {
	if body == nil {
		body = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, target, body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func authorityClaims() models.TokenClaims {
	return models.TokenClaims{
		UserID:    "usr_token_officer",
		UserType:  "agency",
		Role:      "district_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000204",
		District:  "Korle Klottey",
		MFA:       true,
		ExpiresAt: time.Date(2026, 7, 7, 13, 0, 0, 0, time.UTC).Unix(),
	}
}

func signTestToken(t *testing.T, claims models.TokenClaims) string {
	t.Helper()
	return signTestTokenWithSecret(t, claims, testTokenSecret)
}

// signTestTokenWithSecret mirrors auth-service's token signing scheme.
func signTestTokenWithSecret(t *testing.T, claims models.TokenClaims, secret string) string {
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

// createRecord creates and publicly approves a record, returning the created record.
func createRecord(t *testing.T, srv *Server) models.MissingPerson {
	t.Helper()

	createResponse := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/missing-persons", jsonBody(createRequest()))
	createReq.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createReq)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.MissingPerson
	decodeResponse(t, createResponse, &created)

	reviewResponse := httptest.NewRecorder()
	reviewReq := authorityRequest(http.MethodPatch, "/api/v1/authority/missing-persons/"+created.ID+"/review", jsonBody(models.ReviewMissingPersonRequest{
		Decision:      "approve_public",
		PublicSummary: "Public smoke summary with hotline contact through 112.",
		ReviewNotes:   "Guardian consent verified.",
	}))
	srv.Routes().ServeHTTP(reviewResponse, reviewReq)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected review success, got %d: %s", reviewResponse.Code, reviewResponse.Body.String())
	}
	return created
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
