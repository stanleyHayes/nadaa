package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

func newTestServer() *Server {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{Addr: ":8080"}
	return &Server{
		store:        store.NewMemoryStore(now, cfg),
		tokenSecret:  []byte("test-secret"),
		otp:          utils.FixedOTPGenerator{Code: "123456"},
		now:          func() time.Time { return now },
		exposeDevOTP: true,
		config:       cfg,
	}
}

func TestRegisterCitizen(t *testing.T) {
	srv := newTestServer()
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(models.RegisterCitizenRequest{
		Name:              "Ama Mensah",
		Phone:             "+233200000000",
		PreferredLanguage: "en",
		HomeLocation:      &models.Coordinates{Lat: 5.6037, Lng: -0.1870},
		ContactPermission: true,
	}))

	srv.registerCitizenHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	var payload models.RegisterCitizenResponse
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
	body := models.RegisterCitizenRequest{Name: "Ama Mensah", Phone: "+233200000000"}

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
	srv.registerCitizenHandler(register, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(models.RegisterCitizenRequest{
		Name:              "Ama Mensah",
		Phone:             "+233200000000",
		PreferredLanguage: "en",
		ContactPermission: true,
	})))

	login := httptest.NewRecorder()
	srv.loginCitizenHandler(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/login", jsonBody(models.LoginCitizenRequest{
		Phone: "+233200000000",
		OTP:   "123456",
	})))

	if login.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, login.Code, login.Body.String())
	}

	var loginPayload models.LoginCitizenResponse
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

	var profilePayload models.CitizenProfile
	decodeResponse(t, profile, &profilePayload)
	if profilePayload.Phone != "+233200000000" {
		t.Fatalf("expected profile phone, got %#v", profilePayload)
	}
}

func TestLoginCitizenRejectsInvalidOTP(t *testing.T) {
	srv := newTestServer()
	register := httptest.NewRecorder()
	srv.registerCitizenHandler(register, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(models.RegisterCitizenRequest{
		Name:  "Ama Mensah",
		Phone: "+233200000000",
	})))

	login := httptest.NewRecorder()
	srv.loginCitizenHandler(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/login", jsonBody(models.LoginCitizenRequest{
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
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(models.RegisterCitizenRequest{
		Name:         "Ama Mensah",
		Phone:        "+233200000000",
		HomeLocation: &models.Coordinates{Lat: 100, Lng: -0.1870},
	}))

	srv.registerCitizenHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestAgencyRoleCatalogCoversAuthorityRoles(t *testing.T) {
	for _, role := range []string{
		models.RoleSystemAdmin,
		models.RoleAgencyAdmin,
		models.RoleNADMOOfficer,
		models.RoleDistrictOfficer,
		models.RoleDispatcher,
		models.RoleResponder,
		models.RoleAgencyViewer,
	} {
		if !utils.ValidAgencyRole(role) {
			t.Fatalf("expected role %q to be valid", role)
		}
	}

	if utils.ValidAgencyRole(models.RoleCitizen) {
		t.Fatal("citizen must not be valid for agency-user creation")
	}
}

