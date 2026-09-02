package gateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	gateway "github.com/leventkok/tale-role/services/llm-gateway"
	"github.com/leventkok/tale-role/services/llm-gateway/internal/pii"
)

func TestRedactsPIIAndOmitsAdminFromPrompt(t *testing.T) {
	svc := gateway.New()
	ok := true
	n := svc.Narrate(gateway.NarrateRequest{
		Locale:        "en",
		RoomID:        "r1",
		RoomName:      "Ashwood",
		ActorName:     "Iri",
		Kind:          "action",
		Notes:         "open the door, write player@tale.role",
		DiceSystem:    "d20",
		Rolls:         []int{14},
		Total:         17,
		Success:       &ok,
		PresenceNames: []string{"Iri", "Host"},
	})
	if strings.Contains(n.Prose, "player@tale.role") {
		t.Fatalf("email reached storyteller output: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, pii.Marker) && !strings.Contains(n.Prose, "Iri") {
		t.Fatalf("unexpected prose: %s", n.Prose)
	}
	themed := svc.Narrate(gateway.NarrateRequest{
		Locale: "en", RoomName: "Ashwood", ActorName: "Iri", Kind: "wait", ThemeID: "gothic-horror",
	})
	if !strings.Contains(themed.Prose, "[gothic-horror]") {
		t.Fatalf("theme missing from stub prose: %s", themed.Prose)
	}
	intent := svc.ProposeIntent(gateway.IntentRequest{
		Locale: "en",
		RoomID: "r1",
		Kind:   "action",
		Skill:  "str",
		Notes:  "ping 4111111111111111",
	})
	if strings.Contains(intent.Notes, "4111111111111111") {
		t.Fatal("card number leaked into mechanic intent")
	}
	for _, tr := range svc.Traces() {
		if pii.ContainsLeak(tr.RedactedPrompt) {
			t.Fatalf("trace prompt leaked pii: %s", tr.RedactedPrompt)
		}
		if strings.Contains(tr.RedactedPrompt, "system_admin") {
			t.Fatal("system_admin must not enter storyteller traces")
		}
	}
}

func TestLivePromptSwapChangesVoice(t *testing.T) {
	svc := gateway.New()
	a := svc.Narrate(gateway.NarrateRequest{Locale: "en", ActorName: "Iri", Kind: "wait", RoomName: "Ashwood"})
	if strings.Contains(a.Prose, "[v1-terse]") {
		t.Fatal("default pack should not be terse")
	}
	if err := svc.Swap("v1-terse", "stub"); err != nil {
		t.Fatal(err)
	}
	b := svc.Narrate(gateway.NarrateRequest{Locale: "en", ActorName: "Iri", Kind: "wait", RoomName: "Ashwood"})
	if !strings.Contains(b.Prose, "[v1-terse]") {
		t.Fatalf("swap did not change voice: %s", b.Prose)
	}
	rt := svc.Runtime()
	if rt.PromptPack != "v1-terse" || rt.AdapterID != "stub" || rt.Inference != "stub" {
		t.Fatalf("runtime: %+v", rt)
	}
}

func TestLocalAdapterRequiresWeightsOnDisk(t *testing.T) {
	svc := gateway.New()
	if err := svc.Swap("v1", "local"); err == nil {
		t.Fatal("local swap must fail without weights")
	}
	empty := t.TempDir()
	svc.ConfigureLocal(empty)
	if svc.Runtime().WeightsReady {
		t.Fatal("empty dir is not ready")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "adapter_config.json"), []byte(`{"id":"storyteller-v0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	svc.ConfigureLocal(dir)
	rt := svc.Runtime()
	if !rt.WeightsReady || !rt.AdapterDirConfigured || rt.AdapterID != "local" {
		t.Fatalf("expected local ready: %+v", rt)
	}
	if rt.Inference != "stub" {
		t.Fatal("inference stays stub until a runner URL is set")
	}
	svc.SetRunners("http://127.0.0.1:9", "")
	if svc.Runtime().Inference != "local" {
		t.Fatalf("expected local inference with runner url: %+v", svc.Runtime())
	}
}

func TestLocalRunnerNarrateAndFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "adapter_config.json"), []byte(`{"id":"storyteller-v0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	ok := true
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		var req gateway.NarrateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "en",
			Prose:  "Local Mira hits. Engine total " + strconv.Itoa(req.Total) + " for spy@tale.role.",
		})
	})
	mux.HandleFunc("/v1/intent", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.MechanicIntent{Kind: "action", Skill: "dex", DC: 14, Notes: "ok"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := gateway.New()
	svc.ConfigureLocal(dir)
	svc.SetRunners(srv.URL, srv.URL)
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "en", ActorName: "Mira", Kind: "action", Total: 17, Success: &ok, Notes: "kick",
	})
	if !strings.Contains(n.Prose, "Engine total 17") {
		t.Fatalf("runner prose: %s", n.Prose)
	}
	if strings.Contains(n.Prose, "spy@tale.role") {
		t.Fatal("runner email not redacted")
	}
	intent := svc.ProposeIntent(gateway.IntentRequest{Kind: "action", Skill: "str", Notes: "kick"})
	if intent.Skill != "dex" || intent.DC != 14 {
		t.Fatalf("runner intent: %+v", intent)
	}

	down := gateway.New()
	down.ConfigureLocal(dir)
	down.SetRunners("http://127.0.0.1:1", "")
	fallback := down.Narrate(gateway.NarrateRequest{Locale: "en", ActorName: "Mira", Kind: "wait", RoomName: "Ashwood"})
	if !strings.Contains(fallback.Prose, "The engine's dice") {
		t.Fatalf("expected stub fallback: %s", fallback.Prose)
	}
}
