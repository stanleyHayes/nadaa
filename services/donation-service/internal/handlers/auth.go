package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// tokenClaims mirrors the claims auth-service signs into its
// nadaa.<payload>.<sig> tokens (duplicated here because internal packages are
// not importable across Go modules).
type tokenClaims struct {
	Sub       string `json:"sub"`
	Type      string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

// verifyAuthToken validates a nadaa.<payload>.<sig> bearer token against the
// shared HMAC-SHA256 secret and returns its claims. The signature is
// recomputed over the payload and compared in constant time, and the token
// must be unexpired.
func verifyAuthToken(secret, token string, now time.Time) (tokenClaims, bool) {
	if secret == "" {
		return tokenClaims{}, false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return tokenClaims{}, false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[1]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return tokenClaims{}, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, false
	}
	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, false
	}
	if claims.ExpiresAt <= now.Unix() {
		return tokenClaims{}, false
	}
	return claims, true
}

// bearerToken extracts the token from an Authorization: Bearer header.
func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(value) > len("Bearer ") && strings.EqualFold(value[:len("Bearer")], "Bearer") {
		return strings.TrimSpace(value[len("Bearer"):])
	}
	return ""
}
