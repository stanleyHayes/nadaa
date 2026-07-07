package store

import (
	"errors"
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

	phone := utils.NormalizePhone(cfg.BootstrapAdminPhone)
	name := strings.TrimSpace(cfg.BootstrapAdminName)
	mfaCode := strings.TrimSpace(cfg.BootstrapAdminMFACode)
	if mfaCode == "" {
		mfaCode = cfg.MockOTP
	}
	if mfaCode == "" {
		mfaCode = "123456"
	}

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
