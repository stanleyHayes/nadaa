package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSafeLogStripsLineBreaks(t *testing.T) {
	if got := SafeLog("one\ntwo\r\nthree"); got != "onetwothree" {
		t.Fatalf("expected line breaks stripped, got %q", got)
	}
	if got := SafeLog("plain"); got != "plain" {
		t.Fatalf("expected plain value unchanged, got %q", got)
	}
}

func TestCORSLocalhostExceptionRequiresDevelopment(t *testing.T) {
	allowed := map[string]bool{"https://nadaa.gov.gh": true}

	// Outside development a configured allowlist is not bypassed by localhost.
	t.Setenv("NADAA_ENV", "production")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	WithCORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(response, request)
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS origin for localhost outside development, got %q", got)
	}

	// In development the localhost exception applies.
	t.Setenv("NADAA_ENV", "development")
	response = httptest.NewRecorder()
	WithCORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(response, request)
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("expected localhost origin echoed in development, got %q", got)
	}
}
