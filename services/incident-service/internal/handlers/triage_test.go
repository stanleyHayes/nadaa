package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
)

func triageStringPtr(value string) *string { return &value }

func triageIntPtr(value int) *int { return &value }

func suggestTriageForTest(t *testing.T, srv *server, incidentID string) models.TriageSuggestion {
	t.Helper()

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/"+incidentID+"/triage", nil)
	request.SetPathValue("id", incidentID)
	srv.suggestTriageHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.TriageResponse
	decodeResponse(t, response, &payload)
	return payload.Suggestion
}

func TestSuggestTriageReturnsExplainableSuggestion(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	suggestion := suggestTriageForTest(t, srv, incident.ID)

	if suggestion.Severity != "high" {
		t.Fatalf("expected high severity, got %s", suggestion.Severity)
	}
	if suggestion.HumanReviewRequired != true || suggestion.AutoPublishAllowed != false {
		t.Fatalf("expected human review and no auto-publish, got %#v", suggestion)
	}
	if suggestion.Confidence == "" {
		t.Fatalf("expected confidence, got empty")
	}
	if len(suggestion.ExplanationFactors) == 0 {
		t.Fatalf("expected explanation factors, got none")
	}
	if suggestion.SuggestedAgency.AgencyType == "" {
		t.Fatalf("expected suggested agency type, got empty")
	}
	if !strings.HasPrefix(suggestion.SuggestionID, "trs_") {
		t.Fatalf("expected a trs_ suggestion id, got %q", suggestion.SuggestionID)
	}
}

func TestSuggestTriageLogsSuggestionExposure(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	suggestion := suggestTriageForTest(t, srv, incident.ID)

	logs := srv.store.ListAudit(10)
	if !containsAuditAction(logs, "incident.triage_suggested") {
		t.Fatalf("expected triage suggested audit event, got %#v", logs)
	}
	for _, event := range logs {
		if event.Action != "incident.triage_suggested" {
			continue
		}
		snapshot, ok := event.After["triageSuggestion"].(map[string]any)
		if !ok {
			t.Fatalf("expected triage suggestion snapshot in audit event, got %#v", event.After)
		}
		if snapshot["suggestionId"] != suggestion.SuggestionID {
			t.Fatalf("expected audited suggestion id %q, got %#v", suggestion.SuggestionID, snapshot["suggestionId"])
		}
	}
}

func TestSuggestTriageIgnoresFalseReportDuplicates(t *testing.T) {
	srv := newTestServer()
	first := createIncidentForTest(t, srv, validIncidentRequest())
	second := createIncidentForTest(t, srv, validIncidentRequest())

	withDuplicate := suggestTriageForTest(t, srv, first.ID)
	if withDuplicate.DuplicateLikelihood <= 0 {
		t.Fatalf("expected duplicate likelihood above zero, got %v", withDuplicate.DuplicateLikelihood)
	}

	ctx := models.AuthorityContext{ActorUserID: "usr_dispatch", ActorAgencyID: "agc_001", ActorRole: "nadmo_officer", MFACompleted: true}
	if _, code, message := srv.store.TransitionIncident(second.ID, "false_report", ctx, models.IncidentWorkflowRequest{
		Note:            "Confirmed duplicate false report",
		ResolutionNotes: "Report judged false after dispatcher callback.",
	}, srv.now()); code != "" {
		t.Fatalf("expected false_report transition to succeed, got %s: %s", code, message)
	}

	afterDismissal := suggestTriageForTest(t, srv, first.ID)
	if afterDismissal.DuplicateLikelihood != 0 {
		t.Fatalf("expected duplicate likelihood 0 after candidate dismissed, got %v", afterDismissal.DuplicateLikelihood)
	}
	if len(afterDismissal.TopDuplicateIncidentIDs) != 0 {
		t.Fatalf("expected no duplicate ids after candidate dismissed, got %#v", afterDismissal.TopDuplicateIncidentIDs)
	}
	if afterDismissal.AffectedPopulation >= withDuplicate.AffectedPopulation {
		t.Fatalf("expected affected population to drop after candidate dismissed, got %d >= %d", afterDismissal.AffectedPopulation, withDuplicate.AffectedPopulation)
	}
}

