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

func TestGraphQLTableAndAccountMutations(t *testing.T) {
	h, _ := setup(t)
	token := registerAndVerify(t, h, "host@tale.role")
	uni := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `mutation { createUniverse(nameEn: "Ashwood", themeId: "fairytale") { id nameEn compiledPrompt } }`,
	})
	var uniBody struct {
		Data struct {
			CreateUniverse struct {
				ID             string `json:"id"`
				CompiledPrompt string `json:"compiledPrompt"`
			} `json:"createUniverse"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	_ = json.Unmarshal(uni.Body.Bytes(), &uniBody)
	if len(uniBody.Errors) > 0 || uniBody.Data.CreateUniverse.ID == "" || uniBody.Data.CreateUniverse.CompiledPrompt == "" {
		t.Fatalf("createUniverse: %s", uni.Body.String())
	}
	created := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query":     `mutation ($id: ID!) { createRoom(name: "Table", joinMode: "link", universeId: $id) { id } }`,
		"variables": map[string]string{"id": uniBody.Data.CreateUniverse.ID},
	})
	var roomBody struct {
		Data struct {
			CreateRoom struct {
				ID string `json:"id"`
			} `json:"createRoom"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	_ = json.Unmarshal(created.Body.Bytes(), &roomBody)
	if len(roomBody.Errors) > 0 || roomBody.Data.CreateRoom.ID == "" {
		t.Fatalf("createRoom: %s", created.Body.String())
	}
	sheet := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query":     `mutation ($id: ID!) { setCharacter(roomId: $id, name: "Wren", stats: {str: 3, dex: 3, con: 3, int: 3, wis: 3, cha: 3}) }`,
		"variables": map[string]string{"id": roomBody.Data.CreateRoom.ID},
	})
	if bytes.Contains(sheet.Body.Bytes(), []byte(`"errors"`)) {
		t.Fatalf("setCharacter: %s", sheet.Body.String())
	}
	start := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query":     `mutation ($id: ID!) { startRoom(roomId: $id) }`,
		"variables": map[string]string{"id": roomBody.Data.CreateRoom.ID},
	})
	if bytes.Contains(start.Body.Bytes(), []byte(`"errors"`)) {
		t.Fatalf("startRoom: %s", start.Body.String())
	}
	act := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query":     `mutation ($id: ID!) { actTurn(roomId: $id, kind: "action", skill: "str", notes: "knock") { kind total prose } }`,
		"variables": map[string]string{"id": roomBody.Data.CreateRoom.ID},
	})
	if bytes.Contains(act.Body.Bytes(), []byte(`"errors"`)) || !bytes.Contains(act.Body.Bytes(), []byte(`"kind"`)) {
		t.Fatalf("actTurn: %s", act.Body.String())
	}
	snap := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query":     `query ($id: ID!) { room(id: $id) { started turnOrder characters { name stats { str } } turns { actorId kind rolls total prose } } }`,
		"variables": map[string]string{"id": roomBody.Data.CreateRoom.ID},
	})
	if bytes.Contains(snap.Body.Bytes(), []byte(`"errors"`)) || !bytes.Contains(snap.Body.Bytes(), []byte(`"Wren"`)) || !bytes.Contains(snap.Body.Bytes(), []byte(`"turns"`)) {
		t.Fatalf("room snapshot: %s", snap.Body.String())
	}
	if bytes.Contains(snap.Body.Bytes(), []byte("password")) || bytes.Contains(snap.Body.Bytes(), []byte("compiledPrompt")) {
		t.Fatalf("room leaked: %s", snap.Body.String())
	}
	listed := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `{ licenses { deviceId } }`,
	})
	if bytes.Contains(listed.Body.Bytes(), []byte(`"errors"`)) {
		t.Fatalf("licenses empty query: %s", listed.Body.String())
	}
	lic := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `mutation { registerLicense(deviceId: "desk-gql", platform: "win32") { id deviceId } }`,
	})
	if bytes.Contains(lic.Body.Bytes(), []byte(`"errors"`)) {
		t.Fatalf("registerLicense: %s", lic.Body.String())
	}
	erased := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `{ me { totpEnabled } }`,
	})
	if !bytes.Contains(erased.Body.Bytes(), []byte(`"totpEnabled":false`)) {
		t.Fatalf("me totp: %s", erased.Body.String())
	}
	gone := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `mutation { eraseMe }`,
	})
	if bytes.Contains(gone.Body.Bytes(), []byte(`"errors"`)) {
		t.Fatalf("eraseMe: %s", gone.Body.String())
	}
	after := authed(t, h, http.MethodPost, "/graphql", token, map[string]any{
		"query": `{ me { email } }`,
	})
	if after.Code != http.StatusUnauthorized && !bytes.Contains(after.Body.Bytes(), []byte(`"me":null`)) {
		t.Fatalf("me after erase: %d %s", after.Code, after.Body.String())
	}
}
