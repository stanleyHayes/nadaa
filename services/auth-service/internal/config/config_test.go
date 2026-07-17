package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsDevOnlySettingsOutsideDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")
	t.Setenv("NADAA_AUTH_ALLOW_INSECURE_SECRET", "true")

	for name, cfg := range map[string]*Config{
		"mock actor headers": {AllowMockActorHeaders: true},
		"exposed dev OTP":    {ExposeDevOTP: true},
		"mock OTP":           {MockOTP: "123456"},
	} {
		err := cfg.Validate()
		if err == nil || !strings.Contains(err.Error(), "NADAA_ENV=development") {
			t.Fatalf("%s: expected NADAA_ENV=development rejection, got %v", name, err)
		}
	}
}

func TestValidateAllowsDevOnlySettingsInDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "development")
	t.Setenv("NADAA_AUTH_ALLOW_INSECURE_SECRET", "true")

	cfg := &Config{AllowMockActorHeaders: true, ExposeDevOTP: true, MockOTP: "123456"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected development settings to pass, got %v", err)
	}
}

func TestValidateRejectsMockActorsEvenWithInsecureSecretBypass(t *testing.T) {
	// The insecure-secret bypass must not also unlock mock actor headers.
	t.Setenv("NADAA_ENV", "production")
	t.Setenv("NADAA_AUTH_ALLOW_INSECURE_SECRET", "true")

	cfg := &Config{AllowMockActorHeaders: true}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected mock actor headers to be rejected outside development")
	}
}

func TestValidateAcceptsStrongSecretOutsideDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")

	cfg := &Config{TokenSecret: strings.Repeat("s", 32)}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected strong secret to pass, got %v", err)
	}
}
