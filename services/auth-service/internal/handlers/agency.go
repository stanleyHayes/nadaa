package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

func (s *Server) createAgencyUserHandler(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAgencyRole(w, r, models.RoleSystemAdmin, models.RoleAgencyAdmin)
	if !ok {
		return
	}

	var request models.CreateAgencyUserRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = utils.NormalizeEmail(request.Email)
	request.Phone = utils.NormalizePhone(request.Phone)
	request.AgencyID = strings.TrimSpace(request.AgencyID)
	request.Role = utils.NormalizeRole(request.Role)

	if request.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "name_required", "name is required")
		return
	}
	if !utils.ValidEmail(request.Email) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_email", "email must be valid")
		return
	}
	if !utils.ValidPhone(request.Phone) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_phone", "phone must be in E.164 format, for example +233200000000")
		return
	}
	if request.AgencyID == "" {
		utils.WriteError(w, http.StatusBadRequest, "agency_required", "agencyId is required")
		return
	}
	if !utils.ValidAgencyRole(request.Role) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_role", "role must be an authority role")
		return
	}
	if actor.Role == models.RoleAgencyAdmin && actor.Agency.ID != request.AgencyID {
		s.recordAudit(r, utils.AuditActorFromAgency(actor), "auth.rbac.denied", models.AuditTarget{Type: "agency_user", ID: request.Email}, nil, map[string]any{
			"reason":            "agency_scope_forbidden",
			"requestedRole":     request.Role,
			"requestedAgencyId": request.AgencyID,
		})
		utils.WriteError(w, http.StatusForbidden, "agency_scope_forbidden", "agency admins can create users only inside their agency")
		return
	}

	temporaryPassword := utils.NewTemporaryPassword()
	profile, err := s.store.CreateAgencyUser(request, temporaryPassword, s.now())
	if errors.Is(err, store.ErrDuplicateEmail) {
		utils.WriteError(w, http.StatusConflict, "email_already_registered", "email is already registered")
		return
	}
	if errors.Is(err, store.ErrDuplicatePhone) {
		utils.WriteError(w, http.StatusConflict, "phone_already_registered", "phone is already registered")
		return
	}
	if errors.Is(err, store.ErrUnknownAgency) {
		utils.WriteError(w, http.StatusBadRequest, "agency_not_found", "agencyId does not exist")
		return
	}
	if errors.Is(err, store.ErrInvalidRole) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_role", "role must be an authority role")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "agency_user_creation_failed", "could not create agency user")
		return
	}

	s.recordAudit(r, utils.AuditActorFromAgency(actor), "auth.agency_user.created", models.AuditTarget{Type: "agency_user", ID: profile.ID}, nil, utils.AgencyUserAuditSnapshot(profile))
	utils.WriteJSON(w, http.StatusCreated, models.CreateAgencyUserResponse{
		User:              profile,
		TemporaryPassword: temporaryPassword,
		MFASetupRequired:  true,
	})
}

func (s *Server) setupAgencyMFAHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("id"))
	if userID == "" {
		utils.WriteError(w, http.StatusBadRequest, "user_id_required", "agency user id is required")
		return
	}

	var request models.AgencyMFASetupRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Email = utils.NormalizeEmail(request.Email)
	request.TemporaryPassword = strings.TrimSpace(request.TemporaryPassword)
	if !utils.ValidEmail(request.Email) || request.TemporaryPassword == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_mfa_setup_request", "email and temporaryPassword are required")
		return
	}

	code, err := s.otp.Generate()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "mfa_generation_failed", "could not create MFA challenge")
		return
	}

	challenge, err := s.store.StartAgencyMFASetup(userID, request.Email, request.TemporaryPassword, utils.NewMFASecret(), code, s.now())
	if errors.Is(err, store.ErrInvalidCredentials) {
		s.recordAudit(r, models.AuditActor{}, "auth.agency_mfa.setup_failed", models.AuditTarget{Type: "agency_user", ID: userID}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		utils.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "agency user or temporary password is invalid")
		return
	}
	if errors.Is(err, store.ErrMFAAlreadyEnabled) {
		utils.WriteError(w, http.StatusConflict, "mfa_already_enabled", "MFA is already enabled for this agency user")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "mfa_setup_failed", "could not start MFA setup")
		return
	}

	response := models.AgencyMFASetupResponse{
		UserID:      userID,
		ChallengeID: challenge.ID,
		Method:      "mock_totp",
		Secret:      challenge.Secret,
		ExpiresAt:   challenge.ExpiresAt,
	}
	if s.exposeDevOTP {
		response.DevCode = challenge.Code
	}

	profile, _ := s.store.AgencyProfileByID(userID)
	s.recordAudit(r, utils.AuditActorFromAgency(profile), "auth.agency_mfa.setup_started", models.AuditTarget{Type: "agency_user", ID: userID}, nil, map[string]any{
		"challengeId": challenge.ID,
		"method":      response.Method,
		"expiresAt":   challenge.ExpiresAt,
	})
	utils.WriteJSON(w, http.StatusOK, response)
}

