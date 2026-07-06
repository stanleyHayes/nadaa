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

func validAssignmentRequest() assignmentRequest {
	return assignmentRequest{
		AgencyID:      "00000000-0000-0000-0000-000000000201",
		AgencyName:    "Ghana National Fire Service",
		AgencyType:    "fire",
		Priority:      "high",
		Instructions:  "Dispatch rescue team to flooded road.",
		ResponderLead: "Station Officer Mensah",
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

	incidents := srv.store.listIncidents("")
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

func TestDuplicateCandidatesIncludeSameLocationReport(t *testing.T) {
	srv := newTestServer()
	first := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &reporterRef{UserID: "usr_002", Phone: "+233200000002"}
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 1 {
		t.Fatalf("expected one duplicate candidate, got %#v", second.DuplicateCandidates)
	}
	candidate := second.DuplicateCandidates[0]
	if candidate.IncidentID != first.ID || candidate.Reference != first.Reference {
		t.Fatalf("expected first incident as duplicate candidate, got %#v", candidate)
	}
	if candidate.Score < duplicateMinimumScore || candidate.DistanceMeters != 0 || candidate.MinutesApart != 0 {
		t.Fatalf("expected strong same-location candidate, got %#v", candidate)
	}
	if !containsString(candidate.Reasons, "same_hazard") || !containsString(candidate.Reasons, "similar_description") {
		t.Fatalf("expected duplicate reasons to include same hazard and similar description, got %#v", candidate.Reasons)
	}

	incidents := srv.store.listIncidents("")
	var original incidentRecord
	for _, incident := range incidents {
		if incident.ID == first.ID {
			original = incident
			break
		}
	}
	if len(original.DuplicateCandidates) != 1 || original.DuplicateCandidates[0].IncidentID != second.ID {
		t.Fatalf("expected reverse duplicate candidate on original incident, got %#v", original.DuplicateCandidates)
	}
}

func TestDuplicateCandidatesIncludeNearbyReport(t *testing.T) {
	srv := newTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Flood water is rising near the same trapped vehicles"
	body.Location = coordinates{Lat: 5.581, Lng: -0.211}
	body.Reporter = &reporterRef{UserID: "usr_003", Phone: "+233200000003"}
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 1 {
		t.Fatalf("expected nearby duplicate candidate, got %#v", second.DuplicateCandidates)
	}
	candidate := second.DuplicateCandidates[0]
	if candidate.DistanceMeters <= 0 || candidate.DistanceMeters > duplicateDistanceMeters {
		t.Fatalf("expected candidate within dedupe distance, got %#v", candidate)
	}
	if candidate.Score < duplicateMinimumScore {
		t.Fatalf("expected candidate score over threshold, got %#v", candidate)
	}
}

func TestDuplicateCandidatesIgnoreDifferentHazard(t *testing.T) {
	srv := newTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Type = "fire"
	body.Description = "Vehicles are trapped beside a roadside fire"
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 0 {
		t.Fatalf("expected different hazard not to be a duplicate, got %#v", second.DuplicateCandidates)
	}
}

func TestDuplicateCandidatesIgnoreReportsOutsideTimeWindow(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := &server{
		store:       newMemoryStore(),
		rateLimiter: newRateLimiter(100, time.Minute, func() time.Time { return now }),
		now:         func() time.Time { return now },
	}

	createIncidentForTest(t, srv, validIncidentRequest())

	now = now.Add(duplicateReviewWindow + time.Minute)
	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 0 {
		t.Fatalf("expected old report outside duplicate window to be ignored, got %#v", second.DuplicateCandidates)
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

func TestVerifyIncidentAuditsStatusChange(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/verify",
		jsonBody(incidentWorkflowRequest{Note: "Confirmed with caller and duplicate report."}),
	)
	request.SetPathValue("id", incident.ID)
	srv.verifyIncidentHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload incidentRecord
	decodeResponse(t, response, &payload)
	if payload.Status != "verified" || payload.VerifiedBy != "usr_dispatcher_001" || payload.VerifiedAt == nil {
		t.Fatalf("expected verified incident with verifier metadata, got %#v", payload)
	}

	logs := srv.store.listAudit(10)
	if len(logs) != 1 {
		t.Fatalf("expected one audit event, got %#v", logs)
	}
	if logs[0].Action != "incident.verified" || logs[0].Before["status"] != "reported" || logs[0].After["status"] != "verified" {
		t.Fatalf("expected incident verified audit snapshot, got %#v", logs[0])
	}
}

func TestAssignIncidentCreatesAssignmentTimelineAndAudit(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, incident.ID)

	payload := assignIncidentForTest(t, srv, incident.ID, validAssignmentRequest())

	if payload.Status != "assigned" || len(payload.Assignments) != 1 {
		t.Fatalf("expected assigned incident with one assignment, got %#v", payload)
	}
	assignment := payload.Assignments[0]
	if assignment.AgencyID != "00000000-0000-0000-0000-000000000201" || assignment.Priority != "high" || assignment.AssignedBy != "usr_dispatcher_001" {
		t.Fatalf("expected assignment metadata to be stored, got %#v", assignment)
	}
	if !containsTimelineType(payload.Timeline, "incident.reported") ||
		!containsTimelineType(payload.Timeline, "incident.verified") ||
		!containsTimelineType(payload.Timeline, "incident.assigned") {
		t.Fatalf("expected report, verification, and assignment timeline events, got %#v", payload.Timeline)
	}

	logs := srv.store.listAudit(10)
	if len(logs) != 2 {
		t.Fatalf("expected verification and assignment audit events, got %#v", logs)
	}
	if logs[0].Action != "incident.assigned" || logs[0].Before["status"] != "verified" || logs[0].After["status"] != "assigned" {
		t.Fatalf("expected latest audit to capture assignment, got %#v", logs[0])
	}
	if logs[0].After["assignmentCount"] != 1 {
		t.Fatalf("expected assignment count in audit snapshot, got %#v", logs[0].After)
	}
}

func TestAssignIncidentRequiresVerifiedIncident(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/assignments",
		jsonBody(validAssignmentRequest()),
	)
	request.SetPathValue("id", incident.ID)
	srv.assignIncidentHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	incidents := srv.store.listIncidents("")
	if len(incidents[0].Assignments) != 0 || incidents[0].Status != "reported" {
		t.Fatalf("expected reported incident to remain unassigned, got %#v", incidents[0])
	}
}

func TestAgencyAdminCanAssignOnlyOwnAgency(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, incident.ID)

	forbidden := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/assignments",
		jsonBody(validAssignmentRequest()),
	)
	request.Header.Set("X-NADAA-Actor-Role", "agency_admin")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	request.SetPathValue("id", incident.ID)
	srv.assignIncidentHandler(forbidden, request)

	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("expected cross-agency assignment status %d, got %d", http.StatusForbidden, forbidden.Code)
	}

	ownAgency := validAssignmentRequest()
	ownAgency.AgencyID = "00000000-0000-0000-0000-000000000101"
	ownAgency.AgencyName = "NADMO Accra Metro"
	ownAgency.AgencyType = "nadmo"

	response := httptest.NewRecorder()
	request = authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/assignments",
		jsonBody(ownAgency),
	)
	request.Header.Set("X-NADAA-Actor-Role", "agency_admin")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	request.SetPathValue("id", incident.ID)
	srv.assignIncidentHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected own-agency assignment status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
}

