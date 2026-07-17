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

// tokenClaims mirrors auth-service's signed token payload. Keep in sync with
// services/auth-service/internal/models TokenClaims.
type tokenClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId"`
	District  string `json:"district"`
	MFA       bool   `json:"mfa"`
	ExpiresAt int64  `json:"exp"`
}

// verifyToken mirrors auth-service's token scheme: nadaa.<payload>.<sig>,
// where sig is base64url(HMAC-SHA256(secret, payload)). The signature is
// compared with hmac.Equal and the token must not be expired.
func verifyToken(token, secret string, now time.Time) (tokenClaims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" || secret == "" {
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
func bearerToken(r *http.Request) (string, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(header) <= len("bearer ") || !strings.EqualFold(header[:len("bearer ")], "bearer ") {
		return "", false
	}
	token := strings.TrimSpace(header[len("bearer "):])
	return token, token != ""
}
