package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/open-data-service/internal/store"
)

// testNow is the fixed clock every test server runs on; token tests sign
// expirations relative to it.
var testNow = time.Date(2026, 7, 10, 8, 0, 0, 0, time.UTC)

const testTokenSecret = "test-secret-for-open-data-service-tests"

func newTestServer() *Server {
	cfg := &config.Config{
		Addr:                   ":8102",
		AuditLogServiceURL:     "",
		AllowedOrigins:         nil,
		RateLimitRequests:      10,
		RateLimitWindowSeconds: 60,
		// Legacy header path stays exercised here; token-only paths use
		// newSecureTestServer.
		AllowMockActors: true,
	}
	return NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)
}

// newSecureTestServer builds a server with mock actor headers disabled and a
// token secret configured: the production auth posture.
func newSecureTestServer() *Server {
	cfg := &config.Config{
		Addr:                   ":8102",
		AuditLogServiceURL:     "",
		AllowedOrigins:         nil,
		RateLimitRequests:      10,
		RateLimitWindowSeconds: 60,
		AuthTokenSecret:        testTokenSecret,
		AllowMockActors:        false,
	}
	return NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)
}

// signTestToken signs claims the same way auth-service does; tests may use any
// secret because they construct the server directly.
func signTestToken(t *testing.T, secret string, claims map[string]any) string {
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

func adminClaims() map[string]any {
	return map[string]any{
		"sub":      "usr_admin",
		"typ":      "agency",
		"role":     "system_admin",
		"agencyId": "00000000-0000-0000-0000-000000000101",
		"mfa":      true,
		"exp":      testNow.Add(time.Hour).Unix(),
	}
}

func setBearerToken(t *testing.T, r *http.Request, claims map[string]any) {
	t.Helper()
	r.Header.Set("Authorization", "Bearer "+signTestToken(t, testTokenSecret, claims))
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

	// Anonymous catalog reads are approved-only (finding #17).
	for _, dataset := range payload.Datasets {
		if dataset.PrivacyReviewStatus != models.PrivacyReviewApproved {
			t.Fatalf("anonymous catalog returned non-approved dataset %s (%s)", dataset.ID, dataset.PrivacyReviewStatus)
		}
		if dataset.ID == "dataset_raw_incident_feed" {
			t.Fatalf("anonymous catalog leaked the restricted pending dataset")
		}
	}
}

func TestListDatasetsAnonymousStatusFilterIgnored(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets?privacyReviewStatus=pending_review", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.DatasetListResponse
	decodeResponse(t, response, &payload)
	for _, dataset := range payload.Datasets {
		if dataset.PrivacyReviewStatus != models.PrivacyReviewApproved {
			t.Fatalf("anonymous status filter bypassed approved-only rule: %s", dataset.ID)
		}
	}
}

func TestListDatasetsAdminSeesAndFiltersAllStatuses(t *testing.T) {
	srv := newSecureTestServer()

	// Verified admin with no filter sees every dataset, including pending.
	allResponse := httptest.NewRecorder()
	allRequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets", nil)
	setBearerToken(t, allRequest, adminClaims())
	srv.Routes().ServeHTTP(allResponse, allRequest)

	if allResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, allResponse.Code, allResponse.Body.String())
	}
	var all models.DatasetListResponse
	decodeResponse(t, allResponse, &all)
	foundPending := false
	for _, dataset := range all.Datasets {
		if dataset.PrivacyReviewStatus == models.PrivacyReviewPending {
			foundPending = true
			if dataset.SampleRows != nil || dataset.Columns != nil {
				t.Fatalf("non-approved dataset %s exposed sample rows or columns to admin listing", dataset.ID)
			}
		}
	}
	if !foundPending {
		t.Fatalf("expected admin listing to include the pending dataset")
	}

	// Admin may filter by a non-approved status.
	pendingResponse := httptest.NewRecorder()
	pendingRequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets?privacyReviewStatus=pending_review", nil)
	setBearerToken(t, pendingRequest, adminClaims())
	srv.Routes().ServeHTTP(pendingResponse, pendingRequest)

	var pending models.DatasetListResponse
	decodeResponse(t, pendingResponse, &pending)
	if len(pending.Datasets) != 1 || pending.Datasets[0].ID != "dataset_raw_incident_feed" {
		t.Fatalf("expected only the pending dataset, got %#v", pending.Datasets)
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

func TestGetDatasetNonApprovedHiddenFromAnonymous(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_raw_incident_feed", nil)

	srv.Routes().ServeHTTP(response, request)

	// Same as missing: anonymous callers cannot even confirm it exists.
	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestGetDatasetNonApprovedAdminSeesScrubbedDetail(t *testing.T) {
	srv := newSecureTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_raw_incident_feed", nil)
	setBearerToken(t, request, adminClaims())

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DatasetDetailResponse
	decodeResponse(t, response, &payload)
	if payload.Dataset.PrivacyReviewStatus != models.PrivacyReviewPending {
		t.Fatalf("expected pending status, got %s", payload.Dataset.PrivacyReviewStatus)
	}
	if payload.Dataset.SampleRows != nil || payload.Dataset.Columns != nil {
		t.Fatalf("non-approved dataset exposed sample rows or columns")
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

func TestDownloadApprovedDatasetServesRealCSVBytes(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download?format=csv", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if ct := response.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("expected text/csv content type, got %q", ct)
	}
	if cd := response.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") || !strings.Contains(cd, "dataset_flood_reports_2026.csv") {
		t.Fatalf("expected attachment content disposition with csv filename, got %q", cd)
	}
	if got := response.Header().Get("X-NADAA-Audit-Logged"); got != "true" {
		t.Fatalf("expected X-NADAA-Audit-Logged true, got %q", got)
	}
	if got := response.Header().Get("X-RateLimit-Remaining"); got != "9" {
		t.Fatalf("expected X-RateLimit-Remaining 9, got %q", got)
	}

	// The body is the dataset's actual rows serialized as CSV.
	body := response.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected header plus 2 data rows, got %d lines: %q", len(lines), body)
	}
	if lines[0] != "date,district,reportCount,maxUrgency,injuriesReported" {
		t.Fatalf("unexpected csv header %q", lines[0])
	}
	if !strings.Contains(lines[1], "2026-07-01") || !strings.Contains(lines[1], "Accra Metropolitan") {
		t.Fatalf("unexpected first csv data row %q", lines[1])
	}

	// The recorded artifact reports the real byte size and the serving route.
	downloads := srv.store.GetDatasetDownloads("dataset_flood_reports_2026")
	if len(downloads) != 1 {
		t.Fatalf("expected one recorded download, got %#v", downloads)
	}
	if downloads[0].Size != int64(len(body)) {
		t.Fatalf("expected recorded size %d to match served bytes, got %d", len(body), downloads[0].Size)
	}
	if downloads[0].URL != "/api/v1/open-data/datasets/dataset_flood_reports_2026/download?format=csv" {
		t.Fatalf("expected recorded url to be the download route, got %q", downloads[0].URL)
	}

	// The download audit event is persisted locally and queryable by admins.
	events := srv.store.ListAuditEvents()
	if len(events) != 1 || events[0].Action != "dataset_download" || events[0].TargetID != "dataset_flood_reports_2026" {
		t.Fatalf("expected one persisted dataset_download audit event, got %#v", events)
	}
	if events[0].ID == "" {
		t.Fatalf("expected persisted audit event to have an id")
	}
	if events[0].Metadata["size"] != strconv.Itoa(len(body)) {
		t.Fatalf("expected audit metadata size %d, got %q", len(body), events[0].Metadata["size"])
	}
}

func TestDownloadApprovedDatasetServesRealJSONBytes(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download?format=json", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if ct := response.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
	if cd := response.Header().Get("Content-Disposition"); !strings.Contains(cd, "dataset_flood_reports_2026.json") {
		t.Fatalf("expected attachment content disposition with json filename, got %q", cd)
	}

	var rows []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &rows); err != nil {
		t.Fatalf("download body must be the dataset rows as JSON: %v", err)
	}
	if len(rows) != 2 || rows[0]["district"] != "Accra Metropolitan" {
		t.Fatalf("unexpected json rows %#v", rows)
	}

	downloads := srv.store.GetDatasetDownloads("dataset_flood_reports_2026")
	if len(downloads) != 1 || downloads[0].Size != int64(len(response.Body.Bytes())) {
		t.Fatalf("expected recorded size to match served json bytes, got %#v", downloads)
	}
	if downloads[0].Format != "json" {
		t.Fatalf("expected recorded format json, got %q", downloads[0].Format)
	}
}

