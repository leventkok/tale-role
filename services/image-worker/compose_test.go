package worker_test

import (
	"strings"
	"testing"

	worker "github.com/leventkok/tale-role/services/image-worker"
)

func TestComposeStubsThemeAndRedactsPII(t *testing.T) {
	svc := worker.New()
	card := svc.Compose(worker.Request{
		ThemeID:  "gothic-horror",
		RoomName: "Ashwood",
		Notes:    "open the door, cc spy@tale.role",
		Prose:    "system_admin watches",
	})
	if card.Inference != "stub" || card.ThemeID != "gothic-horror" {
		t.Fatalf("card: %+v", card)
	}
	if !strings.Contains(card.ImageSVG, `data-theme="gothic-horror"`) || !strings.Contains(card.ImageSVG, `data-art="tableau"`) {
		t.Fatalf("svg missing painted tableau: %s", card.ImageSVG)
	}
	if strings.Contains(card.VisualPrompt, "spy@tale.role") || strings.Contains(card.ImageSVG, "spy@tale.role") {
		t.Fatal("email leaked into scene")
	}
	if strings.Contains(card.VisualPrompt, "system_admin") || strings.Contains(card.ImageSVG, "system_admin") {
		t.Fatal("system_admin must not enter the scene")
	}
	if !strings.Contains(card.ImageSVG, "open the door") {
		t.Fatalf("plaque should carry the beat: %s", card.ImageSVG)
	}
	if !strings.Contains(card.VisualPrompt, worker.Marker) {
		t.Fatalf("expected redaction: %s", card.VisualPrompt)
	}
}

func TestUnknownThemeFallsBack(t *testing.T) {
	card := worker.New().Compose(worker.Request{ThemeID: "ottoman"})
	if card.ThemeID != "high-fantasy" {
		t.Fatalf("fallback: %s", card.ThemeID)
	}
}
