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

// errInvalidToken marks any token that fails verification.
var errInvalidToken = errors.New("invalid token")

// authTokenClaims mirrors the claims auth-service signs into nadaa tokens.
// Duplicated here because internal packages are not importable across Go modules.
type authTokenClaims struct {
	Subject   string `json:"sub"`
	UserType  string `json:"typ"`
	Phone     string `json:"phone,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role"`
	AgencyID  string `json:"agencyId,omitempty"`
	District  string `json:"district,omitempty"`
	MFA       bool   `json:"mfa,omitempty"`
	ExpiresAt int64  `json:"exp"`
}

// adminActor is the verified authority identity behind an admin action.
type adminActor struct {
	ID       string
	Role     string
	AgencyID string
	District string
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

// verifyAuthToken verifies an auth-service token of the form
// nadaa.<base64url-claims>.<base64url-HMAC-SHA256(secret, payload)>.
// An empty secret can never produce a valid token.
func verifyAuthToken(token, secret string, now time.Time) (authTokenClaims, error) {
	if secret == "" {
		return authTokenClaims{}, errInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return authTokenClaims{}, errInvalidToken
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[1]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expected)) {
		return authTokenClaims{}, errInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return authTokenClaims{}, errInvalidToken
	}
	var claims authTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return authTokenClaims{}, errInvalidToken
	}
	if claims.ExpiresAt <= now.Unix() {
		return authTokenClaims{}, errInvalidToken
	}
	return claims, nil
}
