package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

// AlertServiceClient fetches alerts from the upstream alert service.
type AlertServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAlertServiceClient creates an alert-service client from a base URL.
func NewAlertServiceClient(rawBaseURL string) *AlertServiceClient {
	rawBaseURL = strings.TrimSpace(rawBaseURL)
	if rawBaseURL == "" {
		return nil
	}
	return &AlertServiceClient{
		BaseURL:    strings.TrimRight(rawBaseURL, "/"),
		HTTPClient: &http.Client{Timeout: 2 * time.Second},
	}
}

// ListAlerts fetches current alerts from the upstream alert service.
func (c *AlertServiceClient) ListAlerts(ctx context.Context, now time.Time) ([]models.CitizenAlert, error) {
	parsed, err := url.Parse(c.BaseURL + "/alerts")
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("alert-service returned %d", response.StatusCode)
	}

	var payload models.AuthorityAlertListResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	alerts := make([]models.CitizenAlert, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		if alert.Status != "approved" && alert.Status != "published" {
			continue
		}
		alerts = append(alerts, models.CitizenAlert{
			ID:                 alert.ID,
			Title:              alert.Title,
			HazardType:         alert.HazardType,
			Severity:           alert.Severity,
			Message:            alert.Message,
			Target:             alert.Target,
			TargetLabel:        alert.Target.Label,
			StartsAt:           alert.StartsAt,
			ExpiresAt:          alert.ExpiresAt,
			Status:             alertFeedStatus(alert.StartsAt, alert.ExpiresAt, now),
			RecommendedAction:  alert.RecommendedAction,
			EvacuationRequired: alert.EvacuationRequired,
			ShelterIDs:         alert.ShelterIDs,
			Source:             "alert-service",
			UpdatedAt:          alert.UpdatedAt,
		})
	}

	return alerts, nil
}

func alertFeedStatus(startsAt time.Time, expiresAt time.Time, now time.Time) string {
	if now.Before(startsAt) {
		return "upcoming"
	}
	if !expiresAt.After(now) {
		return "expired"
	}
	return "current"
}

// IncidentServiceClient submits reports to the incident service.
type IncidentServiceClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewIncidentServiceClient creates an incident-service client from a base URL.
func NewIncidentServiceClient(rawBaseURL string) *IncidentServiceClient {
	rawBaseURL = strings.TrimSpace(rawBaseURL)
	if rawBaseURL == "" {
		return nil
	}
	return &IncidentServiceClient{
		BaseURL:    strings.TrimRight(rawBaseURL, "/"),
		HTTPClient: &http.Client{Timeout: 2 * time.Second},
	}
}

// CreateIncident submits an inclusive access report as an incident.
func (c *IncidentServiceClient) CreateIncident(ctx context.Context, report models.InclusiveAccessReport, rawPhone string, profileID string, linkedProfile bool) (models.IncidentIntakeResponse, error) {
	parsed, err := url.Parse(c.BaseURL + "/incidents")
	if err != nil {
		utils.LogError("incident-service handoff url invalid", "baseURL", c.BaseURL, "reportId", report.ID, "error", err)
		return models.IncidentIntakeResponse{}, err
	}
	utils.LogInfo(
		"incident-service handoff request prepared",
		"reportId", report.ID,
		"channel", report.Channel,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"endpoint", parsed.String(),
		"linkedProfile", linkedProfile,
	)

	payload := models.IncidentIntakeRequest{
		Type:               report.Type,
		Description:        report.Description,
		Location:           report.Location,
		PeopleAffected:     0,
		InjuriesReported:   report.Urgency == "life_threatening",
		Urgency:            report.Urgency,
		Anonymous:          !linkedProfile,
		ContactPermission:  linkedProfile,
		AccessibilityNeeds: "Inclusive access channel report",
		Media:              report.Media,
	}
	if linkedProfile {
		payload.Reporter = &models.ReporterRef{UserID: profileID, Phone: rawPhone}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		utils.LogError("incident-service handoff payload marshal failed", "reportId", report.ID, "error", err)
		return models.IncidentIntakeResponse{}, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), strings.NewReader(string(body)))
	if err != nil {
		utils.LogError("incident-service handoff request creation failed", "reportId", report.ID, "error", err)
		return models.IncidentIntakeResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		utils.LogWarn("incident-service handoff request failed", "reportId", report.ID, "endpoint", parsed.String(), "error", err)
		return models.IncidentIntakeResponse{}, err
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		utils.LogWarn("incident-service handoff returned non-success", "reportId", report.ID, "endpoint", parsed.String(), "statusCode", response.StatusCode)
		return models.IncidentIntakeResponse{}, fmt.Errorf("incident-service returned %d", response.StatusCode)
	}

	var result models.IncidentIntakeResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		utils.LogError("incident-service handoff response decode failed", "reportId", report.ID, "statusCode", response.StatusCode, "error", err)
		return models.IncidentIntakeResponse{}, err
	}
	utils.LogInfo("incident-service handoff response decoded", "reportId", report.ID, "incidentId", result.ID, "incidentReference", result.Reference)
	return result, nil
}
