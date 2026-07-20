package utils

import (
	"strings"
	"testing"
	"time"
)

// rfc6238Seed is the base32 encoding of the ASCII secret "12345678901234567890"
// used by the RFC 6238 SHA-1 test vectors.
const rfc6238Seed = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestTOTPCodeMatchesRFC6238Vectors(t *testing.T) {
	// RFC 6238 Appendix B SHA-1 vectors, truncated from 8 to 6 digits.
	for at, expected := range map[int64]string{
		59:          "287082",
		1111111109:  "081804",
		1111111111:  "050471",
		1234567890:  "005924",
		2000000000:  "279037",
		20000000000: "353130",
	} {
		code, err := TOTPCode(rfc6238Seed, time.Unix(at, 0).UTC())
		if err != nil {
			t.Fatalf("TOTPCode(%d): %v", at, err)
		}
		if code != expected {
			t.Fatalf("TOTPCode(%d) = %q, want %q", at, code, expected)
		}
	}
}

func TestVerifyTOTPAcceptsCurrentAndAdjacentSteps(t *testing.T) {
	now := time.Unix(1111111111, 0).UTC()
	for _, offset := range []time.Duration{-30 * time.Second, 0, 30 * time.Second} {
		code, err := TOTPCode(rfc6238Seed, now.Add(offset))
		if err != nil {
			t.Fatalf("TOTPCode(%v): %v", offset, err)
		}
		if !VerifyTOTP(rfc6238Seed, code, now) {
			t.Fatalf("expected code from step offset %v to verify", offset)
		}
	}
}

func TestVerifyTOTPRejectsOutsideWindowAndWrongCodes(t *testing.T) {
	now := time.Unix(1111111111, 0).UTC()

	for _, offset := range []time.Duration{-60 * time.Second, 60 * time.Second} {
		code, err := TOTPCode(rfc6238Seed, now.Add(offset))
		if err != nil {
			t.Fatalf("TOTPCode(%v): %v", offset, err)
		}
		if VerifyTOTP(rfc6238Seed, code, now) {
			t.Fatalf("expected code from step offset %v to be rejected", offset)
		}
	}

	current, err := TOTPCode(rfc6238Seed, now)
	if err != nil {
		t.Fatalf("TOTPCode: %v", err)
	}
	wrong := "000000"
	if current == wrong {
		wrong = "111111"
	}
	if VerifyTOTP(rfc6238Seed, wrong, now) {
		t.Fatal("expected wrong code to be rejected")
	}
	if VerifyTOTP(rfc6238Seed, "12345", now) || VerifyTOTP(rfc6238Seed, "abcdef", now) {
		t.Fatal("expected malformed codes to be rejected")
	}
	if VerifyTOTP("not-valid-base32!!", current, now) {
		t.Fatal("expected undecodable secret to be rejected")
	}
}

func TestNewTOTPSecretGeneratesUsableSecrets(t *testing.T) {
	first, second := NewTOTPSecret(), NewTOTPSecret()
	if first == second {
		t.Fatal("expected distinct random secrets")
	}
	if !ValidTOTPSecret(first) {
		t.Fatalf("expected generated secret to validate, got %q", first)
	}
	code, err := TOTPCode(first, time.Now())
	if err != nil || !ValidSixDigitCode(code) {
		t.Fatalf("expected generated secret to produce a 6-digit code, got %q (%v)", code, err)
	}
}

func TestTOTPAuthURLCarriesSecretAndIssuer(t *testing.T) {
	url := TOTPAuthURL("JBSWY3DPEHPK3PXP", "admin@nadaa.local")
	if !strings.HasPrefix(url, "otpauth://totp/") {
		t.Fatalf("expected otpauth URL, got %q", url)
	}
	if !strings.Contains(url, "secret=JBSWY3DPEHPK3PXP") || !strings.Contains(url, "issuer=NADAA") {
		t.Fatalf("expected secret and issuer parameters, got %q", url)
	}
	if !strings.Contains(url, "NADAA:admin@nadaa.local") {
		t.Fatalf("expected account label, got %q", url)
	}
}