func TestAgencyUserCreationSetupMFALoginAndProfile(t *testing.T) {
	srv := newTestServer()
	_, adminToken := seedVerifiedAgencyUser(t, srv, models.RoleSystemAdmin, "admin@nadaa.local", "+233200000001")

	create := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users", jsonBody(models.CreateAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleDispatcher,
	}))
	createRequest.Header.Set("Authorization", "Bearer "+adminToken)
	srv.createAgencyUserHandler(create, createRequest)

	if create.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, create.Code, create.Body.String())
	}

	var createPayload models.CreateAgencyUserResponse
	decodeResponse(t, create, &createPayload)
	if createPayload.User.Role != models.RoleDispatcher || !createPayload.User.MFARequired || createPayload.User.MFAEnabled {
		t.Fatalf("unexpected created agency user payload: %#v", createPayload.User)
	}
	if createPayload.TemporaryPassword == "" || !createPayload.MFASetupRequired {
		t.Fatalf("expected temporary password and MFA setup requirement: %#v", createPayload)
	}

	loginBeforeMFA := httptest.NewRecorder()
	srv.loginAgencyHandler(loginBeforeMFA, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(models.LoginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
		MFACode:  "123456",
	})))
	if loginBeforeMFA.Code != http.StatusForbidden {
		t.Fatalf("expected status %d before MFA setup, got %d", http.StatusForbidden, loginBeforeMFA.Code)
	}

	setup := httptest.NewRecorder()
	setupRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users/"+createPayload.User.ID+"/mfa/setup", jsonBody(models.AgencyMFASetupRequest{
		Email:             "dispatcher@nadaa.local",
		TemporaryPassword: createPayload.TemporaryPassword,
	}))
	setupRequest.SetPathValue("id", createPayload.User.ID)
	srv.setupAgencyMFAHandler(setup, setupRequest)
	if setup.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, setup.Code, setup.Body.String())
	}

	var setupPayload models.AgencyMFASetupResponse
	decodeResponse(t, setup, &setupPayload)
	if setupPayload.DevCode != "123456" || setupPayload.Secret == "" || setupPayload.ChallengeID == "" {
		t.Fatalf("expected exposed dev MFA challenge, got %#v", setupPayload)
	}

	verify := httptest.NewRecorder()
	verifyRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users/"+createPayload.User.ID+"/mfa/verify", jsonBody(models.AgencyMFAVerifyRequest{
		Email:             "dispatcher@nadaa.local",
		TemporaryPassword: createPayload.TemporaryPassword,
		Code:              setupPayload.DevCode,
	}))
	verifyRequest.SetPathValue("id", createPayload.User.ID)
	srv.verifyAgencyMFAHandler(verify, verifyRequest)
	if verify.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, verify.Code, verify.Body.String())
	}

	var verifyPayload models.AgencyMFAVerifyResponse
	decodeResponse(t, verify, &verifyPayload)
	if !verifyPayload.User.MFAEnabled {
		t.Fatalf("expected MFA to be enabled, got %#v", verifyPayload.User)
	}

	loginMissingMFA := httptest.NewRecorder()
	srv.loginAgencyHandler(loginMissingMFA, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(models.LoginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
	})))
	if loginMissingMFA.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d for missing MFA code, got %d", http.StatusUnauthorized, loginMissingMFA.Code)
	}

	login := httptest.NewRecorder()
	srv.loginAgencyHandler(login, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(models.LoginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
		MFACode:  setupPayload.DevCode,
	})))
	if login.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, login.Code, login.Body.String())
	}

	var loginPayload models.LoginAgencyResponse
	decodeResponse(t, login, &loginPayload)
	if loginPayload.AccessToken == "" || loginPayload.User.Agency.ID != models.DefaultAgencyID {
		t.Fatalf("expected agency token and agency profile, got %#v", loginPayload)
	}

	profile := httptest.NewRecorder()
	profileRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	profileRequest.Header.Set("Authorization", "Bearer "+loginPayload.AccessToken)
	srv.meHandler(profile, profileRequest)
	if profile.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, profile.Code, profile.Body.String())
	}

	var profilePayload models.AgencyUserProfile
	decodeResponse(t, profile, &profilePayload)
	if profilePayload.Role != models.RoleDispatcher || profilePayload.Email != "dispatcher@nadaa.local" {
		t.Fatalf("expected dispatcher profile, got %#v", profilePayload)
	}
}

