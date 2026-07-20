package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActors: true, Env: "production"}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "NADAA_ENV=development") {
		t.Fatalf("expected NADAA_ENV=development rejection, got %v", err)
	}

	// An unset environment must also fail closed.
	cfg = &Config{AllowMockActors: true}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected when NADAA_ENV is unset")
	}
}

func TestValidateAllowsMockActorsInDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActors: true, Env: "development"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors to pass in development, got %v", err)
	}
}

func TestValidateAllowsProductionWithoutMockActors(t *testing.T) {
	cfg := &Config{Env: "production"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected production config without mock actors to pass, got %v", err)
	}
}

func TestFixtureAlertsEnabledOnlyInDevelopmentOrOptIn(t *testing.T) {
	for name, tc := range map[string]struct {
		cfg      *Config
		expected bool
	}{
		"development":       {&Config{Env: "development"}, true},
		"production":        {&Config{Env: "production"}, false},
		"unset env":         {&Config{}, false},
		"production opt-in": {&Config{Env: "production", AllowFixtureAlerts: true}, true},
	} {
		if got := tc.cfg.FixtureAlertsEnabled(); got != tc.expected {
			t.Fatalf("%s: expected FixtureAlertsEnabled=%v, got %v", name, tc.expected, got)
		}
	}
}
