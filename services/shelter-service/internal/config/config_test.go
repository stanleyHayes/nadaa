package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")
	cfg := &Config{AllowMockActors: true}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected when NADAA_ENV is not development")
	}
}

func TestValidateRejectsMockActorsWhenEnvUnset(t *testing.T) {
	t.Setenv("NADAA_ENV", "")
	cfg := &Config{AllowMockActors: true}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected when NADAA_ENV is unset")
	}
}

func TestValidateAllowsMockActorsInDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "development")
	cfg := &Config{AllowMockActors: true}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors to be allowed in development, got %v", err)
	}
}

func TestValidateAllowsSecureDefaults(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")
	cfg := &Config{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default configuration to validate, got %v", err)
	}
}
