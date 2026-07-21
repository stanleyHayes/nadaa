package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
)

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

	var payload models.CreateIncidentResponse
	decodeResponse(t, response, &payload)

	if payload.ID == "" || payload.Reference != "INC-000001" {
		t.Fatalf("expected incident id and reference, got %#v", payload)
	}
	if payload.Status != "reported" || payload.Severity != "high" {
		t.Fatalf("expected reported high incident, got %#v", payload)
	}

	media := srv.store.ListMedia()
	if len(media) != 1 || media[0].Status != "linked" || media[0].IncidentID != payload.ID {
		t.Fatalf("expected media to be linked to incident, got %#v", media)
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
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	incidents := srv.store.ListIncidents("")
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
	body.Location = &models.Coordinates{Lat: 100, Lng: -0.212}

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

	var payload models.CreateIncidentResponse
	decodeResponse(t, response, &payload)
	if !payload.PriorityReview || payload.Severity != "emergency" {
		t.Fatalf("expected emergency priority review, got %#v", payload)
	}
}

func TestCreateDistressRequestForcesRescuePriorityAndAudit(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.RequestKind = "distress_request"
	body.Urgency = "low"
	body.PeopleAffected = 0
	body.Description = "I am trapped by rising water and need rescue now"

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.CreateIncidentResponse
	decodeResponse(t, response, &payload)
	if payload.Reference != "SOS-000001" || payload.RequestKind != "distress_request" || !payload.RescueRequested || !payload.PriorityReview || payload.Severity != "emergency" {
		t.Fatalf("expected emergency rescue request, got %#v", payload)
	}

	incidents := srv.store.ListIncidents("")
	if len(incidents) != 1 || incidents[0].PeopleAffected != 1 || incidents[0].Urgency != "life_threatening" || !containsTimelineType(incidents[0].Timeline, "incident.distress_requested") {
		t.Fatalf("expected normalized distress incident and timeline, got %#v", incidents)
	}
	logs := srv.store.ListAudit(10)
	if len(logs) != 1 || logs[0].Action != "incident.distress_requested" || logs[0].After["rescueRequested"] != true {
		t.Fatalf("expected distress audit event, got %#v", logs)
	}
}

func TestCreateDistressRequestRequiresLocation(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.RequestKind = "distress_request"
	body.Location = nil

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "missing_distress_location") {
		t.Fatalf("expected missing distress location error, got %d: %s", response.Code, response.Body.String())
	}
}

func TestCreateIncidentFlagsSuspiciousReportSignalsWithoutBlocking(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Urgency = "life_threatening"
	body.Description = "Free money promo click here https://example.test emergency "

	payload := createIncidentForTest(t, srv, body)

	if payload.Status != "reported" || !payload.PriorityReview {
		t.Fatalf("expected suspicious life-threatening report to remain live, got %#v", payload)
	}
	if !payload.AbuseReviewRequired || payload.AbuseScore < store.AbuseReviewThreshold {
		t.Fatalf("expected abuse review flag over threshold, got %#v", payload)
	}
	if !containsAbuseSignal(payload.AbuseSignals, "external_link") ||
		!containsAbuseSignal(payload.AbuseSignals, "promotional_language") {
		t.Fatalf("expected link and promotional abuse signals, got %#v", payload.AbuseSignals)
	}

	incidents := srv.store.ListIncidents("")
	if !containsTimelineType(incidents[0].Timeline, "incident.abuse_flagged") {
		t.Fatalf("expected abuse flag timeline event, got %#v", incidents[0].Timeline)
	}
}

func TestCreateIncidentFlagsReporterBurst(t *testing.T) {
	srv := newTestServer()
	first := validIncidentRequest()
	first.Description = "Flood water blocking roadside shops near the bridge."
	createIncidentForTest(t, srv, first)

	second := validIncidentRequest()
	second.Description = "Tree branches fell near the flooded road entrance."
	createIncidentForTest(t, srv, second)

	third := validIncidentRequest()
	third.Description = "Drain cover washed away near the same flooded road."
	payload := createIncidentForTest(t, srv, third)

	if !payload.AbuseReviewRequired || !containsAbuseSignal(payload.AbuseSignals, "reporter_burst") {
		t.Fatalf("expected reporter burst abuse signal, got %#v", payload)
	}
}

