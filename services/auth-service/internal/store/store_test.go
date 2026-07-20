package store

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

var storeTestNow = time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)

func newTestMemoryStore(t *testing.T) *MemoryStore {
	t.Helper()
	m, ok := NewMemoryStore(storeTestNow, &config.Config{}).(*MemoryStore)
	if !ok {
		t.Fatal("expected *MemoryStore from NewMemoryStore")
	}
	return m
}

func TestSeedBootstrapAgencyAdminGeneratesRandomMFASecret(t *testing.T) {
	m := newTestMemoryStore(t)
	cfg := &config.Config{
		BootstrapAdminEmail:    "admin@nadaa.local",
		BootstrapAdminPassword: "bootstrap-pass-123",
		BootstrapAdminPhone:    "+233200000001",
		BootstrapAdminName:     "NADAA System Admin",
		// A mock OTP must never be reused as the bootstrap MFA fallback.
		MockOTP: "123456",
	}

	if err := seedBootstrapAgencyAdmin(m, cfg, storeTestNow); err != nil {
		t.Fatalf("seed bootstrap admin: %v", err)
	}

	userID, exists := m.agencyUsersByEmail["admin@nadaa.local"]
	if !exists {
		t.Fatal("expected bootstrap admin to be seeded")
	}
	user := m.agencyUsersByID[userID]
	if !user.MFAEnabled || !utils.ValidTOTPSecret(user.MFASecret) {
		t.Fatalf("expected generated TOTP secret, got %#v", user)
	}
	if user.Role != models.RoleSystemAdmin {
		t.Fatalf("expected system_admin role, got %q", user.Role)
	}

	// The generated secret must be a usable TOTP seed: the current code from
	// it has to pass login verification.
	code, err := utils.TOTPCode(user.MFASecret, storeTestNow)
	if err != nil {
		t.Fatalf("compute bootstrap TOTP code: %v", err)
	}
	if _, err := m.LoginAgencyUser("admin@nadaa.local", "bootstrap-pass-123", code, storeTestNow); err != nil {
		t.Fatalf("expected bootstrap admin login with TOTP code, got %v", err)
	}
}

func TestSeedBootstrapAgencyAdminHonorsExplicitMFASecret(t *testing.T) {
	m := newTestMemoryStore(t)
	cfg := &config.Config{
		BootstrapAdminEmail:     "admin@nadaa.local",
		BootstrapAdminPassword:  "bootstrap-pass-123",
		BootstrapAdminPhone:     "+233200000001",
		BootstrapAdminName:      "NADAA System Admin",
		BootstrapAdminMFASecret: "JBSWY3DPEHPK3PXP",
	}

	if err := seedBootstrapAgencyAdmin(m, cfg, storeTestNow); err != nil {
		t.Fatalf("seed bootstrap admin: %v", err)
	}

	user := m.agencyUsersByID[m.agencyUsersByEmail["admin@nadaa.local"]]
	if user.MFASecret != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("expected explicit MFA secret to be stored, got %q", user.MFASecret)
	}
}

func TestSeedBootstrapAgencyAdminValidatesCredentialComplexity(t *testing.T) {
	base := config.Config{
		BootstrapAdminEmail:     "admin@nadaa.local",
		BootstrapAdminPassword:  "bootstrap-pass-123",
		BootstrapAdminPhone:     "+233200000001",
		BootstrapAdminName:      "NADAA System Admin",
		BootstrapAdminMFASecret: "JBSWY3DPEHPK3PXP",
	}

	shortPassword := base
	shortPassword.BootstrapAdminPassword = "short"
	if err := seedBootstrapAgencyAdmin(newTestMemoryStore(t), &shortPassword, storeTestNow); err == nil {
		t.Fatal("expected short bootstrap password to be rejected")
	}

	for _, secret := range []string{"!!!!!!", "ABC", "AB", "123456", "mfa_secret_1234"} {
		invalidSecret := base
		invalidSecret.BootstrapAdminMFASecret = secret
		if err := seedBootstrapAgencyAdmin(newTestMemoryStore(t), &invalidSecret, storeTestNow); err == nil {
			t.Fatalf("expected MFA secret %q to be rejected", secret)
		}
	}

	missingPassword := base
	missingPassword.BootstrapAdminPassword = ""
	if err := seedBootstrapAgencyAdmin(newTestMemoryStore(t), &missingPassword, storeTestNow); err == nil {
		t.Fatal("expected email without password to be rejected")
	}
}

func TestHashCredentialUsesSaltedPBKDF2(t *testing.T) {
	hash := utils.HashCredential("correct horse battery staple")
	if !strings.HasPrefix(hash, "pbkdf2$") {
		t.Fatalf("expected pbkdf2 hash format, got %q", hash)
	}
	if !utils.VerifyCredential("correct horse battery staple", hash) {
		t.Fatal("expected credential to verify against its hash")
	}
	if utils.VerifyCredential("wrong password", hash) {
		t.Fatal("expected wrong credential to be rejected")
	}
	first := utils.HashCredential("same")
	second := utils.HashCredential("same")
	if first == second {
		t.Fatal("expected random salts to produce distinct hashes")
	}
}

func TestVerifyCredentialAcceptsLegacySHA256Digests(t *testing.T) {
	sum := sha256.Sum256([]byte("legacy-password"))
	legacyHex := hex.EncodeToString(sum[:])
	legacyBase64 := base64.RawURLEncoding.EncodeToString(sum[:])

	for name, stored := range map[string]string{"hex": legacyHex, "base64url": legacyBase64} {
		if !utils.VerifyCredential("legacy-password", stored) {
			t.Fatalf("expected legacy %s digest to verify", name)
		}
		if utils.VerifyCredential("other-password", stored) {
			t.Fatalf("expected legacy %s digest to reject a wrong password", name)
		}
	}
}

