package mail

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendSendOTP(t *testing.T) {
	var gotAuth, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"test"}`))
	}))
	defer srv.Close()

	m := Resend{APIKey: "re_test", From: "Tale Role <onboarding@resend.dev>", BaseURL: srv.URL, Client: srv.Client()}
	if err := m.SendOTP("host@tale.role", "654321"); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer re_test" {
		t.Fatalf("auth: %s", gotAuth)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["from"] != "Tale Role <onboarding@resend.dev>" {
		t.Fatalf("from: %v", payload["from"])
	}
	if !strings.Contains(gotBody, "654321") {
		t.Fatal("code missing from payload")
	}
}

func TestResendRejectsBadRecipient(t *testing.T) {
	m := Resend{APIKey: "re_test"}
	if err := m.SendOTP("evil\r\nBcc: other@tale.role", "123456"); err == nil {
		t.Fatal("expected reject")
	}
}

func TestResendFromStripsNewlines(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"test"}`))
	}))
	defer srv.Close()
	m := Resend{
		APIKey:  "re_test",
		From:    "Tale Role <onboarding@\nresend.dev\n>",
		BaseURL: srv.URL,
		Client:  srv.Client(),
	}
	if err := m.SendOTP("host@tale.role", "111111"); err != nil {
		t.Fatal(err)
	}
	if got["from"] != "Tale Role <onboarding@resend.dev>" {
		t.Fatalf("from: %v", got["from"])
	}
}

func TestResendSurfacesHTTPStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"domain not verified"}`))
	}))
	defer srv.Close()
	m := Resend{APIKey: "re_test", BaseURL: srv.URL, Client: srv.Client()}
	err := m.SendOTP("host@tale.role", "123456")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("got %v", err)
	}
	if strings.Contains(err.Error(), "123456") {
		t.Fatal("otp leaked in error")
	}
}
