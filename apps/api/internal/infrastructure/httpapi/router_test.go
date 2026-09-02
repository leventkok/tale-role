package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/application/app"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/httpapi"
	"github.com/leventkok/tale-role/apps/api/internal/infrastructure/memory"
	"github.com/leventkok/tale-role/apps/api/internal/shared/config"
)

func setup(t *testing.T) (http.Handler, *app.Service) {
	t.Helper()
	store := memory.NewStore()
	svc := app.NewService(store, "test-secret", time.Hour, 10*time.Minute)
	svc.IssueOTP = func() (string, error) { return "123456", nil }
	cfg := config.Config{
		CORSAllowedOrigins: []string{"http://localhost:3000"},
		MaxBodyBytes:       1 << 20,
	}
	h := httpapi.New(svc, game.NewTable(), slog.New(slog.NewTextHandler(io.Discard, nil)), cfg, "admin@tale.role")
	return h, svc
}

func TestHealth(t *testing.T) {
	h, _ := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("live: %d", rec.Code)
	}
}

func TestRegisterVerifyMeAndLicense(t *testing.T) {
	h, _ := setup(t)

	reg := post(t, h, "/api/v1/auth/register", map[string]string{
		"email": "Player@Tale.Role", "password": "longenough",
	})
	if reg.Code != http.StatusAccepted {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}
	if bytes.Contains(reg.Body.Bytes(), []byte("123456")) {
		t.Fatal("otp leaked in register response")
	}

	bad := post(t, h, "/api/v1/auth/otp/verify", map[string]string{
		"email": "player@tale.role", "code": "000000",
	})
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("bad otp: %d", bad.Code)
	}

	ok := post(t, h, "/api/v1/auth/otp/verify", map[string]string{
		"email": "player@tale.role", "code": "123456",
	})
	if ok.Code != http.StatusOK {
		t.Fatalf("verify: %d %s", ok.Code, ok.Body.String())
	}
	var tok struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(ok.Body.Bytes(), &tok); err != nil || tok.Token == "" {
		t.Fatalf("token: %v %s", err, ok.Body.String())
	}

	me := authed(t, h, http.MethodGet, "/api/v1/me", tok.Token, nil)
	if me.Code != http.StatusOK {
		t.Fatalf("me: %d %s", me.Code, me.Body.String())
	}
	if bytes.Contains(me.Body.Bytes(), []byte("system_admin")) {
		t.Fatal("admin spectator must not appear on player me")
	}

	lic := authed(t, h, http.MethodPost, "/api/v1/licenses/register", tok.Token, map[string]string{
		"device_id": "desk-1", "platform": "win32",
	})
	if lic.Code != http.StatusCreated {
		t.Fatalf("license: %d %s", lic.Code, lic.Body.String())
	}
}

func TestLoginRequiresOTPUntilVerified(t *testing.T) {
	h, _ := setup(t)
	post(t, h, "/api/v1/auth/register", map[string]string{
		"email": "gm@tale.role", "password": "longenough",
	})
	login := post(t, h, "/api/v1/auth/login", map[string]string{
		"email": "gm@tale.role", "password": "longenough",
	})
	if login.Code != http.StatusUnauthorized {
		t.Fatalf("login before verify: %d", login.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(login.Body.Bytes(), &body)
	if body["otp_required"] != true {
		t.Fatalf("expected otp_required, got %v", body)
	}
}

func TestGenericInternalErrorShape(t *testing.T) {
	h, _ := setup(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json: %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("panic")) {
		t.Fatal("must not leak internals")
	}
}

func TestRoomHTTPHidesAdmin(t *testing.T) {
	h, _ := setup(t)
	token := registerAndVerify(t, h, "host@tale.role")
	created := authed(t, h, http.MethodPost, "/api/v1/rooms", token, map[string]string{
		"name": "Ashwood", "join_mode": "link", "dice_system": "d20",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("create room: %d %s", created.Code, created.Body.String())
	}
	var room struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body.Bytes(), &room)
	adminTok := registerAndVerify(t, h, "admin@tale.role")
	join := authed(t, h, http.MethodPost, "/api/v1/rooms/"+room.ID+"/join", adminTok, map[string]string{})
	if join.Code != http.StatusOK {
		t.Fatalf("admin join: %d %s", join.Code, join.Body.String())
	}
	got := authed(t, h, http.MethodGet, "/api/v1/rooms/"+room.ID, token, nil)
	if bytes.Contains(got.Body.Bytes(), []byte("system_admin")) {
		t.Fatalf("admin leaked: %s", got.Body.String())
	}
}

func registerAndVerify(t *testing.T, h http.Handler, email string) string {
	t.Helper()
	post(t, h, "/api/v1/auth/register", map[string]string{"email": email, "password": "longenough"})
	ok := post(t, h, "/api/v1/auth/otp/verify", map[string]string{"email": email, "code": "123456"})
	var tok struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(ok.Body.Bytes(), &tok)
	if tok.Token == "" {
		t.Fatalf("no token for %s: %s", email, ok.Body.String())
	}
	return tok.Token
}

func post(t *testing.T, h http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func authed(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
