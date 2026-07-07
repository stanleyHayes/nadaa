package store

import (
	"errors"
	"sync"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

var (
	ErrDuplicatePhone     = errors.New("duplicate phone")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidRole        = errors.New("invalid role")
	ErrMFAAlreadyEnabled  = errors.New("mfa already enabled")
	ErrMFARequired        = errors.New("mfa required")
	ErrMFASetupRequired   = errors.New("mfa setup required")
	ErrUnknownAgency      = errors.New("unknown agency")
)

// Store is the persistence interface for auth data.
type Store interface {
	RegisterCitizen(request models.RegisterCitizenRequest, code string, now time.Time) (models.CitizenProfile, models.OTPChallenge, error)
	VerifyOTP(phone string, code string, now time.Time) (models.CitizenProfile, error)
	ProfileByID(id string) (models.CitizenProfile, bool)
	CreateAgencyUser(request models.CreateAgencyUserRequest, temporaryPassword string, now time.Time) (models.AgencyUserProfile, error)
	StartAgencyMFASetup(userID string, email string, temporaryPassword string, secret string, code string, now time.Time) (models.MFAChallenge, error)
	VerifyAgencyMFA(userID string, email string, temporaryPassword string, code string, now time.Time) (models.AgencyUserProfile, error)
	LoginAgencyUser(email string, password string, mfaCode string) (models.AgencyUserProfile, error)
	AgencyProfileByID(id string) (models.AgencyUserProfile, bool)
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
	auditLogs           []models.AuditLogRecord
}

// NewMemoryStore creates an in-memory store seeded with fixture data.
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

	seedBootstrapAgencyAdmin(m, cfg, now)
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

	user, agency, err := m.authorityUserByCredentials(userID, email, temporaryPassword)
	if err != nil {
		return models.AgencyUserProfile{}, err
	}

	challenge, exists := m.mfaChallengesByUser[user.ID]
	if !exists || challenge.Code != code || now.After(challenge.ExpiresAt) {
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}

	delete(m.mfaChallengesByUser, user.ID)
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
func (m *MemoryStore) LoginAgencyUser(email string, password string, mfaCode string) (models.AgencyUserProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userID, exists := m.agencyUsersByEmail[email]
	if !exists {
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}

	user := m.agencyUsersByID[userID]
	if user.PasswordHash != utils.HashCredential(password) {
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}
	if user.MFARequired && !user.MFAEnabled {
		return models.AgencyUserProfile{}, ErrMFASetupRequired
	}
	if user.MFARequired && mfaCode == "" {
		return models.AgencyUserProfile{}, ErrMFARequired
	}
	if user.MFARequired && user.MFACode != mfaCode {
		return models.AgencyUserProfile{}, ErrInvalidCredentials
	}

	agency, exists := m.agenciesByID[user.AgencyID]
	if !exists {
		return models.AgencyUserProfile{}, ErrUnknownAgency
	}

	return utils.AgencyProfileFromUser(user, agency), nil
}

func (m *MemoryStore) authorityUserByCredentials(userID string, email string, password string) (models.AgencyUser, models.AgencyRecord, error) {
	user, exists := m.agencyUsersByID[userID]
	if !exists || user.Email != email || user.PasswordHash != utils.HashCredential(password) {
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

	userID, exists := m.usersByPhone[phone]
	if !exists {
		return models.CitizenProfile{}, ErrInvalidCredentials
	}

	challenge, exists := m.challenges[phone]
	if !exists || challenge.Code != code || now.After(challenge.ExpiresAt) {
		return models.CitizenProfile{}, ErrInvalidCredentials
	}

	delete(m.challenges, phone)
	return m.usersByID[userID], nil
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
