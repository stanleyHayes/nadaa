package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/store"
)

func createApprovedCellBroadcast(t *testing.T, srv *Server, languages string) models.CellBroadcastMessage {
	t.Helper()

	createResponse := httptest.NewRecorder()
	createBody := `{"alertId":"alert_feed_current_flood","languages":` + languages + `,"workflowRequestedBy":"dispatcher_001"}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts", bytes.NewBufferString(createBody))
	srv.createCellBroadcastHandler(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.CellBroadcastResponse
	decodeResponse(t, createResponse, &created)

	reviewResponse := httptest.NewRecorder()
	reviewBody := `{"action":"approve","reviewer":"nadmo_cbs_reviewer","note":"Checked telecom template"}`
	reviewRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts/"+created.Message.ID+"/review", bytes.NewBufferString(reviewBody))
	reviewRequest.SetPathValue("id", created.Message.ID)
	withAuthority(reviewRequest)
	srv.reviewCellBroadcastHandler(reviewResponse, reviewRequest)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected review status %d, got %d: %s", http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
	}
	var reviewed models.CellBroadcastResponse
	decodeResponse(t, reviewResponse, &reviewed)
	return reviewed.Message
}

func TestCellBroadcastGenerationReviewAndDelivery(t *testing.T) {
	srv := newTestServer()

	createResponse := httptest.NewRecorder()
	createBody := `{"alertId":"alert_feed_current_flood","languages":["en","tw","ha"],"areas":["Accra Metro","Tema"],"workflowRequestedBy":"dispatcher_001"}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts", bytes.NewBufferString(createBody))
	srv.createCellBroadcastHandler(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var created models.CellBroadcastResponse
	decodeResponse(t, createResponse, &created)
	message := created.Message
	if message.ID == "" || message.ReviewStatus != "pending_review" || len(message.Segments) != 3 {
		t.Fatalf("expected generated pending cell broadcast with 3 segments, got %#v", message)
	}
	// severe_warning maps to the extreme-threat channel, not presidential.
	if message.Channel.MessageIdentifier != 4371 || message.Channel.Label != "extreme" {
		t.Fatalf("expected extreme channel 4371, got %#v", message.Channel)
	}
	if message.EmergencyOverride {
		t.Fatalf("expected no presidential override for severe alert, got %#v", message.Channel)
	}
	if message.Protocol.Standard != "3GPP-CBS" || message.Protocol.Category != "Met" || message.Protocol.CAPSeverity != "Severe" {
		t.Fatalf("expected CAP-classified protocol, got %#v", message.Protocol)
	}
	for _, segment := range message.Segments {
		if segment.DataCodingScheme != "GSM-7" || segment.Pages < 1 || segment.CharacterCount == 0 {
			t.Fatalf("expected encoded segment, got %#v", segment)
		}
		if !strings.Contains(segment.MessageText, "112") || !strings.Contains(segment.MessageText, "Accra") {
			t.Fatalf("expected compliant segment text, got %q", segment.MessageText)
		}
	}

	reviewResponse := httptest.NewRecorder()
	reviewBody := `{"action":"approve","reviewer":"nadmo_cbs_reviewer"}`
	reviewRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts/"+message.ID+"/review", bytes.NewBufferString(reviewBody))
	reviewRequest.SetPathValue("id", message.ID)
	withAuthority(reviewRequest)
	srv.reviewCellBroadcastHandler(reviewResponse, reviewRequest)
	if reviewResponse.Code != http.StatusOK {
		t.Fatalf("expected review status %d, got %d: %s", http.StatusOK, reviewResponse.Code, reviewResponse.Body.String())
	}
	var reviewed models.CellBroadcastResponse
	decodeResponse(t, reviewResponse, &reviewed)
	if reviewed.Message.Status != "approved" || reviewed.Message.ReviewStatus != "approved" {
		t.Fatalf("expected approved cell broadcast, got %#v", reviewed.Message)
	}

	deliverResponse := httptest.NewRecorder()
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts/"+message.ID+"/deliver", bytes.NewBufferString(`{}`))
	deliverRequest.SetPathValue("id", message.ID)
	withAuthority(deliverRequest)
	srv.deliverCellBroadcastHandler(deliverResponse, deliverRequest)
	if deliverResponse.Code != http.StatusAccepted {
		t.Fatalf("expected deliver status %d, got %d: %s", http.StatusAccepted, deliverResponse.Code, deliverResponse.Body.String())
	}

	var delivered models.CellBroadcastDeliveryResponse
	decodeResponse(t, deliverResponse, &delivered)
	if len(delivered.Dispatches) != 3 {
		t.Fatalf("expected three dispatches, got %#v", delivered.Dispatches)
	}
	serials := map[int]bool{}
	for _, dispatch := range delivered.Dispatches {
		if dispatch.Status != "broadcast" || dispatch.Adapter != "sandbox_cbc" {
			t.Fatalf("expected live sandbox broadcast, got %#v", dispatch)
		}
		if dispatch.MessageIdentifier != 4371 || len(dispatch.Areas) != 2 {
			t.Fatalf("expected channel + area metadata, got %#v", dispatch)
		}
		if serials[dispatch.SerialNumber] {
			t.Fatalf("expected unique serial numbers, got duplicate %d", dispatch.SerialNumber)
		}
		serials[dispatch.SerialNumber] = true
	}

	logsResponse := httptest.NewRecorder()
	logsRequest := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/delivery-logs?channel=cell_broadcast&alertId=alert_feed_current_flood", nil)
	srv.listDeliveryLogsHandler(logsResponse, logsRequest)
	var logs models.DeliveryLogListResponse
	decodeResponse(t, logsResponse, &logs)
	if len(logs.Logs) != 3 {
		t.Fatalf("expected persisted cell broadcast audit logs, got %#v", logs.Logs)
	}
}

func TestCellBroadcastReviewerComesFromVerifiedActor(t *testing.T) {
	srv := newTestServer()
	// The review request body self-declares a reviewer; the verified authority
	// actor must win.
	message := createApprovedCellBroadcast(t, srv, `["en"]`)
	if message.Reviewer != "usr_test_authority" {
		t.Fatalf("expected reviewer from the verified actor, got %q", message.Reviewer)
	}
}

func TestCellBroadcastDeliveryRequiresApproval(t *testing.T) {
	srv := newTestServer()

	createResponse := httptest.NewRecorder()
	createBody := `{"alertId":"alert_feed_current_flood","languages":["en"]}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts", bytes.NewBufferString(createBody))
	srv.createCellBroadcastHandler(createResponse, createRequest)
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}
	var created models.CellBroadcastResponse
	decodeResponse(t, createResponse, &created)

	deliverResponse := httptest.NewRecorder()
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts/"+created.Message.ID+"/deliver", bytes.NewBufferString(`{}`))
	deliverRequest.SetPathValue("id", created.Message.ID)
	withAuthority(deliverRequest)
	srv.deliverCellBroadcastHandler(deliverResponse, deliverRequest)
	if deliverResponse.Code != http.StatusConflict {
		t.Fatalf("expected status %d for unapproved delivery, got %d: %s", http.StatusConflict, deliverResponse.Code, deliverResponse.Body.String())
	}
}

func TestCellBroadcastDryRunIsSimulated(t *testing.T) {
	srv := newTestServer()
	message := createApprovedCellBroadcast(t, srv, `["en","tw"]`)

	deliverResponse := httptest.NewRecorder()
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts/"+message.ID+"/deliver", bytes.NewBufferString(`{"dryRun":true}`))
	deliverRequest.SetPathValue("id", message.ID)
	withAuthority(deliverRequest)
	srv.deliverCellBroadcastHandler(deliverResponse, deliverRequest)
	if deliverResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, deliverResponse.Code, deliverResponse.Body.String())
	}

	var delivered models.CellBroadcastDeliveryResponse
	decodeResponse(t, deliverResponse, &delivered)
	for _, dispatch := range delivered.Dispatches {
		if dispatch.Status != "simulated" || !dispatch.DryRun {
			t.Fatalf("expected simulated dry-run dispatch, got %#v", dispatch)
		}
	}
}

func TestCellBroadcastPreview(t *testing.T) {
	srv := newTestServer()
	message := createApprovedCellBroadcast(t, srv, `["en","tw","ha"]`)

	previewResponse := httptest.NewRecorder()
	previewRequest := httptest.NewRequest(http.MethodGet, "/api/v1/notifications/cell-broadcasts/"+message.ID+"/preview", nil)
	previewRequest.SetPathValue("id", message.ID)
	srv.previewCellBroadcastHandler(previewResponse, previewRequest)
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, previewResponse.Code, previewResponse.Body.String())
	}

	var preview models.CellBroadcastPreviewResponse
	decodeResponse(t, previewResponse, &preview)
	if len(preview.Previews) != 3 {
		t.Fatalf("expected three segment previews, got %#v", preview.Previews)
	}
	for _, segment := range preview.Previews {
		if segment.HandsetCategory == "" || segment.Channel == "" || segment.Pages < 1 {
			t.Fatalf("expected handset-accurate preview, got %#v", segment)
		}
	}
}

func TestCellBroadcastDisabledAdapterSkips(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := NewServer(
		store.NewMemoryStore(now),
		nil,
		nil,
		map[string]models.NotificationProvider{},
		models.DisabledCellBroadcastAdapter{Reason: "telecom cell broadcast path is not configured"},
		func() time.Time { return now },
		&config.Config{Addr: ":8090", AllowMockActors: true},
	)
	message := createApprovedCellBroadcast(t, srv, `["en"]`)

	deliverResponse := httptest.NewRecorder()
	deliverRequest := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/cell-broadcasts/"+message.ID+"/deliver", bytes.NewBufferString(`{}`))
	deliverRequest.SetPathValue("id", message.ID)
	withAuthority(deliverRequest)
	srv.deliverCellBroadcastHandler(deliverResponse, deliverRequest)
	if deliverResponse.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d: %s", http.StatusAccepted, deliverResponse.Code, deliverResponse.Body.String())
	}

	var delivered models.CellBroadcastDeliveryResponse
	decodeResponse(t, deliverResponse, &delivered)
	if len(delivered.Dispatches) != 1 {
		t.Fatalf("expected one dispatch, got %#v", delivered.Dispatches)
	}
	dispatch := delivered.Dispatches[0]
	if dispatch.Status != "skipped" || dispatch.Adapter != "disabled_cbc" {
		t.Fatalf("expected disabled adapter to skip broadcast, got %#v", dispatch)
	}
}
