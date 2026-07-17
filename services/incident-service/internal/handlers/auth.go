package handlers

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

// errInvalidToken marks any token that fails format, signature, or expiry checks.
var errInvalidToken = errors.New("invalid token")

// tokenClaims mirrors the signed payload auth-service issues (nadaa.<payload>.<sig>).
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

// bearerToken extracts the token from an Authorization header.
func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" || token == header {
		return "", false
	}
	return token, true
}

// verifyAuthToken verifies a nadaa.<payload>.<sig> token signed with the shared
// HMAC secret and returns its claims. An empty secret rejects every token.
func verifyAuthToken(secret []byte, now func() time.Time, token string) (tokenClaims, error) {
	if len(secret) == 0 {
		return tokenClaims{}, errInvalidToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return tokenClaims{}, errInvalidToken
	}

	expectedSignature := signAuthPayload(secret, parts[1])
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

	if claims.ExpiresAt <= now().Unix() {
		return tokenClaims{}, errInvalidToken
	}

	return claims, nil
}

func signAuthPayload(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
