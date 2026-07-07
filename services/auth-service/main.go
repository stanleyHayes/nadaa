package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	store        *memoryStore
	tokenSecret  []byte
	otp          otpGenerator
	now          func() time.Time
	exposeDevOTP bool
}

type memoryStore struct {
	mu                  sync.RWMutex
	usersByID           map[string]citizenProfile
	usersByPhone        map[string]string
	agenciesByID        map[string]agencyRecord
	agencyUsersByID     map[string]agencyUser
	agencyUsersByEmail  map[string]string
	agencyUsersByPhone  map[string]string
	challenges          map[string]otpChallenge
	mfaChallengesByUser map[string]mfaChallenge
	auditLogs           []auditLogRecord
}

type otpGenerator interface {
	Generate() (string, error)
}

type randomOTPGenerator struct{}

type fixedOTPGenerator struct {
	code string
}

type otpChallenge struct {
	ID        string
	Phone     string
	Code      string
	ExpiresAt time.Time
}

type mfaChallenge struct {
	ID        string
	UserID    string
	Secret    string
	Code      string
	ExpiresAt time.Time
}

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type agencyRecord struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Region        string    `json:"region"`
	District      string    `json:"district"`
	ContactNumber string    `json:"contactNumber,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type agencySummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Region        string `json:"region"`
	District      string `json:"district"`
	ContactNumber string `json:"contactNumber,omitempty"`
}

type citizenProfile struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	Phone             string       `json:"phone"`
	Role              string       `json:"role"`
	PreferredLanguage string       `json:"preferredLanguage"`
	HomeLocation      *coordinates `json:"homeLocation,omitempty"`
	ContactPermission bool         `json:"contactPermission"`
	CreatedAt         time.Time    `json:"createdAt"`
}

type agencyUser struct {
	ID           string
	Name         string
	Email        string
	Phone        string
	Role         string
	AgencyID     string
	MFARequired  bool
	MFAEnabled   bool
	PasswordHash string
	MFACode      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type agencyUserProfile struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	Phone       string        `json:"phone"`
	Role        string        `json:"role"`
	Agency      agencySummary `json:"agency"`
	MFARequired bool          `json:"mfaRequired"`
	MFAEnabled  bool          `json:"mfaEnabled"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type registerCitizenRequest struct {
	Name              string       `json:"name"`
	Phone             string       `json:"phone"`
	PreferredLanguage string       `json:"preferredLanguage"`
	HomeLocation      *coordinates `json:"homeLocation"`
	ContactPermission bool         `json:"contactPermission"`
}

type registerCitizenResponse struct {
	UserID      string `json:"userId"`
	Phone       string `json:"phone"`
	ChallengeID string `json:"challengeId"`
	OTPDelivery string `json:"otpDelivery"`
	DevOTP      string `json:"devOtp,omitempty"`
}

type loginCitizenRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

type loginCitizenResponse struct {
	AccessToken string         `json:"accessToken"`
	TokenType   string         `json:"tokenType"`
	ExpiresAt   time.Time      `json:"expiresAt"`
	User        citizenProfile `json:"user"`
}

type createAgencyUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	AgencyID string `json:"agencyId"`
	Role     string `json:"role"`
}

type createAgencyUserResponse struct {
	User              agencyUserProfile `json:"user"`
	TemporaryPassword string            `json:"temporaryPassword"`
	MFASetupRequired  bool              `json:"mfaSetupRequired"`
}

type agencyMFASetupRequest struct {
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporaryPassword"`
}

type agencyMFASetupResponse struct {
	UserID      string    `json:"userId"`
	ChallengeID string    `json:"challengeId"`
	Method      string    `json:"method"`
	Secret      string    `json:"secret"`
	ExpiresAt   time.Time `json:"expiresAt"`
	DevCode     string    `json:"devCode,omitempty"`
}

type agencyMFAVerifyRequest struct {
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporaryPassword"`
	Code              string `json:"code"`
}

type agencyMFAVerifyResponse struct {
	User agencyUserProfile `json:"user"`
}

type loginAgencyRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MFACode  string `json:"mfaCode"`
}

type loginAgencyResponse struct {
	AccessToken string            `json:"accessToken"`
	TokenType   string            `json:"tokenType"`
	ExpiresAt   time.Time         `json:"expiresAt"`
	User        agencyUserProfile `json:"user"`
}