func TestListIncidentsAssignedToMe(t *testing.T) {
	srv := newTestServer()
	first := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, first.ID)

	ownAgency := validAssignmentRequest()
	ownAgency.AgencyID = "00000000-0000-0000-0000-000000000101"
	ownAgency.AgencyName = "NADMO Accra Metro"
	ownAgency.AgencyType = "nadmo"
	assignIncidentForTest(t, srv, first.ID, ownAgency)

	secondBody := validIncidentRequest()
	secondBody.Description = "Smoke is visible from a roadside electrical kiosk"
	secondBody.Type = "electrical_hazard"
	second := createIncidentForTest(t, srv, secondBody)
	verifyIncidentForTest(t, srv, second.ID)
	otherAgency := validAssignmentRequest()
	otherAgency.AgencyID = "00000000-0000-0000-0000-000000000202"
	assignIncidentForTest(t, srv, second.ID, otherAgency)

	response := httptest.NewRecorder()
	srv.listIncidentsHandler(response, authorityRequest(http.MethodGet, "/api/v1/incidents?assignedToMe=true", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload incidentListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Incidents) != 1 || payload.Incidents[0].ID != first.ID {
		t.Fatalf("expected only own assigned incident, got %#v", payload.Incidents)
	}
}

func TestStatusWorkflowAllowsValidOperationalTransitions(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	for _, nextStatus := range []string{"under_review", "verified", "assigned", "response_en_route", "on_scene", "contained", "recovery_ongoing", "closed"} {
		response := httptest.NewRecorder()
		body := incidentStatusRequest{
			Status: nextStatus,
			Note:   "Operational update from test dispatcher.",
		}
		if nextStatus == "closed" {
			body.ResolutionNotes = "Waters receded and field team closed the incident."
		}
		request := authorityRequest(http.MethodPatch, "/api/v1/incidents/"+incident.ID+"/status", jsonBody(body))
		request.SetPathValue("id", incident.ID)
		srv.updateIncidentStatusHandler(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected status %d for %s, got %d: %s", http.StatusOK, nextStatus, response.Code, response.Body.String())
		}

		var payload incidentRecord
		decodeResponse(t, response, &payload)
		if payload.Status != nextStatus {
			t.Fatalf("expected incident status %s, got %#v", nextStatus, payload)
		}
	}

	logs := srv.store.listAudit(20)
	if len(logs) != 8 {
		t.Fatalf("expected eight audit events, got %d", len(logs))
	}
	if logs[0].Action != "incident.closed" {
		t.Fatalf("expected latest audit event to be closure, got %#v", logs[0])
	}
}

