package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
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
	h := httpapi.New(svc, game.NewTable(), nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg, "admin@tale.role")
	return h, svc
}

type failMailer struct{}

func (failMailer) SendOTP(string, string) error { return errors.New("smtp down") }

func TestRegisterMailFailure(t *testing.T) {
	store := memory.NewStore()
	svc := app.NewService(store, "test-secret", time.Hour, 10*time.Minute)
	svc.IssueOTP = func() (string, error) { return "123456", nil }
	svc.Mailer = failMailer{}
	cfg := config.Config{
		CORSAllowedOrigins: []string{"http://localhost:3000"},
		MaxBodyBytes:       1 << 20,
	}
	h := httpapi.New(svc, game.NewTable(), nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), cfg, "")
	reg := post(t, h, "/api/v1/auth/register", map[string]string{
		"email": "host@tale.role", "password": "longenough",
	})
	if reg.Code != http.StatusServiceUnavailable {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}
	if bytes.Contains(reg.Body.Bytes(), []byte("123456")) || bytes.Contains(reg.Body.Bytes(), []byte("smtp down")) {
		t.Fatalf("leaked: %s", reg.Body.String())
	}
}

func TestHealth(t *testing.T) {
	h, _ := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("live: %d", rec.Code)
	}
	ready := httptest.NewRecorder()
	h.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if ready.Code != http.StatusOK {
		t.Fatalf("ready: %d", ready.Code)
	}
	if !bytes.Contains(ready.Body.Bytes(), []byte(`"persistence":"memory"`)) || !bytes.Contains(ready.Body.Bytes(), []byte(`"llm":"stub"`)) || !bytes.Contains(ready.Body.Bytes(), []byte(`"mail":"none"`)) {
		t.Fatalf("scale signals: %s", ready.Body.String())
	}
}