func TestAbuseReviewCanClearSuspiciousReport(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Description = "Promo link https://example.test but caller confirmed active flood"
	incident := createIncidentForTest(t, srv, body)

	payload := reviewAbuseForTest(t, srv, incident.ID, models.AbuseReviewRequest{
		Decision: "clear",
		Note:     "Dispatcher confirmed caller and live flood conditions.",
	})

	if payload.AbuseReviewRequired || payload.AbuseReviewDecision != "clear" || payload.AbuseReviewedBy != "usr_dispatcher_001" || payload.AbuseReviewedAt == nil {
		t.Fatalf("expected cleared abuse review metadata, got %#v", payload)
	}
	if !containsTimelineType(payload.Timeline, "incident.abuse_cleared") {
		t.Fatalf("expected abuse cleared timeline event, got %#v", payload.Timeline)
	}

	logs := srv.store.ListAudit(10)
	if len(logs) != 1 || logs[0].Action != "incident.abuse_cleared" {
		t.Fatalf("expected abuse cleared audit event, got %#v", logs)
	}
}

func TestAbuseReviewCanMarkFalseReportWithResolution(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Description = "Free money promo click here https://example.test"
	incident := createIncidentForTest(t, srv, body)

	missingResolution := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/abuse-review",
		jsonBody(models.AbuseReviewRequest{Decision: "false_report", Note: "Dispatcher confirmed no emergency."}),
	)
	request.SetPathValue("id", incident.ID)
	srv.reviewAbuseHandler(missingResolution, request)
	if missingResolution.Code != http.StatusBadRequest {
		t.Fatalf("expected missing resolution status %d, got %d", http.StatusBadRequest, missingResolution.Code)
	}

	payload := reviewAbuseForTest(t, srv, incident.ID, models.AbuseReviewRequest{
		Decision:        "false_report",
		Note:            "Dispatcher confirmed no emergency at the reported location.",
		ResolutionNotes: "Field callback and district desk confirmed no active incident.",
	})

	if payload.Status != "false_report" || payload.AbuseReviewRequired || payload.ResolutionNotes == "" || payload.ClosedAt == nil {
		t.Fatalf("expected false report closure metadata, got %#v", payload)
	}
	if !containsTimelineType(payload.Timeline, "incident.false_reported") {
		t.Fatalf("expected false report timeline event, got %#v", payload.Timeline)
	}

	logs := srv.store.ListAudit(10)
	if len(logs) != 1 || logs[0].Action != "incident.false_reported" || logs[0].After["status"] != "false_report" {
		t.Fatalf("expected false report abuse audit event, got %#v", logs)
	}
}

func TestAbuseReviewRequiresWorkflowRole(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/abuse-review",
		jsonBody(models.AbuseReviewRequest{Decision: "clear", Note: "Viewer should not clear moderation."}),
	)
	request.Header.Set("X-NADAA-Actor-Role", "agency_viewer")
	request.SetPathValue("id", incident.ID)
	srv.reviewAbuseHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
}

