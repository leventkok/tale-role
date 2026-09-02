package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGraphQLHealthMeAndRoom(t *testing.T) {
	h, _ := setup(t)
	open := post(t, h, "/graphql", map[string]any{
		"query": "{ health { status persistence llm images mail } me { email } }",
	})
	if open.Code != http.StatusOK {
		t.Fatalf("health: %d %s", open.Code, open.Body.String())
	}
	if !bytes.Contains(open.Body.Bytes(), []byte(`"persistence"`)) || !bytes.Contains(open.Body.Bytes(), []byte(`"llm"`)) || !bytes.Contains(open.Body.Bytes(), []byte(`"mail":"none"`)) {
		t.Fatalf("health payload: %s", open.Body.String())
	}
	if bytes.Contains(open.Body.Bytes(), []byte("password")) || bytes.Contains(open.Body.Bytes(), []byte("$2a$")) {
		t.Fatal("hash leaked")
	}
	var anon struct {
		Data struct {
			Me *struct {
				Email string `json:"email"`
			} `json:"me"`
		} `json:"data"`
	}
	_ = json.Unmarshal(open.Body.Bytes(), &anon)
	if anon.Data.Me != nil {
		t.Fatalf("anonymous me: %s", open.Body.String())
	}

	bad := authed(t, h, http.MethodPost, "/graphql", "not-a-jwt", map[string]any{
		"query": "{ health { status } }",
	})
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token: %d %s", bad.Code, bad.Body.String())
	}

	token := registerAndVerify(t, h, "host@tale.role")
	me := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": "{ me { email verified } }",
	})
	if !bytes.Contains(me.Body.Bytes(), []byte(`"email":"host@tale.role"`)) {
		t.Fatalf("authed me: %s", me.Body.String())
	}
	created := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `mutation { createRoom(name: "Ashwood", joinMode: "link", diceSystem: "d20") { id diceSystem } }`,
	})
	if created.Code != http.StatusOK || !bytes.Contains(created.Body.Bytes(), []byte(`"id"`)) {
		t.Fatalf("createRoom: %d %s", created.Code, created.Body.String())
	}
	var mut struct {
		Data struct {
			CreateRoom struct {
				ID string `json:"id"`
			} `json:"createRoom"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	_ = json.Unmarshal(created.Body.Bytes(), &mut)
	if len(mut.Errors) > 0 || mut.Data.CreateRoom.ID == "" {
		t.Fatalf("mutation: %s", created.Body.String())
	}
	got := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query":     "query ($id: ID!) { room(id: $id) { id name themeId presence { role } } }",
		"variables": map[string]string{"id": mut.Data.CreateRoom.ID},
	})
	if bytes.Contains(got.Body.Bytes(), []byte("compiledPrompt")) || bytes.Contains(got.Body.Bytes(), []byte("compiled_prompt")) {
		t.Fatalf("compiled pack on room: %s", got.Body.String())
	}
	if bytes.Contains(got.Body.Bytes(), []byte("system_admin")) {
		t.Fatal("admin leak via graphql")
	}

	uni := authed(t, h, http.MethodPost, "/api/v1/universes", token, map[string]any{
		"name_en": "Ashwood", "theme_id": "fairytale",
	})
	if uni.Code != http.StatusCreated {
		t.Fatalf("universe: %d %s", uni.Code, uni.Body.String())
	}
	var createdUni struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(uni.Body.Bytes(), &createdUni)
	if createdUni.ID == "" {
		t.Fatalf("universe id: %s", uni.Body.String())
	}
	other := registerAndVerify(t, h, "player@tale.role")
	stolen := authed(t, h, http.MethodPost, "/graphql", other, map[string]any{
		"query":     "query ($id: ID!) { universe(id: $id) { compiledPrompt } }",
		"variables": map[string]string{"id": createdUni.ID},
	})
	if !bytes.Contains(stolen.Body.Bytes(), []byte("errors")) {
		t.Fatalf("expected forbidden universe: %s", stolen.Body.String())
	}
}
