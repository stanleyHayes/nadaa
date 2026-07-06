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
	"net/http"
	"os"
	"regexp"
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
	mu           sync.RWMutex
	usersByID    map[string]citizenProfile
	usersByPhone map[string]string
	challenges   map[string]otpChallenge
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

type coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
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

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func main() {
	srv := newServerFromEnv()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", srv.healthHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/register", srv.registerCitizenHandler)
	mux.HandleFunc("POST /api/v1/auth/citizens/login", srv.loginCitizenHandler)
	mux.HandleFunc("GET /api/v1/auth/me", srv.meHandler)

	addr := envOrDefault("NADAA_AUTH_ADDR", ":8080")
	log.Printf("auth-service listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func newServerFromEnv() *server {
	secret := envOrDefault("NADAA_AUTH_TOKEN_SECRET", "dev-secret-change-me")
	mockOTP := os.Getenv("NADAA_AUTH_MOCK_OTP")

	var otp otpGenerator = randomOTPGenerator{}
	if mockOTP != "" {
		otp = fixedOTPGenerator{code: mockOTP}
	}

	return &server{
		store:        newMemoryStore(),
		tokenSecret:  []byte(secret),
		otp:          otp,
		now:          time.Now,
		exposeDevOTP: os.Getenv("NADAA_AUTH_EXPOSE_DEV_OTP") == "true",
	}
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		usersByID:    map[string]citizenProfile{},
		usersByPhone: map[string]string{},
		challenges:   map[string]otpChallenge{},
	}
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

	writeJSON(w, http.StatusOK, loginCitizenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        profile,
	})
}

func (s *server) meHandler(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" || token == r.Header.Get("Authorization") {
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

	profile, ok := s.store.profileByID(claims.UserID)
	if !ok {
		writeError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

var (
	errDuplicatePhone     = errors.New("duplicate phone")
	errInvalidCredentials = errors.New("invalid credentials")
	errInvalidToken       = errors.New("invalid token")
)

func (m *memoryStore) registerCitizen(request registerCitizenRequest, code string, now time.Time) (citizenProfile, otpChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.usersByPhone[request.Phone]; exists {
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

type tokenClaims struct {
	UserID    string `json:"sub"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

func (s *server) signToken(profile citizenProfile, expiresAt time.Time) (string, error) {
	claims := tokenClaims{
		UserID:    profile.ID,
		Phone:     profile.Phone,
		Role:      profile.Role,
		ExpiresAt: expiresAt.Unix(),
	}
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return phone
}

func normalizeLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return "en"
	}
	return language
}

func validPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

func validCoordinates(location coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
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