func TestExportAndEraseAccount(t *testing.T) {
	h, _ := setup(t)
	token := registerAndVerify(t, h, "host@tale.role")
	authed(t, h, http.MethodPost, "/api/v1/licenses/register", token, map[string]string{
		"device_id": "desk-1", "platform": "win32",
	})
	created := authed(t, h, http.MethodPost, "/api/v1/universes", token, map[string]any{
		"name_en": "Ashwood", "theme_id": "fairytale",
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("universe: %d %s", created.Code, created.Body.String())
	}
	dump := authed(t, h, http.MethodGet, "/api/v1/me/export", token, nil)
	if dump.Code != http.StatusOK {
		t.Fatalf("export: %d %s", dump.Code, dump.Body.String())
	}
	if !bytes.Contains(dump.Body.Bytes(), []byte("host@tale.role")) || !bytes.Contains(dump.Body.Bytes(), []byte("desk-1")) {
		t.Fatalf("export missing subject: %s", dump.Body.String())
	}
	if bytes.Contains(dump.Body.Bytes(), []byte("password")) || bytes.Contains(dump.Body.Bytes(), []byte("$2a$")) || bytes.Contains(dump.Body.Bytes(), []byte("totp_secret")) {
		t.Fatal("export leaked password hash")
	}
	erased := authed(t, h, http.MethodDelete, "/api/v1/me", token, nil)
	if erased.Code != http.StatusOK {
		t.Fatalf("erase: %d %s", erased.Code, erased.Body.String())
	}
	if me := authed(t, h, http.MethodGet, "/api/v1/me", token, nil); me.Code != http.StatusUnauthorized {
		t.Fatalf("token after erase: %d %s", me.Code, me.Body.String())
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
	if !bytes.Contains(me.Body.Bytes(), []byte(`"totp_enabled":false`)) {
		t.Fatalf("me totp: %s", me.Body.String())
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

func TestTOTPLoginHTTP(t *testing.T) {
	h, _ := setup(t)
	token := registerAndVerify(t, h, "mfa@tale.role")
	begin := authed(t, h, http.MethodPost, "/api/v1/me/totp/begin", token, map[string]any{})
	if begin.Code != http.StatusOK {
		t.Fatalf("begin: %d %s", begin.Code, begin.Body.String())
	}
	var enrolled struct {
		Secret string `json:"secret"`
		URL    string `json:"otpauth_url"`
	}
	_ = json.Unmarshal(begin.Body.Bytes(), &enrolled)
	if enrolled.Secret == "" || !strings.Contains(enrolled.URL, "otpauth://") {
		t.Fatalf("enroll payload: %s", begin.Body.String())
	}
	code, err := app.CodeNow(enrolled.Secret)
	if err != nil {
		t.Fatal(err)
	}
	confirm := authed(t, h, http.MethodPost, "/api/v1/me/totp/confirm", token, map[string]string{"code": code})
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm: %d %s", confirm.Code, confirm.Body.String())
	}
	login := post(t, h, "/api/v1/auth/login", map[string]string{
		"email": "mfa@tale.role", "password": "longenough",
	})
	if login.Code != http.StatusUnauthorized {
		t.Fatalf("login: %d %s", login.Code, login.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(login.Body.Bytes(), &body)
	if body["mfa_required"] != true {
		t.Fatalf("expected mfa_required: %v", body)
	}
	if bytes.Contains(login.Body.Bytes(), []byte(`"token"`)) {
		t.Fatal("token on mfa challenge")
	}
	ok := post(t, h, "/api/v1/auth/totp/verify", map[string]string{
		"email": "mfa@tale.role", "password": "longenough", "code": code,
	})
	var tok struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(ok.Body.Bytes(), &tok)
	if ok.Code != http.StatusOK || tok.Token == "" {
		t.Fatalf("totp verify: %d %s", ok.Code, ok.Body.String())
	}
	dump := authed(t, h, http.MethodGet, "/api/v1/me/export", tok.Token, nil)
	if bytes.Contains(dump.Body.Bytes(), []byte(enrolled.Secret)) {
		t.Fatal("export leaked totp secret")
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

func TestStorytellerAfterEngineAndAdminTrace(t *testing.T) {
	h, _ := setup(t)
	token := registerAndVerify(t, h, "host@tale.role")
	created := authed(t, h, http.MethodPost, "/api/v1/rooms", token, map[string]string{
		"name": "Ashwood", "join_mode": "link", "dice_system": "d20",
	})
	var room struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body.Bytes(), &room)
	stats := map[string]int{"str": 3, "dex": 3, "con": 3, "int": 3, "wis": 3, "cha": 3}
	sheet := authed(t, h, http.MethodPost, "/api/v1/rooms/"+room.ID+"/characters", token, map[string]any{
		"name": "Iri", "stats": stats,
	})
	if sheet.Code != http.StatusCreated {
		t.Fatalf("character: %d %s", sheet.Code, sheet.Body.String())
	}
	if start := authed(t, h, http.MethodPost, "/api/v1/rooms/"+room.ID+"/start", token, map[string]string{}); start.Code != http.StatusOK {
		t.Fatalf("start: %d %s", start.Code, start.Body.String())
	}
	turn := authed(t, h, http.MethodPost, "/api/v1/rooms/"+room.ID+"/turns", token, map[string]any{
		"kind": "action", "skill": "str", "notes": "force the door, cc player@tale.role", "dc": 12, "locale": "en",
	})
	if turn.Code != http.StatusOK {
		t.Fatalf("turn: %d %s", turn.Code, turn.Body.String())
	}
	if !bytes.Contains(turn.Body.Bytes(), []byte(`"prose"`)) {
		t.Fatalf("expected storyteller prose: %s", turn.Body.String())
	}
	if bytes.Contains(turn.Body.Bytes(), []byte("mechanic_intent")) {
		t.Fatal("mechanic intent must not appear on the player turn")
	}
	var payload map[string]any
	_ = json.Unmarshal(turn.Body.Bytes(), &payload)
	narr, _ := payload["narrative"].(map[string]any)
	prose, _ := narr["prose"].(string)
	if strings.Contains(prose, "player@tale.role") {
		t.Fatalf("email in storyteller prose: %s", prose)
	}

	playerTraces := authed(t, h, http.MethodGet, "/api/v1/admin/traces", token, nil)
	if playerTraces.Code != http.StatusForbidden {
		t.Fatalf("player traces: %d", playerTraces.Code)
	}
	adminTok := registerAndVerify(t, h, "admin@tale.role")
	traces := authed(t, h, http.MethodGet, "/api/v1/admin/traces", adminTok, nil)
	if traces.Code != http.StatusOK {
		t.Fatalf("admin traces: %d %s", traces.Code, traces.Body.String())
	}
	if bytes.Contains(traces.Body.Bytes(), []byte("player@tale.role")) {
		t.Fatalf("pii in traces: %s", traces.Body.String())
	}
	swap := authed(t, h, http.MethodPut, "/api/v1/admin/runtime", adminTok, map[string]string{
		"prompt_pack": "v1-terse", "adapter_id": "stub",
	})
	if swap.Code != http.StatusOK {
		t.Fatalf("swap: %d %s", swap.Code, swap.Body.String())
	}
}

func TestUniverseWizardBindsThemeToRoom(t *testing.T) {
	h, _ := setup(t)
	token := registerAndVerify(t, h, "host@tale.role")
	created := authed(t, h, http.MethodPost, "/api/v1/universes", token, map[string]any{
		"name_en": "Ashwood", "theme_id": "gothic-horror", "era": "long night",
		"tone": "hushed", "taboos": "no real-world politics",
		"npcs": []map[string]string{{"name_en": "Warden", "alignment": "neutral", "voice": "dry"}},
	})
	if created.Code != http.StatusCreated {
		t.Fatalf("universe: %d %s", created.Code, created.Body.String())
	}
	if !bytes.Contains(created.Body.Bytes(), []byte("compiled_prompt")) || !bytes.Contains(created.Body.Bytes(), []byte("gothic-horror")) {
		t.Fatalf("expected compiled pack: %s", created.Body.String())
	}
	if bytes.Contains(created.Body.Bytes(), []byte("system_admin")) {
		t.Fatal("pack leaked system_admin")
	}
	var uni struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(created.Body.Bytes(), &uni)
	other := registerAndVerify(t, h, "player@tale.role")
	if stolen := authed(t, h, http.MethodGet, "/api/v1/universes/"+uni.ID, other, nil); stolen.Code != http.StatusForbidden {
		t.Fatalf("stolen universe: %d", stolen.Code)
	}
	room := authed(t, h, http.MethodPost, "/api/v1/rooms", token, map[string]string{
		"join_mode": "link", "universe_id": uni.ID,
	})
	if room.Code != http.StatusCreated {
		t.Fatalf("room: %d %s", room.Code, room.Body.String())
	}
	var createdRoom struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(room.Body.Bytes(), &createdRoom)
	got := authed(t, h, http.MethodGet, "/api/v1/rooms/"+createdRoom.ID, token, nil)
	if !bytes.Contains(got.Body.Bytes(), []byte("gothic-horror")) || bytes.Contains(got.Body.Bytes(), []byte("compiled_prompt")) {
		t.Fatalf("room snapshot: %s", got.Body.String())
	}
	if join := authed(t, h, http.MethodPost, "/api/v1/rooms/"+createdRoom.ID+"/join", other, map[string]string{}); join.Code != http.StatusOK {
		t.Fatalf("player join: %d %s", join.Code, join.Body.String())
	}
	playerView := authed(t, h, http.MethodGet, "/api/v1/rooms/"+createdRoom.ID, other, nil)
	if bytes.Contains(playerView.Body.Bytes(), []byte("compiled_prompt")) {
		t.Fatalf("player saw compiled prompt: %s", playerView.Body.String())
	}
	stats := map[string]int{"str": 3, "dex": 3, "con": 3, "int": 3, "wis": 3, "cha": 3}
	if sheet := authed(t, h, http.MethodPost, "/api/v1/rooms/"+createdRoom.ID+"/characters", token, map[string]any{
		"name": "Iri", "stats": stats,
	}); sheet.Code != http.StatusCreated {
		t.Fatalf("character: %d %s", sheet.Code, sheet.Body.String())
	}
	if start := authed(t, h, http.MethodPost, "/api/v1/rooms/"+createdRoom.ID+"/start", token, map[string]string{}); start.Code != http.StatusOK {
		t.Fatalf("start: %d %s", start.Code, start.Body.String())
	}
	turn := authed(t, h, http.MethodPost, "/api/v1/rooms/"+createdRoom.ID+"/turns", token, map[string]any{
		"kind": "action", "skill": "str", "notes": "knock, cc spy@tale.role", "dc": 12, "locale": "en",
	})
	if turn.Code != http.StatusOK {
		t.Fatalf("turn: %d %s", turn.Code, turn.Body.String())
	}
	if !bytes.Contains(turn.Body.Bytes(), []byte("[gothic-horror]")) {
		t.Fatalf("stub narrator should name the theme: %s", turn.Body.String())
	}
	if bytes.Contains(turn.Body.Bytes(), []byte("image_svg")) {
		t.Fatal("scene art must not inline in the chronicle turn")
	}
	deadline := time.Now().Add(2 * time.Second)
	var snap *httptest.ResponseRecorder
	var scene struct {
		ThemeID      string `json:"theme_id"`
		VisualPrompt string `json:"visual_prompt"`
		ImageSVG     string `json:"image_svg"`
		Inference    string `json:"inference"`
	}
	for time.Now().Before(deadline) {
		snap = authed(t, h, http.MethodGet, "/api/v1/rooms/"+createdRoom.ID, token, nil)
		var body struct {
			Scene *struct {
				ThemeID      string `json:"theme_id"`
				VisualPrompt string `json:"visual_prompt"`
				ImageSVG     string `json:"image_svg"`
				Inference    string `json:"inference"`
			} `json:"scene"`
		}
		_ = json.Unmarshal(snap.Body.Bytes(), &body)
		if body.Scene != nil && strings.Contains(body.Scene.ImageSVG, `data-theme="gothic-horror"`) {
			scene = *body.Scene
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if scene.ImageSVG == "" || scene.Inference != "stub" {
		t.Fatalf("expected stub scene on room: %s", snap.Body.String())
	}
	if strings.Contains(scene.VisualPrompt, "spy@tale.role") || strings.Contains(scene.ImageSVG, "system_admin") {
		t.Fatal("scene leaked pii or spectator")
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
