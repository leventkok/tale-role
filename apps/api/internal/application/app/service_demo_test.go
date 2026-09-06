package app

import (
	"errors"
	"testing"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
)

func TestEnsureDemoUserNoopWhenUnset(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	if err := svc.EnsureDemoUser("", ""); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.GetUser("test@talerole.com"); ok {
		t.Fatal("unexpected user")
	}
}

func TestEnsureDemoUserSignsInWithoutOTP(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	mailer := &recordingMailer{}
	svc.Mailer = mailer
	if err := svc.EnsureDemoUser("Test@TaleRole.com", "MasterFabricTest123"); err != nil {
		t.Fatal(err)
	}
	tok, err := svc.Login("test@talerole.com", "MasterFabricTest123")
	if err != nil || tok == "" {
		t.Fatalf("login: %v", err)
	}
	if mailer.code != "" {
		t.Fatal("otp mailed for verified demo user")
	}
}

func TestEnsureDemoUserRejectsPartialConfig(t *testing.T) {
	svc := NewService(memory.NewStore(), "test-secret", time.Hour, 10*time.Minute)
	if err := svc.EnsureDemoUser("test@talerole.com", ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("got %v", err)
	}
}

func TestEnsureDemoUserPromotesExisting(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	svc.IssueOTP = func() (string, error) { return "123456", nil }
	if err := svc.Register("test@talerole.com", "oldpassword"); err != nil {
		t.Fatal(err)
	}
	if err := svc.EnsureDemoUser("test@talerole.com", "newpassword"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login("test@talerole.com", "oldpassword"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("old password: %v", err)
	}
	tok, err := svc.Login("test@talerole.com", "newpassword")
	if err != nil || tok == "" {
		t.Fatalf("new password: %v", err)
	}
}
