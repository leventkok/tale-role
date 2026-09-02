package world_test

import (
	"strings"
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/world"
)

func TestCreateCompilesPromptPack(t *testing.T) {
	cat := world.NewCatalog()
	u, err := cat.Create("host", world.Draft{
		NameEN:  "Ashwood",
		ThemeID: "high-fantasy",
		Era:     "late autumn",
		Tone:    "wary",
		Taboos:  "no real-world politics",
		NPCs:    []world.NPC{{NameEN: "The Warden", Alignment: "neutral", Voice: "dry"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.DiceSystem != "d20" || u.PromptPackVersion != "v1" || u.RulesetID != "tale-core" {
		t.Fatalf("defaults: %+v", u)
	}
	if !strings.Contains(u.CompiledPrompt, "high-fantasy") || !strings.Contains(u.CompiledPrompt, "The Warden") {
		t.Fatalf("compile missed fields: %s", u.CompiledPrompt)
	}
	if strings.Contains(u.CompiledPrompt, "system_admin") {
		t.Fatal("compiled pack must not name system_admin")
	}
	list := cat.List("host")
	if len(list) != 1 || list[0].ID != u.ID {
		t.Fatalf("list: %+v", list)
	}
	if _, err := cat.Get(u.ID, "other"); err == nil {
		t.Fatal("other user must not read the universe")
	}
}
