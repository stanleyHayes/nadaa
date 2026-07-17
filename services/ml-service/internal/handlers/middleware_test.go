package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

var (
	testTokenSecret   = []byte("test-token-secret")
	testServiceToken  = "test-internal-token"
	testNow           = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	securedTestClaims = utils.TokenClaims{
		UserID:   "user_001",
		UserType: "agency",
		Role:     "nadmo_officer",
		AgencyID: "agency_nadmo",
		District: "accra",
		MFA:      true,
	}
)

func newSecuredTestServer(t *testing.T) *server {
	t.Helper()
	cfg := &config.Config{
		Addr:                 ":8094",
		TokenSecret:          testTokenSecret,
		InternalServiceToken: testServiceToken,
	}
	s, err := store.NewMemoryStore("../../../../data/flood-risk/models")
	if err != nil {
		t.Fatalf("new memory store: %v", err)
	}
	return NewServer(s, func() time.Time { return testNow }, cfg)
}

func signTestToken(t *testing.T, secret []byte, claims utils.TokenClaims) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(encodedPayload))
	return "nadaa." + encodedPayload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func getForecasts(t *testing.T, srv *server, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecasts", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	return rr
}

func TestHealthzStaysPublic(t *testing.T) {
	srv := newSecuredTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}
}

func TestInternalAccessRejectsMissingOrWrongCredentials(t *testing.T) {
	srv := newSecuredTestServer(t)

	if rr := getForecasts(t, srv, nil); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 without credentials got %d", rr.Code)
	}
	if rr := getForecasts(t, srv, map[string]string{serviceTokenHeader: "wrong-token"}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 with wrong service token got %d", rr.Code)
	}
	if rr := getForecasts(t, srv, map[string]string{"Authorization": "Bearer not-a-nadaa-token"}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 with malformed bearer token got %d", rr.Code)
	}
}

func TestInternalAccessAllowsServiceToken(t *testing.T) {
	srv := newSecuredTestServer(t)
	rr := getForecasts(t, srv, map[string]string{serviceTokenHeader: testServiceToken})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 with service token got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestInternalAccessAllowsVerifiedBearerToken(t *testing.T) {
	srv := newSecuredTestServer(t)
	claims := securedTestClaims
	claims.ExpiresAt = testNow.Add(time.Hour).Unix()
	token := signTestToken(t, testTokenSecret, claims)

	rr := getForecasts(t, srv, map[string]string{"Authorization": "Bearer " + token})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 with verified bearer token got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestInternalAccessRejectsExpiredAndForgedTokens(t *testing.T) {
	srv := newSecuredTestServer(t)

	expired := securedTestClaims
	expired.ExpiresAt = testNow.Add(-time.Hour).Unix()
	expiredToken := signTestToken(t, testTokenSecret, expired)
	if rr := getForecasts(t, srv, map[string]string{"Authorization": "Bearer " + expiredToken}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 with expired token got %d", rr.Code)
	}

	forged := securedTestClaims
	forged.ExpiresAt = testNow.Add(time.Hour).Unix()
	forgedToken := signTestToken(t, []byte("attacker-secret"), forged)
	if rr := getForecasts(t, srv, map[string]string{"Authorization": "Bearer " + forgedToken}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 with forged signature got %d", rr.Code)
	}
}

func TestInternalAccessEmptySecretNeverVerifies(t *testing.T) {
	srv := newSecuredTestServer(t)
	srv.config.TokenSecret = nil

	claims := securedTestClaims
	claims.ExpiresAt = testNow.Add(time.Hour).Unix()
	token := signTestToken(t, nil, claims)
	if rr := getForecasts(t, srv, map[string]string{"Authorization": "Bearer " + token}); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 with empty token secret got %d", rr.Code)
	}
}

func TestInternalAccessMockActors(t *testing.T) {
	srv := newSecuredTestServer(t)
	mockHeaders := map[string]string{"X-NADAA-Actor-ID": "dev-actor", "X-NADAA-Actor-Role": "nadmo_officer"}

	// Mock actor headers are ignored unless explicitly enabled.
	if rr := getForecasts(t, srv, mockHeaders); rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 with mock actors disabled got %d", rr.Code)
	}

	srv.config.AllowMockActors = true
	if rr := getForecasts(t, srv, mockHeaders); rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 with mock actors enabled got %d", rr.Code)
	}
}

func TestInternalAccessOpenWhenServiceTokenUnset(t *testing.T) {
	// The development default (no NADAA_INTERNAL_SERVICE_TOKEN) keeps the
	// service-token path open, as exercised by the remaining handler tests.
	srv := newTestServer(t)
	if rr := getForecasts(t, srv, nil); rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 in development default got %d", rr.Code)
	}
}
