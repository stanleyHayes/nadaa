package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/school-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/school-service/internal/store"
)

func newTestServer() *Server {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8097", RiskServiceURL: "http://risk.test", AllowedOrigins: nil}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return now }, cfg)
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
