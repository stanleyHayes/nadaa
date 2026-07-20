package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

// serviceTokenHeader carries the shared service-to-service credential
// (NADAA_INTERNAL_SERVICE_TOKEN) on internal calls.
//
//nolint:gosec // G101: header name constant, not a credential.
const serviceTokenHeader = "X-NADAA-Service-Token"

// allAgencyRoles matches every authority role, for endpoints open to any
// authenticated agency user rather than a restricted role set.
var allAgencyRoles = []string{
	models.RoleAgencyViewer,
	models.RoleDispatcher,
	models.RoleResponder,
	models.RoleNADMOOfficer,
	models.RoleDistrictOfficer,
	models.RoleAgencyAdmin,
	models.RoleSystemAdmin,
}

// agencyProfileFromMockHeaders accepts the shared X-NADAA-* actor headers used
// by the rest of the platform's services (see each service's requireAuthority).
// It lets the demo dashboards reach auth-service governance endpoints without a
// real signed session, mirroring the mock-auth scheme the other 17 services
// already trust. Returns false when the headers are absent so callers fall back
// to real Bearer-token verification.
func agencyProfileFromMockHeaders(r *http.Request, allowedRoles []string) (models.AgencyUserProfile, bool) {
	actorID := strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID"))
	agencyID := strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID"))
	role := strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role")))
	mfaCompleted := strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true"

	// No actor headers at all -> not a mock request; let the token path run.
	if actorID == "" && agencyID == "" && role == "" {
		return models.AgencyUserProfile{}, false
	}
	if actorID == "" || agencyID == "" || role == "" || !mfaCompleted {
		return models.AgencyUserProfile{}, false
	}
	if !utils.RoleIn(role, allowedRoles) {
		return models.AgencyUserProfile{}, false
	}

	return models.AgencyUserProfile{
		ID:          actorID,
		Role:        role,
		Agency:      models.AgencySummary{ID: agencyID},
		MFARequired: true,
		MFAEnabled:  true,
	}, true
}

func (s *Server) requireAgencyRole(w http.ResponseWriter, r *http.Request, allowedRoles ...string) (models.AgencyUserProfile, bool) {
	// Demo mock-auth parity (opt-in via NADAA_AUTH_ALLOW_MOCK_ACTORS): when the
	// shared X-NADAA-* actor headers are present and authorize the request,
	// honor them like the other services do. This trusts client-supplied role
	// headers, so it is disabled by default and must stay off in production;
	// real token verification below is the only path when the flag is unset.
	if s.config.AllowMockActorHeaders {
		if profile, ok := agencyProfileFromMockHeaders(r, allowedRoles); ok {
			return profile, true
		}
	}

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
