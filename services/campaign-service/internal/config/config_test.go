package config

import "testing"

func TestValidateRejectsMockActorHeadersOutsideDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActorHeaders: true}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actor headers to be rejected outside development")
	}

	t.Setenv("NADAA_ENV", "")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actor headers to be rejected with NADAA_ENV unset")
	}
}

func TestValidateAllowsMockActorHeadersInDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActorHeaders: true}

	t.Setenv("NADAA_ENV", "development")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actor headers allowed in development, got %v", err)
	}
}

func TestValidateAllowsProductionWithoutMockActorHeaders(t *testing.T) {
	cfg := &Config{}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config, got %v", err)
	}
}
