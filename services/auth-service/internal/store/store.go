package store

import (
	"errors"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

// Sentinel errors returned by the auth store.
var (
	ErrDuplicatePhone     = errors.New("duplicate phone")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrUnknownPhone       = errors.New("unknown phone")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidRole        = errors.New("invalid role")
	ErrMFAAlreadyEnabled  = errors.New("mfa already enabled")
	ErrMFARequired        = errors.New("mfa required")
	ErrMFASetupRequired   = errors.New("mfa setup required")
	ErrTooManyAttempts    = errors.New("too many attempts")
	ErrUnknownAgency      = errors.New("unknown agency")
)

// Brute-force protection: after maxFailedAttempts consecutive failures on a
// credential path, the account key locks for attemptLockout.
const (
	maxFailedAttempts = 5
	attemptLockout    = 15 * time.Minute
)

// failedAttemptLog tracks consecutive credential failures for one account key.
type failedAttemptLog struct {
	count       int
	lockedUntil time.Time
}

// Store is the persistence interface for auth data.
type Store interface {
	RegisterCitizen(request models.RegisterCitizenRequest, code string, now time.Time) (models.CitizenProfile, models.OTPChallenge, error)
	RequestCitizenOTP(phone string, code string, now time.Time) (models.CitizenProfile, models.OTPChallenge, error)
	VerifyOTP(phone string, code string, now time.Time) (models.CitizenProfile, error)
	ProfileByID(id string) (models.CitizenProfile, bool)
	CreateAgencyUser(request models.CreateAgencyUserRequest, temporaryPassword string, now time.Time) (models.AgencyUserProfile, error)
	StartAgencyMFASetup(userID string, email string, temporaryPassword string, secret string, code string, now time.Time) (models.MFAChallenge, error)
	VerifyAgencyMFA(userID string, email string, temporaryPassword string, code string, now time.Time) (models.AgencyUserProfile, error)
	LoginAgencyUser(email string, password string, mfaCode string, now time.Time) (models.AgencyUserProfile, error)
	AgencyProfileByID(id string) (models.AgencyUserProfile, bool)
	ListAgencies() []models.AgencySummary
	AppendAuditLog(record models.AuditLogRecord) models.AuditLogRecord
	ListAuditLogs(limit int) []models.AuditLogRecord
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu                  sync.RWMutex
	usersByID           map[string]models.CitizenProfile
	usersByPhone        map[string]string
	agenciesByID        map[string]models.AgencyRecord
	agencyUsersByID     map[string]models.AgencyUser
	agencyUsersByEmail  map[string]string
	agencyUsersByPhone  map[string]string
	challenges          map[string]models.OTPChallenge
	mfaChallengesByUser map[string]models.MFAChallenge
	failedAttempts      map[string]*failedAttemptLog
	auditLogs           []models.AuditLogRecord
}

// NewMemoryStore creates an in-memory store seeded with fixture data. When
// bootstrap admin credentials are configured but seeding fails, the process
// exits: running without the expected admin (or with partially applied
// credentials) leaves operators locked out of a freshly deployed environment.
func NewMemoryStore(now time.Time, cfg *config.Config) Store {
	m := &MemoryStore{
		usersByID:           map[string]models.CitizenProfile{},
		usersByPhone:        map[string]string{},
		agenciesByID:        map[string]models.AgencyRecord{},
		agencyUsersByID:     map[string]models.AgencyUser{},
		agencyUsersByEmail:  map[string]string{},
		agencyUsersByPhone:  map[string]string{},
		challenges:          map[string]models.OTPChallenge{},
		mfaChallengesByUser: map[string]models.MFAChallenge{},
		failedAttempts:      map[string]*failedAttemptLog{},
		auditLogs:           []models.AuditLogRecord{},
	}
	m.agenciesByID[models.DefaultAgencyID] = models.AgencyRecord{
		ID:            models.DefaultAgencyID,
		Name:          "NADMO Accra Metro",
		Type:          "nadmo",
		Region:        "Greater Accra",
		District:      "Accra Metropolitan",
		ContactNumber: "112",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := seedBootstrapAgencyAdmin(m, cfg, now); err != nil {
		log.Fatalf("failed to seed bootstrap agency admin: %v", err)
	}
	return m
}

// RegisterCitizen creates a new citizen profile and login challenge.
func (m *MemoryStore) RegisterCitizen(request models.RegisterCitizenRequest, code string, now time.Time) (models.CitizenProfile, models.OTPChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.usersByPhone[request.Phone]; exists {
		return models.CitizenProfile{}, models.OTPChallenge{}, ErrDuplicatePhone
	}
	if _, exists := m.agencyUsersByPhone[request.Phone]; exists {
		return models.CitizenProfile{}, models.OTPChallenge{}, ErrDuplicatePhone
	}

	profile := models.CitizenProfile{
		ID:                utils.NewID("usr"),
		Name:              request.Name,
		Phone:             request.Phone,
		Role:              models.RoleCitizen,
		PreferredLanguage: request.PreferredLanguage,
		HomeLocation:      request.HomeLocation,
		ContactPermission: request.ContactPermission,
		CreatedAt:         now.UTC(),
	}
	challenge := models.OTPChallenge{
		ID:        utils.NewID("otp"),
		Phone:     request.Phone,
		Code:      code,
		ExpiresAt: now.Add(10 * time.Minute).UTC(),
	}

	m.usersByID[profile.ID] = profile
	m.usersByPhone[profile.Phone] = profile.ID
	m.challenges[profile.Phone] = challenge

	return profile, challenge, nil
}

// RequestCitizenOTP issues a fresh login challenge for an already-registered
// citizen phone, replacing any outstanding challenge for that phone.
func (m *MemoryStore) RequestCitizenOTP(phone string, code string, now time.Time) (models.CitizenProfile, models.OTPChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	userID, exists := m.usersByPhone[phone]
	if !exists {
		return models.CitizenProfile{}, models.OTPChallenge{}, ErrUnknownPhone
	}

	challenge := models.OTPChallenge{
		ID:        utils.NewID("otp"),
		Phone:     phone,
		Code:      code,
		ExpiresAt: now.Add(10 * time.Minute).UTC(),
	}
	m.challenges[phone] = challenge

	return m.usersByID[userID], challenge, nil
}

// CreateAgencyUser creates a new authority user with a temporary password.
func (m *MemoryStore) CreateAgencyUser(request models.CreateAgencyUserRequest, temporaryPassword string, now time.Time) (models.AgencyUserProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !utils.ValidAgencyRole(request.Role) {
		return models.AgencyUserProfile{}, ErrInvalidRole
	}

	agency, exists := m.agenciesByID[request.AgencyID]
	if !exists {
		return models.AgencyUserProfile{}, ErrUnknownAgency
	}

	if _, exists := m.agencyUsersByEmail[request.Email]; exists {
		return models.AgencyUserProfile{}, ErrDuplicateEmail
	}
	if _, exists := m.usersByPhone[request.Phone]; exists {
		return models.AgencyUserProfile{}, ErrDuplicatePhone
	}
	if _, exists := m.agencyUsersByPhone[request.Phone]; exists {
		return models.AgencyUserProfile{}, ErrDuplicatePhone
	}

	user := models.AgencyUser{
		ID:           utils.NewID("usr"),
		Name:         request.Name,
		Email:        request.Email,
		Phone:        request.Phone,
		Role:         request.Role,
		AgencyID:     request.AgencyID,
		MFARequired:  true,
		MFAEnabled:   false,
		PasswordHash: utils.HashCredential(temporaryPassword),
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC(),
	}

	m.agencyUsersByID[user.ID] = user
	m.agencyUsersByEmail[user.Email] = user.ID
	m.agencyUsersByPhone[user.Phone] = user.ID

	return utils.AgencyProfileFromUser(user, agency), nil
}

// StartAgencyMFASetup begins MFA enrollment for an agency user.
func (m *MemoryStore) StartAgencyMFASetup(userID string, email string, temporaryPassword string, secret string, code string, now time.Time) (models.MFAChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, _, err := m.authorityUserByCredentials(userID, email, temporaryPassword)
	if err != nil {
		return models.MFAChallenge{}, err
	}
	if user.MFAEnabled {
		return models.MFAChallenge{}, ErrMFAAlreadyEnabled
	}

	challenge := models.MFAChallenge{
		ID:        utils.NewID("mfa"),
		UserID:    user.ID,
		Secret:    secret,
		Code:      code,
		ExpiresAt: now.Add(10 * time.Minute).UTC(),
	}
	m.mfaChallengesByUser[user.ID] = challenge

	return challenge, nil
}

// VerifyAgencyMFA confirms an MFA setup challenge and enables MFA.
func (m *MemoryStore) VerifyAgencyMFA(userID string, email string, temporaryPassword string, code string, now time.Time) (models.AgencyUserProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	attemptKey := "agency-mfa:" + email
	if m.lockedOut(attemptKey, now) {
		return models.AgencyUserProfile{}, ErrTooManyAttempts
	}

	user, agency, err := m.authorityUserByCredentials(userID, email, temporaryPassword)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			m.recordFailure(attemptKey, now)
		}
		return models.AgencyUserProfile{}, err
	}

	challenge, exists := m.mfaChallengesByUser[user.ID]
	if !exists || !utils.SecureCompare(challenge.Code, code) || now.After(challenge.ExpiresAt) {
		m.recordFailure(attemptKey, now)
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}

	delete(m.mfaChallengesByUser, user.ID)
	m.resetFailures(attemptKey)
	user.MFAEnabled = true
	user.MFACode = code
	user.UpdatedAt = now.UTC()
	m.agencyUsersByID[user.ID] = user

	return utils.AgencyProfileFromUser(user, agency), nil
}

