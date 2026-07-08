package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
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

	observed := submitVolunteerObservationForTest(t, srv, task.ID, models.VolunteerObservationRequest{
		VolunteerID:         volunteer.ID,
		Observation:         "Water is rising near the footbridge and families are waiting for transport.",
		SafetyStatus:        "needs_authority",
		Location:            &models.Coordinates{Lat: 5.561, Lng: -0.201},
		EscalationRequested: true,
		Media:               []string{"media_volunteer_photo_001"},
	})
	if observed.Status != "needs_escalation" || !observed.EscalationRequired || len(observed.Updates) != 2 {
		t.Fatalf("expected escalated volunteer observation, got %#v", observed)
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