func TestDuplicateCandidatesIncludeSameLocationReport(t *testing.T) {
	srv := newTestServer()
	first := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 1 {
		t.Fatalf("expected one duplicate candidate, got %#v", second.DuplicateCandidates)
	}
	candidate := second.DuplicateCandidates[0]
	if candidate.IncidentID != first.ID || candidate.Reference != first.Reference {
		t.Fatalf("expected first incident as duplicate candidate, got %#v", candidate)
	}
	if candidate.Score < store.DuplicateMinimumScore || candidate.DistanceMeters != 0 || candidate.MinutesApart != 0 {
		t.Fatalf("expected strong same-location candidate, got %#v", candidate)
	}
	if !containsString(candidate.Reasons, "same_hazard") || !containsString(candidate.Reasons, "similar_description") {
		t.Fatalf("expected duplicate reasons to include same hazard and similar description, got %#v", candidate.Reasons)
	}

	incidents := srv.store.ListIncidents("")
	var original models.IncidentRecord
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
	body.Location = &models.Coordinates{Lat: 5.581, Lng: -0.211}
	body.Reporter = &models.ReporterRef{UserID: "usr_003", Phone: "+233200000003"}
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 1 {
		t.Fatalf("expected nearby duplicate candidate, got %#v", second.DuplicateCandidates)
	}
	candidate := second.DuplicateCandidates[0]
	if candidate.DistanceMeters <= 0 || candidate.DistanceMeters > store.DuplicateDistanceMeters {
		t.Fatalf("expected candidate within dedupe distance, got %#v", candidate)
	}
	if candidate.Score < store.DuplicateMinimumScore {
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
	srv := NewServer(store.NewMemoryStore(), func() time.Time { return now }, &config.Config{RateLimit: 100, RateWindowSecs: 60})

	createIncidentForTest(t, srv, validIncidentRequest())

	now = now.Add(store.DuplicateReviewWindow + time.Minute)
	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 0 {
		t.Fatalf("expected old report outside duplicate window to be ignored, got %#v", second.DuplicateCandidates)
	}
}

func TestDuplicateReviewReturnsSideBySideCandidates(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	duplicate := createIncidentForTest(t, srv, body)

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/"+primary.ID+"/duplicates", nil)
	request.SetPathValue("id", primary.ID)
	srv.duplicateReviewHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DuplicateReviewResponse
	decodeResponse(t, response, &payload)
	if payload.Incident.ID != primary.ID || len(payload.Candidates) != 1 {
		t.Fatalf("expected primary with one duplicate candidate, got %#v", payload)
	}
	if payload.Candidates[0].Incident.ID != duplicate.ID || payload.Candidates[0].Candidate.Reference != duplicate.Reference {
		t.Fatalf("expected duplicate incident details in review response, got %#v", payload.Candidates[0])
	}
}

func TestDuplicateReviewSanitizesReporterForResponder(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	createIncidentForTest(t, srv, body)

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/"+primary.ID+"/duplicates", nil)
	request.Header.Set("X-NADAA-Actor-Role", "responder")
	request.SetPathValue("id", primary.ID)
	srv.duplicateReviewHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.DuplicateReviewResponse
	decodeResponse(t, response, &payload)
	if payload.Incident.ReportedBy != nil {
		t.Fatalf("expected responder duplicate review to hide primary reporter, got %#v", payload.Incident.ReportedBy)
	}
	if payload.Candidates[0].Incident.ReportedBy != nil {
		t.Fatalf("expected responder duplicate review to hide candidate reporter, got %#v", payload.Candidates[0].Incident.ReportedBy)
	}
	if payload.Incident.Privacy.ReporterIdentityVisible || payload.Candidates[0].Incident.Privacy.ReporterContactVisible {
		t.Fatalf("expected duplicate review privacy flags to hide reporter details, got primary=%#v candidate=%#v", payload.Incident.Privacy, payload.Candidates[0].Incident.Privacy)
	}
}

func TestMergeDuplicateIncidentsClosesDuplicateAndAudits(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	duplicate := createIncidentForTest(t, srv, body)

	payload := mergeIncidentsForTest(t, srv, primary.ID, models.MergeIncidentsRequest{
		DuplicateIncidentIDs: []string{duplicate.ID},
		Note:                 "Same flooded road confirmed from duplicate calls.",
	})

	if !containsString(payload.Incident.MergedIncidentIDs, duplicate.ID) {
		t.Fatalf("expected primary to keep merged duplicate id, got %#v", payload.Incident.MergedIncidentIDs)
	}
	if hasDuplicateCandidate(payload.Incident.DuplicateCandidates, duplicate.ID) {
		t.Fatalf("expected merged duplicate to be removed from primary candidates, got %#v", payload.Incident.DuplicateCandidates)
	}
	if len(payload.MergedIncidents) != 1 {
		t.Fatalf("expected one merged incident, got %#v", payload.MergedIncidents)
	}
	merged := payload.MergedIncidents[0]
	if merged.ID != duplicate.ID || merged.MergedIntoID != primary.ID || merged.Status != "closed" || merged.ResolutionNotes == "" {
		t.Fatalf("expected duplicate to close as traceable merge, got %#v", merged)
	}
	if !containsTimelineType(payload.Incident.Timeline, "incident.merged") ||
		!containsTimelineType(merged.Timeline, "incident.merged_into") {
		t.Fatalf("expected merge timeline events, primary=%#v duplicate=%#v", payload.Incident.Timeline, merged.Timeline)
	}

	logs := srv.store.ListAudit(10)
	if len(logs) != 2 {
		t.Fatalf("expected primary and duplicate merge audit events, got %#v", logs)
	}
	if logs[0].Action != "incident.merged" || logs[0].TargetID != primary.ID {
		t.Fatalf("expected latest audit event to capture primary merge, got %#v", logs[0])
	}
	if logs[0].After["mergedIncidentIds"] == nil || logs[1].Action != "incident.merged_into" {
		t.Fatalf("expected merge trace fields in audit snapshots, got %#v", logs)
	}
}

func hasDuplicateCandidate(candidates []models.DuplicateCandidate, incidentID string) bool {
	for _, candidate := range candidates {
		if candidate.IncidentID == incidentID {
			return true
		}
	}
	return false
}

func TestMergeDuplicateIncidentsRejectsNonCandidate(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Type = "fire"
	body.Description = "Roadside kiosk fire away from the flooded road."
	nonCandidate := createIncidentForTest(t, srv, body)

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+primary.ID+"/merge",
		jsonBody(models.MergeIncidentsRequest{
			DuplicateIncidentIDs: []string{nonCandidate.ID},
			Note:                 "Trying to merge unrelated incidents.",
		}),
	)
	request.SetPathValue("id", primary.ID)
	srv.mergeIncidentHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestMergeDuplicateIncidentsRequiresWorkflowRole(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	duplicate := createIncidentForTest(t, srv, body)

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+primary.ID+"/merge",
		jsonBody(models.MergeIncidentsRequest{
			DuplicateIncidentIDs: []string{duplicate.ID},
			Note:                 "Viewer should not merge duplicate reports.",
		}),
	)
	request.Header.Set("X-NADAA-Actor-Role", "agency_viewer")
	request.SetPathValue("id", primary.ID)
	srv.mergeIncidentHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
}

