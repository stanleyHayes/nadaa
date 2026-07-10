package models

import (
	"context"
	"fmt"
	"time"
)

// CitizenAlert is a public-facing alert in the notification feed.
type CitizenAlert struct {
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	HazardType         string      `json:"hazardType"`
	Severity           string      `json:"severity"`
	Message            string      `json:"message"`
	Target             AlertTarget `json:"target"`
	TargetLabel        string      `json:"targetLabel"`
	StartsAt           time.Time   `json:"startsAt"`
	ExpiresAt          time.Time   `json:"expiresAt"`
	Status             string      `json:"status"`
	RecommendedAction  string      `json:"recommendedAction"`
	EvacuationRequired bool        `json:"evacuationRequired"`
	ShelterIDs         []string    `json:"shelterIds"`
	Source             string      `json:"source"`
	UpdatedAt          time.Time   `json:"updatedAt"`
}

// AuthorityAlert is the upstream alert representation from the alert service.
type AuthorityAlert struct {
	ID                 string      `json:"id"`
	Title              string      `json:"title"`
	HazardType         string      `json:"hazardType"`
	Severity           string      `json:"severity"`
	Message            string      `json:"message"`
	Target             AlertTarget `json:"target"`
	StartsAt           time.Time   `json:"startsAt"`
	ExpiresAt          time.Time   `json:"expiresAt"`
	RecommendedAction  string      `json:"recommendedAction"`
	EvacuationRequired bool        `json:"evacuationRequired"`
	ShelterIDs         []string    `json:"shelterIds"`
	Status             string      `json:"status"`
	UpdatedAt          time.Time   `json:"updatedAt"`
}

// AlertTarget describes who or where an alert applies to.
type AlertTarget struct {
	Type                string       `json:"type"`
	IDs                 []string     `json:"ids"`
	Label               string       `json:"label"`
	Center              *Coordinates `json:"center,omitempty"`
	RadiusMeters        float64      `json:"radiusMeters,omitempty"`
	AreaSqKm            float64      `json:"areaSqKm,omitempty"`
	EstimatedPopulation int          `json:"estimatedPopulation,omitempty"`
}

// Coordinates is a latitude/longitude pair.
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CitizenAlertListResponse is the payload returned when listing citizen alerts.
type CitizenAlertListResponse struct {
	Alerts      []CitizenAlert `json:"alerts"`
	GeneratedAt time.Time      `json:"generatedAt"`
	Source      string         `json:"source"`
}

// AuthorityAlertListResponse is the payload returned by the upstream alert service.
type AuthorityAlertListResponse struct {
	Alerts []AuthorityAlert `json:"alerts"`
}

// DeliveryRequest requests delivery of an alert over one or more channels.
type DeliveryRequest struct {
	AlertID     string   `json:"alertId,omitempty"`
	RecipientID string   `json:"recipientId"`
	Phone       string   `json:"phone,omitempty"`
	PushToken   string   `json:"pushToken,omitempty"`
	Language    string   `json:"language,omitempty"`
	Channels    []string `json:"channels"`
	DryRun      bool     `json:"dryRun,omitempty"`
}

// DeliveryResponse returns the delivery attempts created for an alert.
type DeliveryResponse struct {
	Attempts []DeliveryAttempt `json:"attempts"`
}

// DeliveryAttempt records a single delivery attempt.
type DeliveryAttempt struct {
	ID           string    `json:"id"`
	AlertID      string    `json:"alertId"`
	AlertTitle   string    `json:"alertTitle"`
	Channel      string    `json:"channel"`
	Provider     string    `json:"provider"`
	RecipientRef string    `json:"recipientRef"`
	Status       string    `json:"status"`
	Reason       string    `json:"reason,omitempty"`
	MessageID    string    `json:"messageId,omitempty"`
	VoiceAssetID string    `json:"voiceAssetId,omitempty"`
	Language     string    `json:"language,omitempty"`
	AudioURL     string    `json:"audioUrl,omitempty"`
	AttemptedAt  time.Time `json:"attemptedAt"`
}

// DeliveryLogListResponse returns persisted delivery attempts.
type DeliveryLogListResponse struct {
	Logs []DeliveryAttempt `json:"logs"`
}

// ProviderMessage is passed to a notification provider when sending.
type ProviderMessage struct {
	Alert       CitizenAlert
	Request     DeliveryRequest
	Channel     string
	Recipient   string
	AttemptedAt time.Time
}

