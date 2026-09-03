package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
)

func TestTOTPEnrollThenLogin(t *testing.T) {
	store := memory.NewStore()
	svc := NewService(store, "test-secret", time.Hour, 10*time.Minute)
	svc.IssueOTP = func() (string, error) { return "123456", nil }
	if err := svc.Register("host@tale.role", "longenough"); err != nil {
		t.Fatal(err)
	}
	tok, err := svc.VerifyOTP("host@tale.role", "123456")
	if err != nil || tok == "" {
		t.Fatalf("verify: %v", err)
	}
	u, err := svc.UserFromToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	secret, uri, err := svc.BeginTOTP(u.ID)
	if err != nil || secret == "" || !strings.HasPrefix(uri, "otpauth://totp/") {
		t.Fatalf("begin: %v %s", err, uri)
	}
	code, err := totpAt(secret, time.Now().Unix()/30)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmTOTP(u.ID, code); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login("host@tale.role", "longenough"); !errors.Is(err, ErrMFARequired) {
		t.Fatalf("want MFA, got %v", err)
	}
	if _, err := svc.LoginTOTP("host@tale.role", "longenough", "000000"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("bad code: %v", err)
	}
	tok2, err := svc.LoginTOTP("host@tale.role", "longenough", code)
	if err != nil || tok2 == "" {
		t.Fatalf("totp login: %v", err)
	}
	dump, err := svc.ExportSubject(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dump["totp_enabled"] != true {
		t.Fatalf("export totp: %#v", dump)
	}
	raw, _ := json.Marshal(dump)
	if bytes.Contains(raw, []byte(secret)) || bytes.Contains(raw, []byte("totp_secret")) {
		t.Fatal("totp secret leaked in export")
	}
	if err := svc.DisableTOTP(u.ID, code); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login("host@tale.role", "longenough"); err != nil {
		t.Fatalf("login after disable: %v", err)
	}
}

func TestRFC6238Vector(t *testing.T) {
	got, err := totpAt("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", 1)
	if err != nil || got != "287082" {
		t.Fatalf("rfc6238: %s %v", got, err)
	}
}