func TestRateLimit(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := NewServer(store.NewMemoryStore(), func() time.Time { return now }, &config.Config{RateLimit: 1, RateWindowSecs: 60})

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
	srv.listIncidentsHandler(response, authorityRequest(http.MethodGet, "/api/v1/incidents", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload models.IncidentListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Incidents) != 1 {
		t.Fatalf("expected one incident, got %#v", payload)
	}
}

func TestListIncidentsRequiresAuthorityContext(t *testing.T) {
	srv := newTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	srv.listIncidentsHandler(response, httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListIncidentsSanitizesReporterPrivacy(t *testing.T) {
	srv := newTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	dispatcherResponse := httptest.NewRecorder()
	srv.listIncidentsHandler(dispatcherResponse, authorityRequest(http.MethodGet, "/api/v1/incidents", nil))

	var dispatcherPayload models.IncidentListResponse
	decodeResponse(t, dispatcherResponse, &dispatcherPayload)
	if dispatcherPayload.Incidents[0].ReportedBy == nil || dispatcherPayload.Incidents[0].ReportedBy.Phone == "" {
		t.Fatalf("expected permitted dispatcher view to include reporter contact, got %#v", dispatcherPayload.Incidents[0])
	}
	if !dispatcherPayload.Incidents[0].Privacy.ReporterIdentityVisible || !dispatcherPayload.Incidents[0].Privacy.ReporterContactVisible {
		t.Fatalf("expected privacy policy to show visible contact, got %#v", dispatcherPayload.Incidents[0].Privacy)
	}

	responderRequest := authorityRequest(http.MethodGet, "/api/v1/incidents", nil)
	responderRequest.Header.Set("X-NADAA-Actor-Role", "responder")
	responderResponse := httptest.NewRecorder()
	srv.listIncidentsHandler(responderResponse, responderRequest)

	var responderPayload models.IncidentListResponse
	decodeResponse(t, responderResponse, &responderPayload)
	if responderPayload.Incidents[0].ReportedBy != nil {
		t.Fatalf("expected responder standard view to hide reporter identity, got %#v", responderPayload.Incidents[0].ReportedBy)
	}
	if responderPayload.Incidents[0].Privacy.ReporterIdentityVisible || responderPayload.Incidents[0].Privacy.ReporterContactVisible {
		t.Fatalf("expected privacy policy to hide reporter contact, got %#v", responderPayload.Incidents[0].Privacy)
	}
}

func TestAnonymousIncidentStaysAnonymousInAuthorityList(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Anonymous = true
	body.ContactPermission = false
	createIncidentForTest(t, srv, body)

	response := httptest.NewRecorder()
	srv.listIncidentsHandler(response, authorityRequest(http.MethodGet, "/api/v1/incidents", nil))

	var payload models.IncidentListResponse
	decodeResponse(t, response, &payload)
	if payload.Incidents[0].ReportedBy != nil {
		t.Fatalf("expected anonymous authority view to hide reporter, got %#v", payload.Incidents[0].ReportedBy)
	}
	if payload.Incidents[0].Privacy.ReporterIdentityVisible || payload.Incidents[0].Privacy.ReporterContactVisible {
		t.Fatalf("expected anonymous privacy policy to hide identity and contact, got %#v", payload.Incidents[0].Privacy)
	}
}

func TestVerifyIncidentAuditsStatusChange(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/verify",
		jsonBody(models.IncidentWorkflowRequest{Note: "Confirmed with caller and duplicate report."}),
	)
	request.SetPathValue("id", incident.ID)
	srv.verifyIncidentHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	if payload.Status != "verified" || payload.VerifiedBy != "usr_dispatcher_001" || payload.VerifiedAt == nil {
		t.Fatalf("expected verified incident with verifier metadata, got %#v", payload)
	}

	logs := srv.store.ListAudit(10)
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

	logs := srv.store.ListAudit(10)
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

	incidents := srv.store.ListIncidents("")
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

	var payload models.IncidentListResponse
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
		body := models.IncidentStatusRequest{
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

		var payload models.IncidentRecord
		decodeResponse(t, response, &payload)
		if payload.Status != nextStatus {
			t.Fatalf("expected incident status %s, got %#v", nextStatus, payload)
		}
	}

	logs := srv.store.ListAudit(20)
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
		jsonBody(models.IncidentStatusRequest{
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

	incidents := srv.store.ListIncidents("")
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
		jsonBody(models.IncidentStatusRequest{Status: "false_report", Note: "Caller recanted."}),
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
		jsonBody(models.IncidentStatusRequest{
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

	var payload models.IncidentRecord
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
		jsonBody(models.IncidentWorkflowRequest{Note: "Missing MFA"}),
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
		jsonBody(models.IncidentWorkflowRequest{Note: "Confirmed."}),
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

	var payload models.IncidentAuditListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Logs) != 1 || payload.Logs[0].Action != "incident.verified" {
		t.Fatalf("expected one incident audit log, got %#v", payload)
	}
}

func TestCreateIncidentInvalidDescriptionReturnsDescriptionCode(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Description = "bad"

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}

	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_description" {
		t.Fatalf("expected invalid_description error code, got %q", payload.Error.Code)
	}
}

func TestCreateIncidentRejectsUnsafeReporterIdentifiers(t *testing.T) {
	srv := newTestServer()

	for name, reporter := range map[string]models.ReporterRef{
		"unsafe user id": {UserID: "bad user!"},
		"unsafe phone":   {UserID: "usr_001", Phone: "+233<script>"},
	} {
		body := validIncidentRequest()
		body.Reporter = &reporter

		response := httptest.NewRecorder()
		srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected status %d, got %d", name, http.StatusBadRequest, response.Code)
		}
		var payload models.APIError
		decodeResponse(t, response, &payload)
		if payload.Error.Code != "invalid_reporter" {
			t.Fatalf("%s: expected invalid_reporter error code, got %q", name, payload.Error.Code)
		}
	}
}

func TestStatusWorkflowRejectsVerifiedForNonVerifierRoles(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPatch,
		"/api/v1/incidents/"+incident.ID+"/status",
		jsonBody(models.IncidentStatusRequest{Status: "verified", Note: "Responder attempting verification."}),
	)
	request.Header.Set("X-NADAA-Actor-Role", "responder")
	request.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}

	allowed := httptest.NewRecorder()
	request = authorityRequest(
		http.MethodPatch,
		"/api/v1/incidents/"+incident.ID+"/status",
		jsonBody(models.IncidentStatusRequest{Status: "under_review", Note: "Responder operational update is still allowed."}),
	)
	request.Header.Set("X-NADAA-Actor-Role", "responder")
	request.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(allowed, request)

	if allowed.Code != http.StatusOK {
		t.Fatalf("expected responder under_review status %d, got %d: %s", http.StatusOK, allowed.Code, allowed.Body.String())
	}

	incidents := srv.store.ListIncidents("")
	if incidents[0].Status != "under_review" || incidents[0].VerifiedBy != "" {
		t.Fatalf("expected incident to stay unverified, got %#v", incidents[0])
	}
}

