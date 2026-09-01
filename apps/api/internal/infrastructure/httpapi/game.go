package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
)

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		JoinMode   string `json:"join_mode"`
		Password   string `json:"password"`
		DiceSystem string `json:"dice_system"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	u := userFrom(r)
	room, err := s.table.Create(u.ID, body.Name, body.JoinMode, body.Password, body.DiceSystem)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusCreated, map[string]any{"id": room.ID, "dice_system": room.DiceSystem})
}

func (s *Server) joinRoom(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	u := userFrom(r)
	role := "player"
	if u.Email != "" && s.adminEmail != "" && u.Email == s.adminEmail {
		role = "system_admin"
	}
	if err := s.table.Join(chi.URLParam(r, "roomID"), u.ID, body.Password, role); err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) getRoom(w http.ResponseWriter, r *http.Request) {
	pub, err := s.table.View(chi.URLParam(r, "roomID"), userFrom(r).ID)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, pub)
}

func (s *Server) setCharacter(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string      `json:"name"`
		Stats game.Stats  `json:"stats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	u := userFrom(r)
	if err := s.table.SetCharacter(chi.URLParam(r, "roomID"), u.ID, body.Name, body.Stats); err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (s *Server) startRoom(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	if err := s.table.Start(chi.URLParam(r, "roomID"), u.ID); err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) actRoom(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind  string `json:"kind"`
		Skill string `json:"skill"`
		Notes string `json:"notes"`
		DC    int    `json:"dc"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	u := userFrom(r)
	turn, err := s.table.Act(chi.URLParam(r, "roomID"), u.ID, body.Kind, body.Skill, body.Notes, body.DC)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, turn)
}
