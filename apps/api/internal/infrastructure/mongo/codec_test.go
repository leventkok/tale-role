package mongostore

import (
	"testing"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"go.mongodb.org/mongo-driver/bson"
)

func TestRoomCodecRoundTrip(t *testing.T) {
	ok := true
	src := &game.Room{
		ID: "r1", Name: "Ashwood", HostID: "h1", DiceSystem: "d20", JoinMode: "link",
		Password: "secret", Members: map[string]game.Member{"h1": {UserID: "h1", Role: "gm"}},
		Characters: map[string]*game.Character{"h1": {UserID: "h1", Name: "Iri", HP: 11}},
		Turns:      []game.Turn{{ActorID: "h1", Kind: "action", Success: &ok}},
		Chronicle:  []string{"Karanlık bir koridor açık."},
		CreatedAt:  time.Unix(0, 0).UTC(),
	}
	raw, err := bson.Marshal(encodeRoom(src))
	if err != nil {
		t.Fatal(err)
	}
	var d roomDoc
	if err := bson.Unmarshal(raw, &d); err != nil {
		t.Fatal(err)
	}
	got := decodeRoom(d)
	if got.Password != "secret" || got.Members["h1"].Role != "gm" || got.Characters["h1"].Name != "Iri" {
		t.Fatalf("roundtrip: %+v", got)
	}
	if len(got.Chronicle) != 1 || got.Chronicle[0] != "Karanlık bir koridor açık." {
		t.Fatalf("chronicle: %+v", got.Chronicle)
	}
}
