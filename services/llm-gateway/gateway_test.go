package gateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	if strings.Contains(themed.Prose, "[gothic-horror]") {
		t.Fatalf("theme id must not leak into player prose: %s", themed.Prose)
	}
	if strings.Contains(strings.ToLower(themed.Prose), "engine") || strings.Contains(themed.Prose, "d20") {
		t.Fatalf("stub must stay literary: %s", themed.Prose)
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

func TestHubAdapterRequiresModelIDs(t *testing.T) {
	svc := gateway.New()
	if err := svc.Swap("v1", "hub"); err == nil {
		t.Fatal("hub swap must fail without model ids")
	}
	svc.ConfigureHub("", "")
	if svc.Runtime().WeightsReady {
		t.Fatal("empty hub is not ready")
	}
	svc.ConfigureHub("your-org/talerole-storyteller", "your-org/talerole-mechanics")
	rt := svc.Runtime()
	if !rt.WeightsReady || !rt.AdapterDirConfigured || rt.AdapterID != "hub" {
		t.Fatalf("expected hub ready: %+v", rt)
	}
	if rt.Inference != "stub" {
		t.Fatal("inference stays stub until a runner URL is set")
	}
	svc.SetRunners("http://127.0.0.1:9", "")
	if svc.Runtime().Inference != "hub" {
		t.Fatalf("expected hub inference with runner url: %+v", svc.Runtime())
	}
}

func TestReplicaFailsOverAndPackOverride(t *testing.T) {
	good := http.NewServeMux()
	good.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{Locale: "en", Prose: "replica two spoke"})
	})
	good.HandleFunc("/v1/intent", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.MechanicIntent{Kind: "action", Skill: "cha", DC: 11})
	})
	live := httptest.NewServer(good)
	t.Cleanup(live.Close)
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(dead.Close)

	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "your-org/talerole-mechanics")
	svc.SetRunners(dead.URL+","+live.URL, dead.URL+","+live.URL)
	n := svc.Narrate(gateway.NarrateRequest{Locale: "en", ActorName: "Mira", Kind: "wait"})
	if !strings.Contains(n.Prose, "replica two spoke") {
		t.Fatalf("failover prose: %s", n.Prose)
	}
	intent := svc.ProposeIntent(gateway.IntentRequest{Kind: "action", Skill: "str"})
	if intent.Skill != "cha" {
		t.Fatalf("failover intent: %+v", intent)
	}
	if err := svc.PutPack("v1", "custom-en-voice never invent dice", "özel-tr"); err != nil {
		t.Fatal(err)
	}
	docs := svc.Packs()
	if len(docs) != 2 || !strings.Contains(docs[0].EN, "custom-en-voice") {
		t.Fatalf("packs: %+v", docs)
	}
}

func TestHubRunnerNarrateAndFallback(t *testing.T) {
	ok := true
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		var req gateway.NarrateRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "en",
			Prose:  "Hub Mira hits. Engine total " + strconv.Itoa(req.Total) + " for spy@tale.role.",
		})
	})
	mux.HandleFunc("/v1/intent", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.MechanicIntent{Kind: "action", Skill: "dex", DC: 14, Notes: "ok"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "your-org/talerole-mechanics")
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
	down.ConfigureHub("your-org/talerole-storyteller", "")
	down.SetRunners("http://127.0.0.1:1", "")
	fallback := down.Narrate(gateway.NarrateRequest{Locale: "en", ActorName: "Mira", Kind: "wait", RoomName: "Ashwood"})
	if !strings.Contains(fallback.Prose, "The lantern holds") {
		t.Fatalf("expected stub fallback: %s", fallback.Prose)
	}
	if strings.Contains(strings.ToLower(fallback.Prose), "engine") || strings.Contains(fallback.Prose, "d20") {
		t.Fatalf("stub must stay literary: %s", fallback.Prose)
	}
}

func TestRunnerGarbageProseFallsBackToStub(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "en",
			Prose:  `Never invent dice or HP. {"actor":"Lute","room":"Hall","kind":"action"}`,
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{Locale: "en", ActorName: "Mira", Kind: "wait", RoomName: "Ashwood"})
	if strings.Contains(n.Prose, `"actor"`) || strings.Contains(n.Prose, "Never invent dice") {
		t.Fatalf("garbage runner prose must not reach players: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, "The lantern holds") {
		t.Fatalf("expected stub fallback: %s", n.Prose)
	}
}

