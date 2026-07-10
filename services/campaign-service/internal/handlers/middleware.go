package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/utils"
)

// withMiddleware applies server-wide middleware to the handler.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, next)
}

var allowedCampaignUpdateRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

var allowedHazards = map[string]bool{
	"flood":             true,
	"fire":              true,
	"road_crash":        true,
	"building_collapse": true,
	"medical_emergency": true,
	"security_incident": true,
	"disease_outbreak":  true,
	"electrical_hazard": true,
	"blocked_drain":     true,
	"landslide":         true,
	"marine_accident":   true,
	"storm":             true,
	"tidal_wave":        true,
	"other":             true,
}

var allowedCampaignStatuses = map[string]bool{
	"draft":     true,
	"published": true,
	"archived":  true,
}

var allowedContentBlockTypes = map[string]bool{
	"article":   true,
	"checklist": true,
	"media":     true,
}

// parseAuthority extracts authority context from headers without failing.
func parseAuthority(r *http.Request) (models.AuthorityContext, bool) {
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

// requireAuthority enforces that the request includes a valid authority context.
func requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx, ok := parseAuthority(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "authority actor id, role, and agency id headers are required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for campaign updates")
		return models.AuthorityContext{}, false
	}
	if !allowedCampaignUpdateRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed to manage campaigns")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}
