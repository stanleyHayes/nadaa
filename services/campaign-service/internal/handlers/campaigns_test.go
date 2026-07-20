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

	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/store"
)

const testTokenSecret = "test-secret-for-unit-tests"

var testNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

func newTestServer() *Server {
	cfg := &config.Config{Addr: ":8103", TokenSecret: testTokenSecret, AllowMockActorHeaders: true}
	return NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)
}

// newStrictTestServer disables mock actor headers so only verified bearer
// tokens authorize authority requests.
func newStrictTestServer() *Server {
	cfg := &config.Config{Addr: ":8103", TokenSecret: testTokenSecret}
	return NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)
}

func signTestToken(t *testing.T, claims models.TokenClaims) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(testTokenSecret))
	_, _ = mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func authorityClaims() models.TokenClaims {
	return models.TokenClaims{
		UserID:    "usr_campaign_officer",
		UserType:  "agency",
		Role:      "nadmo_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		MFA:       true,
		ExpiresAt: testNow.Add(time.Hour).Unix(),
	}
}

func tokenRequest(t *testing.T, method, path string, body io.Reader, token string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, path, body)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func jsonBody(v any) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}

func authorityRequest(method, path string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, path, body)
	request.Header.Set("X-NADAA-Actor-ID", "usr_campaign_officer")
	request.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("Content-Type", "application/json")
	return request
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

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestListCampaignsPublic(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns", nil)

	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Campaigns) != 2 {
		t.Fatalf("expected 2 public campaigns, got %d", len(payload.Campaigns))
	}
}

func TestListCampaignsFiltersByRegion(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns?region=Greater+Accra", nil)

	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(payload.Campaigns))
	}
}

func TestListCampaignsFiltersByLanguage(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns?language=tw", nil)

	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Campaigns) != 1 || payload.Campaigns[0].ID != "campaign_001" {
		t.Fatalf("expected flood campaign in Twi, got %#v", payload.Campaigns)
	}
}

func TestListCampaignsAuthorityIncludesDrafts(t *testing.T) {
	srv := newTestServer()
	create := authorityRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(models.CreateCampaignRequest{
		Title:         "Draft cholera campaign",
		HazardType:    "disease_outbreak",
		TargetRegions: []string{"Greater Accra"},
		Languages:     []string{"en"},
		ContentBlocks: []models.CampaignContentBlock{{Type: "article", Title: "Wash hands", Body: "Handwashing prevents cholera."}},
		PublishWindow: models.CampaignPublishWindow{
			StartsAt: time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC),
		},
		Status: "draft",
	}))
	create.SetPathValue("id", "")
	srv.createCampaignHandler(httptest.NewRecorder(), create)

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/campaigns?status=draft", nil)
	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Campaigns) != 1 {
		t.Fatalf("expected 1 draft campaign, got %d", len(payload.Campaigns))
	}
}

func TestGetCampaign(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/campaign_001", nil)
	request.SetPathValue("id", "campaign_001")

	srv.getCampaignHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignResponse
	decodeResponse(t, response, &payload)
	if payload.Campaign.ID != "campaign_001" {
		t.Fatalf("expected campaign_001, got %#v", payload.Campaign)
	}
}

func TestGetCampaignNotFound(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/missing", nil)
	request.SetPathValue("id", "missing")

	srv.getCampaignHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestCreateCampaignRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(models.CreateCampaignRequest{Title: "Test"}))

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateCampaign(t *testing.T) {
	srv := newTestServer()
	body := models.CreateCampaignRequest{
		Title:         "New flood campaign",
		HazardType:    "flood",
		TargetRegions: []string{"Greater Accra"},
		Languages:     []string{"en"},
		ContentBlocks: []models.CampaignContentBlock{
			{Type: "article", Title: "Stay informed", Body: "Listen to official updates."},
		},
		PublishWindow: models.CampaignPublishWindow{
			StartsAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		},
		Status: "published",
	}
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(body))

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.CampaignResponse
	decodeResponse(t, response, &payload)
	if payload.Campaign.Title != "New flood campaign" || payload.Campaign.CreatedBy != "usr_campaign_officer" {
		t.Fatalf("expected created campaign, got %#v", payload.Campaign)
	}
}

