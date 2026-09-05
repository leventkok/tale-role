package app

import (
	"errors"
	"testing"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
)

func TestRegisterLicenseIdempotentAndPlayGate(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	svc.IssueOTP = func() (string, error) { return "123456", nil }
	if err := svc.Register("host@tale.role", "longenough"); err != nil {
		t.Fatal(err)
	}
	tok, err := svc.VerifyOTP("host@tale.role", "123456")
	if err != nil {
		t.Fatal(err)
	}
	u, err := svc.UserFromToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RequireDesktopLicense(u.ID, ""); err != nil {
		t.Fatalf("browser play: %v", err)
	}
	if err := svc.RequireDesktopLicense(u.ID, "desk-1"); !errors.Is(err, ErrLicenseRequired) {
		t.Fatalf("want license required, got %v", err)
	}
	first, err := svc.RegisterLicense(u.ID, "desk-1", "win32")
	if err != nil || first.ID == "" {
		t.Fatalf("register: %v", err)
	}
	again, err := svc.RegisterLicense(u.ID, "desk-1", "win32")
	if err != nil || again.ID != first.ID {
		t.Fatalf("idempotent: %s %s %v", first.ID, again.ID, err)
	}
	if len(svc.Licenses(u.ID)) != 1 {
		t.Fatalf("duplicate licenses: %d", len(svc.Licenses(u.ID)))
	}
	if err := svc.RequireDesktopLicense(u.ID, "desk-1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RevokeLicense(u.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.RequireDesktopLicense(u.ID, "desk-1"); !errors.Is(err, ErrLicenseRequired) {
		t.Fatalf("revoked still licensed: %v", err)
	}
	if err := svc.RevokeLicense(u.ID, first.ID); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("second revoke: %v", err)
	}
	again2, err := svc.RegisterLicense(u.ID, "desk-1", "win32")
	if err != nil || again2.ID == "" {
		t.Fatalf("reregister: %v", err)
	}
	if n := len(svc.Licenses(u.ID)); n != 1 {
		t.Fatalf("reregister duplicates: %d", n)
	}
}
