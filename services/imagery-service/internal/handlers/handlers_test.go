package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	storageDir := filepath.Join(t.TempDir(), "uploads")
	_ = os.MkdirAll(storageDir, 0o750)
	cfg := &config.Config{
		Addr:            ":8099",
		StoragePath:     storageDir,
		RetentionDays:   90,
		TokenSecret:     testTokenSecret,
		AllowMockActors: true,
	}
	return NewServer(store.NewMemoryStore(now, 90), func() time.Time { return now }, cfg)
}

const testTokenSecret = "test-imagery-token-secret"

var testNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

func TestHealth(t *testing.T) {
	srv := newTestServer(t)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.healthHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestCreateImageryViaMultipart(t *testing.T) {
	srv := newTestServer(t)

	fields := map[string]string{
		"source":            "drone",
		"captureTime":       "2026-07-05T10:30:00Z",
		"geometry":          `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.19,5.59],[-0.22,5.59],[-0.22,5.56]]]}`,
		"coverageAreaKm2":   "4.5",
		"resolutionMeters":  "0.25",
		"license":           "CC-BY-4.0",
		"relatedIncidentId": "incident_123",
	}
	response := httptest.NewRecorder()
	request := uploadRequest(fields, "test.png", []byte("\x89PNG\r\n\x1a\n"), "image/png")

	srv.createImageryHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var record models.ImageryRecord
	decodeResponse(t, response, &record)
	if record.Source != "drone" {
		t.Fatalf("expected source drone, got %s", record.Source)
	}
	if record.Status != "active" {
		t.Fatalf("expected status active, got %s", record.Status)
	}
	if record.UploadedBy != "usr_imagery_operator" {
		t.Fatalf("expected uploadedBy usr_imagery_operator, got %s", record.UploadedBy)
	}
	if record.FileName != "test.png" {
		t.Fatalf("expected fileName test.png, got %s", record.FileName)
	}
	if record.SizeBytes != 8 {
		t.Fatalf("expected size 8, got %d", record.SizeBytes)
	}
	if !strings.HasPrefix(record.ContentType, "image/") {
		t.Fatalf("expected image content type, got %s", record.ContentType)
	}
	expectedExpires := record.CreatedAt.Add(90 * 24 * time.Hour)
	if record.ExpiresAt.Sub(expectedExpires).Abs() > time.Second {
		t.Fatalf("expected expiresAt ~%v, got %v", expectedExpires, record.ExpiresAt)
	}
	if record.RelatedIncidentID != "incident_123" {
		t.Fatalf("expected relatedIncidentId incident_123, got %s", record.RelatedIncidentID)
	}
	if _, err := os.Stat(record.StoragePath); err != nil {
		t.Fatalf("expected stored file to exist at %s: %v", record.StoragePath, err)
	}
}

func TestListImagery(t *testing.T) {
	srv := newTestServer(t)

	uploadRecord(t, srv, "satellite", "satellite.tif", []byte("\x89PNG\r\n"), "image/tiff")
	uploadRecord(t, srv, "drone", "drone.jpg", []byte("\x89PNG\r\n"), "image/jpeg")

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/imagery?source=drone", nil)

	srv.listImageryHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ImageryListResponse
	decodeResponse(t, response, &payload)
	if len(payload.Imagery) < 1 {
		t.Fatalf("expected drone imagery, got %#v", payload)
	}
	for _, record := range payload.Imagery {
		if record.Source != "drone" {
			t.Fatalf("expected only drone records, got %#v", record)
		}
	}
}

func TestGeoJSON(t *testing.T) {
	srv := newTestServer(t)

	uploadRecord(t, srv, "satellite", "satellite.tif", []byte("\x89PNG\r\n"), "image/tiff")

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/imagery/geojson", nil)
	request.Host = "example.com"

	srv.geoJSONHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GeoJSONFeatureCollection
	decodeResponse(t, response, &payload)
	if payload.Type != "FeatureCollection" {
		t.Fatalf("expected FeatureCollection, got %s", payload.Type)
	}
	if len(payload.Features) == 0 {
		t.Fatalf("expected at least one feature, got %#v", payload)
	}
	feature := payload.Features[0]
	if feature.Type != "Feature" {
		t.Fatalf("expected Feature, got %s", feature.Type)
	}
	downloadURL, ok := feature.Properties["downloadUrl"].(string)
	if !ok || !strings.Contains(downloadURL, "/api/v1/imagery/"+feature.Properties["id"].(string)+"/download") {
		t.Fatalf("expected valid downloadUrl, got %#v", feature.Properties["downloadUrl"])
	}
}

func TestLifecycleExpiry(t *testing.T) {
	srv := newTestServer(t)

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodPost, "/api/v1/imagery/lifecycle/run", nil)

	srv.runLifecycleHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.ImageryLifecycleResponse
	decodeResponse(t, response, &payload)
	if payload.ExpiredCount < 1 {
		t.Fatalf("expected at least one expired record, got %d", payload.ExpiredCount)
	}

	listResponse := httptest.NewRecorder()
	listRequest := authorityRequest(http.MethodGet, "/api/v1/imagery?status=expired", nil)
	srv.listImageryHandler(listResponse, listRequest)

	var listPayload models.ImageryListResponse
	decodeResponse(t, listResponse, &listPayload)
	if len(listPayload.Imagery) < 1 {
		t.Fatalf("expected expired imagery after lifecycle, got %#v", listPayload)
	}

	geoResponse := httptest.NewRecorder()
	geoRequest := httptest.NewRequest(http.MethodGet, "/imagery/geojson", nil)
	geoRequest.Host = "example.com"
	srv.geoJSONHandler(geoResponse, geoRequest)

	var geoPayload models.GeoJSONFeatureCollection
	decodeResponse(t, geoResponse, &geoPayload)
	if len(geoPayload.Features) == 0 {
		t.Fatalf("expected active features remaining, got %#v", geoPayload)
	}
	for _, feature := range geoPayload.Features {
		if status, ok := feature.Properties["status"]; ok && status == "expired" {
			t.Fatalf("geojson should not include expired records")
		}
	}
}

