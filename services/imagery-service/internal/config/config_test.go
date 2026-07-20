package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActors: true}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actors to be rejected outside development")
	}
}

func TestValidateAllowsMockActorsInDevelopment(t *testing.T) {
	cfg := &Config{AllowMockActors: true, Development: true}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors allowed in development, got %v", err)
	}
}

func TestValidateAllowsProductionDefaults(t *testing.T) {
	cfg := &Config{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config, got %v", err)
	}
}
