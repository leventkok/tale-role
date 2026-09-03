package iam_test

import (
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
)

func TestGrantLanternNeverTouchesDice(t *testing.T) {
	u := &iam.User{}
	u.GrantLantern(25)
	u.GrantLantern(10)
	u.GrantLantern(8)
	if u.LanternLevel != 1 || u.LanternXP != 43 {
		t.Fatalf("lantern: %+v", u)
	}
	u.GrantLantern(57)
	if u.LanternLevel != 2 || u.LanternXP != 0 {
		t.Fatalf("lantern level: %+v", u)
	}
	if iam.NormalizePortrait("") != "warden" || !iam.KnownPortrait("ranger") || iam.KnownPortrait("ottoman") {
		t.Fatal("portrait catalog")
	}
}
