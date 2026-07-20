package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	t.Setenv("NADAA_AUTH_ALLOW_MOCK_ACTORS", "true")
	cfg := &Config{}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected outside development")
	}

	t.Setenv("NADAA_ENV", "development")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors to be allowed in development, got %v", err)
	}
}

func TestValidateAllowsUnsetMockActors(t *testing.T) {
	cfg := &Config{}

	t.Setenv("NADAA_ENV", "production")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected unset mock actors to validate, got %v", err)
	}
}
