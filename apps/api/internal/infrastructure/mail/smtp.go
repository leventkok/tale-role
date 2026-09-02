package mail

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTP delivers OTP mail. Empty Host is a no-op so tests and CI stay offline.
type SMTP struct {
	Host string
	Port string
	From string
	User string
	Pass string
}

func (s SMTP) Enabled() bool {
	return strings.TrimSpace(s.Host) != ""
}

func (s SMTP) Addr() string {
	port := s.Port
	if port == "" {
		port = "1025"
	}
	return net.JoinHostPort(s.Host, port)
}

func (s SMTP) SendOTP(email, code string) error {
	if !s.Enabled() {
		return nil
	}
	if strings.ContainsAny(email, "\r\n") || !strings.Contains(email, "@") {
		return fmt.Errorf("invalid recipient")
	}
	from := strings.TrimSpace(s.From)
	if from == "" {
		from = "noreply@talerole.local"
	}
	msg := otpMessage(from, email, code)
	var auth smtp.Auth
	if s.User != "" {
		auth = smtp.PlainAuth("", s.User, s.Pass, s.Host)
	}
	if err := smtp.SendMail(s.Addr(), auth, envelopeFrom(from), []string{email}, msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func envelopeFrom(from string) string {
	start := strings.LastIndex(from, "<")
	end := strings.LastIndex(from, ">")
	if start >= 0 && end > start {
		return strings.TrimSpace(from[start+1 : end])
	}
	return from
}

func otpPlainBody(code string) string {
	return "Your Tale Role sign-in code:\n\n" +
		code + "\n\n" +
		"It expires in 10 minutes. If you did not request this, ignore the email.\n\n" +
		"Tale Role doğrulama kodunuz yukarıdaki 6 hanedir. 10 dakika geçerlidir.\n"
}

func otpMessage(from, to, code string) []byte {
	body := strings.ReplaceAll(otpPlainBody(code), "\n", "\r\n")
	return []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Your Tale Role sign-in code\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body)
}
