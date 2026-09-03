package httpapi

import (
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
)

func (s *Server) seatSavedHero(roomID, userID string) {
	pub, err := s.table.Public(roomID)
	if err != nil || pub.UniverseID == "" {
		return
	}
	h, ok := s.worlds.Hero(pub.UniverseID, userID)
	if !ok {
		return
	}
	_ = s.table.SeatHero(roomID, userID, game.Character{
		UserID: userID, Name: h.Name, Species: h.Species, Path: h.Path, Backstory: h.Backstory,
		Skills: h.Skills, Stats: h.Stats, HP: h.HP, XP: h.XP, Level: h.Level,
	})
}

func (s *Server) rememberHero(roomID, userID string) {
	pub, err := s.table.View(roomID, userID)
	if err != nil || pub.UniverseID == "" {
		return
	}
	for _, ch := range pub.Characters {
		if ch.UserID != userID {
			continue
		}
		_ = s.worlds.UpsertHero(pub.UniverseID, world.Hero{
			UserID: ch.UserID, Name: ch.Name, Species: ch.Species, Path: ch.Path, Backstory: ch.Backstory,
			Skills: ch.Skills, Stats: ch.Stats, HP: ch.HP, XP: ch.XP, Level: ch.Level,
		})
		return
	}
}