func TestDownloadUnsupportedFormatRejected(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download?format=parquet", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
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
	for i := range 11 {
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

	review := models.ReviewOpenDataRequest{Reviewer: "spoofed@example.com", Approved: true, Note: "Approved for academic use."}
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
	// Attribution comes from the verified actor, not the body-supplied reviewer.
	if approved.Request.ReviewedBy != "usr_admin" {
		t.Fatalf("expected reviewedBy usr_admin (actor), got %q", approved.Request.ReviewedBy)
	}

	// The review decision is audit-logged with the verified admin actor.
	events := srv.store.ListAuditEvents()
	if len(events) != 1 || events[0].Action != "access_request_review" || events[0].TargetID != created.Request.ID {
		t.Fatalf("expected one persisted access_request_review audit event, got %#v", events)
	}
	if events[0].Metadata["reviewer"] != "usr_admin" || events[0].Metadata["decision"] != "approved" {
		t.Fatalf("expected audit metadata reviewer=usr_admin decision=approved, got %#v", events[0].Metadata)
	}

	// Re-review is refused with 409 and overwrites neither the decision nor
	// the audit trail.
	againResponse := httptest.NewRecorder()
	againRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests/"+created.Request.ID+"/approve", jsonBody(models.ReviewOpenDataRequest{Approved: false}))
	againRequest.Header.Set("Content-Type", "application/json")
	setAdminHeaders(againRequest)
	srv.Routes().ServeHTTP(againResponse, againRequest)
	if againResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d on re-review, got %d: %s", http.StatusConflict, againResponse.Code, againResponse.Body.String())
	}
	record, found := srv.store.GetRequest(created.Request.ID)
	if !found || record.Status != models.OpenDataRequestApproved || record.ReviewedBy != "usr_admin" {
		t.Fatalf("expected re-review to leave the original decision untouched, got %#v", record)
	}
	if len(srv.store.ListAuditEvents()) != 1 {
		t.Fatalf("expected re-review to add no audit event, got %#v", srv.store.ListAuditEvents())
	}
}

