package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/store"
)

const testTokenSecret = "road-closure-test-token-secret"

var testNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

func newTestServer() *Server {
	cfg := &config.Config{Addr: ":8095", AllowMockActorHeaders: true}
	return NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)
}

// newTokenTestServer builds a server that only accepts verified bearer tokens
// (mock actor headers disabled), mirroring production configuration.
func newTokenTestServer() *Server {
	cfg := &config.Config{Addr: ":8095", AuthTokenSecret: testTokenSecret}
	return NewServer(store.NewMemoryStore(testNow), func() time.Time { return testNow }, cfg)
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

func TestListRoadClosuresDefaultsToActive(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 2 {
		t.Fatalf("expected 2 active closures, got %d", len(payload.Closures))
	}
}

func TestListRoadClosuresFiltersByStatus(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?status=scheduled", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 1 || payload.Closures[0].Status != "scheduled" {
		t.Fatalf("expected 1 scheduled closure, got %#v", payload.Closures)
	}
}

// createScheduledClosure creates a scheduled closure with the given validity
// window and returns its id.
func createScheduledClosure(t *testing.T, srv *Server, validFrom, validTo time.Time) string {
	t.Helper()
	body := models.CreateRoadClosureRequest{
		RoadName:  "Windowed Scheduled Road",
		Status:    "scheduled",
		Severity:  "moderate",
		ValidFrom: &validFrom,
		ValidTo:   &validTo,
		Geometry: models.LineStringGeometry{
			Type:        "LineString",
			Coordinates: [][]float64{{-0.220, 5.560}, {-0.215, 5.562}},
		},
	}
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))
	srv.createRoadClosureHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var created models.RoadClosureResponse
	decodeResponse(t, response, &created)
	return created.Closure.ID
}

func listClosureIDs(t *testing.T, srv *Server, query string) models.RoadClosureListResponse {
	t.Helper()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures"+query, nil)
	srv.listRoadClosuresHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	return payload
}

func TestScheduledClosureInsideValidityWindowIsActive(t *testing.T) {
	srv := newTestServer()
	// Boundary values are inclusive: validFrom == now and validTo == now both
	// count as inside the window.
	id := createScheduledClosure(t, srv, testNow, testNow)

	active := listClosureIDs(t, srv, "?status=active")
	for _, closure := range active.Closures {
		if closure.ID == id {
			if closure.Status != "active" {
				t.Fatalf("expected in-window scheduled closure to be served as active, got %q", closure.Status)
			}
			return
		}
	}
	t.Fatalf("expected in-window scheduled closure %s in status=active results, got %#v", id, active.Closures)
}

func TestScheduledClosureInsideValidityWindowMatchesDefaultAndScheduledQueries(t *testing.T) {
	srv := newTestServer()
	id := createScheduledClosure(t, srv, testNow.Add(-time.Hour), testNow.Add(time.Hour))

	defaultList := listClosureIDs(t, srv, "")
	found := false
	for _, closure := range defaultList.Closures {
		if closure.ID == id {
			found = true
			if closure.Status != "active" {
				t.Fatalf("expected in-window scheduled closure served as active, got %q", closure.Status)
			}
		}
	}
	if !found {
		t.Fatalf("expected in-window scheduled closure %s in default results, got %#v", id, defaultList.Closures)
	}

	// A scheduled query must still return it as scheduled: the record itself
	// is not rewritten, only its route-facing effectiveness.
	scheduled := listClosureIDs(t, srv, "?status=scheduled")
	for _, closure := range scheduled.Closures {
		if closure.ID == id {
			t.Fatalf("expected in-window closure to stay out of status=scheduled results, got %#v", closure)
		}
	}
}

