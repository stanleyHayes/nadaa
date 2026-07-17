package store

import (
	"context"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
)

// blockingProvider blocks inside Send until released, simulating a slow
// delivery gateway.
type blockingProvider struct {
	entered chan struct{}
	release chan struct{}
}

func (p blockingProvider) Send(ctx context.Context, _ models.ProviderMessage) models.ProviderResult {
	close(p.entered)
	select {
	case <-p.release:
	case <-ctx.Done():
	}
	return models.ProviderResult{Provider: "blocking", Status: "delivered"}
}

func TestCreateDeliveryAttemptsDoesNotHoldStoreLockDuringSend(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	s := NewMemoryStore(now)
	provider := blockingProvider{entered: make(chan struct{}), release: make(chan struct{})}

	done := make(chan []models.DeliveryAttempt, 1)
	go func() {
		done <- s.CreateDeliveryAttempts(
			context.Background(),
			models.CitizenAlert{ID: "alert_1", Title: "Flood warning"},
			models.DeliveryRequest{Phone: "+233200000000", Channels: []string{"sms"}},
			map[string]models.NotificationProvider{"sms": provider},
			now,
		)
	}()

	<-provider.entered

	// A store read must proceed while the provider send is still blocked.
	listDone := make(chan struct{})
	go func() {
		_ = s.ListAlerts(models.AlertFeedFilters{}, now)
		close(listDone)
	}()
	select {
	case <-listDone:
	case <-time.After(2 * time.Second):
		t.Fatal("store read blocked while a provider send was in flight")
	}

	close(provider.release)
	select {
	case attempts := <-done:
		if len(attempts) != 1 || attempts[0].Status != "delivered" {
			t.Fatalf("expected one delivered attempt, got %#v", attempts)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("delivery attempts never completed after the provider was released")
	}
}

func TestCreateVoiceDeliveryAttemptsDoesNotHoldStoreLockDuringSend(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	s := NewMemoryStore(now)
	provider := blockingProvider{entered: make(chan struct{}), release: make(chan struct{})}
	asset := models.VoiceAlertAsset{
		ID:       "voice_alert_1",
		AlertID:  "alert_1",
		Variants: []models.VoiceVariant{{ID: "voice_variant_1", Language: "en", ReviewStatus: "approved"}},
	}

	done := make(chan []models.DeliveryAttempt, 1)
	go func() {
		done <- s.CreateVoiceDeliveryAttempts(
			context.Background(),
			asset,
			models.VoiceDeliveryRequest{Recipients: []models.VoiceRecipient{{Phone: "+233200000000", Language: "en"}}},
			map[string]models.NotificationProvider{"voice": provider},
			now,
		)
	}()

	<-provider.entered

	listDone := make(chan struct{})
	go func() {
		_ = s.ListDeliveryLogs(models.LogFilters{})
		close(listDone)
	}()
	select {
	case <-listDone:
	case <-time.After(2 * time.Second):
		t.Fatal("store read blocked while a voice provider send was in flight")
	}

	close(provider.release)
	select {
	case attempts := <-done:
		if len(attempts) != 1 || attempts[0].Status != "delivered" {
			t.Fatalf("expected one delivered voice attempt, got %#v", attempts)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("voice delivery attempts never completed after the provider was released")
	}
}

func TestCreateAccessReportDeduplicatesProviderMessage(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	s := NewMemoryStore(now)
	report := models.InclusiveAccessReport{
		Channel:           "sms",
		Provider:          "sms_sandbox",
		ProviderMessageID: "msg-1",
		Type:              "flood",
		Urgency:           "high",
		Status:            "queued",
		CreatedAt:         now,
	}

	first, created := s.CreateAccessReport(report)
	if !created || first.ID == "" {
		t.Fatalf("expected a new report, got %#v (created=%v)", first, created)
	}
	second, created := s.CreateAccessReport(report)
	if created {
		t.Fatalf("expected the duplicate to be rejected, got %#v", second)
	}
	if second.ID != first.ID {
		t.Fatalf("expected the existing report %q, got %q", first.ID, second.ID)
	}

	// Reports without a providerMessageId are never deduplicated.
	anonymous := models.InclusiveAccessReport{Channel: "sms", Type: "flood", Status: "queued", CreatedAt: now}
	if _, created := s.CreateAccessReport(anonymous); !created {
		t.Fatal("expected a report without providerMessageId to always be created")
	}
}
