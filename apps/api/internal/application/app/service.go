package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
	"github.com/leventkok/tale-role/apps/api/internal/domain/license"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOTPRequired        = errors.New("otp required")
	ErrOTPInvalid         = errors.New("invalid otp")
	ErrEmailTaken         = errors.New("email taken")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrMailFailed         = errors.New("mail delivery failed")
)

type Service struct {
	store     Identity
	jwtSecret []byte
	jwtExpiry time.Duration
	otpTTL    time.Duration
	// IssueOTP, when set, overrides random OTP generation (tests / TALEROLE_DEV_OTP).
	IssueOTP func() (string, error)
	// Mailer sends the plaintext OTP. Nil skips delivery (CI without SMTP).
	Mailer Mailer
}

func NewService(store Identity, jwtSecret string, jwtExpiry, otpTTL time.Duration) *Service {
	return &Service{
		store:     store,
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
		otpTTL:    otpTTL,
	}
}

func (s *Service) Register(email, password string) error {
	email = normalizeEmail(email)
	if email == "" || len(password) < 8 {
		return ErrInvalidCredentials
	}
	if _, ok := s.store.GetUser(email); ok {
		return ErrEmailTaken
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.store.PutUser(&iam.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		Verified:     false,
		CreatedAt:    time.Now().UTC(),
	})
	return s.issueOTP(email)
}

func (s *Service) Login(email, password string) (token string, err error) {
	email = normalizeEmail(email)
	u, ok := s.store.GetUser(email)
	if !ok {
		return "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	if !u.Verified {
		if err := s.issueOTP(email); err != nil {
			return "", err
		}
		return "", ErrOTPRequired
	}
	return s.sign(u)
}

func (s *Service) VerifyOTP(email, code string) (token string, err error) {
	email = normalizeEmail(email)
	u, ok := s.store.GetUser(email)
	if !ok {
		return "", ErrOTPInvalid
	}
	otp, ok := s.store.GetOTP(email)
	if !ok {
		return "", ErrOTPInvalid
	}
	if otp.Attempts >= 5 {
		s.store.DeleteOTP(email)
		return "", ErrOTPInvalid
	}
	otp.Attempts++
	s.store.PutOTP(otp)
	if bcrypt.CompareHashAndPassword(otp.Hash, []byte(code)) != nil {
		return "", ErrOTPInvalid
	}
	u.Verified = true
	s.store.PutUser(u)
	s.store.DeleteOTP(email)
	return s.sign(u)
}

func (s *Service) UserFromToken(token string) (*iam.User, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrUnauthorized
		}
		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrUnauthorized
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}
	sub, _ := claims["sub"].(string)
	u, ok := s.store.GetUserByID(sub)
	if !ok {
		return nil, ErrUnauthorized
	}
	return u, nil
}

func (s *Service) RegisterLicense(userID, deviceID, platform string) (*license.ProductLicense, error) {
	if userID == "" || deviceID == "" {
		return nil, ErrInvalidCredentials
	}
	l := &license.ProductLicense{
		ID:        uuid.NewString(),
		UserID:    userID,
		DeviceID:  deviceID,
		Platform:  platform,
		CreatedAt: time.Now().UTC(),
	}
	s.store.PutLicense(l)
	return l, nil
}

func (s *Service) Licenses(userID string) []*license.ProductLicense {
	return s.store.LicensesForUser(userID)
}

func (s *Service) ExportSubject(userID string) (map[string]any, error) {
	u, ok := s.store.GetUserByID(userID)
	if !ok {
		return nil, ErrUnauthorized
	}
	return map[string]any{
		"id":         u.ID,
		"email":      u.Email,
		"verified":   u.Verified,
		"created_at": u.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) Erase(userID string) error {
	if _, ok := s.store.GetUserByID(userID); !ok {
		return ErrUnauthorized
	}
	s.store.DeleteUserByID(userID)
	return nil
}

func (s *Service) issueOTP(email string) error {
	issue := s.IssueOTP
	if issue == nil {
		issue = randomOTP
	}
	code, err := issue()
	if err != nil {
		return err
	}
	if s.Mailer != nil {
		if err := s.Mailer.SendOTP(email, code); err != nil {
			return fmt.Errorf("%w: %v", ErrMailFailed, err)
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.store.PutOTP(&iam.OTP{
		Email:     email,
		Hash:      hash,
		ExpiresAt: time.Now().Add(s.otpTTL),
	})
	return nil
}

func (s *Service) sign(u *iam.User) (string, error) {
	now := time.Now()
	jti := make([]byte, 8)
	_, _ = rand.Read(jti)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID,
		"eml": u.Email,
		"jti": hex.EncodeToString(jti),
		"iat": now.Unix(),
		"exp": now.Add(s.jwtExpiry).Unix(),
	})
	return tok.SignedString(s.jwtSecret)
}

func randomOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