func TestRunnerMixedLocaleFallsBack(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "tr",
			Prose:  "The watch is unblinded. Hold the line. The engine's die reads 0.",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{Locale: "tr", ActorName: "Bram", Kind: "story", RoomName: "Kalekarga"})
	if strings.Contains(n.Prose, "Hold the line") || strings.Contains(n.Prose, "die reads") {
		t.Fatalf("mixed locale runner prose must not reach players: %s", n.Prose)
	}
	if strings.Contains(n.Prose, "[") || strings.Contains(n.Prose, "eşiğe durur") {
		t.Fatalf("opening stub must not dump theme or narrator stage directions: %s", n.Prose)
	}
}

func TestStoryOpeningKeepsHostText(t *testing.T) {
	opening := "You wake on the cold stone floor of an abandoned Shaper temple. Pale blue light leaks through the cracks."
	svc := gateway.New()
	n := svc.Narrate(gateway.NarrateRequest{
		Locale:   "tr",
		Kind:     "story",
		RoomName: "World Of Warcraft",
		ThemeID:  "high-fantasy",
		Notes:    opening,
		Opening:  opening,
	})
	if !strings.Contains(n.Prose, "Shaper temple") {
		t.Fatalf("host opening missing: %s", n.Prose)
	}
	if strings.Contains(n.Prose, "Anlatıcı") || strings.Contains(n.Prose, "[high-fantasy]") || strings.Contains(n.Prose, "sessiz") {
		t.Fatalf("must not append turkish stub to host opening: %s", n.Prose)
	}
}

func TestShortOpeningSaladRejected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "tr",
			Prose:  "Bir sonraki çandan önce. Çığlık gelene dek bekle.",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	opening := "You wake on the cold stone floor of an abandoned Shaper temple. Pale blue light leaks through the cracks."
	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{Locale: "tr", Kind: "story", Notes: opening, Opening: opening, RoomName: "World Of Warcraft"})
	if strings.Contains(n.Prose, "çandan") || strings.Contains(n.Prose, "Çığlık") {
		t.Fatalf("short salad opening reached table: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, "Shaper temple") {
		t.Fatalf("expected host opening: %s", n.Prose)
	}
}

func TestRunnerSaladFallsBackToLiterary(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "tr",
			Prose:  "Nöbet dönmez. Luther, World Of Warcraft içinde, NE oluyor bu ses ne ??. Zar 10 der. Kilit durur. Bir pim kopar.",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	fail := false
	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "tr", ActorName: "Luther", Kind: "action", RoomName: "World Of Warcraft",
		Notes: "NE oluyor bu ses ne ??", Total: 10, Success: &fail,
	})
	if strings.Contains(n.Prose, "Nöbet dönmez") || strings.Contains(n.Prose, "Zar 10 der") || strings.Contains(n.Prose, "pim kopar") {
		t.Fatalf("training salad reached the table: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, "Luther") || !strings.Contains(n.Prose, "10") {
		t.Fatalf("literary fallback missing actor or count: %s", n.Prose)
	}
}

func TestRunnerHitVoiceOnMissFallsBack(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "en",
			Prose: "Luther picks up the medallion and studies it for a moment, feeling the subtle vibrations emanating from the carvings grow stronger. He lets out a low whistle, recognizing the power at work here. Without hesitation, he begins to trace a complex pattern along the wall with his finger, following the hum like a map. The others watch warily as the temperature drops slightly.",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	fail := false
	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "en", ActorName: "Luther", Kind: "action", RoomName: "Friday night",
		Notes: "I pick up the medallion and listen to the humming carvings, without going down the corridor.",
		Total: 5, Success: &fail,
	})
	if strings.Contains(strings.ToLower(n.Prose), "without hesitation") || strings.Contains(n.Prose, "recognizing the") {
		t.Fatalf("hit voice on a miss reached the table: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, "Luther") || !strings.Contains(n.Prose, "misses") || !strings.Contains(n.Prose, "5") {
		t.Fatalf("expected honest miss stub: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, "next beat") || strings.Contains(n.Prose, "nothing shifts") {
		t.Fatalf("miss must fail forward, not freeze: %s", n.Prose)
	}
}

func TestEnglishDeedKeepsEnglishBeat(t *testing.T) {
	fail := false
	svc := gateway.New()
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "tr", ActorName: "Luther", Kind: "action", RoomName: "World Of Warcraft",
		Notes: "Examine the humming carvings", Total: 10, Success: &fail,
	})
	if strings.Contains(n.Prose, "direnir") || strings.Contains(n.Prose, "World Of Warcraft") || strings.Contains(n.Prose, "Alet kayar") {
		t.Fatalf("mixed stub: %s", n.Prose)
	}
	if !strings.Contains(n.Prose, "Luther") || !strings.Contains(n.Prose, "carvings") || !strings.Contains(n.Prose, "10") {
		t.Fatalf("expected english literary beat: %s", n.Prose)
	}
}

