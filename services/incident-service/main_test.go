package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{
		store:       newMemoryStore(),
		rateLimiter: newRateLimiter(100, time.Minute, func() time.Time { return now }),
		now:         func() time.Time { return now },
	}
}

func validIncidentRequest() createIncidentRequest {
	return createIncidentRequest{
		Type:               "flood",
		Description:        "Road is flooded and vehicles are trapped",
		Location:           coordinates{Lat: 5.579, Lng: -0.212},
		PeopleAffected:     12,
		InjuriesReported:   false,
		Urgency:            "high",
		Anonymous:          false,
		ContactPermission:  true,
		AccessibilityNeeds: "Elderly person needs evacuation support",
		Media:              nil,
		Reporter:           &reporterRef{UserID: "usr_001", Phone: "+233200000000"},
	}
}

func validMediaUploadRequest() initiateMediaUploadRequest {
	return initiateMediaUploadRequest{
		Purpose:     "incident_media",
		FileName:    "flooded-road.jpg",
		ContentType: "image/jpeg",
		SizeBytes:   820000,
		UploadedBy:  "usr_001",
	}
}

func TestCreateIncident(t *testing.T) {
	srv := newTestServer()
	mediaID := initiateMediaUpload(t, srv)
	body := validIncidentRequest()
	body.Media = []string{mediaID}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body))

	srv.createIncidentHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload createIncidentResponse
	decodeResponse(t, response, &payload)

	if payload.ID == "" || payload.Reference != "INC-000001" {
		t.Fatalf("expected incident id and reference, got %#v", payload)
	}
	if payload.Status != "reported" || payload.Severity != "high" {
		t.Fatalf("expected reported high incident, got %#v", payload)
	}

	media := srv.store.listMedia()
	if len(media) != 1 || media[0].Status != "linked" || media[0].IncidentID != payload.ID {
		t.Fatalf("expected media to be linked to incident, got %#v", media)
	}
}

func TestInitiateMediaUpload(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(validMediaUploadRequest()))

	srv.initiateMediaUploadHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload mediaUploadResponse
	decodeResponse(t, response, &payload)
	if payload.MediaID == "" || payload.Method != "PUT" || payload.Access != "private" {
		t.Fatalf("expected private media upload response, got %#v", payload)
	}
	if payload.MaxSizeBytes != allowedMediaTypes["image/jpeg"] {
		t.Fatalf("expected max size for image/jpeg, got %d", payload.MaxSizeBytes)
	}
}

func TestInitiateMediaUploadRejectsUnsupportedType(t *testing.T) {
	srv := newTestServer()
	body := validMediaUploadRequest()
	body.ContentType = "application/pdf"

	response := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestInitiateMediaUploadRejectsOversizedFile(t *testing.T) {
	srv := newTestServer()
	body := validMediaUploadRequest()
	body.SizeBytes = allowedMediaTypes["image/jpeg"] + 1

	response := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateAnonymousIncidentHidesReporter(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Anonymous = true
	body.ContactPermission = false

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, response.Code)
	}

	incidents := srv.store.listIncidents()
	if len(incidents) != 1 {
		t.Fatalf("expected one incident, got %d", len(incidents))
	}
	if incidents[0].ReportedBy != nil {
		t.Fatalf("expected anonymous report to hide reporter, got %#v", incidents[0].ReportedBy)
	}
}

func TestCreateIncidentRejectsInvalidLocation(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Location = coordinates{Lat: 100, Lng: -0.212}

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateIncidentRejectsUnsupportedHazard(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Type = "volcano"

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateIncidentRejectsUnknownMedia(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Media = []string{"media_missing"}

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCreateIncidentRejectsAlreadyLinkedMedia(t *testing.T) {
	srv := newTestServer()
	mediaID := initiateMediaUpload(t, srv)
	body := validIncidentRequest()
	body.Media = []string{mediaID}

	first := httptest.NewRecorder()
	srv.createIncidentHandler(first, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first incident status %d, got %d", http.StatusCreated, first.Code)
	}

	second := httptest.NewRecorder()
	srv.createIncidentHandler(second, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if second.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, second.Code)
	}
}

func TestLifeThreateningIncidentIsPriorityReview(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Urgency = "life_threatening"
	body.InjuriesReported = true

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload createIncidentResponse
	decodeResponse(t, response, &payload)
	if !payload.PriorityReview || payload.Severity != "emergency" {
		t.Fatalf("expected emergency priority review, got %#v", payload)
	}
}

func TestRateLimit(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := &server{
		store:       newMemoryStore(),
		rateLimiter: newRateLimiter(1, time.Minute, func() time.Time { return now }),
		now:         func() time.Time { return now },
	}

	first := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(validIncidentRequest()))
	request.RemoteAddr = "192.0.2.1:1111"
	srv.createIncidentHandler(first, request)

	second := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(validIncidentRequest()))
	request.RemoteAddr = "192.0.2.1:2222"
	srv.createIncidentHandler(second, request)

	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, second.Code)
	}
}

func TestListIncidents(t *testing.T) {
	srv := newTestServer()
	create := httptest.NewRecorder()
	srv.createIncidentHandler(create, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(validIncidentRequest())))

	response := httptest.NewRecorder()
	srv.listIncidentsHandler(response, httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload incidentListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Incidents) != 1 {
		t.Fatalf("expected one incident, got %#v", payload)
	}
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

func initiateMediaUpload(t *testing.T, srv *server) string {
	t.Helper()

	response := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(validMediaUploadRequest())))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected media upload status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload mediaUploadResponse
	decodeResponse(t, response, &payload)
	return payload.MediaID
}
