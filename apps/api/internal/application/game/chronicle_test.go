package game_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
)

func TestSceneMemoryKeepsTempleObjects(t *testing.T) {
	opening := "Terk edilmiş bir Şekillendiren tapınağının soğuk taş zemininde uyanırsın. " +
		"Tavan çatlaklarından sızan soluk mavi ışık, duvarları kaplayan eski oymaları yakalar. " +
		"Karanlık koridorun bir yerinde metal sürtünür. Yanında bir torba durur; madalyonda Uyan yazar."
	got := game.SceneMemory(opening)
	blob := strings.Join(got, " ")
	for _, want := range []string{"oym", "koridor", "torba", "madalyon", "çatlak"} {
		if !strings.Contains(blob, want) {
			t.Fatalf("scene memory missing %q: %v", want, got)
		}
	}
}

func TestChroniclePinsSceneAfterManyBeats(t *testing.T) {
	tab := game.NewTable()
	room, err := tab.Create("host", "Ashwood", "public", "", "d20")
	if err != nil {
		t.Fatal(err)
	}
	opening := "Duvarlardaki oymalar uğuldamaktadır. Karanlık koridorun bir yerinde metal sürtünür. " +
		"Yanında bir torba durur; madalyonda Uyan yazar. Tavandan mavi ışık sızar."
	tab.Remember(room.ID, game.SceneMemory(opening)...)
	ok := true
	for i := 0; i < 12; i++ {
		tab.Remember(room.ID, game.BeatMemory("Floc", "hamle "+strconv.Itoa(i), &ok))
	}
	facts := tab.Chronicle(room.ID)
	if len(facts) > 8 {
		t.Fatalf("chronicle grew past 8: %d %v", len(facts), facts)
	}
	blob := strings.Join(facts, " ")
	if !strings.Contains(blob, "oym") || !strings.Contains(blob, "koridor") || !strings.Contains(blob, "torba") {
		t.Fatalf("scene dropped after beats: %v", facts)
	}
	if !strings.Contains(blob, "hamle 11") {
		t.Fatalf("latest beat missing: %v", facts)
	}
}

func TestBeatMemoryStripsPackedDeed(t *testing.T) {
	fail := false
	line := game.BeatMemory("Floc", "deed: Torbayı sırtıma alır. Narrate this outcome.", &fail)
	if !strings.Contains(line, "Torbayı sırtıma alır") || strings.Contains(line, "Narrate") || !strings.Contains(line, "kaçırdı") {
		t.Fatalf("packed deed leaked: %s", line)
	}
}