func (m *MemoryStore) enableAgencyMFA(userID string, code string, now time.Time) {
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

// LoginAgencyUser authenticates an authority user.
func (m *MemoryStore) LoginAgencyUser(email string, password string, mfaCode string, now time.Time) (models.AgencyUserProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	attemptKey := "agency-login:" + email
	if m.lockedOut(attemptKey, now) {
		return models.AgencyUserProfile{}, ErrTooManyAttempts
	}

	userID, exists := m.agencyUsersByEmail[email]
	if !exists {
		m.recordFailure(attemptKey, now)
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}

	user := m.agencyUsersByID[userID]
	if !utils.VerifyCredential(password, user.PasswordHash) {
		m.recordFailure(attemptKey, now)
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}
	if user.MFARequired && !user.MFAEnabled {
		return models.AgencyUserProfile{}, ErrMFASetupRequired
	}
	if user.MFARequired && mfaCode == "" {
		return models.AgencyUserProfile{}, ErrMFARequired
	}
	if user.MFARequired && !utils.SecureCompare(user.MFACode, mfaCode) {
		m.recordFailure(attemptKey, now)
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}

	agency, exists := m.agenciesByID[user.AgencyID]
	if !exists {
		return models.AgencyUserProfile{}, ErrUnknownAgency
	}

	m.resetFailures(attemptKey)
	return utils.AgencyProfileFromUser(user, agency), nil
}

func (m *MemoryStore) authorityUserByCredentials(userID string, email string, password string) (models.AgencyUser, models.AgencyRecord, error) {
	user, exists := m.agencyUsersByID[userID]
	if !exists || user.Email != email || !utils.VerifyCredential(password, user.PasswordHash) {
		return models.AgencyUser{}, models.AgencyRecord{}, ErrInvalidCredentials
	}

	agency, exists := m.agenciesByID[user.AgencyID]
	if !exists {
		return models.AgencyUser{}, models.AgencyRecord{}, ErrUnknownAgency
	}

	return user, agency, nil
}

// VerifyOTP validates a citizen login challenge.
func (m *MemoryStore) VerifyOTP(phone string, code string, now time.Time) (models.CitizenProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	attemptKey := "citizen-otp:" + phone
	if m.lockedOut(attemptKey, now) {
		return models.CitizenProfile{}, ErrTooManyAttempts
	}

	userID, exists := m.usersByPhone[phone]
	if !exists {
		m.recordFailure(attemptKey, now)
		return models.CitizenProfile{}, ErrInvalidCredentials
	}

	challenge, exists := m.challenges[phone]
	if !exists || !utils.SecureCompare(challenge.Code, code) || now.After(challenge.ExpiresAt) {
		m.recordFailure(attemptKey, now)
		return models.CitizenProfile{}, ErrInvalidCredentials
	}

	delete(m.challenges, phone)
	m.resetFailures(attemptKey)
	return m.usersByID[userID], nil
}

// lockedOut reports whether the attempt key is inside an active lockout
// window. Callers must hold the lock.
func (m *MemoryStore) lockedOut(key string, now time.Time) bool {
	attempts, exists := m.failedAttempts[key]
	return exists && attempts.lockedUntil.After(now)
}

// recordFailure increments the consecutive-failure counter for the key and
// starts a lockout window once the threshold is reached. An expired lockout
// resets the counter so a legitimate user gets a fresh set of attempts.
// Callers must hold the lock.
func (m *MemoryStore) recordFailure(key string, now time.Time) {
	attempts, exists := m.failedAttempts[key]
	if !exists {
		attempts = &failedAttemptLog{}
		m.failedAttempts[key] = attempts
	}
	if attempts.count >= maxFailedAttempts && !attempts.lockedUntil.After(now) {
		attempts.count = 0
	}
	attempts.count++
	if attempts.count >= maxFailedAttempts {
		attempts.lockedUntil = now.Add(attemptLockout)
	}
}

// resetFailures clears the failure counter for the key after a successful
// verification. Callers must hold the lock.
func (m *MemoryStore) resetFailures(key string) {
	delete(m.failedAttempts, key)
}

// ProfileByID returns a citizen profile by ID.
func (m *MemoryStore) ProfileByID(id string) (models.CitizenProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	profile, ok := m.usersByID[id]
	return profile, ok
}

// AgencyProfileByID returns an agency user profile by ID.
func (m *MemoryStore) AgencyProfileByID(id string) (models.AgencyUserProfile, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.agencyUsersByID[id]
	if !ok {
		return models.AgencyUserProfile{}, false
	}

	agency, ok := m.agenciesByID[user.AgencyID]
	if !ok {
		return models.AgencyUserProfile{}, false
	}

	return utils.AgencyProfileFromUser(user, agency), true
}

// ListAgencies returns the agency directory, sorted by name for a stable
// response order.
func (m *MemoryStore) ListAgencies() []models.AgencySummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agencies := make([]models.AgencySummary, 0, len(m.agenciesByID))
	for _, agency := range m.agenciesByID {
		agencies = append(agencies, utils.AgencySummaryFromRecord(agency))
	}
	slices.SortFunc(agencies, func(a, b models.AgencySummary) int {
		return strings.Compare(a.Name, b.Name)
	})

	return agencies
}

// AppendAuditLog records an audit event.
func (m *MemoryStore) AppendAuditLog(record models.AuditLogRecord) models.AuditLogRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.auditLogs = append(m.auditLogs, record)
	return record
}

// ListAuditLogs returns the most recent audit events up to limit.
func (m *MemoryStore) ListAuditLogs(limit int) []models.AuditLogRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	logs := make([]models.AuditLogRecord, 0, min(limit, len(m.auditLogs)))
	for i := len(m.auditLogs) - 1; i >= 0 && len(logs) < limit; i-- {
		logs = append(logs, m.auditLogs[i])
	}

	return logs
}
