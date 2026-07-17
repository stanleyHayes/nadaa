package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/auth-service/internal/config"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/auth-service/internal/utils"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	store        store.Store
	tokenSecret  []byte
	otp          utils.OTPGenerator
	now          func() time.Time
	exposeDevOTP bool
	config       *config.Config
}

// NewServer creates a new Server with the given dependencies.
func NewServer(s store.Store, now func() time.Time, cfg *config.Config) *Server {
	var otp utils.OTPGenerator = utils.RandomOTPGenerator{}
	if cfg.MockOTP != "" {
		otp = utils.FixedOTPGenerator{Code: cfg.MockOTP}
	}

	return &Server{
		store:        s,
		tokenSecret:  []byte(cfg.TokenSecret),
		otp:          otp,
		now:          now,
		exposeDevOTP: cfg.ExposeDevOTP,
		config:       cfg,
	}
}

func (s *Server) signToken(profile models.CitizenProfile, expiresAt time.Time) (string, error) {
	claims := models.TokenClaims{
		UserID:    profile.ID,
		UserType:  models.RoleCitizen,
		Phone:     profile.Phone,
		Role:      profile.Role,
		ExpiresAt: expiresAt.Unix(),
	}
	return s.signClaims(claims)
}

func (s *Server) signAgencyToken(profile models.AgencyUserProfile, expiresAt time.Time) (string, error) {
	claims := models.TokenClaims{
		UserID:    profile.ID,
		UserType:  "agency",
		Email:     profile.Email,
		Phone:     profile.Phone,
		Role:      profile.Role,
		AgencyID:  profile.Agency.ID,
		District:  profile.Agency.District,
		MFA:       profile.MFAEnabled,
		ExpiresAt: expiresAt.Unix(),
	}
	return s.signClaims(claims)
}

func (s *Server) signClaims(claims models.TokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := s.sign(encodedPayload)
	return "nadaa." + encodedPayload + "." + signature, nil
}

func (s *Server) verifyToken(token string) (models.TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "nadaa" {
		return models.TokenClaims{}, store.ErrInvalidToken
	}

	expectedSignature := s.sign(parts[1])
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return models.TokenClaims{}, store.ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return models.TokenClaims{}, store.ErrInvalidToken
	}

	var claims models.TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return models.TokenClaims{}, store.ErrInvalidToken
	}

	if claims.ExpiresAt <= s.now().Unix() {
		return models.TokenClaims{}, store.ErrInvalidToken
	}

	return claims, nil
}

func (s *Server) sign(payload string) string {
	mac := hmac.New(sha256.New, s.tokenSecret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
