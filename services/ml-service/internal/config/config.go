package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

// Config holds ml-service configuration loaded from the environment.
type Config struct {
	Addr           string
	ModelDir       string
	AllowedOrigins map[string]bool
}

const defaultBindAddr = ":8094"

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	modelDir, err := resolveModelDir()
	if err != nil {
		return nil, err
	}

	return &Config{
		Addr:           resolveListenAddr("NADAA_ML_ADDR", defaultBindAddr),
		ModelDir:       modelDir,
		AllowedOrigins: allowedOriginsFromEnv(),
	}, nil
}

// resolveListenAddr honors a platform-provided PORT (e.g. Render sets a bare
// number like "10000"), normalizing it to ":PORT", then a service-specific
// address override, then the default. This lets the service bind the port the
// host expects while preserving local defaults.
func resolveListenAddr(addrKey, fallback string) string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	if addrKey != "" {
		if value := strings.TrimSpace(os.Getenv(addrKey)); value != "" {
			return value
		}
	}
	return fallback
}

func resolveModelDir() (string, error) {
	if value := strings.TrimSpace(os.Getenv("NADAA_ML_MODEL_DIR")); value != "" {
		return value, nil
	}

	candidates := []string{
		filepath.Join("data", "flood-risk", "models"),
		filepath.Join("..", "..", "data", "flood-risk", "models"),
		filepath.Join("/app", "data", "flood-risk", "models"),
	}
	for _, candidate := range candidates {
		if utils.FileExists(filepath.Join(candidate, "baseline-logistic.v1.json")) &&
			utils.FileExists(filepath.Join(candidate, "sample-predictions.v1.json")) {
			return candidate, nil
		}
	}
	return "", errors.New("could not find flood-risk model artifacts; set NADAA_ML_MODEL_DIR")
}

func allowedOriginsFromEnv() map[string]bool {
	raw := strings.TrimSpace(os.Getenv("NADAA_ALLOWED_ORIGINS"))
	if raw == "" || raw == "*" {
		return nil
	}

	allowed := map[string]bool{}
	for origin := range strings.SplitSeq(raw, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}
