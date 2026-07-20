package store

import (
	"testing"
	"time"
)

func TestSeedAlertsSkippedInProduction(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	s := NewMemoryStore(now)

	if alerts := s.ListAlerts("", false, false, "", "", now); len(alerts) != 0 {
		t.Fatalf("expected no seeded fixture alerts in production, got %#v", alerts)
	}
}

func TestSeedAlertsPresentOutsideProduction(t *testing.T) {
	t.Setenv("NADAA_ENV", "development")
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	s := NewMemoryStore(now)

	alerts := s.ListAlerts("", false, false, "", "", now)
	if len(alerts) != 1 || alerts[0].ID != "alert_fixture_submitted" {
		t.Fatalf("expected the submitted fixture alert outside production, got %#v", alerts)
	}
}
