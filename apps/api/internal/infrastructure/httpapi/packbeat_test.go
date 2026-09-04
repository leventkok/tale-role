package httpapi

import (
	"strings"
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
)

func TestPackBeatOmitsCount(t *testing.T) {
	miss := false
	got := packBeat("Floc", game.Turn{Kind: "action", Success: &miss, Total: 9}, "Ayağa kalkarım ve torbanın içindekilere göz atarım.")
	if strings.Contains(got, "Player count") || strings.Contains(got, "9") {
		t.Fatalf("count leaked into beat: %s", got)
	}
	if !strings.Contains(got, "Deed:") || !strings.Contains(got, "MISS") {
		t.Fatalf("expected deed wrapper: %s", got)
	}
	if !strings.Contains(got, "third person") {
		t.Fatalf("expected third-person instruction: %s", got)
	}
}
