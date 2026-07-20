package store

import (
	"os"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/alert-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/alert-service/internal/utils"
)

func seedAlerts(now time.Time) []models.AuthorityAlert {
	// The fixture alert exists to exercise the approval queue in development
	// and smoke environments. In production it must never be seeded: a live
	// fixture in the queue could be approved into real citizen sends.
	if strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_ENV")), "production") {
		return nil
	}
	return []models.AuthorityAlert{
		{
			ID:                 "alert_fixture_submitted",
			Title:              "Accra flood watch",
			HazardType:         "flood",
			Severity:           "warning",
			Message:            "Heavy rainfall may cause flooding in low-lying communities.",
			Target:             models.AlertTarget{Type: "district", IDs: []string{"accra-metropolitan"}, Label: "Accra Metropolitan"},
			StartsAt:           now.Add(30 * time.Minute),
			ExpiresAt:          now.Add(12 * time.Hour),
			RecommendedAction:  "Avoid flooded roads and prepare to move to higher ground.",
			EvacuationRequired: false,
			ShelterIDs:         []string{"00000000-0000-0000-0000-000000000301"},
			IssuingAgencyID:    "00000000-0000-0000-0000-000000000101",
			IssuedBy:           "usr_dispatcher_fixture",
			Status:             "submitted",
			CreatedAt:          now.Add(-45 * time.Minute),
			UpdatedAt:          now.Add(-15 * time.Minute),
			SubmittedAt:        utils.TimePtr(now.Add(-15 * time.Minute)),
		},
	}
}