type auditLogRecord struct {
	ID            string         `json:"id"`
	ActorUserID   string         `json:"actorUserId,omitempty"`
	ActorAgencyID string         `json:"actorAgencyId,omitempty"`
	ActorRole     string         `json:"actorRole,omitempty"`
	Action        string         `json:"action"`
	TargetType    string         `json:"targetType"`
	TargetID      string         `json:"targetId,omitempty"`
	RequestID     string         `json:"requestId,omitempty"`
	IPAddress     string         `json:"ipAddress,omitempty"`
	UserAgent     string         `json:"userAgent,omitempty"`
	Before        map[string]any `json:"before,omitempty"`
	After         map[string]any `json:"after,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
}

type auditLogListResponse struct {
	Logs []auditLogRecord `json:"logs"`
}

type auditActor struct {
	UserID   string
	AgencyID string
	Role     string
}

type auditTarget struct {
	Type string
	ID   string
}

type auditRequestContext struct {
	RequestID string
	IPAddress string
	UserAgent string
}

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

const (
	defaultAgencyID = "00000000-0000-0000-0000-000000000101"

	roleCitizen         = "citizen"
	roleAgencyViewer    = "agency_viewer"
	roleDispatcher      = "dispatcher"
	roleResponder       = "responder"
	roleNADMOOfficer    = "nadmo_officer"
	roleDistrictOfficer = "district_officer"
	roleAgencyAdmin     = "agency_admin"
	roleSystemAdmin     = "system_admin"
)

var agencyRoles = map[string]bool{
	roleAgencyViewer:    true,
	roleDispatcher:      true,
	roleResponder:       true,
	roleNADMOOfficer:    true,
	roleDistrictOfficer: true,
	roleAgencyAdmin:     true,
	roleSystemAdmin:     true,
}

func main() {
	srv := newServerFromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/register", srv.registerCitizenHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/login", srv.loginCitizenHandler)
	mux.HandleFunc("GET /api/v1/auth/me", srv.meHandler)
	mux.HandleFunc("POST /api/v1/auth/agency-users", srv.createAgencyUserHandler)
	mux.HandleFunc("POST /api/v1/auth/agency-users/{id}/mfa/setup", srv.setupAgencyMFAHandler)
	mux.HandleFunc("POST /api/v1/auth/agency-users/{id}/mfa/verify", srv.verifyAgencyMFAHandler)
	mux.HandleFunc("POST /api/v1/auth/agency/login", srv.loginAgencyHandler)
	mux.HandleFunc("GET /api/v1/audit/logs", srv.listAuditLogsHandler)

	addr := envOrDefault("NADAA_AUTH_ADDR", ":8080")
	log.Printf("auth-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServerFromEnv() *server {
	secret := envOrDefault("NADAA_AUTH_TOKEN_SECRET", "dev-secret-change-me")
	mockOTP := os.Getenv("NADAA_AUTH_MOCK_OTP")
	store := newMemoryStore()

	var otp otpGenerator = randomOTPGenerator{}
	if mockOTP != "" {
		otp = fixedOTPGenerator{code: mockOTP}
	}
	if err := seedBootstrapAgencyAdmin(store, mockOTP); err != nil {
		log.Printf("bootstrap agency admin skipped: %v", err)
	}

	return &server{
		store:        store,
		tokenSecret:  []byte(secret),
		otp:          otp,
		now:          time.Now,
		exposeDevOTP: os.Getenv("NADAA_AUTH_EXPOSE_DEV_OTP") == "true",
	}
}

func newMemoryStore() *memoryStore {
	now := time.Now().UTC()
	store := &memoryStore{
		usersByID:           map[string]citizenProfile{},
		usersByPhone:        map[string]string{},
		agenciesByID:        map[string]agencyRecord{},
		agencyUsersByID:     map[string]agencyUser{},
		agencyUsersByEmail:  map[string]string{},
		agencyUsersByPhone:  map[string]string{},
		challenges:          map[string]otpChallenge{},
		mfaChallengesByUser: map[string]mfaChallenge{},
		auditLogs:           []auditLogRecord{},
	}
	store.agenciesByID[defaultAgencyID] = agencyRecord{
		ID:            defaultAgencyID,
		Name:          "NADMO Accra Metro",
		Type:          "nadmo",
		Region:        "Greater Accra",
		District:      "Accra Metropolitan",
		ContactNumber: "112",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	return store
}

func seedBootstrapAgencyAdmin(store *memoryStore, fallbackMFACode string) error {
	email := normalizeEmail(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL"))
	password := strings.TrimSpace(os.Getenv("NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD"))
	if email == "" && password == "" {
		return nil
	}
	if !validEmail(email) || password == "" {
		return errors.New("NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL and NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD are required together")
	}

	now := time.Now().UTC()
	phone := normalizePhone(envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_PHONE", "+233200000001"))
	name := strings.TrimSpace(envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_NAME", "NADAA System Admin"))
	mfaCode := strings.TrimSpace(envOrDefault("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE", fallbackMFACode))
	if mfaCode == "" {
		mfaCode = "123456"
	}

	profile, err := store.createAgencyUser(createAgencyUserRequest{
		Name:     name,
		Email:    email,
		Phone:    phone,
		AgencyID: defaultAgencyID,
		Role:     roleSystemAdmin,
	}, password, now)
	if err != nil {
		return err
	}

	store.enableAgencyMFA(profile.ID, mfaCode, now)
	return nil
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "auth-service"})
}

func (s *server) registerCitizenHandler(w http.ResponseWriter, r *http.Request) {
	var request registerCitizenRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Phone = normalizePhone(request.Phone)
	request.PreferredLanguage = normalizeLanguage(request.PreferredLanguage)

	if request.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "name is required")
		return
	}

	if !validPhone(request.Phone) {
		writeError(w, http.StatusBadRequest, "invalid_phone", "phone must be in E.164 format, for example +233200000000")
		return
	}

	if request.HomeLocation != nil && !validCoordinates(*request.HomeLocation) {
		writeError(w, http.StatusBadRequest, "invalid_home_location", "homeLocation must contain valid lat and lng values")
		return
	}

	code, err := s.otp.Generate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "otp_generation_failed", "could not create login challenge")
		return
	}

	profile, challenge, err := s.store.registerCitizen(request, code, s.now())
	if errors.Is(err, errDuplicatePhone) {
		writeError(w, http.StatusConflict, "phone_already_registered", "phone is already registered")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "registration_failed", "could not register citizen")
		return
	}

	response := registerCitizenResponse{
		UserID:      profile.ID,
		Phone:       profile.Phone,
		ChallengeID: challenge.ID,
		OTPDelivery: "mock",
	}
	if s.exposeDevOTP {
		response.DevOTP = challenge.Code
	}

	s.recordAudit(r, auditActorFromCitizen(profile), "auth.citizen.registered", auditTarget{Type: "citizen_user", ID: profile.ID}, nil, citizenAuditSnapshot(profile))
	writeJSON(w, http.StatusCreated, response)
}

func (s *server) loginCitizenHandler(w http.ResponseWriter, r *http.Request) {
	var request loginCitizenRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Phone = normalizePhone(request.Phone)
	request.OTP = strings.TrimSpace(request.OTP)

	if !validPhone(request.Phone) || request.OTP == "" {
		writeError(w, http.StatusBadRequest, "invalid_login_request", "phone and otp are required")
		return
	}

	profile, err := s.store.verifyOTP(request.Phone, request.OTP, s.now())
	if errors.Is(err, errInvalidCredentials) {
		s.recordAudit(r, auditActor{}, "auth.citizen_login.failed", auditTarget{Type: "citizen_phone", ID: request.Phone}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "phone or otp is invalid")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "login_failed", "could not complete login")
		return
	}

	expiresAt := s.now().Add(24 * time.Hour)
	token, err := s.signToken(profile, expiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_generation_failed", "could not create access token")
		return
	}

	s.recordAudit(r, auditActorFromCitizen(profile), "auth.citizen_login.succeeded", auditTarget{Type: "citizen_user", ID: profile.ID}, nil, map[string]any{
		"expiresAt": expiresAt,
	})
	writeJSON(w, http.StatusOK, loginCitizenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        profile,
	})
}

func (s *server) meHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return
	}

	claims, err := s.verifyToken(token)
	if errors.Is(err, errInvalidToken) {
		writeError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_verification_failed", "could not verify token")
		return
	}

	if claims.UserType == "agency" {
		profile, ok := s.store.agencyProfileByID(claims.UserID)
		if !ok {
			writeError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
			return
		}

		writeJSON(w, http.StatusOK, profile)
		return
	}

	profile, ok := s.store.profileByID(claims.UserID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (s *server) createAgencyUserHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAgencyRole(w, r, roleSystemAdmin, roleAgencyAdmin)
	if !ok {
		return
	}

	var request createAgencyUserRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = normalizeEmail(request.Email)
	request.Phone = normalizePhone(request.Phone)
	request.AgencyID = strings.TrimSpace(request.AgencyID)
	request.Role = normalizeRole(request.Role)

	if request.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "name is required")
		return
	}
	if !validEmail(request.Email) {
		writeError(w, http.StatusBadRequest, "invalid_email", "email must be valid")
		return
	}
	if !validPhone(request.Phone) {
		writeError(w, http.StatusBadRequest, "invalid_phone", "phone must be in E.164 format, for example +233200000000")
		return
	}
	if request.AgencyID == "" {
		writeError(w, http.StatusBadRequest, "agency_required", "agencyId is required")
		return
	}
	if !validAgencyRole(request.Role) {
		writeError(w, http.StatusBadRequest, "invalid_role", "role must be an authority role")
		return
	}
	if actor.Role == roleAgencyAdmin && actor.Agency.ID != request.AgencyID {
		s.recordAudit(r, auditActorFromAgency(actor), "auth.rbac.denied", auditTarget{Type: "agency_user", ID: request.Email}, nil, map[string]any{
			"reason":            "agency_scope_forbidden",
			"requestedRole":     request.Role,
			"requestedAgencyId": request.AgencyID,
		})
		writeError(w, http.StatusForbidden, "agency_scope_forbidden", "agency admins can create users only inside their agency")
		return
	}

	temporaryPassword := newTemporaryPassword()
	profile, err := s.store.createAgencyUser(request, temporaryPassword, s.now())
	if errors.Is(err, errDuplicateEmail) {
		writeError(w, http.StatusConflict, "email_already_registered", "email is already registered")
		return
	}
	if errors.Is(err, errDuplicatePhone) {
		writeError(w, http.StatusConflict, "phone_already_registered", "phone is already registered")
		return
	}
	if errors.Is(err, errUnknownAgency) {
		writeError(w, http.StatusBadRequest, "agency_not_found", "agencyId does not exist")
		return
	}
	if errors.Is(err, errInvalidRole) {
		writeError(w, http.StatusBadRequest, "invalid_role", "role must be an authority role")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "agency_user_creation_failed", "could not create agency user")
		return
	}

	s.recordAudit(r, auditActorFromAgency(actor), "auth.agency_user.created", auditTarget{Type: "agency_user", ID: profile.ID}, nil, agencyUserAuditSnapshot(profile))
	writeJSON(w, http.StatusCreated, createAgencyUserResponse{
		User:              profile,
		TemporaryPassword: temporaryPassword,
		MFASetupRequired:  true,
	})
}

func (s *server) setupAgencyMFAHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("id"))
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id_required", "agency user id is required")
		return
	}

	var request agencyMFASetupRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Email = normalizeEmail(request.Email)
	request.TemporaryPassword = strings.TrimSpace(request.TemporaryPassword)
	if !validEmail(request.Email) || request.TemporaryPassword == "" {
		writeError(w, http.StatusBadRequest, "invalid_mfa_setup_request", "email and temporaryPassword are required")
		return
	}

	code, err := s.otp.Generate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mfa_generation_failed", "could not create MFA challenge")
		return
	}

	challenge, err := s.store.startAgencyMFASetup(userID, request.Email, request.TemporaryPassword, newMFASecret(), code, s.now())
	if errors.Is(err, errInvalidCredentials) {
		s.recordAudit(r, auditActor{}, "auth.agency_mfa.setup_failed", auditTarget{Type: "agency_user", ID: userID}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "agency user or temporary password is invalid")
		return
	}
	if errors.Is(err, errMFAAlreadyEnabled) {
		writeError(w, http.StatusConflict, "mfa_already_enabled", "MFA is already enabled for this agency user")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mfa_setup_failed", "could not start MFA setup")
		return
	}

	response := agencyMFASetupResponse{
		UserID:      userID,
		ChallengeID: challenge.ID,
		Method:      "mock_totp",
		Secret:      challenge.Secret,
		ExpiresAt:   challenge.ExpiresAt,
	}
	if s.exposeDevOTP {
		response.DevCode = challenge.Code
	}

	profile, _ := s.store.agencyProfileByID(userID)
	s.recordAudit(r, auditActorFromAgency(profile), "auth.agency_mfa.setup_started", auditTarget{Type: "agency_user", ID: userID}, nil, map[string]any{
		"challengeId": challenge.ID,
		"method":      response.Method,
		"expiresAt":   challenge.ExpiresAt,
	})
	writeJSON(w, http.StatusOK, response)
}

func (s *server) verifyAgencyMFAHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("id"))
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id_required", "agency user id is required")
		return
	}

	var request agencyMFAVerifyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Email = normalizeEmail(request.Email)
	request.TemporaryPassword = strings.TrimSpace(request.TemporaryPassword)
	request.Code = strings.TrimSpace(request.Code)
	if !validEmail(request.Email) || request.TemporaryPassword == "" || request.Code == "" {
		writeError(w, http.StatusBadRequest, "invalid_mfa_verify_request", "email, temporaryPassword, and code are required")
		return
	}

	profile, err := s.store.verifyAgencyMFA(userID, request.Email, request.TemporaryPassword, request.Code, s.now())
	if errors.Is(err, errInvalidCredentials) {
		s.recordAudit(r, auditActor{}, "auth.agency_mfa.verify_failed", auditTarget{Type: "agency_user", ID: userID}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "agency user, temporary password, or MFA code is invalid")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "mfa_verification_failed", "could not verify MFA")
		return
	}

	s.recordAudit(r, auditActorFromAgency(profile), "auth.agency_mfa.verified", auditTarget{Type: "agency_user", ID: profile.ID}, nil, agencyUserAuditSnapshot(profile))
	writeJSON(w, http.StatusOK, agencyMFAVerifyResponse{User: profile})
}

func (s *server) loginAgencyHandler(w http.ResponseWriter, r *http.Request) {
	var request loginAgencyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Email = normalizeEmail(request.Email)
	request.Password = strings.TrimSpace(request.Password)
	request.MFACode = strings.TrimSpace(request.MFACode)
	if !validEmail(request.Email) || request.Password == "" {
		writeError(w, http.StatusBadRequest, "invalid_login_request", "email and password are required")
		return
	}

	profile, err := s.store.loginAgencyUser(request.Email, request.Password, request.MFACode)
	if errors.Is(err, errMFASetupRequired) {
		s.recordAudit(r, auditActor{}, "auth.agency_login.blocked", auditTarget{Type: "agency_email", ID: request.Email}, nil, map[string]any{
			"reason": "mfa_setup_required",
		})
		writeError(w, http.StatusForbidden, "mfa_setup_required", "MFA must be set up before login")
		return
	}
	if errors.Is(err, errMFARequired) {
		s.recordAudit(r, auditActor{}, "auth.agency_login.failed", auditTarget{Type: "agency_email", ID: request.Email}, nil, map[string]any{
			"reason": "mfa_required",
		})
		writeError(w, http.StatusUnauthorized, "mfa_required", "MFA code is required")
		return
	}
	if errors.Is(err, errInvalidCredentials) {
		s.recordAudit(r, auditActor{}, "auth.agency_login.failed", auditTarget{Type: "agency_email", ID: request.Email}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email, password, or MFA code is invalid")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "login_failed", "could not complete agency login")
		return
	}

	expiresAt := s.now().Add(12 * time.Hour)
	token, err := s.signAgencyToken(profile, expiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_generation_failed", "could not create access token")
		return
	}

	s.recordAudit(r, auditActorFromAgency(profile), "auth.agency_login.succeeded", auditTarget{Type: "agency_user", ID: profile.ID}, nil, map[string]any{
		"expiresAt": expiresAt,
	})
	writeJSON(w, http.StatusOK, loginAgencyResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        profile,
	})
}

func (s *server) listAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAgencyRole(w, r, roleSystemAdmin)
	if !ok {
		return
	}

	limit := parseAuditLimit(r.URL.Query().Get("limit"))
	logs := s.store.listAuditLogs(limit)
	writeJSON(w, http.StatusOK, auditLogListResponse{Logs: logs})

	s.recordAudit(r, auditActorFromAgency(actor), "audit.logs.viewed", auditTarget{Type: "audit_logs"}, nil, map[string]any{
		"limit": limit,
		"count": len(logs),
	})
}

var (
	errDuplicatePhone     = errors.New("duplicate phone")
	errDuplicateEmail     = errors.New("duplicate email")
	errInvalidCredentials = errors.New("invalid credentials")
	errInvalidToken       = errors.New("invalid token")
	errInvalidRole        = errors.New("invalid role")
	errMFAAlreadyEnabled  = errors.New("mfa already enabled")
	errMFARequired        = errors.New("mfa required")
	errMFASetupRequired   = errors.New("mfa setup required")
	errUnknownAgency      = errors.New("unknown agency")
)

func (m *memoryStore) registerCitizen(request registerCitizenRequest, code string, now time.Time) (citizenProfile, otpChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.usersByPhone[request.Phone]; exists {
		return citizenProfile{}, otpChallenge{}, errDuplicatePhone
	}
	if _, exists := m.agencyUsersByPhone[request.Phone]; exists {
		return citizenProfile{}, otpChallenge{}, errDuplicatePhone
	}

	profile := citizenProfile{
		ID:                newID("usr"),
		Name:              request.Name,
		Phone:             request.Phone,
		Role:              "citizen",
		PreferredLanguage: request.PreferredLanguage,
		HomeLocation:      request.HomeLocation,
		ContactPermission: request.ContactPermission,
		CreatedAt:         now.UTC(),
	}
	challenge := otpChallenge{
		ID:        newID("otp"),
		Phone:     request.Phone,
		Code:      code,
		ExpiresAt: now.Add(10 * time.Minute).UTC(),
	}

	m.usersByID[profile.ID] = profile
	m.usersByPhone[profile.Phone] = profile.ID
	m.challenges[profile.Phone] = challenge

	return profile, challenge, nil
}

func (m *memoryStore) createAgencyUser(request createAgencyUserRequest, temporaryPassword string, now time.Time) (agencyUserProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !validAgencyRole(request.Role) {
		return agencyUserProfile{}, errInvalidRole
	}

	agency, exists := m.agenciesByID[request.AgencyID]
	if !exists {
		return agencyUserProfile{}, errUnknownAgency
	}

	if _, exists := m.agencyUsersByEmail[request.Email]; exists {
		return agencyUserProfile{}, errDuplicateEmail
	}
	if _, exists := m.usersByPhone[request.Phone]; exists {
		return agencyUserProfile{}, errDuplicatePhone
	}
	if _, exists := m.agencyUsersByPhone[request.Phone]; exists {
		return agencyUserProfile{}, errDuplicatePhone
	}

	user := agencyUser{
		ID:           newID("usr"),
		Name:         request.Name,
		Email:        request.Email,
		Phone:        request.Phone,
		Role:         request.Role,
		AgencyID:     request.AgencyID,
		MFARequired:  true,
		MFAEnabled:   false,
		PasswordHash: hashCredential(temporaryPassword),
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC(),
	}

	m.agencyUsersByID[user.ID] = user
	m.agencyUsersByEmail[user.Email] = user.ID
	m.agencyUsersByPhone[user.Phone] = user.ID

	return agencyProfileFromUser(user, agency), nil
}

func (m *memoryStore) startAgencyMFASetup(userID string, email string, temporaryPassword string, secret string, code string, now time.Time) (mfaChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, _, err := m.authorityUserByCredentials(userID, email, temporaryPassword)
	if err != nil {
		return mfaChallenge{}, err
	}
	if user.MFAEnabled {
		return mfaChallenge{}, errMFAAlreadyEnabled
	}

	challenge := mfaChallenge{
		ID:        newID("mfa"),
		UserID:    user.ID,
		Secret:    secret,
		Code:      code,
		ExpiresAt: now.Add(10 * time.Minute).UTC(),
	}
	m.mfaChallengesByUser[user.ID] = challenge

	return challenge, nil
}

func (m *memoryStore) verifyAgencyMFA(userID string, email string, temporaryPassword string, code string, now time.Time) (agencyUserProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, agency, err := m.authorityUserByCredentials(userID, email, temporaryPassword)
	if err != nil {
		return agencyUserProfile{}, err
	}

	challenge, exists := m.mfaChallengesByUser[user.ID]
	if !exists || challenge.Code != code || now.After(challenge.ExpiresAt) {
		return agencyUserProfile{}, errInvalidCredentials
	}

	delete(m.mfaChallengesByUser, user.ID)
	user.MFAEnabled = true
	user.MFACode = code
	user.UpdatedAt = now.UTC()
	m.agencyUsersByID[user.ID] = user

	return agencyProfileFromUser(user, agency), nil
}

func (m *memoryStore) enableAgencyMFA(userID string, code string, now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.agencyUsersByID[userID]
	if !exists {
		return
	}
	user.MFAEnabled = true
	user.MFACode = code
	user.UpdatedAt = now.UTC()
	m.agencyUsersByID[user.ID] = user
}

func (m *memoryStore) loginAgencyUser(email string, password string, mfaCode string) (agencyUserProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID, exists := m.agencyUsersByEmail[email]
	if !exists {
		return agencyUserProfile{}, errInvalidCredentials
	}

	user := m.agencyUsersByID[userID]
	if user.PasswordHash != hashCredential(password) {
		return agencyUserProfile{}, errInvalidCredentials
	}
	if user.MFARequired && !user.MFAEnabled {
		return agencyUserProfile{}, errMFASetupRequired
	}
	if user.MFARequired && mfaCode == "" {
		return agencyUserProfile{}, errMFARequired
	}
	if user.MFARequired && user.MFACode != mfaCode {
		return agencyUserProfile{}, errInvalidCredentials
	}

	agency, exists := m.agenciesByID[user.AgencyID]
	if !exists {
		return agencyUserProfile{}, errUnknownAgency
	}

	return agencyProfileFromUser(user, agency), nil
}

func (m *memoryStore) authorityUserByCredentials(userID string, email string, password string) (agencyUser, agencyRecord, error) {
	user, exists := m.agencyUsersByID[userID]
	if !exists || user.Email != email || user.PasswordHash != hashCredential(password) {
		return agencyUser{}, agencyRecord{}, errInvalidCredentials
	}

	agency, exists := m.agenciesByID[user.AgencyID]
	if !exists {
		return agencyUser{}, agencyRecord{}, errUnknownAgency
	}

	return user, agency, nil
}

func (m *memoryStore) verifyOTP(phone string, code string, now time.Time) (citizenProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userID, exists := m.usersByPhone[phone]
	if !exists {
		return citizenProfile{}, errInvalidCredentials
	}

	challenge, exists := m.challenges[phone]
	if !exists || challenge.Code != code || now.After(challenge.ExpiresAt) {
		return citizenProfile{}, errInvalidCredentials
	}

	delete(m.challenges, phone)
	return m.usersByID[userID], nil
}

func (m *memoryStore) profileByID(id string) (citizenProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile, ok := m.usersByID[id]
	return profile, ok
}

func (m *memoryStore) agencyProfileByID(id string) (agencyUserProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.agencyUsersByID[id]
	if !ok {
		return agencyUserProfile{}, false
	}

	agency, ok := m.agenciesByID[user.AgencyID]
	if !ok {
		return agencyUserProfile{}, false
	}

	return agencyProfileFromUser(user, agency), true
}

func (m *memoryStore) appendAuditLog(record auditLogRecord) auditLogRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.auditLogs = append(m.auditLogs, record)
	return record
}

func (m *memoryStore) listAuditLogs(limit int) []auditLogRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	logs := make([]auditLogRecord, 0, min(limit, len(m.auditLogs)))
	for i := len(m.auditLogs) - 1; i >= 0 && len(logs) < limit; i-- {
		logs = append(logs, m.auditLogs[i])
	}

	return logs
}

type tokenClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	MFA       bool   `json:"mfa,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

func (s *server) signToken(profile citizenProfile, expiresAt time.Time) (string, error) {
	claims := tokenClaims{
		UserID:    profile.ID,
		UserType:  roleCitizen,
		Phone:     profile.Phone,
		Role:      profile.Role,
		ExpiresAt: expiresAt.Unix(),
	}
	return s.signClaims(claims)
}

func (s *server) signAgencyToken(profile agencyUserProfile, expiresAt time.Time) (string, error) {
	claims := tokenClaims{
		UserID:    profile.ID,
		UserType:  "agency",
		Email:     profile.Email,
		Phone:     profile.Phone,
		Role:      profile.Role,
		AgencyID:  profile.Agency.ID,
		MFA:       profile.MFAEnabled,
		ExpiresAt: expiresAt.Unix(),
	}
	return s.signClaims(claims)
}

func (s *server) signClaims(claims tokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := s.sign(encodedPayload)
	return "nadaa." + encodedPayload + "." + signature, nil
}

func (s *server) verifyToken(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return tokenClaims{}, errInvalidToken
	}

	expectedSignature := s.sign(parts[1])
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return tokenClaims{}, errInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, errInvalidToken
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, errInvalidToken
	}

	if claims.ExpiresAt <= s.now().Unix() {
		return tokenClaims{}, errInvalidToken
	}

	return claims, nil
}