func TestRejectRequestAuditedAndReReviewConflict(t *testing.T) {
	srv := newTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID:     "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{Name: "Ama Kwame", Email: "ama@example.edu.gh", UseCase: "research"},
		Purpose:       "Research on flood response patterns in Accra.",
	}
	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
	createRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createRequest)

	var created models.OpenDataRequestResponse
	decodeResponse(t, createResponse, &created)

	rejectResponse := httptest.NewRecorder()
	rejectRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests/"+created.Request.ID+"/approve", jsonBody(models.ReviewOpenDataRequest{Approved: false, Note: "Insufficient data-use agreement."}))
	rejectRequest.Header.Set("Content-Type", "application/json")
	setAdminHeaders(rejectRequest)
	srv.Routes().ServeHTTP(rejectResponse, rejectRequest)

	if rejectResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rejectResponse.Code, rejectResponse.Body.String())
	}
	var rejected models.OpenDataRequestResponse
	decodeResponse(t, rejectResponse, &rejected)
	if rejected.Request.Status != models.OpenDataRequestRejected {
		t.Fatalf("expected rejected status, got %s", rejected.Request.Status)
	}

	events := srv.store.ListAuditEvents()
	if len(events) != 1 || events[0].Action != "access_request_review" || events[0].Metadata["decision"] != "rejected" {
		t.Fatalf("expected rejected review decision audit-logged, got %#v", events)
	}

	reReviewResponse := httptest.NewRecorder()
	reReviewRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests/"+created.Request.ID+"/approve", jsonBody(models.ReviewOpenDataRequest{Approved: true}))
	reReviewRequest.Header.Set("Content-Type", "application/json")
	setAdminHeaders(reReviewRequest)
	srv.Routes().ServeHTTP(reReviewResponse, reReviewRequest)
	if reReviewResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d on re-review of rejected request, got %d", http.StatusConflict, reReviewResponse.Code)
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
	for range 3 {
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

func TestAdminEndpointsWithValidToken(t *testing.T) {
	srv := newSecureTestServer()

	// No token and mock actors disabled -> 401, even with legacy headers set.
	noAuth := httptest.NewRecorder()
	noAuthRequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
	setAdminHeaders(noAuthRequest)
	srv.Routes().ServeHTTP(noAuth, noAuthRequest)
	if noAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for legacy headers with mock actors disabled, got %d", noAuth.Code)
	}

	// Valid signed agency token with MFA and admin role -> 200.
	valid := httptest.NewRecorder()
	validRequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
	setBearerToken(t, validRequest, adminClaims())
	srv.Routes().ServeHTTP(valid, validRequest)
	if valid.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid admin token, got %d: %s", valid.Code, valid.Body.String())
	}

	cases := []struct {
		name   string
		mutate func(claims map[string]any)
		token  string
		want   int
	}{
		{
			name: "expired token",
			mutate: func(claims map[string]any) {
				claims["exp"] = testNow.Add(-time.Hour).Unix()
			},
			want: http.StatusUnauthorized,
		},
		{
			name: "mfa not completed",
			mutate: func(claims map[string]any) {
				claims["mfa"] = false
			},
			want: http.StatusForbidden,
		},
		{
			name: "citizen token",
			mutate: func(claims map[string]any) {
				claims["typ"] = "citizen"
			},
			want: http.StatusForbidden,
		},
		{
			name: "non-admin role",
			mutate: func(claims map[string]any) {
				claims["role"] = "agency_viewer"
			},
			want: http.StatusForbidden,
		},
		{
			name:  "garbage token",
			token: "nadaa.not-a-token.nope",
			want:  http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.token
			if token == "" {
				claims := adminClaims()
				tc.mutate(claims)
				token = signTestToken(t, testTokenSecret, claims)
			}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
			request.Header.Set("Authorization", "Bearer "+token)
			srv.Routes().ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("expected %d, got %d: %s", tc.want, response.Code, response.Body.String())
			}
		})
	}

	// A token signed with the wrong secret must not verify.
	forged := httptest.NewRecorder()
	forgedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/requests", nil)
	forgedRequest.Header.Set("Authorization", "Bearer "+signTestToken(t, "wrong-secret", adminClaims()))
	srv.Routes().ServeHTTP(forged, forgedRequest)
	if forged.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong-secret token, got %d", forged.Code)
	}
}

