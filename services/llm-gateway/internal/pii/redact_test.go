package pii_test

import (
	"strings"
	"testing"

	"github.com/leventkok/tale-role/services/llm-gateway/internal/pii"
)

func TestRedactEmailAndDigits(t *testing.T) {
	in := "mail player@tale.role then card 4111111111111111"
	out := pii.Redact(in)
	if strings.Contains(out, "player@tale.role") || strings.Contains(out, "4111111111111111") {
		t.Fatalf("pii leaked: %s", out)
	}
	if strings.Count(out, pii.Marker) < 2 {
		t.Fatalf("expected two redactions, got %s", out)
	}
}