func TestCreateCampaignRejectsStaleWindow(t *testing.T) {
	srv := newTestServer()
	body := models.CreateCampaignRequest{
		Title:         "Stale campaign",
		HazardType:    "flood",
		TargetRegions: []string{"Greater Accra"},
		Languages:     []string{"en"},
		ContentBlocks: []models.CampaignContentBlock{{Type: "article", Title: "Old", Body: "Old"}},
		PublishWindow: models.CampaignPublishWindow{
			StartsAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC),
		},
		Status: "published",
	}
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(body))

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateCampaignRejectsOversizedBody(t *testing.T) {
	srv := newTestServer()
	oversized := bytes.NewReader(append([]byte(`{"title":"`), bytes.Repeat([]byte("a"), 2<<20)...))
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/campaigns", oversized)

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestUpdateCampaignRejectsOversizedBody(t *testing.T) {
	srv := newTestServer()
	oversized := bytes.NewReader(append([]byte(`{"title":"`), bytes.Repeat([]byte("a"), 2<<20)...))
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPut, "/api/v1/campaigns/campaign_001", oversized)
	request.SetPathValue("id", "campaign_001")

	srv.updateCampaignHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestUpdateCampaign(t *testing.T) {
	srv := newTestServer()
	body := models.UpdateCampaignRequest{
		Title:  "Updated flood campaign",
		Status: "archived",
	}
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPut, "/api/v1/campaigns/campaign_001", jsonBody(body))
	request.SetPathValue("id", "campaign_001")

	srv.updateCampaignHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignResponse
	decodeResponse(t, response, &payload)
	if payload.Campaign.Title != "Updated flood campaign" || payload.Campaign.Status != "archived" {
		t.Fatalf("expected updated campaign, got %#v", payload.Campaign)
	}
}

func TestUpdateCampaignNotFound(t *testing.T) {
	srv := newTestServer()
	body := models.UpdateCampaignRequest{Title: "Updated"}
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPut, "/api/v1/campaigns/missing", jsonBody(body))
	request.SetPathValue("id", "missing")

	srv.updateCampaignHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestGetCampaignMetrics(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/campaign_001/metrics", nil)
	request.SetPathValue("id", "campaign_001")

	srv.getCampaignMetricsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignMetricListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Metrics) == 0 {
		t.Fatalf("expected metrics, got %#v", payload)
	}
}

func TestListCampaignTemplates(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaign-templates", nil)

	srv.listCampaignTemplatesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.CampaignTemplateListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Templates) == 0 {
		t.Fatalf("expected templates, got %#v", payload)
	}
}

