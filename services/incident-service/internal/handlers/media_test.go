package handlers

import (
	"bytes"
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

func newMediaContentTestServer(t *testing.T) *server {
	t.Helper()

	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		RateLimit:        100,
		RateWindowSecs:   60,
		TokenSecret:      testTokenSecret,
		MediaStoragePath: t.TempDir(),
	}
	return NewServer(store.NewMemoryStore(), func() time.Time { return now }, cfg)
}

func initiateMediaUploadWithClaims(t *testing.T, srv *server, body models.InitiateMediaUploadRequest, claims tokenClaims) models.MediaUploadResponse {
	t.Helper()

	response := httptest.NewRecorder()
	srv.initiateMediaUploadHandler(response, tokenRequest(http.MethodPost, "/api/v1/media/uploads", jsonBody(body), claims))
	if response.Code != http.StatusCreated {
		t.Fatalf("expected media upload status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.MediaUploadResponse
	decodeResponse(t, response, &payload)
	return payload
}

func TestInitiateMediaUploadReturnsAbsoluteContentURL(t *testing.T) {
	srv := newTestServer()
	payload := initiateMediaUploadWithClaims(t, srv, validMediaUploadRequest(), citizenClaims("usr_001"))

	expected := "http://localhost:8084/api/v1/media/" + payload.MediaID + "/content"
	if payload.UploadURL != expected {
		t.Fatalf("expected absolute content upload URL %q, got %q", expected, payload.UploadURL)
	}
}

func TestMediaContentRoundTrip(t *testing.T) {
	srv := newMediaContentTestServer(t)
	payload := initiateMediaUploadWithClaims(t, srv, validMediaUploadRequest(), citizenClaims("usr_001"))

	content := []byte("fake-jpeg-bytes-for-round-trip")
	putResponse := httptest.NewRecorder()
	putRequest := tokenRequest(http.MethodPut, payload.UploadURL, bytes.NewReader(content), citizenClaims("usr_001"))
	putRequest.SetPathValue("id", payload.MediaID)
	srv.putMediaContentHandler(putResponse, putRequest)
	if putResponse.Code != http.StatusOK {
		t.Fatalf("expected PUT status %d, got %d: %s", http.StatusOK, putResponse.Code, putResponse.Body.String())
	}

	var record models.MediaRecord
	decodeResponse(t, putResponse, &record)
	if record.Status != "uploaded" || record.SizeBytes != int64(len(content)) {
		t.Fatalf("expected uploaded record with actual size, got %#v", record)
	}

	// The uploader can read the bytes back.
	getResponse := httptest.NewRecorder()
	getRequest := tokenRequest(http.MethodGet, payload.UploadURL, nil, citizenClaims("usr_001"))
	getRequest.SetPathValue("id", payload.MediaID)
	srv.getMediaContentHandler(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected GET status %d, got %d: %s", http.StatusOK, getResponse.Code, getResponse.Body.String())
	}
	if !bytes.Equal(getResponse.Body.Bytes(), content) {
		t.Fatalf("expected stored bytes to round-trip, got %q", getResponse.Body.String())
	}
	if getResponse.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatalf("expected stored content type, got %q", getResponse.Header().Get("Content-Type"))
	}

	// Authority users can read the bytes for incident review.
	authorityResponse := httptest.NewRecorder()
	authorityRequest := tokenRequest(http.MethodGet, payload.UploadURL, nil, testAuthorityClaims())
	authorityRequest.SetPathValue("id", payload.MediaID)
	srv.getMediaContentHandler(authorityResponse, authorityRequest)
	if authorityResponse.Code != http.StatusOK {
		t.Fatalf("expected authority GET status %d, got %d: %s", http.StatusOK, authorityResponse.Code, authorityResponse.Body.String())
	}
	if !bytes.Equal(authorityResponse.Body.Bytes(), content) {
		t.Fatalf("expected authority read to return stored bytes, got %q", authorityResponse.Body.String())
	}
}

func TestMediaContentRejectsOversizedUpload(t *testing.T) {
	srv := newMediaContentTestServer(t)
	payload := initiateMediaUploadWithClaims(t, srv, validMediaUploadRequest(), citizenClaims("usr_001"))

	oversized := make([]byte, utils.AllowedMediaTypes["image/jpeg"]+1)
	putResponse := httptest.NewRecorder()
	putRequest := tokenRequest(http.MethodPut, payload.UploadURL, bytes.NewReader(oversized), citizenClaims("usr_001"))
	putRequest.SetPathValue("id", payload.MediaID)
	srv.putMediaContentHandler(putResponse, putRequest)
	if putResponse.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d: %s", http.StatusRequestEntityTooLarge, putResponse.Code, putResponse.Body.String())
	}

	record, found := srv.store.MediaByID(payload.MediaID)
	if !found || record.Status != "pending_upload" {
		t.Fatalf("expected media record to stay pending after oversize rejection, got %#v", record)
	}
}

