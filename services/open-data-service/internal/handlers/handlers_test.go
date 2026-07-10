package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/store"
)

func newTestServer() *Server {
	now := time.Date(2026, 7, 10, 8, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		Addr:                   ":8102",
		AuditLogServiceURL:     "",
		AllowedOrigins:         nil,
		RateLimitRequests:      10,
		RateLimitWindowSeconds: 60,
	}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
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

func TestHealthHandler(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestListDatasetsReturnsDatasetsWithPrivacyStatus(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DatasetListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Datasets) == 0 {
		t.Fatalf("expected datasets, got none")
	}

	approved := 0
	for _, dataset := range payload.Datasets {
		if dataset.PrivacyReviewStatus == "" {
			t.Fatalf("expected privacy review status on every dataset")
		}
		if dataset.PrivacyReviewStatus == models.PrivacyReviewApproved {
			approved++
		}
	}
	if approved == 0 {
		t.Fatalf("expected at least one approved dataset")
	}
}

func TestListDatasetsWithCategoryFilter(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets?category=flood", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.DatasetListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Datasets) != 1 || payload.Datasets[0].Category != models.OpenDataCategoryFlood {
		t.Fatalf("expected one flood dataset, got %#v", payload.Datasets)
	}
}

func TestGetDatasetReturnsDetail(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DatasetDetailResponse
	decodeResponse(t, response, &payload)
	if payload.Dataset.ID != "dataset_flood_reports_2026" {
		t.Fatalf("unexpected dataset id %s", payload.Dataset.ID)
	}
}

func TestGetDatasetNotFound(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/missing", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestDownloadApprovedDataset(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download?format=json", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DatasetDownloadResponse
	decodeResponse(t, response, &payload)
	if payload.Download.DatasetID != "dataset_flood_reports_2026" {
		t.Fatalf("unexpected download dataset id %s", payload.Download.DatasetID)
	}
	if payload.Download.Format != "json" {
		t.Fatalf("expected json format, got %s", payload.Download.Format)
	}
	if payload.RateLimit.Remaining != 9 {
		t.Fatalf("expected remaining 9, got %d", payload.RateLimit.Remaining)
	}
	if !payload.AuditLogged {
		t.Fatalf("expected audit logged")
	}
}

func TestDownloadPendingDatasetForbidden(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_raw_incident_feed/download", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestDownloadRateLimit(t *testing.T) {
	srv := newTestServer()
	for i := 0; i < 11; i++ {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download", nil)
		srv.Routes().ServeHTTP(response, request)
		if i < 10 && response.Code != http.StatusOK {
			t.Fatalf("expected ok on request %d, got %d", i, response.Code)
		}
		if i == 10 && response.Code != http.StatusTooManyRequests {
			t.Fatalf("expected rate limit on request %d, got %d", i, response.Code)
		}
	}
}

func TestCreateRequest(t *testing.T) {
	srv := newTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID: "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{
			Name:         "Ama Kwame",
			Organization: "University of Ghana",
			Email:        "ama.kwame@example.edu.gh",
			UseCase:      "academic research",
		},
		Purpose: "Research on flood response patterns in Accra for academic publication.",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.OpenDataRequestResponse
	decodeResponse(t, response, &payload)
	if payload.Request.Status != models.OpenDataRequestPending {
		t.Fatalf("expected pending status, got %s", payload.Request.Status)
	}
}

func TestCreateRequestInvalidEmail(t *testing.T) {
	srv := newTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID: "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{
			Name:    "Ama Kwame",
			Email:   "not-an-email",
			UseCase: "academic research",
		},
		Purpose: "Research on flood response patterns.",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestListRequestsAdminRequired(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)

	srv.Routes().ServeHTTP(response, request)

	// Missing authority context is unauthenticated, not merely forbidden.
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListRequestsAdmin(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
	setAdminHeaders(request)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.OpenDataRequestListResponse
	decodeResponse(t, response, &payload)
	if payload.Requests == nil {
		t.Fatalf("expected requests array")
	}
}

func setAdminHeaders(r *http.Request) {
	r.Header.Set("X-NADAA-Actor-ID", "usr_admin")
	r.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	r.Header.Set("X-NADAA-Actor-Role", "system_admin")
	r.Header.Set("X-NADAA-MFA-Completed", "true")
}

func TestAdminEndpointsRejectMissingAuthority(t *testing.T) {
	srv := newTestServer()

	// No authority headers at all -> 401.
	missing := httptest.NewRecorder()
	srv.Routes().ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil))
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without authority context, got %d", missing.Code)
	}

	// Role present but MFA not completed -> 403.
	noMFA := httptest.NewRecorder()
	noMFARequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
	noMFARequest.Header.Set("X-NADAA-Actor-ID", "usr_admin")
	noMFARequest.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	noMFARequest.Header.Set("X-NADAA-Actor-Role", "system_admin")
	srv.Routes().ServeHTTP(noMFA, noMFARequest)
	if noMFA.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without MFA, got %d", noMFA.Code)
	}

	// Full context but non-admin role -> 403.
	badRole := httptest.NewRecorder()
	badRoleRequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
	setAdminHeaders(badRoleRequest)
	badRoleRequest.Header.Set("X-NADAA-Actor-Role", "citizen")
	srv.Routes().ServeHTTP(badRole, badRoleRequest)
	if badRole.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin role, got %d", badRole.Code)
	}
}

func TestApproveRequest(t *testing.T) {
	srv := newTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID: "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{
			Name:    "Ama Kwame",
			Email:   "ama.kwame@example.edu.gh",
			UseCase: "academic research",
		},
		Purpose: "Research on flood response patterns.",
	}

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
	createRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createRequest)

	var created models.OpenDataRequestResponse
	decodeResponse(t, createResponse, &created)

	review := models.ReviewOpenDataRequest{Reviewer: "admin@nadaa.gov.gh", Approved: true, Note: "Approved for academic use."}
	approveResponse := httptest.NewRecorder()
	approveRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests/"+created.Request.ID+"/approve", jsonBody(review))
	approveRequest.Header.Set("Content-Type", "application/json")
	setAdminHeaders(approveRequest)
	srv.Routes().ServeHTTP(approveResponse, approveRequest)

	if approveResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, approveResponse.Code, approveResponse.Body.String())
	}

	var approved models.OpenDataRequestResponse
	decodeResponse(t, approveResponse, &approved)
	if approved.Request.Status != models.OpenDataRequestApproved {
		t.Fatalf("expected approved status, got %s", approved.Request.Status)
	}
}

func TestCreateRequestAssignsUniqueIDs(t *testing.T) {
	srv := newTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID:     "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{Name: "Ama Kwame", Email: "ama@example.edu.gh", UseCase: "research"},
		Purpose:       "Research on flood response patterns in Accra.",
	}
	ids := map[string]bool{}
	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
		req.Header.Set("Content-Type", "application/json")
		srv.Routes().ServeHTTP(rec, req)
		var payload models.OpenDataRequestResponse
		decodeResponse(t, rec, &payload)
		if payload.Request.ID == "" || ids[payload.Request.ID] {
			t.Fatalf("expected unique non-empty request id, got %q (seen=%v)", payload.Request.ID, ids)
		}
		ids[payload.Request.ID] = true
	}
}
