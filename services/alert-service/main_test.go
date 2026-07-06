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
	response := httptest.NewRecorder()
	request := authorizedRequest(http.MethodPost, "/api/v1/alerts", validAlertBody())
	srv.createAlertHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create alert: %s", response.Body.String())
	}

	var alert authorityAlert
	decodeResponse(t, response, &alert)
	return alert
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