func TestMediaContentRequiresAuthentication(t *testing.T) {
	srv := newMediaContentTestServer(t)
	payload := initiateMediaUploadWithClaims(t, srv, validMediaUploadRequest(), citizenClaims("usr_001"))

	putResponse := httptest.NewRecorder()
	putRequest := httptest.NewRequest(http.MethodPut, payload.UploadURL, bytes.NewReader([]byte("data")))
	putRequest.SetPathValue("id", payload.MediaID)
	srv.putMediaContentHandler(putResponse, putRequest)
	if putResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated PUT %d, got %d: %s", http.StatusUnauthorized, putResponse.Code, putResponse.Body.String())
	}

	getResponse := httptest.NewRecorder()
	getRequest := httptest.NewRequest(http.MethodGet, payload.UploadURL, nil)
	getRequest.SetPathValue("id", payload.MediaID)
	srv.getMediaContentHandler(getResponse, getRequest)
	if getResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated GET %d, got %d: %s", http.StatusUnauthorized, getResponse.Code, getResponse.Body.String())
	}

	// A different citizen cannot write or read another uploader's media.
	otherPut := httptest.NewRecorder()
	otherPutRequest := tokenRequest(http.MethodPut, payload.UploadURL, bytes.NewReader([]byte("data")), citizenClaims("usr_intruder"))
	otherPutRequest.SetPathValue("id", payload.MediaID)
	srv.putMediaContentHandler(otherPut, otherPutRequest)
	if otherPut.Code != http.StatusForbidden {
		t.Fatalf("expected other citizen PUT %d, got %d: %s", http.StatusForbidden, otherPut.Code, otherPut.Body.String())
	}

	otherGet := httptest.NewRecorder()
	otherGetRequest := tokenRequest(http.MethodGet, payload.UploadURL, nil, citizenClaims("usr_intruder"))
	otherGetRequest.SetPathValue("id", payload.MediaID)
	srv.getMediaContentHandler(otherGet, otherGetRequest)
	if otherGet.Code != http.StatusForbidden {
		t.Fatalf("expected other citizen GET %d, got %d: %s", http.StatusForbidden, otherGet.Code, otherGet.Body.String())
	}
}

func TestMediaContentGetMissingFileReturnsNotFound(t *testing.T) {
	srv := newMediaContentTestServer(t)
	payload := initiateMediaUploadWithClaims(t, srv, validMediaUploadRequest(), citizenClaims("usr_001"))

	getResponse := httptest.NewRecorder()
	getRequest := tokenRequest(http.MethodGet, payload.UploadURL, nil, citizenClaims("usr_001"))
	getRequest.SetPathValue("id", payload.MediaID)
	srv.getMediaContentHandler(getResponse, getRequest)
	if getResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status %d before bytes are uploaded, got %d: %s", http.StatusNotFound, getResponse.Code, getResponse.Body.String())
	}
}
