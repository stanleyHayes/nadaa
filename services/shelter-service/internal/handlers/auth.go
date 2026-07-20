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

// tokenClaims mirrors the claims payload signed by auth-service. The struct
// is duplicated here because auth-service is a separate Go module.
type tokenClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa"`
	ExpiresAt int64  `json:"exp"`
}

// bearerToken extracts a Bearer token from the Authorization header.
func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return "", false
	}
	return token, true
}

// verifyToken verifies an auth-service nadaa.<payload>.<sig> token: the
// signature is base64.RawURLEncoding(HMAC-SHA256(secret, payload)) compared
// with hmac.Equal, the claims must not be expired, and the token must be an
// agency token — citizen tokens never mint an authority context.
func verifyToken(token string, secret []byte, now time.Time) (tokenClaims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return tokenClaims{}, false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
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
	if claims.UserType != "agency" {
		return tokenClaims{}, false
	}
	return claims, true
}
