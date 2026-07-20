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

func seedBootstrapAgencyAdmin(m *MemoryStore, cfg *config.Config, now time.Time) error {
	email := utils.NormalizeEmail(cfg.BootstrapAdminEmail)
	password := strings.TrimSpace(cfg.BootstrapAdminPassword)
	if email == "" && password == "" {
		return nil
	}
	if !utils.ValidEmail(email) || password == "" {
		return errors.New("NADAA_AUTH_BOOTSTRAP_ADMIN_EMAIL and NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD are required together")
	}
	if len(password) < MinAgencyPasswordLength {
		return fmt.Errorf("NADAA_AUTH_BOOTSTRAP_ADMIN_PASSWORD must be at least %d characters", MinAgencyPasswordLength)
	}

	// The bootstrap MFA seed must never fall back to a constant: require an
	// explicit operator-provided base32 TOTP secret, or generate a random one
	// and surface its otpauth URL once in the startup log so the operator can
	// enroll an authenticator for the first login.
	mfaSecret := strings.TrimSpace(cfg.BootstrapAdminMFASecret)
	if mfaSecret == "" {
		mfaSecret = utils.NewTOTPSecret()
		log.Printf("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET is not set: generated a random bootstrap admin TOTP secret; enroll an authenticator with %s", utils.TOTPAuthURL(mfaSecret, email))
	} else if !utils.ValidTOTPSecret(mfaSecret) {
		return errors.New("NADAA_AUTH_BOOTSTRAP_ADMIN_MFA_SECRET must be a base32-encoded TOTP secret")
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

	m.enableAgencyMFA(profile.ID, mfaSecret, now)
	return nil
}
