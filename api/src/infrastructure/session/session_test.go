package session

import (
	"testing"
	"time"
)

func TestManagerVerifiesSignedUser(t *testing.T) {
	manager := NewManager([]byte("01234567890123456789012345678901"))
	value, err := manager.Sign(User{Subject: "subject", Email: "user@example.com", EmailVerified: true})
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	user, err := manager.Verify(value)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if user.Subject != "subject" || user.Email != "user@example.com" || !user.EmailVerified {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestManagerRejectsTamperedAndExpiredValues(t *testing.T) {
	manager := NewManager([]byte("01234567890123456789012345678901"))
	value, err := manager.SignAt(User{Subject: "subject"}, time.Now().Add(-DefaultTTL-time.Second))
	if err != nil {
		t.Fatalf("SignAt() error = %v", err)
	}
	if _, err := manager.Verify(value); err == nil {
		t.Fatal("expected expired session to be rejected")
	}
	if _, err := manager.Verify(value + "tampered"); err == nil {
		t.Fatal("expected tampered session to be rejected")
	}
}