func TestAgencyUserCreationRejectsUnauthorizedRoles(t *testing.T) {
	srv := newTestServer()
	_, dispatcherToken := seedVerifiedAgencyUser(t, srv, models.RoleDispatcher, "dispatcher@nadaa.local", "+233200000002")

	create := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users", jsonBody(models.CreateAgencyUserRequest{
		Name:     "Viewer One",
		Email:    "viewer@nadaa.local",
		Phone:    "+233200000003",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleAgencyViewer,
	}))
	request.Header.Set("Authorization", "Bearer "+dispatcherToken)
	srv.createAgencyUserHandler(create, request)

	if create.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, create.Code, create.Body.String())
	}
}

func TestAuditLogsCaptureAuthAndAdminEvents(t *testing.T) {
	srv := newTestServer()
	_, adminToken := seedVerifiedAgencyUser(t, srv, models.RoleSystemAdmin, "admin@nadaa.local", "+233200000001")

	register := httptest.NewRecorder()
	srv.registerCitizenHandler(register, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/register", jsonBody(models.RegisterCitizenRequest{
		Name:  "Ama Mensah",
		Phone: "+233200000000",
	})))
	if register.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, register.Code)
	}

	citizenLogin := httptest.NewRecorder()
	srv.loginCitizenHandler(citizenLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/citizens/login", jsonBody(models.LoginCitizenRequest{
		Phone: "+233200000000",
		OTP:   "123456",
	})))
	if citizenLogin.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, citizenLogin.Code)
	}

	create := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users", jsonBody(models.CreateAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleDispatcher,
	}))
	createRequest.Header.Set("Authorization", "Bearer "+adminToken)
	createRequest.Header.Set("X-Request-ID", "req-create-dispatcher")
	createRequest.Header.Set("X-Forwarded-For", "203.0.113.10")
	createRequest.Header.Set("User-Agent", "nadaa-test/1.0")
	srv.createAgencyUserHandler(create, createRequest)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, create.Code, create.Body.String())
	}

	var createPayload models.CreateAgencyUserResponse
	decodeResponse(t, create, &createPayload)

	setup := httptest.NewRecorder()
	setupRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users/"+createPayload.User.ID+"/mfa/setup", jsonBody(models.AgencyMFASetupRequest{
		Email:             "dispatcher@nadaa.local",
		TemporaryPassword: createPayload.TemporaryPassword,
	}))
	setupRequest.SetPathValue("id", createPayload.User.ID)
	srv.setupAgencyMFAHandler(setup, setupRequest)
	if setup.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, setup.Code, setup.Body.String())
	}

	var setupPayload models.AgencyMFASetupResponse
	decodeResponse(t, setup, &setupPayload)

	verify := httptest.NewRecorder()
	verifyRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency-users/"+createPayload.User.ID+"/mfa/verify", jsonBody(models.AgencyMFAVerifyRequest{
		Email:             "dispatcher@nadaa.local",
		TemporaryPassword: createPayload.TemporaryPassword,
		Code:              setupPayload.DevCode,
	}))
	verifyRequest.SetPathValue("id", createPayload.User.ID)
	srv.verifyAgencyMFAHandler(verify, verifyRequest)
	if verify.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, verify.Code, verify.Body.String())
	}

	agencyLogin := httptest.NewRecorder()
	srv.loginAgencyHandler(agencyLogin, httptest.NewRequest(http.MethodPost, "/api/v1/auth/agency/login", jsonBody(models.LoginAgencyRequest{
		Email:    "dispatcher@nadaa.local",
		Password: createPayload.TemporaryPassword,
		MFACode:  setupPayload.DevCode,
	})))
	if agencyLogin.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, agencyLogin.Code, agencyLogin.Body.String())
	}

	for _, action := range []string{
		"auth.citizen.registered",
		"auth.citizen_login.succeeded",
		"auth.agency_user.created",
		"auth.agency_mfa.setup_started",
		"auth.agency_mfa.verified",
		"auth.agency_login.succeeded",
	} {
		if !hasAuditAction(srv, action) {
			t.Fatalf("expected audit action %q", action)
		}
	}

	createAudit, ok := findAuditAction(srv, "auth.agency_user.created")
	if !ok {
		t.Fatal("expected agency user creation audit record")
	}
	if createAudit.ActorRole != models.RoleSystemAdmin || createAudit.TargetID != createPayload.User.ID {
		t.Fatalf("unexpected actor/target in audit record: %#v", createAudit)
	}
	if createAudit.RequestID != "req-create-dispatcher" || createAudit.IPAddress != "203.0.113.10" || createAudit.UserAgent != "nadaa-test/1.0" {
		t.Fatalf("expected request metadata, got %#v", createAudit)
	}
	if createAudit.After["role"] != models.RoleDispatcher || createAudit.After["mfaEnabled"] != false {
		t.Fatalf("expected sanitized after snapshot, got %#v", createAudit.After)
	}
	if _, containsPassword := createAudit.After["temporaryPassword"]; containsPassword {
		t.Fatal("audit snapshot must not include temporary password")
	}
}

