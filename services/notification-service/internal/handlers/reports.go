package handlers

import (
	"context"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/notification-service/internal/utils"
)

func (s *Server) submitInclusiveReport(ctx context.Context, report models.InclusiveAccessReport, rawPhone string, profileID string, linkedProfile bool) models.InclusiveAccessReport {
	if s.incidentClient == nil {
		report.Status = "queued"
		report.FailureReason = "incident-service handoff is not configured"
		utils.LogWarn(
			"inclusive report queued without incident-service",
			"reportId", report.ID,
			"channel", report.Channel,
			"hazard", report.Type,
			"urgency", report.Urgency,
			"phoneRef", report.PhoneRef,
		)
		s.store.UpdateAccessReport(report)
		return report
	}

	utils.LogInfo(
		"inclusive report incident handoff starting",
		"reportId", report.ID,
		"channel", report.Channel,
		"hazard", report.Type,
		"urgency", report.Urgency,
		"phoneRef", report.PhoneRef,
		"linkedProfile", linkedProfile,
	)
	response, err := s.incidentClient.CreateIncident(ctx, report, rawPhone, profileID, linkedProfile)
	if err != nil {
		report.Status = "queued"
		report.FailureReason = err.Error()
		utils.LogWarn(
			"inclusive report incident handoff failed",
			"reportId", report.ID,
			"channel", report.Channel,
			"phoneRef", report.PhoneRef,
			"error", err,
		)
		s.store.UpdateAccessReport(report)
		return report
	}

	report.Status = "submitted"
	report.IncidentID = response.ID
	report.IncidentReference = response.Reference
	report.FailureReason = ""
	utils.LogInfo(
		"inclusive report incident handoff completed",
		"reportId", report.ID,
		"channel", report.Channel,
		"phoneRef", report.PhoneRef,
		"incidentId", report.IncidentID,
		"incidentReference", report.IncidentReference,
	)
	s.store.UpdateAccessReport(report)
	return report
}
