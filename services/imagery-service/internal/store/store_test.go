package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/models"
)

func TestRunLifecycleExpiresRecordAndDeletesStoredFile(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	// A negative retention period makes every created record already expired.
	s := NewMemoryStore(now, -1)

	path := filepath.Join(t.TempDir(), "expired.png")
	if err := os.WriteFile(path, []byte("\x89PNG\r\n"), 0o600); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	record := s.Create(models.ImageryUploadInput{Source: "drone"}, "expired.png", "image/png", path, "usr_test", 8, now)

	expired := s.RunLifecycle(now)
	if expired < 1 {
		t.Fatalf("expected at least one expired record, got %d", expired)
	}

	got, ok := s.GetByID(record.ID)
	if !ok || got.Status != "expired" {
		t.Fatalf("expected record marked expired, got %#v (found=%t)", got, ok)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected lifecycle to delete the stored file, stat err=%v", err)
	}
}

func TestRunLifecycleToleratesMissingFile(t *testing.T) {
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	s := NewMemoryStore(now, -1)

	// The seed record img_seed_expiring points at a file that does not exist;
	// expiry must still proceed.
	expired := s.RunLifecycle(now)
	if expired < 1 {
		t.Fatalf("expected at least one expired record, got %d", expired)
	}
	got, ok := s.GetByID("img_seed_expiring")
	if !ok || got.Status != "expired" {
		t.Fatalf("expected seed record marked expired despite missing file, got %#v (found=%t)", got, ok)
	}
}
