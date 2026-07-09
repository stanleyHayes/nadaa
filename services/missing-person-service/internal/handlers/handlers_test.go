package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/store"
)

func newTestServer() *Server {
	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8101", AllowedOrigins: nil}
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

func createRequest() models.CreateMissingPersonRequest {
	lat := 5.55
	lng := -0.18
	return models.CreateMissingPersonRequest{
		PersonName:  "Smoke Test Person",
		Age:         17,
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
