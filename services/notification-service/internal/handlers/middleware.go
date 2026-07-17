package handlers

import (
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/client"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// deliveryRoles are authority roles allowed to trigger alert deliveries
// (generic channels and voice) to citizens.
var deliveryRoles = map[string]bool{
	"system_admin":     true,
	"agency_admin":     true,
	"nadmo_officer":    true,
	"district_officer": true,
	"dispatcher":       true,
}

// cellBroadcastReviewRoles are authority roles allowed to approve or reject a
// mass cell broadcast.
var cellBroadcastReviewRoles = map[string]bool{
	"system_admin":  true,
	"agency_admin":  true,
	"nadmo_officer": true,
}

// withMiddleware applies CORS and security headers to a handler and carries the
// caller's Authorization header in the request context so outbound service
// calls can forward it.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return utils.WithCORS(s.config.AllowedOrigins, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authorization := strings.TrimSpace(r.Header.Get("Authorization")); authorization != "" {
			r = r.WithContext(client.WithAuthorization(r.Context(), authorization))
		}
		next.ServeHTTP(w, r)
	}))
}

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

// requireAuthority gates an authority-only endpoint: a verified actor with
// completed MFA and an allowed role is required.
func (s *Server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	ctx, ok := s.authorityContext(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "a valid authority bearer token is required")
		return models.AuthorityContext{}, false
	}
	if !ctx.MFACompleted {
		utils.WriteError(w, http.StatusForbidden, "mfa_required", "MFA must be completed for notification delivery actions")
		return models.AuthorityContext{}, false
	}
	if !allowedRoles[ctx.ActorRole] {
		utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this notification delivery action")
		return models.AuthorityContext{}, false
	}

	return ctx, true
}

// webhookSecretFor returns the configured shared secret for an inbound webhook
// channel, if any.
func (s *Server) webhookSecretFor(channel string) string {
	switch channel {
	case "sms":
		return s.config.WebhookSecrets.SMS
	case "ussd":
		return s.config.WebhookSecrets.USSD
	case "whatsapp":
		return s.config.WebhookSecrets.WhatsApp
	case "voice":
		return s.config.WebhookSecrets.Voice
	default:
		return ""
	}
}

// requireWebhookSecret gates an inbound provider webhook when a shared secret
// is configured for the channel. With no secret configured the webhook stays
// open (local development default; a WARN is logged at startup).
func (s *Server) requireWebhookSecret(w http.ResponseWriter, r *http.Request, channel string) bool {
	secret := s.webhookSecretFor(channel)
	if secret == "" {
		return true
	}
	presented := strings.TrimSpace(r.Header.Get("X-NADAA-Webhook-Secret"))
	if presented == "" || !webhookSecretMatches(secret, presented) {
		utils.LogWarn(channel+" webhook rejected", "code", "invalid_webhook_secret")
		utils.WriteError(w, http.StatusUnauthorized, "invalid_webhook_secret", "a valid X-NADAA-Webhook-Secret header is required")
		return false
	}
	return true
}
