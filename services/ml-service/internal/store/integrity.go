package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// integrityManifestName is the sha256 manifest verified before model artifacts load.
const integrityManifestName = "checksums.sha256"

// integrityRequiredFiles must be covered by the manifest: the model artifacts and
// the generated feature file the simulation grid is scored from.
var integrityRequiredFiles = []string{
	"baseline-logistic.v1.json",
	"sample-predictions.v1.json",
	filepath.Join("..", "generated", "features.v1.json"),
}

// integrityCheckSkipped reports whether the startup integrity check is disabled.
func integrityCheckSkipped() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("NADAA_ML_SKIP_INTEGRITY_CHECK")), "true")
}

// verifyModelIntegrity recomputes the sha256 of every file listed in the
// model directory's checksums.sha256 manifest and fails closed on a missing
// manifest, a missing file, or any digest mismatch.
func verifyModelIntegrity(modelDir string) error {
	manifestPath := filepath.Join(modelDir, integrityManifestName)
	//nolint:gosec // path is resolved from trusted model directory configuration.
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read integrity manifest: %w", err)
	}

	entries, err := parseIntegrityManifest(string(data))
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return errors.New("integrity manifest lists no files")
	}

	covered := make(map[string]bool, len(entries))
	for name, expected := range entries {
		cleaned := filepath.Clean(name)
		if filepath.IsAbs(cleaned) {
			return fmt.Errorf("integrity manifest has absolute path %q", name)
		}
		covered[cleaned] = true

		path := filepath.Join(modelDir, cleaned)
		//nolint:gosec // paths come from the trusted integrity manifest under the model directory.
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open integrity-checked file %q: %w", cleaned, err)
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return fmt.Errorf("hash integrity-checked file %q: %w", cleaned, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close integrity-checked file %q: %w", cleaned, closeErr)
		}
		if actual := hex.EncodeToString(hash.Sum(nil)); actual != expected {
			return fmt.Errorf("integrity check failed for %q: expected sha256 %s, got %s", cleaned, expected, actual)
		}
	}

	for _, required := range integrityRequiredFiles {
		if !covered[filepath.Clean(required)] {
			return fmt.Errorf("integrity manifest does not cover required artifact %q", required)
		}
	}
	return nil
}

// parseIntegrityManifest parses standard `shasum -a 256` output lines of the
// form "<hex digest>  <path>" (or "<hex digest> *<path>" in binary mode).
func parseIntegrityManifest(data string) (map[string]string, error) {
	entries := map[string]string{}
	for line := range strings.Lines(data) {
		line = strings.TrimRight(line, "\r\n")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed integrity manifest line %q", line)
		}
		sum := strings.ToLower(strings.TrimSpace(parts[0]))
		name := strings.TrimPrefix(strings.TrimLeft(parts[1], " "), "*")
		if len(sum) != sha256.Size*2 {
			return nil, fmt.Errorf("integrity manifest line %q has an invalid sha256 digest", line)
		}
		if name == "" {
			return nil, fmt.Errorf("integrity manifest line %q has no file path", line)
		}
		entries[name] = sum
	}
	return entries, nil
}

// logIntegritySkip warns once per store load when the check is disabled.
func logIntegritySkip() {
	log.Printf("WARN ml-service model integrity check skipped via NADAA_ML_SKIP_INTEGRITY_CHECK")
}
