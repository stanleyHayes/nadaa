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

// bearerToken extracts the Bearer token from the Authorization header.
func bearerToken(r *http.Request) (string, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	return token, token != ""
}

// parseAuthority resolves the authority context from a verified bearer token
// issued by auth-service. The legacy X-NADAA-* actor headers are honored only
// when NADAA_AUTH_ALLOW_MOCK_ACTORS=true (local dev and smoke tests); otherwise
// they are ignored entirely. Returns false when no valid credentials exist so
// public endpoints keep working while authority endpoints reject with 401.
func (s *Server) parseAuthority(r *http.Request) (models.AuthorityContext, bool) {
	requestID := strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID"))

	if token, ok := bearerToken(r); ok {
		claims, valid := utils.VerifyToken(token, []byte(s.config.TokenSecret), s.now())
		if !valid {
			return models.AuthorityContext{}, false
		}
		ctx := models.AuthorityContext{
			ActorUserID:   strings.TrimSpace(claims.UserID),
			ActorAgencyID: strings.TrimSpace(claims.AgencyID),
			ActorRole:     utils.NormalizeQueryValue(claims.Role),
			ActorDistrict: strings.TrimSpace(claims.District),
			MFACompleted:  claims.MFA,
			RequestID:     requestID,
		}
		if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
			return models.AuthorityContext{}, false
		}
		return ctx, true
	}

	if !s.config.AllowMockActorHeaders {
		return models.AuthorityContext{}, false
	}

	ctx := models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")),
		ActorAgencyID: strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")),
		ActorRole:     utils.NormalizeQueryValue(r.Header.Get("X-NADAA-Actor-Role")),
		MFACompleted:  utils.NormalizeQueryValue(r.Header.Get("X-NADAA-MFA-Completed")) == "true",
		RequestID:     requestID,
	}
	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// requireAuthority enforces that the request includes a valid authority context.
func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	ctx, ok := s.parseAuthority(r)
	if !ok {
		if _, hasToken := bearerToken(r); hasToken {
			utils.WriteError(w, http.StatusUnauthorized, "invalid_token", "bearer token is invalid or expired")
			return models.AuthorityContext{}, false
		}
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "an authority bearer token is required")
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
