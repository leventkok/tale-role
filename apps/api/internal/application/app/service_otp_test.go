package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
)

type recordingMailer struct {
	email string
	code  string
	err   error
}

func (m *recordingMailer) SendOTP(email, code string) error {
	m.email = email
	m.code = code
	return m.err
}

func TestIssueOTPMailsThenStoresHash(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	mailer := &recordingMailer{}
	svc.Mailer = mailer
	svc.IssueOTP = func() (string, error) { return "654321", nil }

	if err := svc.Register("Host@Tale.Role", "longenough"); err != nil {
		t.Fatal(err)
	}
	if mailer.email != "host@tale.role" || mailer.code != "654321" {
		t.Fatalf("mailed %+v", mailer)
	}
	otp, ok := store.GetOTP("host@tale.role")
	if !ok || len(otp.Hash) == 0 {
		t.Fatal("otp not stored")
	}
	if strings.Contains(string(otp.Hash), "654321") {
		t.Fatal("plaintext otp stored")
	}
	tok, err := svc.VerifyOTP("host@tale.role", "654321")
	if err != nil || tok == "" {
		t.Fatalf("verify: %v", err)
	}
}

func TestIssueOTPDoesNotStoreWhenMailFails(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	svc.Mailer = &recordingMailer{err: errors.New("smtp down")}
	svc.IssueOTP = func() (string, error) { return "111222", nil }

	err := svc.Register("player@tale.role", "longenough")
	if !errors.Is(err, ErrMailFailed) {
		t.Fatalf("want ErrMailFailed, got %v", err)
	}
	if _, ok := store.GetOTP("player@tale.role"); ok {
		t.Fatal("otp stored after mail failure")
	}
}