func TestSuggestTriageForMissingIncident(t *testing.T) {
	srv := newTestServer()

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/inc_missing/triage", nil)
	request.SetPathValue("id", "inc_missing")
	srv.suggestTriageHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestSuggestTriageRequiresReaderRole(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID+"/triage", nil)
	request.Header.Set("X-NADAA-Actor-Role", "citizen")
	request.SetPathValue("id", incident.ID)
	srv.suggestTriageHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestRecordTriageOverrideLogsAuditAndTimeline(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	suggestion := suggestTriageForTest(t, srv, incident.ID)

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/triage-review",
		jsonBody(models.TriageReviewRequest{
			Accepted:     false,
			SuggestionID: suggestion.SuggestionID,
			OverriddenFields: &models.TriageOverrideFields{
				Severity:            triageStringPtr("emergency"),
				AffectedPopulation:  triageIntPtr(45),
				SuggestedAgencyType: triageStringPtr("fire"),
				SuggestedAgencyID:   triageStringPtr("00000000-0000-0000-0000-000000000201"),
			},
			Reason: "Scene upgraded after dispatcher callback confirmed trapped vehicles.",
		}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewTriageHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.TriageReviewResponse
	decodeResponse(t, response, &payload)

	if !containsTimelineType(payload.Incident.Timeline, "incident.triage_overridden") {
		t.Fatalf("expected triage overridden timeline event, got %#v", payload.Incident.Timeline)
	}

	logs := srv.store.ListAudit(10)
	if !containsAuditAction(logs, "incident.triage_overridden") {
		t.Fatalf("expected triage overridden audit event, got %#v", logs)
	}
	for _, event := range logs {
		if event.Action != "incident.triage_overridden" {
			continue
		}
		snapshot, ok := event.After["triageSuggestion"].(map[string]any)
		if !ok || snapshot["suggestionId"] != suggestion.SuggestionID {
			t.Fatalf("expected audit to reference reviewed suggestion %q, got %#v", suggestion.SuggestionID, event.After)
		}
		if event.After["triageSuggestionSource"] != "logged" {
			t.Fatalf("expected logged suggestion source, got %#v", event.After["triageSuggestionSource"])
		}
	}
}

func TestRecordTriageOverrideAuditsOnlySuppliedFields(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/triage-review",
		jsonBody(models.TriageReviewRequest{
			Accepted: false,
			OverriddenFields: &models.TriageOverrideFields{
				Severity: triageStringPtr("emergency"),
			},
			Reason: "Callback confirmed the situation is worse than reported.",
		}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewTriageHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	logs := srv.store.ListAudit(10)
	for _, event := range logs {
		if event.Action != "incident.triage_overridden" {
			continue
		}
		override, ok := event.After["triageOverride"].(map[string]any)
		if !ok {
			t.Fatalf("expected triage override snapshot, got %#v", event.After)
		}
		fields, ok := override["overriddenFields"].(map[string]any)
		if !ok {
			t.Fatalf("expected overridden fields snapshot, got %#v", override)
		}
		if fields["severity"] != "emergency" {
			t.Fatalf("expected severity override to be audited, got %#v", fields)
		}
		if _, present := fields["affectedPopulation"]; present {
			t.Fatalf("expected unsupplied affectedPopulation to be absent from audit, got %#v", fields)
		}
		if _, present := fields["suggestedAgencyType"]; present {
			t.Fatalf("expected unsupplied suggestedAgencyType to be absent from audit, got %#v", fields)
		}
	}
}

func TestRecordTriageAcceptLogsAcceptedEvent(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	suggestion := suggestTriageForTest(t, srv, incident.ID)

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/triage-review",
		jsonBody(models.TriageReviewRequest{Accepted: true, SuggestionID: suggestion.SuggestionID}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewTriageHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	logs := srv.store.ListAudit(10)
	if !containsAuditAction(logs, "incident.triage_accepted") {
		t.Fatalf("expected triage accepted audit event, got %#v", logs)
	}
}

func TestRecordTriageReviewRejectsUnknownSuggestionID(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/triage-review",
		jsonBody(models.TriageReviewRequest{Accepted: true, SuggestionID: "trs_does_not_exist"}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewTriageHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestRecordTriageOverrideRejectsEmptyOverride(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/triage-review",
		jsonBody(models.TriageReviewRequest{
			Accepted:         false,
			OverriddenFields: &models.TriageOverrideFields{},
			Reason:           "Override with no fields should be rejected.",
		}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewTriageHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestRecordTriageOverrideRequiresReason(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/triage-review",
		jsonBody(models.TriageReviewRequest{
			Accepted: false,
			OverriddenFields: &models.TriageOverrideFields{
				Severity: triageStringPtr("emergency"),
			},
		}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewTriageHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}
