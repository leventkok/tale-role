package world_test

import (
	"strings"
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
)

func TestCreateCompilesPromptPack(t *testing.T) {
	cat := world.NewCatalog()
	u, err := cat.Create("host", world.Draft{
		NameEN:      "Ashwood",
		ThemeID:     "high-fantasy",
		Era:         "late autumn",
		Tone:        "wary",
		Description: "A hush in the pines. The Warden keeps the last lantern.",
		Opening:     "You wake on the Ashwood road.",
		NPCs:        []world.NPC{{NameEN: "The Warden", Alignment: "neutral", Voice: "dry"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.DiceSystem != "d20" || u.PromptPackVersion != "v1" || u.RulesetID != "tale-core" {
		t.Fatalf("defaults: %+v", u)
	}
	if !strings.Contains(u.CompiledPrompt, "high-fantasy") || !strings.Contains(u.CompiledPrompt, "The Warden") || !strings.Contains(u.CompiledPrompt, "Ashwood road") {
		t.Fatalf("compile missed fields: %s", u.CompiledPrompt)
	}
	if strings.Contains(u.CompiledPrompt, "system_admin") {
		t.Fatal("compiled pack must not name system_admin")
	}
	brief, ok := cat.TableBrief(u.ID)
	if !ok || !strings.Contains(brief, "late autumn") || !strings.Contains(brief, "wary") || !strings.Contains(brief, "The Warden") || !strings.Contains(brief, "Ashwood road") {
		t.Fatalf("table brief missed world fields: %s", brief)
	}
	if strings.Contains(brief, "high-fantasy") || strings.Contains(brief, "system_admin") {
		t.Fatalf("table brief must not leak theme ids: %s", brief)
	}
	list := cat.List("host")
	if len(list) != 1 || list[0].ID != u.ID {
		t.Fatalf("list: %+v", list)
	}
	if _, err := cat.Get(u.ID, "other"); err == nil {
		t.Fatal("other user must not read the universe")
	}
	opening, desc, nameEN, _, ok := cat.SceneSeed(u.ID)
	if !ok || opening != "You wake on the Ashwood road." || desc == "" || nameEN != "Ashwood" {
		t.Fatalf("scene seed: %q %q %q", opening, desc, nameEN)
	}
	stats := game.Stats{STR: 3, DEX: 3, CON: 3, INT: 3, WIS: 3, CHA: 3}
	if err := cat.UpsertHero(u.ID, world.Hero{UserID: "p1", Name: "Iri", Path: "warden", Skills: []string{"athletics"}, Stats: stats, HP: 12, XP: 20, Level: 2}); err != nil {
		t.Fatal(err)
	}
	h, ok := cat.Hero(u.ID, "p1")
	if !ok || h.Name != "Iri" || h.Level != 2 {
		t.Fatalf("hero: %+v %v", h, ok)
	}
	cat.ForgetPlayer("p1")
	if _, ok := cat.Hero(u.ID, "p1"); ok {
		t.Fatal("erase must drop the player's hero")
	}
	cat.ForgetOwner("host")
	if len(cat.List("host")) != 0 {
		t.Fatal("erasure must drop owned universes")
	}
}
