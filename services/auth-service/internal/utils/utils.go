package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
)

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// OTPGenerator produces one-time codes.
type OTPGenerator interface {
	Generate() (string, error)
}

// RandomOTPGenerator generates cryptographically random 6-digit codes.
type RandomOTPGenerator struct{}

// Generate returns a random six-digit OTP.
func (RandomOTPGenerator) Generate() (string, error) {
	upperBound := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, upperBound)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// FixedOTPGenerator always returns a configured code, falling back to 123456.
type FixedOTPGenerator struct {
	Code string
}

// Generate returns the fixed code.
func (f FixedOTPGenerator) Generate() (string, error) {
	if f.Code == "" {
		return "123456", nil
	}
	return f.Code, nil
}

// DecodeJSON decodes a JSON request body into target.
func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write json response: %v", err)
	}
}

// WriteError writes a structured API error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, models.APIError{Error: models.APIErrorBody{Code: code, Message: message}})
}

// WithCORS wraps a handler with security and CORS headers.
func WithCORS(allowedOrigins map[string]bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		applySecurityHeaders(w)
		applyCORSHeaders(w, r, allowedOrigins)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func applySecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	w.Header().Set("Cache-Control", "no-store")
}

func applyCORSHeaders(w http.ResponseWriter, r *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if len(allowedOrigins) == 0 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Add("Vary", "Origin")
		if allowedOrigins[origin] || strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

// AllowedOriginsFromEnv parses NADAA_ALLOWED_ORIGINS into a lookup set.
func AllowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for origin := range strings.SplitSeq(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}

// EnvOrDefault returns the value of key or fallback if unset.
func EnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// NormalizePhone strips spaces and dashes from a phone number.
func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	return phone
}

// NormalizeEmail lowercases and trims an email address.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// NormalizeLanguage lowercases and trims a language tag, defaulting to en.
func NormalizeLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return "en"
	}
	return language
}

// NormalizeRole lowercases and trims a role value.
func NormalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

// ValidPhone reports whether phone is in E.164 format.
func ValidPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

// ValidEmail performs a simple syntactic email check.
func ValidEmail(email string) bool {
	parts := strings.Split(email, "@")
	return len(parts) == 2 && parts[0] != "" && strings.Contains(parts[1], ".")
}

// ValidCoordinates reports whether location is within valid lat/lng ranges.
func ValidCoordinates(location models.Coordinates) bool {
	return location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180
}

// ParseAuditLimit parses and clamps the audit log limit query parameter.
func ParseAuditLimit(raw string) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 || limit > 100 {
		return 50
	}
	return limit
}

// ValidAgencyRole reports whether role is an authority role.
func ValidAgencyRole(role string) bool {
	return models.AgencyRoles[role]
}

// RoleIn reports whether role is one of allowedRoles.
func RoleIn(role string, allowedRoles []string) bool {
	return slices.Contains(allowedRoles, role)
}

// BearerToken extracts the token from an Authorization header.
func BearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return "", false
	}
	return token, true
}

// AuditContextFromRequest extracts request metadata for audit logging.
func AuditContextFromRequest(r *http.Request) models.AuditRequestContext {
	if r == nil {
		return models.AuditRequestContext{RequestID: NewID("req")}
	}

	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = NewID("req")
	}

	return models.AuditRequestContext{
		RequestID: requestID,
		IPAddress: RequestIPAddress(r),
		UserAgent: strings.TrimSpace(r.UserAgent()),
	}
}

// RequestIPAddress extracts the client IP from proxy headers or RemoteAddr.
func RequestIPAddress(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}
	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

// AuditActorFromCitizen builds an audit actor from a citizen profile.
func AuditActorFromCitizen(profile models.CitizenProfile) models.AuditActor {
	return models.AuditActor{UserID: profile.ID, Role: profile.Role}
}

// AuditActorFromAgency builds an audit actor from an agency user profile.
func AuditActorFromAgency(profile models.AgencyUserProfile) models.AuditActor {
	return models.AuditActor{UserID: profile.ID, AgencyID: profile.Agency.ID, Role: profile.Role}
}

// CitizenAuditSnapshot builds a sanitized audit snapshot for a citizen.
func CitizenAuditSnapshot(profile models.CitizenProfile) map[string]any {
	return map[string]any{
		"id":                profile.ID,
		"phone":             profile.Phone,
		"role":              profile.Role,
		"preferredLanguage": profile.PreferredLanguage,
		"contactPermission": profile.ContactPermission,
	}
}

// AgencyUserAuditSnapshot builds a sanitized audit snapshot for an agency user.
func AgencyUserAuditSnapshot(profile models.AgencyUserProfile) map[string]any {
	return map[string]any{
		"id":          profile.ID,
		"name":        profile.Name,
		"email":       profile.Email,
		"phone":       profile.Phone,
		"role":        profile.Role,
		"agencyId":    profile.Agency.ID,
		"mfaRequired": profile.MFARequired,
		"mfaEnabled":  profile.MFAEnabled,
	}
}

// AgencySummaryFromRecord builds a summary from an agency record.
func AgencySummaryFromRecord(agency models.AgencyRecord) models.AgencySummary {
	return models.AgencySummary{
		ID:            agency.ID,
		Name:          agency.Name,
		Type:          agency.Type,
		Region:        agency.Region,
		District:      agency.District,
		ContactNumber: agency.ContactNumber,
	}
}

// AgencyProfileFromUser builds the public profile for an agency user.
func AgencyProfileFromUser(user models.AgencyUser, agency models.AgencyRecord) models.AgencyUserProfile {
	return models.AgencyUserProfile{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Phone:       user.Phone,
		Role:        user.Role,
		Agency:      AgencySummaryFromRecord(agency),
		MFARequired: user.MFARequired,
		MFAEnabled:  user.MFAEnabled,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// HashCredential returns a base64-url SHA-256 hash of a credential.
func HashCredential(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// NewTemporaryPassword returns a new random temporary password.
func NewTemporaryPassword() string {
	return NewID("tmp")
}

// NewMFASecret returns a new random MFA secret identifier.
func NewMFASecret() string {
	return NewID("mfa_secret")
}

// NewID returns a random identifier with the given prefix.
func NewID(prefix string) string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, bytes)
}
