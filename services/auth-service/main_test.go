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

func TestAgencyRoleCatalogCoversAuthorityRoles(t *testing.T) {
	for _, role := range []string{
		roleSystemAdmin,
		roleAgencyAdmin,
		roleNADMOOfficer,
		roleDistrictOfficer,
		roleDispatcher,
		roleResponder,
		roleAgencyViewer,
	} {
		if !validAgencyRole(role) {
			t.Fatalf("expected role %q to be valid", role)
		}
	}

	if validAgencyRole(roleCitizen) {
		t.Fatal("citizen must not be valid for agency-user creation")
	}
}

func TestAgencyUserCreationSetupMFALoginAndProfile(t *testing.T) {
	srv := newTestServer()
	_, adminToken := seedVerifiedAgencyUser(t, srv, roleSystemAdmin, "admin@nadaa.local", "+233200000001")

	create := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users", jsonBody(createAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: defaultAgencyID,
		Role:     roleDispatcher,
	}))
	createRequest.Header.Set("Authorization", "Bearer "+adminToken)
	srv.createAgencyUserHandler(create, createRequest)

	if create.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, create.Code, create.Body.String())
	}

	var createPayload createAgencyUserResponse
	decodeResponse(t, create, &createPayload)
	if createPayload.User.Role != roleDispatcher || !createPayload.User.MFARequired || createPayload.User.MFAEnabled {
		t.Fatalf("unexpected created agency user payload: %#v", createPayload.User)
	}
	if createPayload.TemporaryPassword == "" || !createPayload.MFASetupRequired {
		t.Fatalf("expected temporary password and MFA setup requirement: %#v", createPayload)
	}

	loginBeforeMFA := httptest.NewRecorder()
	srv.loginAgencyHandler(loginBeforeMFA, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(loginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
		MFACode:  "123456",
	})))
	if loginBeforeMFA.Code != http.StatusForbidden {
		t.Fatalf("expected status %d before MFA setup, got %d", http.StatusForbidden, loginBeforeMFA.Code)
	}

	setup := httptest.NewRecorder()
	setupRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users/"+createPayload.User.ID+"/mfa/setup", jsonBody(agencyMFASetupRequest{
		Email:             "dispatcher@nadaa.local",
		TemporaryPassword: createPayload.TemporaryPassword,
	}))
	setupRequest.SetPathValue("id", createPayload.User.ID)
	srv.setupAgencyMFAHandler(setup, setupRequest)
	if setup.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, setup.Code, setup.Body.String())
	}

	var setupPayload agencyMFASetupResponse
	decodeResponse(t, setup, &setupPayload)
	if setupPayload.DevCode != "123456" || setupPayload.Secret == "" || setupPayload.ChallengeID == "" {
		t.Fatalf("expected exposed dev MFA challenge, got %#v", setupPayload)
	}

	verify := httptest.NewRecorder()
	verifyRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users/"+createPayload.User.ID+"/mfa/verify", jsonBody(agencyMFAVerifyRequest{
		Email:             "dispatcher@nadaa.local",
		TemporaryPassword: createPayload.TemporaryPassword,
		Code:              setupPayload.DevCode,
	}))
	verifyRequest.SetPathValue("id", createPayload.User.ID)
	srv.verifyAgencyMFAHandler(verify, verifyRequest)
	if verify.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, verify.Code, verify.Body.String())
	}

	var verifyPayload agencyMFAVerifyResponse
	decodeResponse(t, verify, &verifyPayload)
	if !verifyPayload.User.MFAEnabled {
		t.Fatalf("expected MFA to be enabled, got %#v", verifyPayload.User)
	}

	loginMissingMFA := httptest.NewRecorder()
	srv.loginAgencyHandler(loginMissingMFA, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(loginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
	})))
	if loginMissingMFA.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for missing MFA code, got %d", http.StatusUnauthorized, loginMissingMFA.Code)
	}

	login := httptest.NewRecorder()
	srv.loginAgencyHandler(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(loginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
		MFACode:  setupPayload.DevCode,
	})))
	if login.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, login.Code, login.Body.String())
	}

	var loginPayload loginAgencyResponse
	decodeResponse(t, login, &loginPayload)
	if loginPayload.AccessToken == "" || loginPayload.User.Agency.ID != defaultAgencyID {
		t.Fatalf("expected agency token and agency profile, got %#v", loginPayload)
	}

	profile := httptest.NewRecorder()
	profileRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	profileRequest.Header.Set("Authorization", "Bearer "+loginPayload.AccessToken)
	srv.meHandler(profile, profileRequest)
	if profile.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, profile.Code, profile.Body.String())
	}

	var profilePayload agencyUserProfile
	decodeResponse(t, profile, &profilePayload)
	if profilePayload.Role != roleDispatcher || profilePayload.Email != "dispatcher@nadaa.local" {
		t.Fatalf("expected dispatcher profile, got %#v", profilePayload)
	}
}

func TestAgencyUserCreationRejectsUnauthorizedRoles(t *testing.T) {
	srv := newTestServer()
	_, dispatcherToken := seedVerifiedAgencyUser(t, srv, roleDispatcher, "dispatcher@nadaa.local", "+233200000002")

	create := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users", jsonBody(createAgencyUserRequest{
		Name:     "Viewer One",
		Email:    "viewer@nadaa.local",
		Phone:    "+233200000003",
		AgencyID: defaultAgencyID,
		Role:     roleAgencyViewer,
	}))
	request.Header.Set("Authorization", "Bearer "+dispatcherToken)
	srv.createAgencyUserHandler(create, request)

	if create.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, create.Code, create.Body.String())
	}
}

func seedVerifiedAgencyUser(t *testing.T, srv *server, role string, email string, phone string) (agencyUserProfile, string) {
	t.Helper()

	password := "Password123!"
	profile, err := srv.store.createAgencyUser(createAgencyUserRequest{
		Name:     "Seed " + role,
		Email:    email,
		Phone:    phone,
		AgencyID: defaultAgencyID,
		Role:     role,
	}, password, srv.now())
	if err != nil {
		t.Fatalf("seed agency user: %v", err)
	}

	challenge, err := srv.store.startAgencyMFASetup(profile.ID, email, password, newMFASecret(), "123456", srv.now())
	if err != nil {
		t.Fatalf("seed MFA setup: %v", err)
	}

	profile, err = srv.store.verifyAgencyMFA(profile.ID, email, password, challenge.Code, srv.now())
	if err != nil {
		t.Fatalf("seed MFA verify: %v", err)
	}

	token, err := srv.signAgencyToken(profile, srv.now().Add(12*time.Hour))
	if err != nil {
		t.Fatalf("seed agency token: %v", err)
	}

	return profile, token
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
