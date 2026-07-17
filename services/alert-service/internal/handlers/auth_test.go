package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/store"
)

const testTokenSecret = "test-alert-token-secret"

// newTokenTestServer builds a server with a token secret and mock actors off,
// so only verified bearer tokens authenticate authority requests.
func newTokenTestServer() *Server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8089", TokenSecret: testTokenSecret}
	return NewServer(store.NewMemoryStore(now), func() time.Time { return time.Now().UTC() }, cfg)
}

func signTestToken(t *testing.T, secret string, claims tokenClaims) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(encodedPayload))
	return "nadaa." + encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func drafterToken(t *testing.T) string {
	t.Helper()
	return signTestToken(t, testTokenSecret, tokenClaims{
		UserID:    "usr_drafter",
		UserType:  "agency",
		Role:      "district_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		District:  "accra-metropolitan",
		MFA:       true,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
}

func approverToken(t *testing.T) string {
	t.Helper()
	return signTestToken(t, testTokenSecret, tokenClaims{
		UserID:    "usr_approver",
		UserType:  "agency",
		Role:      "nadmo_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		MFA:       true,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
}

func tokenRequest(method string, path string, body string, token string) *http.Request {
	var reader *bytes.Buffer
	if body == "" {
		reader = bytes.NewBuffer(nil)
	} else {
		reader = bytes.NewBufferString(body)
	}
	request := httptest.NewRequest(method, path, reader)
	request.SetPathValue("id", pathID(path))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func TestBearerTokenAuthenticatesAuthorityWorkflow(t *testing.T) {
	srv := newTokenTestServer()

	createResponse := httptest.NewRecorder()
	srv.createAlertHandler(createResponse, tokenRequest(http.MethodPost, "/api/v1/alerts", validAlertBody(), drafterToken(t)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d: %s", http.StatusCreated, createResponse.Code, createResponse.Body.String())
	}

	var draft models.AuthorityAlert
	decodeResponse(t, createResponse, &draft)
	if draft.IssuedBy != "usr_drafter" || draft.IssuingAgencyID != "00000000-0000-0000-0000-000000000101" {
		t.Fatalf("expected actor context from verified claims, got %#v", draft)
	}

	submitResponse := httptest.NewRecorder()
	srv.submitAlertHandler(submitResponse, tokenRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/submit", "", drafterToken(t)))
	if submitResponse.Code != http.StatusOK {
		t.Fatalf("expected submit status %d, got %d: %s", http.StatusOK, submitResponse.Code, submitResponse.Body.String())
	}

	approveResponse := httptest.NewRecorder()
	srv.approveAlertHandler(approveResponse, tokenRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/approve", `{"note":"Reviewed"}`, approverToken(t)))
	if approveResponse.Code != http.StatusOK {
		t.Fatalf("expected approve status %d, got %d: %s", http.StatusOK, approveResponse.Code, approveResponse.Body.String())
	}

	logs := srv.store.ListAudit(10)
	if len(logs) != 3 || logs[2].ActorUserID != "usr_drafter" || logs[0].ActorUserID != "usr_approver" {
		t.Fatalf("expected audit actors from verified claims, got %#v", logs)
	}
}

func TestBearerTokenWithoutMFACannotWrite(t *testing.T) {
	srv := newTokenTestServer()
	token := signTestToken(t, testTokenSecret, tokenClaims{
		UserID:    "usr_drafter",
		UserType:  "agency",
		Role:      "district_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		MFA:       false,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})

	response := httptest.NewRecorder()
	srv.createAlertHandler(response, tokenRequest(http.MethodPost, "/api/v1/alerts", validAlertBody(), token))
	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, response.Code)
	}
}

func TestMissingForgedOrInvalidTokensAreUnauthorized(t *testing.T) {
	srv := newTokenTestServer()

	// No credentials at all.
	response := httptest.NewRecorder()
	srv.createAlertHandler(response, httptest.NewRequest(http.MethodPost, "/api/v1/alerts", bytes.NewBufferString(validAlertBody())))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d without credentials, got %d", http.StatusUnauthorized, response.Code)
	}

	// Forged legacy headers are ignored when mock actors are off.
	forged := authorizedRequest(http.MethodPost, "/api/v1/alerts", validAlertBody())
	response = httptest.NewRecorder()
	srv.createAlertHandler(response, forged)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for forged headers, got %d", http.StatusUnauthorized, response.Code)
	}

	// Token signed with the wrong secret.
	wrongSecret := signTestToken(t, "attacker-secret", tokenClaims{
		UserID:    "usr_attacker",
		UserType:  "agency",
		Role:      "system_admin",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		MFA:       true,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
	response = httptest.NewRecorder()
	srv.createAlertHandler(response, tokenRequest(http.MethodPost, "/api/v1/alerts", validAlertBody(), wrongSecret))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for wrong-secret token, got %d", http.StatusUnauthorized, response.Code)
	}

	// Expired token.
	expired := signTestToken(t, testTokenSecret, tokenClaims{
		UserID:    "usr_drafter",
		UserType:  "agency",
		Role:      "district_officer",
		AgencyID:  "00000000-0000-0000-0000-000000000101",
		MFA:       true,
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
	})
	response = httptest.NewRecorder()
	srv.createAlertHandler(response, tokenRequest(http.MethodPost, "/api/v1/alerts", validAlertBody(), expired))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for expired token, got %d", http.StatusUnauthorized, response.Code)
	}

	// Malformed token.
	response = httptest.NewRecorder()
	srv.createAlertHandler(response, tokenRequest(http.MethodPost, "/api/v1/alerts", validAlertBody(), "nadaa.not-a-token"))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for malformed token, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestListAlertsTreatsForgedHeadersAsPublic(t *testing.T) {
	srv := newTokenTestServer()

	createResponse := httptest.NewRecorder()
	srv.createAlertHandler(createResponse, tokenRequest(http.MethodPost, "/api/v1/alerts", alertBodyWithSourcePrediction(), drafterToken(t)))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create alert: %s", createResponse.Body.String())
	}
	var draft models.AuthorityAlert
	decodeResponse(t, createResponse, &draft)

	overrideResponse := httptest.NewRecorder()
	srv.emergencyOverrideHandler(overrideResponse, tokenRequest(http.MethodPost, "/api/v1/alerts/"+draft.ID+"/emergency-override", `{"reason":"Immediate life-safety warning"}`, approverToken(t)))
	if overrideResponse.Code != http.StatusOK {
		t.Fatalf("override failed: %s", overrideResponse.Body.String())
	}

	// Forged headers without a valid token must get the public view.
	forged := authorizedRequest(http.MethodGet, "/api/v1/alerts?current=true", "")
	response := httptest.NewRecorder()
	srv.listAlertsHandler(response, forged)
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, response.Code)
	}
	var publicPayload models.AlertListResponse
	decodeResponse(t, response, &publicPayload)
	if len(publicPayload.Alerts) != 1 || publicPayload.Alerts[0].SourcePrediction != nil {
		t.Fatalf("expected public view to hide source prediction from forged headers, got %#v", publicPayload.Alerts)
	}

	// A verified authority token gets the internal view.
	response = httptest.NewRecorder()
	srv.listAlertsHandler(response, tokenRequest(http.MethodGet, "/api/v1/alerts?current=true", "", approverToken(t)))
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, response.Code)
	}
	var authorityPayload models.AlertListResponse
	decodeResponse(t, response, &authorityPayload)
	if len(authorityPayload.Alerts) != 1 || authorityPayload.Alerts[0].SourcePrediction == nil {
		t.Fatalf("expected verified authority to see source prediction, got %#v", authorityPayload.Alerts)
	}

	// An invalid token also gets the public view, not the internal one.
	response = httptest.NewRecorder()
	srv.listAlertsHandler(response, tokenRequest(http.MethodGet, "/api/v1/alerts?current=true", "", strings.TrimSuffix(drafterToken(t), "a")+"x"))
	if response.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d", http.StatusOK, response.Code)
	}
	var invalidTokenPayload models.AlertListResponse
	decodeResponse(t, response, &invalidTokenPayload)
	if len(invalidTokenPayload.Alerts) != 1 || invalidTokenPayload.Alerts[0].SourcePrediction != nil {
		t.Fatalf("expected public view for invalid token, got %#v", invalidTokenPayload.Alerts)
	}
}