func TestStatusWorkflowRejectsInvalidTransition(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPatch,
		"/api/v1/incidents/"+incident.ID+"/status",
		jsonBody(incidentStatusRequest{
			Status:          "closed",
			Note:            "Trying to close before review.",
			ResolutionNotes: "Closure should not be accepted from reported.",
		}),
	)
	request.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	incidents := srv.store.listIncidents("")
	if incidents[0].Status != "reported" {
		t.Fatalf("expected incident to remain reported, got %#v", incidents[0])
	}
}

func TestStatusWorkflowRequiresResolutionNotesForFalseReport(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	missingNotes := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPatch,
		"/api/v1/incidents/"+incident.ID+"/status",
		jsonBody(incidentStatusRequest{Status: "false_report", Note: "Caller recanted."}),
	)
	request.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(missingNotes, request)

	if missingNotes.Code != http.StatusBadRequest {
		t.Fatalf("expected missing resolution notes status %d, got %d", http.StatusBadRequest, missingNotes.Code)
	}

	response := httptest.NewRecorder()
	request = authorityRequest(
		http.MethodPatch,
		"/api/v1/incidents/"+incident.ID+"/status",
		jsonBody(incidentStatusRequest{
			Status:          "false report",
			Note:            "Caller recanted.",
			ResolutionNotes: "Dispatcher confirmed the location has no active incident.",
		}),
	)
	request.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload incidentRecord
	decodeResponse(t, response, &payload)
	if payload.Status != "false_report" || payload.ResolutionNotes == "" || payload.ClosedAt == nil {
		t.Fatalf("expected false report resolution metadata, got %#v", payload)
	}
}

func TestStatusWorkflowRequiresMFA(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/verify",
		jsonBody(incidentWorkflowRequest{Note: "Missing MFA"}),
	)
	request.Header.Set("X-NADAA-MFA-Completed", "false")
	request.SetPathValue("id", incident.ID)
	srv.verifyIncidentHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestListIncidentAuditRequiresApproverRole(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/verify",
		jsonBody(incidentWorkflowRequest{Note: "Confirmed."}),
	)
	request.SetPathValue("id", incident.ID)
	srv.verifyIncidentHandler(httptest.NewRecorder(), request)

	forbidden := httptest.NewRecorder()
	viewerRequest := authorityRequest(http.MethodGet, "/api/v1/incidents/audit", nil)
	viewerRequest.Header.Set("X-NADAA-Actor-Role", "dispatcher")
	srv.listIncidentAuditHandler(forbidden, viewerRequest)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("expected dispatcher audit read status %d, got %d", http.StatusForbidden, forbidden.Code)
	}

	response := httptest.NewRecorder()
	srv.listIncidentAuditHandler(response, authorityRequest(http.MethodGet, "/api/v1/incidents/audit?limit=1", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload incidentAuditListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Logs) != 1 || payload.Logs[0].Action != "incident.verified" {
		t.Fatalf("expected one incident audit log, got %#v", payload)
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

	var payload mediaUploadResponse
	decodeResponse(t, response, &payload)
	return payload.MediaID
}

func createIncidentForTest(t *testing.T, srv *server, body createIncidentRequest) createIncidentResponse {
	t.Helper()

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected incident status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload createIncidentResponse
	decodeResponse(t, response, &payload)
	return payload
}

func verifyIncidentForTest(t *testing.T, srv *server, incidentID string) incidentRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/verify",
		jsonBody(incidentWorkflowRequest{Note: "Confirmed by test dispatcher."}),
	)
	request.SetPathValue("id", incidentID)
	srv.verifyIncidentHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected verify status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload incidentRecord
	decodeResponse(t, response, &payload)
	return payload
}

func assignIncidentForTest(t *testing.T, srv *server, incidentID string, body assignmentRequest) incidentRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incidentID+"/assignments",
		jsonBody(body),
	)
	request.SetPathValue("id", incidentID)
	srv.assignIncidentHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected assignment status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload incidentRecord
	decodeResponse(t, response, &payload)
	return payload
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func containsTimelineType(values []timelineEvent, needle string) bool {
	for _, value := range values {
		if value.Type == needle {
			return true
		}
	}
	return false
}
