package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	cases := []struct {
		name      string
		env       string
		allowMock bool
		wantErr   bool
	}{
		{"production rejects mock actors", "production", true, true},
		{"staging rejects mock actors", "staging", true, true},
		{"unset env rejects mock actors", "", true, true},
		{"development allows mock actors", "development", true, false},
		{"production without mock actors passes", "production", false, false},
		{"unset env without mock actors passes", "", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NADAA_ENV", tc.env)
			cfg := &Config{AllowMockActors: tc.allowMock}
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error for NADAA_ENV=%q with mock actors enabled", tc.env)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