func TestMergeRejectsVerifiedDuplicate(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	duplicate := createIncidentForTest(t, srv, body)
	verifyIncidentForTest(t, srv, duplicate.ID)

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+primary.ID+"/merge",
		jsonBody(models.MergeIncidentsRequest{
			DuplicateIncidentIDs: []string{duplicate.ID},
			Note:                 "Verified duplicate must not be force-closed.",
		}),
	)
	request.SetPathValue("id", primary.ID)
	srv.mergeIncidentHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
}

func TestMergeRejectsAssignedDuplicate(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	duplicate := createIncidentForTest(t, srv, body)
	verifyIncidentForTest(t, srv, duplicate.ID)
	assignIncidentForTest(t, srv, duplicate.ID, validAssignmentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(
		http.MethodPost,
		"/api/v1/incidents/"+primary.ID+"/merge",
		jsonBody(models.MergeIncidentsRequest{
			DuplicateIncidentIDs: []string{duplicate.ID},
			Note:                 "Assigned duplicate must not be force-closed.",
		}),
	)
	request.SetPathValue("id", primary.ID)
	srv.mergeIncidentHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}

	incidents := srv.store.ListIncidents("")
	for _, item := range incidents {
		if item.ID == duplicate.ID && (item.Status == "closed" || item.MergedIntoID != "") {
			t.Fatalf("expected assigned duplicate to remain untouched, got %#v", item)
		}
	}
}

func TestMergeClearsAbuseFlagAndAssignmentsOnMergedRecord(t *testing.T) {
	srv := newTestServer()
	primary := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Description = "Free money promo click here https://example.test flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	duplicate := createIncidentForTest(t, srv, body)
	if !duplicate.AbuseReviewRequired {
		t.Fatalf("expected duplicate to require abuse review, got %#v", duplicate)
	}

	payload := mergeIncidentsForTest(t, srv, primary.ID, models.MergeIncidentsRequest{
		DuplicateIncidentIDs: []string{duplicate.ID},
		Note:                 "Same flooded road confirmed from duplicate calls.",
	})

	if len(payload.MergedIncidents) != 1 {
		t.Fatalf("expected one merged incident, got %#v", payload.MergedIncidents)
	}
	merged := payload.MergedIncidents[0]
	if merged.AbuseReviewRequired || len(merged.Assignments) != 0 {
		t.Fatalf("expected merged record to clear abuse flag and assignments, got %#v", merged)
	}
}

