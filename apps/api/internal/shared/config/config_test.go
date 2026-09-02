package config

import "testing"

func TestListenAddrLaptopDefault(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("SERVER_HOST", "")
	host, port := listenAddr()
	if host != "127.0.0.1" || port != "8080" {
		t.Fatalf("laptop default: got %s:%s", host, port)
	}
}

func TestListenAddrPaaSPort(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("SERVER_HOST", "")
	host, port := listenAddr()
	if host != "0.0.0.0" || port != "10000" {
		t.Fatalf("paas PORT: got %s:%s", host, port)
	}
}

func TestListenAddrExplicitHostWins(t *testing.T) {
	t.Setenv("PORT", "10000")
	t.Setenv("SERVER_HOST", "127.0.0.1")
	host, port := listenAddr()
	if host != "127.0.0.1" || port != "10000" {
		t.Fatalf("explicit host: got %s:%s", host, port)
	}
}

func TestMailPrefersResend(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("SMTP_HOST", "smtp.hostinger.com")
	cfg := Load()
	if cfg.Mail() != "resend" {
		t.Fatalf("mail: %s", cfg.Mail())
	}
}
