package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
)

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
		LogError("write json response failed", "status", status, "error", err)
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
		if allowedOrigins[origin] || (isDevelopmentEnv() && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:"))) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
}

// isDevelopmentEnv reports whether the service runs in a development
// environment; the localhost CORS exception is only honored there.
func isDevelopmentEnv() bool {
	return NormalizeQueryValue(os.Getenv("NADAA_ENV")) == "development"
}

// EnvOrDefault returns the value of key or fallback if unset.
func EnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// EnvBool parses a boolean from the environment.
func EnvBool(key string, fallback bool) bool {
	value := NormalizeQueryValue(os.Getenv(key))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on", "enabled", "mock":
		return true
	case "0", "false", "no", "off", "disabled":
		return false
	default:
		return fallback
	}
}

// AllowedOriginsFromEnv parses the NADAA_ALLOWED_ORIGINS environment variable.
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

// NormalizeQueryValue trims, lowercases, replaces spaces with underscores, and trims underscores.
func NormalizeQueryValue(value string) string {
	return strings.Trim(strings.ToLower(strings.ReplaceAll(value, " ", "_")), "_")
}

// NormalizeID trims whitespace from an identifier.
func NormalizeID(value string) string {
	return strings.TrimSpace(value)
}

// NormalizeLanguage normalizes a language query parameter.
func NormalizeLanguage(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "en"
	}
	return value
}

// ContainsString reports whether values contains expected.
func ContainsString(values []string, expected string) bool {
	return slices.Contains(values, expected)
}

// ProviderName returns a stable provider identifier for logging.
func ProviderName(provider models.NotificationProvider) string {
	if provider == nil {
		return "missing"
	}
	switch provider := provider.(type) {
	case models.MockProvider:
		return "mock_" + provider.Channel
	case models.DisabledProvider:
		return provider.Channel + "_disabled"
	case models.ArkeselSMSProvider:
		return "arkesel_sms"
	case models.ExpoPushProvider:
		return "expo_push"
	default:
		return fmt.Sprintf("%T", provider)
	}
}

// RecipientRef returns a privacy-safe reference for a delivery request and channel.
func RecipientRef(request models.DeliveryRequest, channel string) string {
	if request.RecipientID != "" {
		return request.RecipientID
	}
	if channel == "sms" && request.Phone != "" {
		if len(request.Phone) <= 4 {
			return "phone:" + request.Phone
		}
		return "phone:..." + request.Phone[len(request.Phone)-4:]
	}
	if channel == "push" && request.PushToken != "" {
		if len(request.PushToken) <= 6 {
			return "push:" + request.PushToken
		}
		return "push:..." + request.PushToken[len(request.PushToken)-6:]
	}
	return "anonymous"
}

// PreferredRecipientChannel chooses the preferred channel for logging a recipient.
func PreferredRecipientChannel(channels []string) string {
	if slices.Contains(channels, "sms") {
		return "sms"
	}
	if len(channels) > 0 {
		return channels[0]
	}
	return ""
}

// VoiceRecipientRef returns a privacy-safe reference for a voice recipient.
func VoiceRecipientRef(recipient models.VoiceRecipient) string {
	if recipient.RecipientID != "" {
		return recipient.RecipientID
	}
	if recipient.Phone != "" {
		if len(recipient.Phone) <= 4 {
			return "phone:" + recipient.Phone
		}
		return "phone:..." + recipient.Phone[len(recipient.Phone)-4:]
	}
	return "anonymous"
}

// PhoneRef returns a privacy-safe phone reference.
func PhoneRef(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "phone:unknown"
	}
	if len(phone) <= 4 {
		return "phone:" + phone
	}
	return "phone:..." + phone[len(phone)-4:]
}

