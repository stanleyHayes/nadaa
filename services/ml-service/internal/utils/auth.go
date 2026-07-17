package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// ErrInvalidToken is returned when a bearer token fails verification.
var ErrInvalidToken = errors.New("invalid token")

// TokenClaims is the signed payload of a NADAA access token issued by auth-service.
type TokenClaims struct {
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

// BearerToken extracts the token from an Authorization header.
func BearerToken(r *http.Request) (string, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return "", false
	}
	return token, true
}

// VerifyToken verifies a nadaa.<payload>.<sig> token against the shared
// NADAA_AUTH_TOKEN_SECRET, mirroring auth-service's signing scheme. An empty
// secret never verifies, so authority requests fail closed when it is unset.
func VerifyToken(token string, secret []byte, now time.Time) (TokenClaims, error) {
	if len(secret) == 0 {
		return TokenClaims{}, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return TokenClaims{}, ErrInvalidToken
	}

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return TokenClaims{}, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return TokenClaims{}, ErrInvalidToken
	}
	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return TokenClaims{}, ErrInvalidToken
	}
	if claims.ExpiresAt <= now.Unix() {
		return TokenClaims{}, ErrInvalidToken
	}
	return claims, nil
}