func (s *server) requireAgencyRole(w http.ResponseWriter, r *http.Request, allowedRoles ...string) (agencyUserProfile, bool) {
	token, ok := bearerToken(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return agencyUserProfile{}, false
	}

	claims, err := s.verifyToken(token)
	if errors.Is(err, errInvalidToken) {
		writeError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return agencyUserProfile{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_verification_failed", "could not verify token")
		return agencyUserProfile{}, false
	}
	if claims.UserType != "agency" {
		s.recordAudit(r, auditActor{UserID: claims.UserID, Role: claims.Role}, "auth.rbac.denied", auditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"reason":     "authority_user_required",
			"actualRole": claims.Role,
		})
		writeError(w, http.StatusForbidden, "authority_user_required", "authority user access is required")
		return agencyUserProfile{}, false
	}
	if !claims.MFA {
		s.recordAudit(r, auditActor{UserID: claims.UserID, AgencyID: claims.AgencyID, Role: claims.Role}, "auth.rbac.denied", auditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"reason":     "mfa_required",
			"actualRole": claims.Role,
		})
		writeError(w, http.StatusForbidden, "mfa_required", "MFA is required for authority workflows")
		return agencyUserProfile{}, false
	}

	profile, ok := s.store.agencyProfileByID(claims.UserID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
		return agencyUserProfile{}, false
	}
	if !profile.MFAEnabled {
		s.recordAudit(r, auditActorFromAgency(profile), "auth.rbac.denied", auditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"reason":     "mfa_required",
			"actualRole": profile.Role,
		})
		writeError(w, http.StatusForbidden, "mfa_required", "MFA is required for authority workflows")
		return agencyUserProfile{}, false
	}
	if !roleIn(profile.Role, allowedRoles) {
		s.recordAudit(r, auditActorFromAgency(profile), "auth.rbac.denied", auditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"allowedRoles": allowedRoles,
			"actualRole":   profile.Role,
		})
		writeError(w, http.StatusForbidden, "forbidden", "role is not allowed to perform this action")
		return agencyUserProfile{}, false
	}

	return profile, true
}