func TestTableTitleNeverBecomesPlace(t *testing.T) {
	fail := false
	svc := gateway.New()
	for _, title := range []string{"Star Wars", "Dragon Age", "Harry Potter"} {
		n := svc.Narrate(gateway.NarrateRequest{
			Locale: "tr", ActorName: "Luther", Kind: "action", RoomName: title,
			Notes: "Examine the humming carvings", Total: 10, Success: &fail,
		})
		if strings.Contains(n.Prose, title) {
			t.Fatalf("lobby title leaked into beat (%s): %s", title, n.Prose)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "tr",
			Prose:  "Luther looks. Star Wars resists. Number 10. The tool slips. Time ends.",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	opening := "You wake on the cold stone floor of an abandoned Shaper temple."
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "tr", ActorName: "Luther", Kind: "action", RoomName: "Star Wars",
		Notes: "Examine the humming carvings", Opening: opening, Total: 10, Success: &fail,
	})
	if strings.Contains(n.Prose, "Star Wars") || strings.Contains(n.Prose, "The tool slips") {
		t.Fatalf("table title or staccato salad reached the table: %s", n.Prose)
	}
}

func TestTaleFollowsEnglishOpening(t *testing.T) {
	svc := gateway.New()
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "tr", ActorName: "Luther", Kind: "wait",
		Opening: "You wake on the cold stone floor of an abandoned Shaper temple.",
	})
	if n.Locale != "en" || strings.Contains(n.Prose, "nefesini") {
		t.Fatalf("english opening must keep english tale voice: locale=%s prose=%s", n.Locale, n.Prose)
	}
}

func TestNarrateForwardsWorldAndCast(t *testing.T) {
	var got gateway.NarrateRequest
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_ = json.NewEncoder(w).Encode(gateway.Narrative{
			Locale: "en",
			Prose:  "Pale blue light finds Luther at the carvings. The stone keeps its secret. Dust hangs in the nave. The ranger lowers his hand. Nothing yields yet.",
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	fail := false
	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{
		Locale: "tr", ActorName: "Luther", Kind: "action", RoomName: "Friday night",
		Opening:    "You wake on the cold stone floor of an abandoned Shaper temple.",
		Notes:      "Examine the humming carvings",
		WorldBrief: "Age: first winter\nMood: wary\nLook: high fantasy",
		Cast:       []gateway.CastMember{{Name: "Luther", Species: "human", Path: "ranger"}},
		Total:      10, Success: &fail,
	})
	if !strings.Contains(got.WorldBrief, "first winter") || len(got.Cast) != 1 || got.Cast[0].Path != "ranger" {
		t.Fatalf("world/cast not forwarded: %+v", got)
	}
	if strings.Contains(n.Prose, "Friday night") {
		t.Fatalf("lobby title in runner prose: %s", n.Prose)
	}
}

func TestStoryRunnerEnglishOpeningAccepted(t *testing.T) {
	opening := "You wake on the cold stone floor of an abandoned Shaper temple. Pale blue light leaks through the cracks."
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/narrate", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gateway.Narrative{Locale: "en", Prose: opening})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	svc := gateway.New()
	svc.ConfigureHub("your-org/talerole-storyteller", "")
	svc.SetRunners(srv.URL, "")
	n := svc.Narrate(gateway.NarrateRequest{Locale: "tr", Kind: "story", Notes: opening, Opening: opening, RoomName: "World Of Warcraft", ThemeID: "high-fantasy"})
	if n.Prose != opening {
		t.Fatalf("runner opening rewritten: %s", n.Prose)
	}
}
