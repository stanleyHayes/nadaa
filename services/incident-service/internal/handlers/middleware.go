package handlers

import (
	"crypto/hmac"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/incident-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/incident-service/internal/utils"
)

// serviceTokenHeader carries the shared internal token for service-to-service calls.
//
//nolint:gosec // G101: header name constant, not a credential.
const serviceTokenHeader = "X-NADAA-Service-Token"

// internalServiceRole is the read-only actor role granted to verified
// service-to-service calls. It is deliberately present only in
// incidentReadRoles — never in reporterContactRoles or any workflow role map.
const internalServiceRole = "internal_service"

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
		// Read-only service-to-service context; absent from
		// reporterContactRoles and every workflow role map by design.
		internalServiceRole: true,
	}
	triageReviewRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	reporterContactRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
	}
	// volunteerTaskWorkflowRoles gates authority actors on volunteer-task
	// mutations; read-only roles (agency_viewer, internal_service) are excluded.
	volunteerTaskWorkflowRoles = map[string]bool{
		"system_admin":     true,
		"agency_admin":     true,
		"nadmo_officer":    true,
		"district_officer": true,
		"dispatcher":       true,
		"responder":        true,
	}
)

func (s *server) requireAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
	// A verified internal service token yields a read-only authority context;
	// its role passes only the read role map, so workflow endpoints stay closed.
	if ctx, ok := s.serviceTokenContext(r); ok {
		if !allowedRoles[ctx.ActorRole] {
			utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for this incident workflow action")
			return models.AuthorityContext{}, false
		}
		return ctx, true
	}

	// Legacy X-NADAA-* actor headers are honored only when mock actors are
	// explicitly enabled (local development and smoke tests).
	if s.allowMockActors && hasMockActorHeaders(r) {
		return s.requireMockAuthority(w, r, allowedRoles)
	}

	ctx, ok := s.authorityContextFromToken(w, r)
	if !ok {
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

// serviceTokenContext authenticates X-NADAA-Service-Token as a read-only
// service-to-service credential using a constant-time compare. Incident data
// is sensitive, so unlike ml-service the path stays closed when
// NADAA_INTERNAL_SERVICE_TOKEN is unset: the header is ignored entirely and
// the caller falls through to the bearer-token path.
func (s *server) serviceTokenContext(r *http.Request) (models.AuthorityContext, bool) {
	if s.internalServiceToken == "" {
		return models.AuthorityContext{}, false
	}
	if !hmac.Equal([]byte(r.Header.Get(serviceTokenHeader)), []byte(s.internalServiceToken)) {
		return models.AuthorityContext{}, false
	}
	return models.AuthorityContext{
		ActorUserID:  internalServiceRole,
		ActorRole:    internalServiceRole,
		MFACompleted: true,
		RequestID:    strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}, true
}

// requireVolunteerActor authenticates volunteer task endpoints: either a
// verified authority user whose role is in allowedAgencyRoles, or the citizen
// whose token subject matches the volunteer profile's registered user id.
func (s *server) requireVolunteerActor(w http.ResponseWriter, r *http.Request, citizenUserID string, allowedAgencyRoles map[string]bool) (models.AuthorityContext, bool) {
	if s.allowMockActors && hasMockActorHeaders(r) {
		return s.requireMockAuthority(w, r, allowedAgencyRoles)
	}

	token, ok := bearerToken(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return models.AuthorityContext{}, false
	}
	claims, err := verifyAuthToken(s.tokenSecret, s.now, token)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return models.AuthorityContext{}, false
	}

	if claims.UserType == "agency" {
		ctx := authorityContextFromClaims(claims, r)
		if ctx.ActorUserID == "" || ctx.ActorRole == "" {
			utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "token must carry an authority user id and role")
			return models.AuthorityContext{}, false
		}
		if !allowedAgencyRoles[ctx.ActorRole] {
			utils.WriteError(w, http.StatusForbidden, "forbidden", "actor role is not allowed for volunteer task workflow actions")
			return models.AuthorityContext{}, false
		}
		return ctx, true
	}

	if citizenUserID != "" && strings.TrimSpace(claims.UserID) == citizenUserID {
		return models.AuthorityContext{
			ActorUserID:  strings.TrimSpace(claims.UserID),
			ActorRole:    "citizen",
			MFACompleted: claims.MFA,
			RequestID:    strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
		}, true
	}

	utils.WriteError(w, http.StatusForbidden, "forbidden", "only the owning volunteer or an authority user can access volunteer tasks")
	return models.AuthorityContext{}, false
}

// requireMockAuthority preserves the legacy header-based actor context for
// local development when NADAA_AUTH_ALLOW_MOCK_ACTORS=true.
func (s *server) requireMockAuthority(w http.ResponseWriter, r *http.Request, allowedRoles map[string]bool) (models.AuthorityContext, bool) {
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

func (s *server) authorityContextFromToken(w http.ResponseWriter, r *http.Request) (models.AuthorityContext, bool) {
	token, ok := bearerToken(r)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "missing_token", "Bearer token is required")
		return models.AuthorityContext{}, false
	}
	claims, err := verifyAuthToken(s.tokenSecret, s.now, token)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return models.AuthorityContext{}, false
	}

	ctx := authorityContextFromClaims(claims, r)
	if ctx.ActorUserID == "" || ctx.ActorAgencyID == "" || ctx.ActorRole == "" {
		utils.WriteError(w, http.StatusUnauthorized, "missing_authority_context", "token must carry an authority user id, role, and agency id")
		return models.AuthorityContext{}, false
	}
	return ctx, true
}

// authorityContextFromClaims builds the actor context from verified token
// claims only; client-supplied actor headers are never consulted here.
func authorityContextFromClaims(claims tokenClaims, r *http.Request) models.AuthorityContext {
	return models.AuthorityContext{
		ActorUserID:   strings.TrimSpace(claims.UserID),
		ActorAgencyID: strings.TrimSpace(claims.AgencyID),
		ActorRole:     strings.TrimSpace(strings.ToLower(claims.Role)),
		ActorDistrict: strings.TrimSpace(claims.District),
		MFACompleted:  claims.MFA,
		RequestID:     strings.TrimSpace(r.Header.Get("X-NADAA-Request-ID")),
	}
}

// hasMockActorHeaders reports whether any legacy actor header is present.
func hasMockActorHeaders(r *http.Request) bool {
	return strings.TrimSpace(r.Header.Get("X-NADAA-Actor-ID")) != "" ||
		strings.TrimSpace(r.Header.Get("X-NADAA-Actor-Role")) != "" ||
		strings.TrimSpace(r.Header.Get("X-NADAA-Agency-ID")) != ""
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