func TestDuplicateCandidatesSkipClosedIncidents(t *testing.T) {
	srv := newTestServer()
	first := createIncidentForTest(t, srv, validIncidentRequest())

	ctx := models.AuthorityContext{ActorUserID: "usr_dispatch", ActorAgencyID: "agc_001", ActorRole: "nadmo_officer", MFACompleted: true}
	for _, next := range []string{"under_review", "verified", "assigned"} {
		if _, code, message := srv.store.TransitionIncident(first.ID, next, ctx, models.IncidentWorkflowRequest{Note: "Operational step."}, srv.now()); code != "" {
			t.Fatalf("expected transition to %s to succeed, got %s: %s", next, code, message)
		}
	}
	if _, code, message := srv.store.TransitionIncident(first.ID, "closed", ctx, models.IncidentWorkflowRequest{
		Note:            "Response complete.",
		ResolutionNotes: "Waters receded and the road reopened.",
	}, srv.now()); code != "" {
		t.Fatalf("expected close to succeed, got %s: %s", code, message)
	}

	body := validIncidentRequest()
	body.Description = "Vehicles are trapped on the same flooded road"
	body.Reporter = &models.ReporterRef{UserID: "usr_002", Phone: "+233200000002"}
	second := createIncidentForTest(t, srv, body)

	if len(second.DuplicateCandidates) != 0 {
		t.Fatalf("expected closed incident to be skipped in duplicate scoring, got %#v", second.DuplicateCandidates)
	}
}

func TestGetIncidentByID(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil)
	request.SetPathValue("id", incident.ID)
	srv.getIncidentHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	if payload.ID != incident.ID || payload.Reference != incident.Reference {
		t.Fatalf("expected incident %s, got %#v", incident.ID, payload)
	}
	if payload.ReportedBy == nil || payload.ReportedBy.Phone == "" {
		t.Fatalf("expected dispatcher view to include reporter contact, got %#v", payload.ReportedBy)
	}

	responderResponse := httptest.NewRecorder()
	responderRequest := authorityRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil)
	responderRequest.Header.Set("X-NADAA-Actor-Role", "responder")
	responderRequest.SetPathValue("id", incident.ID)
	srv.getIncidentHandler(responderResponse, responderRequest)

	if responderResponse.Code != http.StatusOK {
		t.Fatalf("expected responder status %d, got %d: %s", http.StatusOK, responderResponse.Code, responderResponse.Body.String())
	}
	var responderPayload models.IncidentRecord
	decodeResponse(t, responderResponse, &responderPayload)
	if responderPayload.ReportedBy != nil || responderPayload.Privacy.ReporterContactVisible {
		t.Fatalf("expected responder view to hide reporter identity, got %#v", responderPayload.ReportedBy)
	}
}

func TestGetIncidentByIDNotFound(t *testing.T) {
	srv := newTestServer()

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/inc_missing", nil)
	request.SetPathValue("id", "inc_missing")
	srv.getIncidentHandler(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, response.Code)
	}
}

func TestGetIncidentByIDRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil)
	request.SetPathValue("id", incident.ID)
	srv.getIncidentHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestServiceTokenGrantsReadOnlyAccess(t *testing.T) {
	srv := newServiceTokenTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil)
	request.Header.Set(serviceTokenHeader, testInternalServiceToken)
	request.SetPathValue("id", incident.ID)
	srv.getIncidentHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected service token status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	if payload.ID != incident.ID || payload.Reference != incident.Reference {
		t.Fatalf("expected incident %s, got %#v", incident.ID, payload)
	}
	if payload.ReportedBy != nil || payload.Privacy.ReporterIdentityVisible || payload.Privacy.ReporterContactVisible {
		t.Fatalf("expected service token view to hide reporter identity and contact, got %#v", payload.ReportedBy)
	}

	listResponse := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
	listRequest.Header.Set(serviceTokenHeader, testInternalServiceToken)
	srv.listIncidentsHandler(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected service token list status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}
}

