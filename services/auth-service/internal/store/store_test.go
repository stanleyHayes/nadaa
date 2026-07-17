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

func TestSeedBootstrapAgencyAdminGeneratesRandomMFACode(t *testing.T) {
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
	if !user.MFAEnabled || !utils.ValidSixDigitCode(user.MFACode) {
		t.Fatalf("expected generated 6-digit MFA code, got %#v", user)
	}
	if user.Role != models.RoleSystemAdmin {
		t.Fatalf("expected system_admin role, got %q", user.Role)
	}
}

func TestSeedBootstrapAgencyAdminHonorsExplicitMFACode(t *testing.T) {
	m := newTestMemoryStore(t)
	cfg := &config.Config{
		BootstrapAdminEmail:    "admin@nadaa.local",
		BootstrapAdminPassword: "bootstrap-pass-123",
		BootstrapAdminPhone:    "+233200000001",
		BootstrapAdminName:     "NADAA System Admin",
		BootstrapAdminMFACode:  "456789",
	}

	if err := seedBootstrapAgencyAdmin(m, cfg, storeTestNow); err != nil {
		t.Fatalf("seed bootstrap admin: %v", err)
	}

	user := m.agencyUsersByID[m.agencyUsersByEmail["admin@nadaa.local"]]
	if user.MFACode != "456789" {
		t.Fatalf("expected explicit MFA code to be stored, got %q", user.MFACode)
	}
}

func TestSeedBootstrapAgencyAdminValidatesCredentialComplexity(t *testing.T) {
	base := config.Config{
		BootstrapAdminEmail:    "admin@nadaa.local",
		BootstrapAdminPassword: "bootstrap-pass-123",
		BootstrapAdminPhone:    "+233200000001",
		BootstrapAdminName:     "NADAA System Admin",
		BootstrapAdminMFACode:  "456789",
	}

	shortPassword := base
	shortPassword.BootstrapAdminPassword = "short"
	if err := seedBootstrapAgencyAdmin(newTestMemoryStore(t), &shortPassword, storeTestNow); err == nil {
		t.Fatal("expected short bootstrap password to be rejected")
	}

	for _, code := range []string{"12345", "1234567", "abcdef", "12345 "} {
		invalidCode := base
		invalidCode.BootstrapAdminMFACode = code
		if err := seedBootstrapAgencyAdmin(newTestMemoryStore(t), &invalidCode, storeTestNow); err == nil {
			t.Fatalf("expected MFA code %q to be rejected", code)
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
	m.enableAgencyMFA(profile.ID, "123456", storeTestNow)

	// Rewrite the stored hash to the legacy unsalted SHA-256 format.
	user := m.agencyUsersByID[profile.ID]
	sum := sha256.Sum256([]byte("Password123!"))
	user.PasswordHash = hex.EncodeToString(sum[:])
	m.agencyUsersByID[profile.ID] = user

	if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "Password123!", "123456", storeTestNow); err != nil {
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
	m.enableAgencyMFA(profile.ID, "123456", storeTestNow)

	for attempt := range maxFailedAttempts {
		if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "wrong-password", "", storeTestNow); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("attempt %d: expected ErrInvalidCredentials, got %v", attempt+1, err)
		}
	}
	if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "Password123!", "123456", storeTestNow); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected ErrTooManyAttempts during lockout, got %v", err)
	}

	afterLockout := storeTestNow.Add(attemptLockout + time.Minute)
	if _, err := m.LoginAgencyUser("dispatcher@nadaa.local", "Password123!", "123456", afterLockout); err != nil {
		t.Fatalf("expected login to succeed after lockout expiry, got %v", err)
	}
}
