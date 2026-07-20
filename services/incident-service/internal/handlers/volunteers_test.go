package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
)

func TestVolunteerRegistrationVerificationTaskAndObservationTimeline(t *testing.T) {
	srv := newTestServer()
	volunteer := registerVolunteerForTest(t, srv, validVolunteerRegistrationRequest())
	if volunteer.VerificationStatus != "pending" || volunteer.GroupID == "" || len(volunteer.SafetyNotes) == 0 {
		t.Fatalf("expected pending volunteer with group and safety rules, got %#v", volunteer)
	}

	verified := verifyVolunteerForTest(t, srv, volunteer.ID, models.VerifyVolunteerRequest{
		Decision: "verify",
		Note:     "District officer checked ID and community lead reference.",
	})
	if verified.VerificationStatus != "verified" || verified.VerifiedBy != "usr_dispatcher_001" || verified.VerifiedAt == nil {
		t.Fatalf("expected verified volunteer metadata, got %#v", verified)
	}

	incident := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, incident.ID)
	task := assignVolunteerTaskForTest(t, srv, incident.ID, validVolunteerTaskRequest(volunteer.ID))
	if task.Status != "assigned" || task.IncidentReference != incident.Reference || task.VolunteerID != volunteer.ID {
		t.Fatalf("expected assigned volunteer task, got %#v", task)
	}
	if len(task.SafetyRules) == 0 {
		t.Fatalf("expected task safety rules, got %#v", task)
	}

	status := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID:  volunteer.ID,
		Status:       "accepted",
		Note:         "I can check the shelter approach from a safe public road.",
		SafetyStatus: "safe",
		Location:     &models.Coordinates{Lat: 5.56, Lng: -0.2},
	})
	if status.Status != "accepted" || status.AcceptedAt == nil || len(status.Updates) != 1 {
		t.Fatalf("expected accepted volunteer status update, got %#v", status)
	}

	observationMediaID := initiateMediaUpload(t, srv)
	observed := submitVolunteerObservationForTest(t, srv, task.ID, models.VolunteerObservationRequest{
		VolunteerID:         volunteer.ID,
		Observation:         "Water is rising near the footbridge and families are waiting for transport.",
		SafetyStatus:        "needs_authority",
		Location:            &models.Coordinates{Lat: 5.561, Lng: -0.201},
		EscalationRequested: true,
		Media:               []string{observationMediaID},
	})
	if observed.Status != "needs_escalation" || !observed.EscalationRequired || len(observed.Updates) != 2 {
		t.Fatalf("expected escalated volunteer observation, got %#v", observed)
	}
	latestUpdate := observed.Updates[len(observed.Updates)-1]
	if len(latestUpdate.Media) != 1 || latestUpdate.Media[0] != observationMediaID {
		t.Fatalf("expected observation media reference to persist on the task update, got %#v", latestUpdate)
	}

	incidents := srv.store.ListIncidents("")
	var updated models.IncidentRecord
	for _, item := range incidents {
		if item.ID == incident.ID {
			updated = item
			break
		}
	}
	for _, expected := range []string{"incident.volunteer_assigned", "incident.volunteer_status_updated", "incident.volunteer_observation", "incident.volunteer_escalation"} {
		if !containsTimelineType(updated.Timeline, expected) {
			t.Fatalf("expected volunteer timeline event %s, got %#v", expected, updated.Timeline)
		}
	}

	logs := srv.store.ListAudit(10)
	if !containsAuditAction(logs, "incident.volunteer_assigned") || !containsAuditAction(logs, "volunteer_task.assigned") {
		t.Fatalf("expected volunteer assignment audit events, got %#v", logs)
	}
}

