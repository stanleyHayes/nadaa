package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateSubmitAndApproveAlert(t *testing.T) {
	srv := &server{store: newMemoryStore()}

	createResponse := httptest.NewRecorder()
	createRequest := authorizedRequest(http.MethodPost, "/api/v1/alerts", validAlertBody())
	srv.createAlertHandler(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var draft authorityAlert
	decodeResponse(t, createResponse, &draft)
	if draft.Status != "draft" || draft.IssuingAgencyID == "" {
		t.Fatalf("unexpected draft alert: %#v", draft)
	}

	submitResponse := httptest.NewRecorder()
	submitRequest := authorizedRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/submit", "")
	srv.submitAlertHandler(submitResponse, submitRequest.WithContext(submitRequest.Context()))

	if submitResponse.Code != http.StatusOK {
		t.Fatalf("expected submit status %d, got %d: %s", http.StatusOK, submitResponse.Code, submitResponse.Body.String())
	}

	var submitted authorityAlert
	decodeResponse(t, submitResponse, &submitted)
	if submitted.Status != "submitted" || submitted.SubmittedAt == nil {
		t.Fatalf("unexpected submitted alert: %#v", submitted)
	}

	approveResponse := httptest.NewRecorder()
	approveRequest := approverRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/approve", `{"note":"Reviewed by NADMO approver"}`)
	srv.approveAlertHandler(approveResponse, approveRequest)

	if approveResponse.Code != http.StatusOK {
		t.Fatalf("expected approve status %d, got %d: %s", http.StatusOK, approveResponse.Code, approveResponse.Body.String())
	}

	var approved authorityAlert
	decodeResponse(t, approveResponse, &approved)
	if approved.Status != "approved" || approved.ApprovedBy != "usr_approver" || approved.ApprovedAt == nil {
		t.Fatalf("unexpected approved alert: %#v", approved)
	}

	auditLogs := srv.store.listAudit(10)
	if len(auditLogs) != 3 {
		t.Fatalf("expected create, submit, and approve audit logs, got %#v", auditLogs)
	}
}

func TestCreateAlertStoresSourcePredictionInAudit(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	draft := createAlertWithBody(t, srv, alertBodyWithSourcePrediction())

	if draft.SourcePrediction == nil {
		t.Fatalf("expected source prediction on draft alert: %#v", draft)
	}
	if draft.SourcePrediction.PredictionID != "pred_grid-accra-central-001" {
		t.Fatalf("unexpected source prediction: %#v", draft.SourcePrediction)
	}
	if draft.SourcePrediction.AutoPublishAllowed {
		t.Fatalf("source prediction must not allow auto-publish: %#v", draft.SourcePrediction)
	}

	logs := srv.store.listAudit(10)
	if len(logs) != 1 || logs[0].Action != "alert.created" {
		t.Fatalf("expected alert.created audit log, got %#v", logs)
	}
	afterPrediction, ok := logs[0].After["sourcePrediction"].(*alertSourcePrediction)
	if !ok || afterPrediction.PredictionID != draft.SourcePrediction.PredictionID {
		t.Fatalf("expected source prediction in audit snapshot, got %#v", logs[0].After)
	}
}

func TestCreateAlertRejectsUnsafeSourcePrediction(t *testing.T) {
	srv := &server{store: newMemoryStore()}

	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts", strings.Replace(alertBodyWithSourcePrediction(), `"autoPublishAllowed": false`, `"autoPublishAllowed": true`, 1))
	srv.createAlertHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected unsafe source prediction status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestRejectRequiresReasonAndAudits(t *testing.T) {
	srv := &server{store: newMemoryStore()}

	response := httptest.NewRecorder()
	request := approverRequest(http.MethodPost, "/api/v1/alerts/alert_fixture_submitted/reject", `{"reason":"Needs clearer target wording"}`)
	srv.rejectAlertHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected reject status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var rejected authorityAlert
	decodeResponse(t, response, &rejected)
	if rejected.Status != "rejected" || rejected.StatusReason == "" || rejected.RejectedBy != "usr_approver" {
		t.Fatalf("unexpected rejected alert: %#v", rejected)
	}

	logs := srv.store.listAudit(10)
	if len(logs) != 1 || logs[0].Action != "alert.rejected" {
		t.Fatalf("expected rejection audit log, got %#v", logs)
	}
}

func TestRejectWithoutReasonFails(t *testing.T) {
	srv := &server{store: newMemoryStore()}

	response := httptest.NewRecorder()
	request := approverRequest(http.MethodPost, "/api/v1/alerts/alert_fixture_submitted/reject", `{"reason":""}`)
	srv.rejectAlertHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestApprovalRequiresDifferentActor(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	draft := createAndSubmitAlert(t, srv)

	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/approve", `{"note":"Trying to approve own draft"}`)
	request.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
	srv.approveAlertHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestEmergencyOverrideRequiresAuthorizedRoleAndAudit(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	draft := createAlert(t, srv)

	viewerResponse := httptest.NewRecorder()
	viewerRequest := authorizedRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/emergency-override", `{"reason":"Immediate life-safety warning"}`)
	viewerRequest.Header.Set("X-NADAA-Actor-Role", "agency_viewer")
	srv.emergencyOverrideHandler(viewerResponse, viewerRequest)

	if viewerResponse.Code != http.StatusForbidden {
		t.Fatalf("expected viewer override status %d, got %d", http.StatusForbidden, viewerResponse.Code)
	}

	overrideResponse := httptest.NewRecorder()
	overrideRequest := approverRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/emergency-override", `{"reason":"Immediate life-safety warning"}`)
	srv.emergencyOverrideHandler(overrideResponse, overrideRequest)

	if overrideResponse.Code != http.StatusOK {
		t.Fatalf("expected override status %d, got %d: %s", http.StatusOK, overrideResponse.Code, overrideResponse.Body.String())
	}

	var overridden authorityAlert
	decodeResponse(t, overrideResponse, &overridden)
	if overridden.Status != "approved" || !overridden.EmergencyOverride {
		t.Fatalf("unexpected override alert: %#v", overridden)
	}

	logs := srv.store.listAudit(10)
	if logs[0].Action != "alert.emergency_override" {
		t.Fatalf("expected emergency override audit log first, got %#v", logs)
	}
}

func TestWriteRequiresMFA(t *testing.T) {
	srv := &server{store: newMemoryStore()}

	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts", validAlertBody())
	request.Header.Set("X-NADAA-MFA-Completed", "false")
	srv.createAlertHandler(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestPublicListOnlyReturnsApprovedCurrentAlerts(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	draft := createAndSubmitAlert(t, srv)

	overrideResponse := httptest.NewRecorder()
	overrideRequest := approverRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/emergency-override", `{"reason":"Immediate life-safety warning"}`)
	srv.emergencyOverrideHandler(overrideResponse, overrideRequest)
	if overrideResponse.Code != http.StatusOK {
		t.Fatalf("override failed: %s", overrideResponse.Body.String())
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?current=true", nil)
	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload alertListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Alerts) != 1 || payload.Alerts[0].Status != "approved" {
		t.Fatalf("expected one current approved alert, got %#v", payload.Alerts)
	}
}

func TestPublicListHidesSourcePredictionMetadata(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	draft := createAlertWithBody(t, srv, alertBodyWithSourcePrediction())

	overrideResponse := httptest.NewRecorder()
	overrideRequest := approverRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/emergency-override", `{"reason":"Immediate life-safety warning"}`)
	srv.emergencyOverrideHandler(overrideResponse, overrideRequest)
	if overrideResponse.Code != http.StatusOK {
		t.Fatalf("override failed: %s", overrideResponse.Body.String())
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?current=true", nil)
	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var payload alertListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Alerts) != 1 || payload.Alerts[0].SourcePrediction != nil {
		t.Fatalf("expected public alert without source prediction metadata, got %#v", payload.Alerts)
	}
}

func TestPreviewDistrictTargetReturnsGeometry(t *testing.T) {
	srv := &server{store: newMemoryStore()}

	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts/targets/preview", `{
		"type": "district",
		"ids": ["accra-metropolitan"],
		"label": "Accra Metropolitan"
	}`)
	srv.previewTargetHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected preview status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload targetPreviewResponse
	decodeResponse(t, response, &payload)
	if payload.Target.Geometry == nil || payload.Target.Center == nil || payload.Target.AreaSqKm <= 0 {
		t.Fatalf("expected district target geometry and area, got %#v", payload.Target)
	}
	if payload.Target.EstimatedPopulation == 0 || payload.Summary == "" {
		t.Fatalf("expected population and summary, got %#v", payload)
	}
}

func TestRadiusTargetIsStoredAndQueryable(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	target := `{
		"type": "radius",
		"ids": ["accra-central-radius"],
		"label": "Accra Central 5km",
		"center": { "lat": 5.556, "lng": -0.202 },
		"radiusMeters": 5000
	}`
	draft := createAlertWithBody(t, srv, alertBodyWithTarget(target))

	if draft.Target.Center == nil || draft.Target.RadiusMeters != 5000 || draft.Target.AreaSqKm <= 0 {
		t.Fatalf("expected radius target metadata, got %#v", draft.Target)
	}

	overrideResponse := httptest.NewRecorder()
	overrideRequest := approverRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/emergency-override", `{"reason":"Immediate life-safety warning"}`)
	srv.emergencyOverrideHandler(overrideResponse, overrideRequest)
	if overrideResponse.Code != http.StatusOK {
		t.Fatalf("override failed: %s", overrideResponse.Body.String())
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/alerts?current=true&targetType=radius&targetId=accra-central-radius", nil)
	srv.listAlertsHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, response.Code)
	}

	var payload alertListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Alerts) != 1 || payload.Alerts[0].ID != draft.ID {
		t.Fatalf("expected radius alert from target query, got %#v", payload.Alerts)
	}
}

func TestCustomTargetRejectsInvalidPolygon(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	target := `{
		"type": "custom",
		"ids": ["bad-polygon"],
		"label": "Bad Polygon",
		"geometry": {
			"type": "Polygon",
			"coordinates": [[[-0.22,5.55],[-0.18,5.55],[-0.18,5.59]]]
		}
	}`

	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts", alertBodyWithTarget(target))
	srv.createAlertHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid custom polygon status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestCustomTargetStoresPolygonGeometry(t *testing.T) {
	srv := &server{store: newMemoryStore()}
	target := `{
		"type": "custom",
		"ids": ["kaneshie-flood-box"],
		"label": "Kaneshie Flood Box",
		"geometry": {
			"type": "Polygon",
			"coordinates": [[[-0.24,5.55],[-0.19,5.55],[-0.19,5.59],[-0.24,5.59],[-0.24,5.55]]]
		}
	}`

	draft := createAlertWithBody(t, srv, alertBodyWithTarget(target))
	if draft.Target.Geometry == nil || draft.Target.AreaSqKm <= 0 || draft.Target.Center == nil {
		t.Fatalf("expected custom target geometry metadata, got %#v", draft.Target)
	}
}

func createAndSubmitAlert(t *testing.T, srv *server) authorityAlert {
	t.Helper()
	draft := createAlert(t, srv)

	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/submit", "")
	srv.submitAlertHandler(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("submit alert: %s", response.Body.String())
	}

	var submitted authorityAlert
	decodeResponse(t, response, &submitted)
	return submitted
}

func createAlert(t *testing.T, srv *server) authorityAlert {
	t.Helper()
	return createAlertWithBody(t, srv, validAlertBody())
}

func createAlertWithBody(t *testing.T, srv *server, body string) authorityAlert {
	t.Helper()
	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts", body)
	srv.createAlertHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create alert: %s", response.Body.String())
	}

	var alert authorityAlert
	decodeResponse(t, response, &alert)
	return alert
}

func alertBodyWithTarget(target string) string {
	startsAt := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	expiresAt := time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339)
	return `{
		"title": "Severe flood warning",
		"hazardType": "flood",
		"severity": "severe_warning",
		"message": "Avoid low-lying roads and move to higher ground immediately.",
		"target": ` + target + `,
		"startsAt": "` + startsAt + `",
		"expiresAt": "` + expiresAt + `",
		"recommendedAction": "Prepare to evacuate if instructed by NADMO.",
		"evacuationRequired": false,
		"shelterIds": ["00000000-0000-0000-0000-000000000301"]
	}`
}

func validAlertBody() string {
	startsAt := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	expiresAt := time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339)
	return `{
		"title": "Severe flood warning",
		"hazardType": "flood",
		"severity": "severe_warning",
		"message": "Avoid low-lying roads and move to higher ground immediately.",
		"target": {
			"type": "district",
			"ids": ["accra-metropolitan"],
			"label": "Accra Metropolitan"
		},
		"startsAt": "` + startsAt + `",
		"expiresAt": "` + expiresAt + `",
		"recommendedAction": "Prepare to evacuate if instructed by NADMO.",
		"evacuationRequired": false,
		"shelterIds": ["00000000-0000-0000-0000-000000000301"]
	}`
}

func alertBodyWithSourcePrediction() string {
	startsAt := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	expiresAt := time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339)
	return `{
		"title": "ML reviewed flood warning",
		"hazardType": "flood",
		"severity": "severe_warning",
		"message": "Reviewed ML flood prediction indicates severe flood risk near Accra Central.",
		"target": {
			"type": "district",
			"ids": ["accra-metropolitan"],
			"label": "Accra Metropolitan"
		},
		"startsAt": "` + startsAt + `",
		"expiresAt": "` + expiresAt + `",
		"recommendedAction": "Prepare to evacuate if instructed by NADMO.",
		"evacuationRequired": false,
		"shelterIds": ["00000000-0000-0000-0000-000000000301"],
		"sourcePrediction": {
			"predictionId": "pred_grid-accra-central-001",
			"predictionLogId": "ml_log_fixture",
			"modelVersion": "flood-logistic-baseline-0.1.0",
			"inputFeatureSetVersion": "flood-risk-features.v1",
			"probability": 0.9993,
			"severity": "severe",
			"confidence": "medium",
			"humanReviewRequired": true,
			"autoPublishAllowed": false,
			"reviewNote": "Dispatcher reviewed explanation factors."
		}
	}`
}

func authorizedRequest(method string, path string, body string) *http.Request {
	var reader *bytes.Buffer
	if body == "" {
		reader = bytes.NewBuffer(nil)
	} else {
		reader = bytes.NewBufferString(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.SetPathValue("id", pathID(path))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-NADAA-Actor-ID", "usr_drafter")
	request.Header.Set("X-NADAA-Actor-Role", "district_officer")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000101")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "req_test")
	return request
}

func approverRequest(method string, path string, body string) *http.Request {
	request := authorizedRequest(method, path, body)
	request.Header.Set("X-NADAA-Actor-ID", "usr_approver")
	request.Header.Set("X-NADAA-Actor-Role", "nadmo_officer")
	request.Header.Set("X-NADAA-Request-ID", "req_approve")
	return request
}

func pathID(path string) string {
	parts := strings.Split(path, "/")
	for index, part := range parts {
		if part == "alerts" && index+1 < len(parts) {
			return parts[index+1]
		}
	}
	return ""
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
