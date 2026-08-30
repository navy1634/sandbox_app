package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/sandbox-nextjs/src/infrastructure/oidc"
)

const (
	DefaultTTL     = 24 * time.Hour
	TransactionTTL = 5 * time.Minute
)

type User struct {
	Subject       string `json:"sub"`
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
}

type signedUser struct {
	Kind      string `json:"kind"`
	User      User   `json:"user"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	JTI       string `json:"jti"`
}

type signedTransaction struct {
	Kind      string                    `json:"kind"`
	Request   oidc.AuthorizationRequest `json:"request"`
	IssuedAt  int64                     `json:"iat"`
	ExpiresAt int64                     `json:"exp"`
	JTI       string                    `json:"jti"`
}

type Manager struct {
	secret []byte
}

func NewManager(secret []byte) *Manager {
	return &Manager{secret: append([]byte(nil), secret...)}
}

func (m *Manager) Sign(user User) (string, error) {
	return m.signUserAt(user, time.Now())
}

func (m *Manager) SignAt(user User, now time.Time) (string, error) {
	return m.signUserAt(user, now)
}

func (m *Manager) Verify(value string) (User, error) {
	var signed signedUser
	if err := m.verify(value, &signed); err != nil {
		return User{}, err
	}
	if signed.Kind != "user" || signed.User.Subject == "" {
		return User{}, errors.New("invalid user session")
	}
	return signed.User, nil
}

func (m *Manager) SignTransaction(request oidc.AuthorizationRequest) (string, error) {
	return m.signTransactionAt(request, time.Now())
}

func (m *Manager) VerifyTransaction(value string) (oidc.AuthorizationRequest, error) {
	var signed signedTransaction
	if err := m.verify(value, &signed); err != nil {
		return oidc.AuthorizationRequest{}, err
	}
	if signed.Kind != "transaction" || signed.Request.State == "" || signed.Request.Nonce == "" || signed.Request.CodeVerifier == "" {
		return oidc.AuthorizationRequest{}, errors.New("invalid OIDC transaction")
	}
	return signed.Request, nil
}

func (m *Manager) signUserAt(user User, now time.Time) (string, error) {
	return m.encode(signedUser{
		Kind:      "user",
		User:      user,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(DefaultTTL).Unix(),
		JTI:       randomID(),
	})
}

func (m *Manager) signTransactionAt(request oidc.AuthorizationRequest, now time.Time) (string, error) {
	return m.encode(signedTransaction{
		Kind:      "transaction",
		Request:   request,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(TransactionTTL).Unix(),
		JTI:       randomID(),
	})
}

func (m *Manager) encode(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func (m *Manager) verify(value string, target any) error {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return errors.New("invalid session format")
	}
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(expected, actual) {
		return errors.New("invalid session signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return err
	}
	var expiry struct {
		ExpiresAt int64  `json:"exp"`
		JTI       string `json:"jti"`
	}
	if err := json.Unmarshal(payload, &expiry); err != nil {
		return err
	}
	if expiry.ExpiresAt <= time.Now().Unix() {
		return errors.New("session expired")
	}
	if expiry.JTI == "" {
		return errors.New("session id is missing")
	}
	return nil
}

func randomID() string {
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(value)
}