func TestAuditLogEndpointRequiresSystemAdmin(t *testing.T) {
	srv := newTestServer()
	_, systemAdminToken := seedVerifiedAgencyUser(t, srv, models.RoleSystemAdmin, "admin@nadaa.local", "+233200000001")
	_, dispatcherToken := seedVerifiedAgencyUser(t, srv, models.RoleDispatcher, "dispatcher@nadaa.local", "+233200000002")

	forbidden := httptest.NewRecorder()
	forbiddenRequest := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs", nil)
	forbiddenRequest.Header.Set("Authorization", "Bearer "+dispatcherToken)
	srv.listAuditLogsHandler(forbidden, forbiddenRequest)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, forbidden.Code, forbidden.Body.String())
	}
	if !hasAuditAction(srv, "auth.rbac.denied") {
		t.Fatal("expected RBAC denial audit event")
	}

	list := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logs?limit=5", nil)
	listRequest.Header.Set("Authorization", "Bearer "+systemAdminToken)
	srv.listAuditLogsHandler(list, listRequest)
	if list.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, list.Code, list.Body.String())
	}

	var payload models.AuditLogListResponse
	decodeResponse(t, list, &payload)
	if len(payload.Logs) == 0 || len(payload.Logs) > 5 {
		t.Fatalf("expected up to five audit logs, got %#v", payload.Logs)
	}
	if !hasAuditAction(srv, "audit.logs.viewed") {
		t.Fatal("expected audit log view event")
	}
}

func seedVerifiedAgencyUser(t *testing.T, srv *Server, role string, email string, phone string) (models.AgencyUserProfile, string) {
	t.Helper()

	password := "Password123!"
	profile, err := srv.store.CreateAgencyUser(models.CreateAgencyUserRequest{
		Name:     "Seed " + role,
		Email:    email,
		Phone:    phone,
		AgencyID: models.DefaultAgencyID,
		Role:     role,
	}, password, srv.now())
	if err != nil {
		t.Fatalf("seed agency user: %v", err)
	}

	challenge, err := srv.store.StartAgencyMFASetup(profile.ID, email, password, utils.NewMFASecret(), "123456", srv.now())
	if err != nil {
		t.Fatalf("seed MFA setup: %v", err)
	}

	profile, err = srv.store.VerifyAgencyMFA(profile.ID, email, password, challenge.Code, srv.now())
	if err != nil {
		t.Fatalf("seed MFA verify: %v", err)
	}

	token, err := srv.signAgencyToken(profile, srv.now().Add(12*time.Hour))
	if err != nil {
		t.Fatalf("seed agency token: %v", err)
	}

	return profile, token
}

func hasAuditAction(srv *Server, action string) bool {
	_, ok := findAuditAction(srv, action)
	return ok
}

func findAuditAction(srv *Server, action string) (models.AuditLogRecord, bool) {
	for _, record := range srv.store.ListAuditLogs(100) {
		if record.Action == action {
			return record, true
		}
	}
	return models.AuditLogRecord{}, false
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
