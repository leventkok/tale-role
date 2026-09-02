package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const resendDefaultURL = "https://api.resend.com/emails"

// Resend delivers OTP over HTTPS. Render's free plan blocks SMTP ports;
// this path uses 443 so Hostinger SMTP is not required at runtime.
type Resend struct {
	APIKey  string
	From    string
	BaseURL string
	Client  *http.Client
}

func (r Resend) Enabled() bool {
	return strings.TrimSpace(r.APIKey) != ""
}

func (r Resend) SendOTP(email, code string) error {
	if !r.Enabled() {
		return fmt.Errorf("resend api key unset")
	}
	if strings.ContainsAny(email, "\r\n") || !strings.Contains(email, "@") {
		return fmt.Errorf("invalid recipient")
	}
	from := strings.TrimSpace(r.From)
	if from == "" {
		from = "Tale Role <onboarding@resend.dev>"
	}
	payload, err := json.Marshal(map[string]any{
		"from":    from,
		"to":      []string{email},
		"subject": "Your Tale Role sign-in code",
		"text":    otpPlainBody(code),
	})
	if err != nil {
		return err
	}
	url := r.BaseURL
	if url == "" {
		url = resendDefaultURL
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")
	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend status %d", resp.StatusCode)
	}
	return nil
}
