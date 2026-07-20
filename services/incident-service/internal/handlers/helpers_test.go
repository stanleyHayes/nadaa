package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
)

const testTokenSecret = "test-incident-token-secret"

// testInternalServiceToken is the shared service-to-service token used by the
// X-NADAA-Service-Token tests.
const testInternalServiceToken = "test-internal-service-token"

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{RateLimit: 100, RateWindowSecs: 60, TokenSecret: testTokenSecret, AllowMockActors: true}
	return NewServer(store.NewMemoryStore(), func() time.Time { return now }, cfg)
}

// newTokenOnlyTestServer builds a server with mock actor headers disabled so
// the signed-token path is the only way to reach authority endpoints.
func newTokenOnlyTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{RateLimit: 100, RateWindowSecs: 60, TokenSecret: testTokenSecret}
	return NewServer(store.NewMemoryStore(), func() time.Time { return now }, cfg)
}

// newServiceTokenTestServer configures the internal service-to-service token
// so X-NADAA-Service-Token credentials are honored.
func newServiceTokenTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{RateLimit: 100, RateWindowSecs: 60, TokenSecret: testTokenSecret, InternalServiceToken: testInternalServiceToken}
	return NewServer(store.NewMemoryStore(), func() time.Time { return now }, cfg)
}

func testAuthorityClaims() tokenClaims {
	return tokenClaims{
		UserID:    "usr_dispatcher_001",
		UserType:  "agency",
		Role:      "nadmo_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		MFA:       true,
		ExpiresAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
}

func signTestToken(secret string, claims tokenClaims) string {
	payload, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// tokenRequest builds a request carrying a signed Bearer token instead of
// legacy actor headers.
func tokenRequest(method string, target string, body *bytes.Reader, claims tokenClaims) *http.Request {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = body
	}
	request := httptest.NewRequest(method, target, reader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+signTestToken(testTokenSecret, claims))
	return request
}

func validIncidentRequest() models.CreateIncidentRequest {
	return models.CreateIncidentRequest{
		Type:               "flood",
		Description:        "Road is flooded and vehicles are trapped",
		Location:           &models.Coordinates{Lat: 5.579, Lng: -0.212},
		PeopleAffected:     12,
		InjuriesReported:   false,
		Urgency:            "high",
		Anonymous:          false,
		ContactPermission:  true,
		AccessibilityNeeds: "Elderly person needs evacuation support",
		Media:              nil,
		Reporter:           &models.ReporterRef{UserID: "usr_001", Phone: "+233200000000"},
	}
}

func validMediaUploadRequest() models.InitiateMediaUploadRequest {
	return models.InitiateMediaUploadRequest{
		Purpose:     "incident_media",
		FileName:    "flooded-road.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   820000,
		UploadedBy:  "usr_001",
	}
}

func validAssignmentRequest() models.AssignmentRequest {
	return models.AssignmentRequest{
		AgencyID:      "00000000-0000-0000-0000-000000000201",
		AgencyName:    "Ghana National Fire Service",
		AgencyType:    "fire",
		Priority:      "high",
		Instructions:  "Dispatch rescue team to flooded road.",
		ResponderLead: "Station Officer Mensah",
	}
}

func validVolunteerRegistrationRequest() models.RegisterVolunteerRequest {
	return models.RegisterVolunteerRequest{
		CitizenUserID:      "usr_volunteer_001",
		Name:               "Ama Volunteer",
		Phone:              "+233200000111",
		Region:             "Greater Accra",
		District:           "Accra Metropolitan",
		Community:          "Jamestown",
		Skills:             []string{"first aid", "community alerts"},
		Languages:          []string{"en", "tw"},
		AvailabilityStatus: "available",
	}
}

func validVolunteerTaskRequest(volunteerID string) models.VolunteerTaskRequest {
	return models.VolunteerTaskRequest{
		VolunteerID:   volunteerID,
		Type:          "welfare_check",
		Priority:      "high",
		Instructions:  "Check whether households near the shelter need water or accessible transport. Stay outside unsafe areas.",
		LocationLabel: "Jamestown shelter approach",
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
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = body
	}
	request := httptest.NewRequest(method, target, reader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_dispatcher_001")
	request.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-request-001")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func initiateMediaUpload(t *testing.T, srv *server) string {
	t.Helper()

	response := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(validMediaUploadRequest())))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected media upload status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.MediaUploadResponse
	decodeResponse(t, response, &payload)
	return payload.MediaID
}

