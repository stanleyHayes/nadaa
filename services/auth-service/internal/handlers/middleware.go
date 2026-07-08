package handlers

import (
	"errors"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

func (s *Server) requireAgencyRole(w http.ResponseWriter, r *http.Request, allowedRoles ...string) (models.AgencyUserProfile, bool) {
	token, ok := utils.BearerToken(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return models.AgencyUserProfile{}, false
	}

	claims, err := s.verifyToken(token)
	if errors.Is(err, store.ErrInvalidToken) {
		utils.WriteError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return models.AgencyUserProfile{}, false
	}
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "token_verification_failed", "could not verify token")
		return models.AgencyUserProfile{}, false
	}
	if claims.UserType != "agency" {
		s.recordAudit(r, models.AuditActor{UserID: claims.UserID, Role: claims.Role}, "auth.rbac.denied", models.AuditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"reason":     "authority_user_required",
			"actualRole": claims.Role,
		})
		utils.WriteError(w, http.StatusForbidden, "authority_user_required", "authority user access is required")
		return models.AgencyUserProfile{}, false
	}
	if !claims.MFA {
		s.recordAudit(r, models.AuditActor{UserID: claims.UserID, AgencyID: claims.AgencyID, Role: claims.Role}, "auth.rbac.denied", models.AuditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"reason":     "mfa_required",
			"actualRole": claims.Role,
		})
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA is required for authority workflows")
		return models.AgencyUserProfile{}, false
	}

	profile, ok := s.store.AgencyProfileByID(claims.UserID)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "user_not_found", "token user no longer exists")
		return models.AgencyUserProfile{}, false
	}
	if !profile.MFAEnabled {
		s.recordAudit(r, utils.AuditActorFromAgency(profile), "auth.rbac.denied", models.AuditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"reason":     "mfa_required",
			"actualRole": profile.Role,
		})
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA is required for authority workflows")
		return models.AgencyUserProfile{}, false
	}
	if !utils.RoleIn(profile.Role, allowedRoles) {
		s.recordAudit(r, utils.AuditActorFromAgency(profile), "auth.rbac.denied", models.AuditTarget{Type: "route", ID: r.URL.Path}, nil, map[string]any{
			"allowedRoles": allowedRoles,
			"actualRole":   profile.Role,
		})
		utils.WriteError(w, http.StatusForbidden, "forbidden", "role is not allowed to perform this action")
		return models.AgencyUserProfile{}, false
	}

	return profile, true
}
