package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	for name, env := range map[string]string{
		"production": "production",
		"staging":    "staging",
		"unset":      "",
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("NADAA_ENV", env)
			cfg := &Config{AllowMockActors: true}
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected NADAA_AUTH_ALLOW_MOCK_ACTORS to be rejected when NADAA_ENV=%q", env)
			}
		})
	}
}

func TestValidateAllowsMockActorsInDevelopment(t *testing.T) {
	t.Setenv("NADAA_ENV", "development")
	cfg := &Config{AllowMockActors: true}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected mock actors to be allowed in development, got %v", err)
	}
}

func TestValidateAllowsDisabledMockActorsAnywhere(t *testing.T) {
	t.Setenv("NADAA_ENV", "production")
	cfg := &Config{AllowMockActors: false}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected disabled mock actors to validate, got %v", err)
	}
}

func TestLoadMediaDefaults(t *testing.T) {
	cfg := Load()
	if cfg.MediaStoragePath != "./uploads/media" {
		t.Fatalf("expected default media storage path, got %q", cfg.MediaStoragePath)
	}
	if cfg.PublicBaseURL != "http://localhost:8084" {
		t.Fatalf("expected default public base URL, got %q", cfg.PublicBaseURL)
	}
}

func TestLoadMediaOverrides(t *testing.T) {
	t.Setenv("NADAA_INCIDENT_MEDIA_STORAGE_PATH", "/data/incident-media")
	t.Setenv("NADAA_INCIDENT_PUBLIC_BASE_URL", "https://incident.example.com/")
	cfg := Load()
	if cfg.MediaStoragePath != "/data/incident-media" {
		t.Fatalf("expected media storage path override, got %q", cfg.MediaStoragePath)
	}
	if cfg.PublicBaseURL != "https://incident.example.com" {
		t.Fatalf("expected public base URL override without trailing slash, got %q", cfg.PublicBaseURL)
	}
}
