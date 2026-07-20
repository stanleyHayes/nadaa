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

	"github.com/stanleyHayes/nadaa/services/school-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/store"
)

const testTokenSecret = "test-token-secret-for-school-service-tests"

func newTestServer() *Server {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8097", TokenSecret: testTokenSecret, AllowMockActors: true, AllowedOrigins: nil}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
}

// newTokenOnlyServer builds a server with mock actor headers disabled, so only
// verified bearer tokens establish authority context.
func newTokenOnlyServer() *Server {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8097", TokenSecret: testTokenSecret, AllowMockActors: false, AllowedOrigins: nil}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
}

func signTestToken(t *testing.T, secret string, claims tokenClaims) string {
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

func districtOfficerTokenClaims() tokenClaims {
	return tokenClaims{
		UserID:    "usr_district_officer_001",
		UserType:  "agency",
		Role:      "district_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000204",
		District:  "accra metropolitan",
		MFA:       true,
		ExpiresAt: time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC).Unix(),
	}
}

func applyDistrictOfficerHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_district_officer_001")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Actor-District", "accra metropolitan")
}

func applySystemAdminHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_system_admin")
	request.Header.Set("X-NADAA-Actor-Role", "system_admin")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000001")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func decodeResponse[T any](t *testing.T, response *httptest.ResponseRecorder, target *T) {
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

func TestListSchoolsRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListSchoolsDistrictScope(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applyDistrictOfficerHeaders(request)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.SchoolListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Schools) != 2 {
		t.Fatalf("expected 2 Accra schools, got %d", len(payload.Schools))
	}
}

func TestListSchoolsAdminSeesAll(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applySystemAdminHeaders(request)

	srv.Routes().ServeHTTP(response, request)

	var payload models.SchoolListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Schools) != 3 {
		t.Fatalf("expected 3 schools for admin, got %d", len(payload.Schools))
	}
}

