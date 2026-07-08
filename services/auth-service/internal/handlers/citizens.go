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

func (s *Server) registerCitizenHandler(w http.ResponseWriter, r *http.Request) {
	var request models.RegisterCitizenRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Phone = utils.NormalizePhone(request.Phone)
	request.PreferredLanguage = utils.NormalizeLanguage(request.PreferredLanguage)

	if request.Name == "" {
		utils.WriteError(w, http.StatusBadRequest, "name_required", "name is required")
		return
	}

	if !utils.ValidPhone(request.Phone) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_phone", "phone must be in E.164 format, for example +233200000000")
		return
	}

	if request.HomeLocation != nil && !utils.ValidCoordinates(*request.HomeLocation) {
		utils.WriteError(w, http.StatusBadRequest, "invalid_home_location", "homeLocation must contain valid lat and lng values")
		return
	}

	code, err := s.otp.Generate()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "otp_generation_failed", "could not create login challenge")
		return
	}

	profile, challenge, err := s.store.RegisterCitizen(request, code, s.now())
	if errors.Is(err, store.ErrDuplicatePhone) {
		utils.WriteError(w, http.StatusConflict, "phone_already_registered", "phone is already registered")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "registration_failed", "could not register citizen")
		return
	}

	response := models.RegisterCitizenResponse{
		UserID:      profile.ID,
		Phone:       profile.Phone,
		ChallengeID: challenge.ID,
		OTPDelivery: "mock",
	}
	if s.exposeDevOTP {
		response.DevOTP = challenge.Code
	}

	s.recordAudit(r, utils.AuditActorFromCitizen(profile), "auth.citizen.registered", models.AuditTarget{Type: "citizen_user", ID: profile.ID}, nil, utils.CitizenAuditSnapshot(profile))
	utils.WriteJSON(w, http.StatusCreated, response)
}

func (s *Server) loginCitizenHandler(w http.ResponseWriter, r *http.Request) {
	var request models.LoginCitizenRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	request.Phone = utils.NormalizePhone(request.Phone)
	request.OTP = strings.TrimSpace(request.OTP)

	if !utils.ValidPhone(request.Phone) || request.OTP == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid_login_request", "phone and otp are required")
		return
	}

	profile, err := s.store.VerifyOTP(request.Phone, request.OTP, s.now())
	if errors.Is(err, store.ErrInvalidCredentials) {
		s.recordAudit(r, models.AuditActor{}, "auth.citizen_login.failed", models.AuditTarget{Type: "citizen_phone", ID: request.Phone}, nil, map[string]any{
			"reason": "invalid_credentials",
		})
		utils.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "phone or otp is invalid")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "login_failed", "could not complete login")
		return
	}

	expiresAt := s.now().Add(24 * time.Hour)
	token, err := s.signToken(profile, expiresAt)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "token_generation_failed", "could not create access token")
		return
	}

	s.recordAudit(r, utils.AuditActorFromCitizen(profile), "auth.citizen_login.succeeded", models.AuditTarget{Type: "citizen_user", ID: profile.ID}, nil, map[string]any{
		"expiresAt": expiresAt,
	})
	utils.WriteJSON(w, http.StatusOK, models.LoginCitizenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        profile,
	})
}

func (s *Server) meHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := utils.BearerToken(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return
	}

	claims, err := s.verifyToken(token)
	if errors.Is(err, store.ErrInvalidToken) {
		utils.WriteError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "token_verification_failed", "could not verify token")
		return
	}

	if claims.UserType == "agency" {
		profile, ok := s.store.AgencyProfileByID(claims.UserID)
		if !ok {
			utils.WriteError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
			return
		}

		utils.WriteJSON(w, http.StatusOK, profile)
		return
	}

	profile, ok := s.store.ProfileByID(claims.UserID)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
		return
	}

	utils.WriteJSON(w, http.StatusOK, profile)
}
