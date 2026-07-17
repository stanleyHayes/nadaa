package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/road-closure-service/internal/models"
)

// bearerToken extracts the token from an Authorization: Bearer header.
func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
}

// verifyAuthToken validates a nadaa.<payload>.<sig> token issued by
// auth-service: sig is base64url(HMAC-SHA256(secret, payload)) and the claims
// must not be expired. Mirrors auth-service's scheme, which is not importable
// across Go modules.
func verifyAuthToken(token string, secret []byte, now time.Time) (models.TokenClaims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return models.TokenClaims{}, false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[1]))
	if !hmac.Equal([]byte(parts[2]), []byte(base64.RawURLEncoding.EncodeToString(mac.Sum(nil)))) {
		return models.TokenClaims{}, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return models.TokenClaims{}, false
	}
	var claims models.TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return models.TokenClaims{}, false
	}
	if claims.ExpiresAt <= now.Unix() {
		return models.TokenClaims{}, false
	}
	return claims, true
}
