package store

import (
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
)

func TestDonationReferencesDoNotCollideAcrossRestart(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	input := models.CreateDonationInput{
		DonorName:   "Ama",
		Email:       "ama@example.com",
		AmountMinor: 5000,
		Currency:    "GHS",
		Provider:    "sandbox",
	}

	first := NewMemoryStore(now)
	seen := map[string]bool{}
	for range 3 {
		donation := first.CreateDonation(input, now)
		seen[donation.Reference] = true
	}

	// Simulate a same-day restart: a new process reloads the day's persisted
	// donations and must continue the reference counter past them instead of
	// re-issuing references the payment gateway still holds.
	restarted := &MemoryStore{donations: first.ListDonations(models.DonationFilter{})}
	restarted.advanceSeqPastLoadedReferences(now)

	for range 3 {
		donation := restarted.CreateDonation(input, now)
		if seen[donation.Reference] {
			t.Fatalf("reference %q re-issued after a same-day restart", donation.Reference)
		}
		seen[donation.Reference] = true
	}
}

func TestReferenceDaySuffix(t *testing.T) {
	suffix, ok := referenceDaySuffix("GIFT-20260720-042", "20260720")
	if !ok || suffix != 42 {
		t.Fatalf("expected suffix 42, got %d (ok=%v)", suffix, ok)
	}
	if _, ok := referenceDaySuffix("GIFT-20260719-042", "20260720"); ok {
		t.Fatal("expected a different day's reference to be ignored")
	}
	if _, ok := referenceDaySuffix("not-a-reference", "20260720"); ok {
		t.Fatal("expected a malformed reference to be ignored")
	}
}