// ProviderResult is the outcome of a provider send attempt.
type ProviderResult struct {
	Provider  string
	Status    string
	Reason    string
	MessageID string
}

// NotificationProvider abstracts a push/sms/voice backend.
type NotificationProvider interface {
	Send(ctx context.Context, message ProviderMessage) ProviderResult
}

// MockProvider simulates a successful delivery for a channel.
type MockProvider struct {
	Channel string
}

// Send returns a mock delivered result.
func (p MockProvider) Send(_ context.Context, message ProviderMessage) ProviderResult {
	providerID := "mock_push"
	switch p.Channel {
	case "sms":
		providerID = "mock_sms"
	case "voice":
		providerID = "mock_voice"
	}
	return ProviderResult{
		Provider:  providerID,
		Status:    "delivered",
		MessageID: fmt.Sprintf("%s_%s_%d", providerID, message.Alert.ID, message.AttemptedAt.Unix()),
	}
}

// DisabledProvider returns a skipped result for a disabled channel.
type DisabledProvider struct {
	Channel string
	Reason  string
}

// Send returns a skipped result.
func (p DisabledProvider) Send(_ context.Context, _ ProviderMessage) ProviderResult {
	providerID := p.Channel + "_disabled"
	return ProviderResult{
		Provider: providerID,
		Status:   "skipped",
		Reason:   p.Reason,
	}
}

// AlertFeedFilters captures accepted query parameters for the alert feed.
type AlertFeedFilters struct {
	Hazard         string
	Severity       string
	Status         string
	IncludeExpired bool
	TargetType     string
	TargetID       string
}

// LogFilters captures accepted query parameters for delivery logs.
type LogFilters struct {
	AlertID string
	Channel string
	Status  string
}

// AccessLogFilters captures accepted query parameters for inclusive access logs.
type AccessLogFilters struct {
	Channel string
	Intent  string
	Status  string
}

// USSDWebhookRequest is an inbound USSD message.
type USSDWebhookRequest struct {
	SessionID         string       `json:"sessionId"`
	Phone             string       `json:"phone"`
	ServiceCode       string       `json:"serviceCode,omitempty"`
	Text              string       `json:"text"`
	Language          string       `json:"language,omitempty"`
	Network           string       `json:"network,omitempty"`
	Provider          string       `json:"provider,omitempty"`
	ProviderMessageID string       `json:"providerMessageId,omitempty"`
	ProviderError     string       `json:"providerError,omitempty"`
	ProfileID         string       `json:"profileId,omitempty"`
	LinkProfile       bool         `json:"linkProfile,omitempty"`
	Location          *Coordinates `json:"location,omitempty"`
}

// USSDWebhookResponse is the response to an inbound USSD message.
type USSDWebhookResponse struct {
	SessionID string                 `json:"sessionId"`
	Action    string                 `json:"action"`
	Message   string                 `json:"message"`
	Language  string                 `json:"language"`
	Log       InclusiveAccessLog     `json:"log"`
	Report    *InclusiveAccessReport `json:"report,omitempty"`
}

// SMSInboundRequest is an inbound SMS message.
type SMSInboundRequest struct {
	From              string       `json:"from"`
	Body              string       `json:"body"`
	Language          string       `json:"language,omitempty"`
	Provider          string       `json:"provider,omitempty"`
	ProviderMessageID string       `json:"providerMessageId,omitempty"`
	ProviderError     string       `json:"providerError,omitempty"`
	ProfileID         string       `json:"profileId,omitempty"`
	LinkProfile       bool         `json:"linkProfile,omitempty"`
	Location          *Coordinates `json:"location,omitempty"`
}

// SMSInboundResponse is the response to an inbound SMS message.
type SMSInboundResponse struct {
	Message string                 `json:"message"`
	Log     InclusiveAccessLog     `json:"log"`
	Report  *InclusiveAccessReport `json:"report,omitempty"`
}

// WhatsAppInboundRequest is an inbound WhatsApp message.
type WhatsAppInboundRequest struct {
	From              string          `json:"from"`
	Body              string          `json:"body"`
	Language          string          `json:"language,omitempty"`
	Provider          string          `json:"provider,omitempty"`
	ProviderMessageID string          `json:"providerMessageId,omitempty"`
	ProviderError     string          `json:"providerError,omitempty"`
	ProfileID         string          `json:"profileId,omitempty"`
	LinkProfile       bool            `json:"linkProfile,omitempty"`
	Location          *Coordinates    `json:"location,omitempty"`
	Media             []WhatsAppMedia `json:"media,omitempty"`
}