func TestVolunteerTaskRequiresVerifiedVolunteer(t *testing.T) {
	srv := newTestServer()
	volunteer := registerVolunteerForTest(t, srv, validVolunteerRegistrationRequest())
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, incident.ID)

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/volunteer-tasks",
		jsonBody(validVolunteerTaskRequest(volunteer.ID)),
	)
	request.SetPathValue("id", incident.ID)
	srv.assignVolunteerTaskHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected unverified volunteer status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestVolunteerTaskRejectsUnsafeInstructions(t *testing.T) {
	srv := newTestServer()
	volunteer := registerVolunteerForTest(t, srv, validVolunteerRegistrationRequest())
	verifyVolunteerForTest(t, srv, volunteer.ID, models.VerifyVolunteerRequest{
		Decision: "verify",
		Note:     "District officer checked ID and community lead reference.",
	})
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, incident.ID)

	body := validVolunteerTaskRequest(volunteer.ID)
	body.Instructions = "Enter floodwater and rescue trapped residents before responders arrive."
	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/volunteer-tasks",
		jsonBody(body),
	)
	request.SetPathValue("id", incident.ID)
	srv.assignVolunteerTaskHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected unsafe instruction status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func assignedVolunteerTaskForTest(t *testing.T, srv *server) (models.VolunteerProfile, models.VolunteerTaskRecord) {
	t.Helper()

	volunteer := registerVolunteerForTest(t, srv, validVolunteerRegistrationRequest())
	verifyVolunteerForTest(t, srv, volunteer.ID, models.VerifyVolunteerRequest{
		Decision: "verify",
		Note:     "District officer checked ID and community lead reference.",
	})
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	verifyIncidentForTest(t, srv, incident.ID)
	task := assignVolunteerTaskForTest(t, srv, incident.ID, validVolunteerTaskRequest(volunteer.ID))
	return volunteer, task
}

func volunteerTaskStatusRequest(t *testing.T, srv *server, taskID string, body any, mutate func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/volunteer-tasks/"+taskID+"/status", jsonBody(body))
	request.Header.Set("Content-Type", "application/json")
	if mutate != nil {
		mutate(request)
	}
	request.SetPathValue("id", taskID)
	srv.updateVolunteerTaskStatusHandler(response, request)
	return response
}

func TestVolunteerTaskStatusTransitionGuard(t *testing.T) {
	srv := newTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	// assigned -> completed skips the operational flow.
	response := volunteerTaskStatusRequest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "completed",
	}, func(r *http.Request) {
		r.Header.Set("X-NADAA-Actor-ID", "usr_dispatcher_001")
		r.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
		r.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
		r.Header.Set("X-NADAA-MFA-Completed", "true")
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected assigned->completed status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	// assigned -> accepted -> on_scene -> completed is the valid chain.
	for _, next := range []string{"accepted", "on_scene", "completed"} {
		updated := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
			VolunteerID: volunteer.ID,
			Status:      next,
		})
		if updated.Status != next {
			t.Fatalf("expected task status %s, got %#v", next, updated)
		}
	}

	// completed is terminal.
	response = volunteerTaskStatusRequest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
	}, func(r *http.Request) {
		r.Header.Set("X-NADAA-Actor-ID", "usr_dispatcher_001")
		r.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
		r.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
		r.Header.Set("X-NADAA-MFA-Completed", "true")
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected completed->accepted status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	final, found := srv.store.VolunteerTaskByID(task.ID)
	if !found || final.Status != "completed" || final.CompletedAt == nil {
		t.Fatalf("expected task to remain completed, got %#v", final)
	}
}

func TestVolunteerTaskCancelledIsTerminal(t *testing.T) {
	srv := newTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	cancelled := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "cancelled",
		Note:        "Volunteer unavailable due to family emergency.",
	})
	if cancelled.Status != "cancelled" {
		t.Fatalf("expected cancelled task, got %#v", cancelled)
	}

	response := volunteerTaskStatusRequest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
	}, func(r *http.Request) {
		r.Header.Set("X-NADAA-Actor-ID", "usr_dispatcher_001")
		r.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
		r.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
		r.Header.Set("X-NADAA-MFA-Completed", "true")
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected cancelled->accepted status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestVolunteerEscalationTimelineEventOnlyOnNewRequest(t *testing.T) {
	srv := newTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	first := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID:  volunteer.ID,
		Status:       "accepted",
		SafetyStatus: "unsafe",
	})
	if !first.EscalationRequired {
		t.Fatalf("expected escalation after unsafe status, got %#v", first)
	}

	second := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID:  volunteer.ID,
		Status:       "on_scene",
		SafetyStatus: "unsafe",
	})
	if !second.EscalationRequired {
		t.Fatalf("expected escalation flag to stay set, got %#v", second)
	}

	incidents := srv.store.ListIncidents("")
	var incident models.IncidentRecord
	for _, item := range incidents {
		if item.ID == task.IncidentID {
			incident = item
			break
		}
	}
	count := 0
	for _, event := range incident.Timeline {
		if event.Type == "incident.volunteer_escalation" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one escalation timeline event, got %d: %#v", count, incident.Timeline)
	}
}

