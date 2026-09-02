package mail

import (
	"strings"
	"testing"
)

func TestOTPMessageKeepsCodeInBody(t *testing.T) {
	msg := string(otpMessage("Tale Role <noreply@talerole.local>", "host@tale.role", "654321"))
	if !strings.Contains(msg, "Subject: Your Tale Role sign-in code") {
		t.Fatal("subject")
	}
	if strings.Contains(strings.SplitN(msg, "\r\n\r\n", 2)[0], "654321") {
		t.Fatal("code in headers")
	}
	if !strings.Contains(msg, "654321") {
		t.Fatal("code missing from body")
	}
	if envelopeFrom("Tale Role <noreply@talerole.local>") != "noreply@talerole.local" {
		t.Fatal("envelope from")
	}
}

func TestSendOTPNoopWithoutHost(t *testing.T) {
	if err := (SMTP{}).SendOTP("host@tale.role", "123456"); err != nil {
		t.Fatal(err)
	}
}

func TestSendOTPRejectsHeaderInjection(t *testing.T) {
	s := SMTP{Host: "127.0.0.1", Port: "1025", From: "noreply@talerole.local"}
	if err := s.SendOTP("evil\r\nBcc: other@tale.role", "123456"); err == nil {
		t.Fatal("expected reject")
	}
}