func TestLoginAgencyUserAcceptsLegacyPasswordHash(t *testing.T) {
	m := newTestMemoryStore(t)
	profile, err := m.CreateAgencyUser(models.CreateAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleDispatcher,
	}, "Password123!", storeTestNow)
	if err != nil {
		t.Fatalf("create agency user: %v", err)
	}
	m.enableAgencyMFA(profile.ID, "JBSWY3DPEHPK3PXP", storeTestNow)

	// Rewrite the stored hash to the legacy unsalted SHA-256 format.
	user := m.agencyUsersByID[profile.ID]
	sum := sha256.Sum256([]byte("Password123!"))
	user.PasswordHash = hex.EncodeToString(sum[:])
	m.agencyUsersByID[profile.ID] = user

	code, err := utils.TOTPCode("JBSWY3DPEHPK3PXP", storeTestNow)
	if err != nil {
		t.Fatalf("compute TOTP code: %v", err)
	}
	if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "Password123!", code, storeTestNow); err != nil {
		t.Fatalf("expected legacy hash to verify, got %v", err)
	}
}

func TestLoginAgencyUserLockoutExpires(t *testing.T) {
	m := newTestMemoryStore(t)
	profile, err := m.CreateAgencyUser(models.CreateAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleDispatcher,
	}, "Password123!", storeTestNow)
	if err != nil {
		t.Fatalf("create agency user: %v", err)
	}
	m.enableAgencyMFA(profile.ID, "JBSWY3DPEHPK3PXP", storeTestNow)

	for attempt := range maxFailedAttempts {
		if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "wrong-password", "", storeTestNow); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", attempt+1, err)
		}
	}
	code, err := utils.TOTPCode("JBSWY3DPEHPK3PXP", storeTestNow)
	if err != nil {
		t.Fatalf("compute TOTP code: %v", err)
	}
	if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "Password123!", code, storeTestNow); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected ErrTooManyAttempts during lockout, got %v", err)
	}

	afterLockout := storeTestNow.Add(attemptLockout + time.Minute)
	freshCode, err := utils.TOTPCode("JBSWY3DPEHPK3PXP", afterLockout)
	if err != nil {
		t.Fatalf("compute TOTP code: %v", err)
	}
	if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "Password123!", freshCode, afterLockout); err != nil {
		t.Fatalf("expected login to succeed after lockout expiry, got %v", err)
	}
}

func TestChangeAgencyPasswordVerifiesCurrentAndLocksOut(t *testing.T) {
	m := newTestMemoryStore(t)
	profile, err := m.CreateAgencyUser(models.CreateAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleDispatcher,
	}, "Password123!", storeTestNow)
	if err != nil {
		t.Fatalf("create agency user: %v", err)
	}

	if err := m.ChangeAgencyPassword(profile.ID, "wrong-password", "NewPassword456!", storeTestNow); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for wrong current password, got %v", err)
	}
	if err := m.ChangeAgencyPassword(profile.ID, "Password123!", "short", storeTestNow); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword for weak new password, got %v", err)
	}
	if err := m.ChangeAgencyPassword(profile.ID, "Password123!", "NewPassword456!", storeTestNow); err != nil {
		t.Fatalf("expected password change to succeed, got %v", err)
	}

	user := m.agencyUsersByID[profile.ID]
	if !utils.VerifyCredential("NewPassword456!", user.PasswordHash) {
		t.Fatal("expected stored hash to match the new password")
	}
	if utils.VerifyCredential("Password123!", user.PasswordHash) {
		t.Fatal("expected stored hash to reject the old password")
	}

	for attempt := range maxFailedAttempts {
		if err := m.ChangeAgencyPassword(profile.ID, "wrong-password", "AnotherPassword789!", storeTestNow); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", attempt+1, err)
		}
	}
	if err := m.ChangeAgencyPassword(profile.ID, "NewPassword456!", "AnotherPassword789!", storeTestNow); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected ErrTooManyAttempts during lockout, got %v", err)
	}
}

func TestListAgencyUsersSanitizesAndReportsLockout(t *testing.T) {
	m := newTestMemoryStore(t)
	profile, err := m.CreateAgencyUser(models.CreateAgencyUserRequest{
		Name:     "Dispatcher One",
		Email:    "dispatcher@nadaa.local",
		Phone:    "+233200000002",
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleDispatcher,
	}, "Password123!", storeTestNow)
	if err != nil {
		t.Fatalf("create agency user: %v", err)
	}

	users := m.ListAgencyUsers(storeTestNow)
	if len(users) != 1 {
		t.Fatalf("expected one directory entry, got %#v", users)
	}
	entry := users[0]
	if entry.ID != profile.ID || entry.Email != "dispatcher@nadaa.local" || entry.Role != models.RoleDispatcher || entry.AgencyID != models.DefaultAgencyID {
		t.Fatalf("unexpected directory entry: %#v", entry)
	}
	if entry.MFAEnabled || entry.LockedUntil != nil {
		t.Fatalf("expected mfaEnabled=false and no lockout, got %#v", entry)
	}

	// Drive the login lockout and confirm the directory reports its deadline.
	for range maxFailedAttempts {
		_, _ = m.LoginAgencyUser("dispatcher@nadaa.local", "wrong-password", "", storeTestNow)
	}
	users = m.ListAgencyUsers(storeTestNow)
	if users[0].LockedUntil == nil || !users[0].LockedUntil.After(storeTestNow) {
		t.Fatalf("expected active lockout deadline, got %#v", users[0])
	}
}