func TestDeleteImagery(t *testing.T) {
	srv := newTestServer(t)

	record := uploadRecord(t, srv, "drone", "delete-me.png", []byte("\x89PNG\r\n"), "image/png")

	deleteResponse := httptest.NewRecorder()
	deleteRequest := authorityRequest(http.MethodDelete, "/api/v1/imagery/"+record.ID, nil)
	deleteRequest.SetPathValue("id", record.ID)
	srv.deleteImageryHandler(deleteResponse, deleteRequest)

	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, deleteResponse.Code, deleteResponse.Body.String())
	}

	getResponse := httptest.NewRecorder()
	getRequest := authorityRequest(http.MethodGet, "/api/v1/imagery/"+record.ID, nil)
	getRequest.SetPathValue("id", record.ID)
	srv.getImageryHandler(getResponse, getRequest)

	if getResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status %d after delete, got %d", http.StatusNotFound, getResponse.Code)
	}

	if _, err := os.Stat(record.StoragePath); !os.IsNotExist(err) {
		t.Fatalf("expected stored file to be removed, got %v", err)
	}
}

func TestCreateImageryRejectsTooLarge(t *testing.T) {
	srv := newTestServer(t)

	fields := map[string]string{
		"source":           "drone",
		"captureTime":      "2026-07-05T10:30:00Z",
		"geometry":         `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.19,5.59],[-0.22,5.59],[-0.22,5.56]]]}`,
		"coverageAreaKm2":  "4.5",
		"resolutionMeters": "0.25",
	}
	large := make([]byte, maxUploadSize+1)
	response := httptest.NewRecorder()
	request := uploadRequest(fields, "huge.png", large, "image/png")

	srv.createImageryHandler(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d: %s", http.StatusRequestEntityTooLarge, response.Code, response.Body.String())
	}
}

func uploadRecord(t *testing.T, srv *Server, source string, filename string, data []byte, contentType string) models.ImageryRecord {
	t.Helper()
	fields := map[string]string{
		"source":           source,
		"captureTime":      "2026-07-05T10:30:00Z",
		"geometry":         `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.19,5.59],[-0.22,5.59],[-0.22,5.56]]]}`,
		"coverageAreaKm2":  "1.0",
		"resolutionMeters": "1.0",
	}
	response := httptest.NewRecorder()
	request := uploadRequest(fields, filename, data, contentType)
	srv.createImageryHandler(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("upload failed: status %d, %s", response.Code, response.Body.String())
	}
	var record models.ImageryRecord
	decodeResponse(t, response, &record)
	return record
}

