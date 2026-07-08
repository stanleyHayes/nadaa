package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

func requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     utils.NormalizeQueryValue(r.Header.Get("X-NADAA-Actor-Role")),
		MFACompleted:  utils.NormalizeQueryValue(r.Header.Get("X-NADAA-MFA-Completed")) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
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

func hasAuthorityHeaders(r *http.Request) bool {
	return strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")) != "" &&
		strings.TrimSpace(r.Header.Get("X-NADAA-Actor-Role")) != "" &&
		strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")) != ""
}