func createDraftForTest(t *testing.T, srv *Server, window models.CampaignPublishWindow) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := authorityRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(models.CreateCampaignRequest{
		Title:         "Draft campaign",
		HazardType:    "flood",
		TargetRegions: []string{"Greater Accra"},
		Languages:     []string{"en"},
		ContentBlocks: []models.CampaignContentBlock{{Type: "article", Title: "T", Body: "Body content for the draft."}},
		PublishWindow: window,
		Status:        "draft",
	}))
	srv.createCampaignHandler(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("draft create expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created models.CampaignResponse
	decodeResponse(t, rec, &created)
	return created.Campaign.ID
}

func TestGetCampaignHidesDraftFromPublic(t *testing.T) {
	srv := newTestServer()
	id := createDraftForTest(t, srv, models.CampaignPublishWindow{
		StartsAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})

	pub := httptest.NewRecorder()
	pubReq := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/"+id, nil)
	pubReq.SetPathValue("id", id)
	srv.getCampaignHandler(pub, pubReq)
	if pub.Code != http.StatusNotFound {
		t.Fatalf("public draft access must 404, got %d", pub.Code)
	}

	pubMetrics := httptest.NewRecorder()
	pubMetricsReq := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/"+id+"/metrics", nil)
	pubMetricsReq.SetPathValue("id", id)
	srv.getCampaignMetricsHandler(pubMetrics, pubMetricsReq)
	if pubMetrics.Code != http.StatusNotFound {
		t.Fatalf("public draft metrics must 404, got %d", pubMetrics.Code)
	}

	auth := httptest.NewRecorder()
	authReq := authorityRequest(http.MethodGet, "/api/v1/campaigns/"+id, nil)
	authReq.SetPathValue("id", id)
	srv.getCampaignHandler(auth, authReq)
	if auth.Code != http.StatusOK {
		t.Fatalf("authority draft access must 200, got %d: %s", auth.Code, auth.Body.String())
	}
}

func TestStatusOnlyPublishRevalidatesWindow(t *testing.T) {
	srv := newTestServer()
	// Draft whose window already ended before the test clock (2026-07-06).
	id := createDraftForTest(t, srv, models.CampaignPublishWindow{
		StartsAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
	})

	upd := httptest.NewRecorder()
	updReq := authorityRequest(http.MethodPut, "/api/v1/campaigns/"+id, jsonBody(models.UpdateCampaignRequest{Status: "published"}))
	updReq.SetPathValue("id", id)
	srv.updateCampaignHandler(upd, updReq)
	if upd.Code != http.StatusBadRequest {
		t.Fatalf("publishing a stale-window campaign must 400, got %d: %s", upd.Code, upd.Body.String())
	}
}

func TestUpdateDraftFutureWindowAllowed(t *testing.T) {
	srv := newTestServer()
	id := createDraftForTest(t, srv, models.CampaignPublishWindow{
		StartsAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	})

	// Moving a DRAFT's window to future dates must not be rejected as premature.
	upd := httptest.NewRecorder()
	updReq := authorityRequest(http.MethodPut, "/api/v1/campaigns/"+id, jsonBody(models.UpdateCampaignRequest{
		PublishWindow: &models.CampaignPublishWindow{
			StartsAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}))
	updReq.SetPathValue("id", id)
	srv.updateCampaignHandler(upd, updReq)
	if upd.Code != http.StatusOK {
		t.Fatalf("updating a draft's future window must 200, got %d: %s", upd.Code, upd.Body.String())
	}
}

func TestListCampaignsForbiddenStatusFilterReturns403(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns?status=draft", nil)

	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "forbidden" {
		t.Fatalf("expected forbidden error code, got %#v", payload.Error)
	}
}

func TestCreateCampaignDefaultsStatusToDraft(t *testing.T) {
	srv := newTestServer()
	body := models.CreateCampaignRequest{
		Title:         "Unstated campaign",
		HazardType:    "flood",
		TargetRegions: []string{"Greater Accra"},
		Languages:     []string{"en"},
		ContentBlocks: []models.CampaignContentBlock{{Type: "article", Title: "T", Body: "Body content."}},
		PublishWindow: models.CampaignPublishWindow{
			StartsAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(body))

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.CampaignResponse
	decodeResponse(t, response, &payload)
	if payload.Campaign.Status != "draft" {
		t.Fatalf("expected empty status to default to draft, got %#v", payload.Campaign.Status)
	}
}

func TestCreateCampaignWithBearerToken(t *testing.T) {
	srv := newStrictTestServer()
	token := signTestToken(t, authorityClaims())
	body := models.CreateCampaignRequest{
		Title:         "Token-created campaign",
		HazardType:    "flood",
		TargetRegions: []string{"Greater Accra"},
		Languages:     []string{"en"},
		ContentBlocks: []models.CampaignContentBlock{{Type: "article", Title: "T", Body: "Body content."}},
		PublishWindow: models.CampaignPublishWindow{
			StartsAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	response := httptest.NewRecorder()
	request := tokenRequest(t, http.MethodPost, "/api/v1/campaigns", jsonBody(body), token)

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.CampaignResponse
	decodeResponse(t, response, &payload)
	if payload.Campaign.CreatedBy != "usr_campaign_officer" {
		t.Fatalf("expected created-by from verified claims, got %#v", payload.Campaign.CreatedBy)
	}
}

func TestBearerTokenAuthoritySeesDrafts(t *testing.T) {
	srv := newStrictTestServer()
	token := signTestToken(t, authorityClaims())

	response := httptest.NewRecorder()
	request := tokenRequest(t, http.MethodGet, "/api/v1/campaigns?status=draft", nil, token)

	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestForgedHeadersRejectedWhenMockActorsOff(t *testing.T) {
	srv := newStrictTestServer()
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/campaigns", jsonBody(models.CreateCampaignRequest{Title: "Forged"}))

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}

	// Forged headers must not widen the public list view either.
	list := httptest.NewRecorder()
	listReq := authorityRequest(http.MethodGet, "/api/v1/campaigns?status=draft", nil)
	srv.listCampaignsHandler(list, listReq)
	if list.Code != http.StatusForbidden {
		t.Fatalf("forged authority headers must not include drafts, got %d", list.Code)
	}
}

func TestCreateCampaignRejectsInvalidToken(t *testing.T) {
	srv := newStrictTestServer()
	response := httptest.NewRecorder()
	request := tokenRequest(t, http.MethodPost, "/api/v1/campaigns", jsonBody(models.CreateCampaignRequest{Title: "Bad token"}), "nadaa.garbage.signature")

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_token" {
		t.Fatalf("expected invalid_token error code, got %#v", payload.Error)
	}
}

func TestCreateCampaignRejectsExpiredToken(t *testing.T) {
	srv := newStrictTestServer()
	claims := authorityClaims()
	claims.ExpiresAt = testNow.Add(-time.Hour).Unix()
	token := signTestToken(t, claims)
	response := httptest.NewRecorder()
	request := tokenRequest(t, http.MethodPost, "/api/v1/campaigns", jsonBody(models.CreateCampaignRequest{Title: "Expired"}), token)

	srv.createCampaignHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestPublicListStaysPublicWithoutCredentials(t *testing.T) {
	srv := newStrictTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns", nil)

	srv.listCampaignsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}
