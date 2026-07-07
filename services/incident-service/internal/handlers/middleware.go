package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

var (
	statusWorkflowRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"responder":        true,
	}
	verificationRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	incidentAuditRoles = map[string]bool{
		"system_admin":  true,
		"agency_admin":  true,
		"nadmo_officer": true,
	}
	assignmentRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	mergeRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	abuseReviewRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	incidentReadRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"responder":        true,
		"agency_viewer":    true,
	}
	reporterContactRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
)

func requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-Actor-Role"))),
		MFACompleted:  strings.TrimSpace(strings.ToLower(r.Header.Get("X-NADAA-MFA-Completed"))) == "true",
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}

	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for incident workflow actions")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this incident workflow action")
		return models.AuthorityContext{}, false
	}

	return ctx, true
}

func statusForCode(code string) int {
	switch code {
	case "not_found":
		return http.StatusNotFound
	case "forbidden":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}