func (s *server) recordAudit(r *http.Request, actor auditActor, action string, target auditTarget, before map[string]any, after map[string]any) auditLogRecord {
	context := auditContextFromRequest(r)
	record := auditLogRecord{
		ID:            newID("aud"),
		ActorUserID:   actor.UserID,
		ActorAgencyID: actor.AgencyID,
		ActorRole:     actor.Role,
		Action:        action,
		TargetType:    target.Type,
		TargetID:      target.ID,
		RequestID:     context.RequestID,
		IPAddress:     context.IPAddress,
		UserAgent:     context.UserAgent,
		Before:        before,
		After:         after,
		CreatedAt:     s.now().UTC(),
	}
	return s.store.appendAuditLog(record)
}

func (s *server) sign(payload string) string {
	mac := hmac.New(sha256.New, s.tokenSecret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (randomOTPGenerator) Generate() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (f fixedOTPGenerator) Generate() (string, error) {
	if f.code == "" {
		return "123456", nil
	}
	return f.code, nil
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func withCORS(next http.Handler) http.Handler {
	allowedOrigins := allowedOriginsFromEnv()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Cache-Control", "no-store")
}

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for _, origin := range strings.Split(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return phone
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return "en"
	}
	return language
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func validPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

func validEmail(email string) bool {
	parts := strings.Split(email, "@")
	return len(parts) == 2 && parts[0] != "" && strings.Contains(parts[1], ".")
}

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

func parseAuditLimit(raw string) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

func validAgencyRole(role string) bool {
	return agencyRoles[role]
}

func roleIn(role string, allowedRoles []string) bool {
	for _, allowed := range allowedRoles {
		if role == allowed {
			return true
		}
	}
	return false
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return "", false
	}
	return token, true
}

func auditContextFromRequest(r *http.Request) auditRequestContext {
	if r == nil {
		return auditRequestContext{RequestID: newID("req")}
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = newID("req")
	}

	return auditRequestContext{
		RequestID: requestID,
		IPAddress: requestIPAddress(r),
		UserAgent: strings.TrimSpace(r.UserAgent()),
	}
}

func requestIPAddress(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}
	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func auditActorFromCitizen(profile citizenProfile) auditActor {
	return auditActor{UserID: profile.ID, Role: profile.Role}
}

func auditActorFromAgency(profile agencyUserProfile) auditActor {
	return auditActor{UserID: profile.ID, AgencyID: profile.Agency.ID, Role: profile.Role}
}

func agencyUserAuditSnapshot(profile agencyUserProfile) map[string]any {
	return map[string]any{
		"id":          profile.ID,
		"name":        profile.Name,
		"email":       profile.Email,
		"phone":       profile.Phone,
		"role":        profile.Role,
		"agencyId":    profile.Agency.ID,
		"mfaRequired": profile.MFARequired,
		"mfaEnabled":  profile.MFAEnabled,
	}
}

func citizenAuditSnapshot(profile citizenProfile) map[string]any {
	return map[string]any{
		"id":                profile.ID,
		"phone":             profile.Phone,
		"role":              profile.Role,
		"preferredLanguage": profile.PreferredLanguage,
		"contactPermission": profile.ContactPermission,
	}
}

func agencyProfileFromUser(user agencyUser, agency agencyRecord) agencyUserProfile {
	return agencyUserProfile{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		Role:        user.Role,
		Agency:      agencySummaryFromRecord(agency),
		MFARequired: user.MFARequired,
		MFAEnabled:  user.MFAEnabled,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func agencySummaryFromRecord(agency agencyRecord) agencySummary {
	return agencySummary{
		ID:            agency.ID,
		Name:          agency.Name,
		Type:          agency.Type,
		Region:        agency.Region,
		District:      agency.District,
		ContactNumber: agency.ContactNumber,
	}
}

func hashCredential(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newTemporaryPassword() string {
	return newID("tmp")
}

func newMFASecret() string {
	return newID("mfa_secret")
}

func newID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, bytes)
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
