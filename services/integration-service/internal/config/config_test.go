package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActors: true}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected outside development")
	}

	t.Setenv("NADAA_ENV", "")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected with NADAA_ENV unset")
	}
}

func TestValidateAllowsMockActorsInDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActors: true}

	t.Setenv("NADAA_ENV", "development")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors allowed in development, got %v", err)
	}
}

func TestValidateAllowsProductionDefaults(t *testing.T) {
	cfg := &Config{}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config, got %v", err)
	}
}
