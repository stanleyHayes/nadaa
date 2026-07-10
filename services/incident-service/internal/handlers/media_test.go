package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
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
