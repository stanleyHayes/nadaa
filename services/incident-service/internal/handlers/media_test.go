package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

func TestInitiateMediaUpload(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(validMediaUploadRequest()))

	srv.initiateMediaUploadHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.MediaUploadResponse
	decodeResponse(t, response, &payload)
	if payload.MediaID == "" || payload.Method != http.MethodPut || payload.Access != "private" {
		t.Fatalf("expected private media upload response, got %#v", payload)
	}
	if payload.MaxSizeBytes != utils.AllowedMediaTypes["image/jpeg"] {
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
	body.SizeBytes = utils.AllowedMediaTypes["image/jpeg"] + 1

	response := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(body)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestListMediaRequiresAuthority(t *testing.T) {
	srv := newTestServer()
	initiateMediaUpload(t, srv)

	unauthenticated := httptest.NewRecorder()
	srv.listMediaHandler(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/media", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, unauthenticated.Code)
	}

	response := httptest.NewRecorder()
	srv.listMediaHandler(response, authorityRequest(http.MethodGet, "/api/v1/media", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.MediaListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Media) != 1 {
		t.Fatalf("expected one media record, got %#v", payload.Media)
	}
}

func TestInitiateMediaUploadIsRateLimited(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	srv := NewServer(store.NewMemoryStore(), func() time.Time { return now }, &config.Config{RateLimit: 1, RateWindowSecs: 60})

	first := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(first, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(validMediaUploadRequest())))
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first upload status %d, got %d: %s", http.StatusCreated, first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(second, httptest.NewRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(validMediaUploadRequest())))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, second.Code)
	}
}
