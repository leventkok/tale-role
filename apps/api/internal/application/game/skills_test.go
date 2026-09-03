package game_test

import (
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
)

func TestCheckBonusProficiencyAndXP(t *testing.T) {
	ch := game.Character{
		Stats:  game.Stats{STR: 4, DEX: 3, CON: 3, INT: 3, WIS: 3, CHA: 2},
		Skills: []string{"athletics"},
		Level:  1,
		HP:     game.MaxHP(game.Stats{STR: 4, DEX: 3, CON: 3, INT: 3, WIS: 3, CHA: 2}, 1),
	}
	if game.MaxHP(ch.Stats, 1) != 12 {
		t.Fatalf("max hp: %d", game.MaxHP(ch.Stats, 1))
	}
	got, err := ch.CheckBonus("athletics")
	if err != nil || got != 6 {
		t.Fatalf("marked skill: %d %v", got, err)
	}
	plain, err := ch.CheckBonus("str")
	if err != nil || plain != 4 {
		t.Fatalf("ability roll has no proficiency: %d %v", plain, err)
	}
	unmarked, err := ch.CheckBonus("stealth")
	if err != nil || unmarked != 3 {
		t.Fatalf("unmarked skill: %d %v", unmarked, err)
	}
	if game.ValidSkills([]string{"athletics", "stealth", "persuasion", "arcana"}) {
		t.Fatal("four skills must fail")
	}
	if !game.ValidSkills(nil) {
		t.Fatal("empty skills stay valid for old sheets")
	}
	ch.GrantXP(100)
	if ch.Level != 2 || ch.XP != 0 || ch.HP != 13 {
		t.Fatalf("level up: %+v", ch)
	}
}

func TestSetSheetAndSeatHero(t *testing.T) {
	tab := game.NewTable()
	room, err := tab.Create("host", "Ashwood", "public", "", "d20")
	if err != nil {
		t.Fatal(err)
	}
	sheet := game.Sheet{
		Name: "Iri", Species: "human", Path: "warden", Backstory: "keeps the last lantern",
		Stats:  game.Stats{STR: 3, DEX: 3, CON: 3, INT: 3, WIS: 3, CHA: 3},
		Skills: []string{"athletics", "perception", "insight"},
	}
	if err := tab.SetSheet(room.ID, "host", sheet); err != nil {
		t.Fatal(err)
	}
	if err := tab.SetSheet(room.ID, "host", sheet); err == nil {
		t.Fatal("second sheet in the same room must fail")
	}
	pub, err := tab.Public(room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pub.Characters) != 1 || pub.Characters[0].Path != "warden" || pub.Characters[0].HP != 12 {
		t.Fatalf("sheet: %+v", pub.Characters)
	}
	next, err := tab.Create("host", "Ashwood II", "public", "", "d20")
	if err != nil {
		t.Fatal(err)
	}
	saved := pub.Characters[0]
	saved.Level = 4
	saved.XP = 20
	saved.HP = game.MaxHP(saved.Stats, 4)
	if err := tab.SeatHero(next.ID, "host", saved); err != nil {
		t.Fatal(err)
	}
	seated, err := tab.Public(next.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(seated.Characters) != 1 || seated.Characters[0].Level != 4 || seated.Characters[0].XP != 20 || seated.Characters[0].Name != "Iri" {
		t.Fatalf("returning hero: %+v", seated.Characters)
	}
}