func TestGetSchoolOutsideDistrictDenied(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools/school_002", nil)
	applyDistrictOfficerHeaders(request)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestCreateAndGetSchool(t *testing.T) {
	srv := newTestServer()
	body := models.CreateSchoolRequest{
		Name:              "New Accra School",
		Location:          models.Coordinates{Lat: 5.55, Lng: -0.19},
		Region:            "Greater Accra",
		District:          "Accra Metropolitan",
		StudentPopulation: 300,
		EmergencyContacts: []models.EmergencyContact{
			{Name: "Head Teacher", Role: "headteacher", Phone: "+233200000999", IsPrimary: true},
		},
		Hazards: []string{"flood"},
		EvacuationPoints: []models.EvacuationPoint{
			{Label: "Assembly ground", Location: models.Coordinates{Lat: 5.551, Lng: -0.191}, Capacity: 350},
		},
	}

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/schools", jsonBody(body))
	applyDistrictOfficerHeaders(createRequest)
	srv.Routes().ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.SchoolProfile
	decodeResponse(t, createResponse, &created)
	if created.ID == "" {
		t.Fatal("expected created school to have an id")
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/schools/"+created.ID, nil)
	applyDistrictOfficerHeaders(getRequest)
	srv.Routes().ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	var detail models.SchoolDetailResponse
	decodeResponse(t, getResponse, &detail)
	if detail.School.Name != body.Name {
		t.Fatalf("expected name %s, got %s", body.Name, detail.School.Name)
	}
}

func TestCreateSchoolRejectsInvalidEvacuationPoint(t *testing.T) {
	cases := []struct {
		name   string
		points []models.EvacuationPoint
	}{
		{
			name:   "out-of-range coordinates",
			points: []models.EvacuationPoint{{Label: "Field", Location: models.Coordinates{Lat: 999, Lng: -0.19}, Capacity: 100}},
		},
		{
			name:   "negative capacity",
			points: []models.EvacuationPoint{{Label: "Field", Location: models.Coordinates{Lat: 5.551, Lng: -0.191}, Capacity: -50}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer()
			body := models.CreateSchoolRequest{
				Name:             "Evac Point School",
				Location:         models.Coordinates{Lat: 5.55, Lng: -0.19},
				Region:           "Greater Accra",
				District:         "Accra Metropolitan",
				EvacuationPoints: tc.points,
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/schools", jsonBody(body))
			applyDistrictOfficerHeaders(request)
			srv.Routes().ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
			}
			var payload models.APIError
			decodeResponse(t, response, &payload)
			if payload.Error.Code != "invalid_evacuation_point" {
				t.Fatalf("expected invalid_evacuation_point, got %q", payload.Error.Code)
			}
		})
	}
}

func TestUpdateSchoolRejectsInvalidEvacuationPoint(t *testing.T) {
	srv := newTestServer()
	update := models.UpdateSchoolRequest{
		EvacuationPoints: []models.EvacuationPoint{
			{Label: "Assembly", Location: models.Coordinates{Lat: 5.551, Lng: -0.191}, Capacity: -1},
		},
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/schools/school_001", jsonBody(update))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_evacuation_point" {
		t.Fatalf("expected invalid_evacuation_point, got %q", payload.Error.Code)
	}
}

func TestUpdateSchool(t *testing.T) {
	srv := newTestServer()
	update := models.UpdateSchoolRequest{
		StudentPopulation: intPtr(999),
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/schools/school_001", jsonBody(update))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var school models.SchoolProfile
	decodeResponse(t, response, &school)
	if school.StudentPopulation != 999 {
		t.Fatalf("expected student population 999, got %d", school.StudentPopulation)
	}
}

func TestCreateDrill(t *testing.T) {
	srv := newTestServer()
	body := models.CreateDrillRequest{
		Date:         time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		Type:         "flood",
		Participants: 600,
		Notes:        "Wet season drill.",
		Completed:    true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/drills", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var drill models.DrillRecord
	decodeResponse(t, response, &drill)
	if drill.SchoolID != "school_001" {
		t.Fatalf("expected schoolId school_001, got %s", drill.SchoolID)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/schools/school_001/drills", nil)
	applyDistrictOfficerHeaders(listRequest)
	srv.Routes().ServeHTTP(listResponse, listRequest)
	var list models.DrillListResponse
	decodeResponse(t, listResponse, &list)
	if len(list.Drills) != 3 {
		t.Fatalf("expected 3 drills, got %d", len(list.Drills))
	}
}

func TestCreateReadinessCheck(t *testing.T) {
	srv := newTestServer()
	body := models.CreateReadinessRequest{
		CheckDate:     time.Date(2026, 7, 9, 9, 0, 0, 0, time.UTC),
		RiskLevel:     "high",
		AreaRiskRef:   "risk_accra_north_002",
		OverallStatus: "ready",
		ChecklistItems: []models.ChecklistItem{
			{Label: "Emergency contacts updated", Checked: true, Category: "admin"},
		},
		Notes: "All set for rainy season.",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/readiness", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var check models.ReadinessCheck
	decodeResponse(t, response, &check)
	if check.OverallStatus != "ready" {
		t.Fatalf("expected status ready, got %s", check.OverallStatus)
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/schools/school_001/readiness", nil)
	applyDistrictOfficerHeaders(getRequest)
	srv.Routes().ServeHTTP(getResponse, getRequest)
	var readiness models.ReadinessResponse
	decodeResponse(t, getResponse, &readiness)
	if readiness.Readiness == nil || readiness.Readiness.OverallStatus != "ready" {
		t.Fatalf("expected latest readiness to be ready, got %#v", readiness.Readiness)
	}
}

func TestCreateSchoolOutsideDistrictDenied(t *testing.T) {
	srv := newTestServer()
	body := models.CreateSchoolRequest{
		Name:              "Tema School",
		Location:          models.Coordinates{Lat: 5.64, Lng: -0.03},
		Region:            "Greater Accra",
		District:          "Tema Metropolitan",
		StudentPopulation: 200,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func intPtr(value int) *int {
	return &value
}

func applyAgencyViewerHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_agency_viewer_001")
	request.Header.Set("X-NADAA-Actor-Role", "agency_viewer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000205")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Actor-District", "accra metropolitan")
}

func applyNadmoOfficerHeaders(request *http.Request) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_nadmo_officer_001")
	request.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000002")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
}

func TestListSchoolsWithBearerToken(t *testing.T) {
	srv := newTokenOnlyServer()
	token := signTestToken(t, testTokenSecret, districtOfficerTokenClaims())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.SchoolListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Schools) != 2 {
		t.Fatalf("expected 2 Accra schools for district-scoped token, got %d", len(payload.Schools))
	}
}

func TestBearerTokenOverridesForgedActorHeaders(t *testing.T) {
	srv := newTokenOnlyServer()
	token := signTestToken(t, testTokenSecret, districtOfficerTokenClaims())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applySystemAdminHeaders(request)
	request.Header.Set("Authorization", "Bearer "+token)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.SchoolListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Schools) != 2 {
		t.Fatalf("expected forged system_admin headers to be ignored and 2 district schools returned, got %d", len(payload.Schools))
	}
}

func TestMockActorHeadersRejectedWhenDisabled(t *testing.T) {
	srv := newTokenOnlyServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applyDistrictOfficerHeaders(request)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestExpiredBearerTokenRejected(t *testing.T) {
	srv := newTokenOnlyServer()
	claims := districtOfficerTokenClaims()
	claims.ExpiresAt = time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC).Unix()
	token := signTestToken(t, testTokenSecret, claims)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestForgedBearerTokenRejected(t *testing.T) {
	srv := newTokenOnlyServer()
	token := signTestToken(t, "a-different-secret", districtOfficerTokenClaims())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestEmptyTokenSecretRejectsBearerToken(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8097", TokenSecret: "", AllowMockActors: false, AllowedOrigins: nil}
	srv := NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
	token := signTestToken(t, testTokenSecret, districtOfficerTokenClaims())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAgencyViewerCanReadButNotWrite(t *testing.T) {
	srv := newTestServer()

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applyAgencyViewerHeaders(listRequest)
	srv.Routes().ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}

	createResponse := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/schools", jsonBody(models.CreateSchoolRequest{
		Name:              "Viewer School",
		Location:          models.Coordinates{Lat: 5.55, Lng: -0.19},
		Region:            "Greater Accra",
		District:          "Accra Metropolitan",
		StudentPopulation: 100,
	}))
	applyAgencyViewerHeaders(createRequest)
	srv.Routes().ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, createResponse.Code, createResponse.Body.String())
	}

	drillResponse := httptest.NewRecorder()
	drillRequest := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/drills", jsonBody(models.CreateDrillRequest{
		Date:         time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		Type:         "fire",
		Participants: 100,
		Completed:    true,
	}))
	applyAgencyViewerHeaders(drillRequest)
	srv.Routes().ServeHTTP(drillResponse, drillRequest)
	if drillResponse.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, drillResponse.Code, drillResponse.Body.String())
	}
}

