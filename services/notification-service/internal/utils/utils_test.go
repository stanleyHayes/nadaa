package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stanleyHayes/nadaa/services/notification-service/internal/models"
)

func TestWithCORSLocalhostExceptionOnlyInDevelopment(t *testing.T) {
	t.Setenv("NADAA_ALLOWED_ORIGINS", "https://app.nadaa.gov.gh")
	allowed := AllowedOriginsFromEnv()

	handler := WithCORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := func(origin string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/x", nil)
		r.Header.Set("Origin", origin)
		return r
	}

	// A configured origin is always echoed.
	configured := httptest.NewRecorder()
	handler.ServeHTTP(configured, request("https://app.nadaa.gov.gh"))
	if got := configured.Header().Get("Access-Control-Allow-Origin"); got != "https://app.nadaa.gov.gh" {
		t.Fatalf("expected configured origin echoed, got %q", got)
	}

	// Outside development the localhost exception must not bypass the allowlist.
	t.Setenv("NADAA_ENV", "production")
	prod := httptest.NewRecorder()
	handler.ServeHTTP(prod, request("http://localhost:3000"))
	if got := prod.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected localhost origin rejected outside development, got %q", got)
	}

	// In development the localhost exception still applies.
	t.Setenv("NADAA_ENV", "development")
	dev := httptest.NewRecorder()
	handler.ServeHTTP(dev, request("http://localhost:3000"))
	if got := dev.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected localhost origin echoed in development, got %q", got)
	}
}

func TestInclusiveLocationOmitsMissingAndZeroCoordinates(t *testing.T) {
	location, label := InclusiveLocation(nil, nil)
	if location != nil {
		t.Fatalf("expected nil coordinates without caller input, got %#v", location)
	}
	if label == "" {
		t.Fatal("expected a follow-up label without caller input")
	}

	zero, _ := InclusiveLocation(&models.Coordinates{Lat: 0, Lng: 0}, nil)
	if zero != nil {
		t.Fatalf("expected (0,0) treated as no-location, got %#v", zero)
	}

	outOfRange, _ := InclusiveLocation(&models.Coordinates{Lat: 95, Lng: 0}, nil)
	if outOfRange != nil {
		t.Fatalf("expected out-of-range coordinates treated as no-location, got %#v", outOfRange)
	}

	shared, sharedLabel := InclusiveLocation(&models.Coordinates{Lat: 5.566, Lng: -0.242}, nil)
	if shared == nil || shared.Lat != 5.566 || shared.Lng != -0.242 {
		t.Fatalf("expected caller-shared coordinates kept, got %#v", shared)
	}
	if sharedLabel == "" {
		t.Fatal("expected a label for caller-shared coordinates")
	}
}
