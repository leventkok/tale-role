package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
	gateway "github.com/leventkok/tale-role/services/llm-gateway"
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
		Name  string     `json:"name"`
		Stats game.Stats `json:"stats"`
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
		Kind   string `json:"kind"`
		Skill  string `json:"skill"`
		Notes  string `json:"notes"`
		DC     int    `json:"dc"`
		Locale string `json:"locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	u := userFrom(r)
	roomID := chi.URLParam(r, "roomID")
	_ = s.llm.ProposeIntent(gateway.IntentRequest{
		Locale: body.Locale,
		RoomID: roomID,
		Kind:   body.Kind,
		Skill:  body.Skill,
		Notes:  body.Notes,
	})
	turn, err := s.table.Act(roomID, u.ID, body.Kind, body.Skill, body.Notes, body.DC)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	pub, err := s.table.View(roomID, u.ID)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	actorName := ""
	names := make([]string, 0, len(pub.Characters))
	for _, ch := range pub.Characters {
		names = append(names, ch.Name)
		if ch.UserID == u.ID {
			actorName = ch.Name
		}
	}
	n := s.llm.Narrate(gateway.NarrateRequest{
		Locale:        body.Locale,
		RoomID:        roomID,
		RoomName:      pub.Name,
		ActorName:     actorName,
		Kind:          turn.Kind,
		Notes:         body.Notes,
		DiceSystem:    turn.DiceSystem,
		Rolls:         turn.Rolls,
		Total:         turn.Total,
		Success:       turn.Success,
		PresenceNames: names,
	})
	narr := game.Narrative{Locale: n.Locale, Prose: n.Prose, NPCLines: []game.NPCLine{}}
	for _, line := range n.NPCLines {
		narr.NPCLines = append(narr.NPCLines, game.NPCLine{NPCID: line.NPCID, Text: line.Text})
	}
	_ = s.table.AttachNarrative(roomID, narr)
	turn.Narrative = &narr
	httperr.JSON(w, http.StatusOK, turn)
}

func (s *Server) adminRuntime(w http.ResponseWriter, _ *http.Request) {
	httperr.JSON(w, http.StatusOK, s.llm.Runtime())
}

func (s *Server) adminSwap(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PromptPack string `json:"prompt_pack"`
		AdapterID  string `json:"adapter_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	if err := s.llm.Swap(body.PromptPack, body.AdapterID); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	httperr.JSON(w, http.StatusOK, s.llm.Runtime())
}

func (s *Server) adminTraces(w http.ResponseWriter, _ *http.Request) {
	httperr.JSON(w, http.StatusOK, map[string]any{"traces": s.llm.Traces()})
}
