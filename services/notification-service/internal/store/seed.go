package store

import (
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
)

func seedCitizenAlerts(now time.Time) []models.CitizenAlert {
	return []models.CitizenAlert{
		newCitizenAlert(
			"alert_feed_current_flood",
			"Severe flood warning",
			"flood",
			"severe_warning",
			"Heavy rainfall and rising drains may flood low-lying parts of Accra Metro and Tema.",
			models.AlertTarget{Type: "district", IDs: []string{"accra-metropolitan", "tema-metropolitan"}, Label: "Accra Metro and Tema"},
			now.Add(-30*time.Minute),
			now.Add(5*time.Hour),
			"Move away from drains, avoid flooded roads, and prepare to go to a shelter if directed.",
			true,
			[]string{"shelter-ama-001", "shelter-osu-002"},
			"fixture",
			now.Add(-20*time.Minute),
			now,
		),
		newCitizenAlert(
			"alert_feed_current_fire",
			"Market fire watch",
			"fire",
			"watch",
			"Responders are monitoring dense market areas after smoke reports near electrical kiosks.",
			models.AlertTarget{Type: "community", IDs: []string{"accra-central"}, Label: "Accra Central"},
			now.Add(-20*time.Minute),
			now.Add(3*time.Hour),
			"Keep access lanes open, avoid overloaded sockets, and call 112 if you see flames or heavy smoke.",
			false,
			nil,
			"fixture",
			now.Add(-15*time.Minute),
			now,
		),
		newCitizenAlert(
			"alert_feed_expired_road",
			"Road hazard resolved",
			"road_crash",
			"advisory",
			"Earlier congestion near Kaneshie Market Road has cleared after responders reopened the lane.",
			models.AlertTarget{Type: "radius", IDs: []string{"kaneshie-market-road"}, Label: "Kaneshie Market Road", Center: &models.Coordinates{Lat: 5.566, Lng: -0.242}, RadiusMeters: 1500},
			now.Add(-8*time.Hour),
			now.Add(-2*time.Hour),
			"Continue to drive carefully and give way to emergency vehicles.",
			false,
			nil,
			"fixture",
			now.Add(-2*time.Hour),
			now,
		),
	}
}

func newCitizenAlert(id string, title string, hazard string, severity string, message string, target models.AlertTarget, startsAt time.Time, expiresAt time.Time, action string, evacuation bool, shelters []string, source string, updatedAt time.Time, now time.Time) models.CitizenAlert {
	return models.CitizenAlert{
		ID:                 id,
		Title:              title,
		HazardType:         hazard,
		Severity:           severity,
		Message:            message,
		Target:             target,
		TargetLabel:        target.Label,
		StartsAt:           startsAt,
		ExpiresAt:          expiresAt,
		Status:             alertFeedStatus(startsAt, expiresAt, now),
		RecommendedAction:  action,
		EvacuationRequired: evacuation,
		ShelterIDs:         shelters,
		Source:             source,
		UpdatedAt:          updatedAt,
	}
}
