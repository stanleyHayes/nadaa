package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/missing-person-service/internal/models"
)

// verifyBearerToken verifies a NADAA access token of the form
// "nadaa.<payload>.<sig>" issued by auth-service, where payload is
// base64.RawURLEncoding(JSON claims) and sig is
// base64.RawURLEncoding(HMAC-SHA256(key=secret, message=payload)).
// It mirrors auth-service's signing scheme; keep the two in sync.
func verifyBearerToken(r *http.Request, secret string, now time.Time) (models.TokenClaims, bool) {
	if secret == "" {
		return models.TokenClaims{}, false
	}
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found {
		return models.TokenClaims{}, false
	}
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return models.TokenClaims{}, false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
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
