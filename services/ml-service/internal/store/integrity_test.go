package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// writeMinimalModelDir seeds the smallest artifact set loadPredictionStore accepts.
func writeMinimalModelDir(t *testing.T, root string) string {
	t.Helper()
	modelDir := filepath.Join(root, "models")
	writeTestFile(t, filepath.Join(modelDir, "baseline-logistic.v1.json"),
		`{"modelVersion":"test-v1","trainingFeatureSetVersion":"fs-v1","featureColumns":[],"coefficients":{"intercept":0},"preprocessing":{"numericStandardization":{}},"limitations":[]}`)
	writeTestFile(t, filepath.Join(modelDir, "sample-predictions.v1.json"),
		`{"modelVersion":"test-v1","featureSetVersion":"fs-v1","predictionCount":0,"predictions":[]}`)
	writeTestFile(t, filepath.Join(root, "generated", "features.v1.json"),
		`{"featureSetVersion":"fs-v1","rows":[]}`)
	return modelDir
}

// writeTestManifest writes a checksums.sha256 over files relative to modelDir.
func writeTestManifest(t *testing.T, modelDir string, names ...string) {
	t.Helper()
	var builder strings.Builder
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(modelDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		sum := sha256.Sum256(data)
		fmt.Fprintf(&builder, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	writeTestFile(t, filepath.Join(modelDir, integrityManifestName), builder.String())
}

func TestVerifyModelIntegrityAcceptsCheckedInManifest(t *testing.T) {
	if err := verifyModelIntegrity("../../../../data/flood-risk/models"); err != nil {
		t.Fatalf("checked-in model artifacts failed integrity verification: %v", err)
	}
}

func TestVerifyModelIntegrityDetectsTampering(t *testing.T) {
	root := t.TempDir()
	modelDir := writeMinimalModelDir(t, root)
	writeTestManifest(t, modelDir, "baseline-logistic.v1.json", "sample-predictions.v1.json", filepath.Join("..", "generated", "features.v1.json"))

	writeTestFile(t, filepath.Join(modelDir, "baseline-logistic.v1.json"),
		`{"modelVersion":"test-v1","trainingFeatureSetVersion":"fs-v1","featureColumns":[],"coefficients":{"intercept":0.9},"preprocessing":{"numericStandardization":{}},"limitations":[]}`)

	if err := verifyModelIntegrity(modelDir); err == nil {
		t.Fatal("expected integrity mismatch to fail")
	}
}

func TestVerifyModelIntegrityRequiresManifest(t *testing.T) {
	root := t.TempDir()
	modelDir := writeMinimalModelDir(t, root)

	if err := verifyModelIntegrity(modelDir); err == nil {
		t.Fatal("expected missing manifest to fail")
	}
}

func TestVerifyModelIntegrityRequiresCoreArtifacts(t *testing.T) {
	root := t.TempDir()
	modelDir := writeMinimalModelDir(t, root)
	// A manifest that covers only a side artifact must not satisfy the check.
	writeTestManifest(t, modelDir, "sample-predictions.v1.json", filepath.Join("..", "generated", "features.v1.json"))

	err := verifyModelIntegrity(modelDir)
	if err == nil {
		t.Fatal("expected manifest without the model artifact to fail")
	}
	if !strings.Contains(err.Error(), "baseline-logistic.v1.json") {
		t.Fatalf("expected error to name the missing required artifact, got %v", err)
	}
}

func TestNewMemoryStoreFailsClosedWithoutManifest(t *testing.T) {
	t.Setenv("NADAA_ML_SKIP_INTEGRITY_CHECK", "")
	root := t.TempDir()
	modelDir := writeMinimalModelDir(t, root)

	if _, err := NewMemoryStore(modelDir); err == nil {
		t.Fatal("expected store load to fail without an integrity manifest")
	}
}

func TestNewMemoryStoreSkipsIntegrityCheckWhenEnvSet(t *testing.T) {
	t.Setenv("NADAA_ML_SKIP_INTEGRITY_CHECK", "true")
	root := t.TempDir()
	modelDir := writeMinimalModelDir(t, root)

	if _, err := NewMemoryStore(modelDir); err != nil {
		t.Fatalf("expected store load with integrity check skipped, got %v", err)
	}
}