func TestVolunteerTaskEndpointsRequireAuthentication(t *testing.T) {
	srv := newTokenOnlyTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	statusResponse := volunteerTaskStatusRequest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
	}, nil)
	if statusResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated status update %d, got %d: %s", http.StatusUnauthorized, statusResponse.Code, statusResponse.Body.String())
	}

	observationResponse := httptest.NewRecorder()
	observationRequest := httptest.NewRequest(http.MethodPost, "/api/v1/volunteer-tasks/"+task.ID+"/observations", jsonBody(models.VolunteerObservationRequest{
		VolunteerID: volunteer.ID,
		Observation: "Water is rising near the footbridge.",
	}))
	observationRequest.SetPathValue("id", task.ID)
	srv.submitVolunteerObservationHandler(observationResponse, observationRequest)
	if observationResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated observation %d, got %d: %s", http.StatusUnauthorized, observationResponse.Code, observationResponse.Body.String())
	}

	tasksResponse := httptest.NewRecorder()
	tasksRequest := httptest.NewRequest(http.MethodGet, "/api/v1/volunteers/"+volunteer.ID+"/tasks", nil)
	tasksRequest.SetPathValue("id", volunteer.ID)
	srv.listVolunteerTasksHandler(tasksResponse, tasksRequest)
	if tasksResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated task list %d, got %d: %s", http.StatusUnauthorized, tasksResponse.Code, tasksResponse.Body.String())
	}
}

func TestVolunteerTaskEndpointsAllowOwningCitizen(t *testing.T) {
	srv := newTokenOnlyTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	ownerClaims := citizenClaims(volunteer.CitizenUserID)

	tasksResponse := httptest.NewRecorder()
	tasksRequest := tokenRequest(http.MethodGet, "/api/v1/volunteers/"+volunteer.ID+"/tasks", nil, ownerClaims)
	tasksRequest.SetPathValue("id", volunteer.ID)
	srv.listVolunteerTasksHandler(tasksResponse, tasksRequest)
	if tasksResponse.Code != http.StatusOK {
		t.Fatalf("expected owner task list status %d, got %d: %s", http.StatusOK, tasksResponse.Code, tasksResponse.Body.String())
	}

	statusResponse := httptest.NewRecorder()
	statusRequest := tokenRequest(http.MethodPatch, "/api/v1/volunteer-tasks/"+task.ID+"/status", jsonBody(models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
		Note:        "On my way via the safe route.",
	}), ownerClaims)
	statusRequest.SetPathValue("id", task.ID)
	srv.updateVolunteerTaskStatusHandler(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("expected owner status update %d, got %d: %s", http.StatusOK, statusResponse.Code, statusResponse.Body.String())
	}
}

func TestVolunteerTaskEndpointsRejectOtherCitizen(t *testing.T) {
	srv := newTokenOnlyTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	otherClaims := citizenClaims("usr_someone_else")

	tasksResponse := httptest.NewRecorder()
	tasksRequest := tokenRequest(http.MethodGet, "/api/v1/volunteers/"+volunteer.ID+"/tasks", nil, otherClaims)
	tasksRequest.SetPathValue("id", volunteer.ID)
	srv.listVolunteerTasksHandler(tasksResponse, tasksRequest)
	if tasksResponse.Code != http.StatusForbidden {
		t.Fatalf("expected other citizen task list %d, got %d: %s", http.StatusForbidden, tasksResponse.Code, tasksResponse.Body.String())
	}

	statusResponse := httptest.NewRecorder()
	statusRequest := tokenRequest(http.MethodPatch, "/api/v1/volunteer-tasks/"+task.ID+"/status", jsonBody(models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
	}), otherClaims)
	statusRequest.SetPathValue("id", task.ID)
	srv.updateVolunteerTaskStatusHandler(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusForbidden {
		t.Fatalf("expected other citizen status update %d, got %d: %s", http.StatusForbidden, statusResponse.Code, statusResponse.Body.String())
	}
}

