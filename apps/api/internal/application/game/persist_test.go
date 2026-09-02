package game_test

import (
	"testing"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
)

type memSink struct {
	n       int
	deleted int
}

func (m *memSink) UpsertRoom(*game.Room) error {
	m.n++
	return nil
}

func (m *memSink) DeleteRoom(string) error {
	m.deleted++
	return nil
}

func TestTablePersistsOnCreate(t *testing.T) {
	tab := game.NewTable()
	sink := &memSink{}
	tab.SetSink(sink)
	if _, err := tab.Create("host", "Ashwood", "link", "", "d20"); err != nil {
		t.Fatal(err)
	}
	if sink.n != 1 {
		t.Fatalf("upserts: %d", sink.n)
	}
}