func TestApproveRequestReviewedByFromToken(t *testing.T) {
	srv := newSecureTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID:     "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{Name: "Ama Kwame", Email: "ama@example.edu.gh", UseCase: "research"},
		Purpose:       "Research on flood response patterns in Accra.",
	}
	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
	createRequest.Header.Set("Content-Type", "application/json")
	srv.Routes().ServeHTTP(createResponse, createRequest)

	var created models.OpenDataRequestResponse
	decodeResponse(t, createResponse, &created)

	claims := adminClaims()
	claims["sub"] = "usr_admin_9"
	review := models.ReviewOpenDataRequest{Reviewer: "spoofed@example.com", Approved: true}
	approveResponse := httptest.NewRecorder()
	approveRequest := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests/"+created.Request.ID+"/approve", jsonBody(review))
	approveRequest.Header.Set("Content-Type", "application/json")
	setBearerToken(t, approveRequest, claims)
	srv.Routes().ServeHTTP(approveResponse, approveRequest)

	if approveResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, approveResponse.Code, approveResponse.Body.String())
	}
	var approved models.OpenDataRequestResponse
	decodeResponse(t, approveResponse, &approved)
	if approved.Request.ReviewedBy != "usr_admin_9" {
		t.Fatalf("expected reviewedBy from token subject usr_admin_9, got %q", approved.Request.ReviewedBy)
	}
}