func TestVolunteerObservationRejectsUnknownMedia(t *testing.T) {
	srv := newTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/volunteer-tasks/"+task.ID+"/observations", jsonBody(models.VolunteerObservationRequest{
		VolunteerID: volunteer.ID,
		Observation: "Water is rising near the footbridge and families are waiting.",
		Media:       []string{"media_not_registered"},
	}))
	request.SetPathValue("id", task.ID)
	srv.submitVolunteerObservationHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "unknown_media" {
		t.Fatalf("expected unknown_media error code, got %q", payload.Error.Code)
	}
}

func TestVolunteerRegistrationRequiresCitizenToken(t *testing.T) {
	srv := newTokenOnlyTestServer()

	// No token at all.
	unauthenticated := httptest.NewRecorder()
	srv.registerVolunteerHandler(unauthenticated, httptest.NewRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(validVolunteerRegistrationRequest())))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated registration %d, got %d: %s", http.StatusUnauthorized, unauthenticated.Code, unauthenticated.Body.String())
	}

	// An agency token is not a citizen account.
	agency := httptest.NewRecorder()
	srv.registerVolunteerHandler(agency, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(validVolunteerRegistrationRequest()), testAuthorityClaims()))
	if agency.Code != http.StatusForbidden {
		t.Fatalf("expected agency registration %d, got %d: %s", http.StatusForbidden, agency.Code, agency.Body.String())
	}

	// Body citizenUserId must match the verified token subject.
	mismatched := validVolunteerRegistrationRequest()
	mismatch := httptest.NewRecorder()
	srv.registerVolunteerHandler(mismatch, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(mismatched), citizenClaims("usr_someone_else")))
	if mismatch.Code != http.StatusForbidden {
		t.Fatalf("expected mismatched citizen %d, got %d: %s", http.StatusForbidden, mismatch.Code, mismatch.Body.String())
	}

	// A matching citizen token registers and derives citizenUserId from the token.
	response := httptest.NewRecorder()
	srv.registerVolunteerHandler(response, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(mismatched), citizenClaims(mismatched.CitizenUserID)))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected citizen registration %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.VolunteerProfileResponse
	decodeResponse(t, response, &payload)
	if payload.Volunteer.CitizenUserID != mismatched.CitizenUserID {
		t.Fatalf("expected citizenUserId from token, got %#v", payload.Volunteer)
	}
}

func TestVolunteerRegistrationDerivesCitizenUserIDFromToken(t *testing.T) {
	srv := newTestServer()
	body := validVolunteerRegistrationRequest()
	body.CitizenUserID = ""

	response := httptest.NewRecorder()
	srv.registerVolunteerHandler(response, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(body), citizenClaims("usr_volunteer_001")))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}
	var payload models.VolunteerProfileResponse
	decodeResponse(t, response, &payload)
	if payload.Volunteer.CitizenUserID != "usr_volunteer_001" {
		t.Fatalf("expected derived citizenUserId, got %#v", payload.Volunteer)
	}
}

