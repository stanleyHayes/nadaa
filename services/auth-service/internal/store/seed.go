package store

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

// minBootstrapPasswordLength is the minimum length for the bootstrap admin
// password; the most privileged account on the platform must not start with a
// weak credential.
const minBootstrapPasswordLength = 12

func seedBootstrapAgencyAdmin(m *MemoryStore, cfg *config.Config, now time.Time) error {
	email := utils.NormalizeEmail(cfg.BootstrapAdminEmail)
	password := strings.TrimSpace(cfg.BootstrapAdminPassword)
	if email == "" && password == "" {
		return nil
	}
	if !utils.ValidEmail(email) || password == "" {
		return errors.New("NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL and NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD are required together")
	}
	if len(password) < minBootstrapPasswordLength {
		return fmt.Errorf("NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD must be at least %d characters", minBootstrapPasswordLength)
	}

	// The bootstrap MFA code must never fall back to a constant: require an
	// explicit operator-provided code, or generate a random one and surface it
	// once in the startup log so the operator can complete the first login.
	mfaCode := strings.TrimSpace(cfg.BootstrapAdminMFACode)
	if mfaCode == "" {
		generated, err := (utils.RandomOTPGenerator{}).Generate()
		if err != nil {
			return fmt.Errorf("generate bootstrap admin MFA code: %w", err)
		}
		mfaCode = generated
		log.Printf("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE is not set: generated one-time bootstrap admin MFA code %s", mfaCode)
	}
	if !utils.ValidSixDigitCode(mfaCode) {
		return errors.New("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_CODE must be exactly 6 digits")
	}

	phone := utils.NormalizePhone(cfg.BootstrapAdminPhone)
	name := strings.TrimSpace(cfg.BootstrapAdminName)

	profile, err := m.CreateAgencyUser(models.CreateAgencyUserRequest{
		Name:     name,
		Email:    email,
		Phone:    phone,
		AgencyID: models.DefaultAgencyID,
		Role:     models.RoleSystemAdmin,
	}, password, now)
	if err != nil {
		return err
	}

	m.enableAgencyMFA(profile.ID, mfaCode, now)
	return nil
}
