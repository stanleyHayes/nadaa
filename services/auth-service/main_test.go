package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer() *server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	return &server{
		store:        newMemoryStore(),
		tokenSecret:  []byte("test-secret"),
		otp:          fixedOTPGenerator{code: "123456"},
		now:          func() time.Time { return now },
		exposeDevOTP: true,
	}
}

func TestRegisterCitizen(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(registerCitizenRequest{
		Name:              "Ama Mensah",
		Phone:             "+233200000000",
		PreferredLanguage: "en",
		HomeLocation:      &coordinates{Lat: 5.6037, Lng: -0.1870},
		ContactPermission: true,
	}))

	srv.registerCitizenHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload registerCitizenResponse
	decodeResponse(t, response, &payload)

	if payload.UserID == "" || payload.ChallengeID == "" {
		t.Fatalf("expected user and challenge ids, got %#v", payload)
	}

	if payload.DevOTP != "123456" {
		t.Fatalf("expected dev OTP to be exposed in test server, got %q", payload.DevOTP)
	}
}

func TestRegisterCitizenRejectsDuplicatePhone(t *testing.T) {
	srv := newTestServer()
	body := registerCitizenRequest{Name: "Ama Mensah", Phone: "+233200000000"}

	first := httptest.NewRecorder()
	srv.registerCitizenHandler(first, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(body)))

	second := httptest.NewRecorder()
	srv.registerCitizenHandler(second, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(body)))

	if second.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, second.Code)
	}
}

func TestLoginCitizenAndReadProfile(t *testing.T) {
	srv := newTestServer()
	register := httptest.NewRecorder()
	srv.registerCitizenHandler(register, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(registerCitizenRequest{
		Name:              "Ama Mensah",
		Phone:             "+233200000000",
		PreferredLanguage: "en",
		ContactPermission: true,
	})))

	login := httptest.NewRecorder()
	srv.loginCitizenHandler(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/login", jsonBody(loginCitizenRequest{
		Phone: "+233200000000",
		OTP:   "123456",
	})))

	if login.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, login.Code, login.Body.String())
	}

	var loginPayload loginCitizenResponse
	decodeResponse(t, login, &loginPayload)
	if loginPayload.AccessToken == "" {
		t.Fatal("expected access token")
	}

	profile := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+loginPayload.AccessToken)
	srv.meHandler(profile, request)

	if profile.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, profile.Code, profile.Body.String())
	}

	var profilePayload citizenProfile
	decodeResponse(t, profile, &profilePayload)
	if profilePayload.Phone != "+233200000000" {
		t.Fatalf("expected profile phone, got %#v", profilePayload)
	}
}

func TestLoginCitizenRejectsInvalidOTP(t *testing.T) {
	srv := newTestServer()
	register := httptest.NewRecorder()
	srv.registerCitizenHandler(register, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(registerCitizenRequest{
		Name:  "Ama Mensah",
		Phone: "+233200000000",
	})))

	login := httptest.NewRecorder()
	srv.loginCitizenHandler(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/login", jsonBody(loginCitizenRequest{
		Phone: "+233200000000",
		OTP:   "000000",
	})))

	if login.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, login.Code)
	}
}

func TestRegisterCitizenRejectsInvalidCoordinates(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(registerCitizenRequest{
		Name:         "Ama Mensah",
		Phone:        "+233200000000",
		HomeLocation: &coordinates{Lat: 100, Lng: -0.1870},
	}))

	srv.registerCitizenHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func jsonBody(value any) *bytes.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