func TestVolunteerRegistrationIsIdempotentPerCitizen(t *testing.T) {
	srv := newTestServer()
	body := validVolunteerRegistrationRequest()
	first := registerVolunteerForTest(t, srv, body)

	// Re-registering the same citizen returns the existing profile with 200
	// instead of minting a duplicate vol_ id that would orphan assignments.
	changed := validVolunteerRegistrationRequest()
	changed.Phone = "+233200000999"
	response := httptest.NewRecorder()
	srv.registerVolunteerHandler(response, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(changed), citizenClaims(body.CitizenUserID)))
	if response.Code != http.StatusOK {
		t.Fatalf("expected idempotent re-registration %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	var payload models.VolunteerProfileResponse
	decodeResponse(t, response, &payload)
	if payload.Volunteer.ID != first.ID {
		t.Fatalf("expected existing volunteer id %s, got %s", first.ID, payload.Volunteer.ID)
	}
	if payload.Volunteer.Phone != first.Phone {
		t.Fatalf("expected existing profile to be returned unchanged, got %#v", payload.Volunteer)
	}
	if len(srv.store.ListVolunteers("", "")) != 1 {
		t.Fatalf("expected exactly one volunteer profile, got %#v", srv.store.ListVolunteers("", ""))
	}
}

func TestVolunteerRegistrationIsRateLimited(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := NewServer(store.NewMemoryStore(), func() time.Time { return now }, &config.Config{RateLimit: 1, RateWindowSecs: 60, TokenSecret: testTokenSecret})

	first := httptest.NewRecorder()
	srv.registerVolunteerHandler(first, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(validVolunteerRegistrationRequest()), citizenClaims("usr_volunteer_001")))
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first registration %d, got %d: %s", http.StatusCreated, first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	other := validVolunteerRegistrationRequest()
	other.CitizenUserID = "usr_volunteer_002"
	srv.registerVolunteerHandler(second, tokenRequest(http.MethodPost, "/api/v1/volunteers", jsonBody(other), citizenClaims(other.CitizenUserID)))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limited registration %d, got %d: %s", http.StatusTooManyRequests, second.Code, second.Body.String())
	}
}

func TestVolunteerTaskMutationsRejectReadOnlyAgencyRoles(t *testing.T) {
	srv := newTokenOnlyTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	viewerClaims := testAuthorityClaims()
	viewerClaims.Role = "agency_viewer"

	statusResponse := httptest.NewRecorder()
	statusRequest := tokenRequest(http.MethodPatch, "/api/v1/volunteer-tasks/"+task.ID+"/status", jsonBody(models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
	}), viewerClaims)
	statusRequest.SetPathValue("id", task.ID)
	srv.updateVolunteerTaskStatusHandler(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusForbidden {
		t.Fatalf("expected agency_viewer status update %d, got %d: %s", http.StatusForbidden, statusResponse.Code, statusResponse.Body.String())
	}

	observationResponse := httptest.NewRecorder()
	observationRequest := tokenRequest(http.MethodPost, "/api/v1/volunteer-tasks/"+task.ID+"/observations", jsonBody(models.VolunteerObservationRequest{
		VolunteerID: volunteer.ID,
		Observation: "Viewer attempting to mutate task state.",
	}), viewerClaims)
	observationRequest.SetPathValue("id", task.ID)
	srv.submitVolunteerObservationHandler(observationResponse, observationRequest)
	if observationResponse.Code != http.StatusForbidden {
		t.Fatalf("expected agency_viewer observation %d, got %d: %s", http.StatusForbidden, observationResponse.Code, observationResponse.Body.String())
	}

	// Read access is preserved for the viewer role.
	tasksResponse := httptest.NewRecorder()
	tasksRequest := tokenRequest(http.MethodGet, "/api/v1/volunteers/"+volunteer.ID+"/tasks", nil, viewerClaims)
	tasksRequest.SetPathValue("id", volunteer.ID)
	srv.listVolunteerTasksHandler(tasksResponse, tasksRequest)
	if tasksResponse.Code != http.StatusOK {
		t.Fatalf("expected agency_viewer task list %d, got %d: %s", http.StatusOK, tasksResponse.Code, tasksResponse.Body.String())
	}
}

