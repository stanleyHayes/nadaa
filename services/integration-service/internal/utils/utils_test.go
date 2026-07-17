package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHonorsLocalhostOnlyInDevelopment(t *testing.T) {
	allowed := map[string]bool{"https://nadaa.gov.gh": true}
	handler := WithCORS(allowed, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, tc := range []struct {
		name        string
		env         string
		origin      string
		expectAllow bool
	}{
		{"development allows localhost", "development", "http://localhost:5173", true},
		{"production ignores localhost", "production", "http://localhost:5173", false},
		{"unset env ignores localhost", "", "http://localhost:5173", false},
		{"production still allows configured origin", "production", "https://nadaa.gov.gh", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NADAA_ENV", tc.env)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/integrations/contracts", nil)
			request.Header.Set("Origin", tc.origin)

			handler.ServeHTTP(response, request)

			allowedOrigin := response.Header().Get("Access-Control-Allow-Origin")
			if tc.expectAllow && allowedOrigin != tc.origin {
				t.Fatalf("expected origin %q echoed, got %q", tc.origin, allowedOrigin)
			}
			if !tc.expectAllow && allowedOrigin == tc.origin {
				t.Fatalf("expected origin %q to not be echoed", tc.origin)
			}
		})
	}
}

func TestValidateWKTLineString(t *testing.T) {
	for _, tc := range []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"valid linestring", "LINESTRING(-0.187 5.6037, -0.20 5.61)", false},
		{"lowercase prefix accepted", "linestring(-0.187 5.6037, -0.20 5.61)", false},
		{"not a linestring", "POINT(-0.187 5.6037)", true},
		{"single coordinate", "LINESTRING(-0.187 5.6037)", true},
		{"non numeric", "LINESTRING(east 5.6037, -0.20 5.61)", true},
		{"out of bounds", "LINESTRING(-0.187 5.6037, -200 95.0)", true},
		{"missing suffix", "LINESTRING(-0.187 5.6037, -0.20 5.61", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, _ := ValidateWKTLineString(tc.value)
			if tc.expectErr && code == "" {
				t.Fatalf("expected error for %q", tc.value)
			}
			if !tc.expectErr && code != "" {
				t.Fatalf("expected %q to be valid, got code %q", tc.value, code)
			}
		})
	}
}

func TestSanitizeLogValueStripsLineBreaks(t *testing.T) {
	if got := SanitizeLogValue("road\nINFO forged\r\nline"); got != "roadINFO forgedline" {
		t.Fatalf("expected line breaks stripped, got %q", got)
	}
}
