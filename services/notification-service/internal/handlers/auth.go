package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// tokenClaims mirrors the claims auth-service signs into nadaa.<payload>.<sig> tokens.
type tokenClaims struct {
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

var errInvalidToken = errors.New("invalid token")

// bearerToken extracts the token from an Authorization header.
func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return ""
	}
	return token
}

// verifyToken validates a nadaa.<payload>.<sig> token against secret and
// returns its claims. The signature is HMAC-SHA256 keyed by secret over the
// encoded payload, compared in constant time.
func verifyToken(token string, secret []byte, now time.Time) (tokenClaims, error) {
	if len(secret) == 0 {
		return tokenClaims{}, errInvalidToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return tokenClaims{}, errInvalidToken
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return tokenClaims{}, errInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return tokenClaims{}, errInvalidToken
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return tokenClaims{}, errInvalidToken
	}
	if claims.ExpiresAt <= now.Unix() {
		return tokenClaims{}, errInvalidToken
	}

	return claims, nil
}

// webhookSecretMatches compares a presented webhook secret against the
// configured one in constant time.
func webhookSecretMatches(configured string, presented string) bool {
	return subtle.ConstantTimeCompare([]byte(configured), []byte(presented)) == 1
}
