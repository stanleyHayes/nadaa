package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

// authorityContext builds the actor context from a verified bearer token. When
// no bearer token is present, legacy X-NADAA-Actor-* headers are honored only
// if mock actors are allowed (local development and smoke tests).
func (s *Server) authorityContext(r *http.Request) (models.AuthorityContext, bool) {
	if token := bearerToken(r); token != "" {
		claims, err := verifyToken(token, []byte(s.config.TokenSecret), s.now())
		if err != nil {
			return models.AuthorityContext{}, false
		}
		return models.AuthorityContext{
			ActorUserID:   claims.UserID,
			ActorAgencyID: claims.AgencyID,
			ActorRole:     utils.NormalizeQueryValue(claims.Role),
			ActorDistrict: claims.District,
			MFACompleted:  claims.MFA,
			RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
		}, true
	}

	if !s.config.AllowMockActors {
		return models.AuthorityContext{}, false
	}
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     utils.NormalizeQueryValue(r.Header.Get("X-NADAA-Actor-Role")),
		MFACompleted:  utils.NormalizeQueryValue(r.Header.Get("X-NADAA-MFA-Completed")) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}
	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx, ok := s.authorityContext(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid authority bearer token is required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for alert workflow actions")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this alert workflow action")
		return models.AuthorityContext{}, false
	}

	return ctx, true
}