func createIncidentForTest(t *testing.T, srv *server, body models.CreateIncidentRequest) models.CreateIncidentResponse {
	t.Helper()

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected incident status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.CreateIncidentResponse
	decodeResponse(t, response, &payload)
	return payload
}

func verifyIncidentForTest(t *testing.T, srv *server, incidentID string) models.IncidentRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/verify",
		jsonBody(models.IncidentWorkflowRequest{Note: "Confirmed by test dispatcher."}),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", incidentID)
	srv.verifyIncidentHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected verify status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	return payload
}

func assignIncidentForTest(t *testing.T, srv *server, incidentID string, body models.AssignmentRequest) models.IncidentRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/assignments",
		jsonBody(body),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", incidentID)
	srv.assignIncidentHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected assignment status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	return payload
}

func mergeIncidentsForTest(t *testing.T, srv *server, incidentID string, body models.MergeIncidentsRequest) models.MergeIncidentsResponse {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/merge",
		jsonBody(body),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", incidentID)
	srv.mergeIncidentHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected merge status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.MergeIncidentsResponse
	decodeResponse(t, response, &payload)
	return payload
}

func reviewAbuseForTest(t *testing.T, srv *server, incidentID string, body models.AbuseReviewRequest) models.IncidentRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/abuse-review",
		jsonBody(body),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", incidentID)
	srv.reviewAbuseHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected abuse review status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	return payload
}

// citizenClaims builds verified citizen token claims for the given user id.
func citizenClaims(userID string) tokenClaims {
	return tokenClaims{
		UserID:    userID,
		UserType:  "citizen",
		Role:      "citizen",
		ExpiresAt: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
}

func registerVolunteerForTest(t *testing.T, srv *server, body models.RegisterVolunteerRequest) models.VolunteerProfile {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(body), citizenClaims(body.CitizenUserID))
	srv.registerVolunteerHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected volunteer registration status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.VolunteerProfileResponse
	decodeResponse(t, response, &payload)
	return payload.Volunteer
}

func verifyVolunteerForTest(t *testing.T, srv *server, volunteerID string, body models.VerifyVolunteerRequest) models.VolunteerProfile {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/volunteers/"+volunteerID+"/verify",
		jsonBody(body),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", volunteerID)
	srv.verifyVolunteerHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected volunteer verify status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.VolunteerProfileResponse
	decodeResponse(t, response, &payload)
	return payload.Volunteer
}

func assignVolunteerTaskForTest(t *testing.T, srv *server, incidentID string, body models.VolunteerTaskRequest) models.VolunteerTaskRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/volunteer-tasks",
		jsonBody(body),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", incidentID)
	srv.assignVolunteerTaskHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected volunteer task status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.VolunteerTaskRecord
	decodeResponse(t, response, &payload)
	return payload
}

func updateVolunteerTaskStatusForTest(t *testing.T, srv *server, taskID string, body models.VolunteerTaskStatusRequest) models.VolunteerTaskRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/volunteer-tasks/"+taskID+"/status", jsonBody(body))
	request.SetPathValue("id", taskID)
	srv.updateVolunteerTaskStatusHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected volunteer status update status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.VolunteerTaskRecord
	decodeResponse(t, response, &payload)
	return payload
}

func submitVolunteerObservationForTest(t *testing.T, srv *server, taskID string, body models.VolunteerObservationRequest) models.VolunteerTaskRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/volunteer-tasks/"+taskID+"/observations", jsonBody(body))
	request.SetPathValue("id", taskID)
	srv.submitVolunteerObservationHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected volunteer observation status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.VolunteerTaskRecord
	decodeResponse(t, response, &payload)
	return payload
}

func containsString(values []string, needle string) bool {
	return slices.Contains(values, needle)
}

func containsAuditAction(values []models.AuditEvent, needle string) bool {
	for _, value := range values {
		if value.Action == needle {
			return true
		}
	}
	return false
}

func containsAbuseSignal(values []models.AbuseSignal, needle string) bool {
	for _, value := range values {
		if value.Code == needle {
			return true
		}
	}
	return false
}

func containsTimelineType(values []models.TimelineEvent, needle string) bool {
	for _, value := range values {
		if value.Type == needle {
			return true
		}
	}
	return false
}