func uploadRequest(fields map[string]string, filename string, fileContent []byte, contentType string) *http.Request {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		_ = writer.WriteField(key, value)
	}

	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		panic(err)
	}
	if _, err := part.Write(fileContent); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	request := authorityRequest(http.MethodPost, "/api/v1/imagery", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func authorityRequest(method string, target string, body *bytes.Buffer) *http.Request {
	var request *http.Request
	if body != nil {
		request = httptest.NewRequest(method, target, body)
	} else {
		request = httptest.NewRequest(method, target, nil)
	}
	request.Header.Set("X-NADAA-Actor-ID", "usr_imagery_operator")
	request.Header.Set("X-NADAA-Actor-Role", "analyst")
	request.Header.Set("X-NADAA-Agency-ID", "00000000-0000-0000-0000-000000000204")
	request.Header.Set("X-NADAA-MFA-Completed", "true")
	request.Header.Set("X-NADAA-Request-ID", "test-imagery")
	return request
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func signTestToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encoded))
	return "nadaa." + encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func authorityClaims() map[string]any {
	return map[string]any{
		"sub":      "usr_imagery_operator",
		"typ":      "agency",
		"role":     "analyst",
		"agencyId": "00000000-0000-0000-0000-000000000204",
		"mfa":      true,
		"exp":      testNow.Add(time.Hour).Unix(),
	}
}

func bearerRequest(t *testing.T, method, target string, claims map[string]any) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, target, nil)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, testTokenSecret, claims))
	return request
}

