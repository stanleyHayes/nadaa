package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidAuthToken is returned when a bearer token fails verification.
var ErrInvalidAuthToken = errors.New("invalid auth token")

// AuthClaims are the verified claims carried by a nadaa.<payload>.<sig> token
// issued by auth-service.
type AuthClaims struct {
	UserID    string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

// VerifyAuthToken verifies a nadaa.<payload>.<sig> bearer token issued by
// auth-service: the signature is HMAC-SHA256 over the base64url payload, keyed
// with the shared NADAA_AUTH_TOKEN_SECRET. It mirrors auth-service's scheme,
// which is not importable across Go modules.
func VerifyAuthToken(secret, token string, now time.Time) (AuthClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" || secret == "" {
		return AuthClaims{}, ErrInvalidAuthToken
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[1]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return AuthClaims{}, ErrInvalidAuthToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AuthClaims{}, ErrInvalidAuthToken
	}

	var claims AuthClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return AuthClaims{}, ErrInvalidAuthToken
	}
	if claims.ExpiresAt <= now.Unix() {
		return AuthClaims{}, ErrInvalidAuthToken
	}
	return claims, nil
}
