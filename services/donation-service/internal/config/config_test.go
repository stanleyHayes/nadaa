package config

import "testing"

func TestValidateRejectsMockActorsOutsideDevelopment(t *testing.T) {
	cases := []struct {
		name            string
		env             string
		allowMockActors bool
		wantErr         bool
	}{
		{name: "mock actors without env", env: "", allowMockActors: true, wantErr: true},
		{name: "mock actors in production", env: "production", allowMockActors: true, wantErr: true},
		{name: "mock actors in staging", env: "staging", allowMockActors: true, wantErr: true},
		{name: "mock actors in development", env: "development", allowMockActors: true, wantErr: false},
		{name: "mock actors disabled without env", env: "", allowMockActors: false, wantErr: false},
		{name: "mock actors disabled in production", env: "production", allowMockActors: false, wantErr: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{Env: tc.env, AllowMockActors: tc.allowMockActors}
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected Validate to reject NADAA_AUTH_ALLOW_MOCK_ACTORS=true with NADAA_ENV=%q", tc.env)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected Validate to pass with NADAA_ENV=%q allowMockActors=%v, got %v", tc.env, tc.allowMockActors, err)
			}
		})
	}
}
