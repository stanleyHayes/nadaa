package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidDecimal(t *testing.T) {
	valid := []string{"0", "1000", "1000.00", "0.5", "12500.00", "123456789"}
	for _, value := range valid {
		if !ValidDecimal(value) {
			t.Errorf("expected %q to be a valid decimal", value)
		}
	}

	invalid := []string{"", "NaN", "Inf", "+Inf", "-Inf", "1e5", "1E5", "1e-5", "0x10", "0x1p4", "1.2.3", "abc", " 1", "1 ", ".5", "1.", "+1", "-1"}
	for _, value := range invalid {
		if ValidDecimal(value) {
			t.Errorf("expected %q to be rejected", value)
		}
	}
}

func TestSafeLogValueStripsNewlines(t *testing.T) {
	if got := SafeLogValue("line1\nline2\r\nline3"); got != "line1 line2  line3" {
		t.Fatalf("expected newlines stripped, got %q", got)
	}
	if got := SafeLogValue("clean"); got != "clean" {
		t.Fatalf("expected clean value untouched, got %q", got)
	}
}

func TestWithCORSLocalhostGating(t *testing.T) {
	allowed := map[string]bool{"https://app.nadaa.gov.gh": true}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	origin := "http://localhost:3000"

	t.Run("localhost echo disabled outside development", func(t *testing.T) {
		t.Setenv("NADAA_ENV", "production")
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/claims", nil)
		request.Header.Set("Origin", origin)
		WithCORS(allowed, next).ServeHTTP(response, request)

		if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("expected no allow-origin for localhost outside development, got %q", got)
		}
	})

	t.Run("localhost echo allowed in development", func(t *testing.T) {
		t.Setenv("NADAA_ENV", "development")
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/claims", nil)
		request.Header.Set("Origin", origin)
		WithCORS(allowed, next).ServeHTTP(response, request)

		if got := response.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("expected localhost origin echoed in development, got %q", got)
		}
	})

	t.Run("allowlisted origin always echoed", func(t *testing.T) {
		t.Setenv("NADAA_ENV", "production")
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/claims", nil)
		request.Header.Set("Origin", "https://app.nadaa.gov.gh")
		WithCORS(allowed, next).ServeHTTP(response, request)

		if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://app.nadaa.gov.gh" {
			t.Fatalf("expected allowlisted origin echoed, got %q", got)
		}
	})
}