func TestGetIncidentByIDRejectsWrongServiceToken(t *testing.T) {
	srv := newServiceTokenTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil)
	request.Header.Set(serviceTokenHeader, "wrong-service-token")
	request.SetPathValue("id", incident.ID)
	srv.getIncidentHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestGetIncidentByIDIgnoresServiceTokenWhenUnset(t *testing.T) {
	srv := newTokenOnlyTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil)
	request.Header.Set(serviceTokenHeader, testInternalServiceToken)
	request.SetPathValue("id", incident.ID)
	srv.getIncidentHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityEndpointsAcceptValidBearerToken(t *testing.T) {
	srv := newTokenOnlyTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	listResponse := httptest.NewRecorder()
	srv.listIncidentsHandler(listResponse, tokenRequest(http.MethodGet, "/api/v1/incidents", nil, testAuthorityClaims()))
	if listResponse.Code != http.StatusOK {
		t.Fatalf("expected token list status %d, got %d: %s", http.StatusOK, listResponse.Code, listResponse.Body.String())
	}

	verifyResponse := httptest.NewRecorder()
	request := tokenRequest(
		http.MethodPost,
		"/api/v1/incidents/"+incident.ID+"/verify",
		jsonBody(models.IncidentWorkflowRequest{Note: "Verified with a signed token."}),
		testAuthorityClaims(),
	)
	request.SetPathValue("id", incident.ID)
	srv.verifyIncidentHandler(verifyResponse, request)
	if verifyResponse.Code != http.StatusOK {
		t.Fatalf("expected token verify status %d, got %d: %s", http.StatusOK, verifyResponse.Code, verifyResponse.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, verifyResponse, &payload)
	if payload.VerifiedBy != testAuthorityClaims().UserID {
		t.Fatalf("expected verifiedBy from token subject, got %#v", payload.VerifiedBy)
	}
}

func TestAuthorityEndpointsIgnoreForgedHeadersWhenMockActorsDisabled(t *testing.T) {
	srv := newTokenOnlyTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	response := httptest.NewRecorder()
	srv.listIncidentsHandler(response, authorityRequest(http.MethodGet, "/api/v1/incidents", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
	}
}

func TestAuthorityEndpointsRejectInvalidAndExpiredTokens(t *testing.T) {
	srv := newTokenOnlyTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	for name, token := range map[string]string{
		"wrong signature": signTestToken("another-secret", testAuthorityClaims()),
		"expired": signTestToken(testTokenSecret, tokenClaims{
			UserID:    "usr_dispatcher_001",
			UserType:  "agency",
			Role:      "nadmo_officer",
			AgencyID:  "00000000-0000-0000-0000-000000000101",
			MFA:       true,
			ExpiresAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		}),
		"malformed": "nadaa.not-a-token",
	} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/incidents", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		srv.listIncidentsHandler(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("%s: expected status %d, got %d: %s", name, http.StatusUnauthorized, response.Code, response.Body.String())
		}
	}
}

func TestTokenWithoutMFAGetsForbidden(t *testing.T) {
	srv := newTokenOnlyTestServer()
	createIncidentForTest(t, srv, validIncidentRequest())

	claims := testAuthorityClaims()
	claims.MFA = false

	response := httptest.NewRecorder()
	srv.listIncidentsHandler(response, tokenRequest(http.MethodGet, "/api/v1/incidents", nil, claims))

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}
}

func TestRoutesResolveIncidentByIDAndAudit(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())
	handler := srv.Routes(nil)

	getResponse := httptest.NewRecorder()
	getRequest := tokenRequest(http.MethodGet, "/api/v1/incidents/"+incident.ID, nil, testAuthorityClaims())
	handler.ServeHTTP(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected GET by id status %d, got %d: %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	var incidentPayload models.IncidentRecord
	decodeResponse(t, getResponse, &incidentPayload)
	if incidentPayload.ID != incident.ID {
		t.Fatalf("expected incident %s from mux route, got %#v", incident.ID, incidentPayload)
	}

	auditResponse := httptest.NewRecorder()
	auditRequest := tokenRequest(http.MethodGet, "/api/v1/incidents/audit", nil, testAuthorityClaims())
	handler.ServeHTTP(auditResponse, auditRequest)
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("expected audit route status %d, got %d: %s", http.StatusOK, auditResponse.Code, auditResponse.Body.String())
	}
}

func TestCreateIncidentRejectsZeroCoordinates(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Location = &models.Coordinates{Lat: 0, Lng: 0}

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_location" {
		t.Fatalf("expected invalid_location error code, got %q", payload.Error.Code)
	}
}