// WhatsAppMedia describes an attached WhatsApp media item.
type WhatsAppMedia struct {
	ID          string `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Caption     string `json:"caption,omitempty"`
}

// WhatsAppInboundResponse is the response to an inbound WhatsApp message.
type WhatsAppInboundResponse struct {
	Message       string                 `json:"message"`
	Conversation  WhatsAppConversation   `json:"conversation"`
	Log           InclusiveAccessLog     `json:"log"`
	Report        *InclusiveAccessReport `json:"report,omitempty"`
	TranscriptIDs []string               `json:"transcriptIds,omitempty"`
}

// WhatsAppConversation tracks an interactive reporting session.
type WhatsAppConversation struct {
	ID                 string    `json:"id"`
	Key                string    `json:"-"`
	Channel            string    `json:"channel"`
	PhoneRef           string    `json:"phoneRef"`
	ProfileID          string    `json:"profileId,omitempty"`
	LinkedProfile      bool      `json:"linkedProfile"`
	Language           string    `json:"language"`
	Intent             string    `json:"intent"`
	State              string    `json:"state"`
	Hazard             string    `json:"hazard,omitempty"`
	Urgency            string    `json:"urgency,omitempty"`
	LastMessageSummary string    `json:"lastMessageSummary,omitempty"`
	LastMediaSummary   string    `json:"lastMediaSummary,omitempty"`
	StartedAt          time.Time `json:"startedAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	ExpiresAt          time.Time `json:"expiresAt"`
	RetentionUntil     time.Time `json:"retentionUntil"`
}

// WhatsAppTranscript is a privacy-safe summary of a WhatsApp message.
type WhatsAppTranscript struct {
	ID                string    `json:"id"`
	ConversationID    string    `json:"conversationId"`
	Provider          string    `json:"provider"`
	ProviderMessageID string    `json:"providerMessageId,omitempty"`
	PhoneRef          string    `json:"phoneRef"`
	ProfileID         string    `json:"profileId,omitempty"`
	LinkedProfile     bool      `json:"linkedProfile"`
	Direction         string    `json:"direction"`
	Intent            string    `json:"intent"`
	State             string    `json:"state"`
	MessageSummary    string    `json:"messageSummary,omitempty"`
	MediaSummary      string    `json:"mediaSummary,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
	RetentionUntil    time.Time `json:"retentionUntil"`
}

// InclusiveAccessLog records an interaction over an inclusive access channel.
type InclusiveAccessLog struct {
	ID                string    `json:"id"`
	Channel           string    `json:"channel"`
	Provider          string    `json:"provider"`
	ProviderMessageID string    `json:"providerMessageId,omitempty"`
	SessionID         string    `json:"sessionId,omitempty"`
	PhoneRef          string    `json:"phoneRef"`
	ProfileID         string    `json:"profileId,omitempty"`
	LinkedProfile     bool      `json:"linkedProfile"`
	Language          string    `json:"language"`
	Intent            string    `json:"intent"`
	Status            string    `json:"status"`
	ProviderError     string    `json:"providerError,omitempty"`
	IncidentID        string    `json:"incidentId,omitempty"`
	IncidentReference string    `json:"incidentReference,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
}

// InclusiveAccessReport is an emergency report received through an inclusive channel.
type InclusiveAccessReport struct {
	ID                string      `json:"id"`
	Channel           string      `json:"channel"`
	Type              string      `json:"type"`
	Urgency           string      `json:"urgency"`
	Description       string      `json:"description"`
	Location          Coordinates `json:"location"`
	LocationLabel     string      `json:"locationLabel"`
	PhoneRef          string      `json:"phoneRef"`
	ProfileID         string      `json:"profileId,omitempty"`
	LinkedProfile     bool        `json:"linkedProfile"`
	Status            string      `json:"status"`
	Media             []string    `json:"media,omitempty"`
	IncidentID        string      `json:"incidentId,omitempty"`
	IncidentReference string      `json:"incidentReference,omitempty"`
	FailureReason     string      `json:"failureReason,omitempty"`
	CreatedAt         time.Time   `json:"createdAt"`
}