func TestScheduledClosureOutsideValidityWindowStaysScheduled(t *testing.T) {
	srv := newTestServer()
	id := createScheduledClosure(t, srv, testNow.Add(time.Hour), testNow.Add(2*time.Hour))

	active := listClosureIDs(t, srv, "?status=active")
	for _, closure := range active.Closures {
		if closure.ID == id {
			t.Fatalf("expected future scheduled closure to stay out of status=active results, got %#v", closure)
		}
	}

	scheduled := listClosureIDs(t, srv, "?status=scheduled")
	found := false
	for _, closure := range scheduled.Closures {
		if closure.ID == id && closure.Status == "scheduled" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected future scheduled closure %s in status=scheduled results, got %#v", id, scheduled.Closures)
	}
}

func TestListRoadClosuresNearby(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?lat=5.570&lng=-0.200", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) == 0 {
		t.Fatalf("expected nearby closures, got %#v", payload.Closures)
	}
	if payload.Closures[0].DistanceMeters <= 0 {
		t.Fatalf("expected distance populated, got %#v", payload.Closures[0])
	}
}

func TestListRoadClosuresByBBox(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?bbox=-0.30,5.50,-0.15,5.60", nil)

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 2 {
		t.Fatalf("expected 2 closures in bbox, got %d", len(payload.Closures))
	}
}

func TestCreateRoadClosureRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(models.CreateRoadClosureRequest{RoadName: "Test Road"}))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateRoadClosure(t *testing.T) {
	srv := newTestServer()
	body := models.CreateRoadClosureRequest{
		RoadName: "Ring Road Central",
		Reason:   "Flooding",
		Status:   "active",
		Severity: "high",
		Geometry: models.LineStringGeometry{
			Type: "LineString",
			Coordinates: [][]float64{
				{-0.210, 5.550},
				{-0.200, 5.552},
			},
		},
		DetourNote: "Use Independence Avenue",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.RoadClosureResponse
	decodeResponse(t, response, &payload)
	if payload.Closure.RoadName != "Ring Road Central" || payload.Closure.Status != "active" || payload.Closure.CreatedBy != "usr_road_closure_officer" {
		t.Fatalf("expected created closure, got %#v", payload.Closure)
	}
}

func TestCreateRoadClosureRejectsInvalidGeometry(t *testing.T) {
	srv := newTestServer()
	body := models.CreateRoadClosureRequest{
		RoadName: "Bad Road",
		Status:   "active",
		Geometry: models.LineStringGeometry{
			Type: "Polygon",
			Coordinates: [][]float64{
				{-0.210, 5.550},
			},
		},
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestUpdateRoadClosure(t *testing.T) {
	srv := newTestServer()
	body := models.UpdateRoadClosureRequest{
		Status: "lifted",
		Reason: "Water receded",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/road-closures/road_closure_001", jsonBody(body))
	request.SetPathValue("id", "road_closure_001")

	srv.updateRoadClosureHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureResponse
	decodeResponse(t, response, &payload)
	if payload.Closure.Status != "lifted" || payload.Closure.UpdatedBy != "usr_road_closure_officer" {
		t.Fatalf("expected lifted closure, got %#v", payload.Closure)
	}
}

func TestUpdateRoadClosureNotFound(t *testing.T) {
	srv := newTestServer()
	body := models.UpdateRoadClosureRequest{Status: "lifted"}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/road-closures/missing", jsonBody(body))
	request.SetPathValue("id", "missing")

	srv.updateRoadClosureHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestImportAdapter(t *testing.T) {
	srv := newTestServer()
	body := models.AdapterImportRequest{
		Source:    "ghana-police",
		SourceRef: "police-road-closure-feed",
		RoadName:  "Sample Market Road",
		Status:    "active",
		Reason:    "Flooding",
		Geometry:  "LINESTRING(-0.20 5.56, -0.19 5.57)",
		ValidFrom: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC),
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures/imports/adapter", jsonBody(body))

	srv.importAdapterHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.AdapterImportResponse
	decodeResponse(t, response, &payload)
	if payload.Imported != 1 || payload.Source != "ghana-police" {
		t.Fatalf("expected one imported closure from ghana-police, got %#v", payload)
	}
	if payload.Closures[0].Geometry.Type != "LineString" || len(payload.Closures[0].Geometry.Coordinates) != 2 {
		t.Fatalf("expected parsed LineString geometry, got %#v", payload.Closures[0].Geometry)
	}
}

func TestImportAdapterRejectsInvalidWKT(t *testing.T) {
	srv := newTestServer()
	body := models.AdapterImportRequest{
		Source:    "ghana-police",
		RoadName:  "Bad Road",
		Status:    "active",
		Geometry:  "POINT(0 0)",
		ValidFrom: time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC),
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures/imports/adapter", jsonBody(body))

	srv.importAdapterHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateRoadClosureIDsDoNotCollideWithSeeds(t *testing.T) {
	srv := newTestServer()
	body := models.CreateRoadClosureRequest{
		RoadName: "New Spintex Road",
		Status:   "active",
		Geometry: models.LineStringGeometry{
			Type:        "LineString",
			Coordinates: [][]float64{{-0.100, 5.600}, {-0.090, 5.610}},
		},
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var created models.RoadClosureResponse
	decodeResponse(t, response, &created)
	if created.Closure.ID != "road_closure_004" {
		t.Fatalf("expected first created closure to be road_closure_004, got %s", created.Closure.ID)
	}

	// Updating the new closure by its own ID must not touch the seed that
	// previously shared it.
	update := models.UpdateRoadClosureRequest{Status: "lifted"}
	updateResponse := httptest.NewRecorder()
	updateRequest := authorityRequest(http.MethodPatch, "/api/v1/road-closures/road_closure_004", jsonBody(update))
	updateRequest.SetPathValue("id", "road_closure_004")

	srv.updateRoadClosureHandler(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, updateResponse.Code, updateResponse.Body.String())
	}
	var updated models.RoadClosureResponse
	decodeResponse(t, updateResponse, &updated)
	if updated.Closure.RoadName != "New Spintex Road" || updated.Closure.Status != "lifted" {
		t.Fatalf("expected the new closure to be updated, got %#v", updated.Closure)
	}
}

func TestUpdateRoadClosureRejectsMergedInvalidWindow(t *testing.T) {
	// Seed road_closure_001 is valid [now-2h, now+6h]; each request supplies
	// only one bound, so the invalid pair is only visible after merging.
	cases := []struct {
		name string
		body models.UpdateRoadClosureRequest
	}{
		{
			name: "validTo before stored validFrom",
			body: models.UpdateRoadClosureRequest{ValidTo: timePtr(testNow.Add(-3 * time.Hour))},
		},
		{
			name: "validFrom after stored validTo",
			body: models.UpdateRoadClosureRequest{ValidFrom: timePtr(testNow.Add(7 * time.Hour))},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer()
			response := httptest.NewRecorder()
			request := authorityRequest(http.MethodPatch, "/api/v1/road-closures/road_closure_001", jsonBody(tc.body))
			request.SetPathValue("id", "road_closure_001")

			srv.updateRoadClosureHandler(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
			}
		})
	}
}

func TestListRoadClosuresNearbyMatchesSegmentMidpoint(t *testing.T) {
	srv := newTestServer()
	// Seed road_closure_001 runs (-0.205,5.570)-(-0.190,5.580); this point is
	// on the segment midpoint but ~1 km from either vertex.
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?lat=5.575&lng=-0.1975&radius=500", nil)
	response := httptest.NewRecorder()

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 1 || payload.Closures[0].ID != "road_closure_001" {
		t.Fatalf("expected road_closure_001 near its segment midpoint, got %#v", payload.Closures)
	}
}

func TestListRoadClosuresBBoxMatchesCrossingSegment(t *testing.T) {
	srv := newTestServer()
	// The box covers only the midpoint of road_closure_001's segment; both
	// vertices lie outside it.
	request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?bbox=-0.200,5.572,-0.195,5.578", nil)
	response := httptest.NewRecorder()

	srv.listRoadClosuresHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.RoadClosureListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Closures) != 1 || payload.Closures[0].ID != "road_closure_001" {
		t.Fatalf("expected road_closure_001 crossing the bbox, got %#v", payload.Closures)
	}
}

func TestListRoadClosuresRejectsInvalidBBox(t *testing.T) {
	cases := []struct {
		name string
		bbox string
	}{
		{"min lng above max lng", "-0.15,5.50,-0.30,5.60"},
		{"min lat above max lat", "-0.30,5.60,-0.15,5.50"},
		{"latitude out of WGS84 range", "-0.30,95.0,-0.15,96.0"},
		{"longitude out of WGS84 range", "-190.0,5.50,-0.15,5.60"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/road-closures?bbox="+tc.bbox, nil)
			response := httptest.NewRecorder()

			srv.listRoadClosuresHandler(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
			}
		})
	}
}

func TestCreateRoadClosureWithBearerToken(t *testing.T) {
	srv := newTokenTestServer()
	body := models.CreateRoadClosureRequest{
		RoadName: "Token Road",
		Status:   "active",
		Geometry: models.LineStringGeometry{
			Type:        "LineString",
			Coordinates: [][]float64{{-0.210, 5.550}, {-0.200, 5.552}},
		},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, models.TokenClaims{
		UserID:    "usr_verified_officer",
		UserType:  "agency",
		Role:      "district_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000204",
		MFA:       true,
		ExpiresAt: testNow.Add(time.Hour).Unix(),
	}))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.RoadClosureResponse
	decodeResponse(t, response, &payload)
	if payload.Closure.CreatedBy != "usr_verified_officer" {
		t.Fatalf("expected actor from verified claims, got %#v", payload.Closure.CreatedBy)
	}
}

func TestCreateRoadClosureIgnoresMockHeadersWhenDisabled(t *testing.T) {
	srv := newTokenTestServer()
	body := models.CreateRoadClosureRequest{
		RoadName: "Token Road",
		Status:   "active",
		Geometry: models.LineStringGeometry{
			Type:        "LineString",
			Coordinates: [][]float64{{-0.210, 5.550}, {-0.200, 5.552}},
		},
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))

	srv.createRoadClosureHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestCreateRoadClosureRejectsForgedAndExpiredTokens(t *testing.T) {
	valid := signTestToken(t, models.TokenClaims{
		UserID: "usr_verified_officer", UserType: "agency", Role: "district_officer",
		AgencyID: "00000000-0000-0000-0000-000000000204", MFA: true, ExpiresAt: testNow.Add(time.Hour).Unix(),
	})
	expired := signTestToken(t, models.TokenClaims{
		UserID: "usr_verified_officer", UserType: "agency", Role: "district_officer",
		AgencyID: "00000000-0000-0000-0000-000000000204", MFA: true, ExpiresAt: testNow.Add(-time.Hour).Unix(),
	})
	cases := []struct {
		name  string
		token string
	}{
		{"tampered signature", valid[:len(valid)-2] + "zz"},
		{"wrong prefix", "fake." + valid[len("nadaa."):]},
		{"expired", expired},
		{"malformed", "not-a-token"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTokenTestServer()
			body := models.CreateRoadClosureRequest{
				RoadName: "Token Road",
				Status:   "active",
				Geometry: models.LineStringGeometry{
					Type:        "LineString",
					Coordinates: [][]float64{{-0.210, 5.550}, {-0.200, 5.552}},
				},
			}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/road-closures", jsonBody(body))
			request.Header.Set("Authorization", "Bearer "+tc.token)

			srv.createRoadClosureHandler(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
			}
		})
	}
}

func timePtr(value time.Time) *time.Time {
	return &value
}

// signTestToken signs claims with the test secret using the same
// nadaa.<payload>.<sig> HMAC-SHA256 scheme as auth-service.
func signTestToken(t *testing.T, claims models.TokenClaims) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(testTokenSecret))
	mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
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
	request.Header.Set("X-NADAA-Actor-ID", "usr_road_closure_officer")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-road-closure")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
