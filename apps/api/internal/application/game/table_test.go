package game_test

import (
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
)

func TestStatsBudget(t *testing.T) {
	ok := game.Stats{STR: 3, DEX: 3, CON: 3, INT: 3, WIS: 3, CHA: 3}
	if !ok.Valid() {
		t.Fatal("18 points even spread should be valid")
	}
	bad := game.Stats{STR: 6, DEX: 6, CON: 6, INT: 6, WIS: 6, CHA: 6}
	if bad.Valid() {
		t.Fatal("over budget")
	}
}

func TestHiddenAdminAndDice(t *testing.T) {
	tab := game.NewTable()
	seq := []int{10, 15, 4, 12}
	i := 0
	tab.UseDie(func(sides int) int {
		n := seq[i%len(seq)]
		i++
		if n > sides {
			return sides
		}
		return n
	})

	room, err := tab.Create("host", "Ashwood", "password", "secret", "d20")
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.Join(room.ID, "p1", "secret", "player"); err != nil {
		t.Fatal(err)
	}
	if err := tab.Join(room.ID, "admin", "wrong", "system_admin"); err != nil {
		t.Fatal(err)
	}
	stats := game.Stats{STR: 3, DEX: 4, CON: 3, INT: 3, WIS: 3, CHA: 2}
	if err := tab.SetCharacter(room.ID, "p1", "Iri", stats); err != nil {
		t.Fatal(err)
	}
	if err := tab.SetCharacter(room.ID, "host", "GM-PC", game.Stats{STR: 3, DEX: 3, CON: 3, INT: 3, WIS: 3, CHA: 3}); err != nil {
		t.Fatal(err)
	}
	if err := tab.SetCharacter(room.ID, "admin", "Ghost", stats); err == nil {
		t.Fatal("admin must not create a table character")
	}
	if _, err := tab.RollInitiative(room.ID, "p1"); err != nil {
		t.Fatal(err)
	}
	if _, err := tab.RollInitiative(room.ID, "host"); err != nil {
		t.Fatal(err)
	}
	if err := tab.Start(room.ID, "host"); err != nil {
		t.Fatal(err)
	}
	if _, err := tab.Act(room.ID, "p1", "action", "str", "force the door", 12); err == nil {
		t.Fatal("p1 must wait until their seat")
	}
	first := ""
	pubOrder, err := tab.Public(room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pubOrder.TurnOrder) == 0 {
		t.Fatal("expected seating order")
	}
	first = pubOrder.TurnOrder[0]
	if _, err := tab.Act(room.ID, first, "say", "", "look around", 0); err != nil {
		t.Fatal(err)
	}
	second := pubOrder.TurnOrder[1]
	turn, err := tab.Act(room.ID, second, "action", "str", "force the door", 12)
	if err != nil {
		t.Fatal(err)
	}
	if turn.Success == nil {
		t.Fatal("action needs a success flag")
	}
	pass, err := tab.Act(room.ID, first, "pass", "", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if pass.Success != nil || len(pass.Rolls) != 0 {
		t.Fatal("pass skips dice")
	}

	pub, err := tab.Public(room.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range pub.Presence {
		if m.Role == "system_admin" || m.UserID == "admin" {
			t.Fatal("system_admin leaked into public presence")
		}
	}
	if len(pub.TurnOrder) == 0 {
		t.Fatal("expected initiative order")
	}
	earned, err := tab.Complete(room.ID, "host")
	if err != nil {
		t.Fatal(err)
	}
	if len(earned) != 2 {
		t.Fatalf("expected two players who acted, got %v", earned)
	}
	if _, err := tab.Act(room.ID, first, "pass", "", "", 0); err == nil {
		t.Fatal("ended tale must refuse turns")
	}
	again, err := tab.Complete(room.ID, "host")
	if err != game.ErrAlreadyEnded || again != nil {
		t.Fatalf("complete twice: %v %v", err, again)
	}
}

func TestAdminCloseEndsLobbyWithoutHost(t *testing.T) {
	tab := game.NewTable()
	room, err := tab.Create("host", "Ashwood", "public", "", "d20")
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.AdminClose(room.ID); err != nil {
		t.Fatal(err)
	}
	if err := tab.AdminClose(room.ID); err != game.ErrAlreadyEnded {
		t.Fatalf("admin close twice: %v", err)
	}
	pub, err := tab.View(room.ID, "host")
	if err != nil || !pub.Completed {
		t.Fatalf("lobby not closed: %v completed=%v", err, pub.Completed)
	}
}