// AccessLogListResponse returns inclusive access logs.
type AccessLogListResponse struct {
	Logs []InclusiveAccessLog `json:"logs"`
}

// VoiceAlertRequest requests generation of a voice alert asset.
type VoiceAlertRequest struct {
	AlertID             string   `json:"alertId"`
	Languages           []string `json:"languages,omitempty"`
	WorkflowRequestedBy string   `json:"workflowRequestedBy,omitempty"`
	Source              string   `json:"source,omitempty"`
}

// VoiceAlertResponse returns the generated voice alert asset.
type VoiceAlertResponse struct {
	Asset VoiceAlertAsset `json:"asset"`
}

// VoiceAlertListResponse returns all voice alert assets.
type VoiceAlertListResponse struct {
	Assets []VoiceAlertAsset `json:"assets"`
}

// VoiceReviewRequest requests review of a voice alert asset.
type VoiceReviewRequest struct {
	Action    string   `json:"action"`
	Reviewer  string   `json:"reviewer"`
	Note      string   `json:"note,omitempty"`
	Languages []string `json:"languages,omitempty"`
}

// VoiceDeliveryRequest requests delivery of an approved voice alert.
type VoiceDeliveryRequest struct {
	Recipients []VoiceRecipient `json:"recipients"`
	DryRun     bool             `json:"dryRun,omitempty"`
}

// VoiceRecipient is a single voice delivery target.
type VoiceRecipient struct {
	RecipientID string `json:"recipientId,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Language    string `json:"language"`
}

// VoiceDeliveryResponse returns voice delivery attempts.
type VoiceDeliveryResponse struct {
	Attempts []DeliveryAttempt `json:"attempts"`
}

// VoiceAlertAsset is a generated multi-language voice alert.
type VoiceAlertAsset struct {
	ID                  string         `json:"id"`
	AlertID             string         `json:"alertId"`
	AlertTitle          string         `json:"alertTitle"`
	HazardType          string         `json:"hazardType"`
	Severity            string         `json:"severity"`
	TargetLabel         string         `json:"targetLabel"`
	Status              string         `json:"status"`
	ReviewStatus        string         `json:"reviewStatus"`
	Source              string         `json:"source"`
	WorkflowRequestedBy string         `json:"workflowRequestedBy,omitempty"`
	Reviewer            string         `json:"reviewer,omitempty"`
	ReviewNote          string         `json:"reviewNote,omitempty"`
	Variants            []VoiceVariant `json:"variants"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	ReviewedAt          *time.Time     `json:"reviewedAt,omitempty"`
}

// VoiceVariant is a single-language voice recording.
type VoiceVariant struct {
	ID                  string    `json:"id"`
	Language            string    `json:"language"`
	Locale              string    `json:"locale"`
	VoiceName           string    `json:"voiceName"`
	MessageText         string    `json:"messageText"`
	AudioURL            string    `json:"audioUrl"`
	DurationSeconds     int       `json:"durationSeconds"`
	Status              string    `json:"status"`
	ReviewStatus        string    `json:"reviewStatus"`
	AccessibilityChecks []string  `json:"accessibilityChecks"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// IncidentIntakeRequest is the payload sent to the incident service.
type IncidentIntakeRequest struct {
	Type               string       `json:"type"`
	Description        string       `json:"description"`
	Location           Coordinates  `json:"location"`
	PeopleAffected     int          `json:"peopleAffected"`
	InjuriesReported   bool         `json:"injuriesReported"`
	Urgency            string       `json:"urgency"`
	Anonymous          bool         `json:"anonymous"`
	ContactPermission  bool         `json:"contactPermission"`
	AccessibilityNeeds string       `json:"accessibilityNeeds"`
	Media              []string     `json:"media"`
	Reporter           *ReporterRef `json:"reporter,omitempty"`
}

// ReporterRef identifies the reporter when profile linking is allowed.
type ReporterRef struct {
	UserID string `json:"userId"`
	Phone  string `json:"phone,omitempty"`
}

// IncidentIntakeResponse is returned by the incident service.
type IncidentIntakeResponse struct {
	ID        string `json:"id"`
	Reference string `json:"reference"`
	Status    string `json:"status"`
}

// APIError is the standard error response envelope.
type APIError struct {
	Error APIErrorBody `json:"error"`
}

// APIErrorBody is the standard error response body.
type APIErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