func TestCreateIncidentWithoutLocationRoundTripsNull(t *testing.T) {
	srv := newTestServer()

	// A located same-hazard report exists; locationless reports must not be
	// distance-scored against it.
	located := createIncidentForTest(t, srv, validIncidentRequest())

	body := validIncidentRequest()
	body.Location = nil
	body.Description = "USSD report: drain overflow flooding the Circle underpass."
	body.Reporter = &models.ReporterRef{UserID: "usr_ussd_001", Phone: "+233200000009"}
	locationless := createIncidentForTest(t, srv, body)

	if len(locationless.DuplicateCandidates) != 0 {
		t.Fatalf("expected locationless report to skip duplicate scoring, got %#v", locationless.DuplicateCandidates)
	}

	stored, found := srv.store.GetIncident(locationless.ID)
	if !found {
		t.Fatalf("expected stored incident %s", locationless.ID)
	}
	if stored.Location != nil {
		t.Fatalf("expected stored location to be nil, got %#v", stored.Location)
	}

	// Locationless pairs must not be scored against each other either.
	second := validIncidentRequest()
	second.Location = nil
	second.Description = "Another USSD report about the flooded underpass."
	second.Reporter = &models.ReporterRef{UserID: "usr_ussd_002", Phone: "+233200000010"}
	secondLocationless := createIncidentForTest(t, srv, second)
	if len(secondLocationless.DuplicateCandidates) != 0 {
		t.Fatalf("expected locationless pair to skip duplicate scoring, got %#v", secondLocationless.DuplicateCandidates)
	}

	// The located incident must not gain reverse candidates from locationless reports.
	locatedStored, found := srv.store.GetIncident(located.ID)
	if !found || len(locatedStored.DuplicateCandidates) != 0 {
		t.Fatalf("expected located incident to keep zero candidates, got %#v", locatedStored.DuplicateCandidates)
	}

	// The authority read round-trips the missing location as JSON null.
	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/incidents/"+locationless.ID, nil)
	request.SetPathValue("id", locationless.ID)
	srv.getIncidentHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"location":null`) {
		t.Fatalf("expected location to round-trip as null, got %s", response.Body.String())
	}
}

func TestUpdateIncidentStatusRejectsFalseReportForNonAbuseRoles(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	falseReportBody := models.IncidentStatusRequest{
		Status:          "false_report",
		Note:            "Responder attempting to hide a report.",
		ResolutionNotes: "Responder claims nothing is happening at the scene.",
	}

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/incidents/"+incident.ID+"/status", jsonBody(falseReportBody))
	request.Header.Set("X-NADAA-Actor-Role", "responder")
	request.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected responder false_report status %d, got %d: %s", http.StatusForbidden, response.Code, response.Body.String())
	}

	incidents := srv.store.ListIncidents("")
	if incidents[0].Status != "reported" {
		t.Fatalf("expected incident to remain reported, got %#v", incidents[0])
	}

	// Abuse-review roles keep the ability to mark false reports via status.
	allowed := httptest.NewRecorder()
	allowedRequest := authorityRequest(http.MethodPatch, "/api/v1/incidents/"+incident.ID+"/status", jsonBody(falseReportBody))
	allowedRequest.Header.Set("X-NADAA-Actor-Role", "dispatcher")
	allowedRequest.SetPathValue("id", incident.ID)
	srv.updateIncidentStatusHandler(allowed, allowedRequest)
	if allowed.Code != http.StatusOK {
		t.Fatalf("expected dispatcher false_report status %d, got %d: %s", http.StatusOK, allowed.Code, allowed.Body.String())
	}
}

func TestUpdateIncidentStatusAllowsResponderToClose(t *testing.T) {
	srv := newTestServer()
	incident := createIncidentForTest(t, srv, validIncidentRequest())

	// closed is an operational terminal status, not owned by the abuse flow;
	// responders keep the ability to close after response.
	for _, next := range []string{"under_review", "verified", "assigned", "response_en_route", "on_scene", "contained", "recovery_ongoing"} {
		updated := updateIncidentStatusForTest(t, srv, incident.ID, "", models.IncidentStatusRequest{Status: next, Note: "Operational step."})
		if updated.Status != next {
			t.Fatalf("expected status %s, got %#v", next, updated)
		}
	}

	closed := updateIncidentStatusForTest(t, srv, incident.ID, "responder", models.IncidentStatusRequest{
		Status:          "closed",
		Note:            "Response complete.",
		ResolutionNotes: "Waters receded and the road reopened.",
	})
	if closed.Status != "closed" || closed.ClosedAt == nil {
		t.Fatalf("expected responder to close incident, got %#v", closed)
	}
}

func updateIncidentStatusForTest(t *testing.T, srv *server, incidentID string, role string, body models.IncidentStatusRequest) models.IncidentRecord {
	t.Helper()

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPatch, "/api/v1/incidents/"+incidentID+"/status", jsonBody(body))
	if role != "" {
		request.Header.Set("X-NADAA-Actor-Role", role)
	}
	request.SetPathValue("id", incidentID)
	srv.updateIncidentStatusHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected status update %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.IncidentRecord
	decodeResponse(t, response, &payload)
	return payload
}

func TestCreateIncidentRejectsOversizedBody(t *testing.T) {
	srv := newTestServer()
	body := validIncidentRequest()
	body.Description = strings.Repeat("flood", 300000) // ~1.5 MB, over the 1 MiB cap

	response := httptest.NewRecorder()
	srv.createIncidentHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/incidents", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	var payload models.APIError
	decodeResponse(t, response, &payload)
	if payload.Error.Code != "invalid_json" {
		t.Fatalf("expected invalid_json error code, got %q", payload.Error.Code)
	}
}