func TestCreateRequestRateLimited(t *testing.T) {
	srv := newTestServer()
	body := models.CreateOpenDataRequest{
		DatasetID:     "dataset_raw_incident_feed",
		RequesterInfo: models.RequesterInfo{Name: "Ama Kwame", Email: "ama@example.edu.gh", UseCase: "research"},
		Purpose:       "Research on flood response patterns in Accra.",
	}
	for i := range 11 {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/open-data/requests", jsonBody(body))
		request.Header.Set("Content-Type", "application/json")
		srv.Routes().ServeHTTP(response, request)
		if i < 10 && response.Code != http.StatusCreated {
			t.Fatalf("expected created on request %d, got %d", i, response.Code)
		}
		if i == 10 && response.Code != http.StatusTooManyRequests {
			t.Fatalf("expected rate limit on request %d, got %d", i, response.Code)
		}
	}
}

func TestRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	srv := newTestServer()
	// Every request spoofs a different XFF, but proxy headers are not trusted by
	// default, so they all share the single RemoteAddr bucket.
	for i := range 11 {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download", nil)
		request.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(i))
		srv.Routes().ServeHTTP(response, request)
		if i < 10 && response.Code != http.StatusOK {
			t.Fatalf("expected ok on request %d, got %d", i, response.Code)
		}
		if i == 10 && response.Code != http.StatusTooManyRequests {
			t.Fatalf("expected rate limit despite spoofed XFF on request %d, got %d", i, response.Code)
		}
	}
}

func TestClientIP(t *testing.T) {
	srv := newTestServer()

	direct := httptest.NewRequest(http.MethodGet, "/", nil)
	direct.RemoteAddr = "198.51.100.7:43210"
	if ip := srv.clientIP(direct); ip != "198.51.100.7" {
		t.Fatalf("expected host without port, got %q", ip)
	}

	spoofed := httptest.NewRequest(http.MethodGet, "/", nil)
	spoofed.RemoteAddr = "198.51.100.7:43210"
	spoofed.Header.Set("X-Forwarded-For", "203.0.113.99")
	if ip := srv.clientIP(spoofed); ip != "198.51.100.7" {
		t.Fatalf("expected XFF ignored by default, got %q", ip)
	}

	srv.config.TrustProxyHeaders = true
	if ip := srv.clientIP(spoofed); ip != "203.0.113.99" {
		t.Fatalf("expected trusted XFF honored, got %q", ip)
	}
	chained := httptest.NewRequest(http.MethodGet, "/", nil)
	chained.RemoteAddr = "198.51.100.7:43210"
	chained.Header.Set("X-Forwarded-For", "203.0.113.99, 10.0.0.1")
	if ip := srv.clientIP(chained); ip != "203.0.113.99" {
		t.Fatalf("expected first XFF entry, got %q", ip)
	}
}