func (s *Server) verifyAgencyMFAHandler(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.PathValue("id"))
	if userID == "" {
		utils.WriteError(w, http.StatusBadRequest, "user_id_required", "agency user id is required")
		return
	}

	var request models.AgencyMFAVerifyRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Email = utils.NormalizeEmail(request.Email)
	request.TemporaryPassword = strings.TrimSpace(request.TemporaryPassword)
	request.Code = strings.TrimSpace(request.Code)
	if !utils.ValidEmail(request.Email) || request.TemporaryPassword == "" || request.Code == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_mfa_verify_request", "email, temporaryPassword, and code are required")
		return
	}

	profile, err := s.store.VerifyAgencyMFA(userID, request.Email, request.TemporaryPassword, request.Code, s.now())
	if errors.Is(err, store.ErrInvalidCredentials) {
		s.recordAudit(r, models.AuditActor{}, "auth.agency_mfa.verify_failed", models.AuditTarget{Type: "agency_user", ID: userID}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		utils.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "agency user, temporary password, or MFA code is invalid")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "mfa_verification_failed", "could not verify MFA")
		return
	}

	s.recordAudit(r, utils.AuditActorFromAgency(profile), "auth.agency_mfa.verified", models.AuditTarget{Type: "agency_user", ID: profile.ID}, nil, utils.AgencyUserAuditSnapshot(profile))
	utils.WriteJSON(w, http.StatusOK, models.AgencyMFAVerifyResponse{User: profile})
}

func (s *Server) loginAgencyHandler(w http.ResponseWriter, r *http.Request) {
	var request models.LoginAgencyRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Email = utils.NormalizeEmail(request.Email)
	request.Password = strings.TrimSpace(request.Password)
	request.MFACode = strings.TrimSpace(request.MFACode)
	if !utils.ValidEmail(request.Email) || request.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_login_request", "email and password are required")
		return
	}

	profile, err := s.store.LoginAgencyUser(request.Email, request.Password, request.MFACode)
	if errors.Is(err, store.ErrMFASetupRequired) {
		s.recordAudit(r, models.AuditActor{}, "auth.agency_login.blocked", models.AuditTarget{Type: "agency_email", ID: request.Email}, nil, map[string]any{
			"reason": "mfa_setup_required",
		})
		utils.WriteError(w, http.StatusForbidden, "mfa_setup_required", "MFA must be set up before login")
		return
	}
	if errors.Is(err, store.ErrMFARequired) {
		s.recordAudit(r, models.AuditActor{}, "auth.agency_login.failed", models.AuditTarget{Type: "agency_email", ID: request.Email}, nil, map[string]any{
			"reason": "mfa_required",
		})
		utils.WriteError(w, http.StatusUnauthorized, "mfa_required", "MFA code is required")
		return
	}
	if errors.Is(err, store.ErrInvalidCredentials) {
		s.recordAudit(r, models.AuditActor{}, "auth.agency_login.failed", models.AuditTarget{Type: "agency_email", ID: request.Email}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		utils.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "email, password, or MFA code is invalid")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "login_failed", "could not complete agency login")
		return
	}

	expiresAt := s.now().Add(12 * time.Hour)
	token, err := s.signAgencyToken(profile, expiresAt)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "token_generation_failed", "could not create access token")
		return
	}

	s.recordAudit(r, utils.AuditActorFromAgency(profile), "auth.agency_login.succeeded", models.AuditTarget{Type: "agency_user", ID: profile.ID}, nil, map[string]any{
		"expiresAt": expiresAt,
	})
	utils.WriteJSON(w, http.StatusOK, models.LoginAgencyResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        profile,
	})
}