func TestVolunteerTaskEventsAttributedToVerifiedActor(t *testing.T) {
	srv := newTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	// Authority actor via verified headers: attribution must use the verified
	// actor id, not the client-supplied volunteerId.
	updated := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
		Note:        "Dispatcher confirmed task acceptance by phone.",
	})
	latest := updated.Updates[len(updated.Updates)-1]
	if latest.CreatedBy != "usr_dispatcher_001" {
		t.Fatalf("expected update createdBy verified actor, got %q", latest.CreatedBy)
	}

	incidents := srv.store.ListIncidents("")
	var incident models.IncidentRecord
	for _, item := range incidents {
		if item.ID == task.IncidentID {
			incident = item
			break
		}
	}
	var statusEvent *models.TimelineEvent
	for index := range incident.Timeline {
		if incident.Timeline[index].Type == "incident.volunteer_status_updated" {
			statusEvent = &incident.Timeline[index]
		}
	}
	if statusEvent == nil || statusEvent.ActorUserID != "usr_dispatcher_001" || statusEvent.ActorRole != "nadmo_officer" {
		t.Fatalf("expected timeline attribution to verified actor, got %#v", statusEvent)
	}

	auditFound := false
	for _, logEntry := range srv.store.ListAudit(10) {
		if logEntry.Action == "incident.volunteer_status_updated" {
			auditFound = true
			if logEntry.ActorUserID != "usr_dispatcher_001" {
				t.Fatalf("expected audit attribution to verified actor, got %#v", logEntry)
			}
		}
	}
	if !auditFound {
		t.Fatal("expected incident.volunteer_status_updated audit event")
	}
}

func TestVolunteerTaskEventsAttributedToOwningCitizen(t *testing.T) {
	srv := newTokenOnlyTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	response := httptest.NewRecorder()
	request := tokenRequest(http.MethodPatch, "/api/v1/volunteer-tasks/"+task.ID+"/status", jsonBody(models.VolunteerTaskStatusRequest{
		VolunteerID: volunteer.ID,
		Status:      "accepted",
	}), citizenClaims(volunteer.CitizenUserID))
	request.SetPathValue("id", task.ID)
	srv.updateVolunteerTaskStatusHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected citizen status update %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.VolunteerTaskRecord
	decodeResponse(t, response, &payload)
	latest := payload.Updates[len(payload.Updates)-1]
	if latest.CreatedBy != volunteer.CitizenUserID {
		t.Fatalf("expected update createdBy owning citizen, got %q", latest.CreatedBy)
	}
}

func TestVolunteerObservationCannotResurrectTerminalTask(t *testing.T) {
	srv := newTestServer()
	volunteer, task := assignedVolunteerTaskForTest(t, srv)

	for _, next := range []string{"accepted", "on_scene", "completed"} {
		updated := updateVolunteerTaskStatusForTest(t, srv, task.ID, models.VolunteerTaskStatusRequest{
			VolunteerID: volunteer.ID,
			Status:      next,
		})
		if updated.Status != next {
			t.Fatalf("expected task status %s, got %#v", next, updated)
		}
	}

	// An escalating observation on a completed task must be rejected by the
	// same transition guard as status updates, not resurrect the task.
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/volunteer-tasks/"+task.ID+"/observations", jsonBody(models.VolunteerObservationRequest{
		VolunteerID:         volunteer.ID,
		Observation:         "Late escalation attempt after task completion.",
		SafetyStatus:        "unsafe",
		EscalationRequested: true,
	}))
	request.SetPathValue("id", task.ID)
	srv.submitVolunteerObservationHandler(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected terminal task observation %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_transition" {
		t.Fatalf("expected invalid_transition error code, got %q", payload.Error.Code)
	}

	final, found := srv.store.VolunteerTaskByID(task.ID)
	if !found || final.Status != "completed" || final.EscalationRequired {
		t.Fatalf("expected task to remain completed without escalation, got %#v", final)
	}

	// A plain (non-escalating) observation still records without a status change.
	plain := submitVolunteerObservationForTest(t, srv, task.ID, models.VolunteerObservationRequest{
		VolunteerID:  volunteer.ID,
		Observation:  "Final photo of the cleared drainage for the record.",
		SafetyStatus: "safe",
	})
	if plain.Status != "completed" || len(plain.Updates) != 4 {
		t.Fatalf("expected plain observation to append without transition, got %#v", plain)
	}
}