func TestListAuditEventsEndpoint(t *testing.T) {
	srv := newSecureTestServer()

	// Anonymous -> 401.
	denied := httptest.NewRecorder()
	srv.Routes().ServeHTTP(denied, httptest.NewRequest(http.MethodGet, "/api/v1/open-data/audit", nil))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for anonymous audit read, got %d", denied.Code)
	}

	// One public download -> one persisted audit event.
	download := httptest.NewRecorder()
	srv.Routes().ServeHTTP(download, httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download", nil))
	if download.Code != http.StatusOK {
		t.Fatalf("expected download ok, got %d", download.Code)
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/audit", nil)
	setBearerToken(t, request, adminClaims())
	srv.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.AuditEventListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Events) != 1 {
		t.Fatalf("expected one audit event, got %#v", payload.Events)
	}
	event := payload.Events[0]
	if event.Action != "dataset_download" || event.TargetID != "dataset_flood_reports_2026" || event.ID == "" {
		t.Fatalf("unexpected audit event %#v", event)
	}
}

func TestAuditForwardingFailureStillHonest(t *testing.T) {
	// The remote audit service rejects the event; local persistence is the
	// source of truth, so the response must still report the local record.
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer remote.Close()

	cfg := &config.Config{
		Addr:                   ":8102",
		AuditLogServiceURL:     remote.URL,
		RateLimitRequests:      10,
		RateLimitWindowSeconds: 60,
		AllowMockActors:        true,
	}
	srv := NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/open-data/datasets/dataset_flood_reports_2026/download", nil)
	srv.Routes().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	if got := response.Header().Get("X-NADAA-Audit-Logged"); got != "true" {
		t.Fatalf("expected X-NADAA-Audit-Logged true from local persistence even when forwarding fails, got %q", got)
	}
	if len(srv.store.ListAuditEvents()) != 1 {
		t.Fatalf("expected one locally persisted audit event")
	}
}

func TestSendAuditEventForwardsWithServiceToken(t *testing.T) {
	received := make(chan *http.Request, 1)
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"audit_000001"}`))
	}))
	defer sink.Close()

	cfg := &config.Config{
		Addr:                 ":8102",
		AuditLogServiceURL:   sink.URL,
		InternalServiceToken: "svc-secret",
	}
	srv := NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer caller-credential")
	srv.sendAuditEvent(request, models.AuditEvent{ID: "audit_000001", Action: "dataset_download", TargetID: "dataset_flood_reports_2026"})

	select {
	case got := <-received:
		if token := got.Header.Get("X-NADAA-Service-Token"); token != "svc-secret" {
			t.Fatalf("expected X-NADAA-Service-Token svc-secret, got %q", token)
		}
		// The caller's credentials are never forwarded to the audit sink.
		if authorization := got.Header.Get("Authorization"); authorization != "" {
			t.Fatalf("expected no forwarded Authorization header, got %q", authorization)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected the audit event to be forwarded to the sink")
	}
}

func TestSendAuditEventLogsRejectedForward(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusMethodNotAllowed} {
		received := make(chan struct{}, 1)
		sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			received <- struct{}{}
			w.WriteHeader(status)
		}))

		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		cfg := &config.Config{
			Addr:                 ":8102",
			AuditLogServiceURL:   sink.URL,
			InternalServiceToken: "svc-secret",
		}
		srv := NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)

		srv.sendAuditEvent(httptest.NewRequest(http.MethodGet, "/", nil), models.AuditEvent{ID: "audit_000002", Action: "dataset_download", TargetID: "dataset_flood_reports_2026"})

		select {
		case <-received:
		case <-time.After(2 * time.Second):
			log.SetOutput(os.Stderr)
			sink.Close()
			t.Fatalf("expected the audit event to reach the sink for status %d", status)
		}
		// The rejection is logged after the sink responds; wait for it.
		deadline := time.Now().Add(2 * time.Second)
		for !strings.Contains(logBuf.String(), "audit_forward_rejected") && time.Now().Before(deadline) {
			time.Sleep(5 * time.Millisecond)
		}
		log.SetOutput(os.Stderr)
		sink.Close()
		if !strings.Contains(logBuf.String(), "audit_forward_rejected") {
			t.Fatalf("expected rejected forward (status %d) to be logged, got %q", status, logBuf.String())
		}
	}
}
