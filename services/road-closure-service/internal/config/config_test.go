package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActorHeaders: true}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected outside development")
	}

	t.Setenv("NADAA_ENV", "development")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors to be allowed in development, got %v", err)
	}
}

func TestValidateAllowsDisabledMockActors(t *testing.T) {
	cfg := &Config{AllowMockActorHeaders: false}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected disabled mock actors to validate, got %v", err)
	}
}
