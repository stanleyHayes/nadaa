package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")

	cfg := &Config{AllowMockActors: true}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "NADAA_ENV=development") {
		t.Fatalf("expected NADAA_ENV=development rejection, got %v", err)
	}
}

func TestValidateAllowsMockActorsInDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "development")

	cfg := &Config{AllowMockActors: true}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors to pass in development, got %v", err)
	}
}

func TestValidateAllowsProductionWithoutMockActors(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")

	if err := (&Config{}).Validate(); err != nil {
		t.Fatalf("expected production config without mock actors to pass, got %v", err)
	}
}