// FirstToken returns the first whitespace-separated token in body.
func FirstToken(body string) string {
	fields := strings.Fields(body)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// SMSCommandName returns the uppercased first token of an SMS-style body.
func SMSCommandName(body string) string {
	parts := strings.Fields(strings.TrimSpace(body))
	if len(parts) == 0 {
		return ""
	}
	return strings.ToUpper(parts[0])
}

// NormalizeSMSHazard maps common SMS hazard names to canonical NADAA hazard types.
func NormalizeSMSHazard(value string) string {
	value = NormalizeQueryValue(value)
	switch value {
	case "road", "crash", "accident":
		return "road_crash"
	case "medical", "ambulance":
		return "medical_emergency"
	case "security":
		return "security_incident"
	default:
		return value
	}
}

// NormalizeSMSUrgency maps common SMS urgency names to canonical urgency values.
func NormalizeSMSUrgency(value string) string {
	value = NormalizeQueryValue(value)
	switch value {
	case "life", "life_threatening", "emergency":
		return "life_threatening"
	case "high", "moderate", "low":
		return value
	default:
		return ""
	}
}

// ProviderOrDefault returns provider if non-empty, otherwise fallback.
func ProviderOrDefault(provider string, fallback string) string {
	provider = NormalizeQueryValue(provider)
	if provider == "" {
		return fallback
	}
	return provider
}

// ProfileIDForLog returns profileID only when linked.
func ProfileIDForLog(profileID string, linked bool) string {
	if linked {
		return profileID
	}
	return ""
}

// NormalizeWhatsAppMedia cleans and filters WhatsApp media attachments.
func NormalizeWhatsAppMedia(media []models.WhatsAppMedia) []models.WhatsAppMedia {
	result := make([]models.WhatsAppMedia, 0, len(media))
	for _, item := range media {
		item.ID = strings.TrimSpace(item.ID)
		item.URL = strings.TrimSpace(item.URL)
		item.ContentType = strings.TrimSpace(item.ContentType)
		item.Caption = strings.TrimSpace(item.Caption)
		if item.ID == "" && item.URL == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}

// WhatsAppMediaRefs returns media references suitable for incident handoff.
func WhatsAppMediaRefs(media []models.WhatsAppMedia) []string {
	refs := make([]string, 0, len(media))
	for _, item := range media {
		switch {
		case item.ID != "":
			refs = append(refs, item.ID)
		case item.URL != "":
			refs = append(refs, item.URL)
		}
	}
	return refs
}

// WhatsAppConversationKey returns the lookup key for a WhatsApp conversation.
// It keys on the full E.164 phone number so users sharing a last-4 suffix never
// share a conversation state machine; the masked phone ref is for logs only.
func WhatsAppConversationKey(phone string, _ string, _ bool) string {
	return strings.TrimSpace(phone)
}

// IsWhatsAppTopLevelCommand reports whether command is a recognized top-level command.
func IsWhatsAppTopLevelCommand(command string) bool {
	switch command {
	case "ALERT", "ALERTS", "RISK", "GUIDE", "GUIDES", "SHELTER", "SHELTERS", "HELP", "112", "REPORT", "CANCEL", "MENU", "START", "HI", "HELLO":
		return true
	default:
		return false
	}
}

// WhatsAppCommandArg returns the whitespace-separated token at index.
func WhatsAppCommandArg(body string, index int) string {
	fields := strings.Fields(body)
	if index < 0 || index >= len(fields) {
		return ""
	}
	return fields[index]
}

// WhatsAppRetentionUntil returns the privacy retention timestamp.
func WhatsAppRetentionUntil(now time.Time) time.Time {
	return now.Add(90 * 24 * time.Hour)
}

// WhatsAppMessageSummary returns a privacy-safe message summary.
func WhatsAppMessageSummary(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	command := SMSCommandName(body)
	if command == "" {
		return fmt.Sprintf("provided:%d_chars", len(body))
	}
	return fmt.Sprintf("command:%s:%d_chars", command, len(body))
}

// WhatsAppMediaSummary returns a privacy-safe media summary.
func WhatsAppMediaSummary(media []models.WhatsAppMedia) string {
	if len(media) == 0 {
		return ""
	}
	return fmt.Sprintf("attachments:%d", len(media))
}

// InclusiveLocation resolves location from coordinates or free-text tokens. It
// returns nil coordinates when the caller did not share a usable location so the
// incident handoff omits the field instead of plotting a fabricated point;
// (0,0) is treated as no-location because it is the default zero value, not a
// real caller fix.
func InclusiveLocation(location *models.Coordinates, locationTokens []string) (*models.Coordinates, string) {
	if location != nil && location.Lat >= -90 && location.Lat <= 90 && location.Lng >= -180 && location.Lng <= 180 && (location.Lat != 0 || location.Lng != 0) {
		resolved := *location
		return &resolved, "caller shared approximate coordinates"
	}
	locationLabel := strings.TrimSpace(strings.Join(locationTokens, " "))
	if locationLabel == "" {
		locationLabel = "caller did not provide location details; use phone follow-up"
	}
	return nil, locationLabel
}

// LogTextSummary returns a privacy-safe text length summary.
func LogTextSummary(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("provided:%d_chars", len(value))
}

// LogInfo logs an INFO level structured message.
func LogInfo(message string, fields ...any) {
	logWithLevel("INFO", message, fields...)
}

// LogWarn logs a WARN level structured message.
func LogWarn(message string, fields ...any) {
	logWithLevel("WARN", message, fields...)
}

// LogError logs an ERROR level structured message.
func LogError(message string, fields ...any) {
	logWithLevel("ERROR", message, fields...)
}

func logWithLevel(level string, message string, fields ...any) {
	pairs := make([]string, 0, (len(fields)+1)/2)
	for index := 0; index < len(fields); index += 2 {
		key := strings.TrimSpace(fmt.Sprint(fields[index]))
		if key == "" {
			key = fmt.Sprintf("field_%d", index/2)
		}
		key = strings.ReplaceAll(key, " ", "_")

		value := "missing_value"
		if index+1 < len(fields) {
			value = fmt.Sprint(fields[index+1])
		}
		value = strings.ReplaceAll(value, "\n", " ")
		value = strings.ReplaceAll(value, "\r", " ")
		pairs = append(pairs, fmt.Sprintf("%s=%q", key, value))
	}

	if len(pairs) == 0 {
		log.Printf("level=%s msg=%q", level, message)
		return
	}
	log.Printf("level=%s msg=%q %s", level, message, strings.Join(pairs, " "))
}