func TestAuthorityAcceptsValidBearerToken(t *testing.T) {
	srv := newTestServer(t)

	response := httptest.NewRecorder()
	request := bearerRequest(t, http.MethodGet, "/api/v1/imagery", authorityClaims())

	srv.listImageryHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestAuthorityRejectsMissingToken(t *testing.T) {
	srv := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/imagery", nil)

	srv.listImageryHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityRejectsForgedToken(t *testing.T) {
	srv := newTestServer(t)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/imagery", nil)
	request.Header.Set("Authorization", "Bearer "+signTestToken(t, "wrong-secret", authorityClaims()))

	srv.listImageryHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityRejectsExpiredToken(t *testing.T) {
	srv := newTestServer(t)

	claims := authorityClaims()
	claims["exp"] = testNow.Add(-time.Hour).Unix()
	response := httptest.NewRecorder()
	request := bearerRequest(t, http.MethodGet, "/api/v1/imagery", claims)

	srv.listImageryHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthorityIgnoresMockHeadersWhenDisabled(t *testing.T) {
	storageDir := filepath.Join(t.TempDir(), "uploads")
	_ = os.MkdirAll(storageDir, 0o750)
	cfg := &config.Config{Addr: ":8099", StoragePath: storageDir, RetentionDays: 90, TokenSecret: testTokenSecret}
	srv := NewServer(store.NewMemoryStore(testNow, 90), func() time.Time { return testNow }, cfg)

	response := httptest.NewRecorder()
	request := authorityRequest(http.MethodGet, "/api/v1/imagery", nil)

	srv.listImageryHandler(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d when mock actors disabled, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestIDsDoNotRepeatAfterDelete(t *testing.T) {
	srv := newTestServer(t)

	first := uploadRecord(t, srv, "drone", "first.png", []byte("\x89PNG\r\nfirst"), "image/png")
	second := uploadRecord(t, srv, "drone", "second.png", []byte("\x89PNG\r\nsecond"), "image/png")

	deleteResponse := httptest.NewRecorder()
	deleteRequest := authorityRequest(http.MethodDelete, "/api/v1/imagery/"+first.ID, nil)
	deleteRequest.SetPathValue("id", first.ID)
	srv.deleteImageryHandler(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResponse.Code)
	}

	third := uploadRecord(t, srv, "drone", "third.png", []byte("\x89PNG\r\nthird"), "image/png")
	if third.ID == first.ID || third.ID == second.ID {
		t.Fatalf("expected fresh id after delete, got %s (deleted %s, existing %s)", third.ID, first.ID, second.ID)
	}

	getResponse := httptest.NewRecorder()
	getRequest := authorityRequest(http.MethodGet, "/api/v1/imagery/"+third.ID, nil)
	getRequest.SetPathValue("id", third.ID)
	srv.getImageryHandler(getResponse, getRequest)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d for new record, got %d", http.StatusOK, getResponse.Code)
	}

	downloadResponse := httptest.NewRecorder()
	downloadRequest := authorityRequest(http.MethodGet, "/api/v1/imagery/"+third.ID+"/download", nil)
	downloadRequest.SetPathValue("id", third.ID)
	srv.downloadImageryHandler(downloadResponse, downloadRequest)
	if downloadResponse.Code != http.StatusOK {
		t.Fatalf("expected status %d for new record download, got %d", http.StatusOK, downloadResponse.Code)
	}
	if downloadResponse.Body.String() != "\x89PNG\r\nthird" {
		t.Fatalf("expected downloaded file contents, got %q", downloadResponse.Body.String())
	}
}

func TestCreateImageryRejectsNonFiniteNumbers(t *testing.T) {
	cases := []struct {
		name   string
		fields map[string]string
	}{
		{"nan_coverage", map[string]string{"coverageAreaKm2": "NaN"}},
		{"inf_coverage", map[string]string{"coverageAreaKm2": "Inf"}},
		{"nan_resolution", map[string]string{"resolutionMeters": "NaN"}},
		{"inf_resolution", map[string]string{"resolutionMeters": "+Inf"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(t)
			fields := map[string]string{
				"source":           "drone",
				"captureTime":      "2026-07-05T10:30:00Z",
				"geometry":         `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.19,5.59],[-0.22,5.59],[-0.22,5.56]]]}`,
				"coverageAreaKm2":  "4.5",
				"resolutionMeters": "0.25",
			}
			maps.Copy(fields, tc.fields)
			response := httptest.NewRecorder()
			request := uploadRequest(fields, "test.png", []byte("\x89PNG\r\n\x1a\n"), "image/png")

			srv.createImageryHandler(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
			}
		})
	}
}

func TestCreateImageryRejectsInvalidGeometry(t *testing.T) {
	cases := []struct {
		name     string
		geometry string
	}{
		{"missing_coordinates", `{"type":"Polygon"}`},
		{"empty_coordinates", `{"type":"Polygon","coordinates":[]}`},
		{"short_ring", `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.22,5.56]]]}`},
		{"open_ring", `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.19,5.59],[-0.22,5.59]]]}`},
		{"invalid_position", `{"type":"Polygon","coordinates":[[[-0.22],[-0.19,5.56],[-0.19,5.59],[-0.22]]]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(t)
			fields := map[string]string{
				"source":           "drone",
				"captureTime":      "2026-07-05T10:30:00Z",
				"geometry":         tc.geometry,
				"coverageAreaKm2":  "4.5",
				"resolutionMeters": "0.25",
			}
			response := httptest.NewRecorder()
			request := uploadRequest(fields, "test.png", []byte("\x89PNG\r\n\x1a\n"), "image/png")

			srv.createImageryHandler(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, response.Code, response.Body.String())
			}
		})
	}
}

func TestCreateImageryWriteFailureLeavesNoPhantomRecord(t *testing.T) {
	now := testNow
	// StoragePath points into a directory that does not exist, so the file
	// write fails after the metadata record is created.
	cfg := &config.Config{
		Addr:            ":8099",
		StoragePath:     filepath.Join(t.TempDir(), "missing", "uploads"),
		RetentionDays:   90,
		TokenSecret:     testTokenSecret,
		AllowMockActors: true,
	}
	srv := NewServer(store.NewMemoryStore(now, 90), func() time.Time { return now }, cfg)

	fields := map[string]string{
		"source":           "drone",
		"captureTime":      "2026-07-05T10:30:00Z",
		"geometry":         `{"type":"Polygon","coordinates":[[[-0.22,5.56],[-0.19,5.56],[-0.19,5.59],[-0.22,5.59],[-0.22,5.56]]]}`,
		"coverageAreaKm2":  "4.5",
		"resolutionMeters": "0.25",
	}
	response := httptest.NewRecorder()
	request := uploadRequest(fields, "test.png", []byte("\x89PNG\r\n\x1a\n"), "image/png")

	srv.createImageryHandler(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d: %s", http.StatusInternalServerError, response.Code, response.Body.String())
	}

	listResponse := httptest.NewRecorder()
	listRequest := authorityRequest(http.MethodGet, "/api/v1/imagery", nil)
	srv.listImageryHandler(listResponse, listRequest)

	var payload models.ImageryListResponse
	decodeResponse(t, listResponse, &payload)
	seedCount := 2 // seeded fixtures only; the failed upload must not linger
	if len(payload.Imagery) != seedCount {
		t.Fatalf("expected %d records after failed write, got %d: %#v", seedCount, len(payload.Imagery), payload.Imagery)
	}
}

func TestGeoJSONUsesConfiguredPublicBaseURL(t *testing.T) {
	srv := newTestServer(t)
	srv.config.PublicBaseURL = "https://imagery.nadaa.gov.gh"

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/imagery/geojson", nil)
	request.Host = "untrusted.example.com"

	srv.geoJSONHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var payload models.GeoJSONFeatureCollection
	decodeResponse(t, response, &payload)
	if len(payload.Features) == 0 {
		t.Fatalf("expected at least one feature, got %#v", payload)
	}
	for _, feature := range payload.Features {
		downloadURL, ok := feature.Properties["downloadUrl"].(string)
		if !ok || !strings.HasPrefix(downloadURL, "https://imagery.nadaa.gov.gh/api/v1/imagery/") {
			t.Fatalf("expected downloadUrl on configured base URL, got %#v", feature.Properties["downloadUrl"])
		}
	}
}