func TestDistrictOfficerWithoutDistrictDenied(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applyDistrictOfficerHeaders(request)
	request.Header.Del("X-NADAA-Actor-District")

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "district_scope_required" {
		t.Fatalf("expected error code district_scope_required, got %s", payload.Error.Code)
	}
}

func TestNadmoOfficerUnscopedWithoutDistrict(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/schools", nil)
	applyNadmoOfficerHeaders(request)

	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.SchoolListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Schools) != 3 {
		t.Fatalf("expected 3 schools for unscoped nadmo_officer, got %d", len(payload.Schools))
	}
}

func TestCreateDrillRejectsFutureDate(t *testing.T) {
	srv := newTestServer()
	body := models.CreateDrillRequest{
		Date:         time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC),
		Type:         "fire",
		Participants: 100,
		Completed:    true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/drills", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_date" {
		t.Fatalf("expected error code invalid_date, got %s", payload.Error.Code)
	}
}

func TestCreateDrillRejectsZeroDate(t *testing.T) {
	srv := newTestServer()
	body := models.CreateDrillRequest{
		Type:         "fire",
		Participants: 100,
		Completed:    true,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/drills", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_date" {
		t.Fatalf("expected error code invalid_date, got %s", payload.Error.Code)
	}
}

func TestCreateDrillAllowsDateWithinSkewTolerance(t *testing.T) {
	srv := newTestServer()
	body := models.CreateDrillRequest{
		Date:         time.Date(2026, 7, 10, 18, 0, 0, 0, time.UTC),
		Type:         "fire",
		Participants: 100,
		Completed:    false,
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/drills", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
}

func TestCreateReadinessRejectsFutureCheckDate(t *testing.T) {
	srv := newTestServer()
	body := models.CreateReadinessRequest{
		CheckDate:     time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC),
		RiskLevel:     "high",
		OverallStatus: "ready",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/readiness", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_check_date" {
		t.Fatalf("expected error code invalid_check_date, got %s", payload.Error.Code)
	}
}

func TestCreateReadinessRejectsZeroCheckDate(t *testing.T) {
	srv := newTestServer()
	body := models.CreateReadinessRequest{
		RiskLevel:     "high",
		OverallStatus: "ready",
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/schools/school_001/readiness", jsonBody(body))
	applyDistrictOfficerHeaders(request)
	srv.Routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_check_date" {
		t.Fatalf("expected error code invalid_check_date, got %s", payload.Error.Code)
	}
}
