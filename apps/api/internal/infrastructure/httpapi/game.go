package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
	worker "github.com/leventkok/tale-role/services/image-worker"
	gateway "github.com/leventkok/tale-role/services/llm-gateway"
)

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		JoinMode   string `json:"join_mode"`
		Password   string `json:"password"`
		DiceSystem string `json:"dice_system"`
		UniverseID string `json:"universe_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	if s.denyUnlicensedPlayHTTP(w, r) {
		return
	}
	u := userFrom(r)
	dice := body.DiceSystem
	name := body.Name
	if body.UniverseID != "" {
		uni, err := s.worlds.GetForHost(body.UniverseID, u.ID)
		if err != nil {
			s.writeAppError(w, err)
			return
		}
		dice = uni.DiceSystem
		if strings.TrimSpace(name) == "" {
			name = uni.NameEN
		}
		room, err := s.table.Create(u.ID, name, body.JoinMode, body.Password, dice)
		if err != nil {
			s.writeAppError(w, err)
			return
		}
		_ = s.table.BindUniverse(room.ID, uni.ID, uni.ThemeID, uni.PromptPackVersion)
		s.seatSavedHero(room.ID, u.ID)
		httperr.JSON(w, http.StatusCreated, map[string]any{"id": room.ID, "dice_system": room.DiceSystem, "universe_id": uni.ID})
		return
	}
	room, err := s.table.Create(u.ID, name, body.JoinMode, body.Password, dice)
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
	if s.denyUnlicensedPlayHTTP(w, r) {
		return
	}
	u := userFrom(r)
	role := "player"
	if u.Email != "" && s.adminEmail != "" && u.Email == s.adminEmail {
		role = "system_admin"
	}
	if err := s.table.Join(chi.URLParam(r, "roomID"), u.ID, body.Password, role); err != nil {
		s.writeAppError(w, err)
		return
	}
	s.seatSavedHero(chi.URLParam(r, "roomID"), u.ID)
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
		Name      string     `json:"name"`
		Species   string     `json:"species"`
		Path      string     `json:"path"`
		Backstory string     `json:"backstory"`
		Skills    []string   `json:"skills"`
		Stats     game.Stats `json:"stats"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	if s.denyUnlicensedPlayHTTP(w, r) {
		return
	}
	u := userFrom(r)
	if err := s.table.SetSheet(chi.URLParam(r, "roomID"), u.ID, game.Sheet{
		Name: body.Name, Species: body.Species, Path: body.Path, Backstory: body.Backstory,
		Stats: body.Stats, Skills: body.Skills,
	}); err != nil {
		s.writeAppError(w, err)
		return
	}
	s.rememberHero(chi.URLParam(r, "roomID"), u.ID)
	httperr.JSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (s *Server) rollInitiative(w http.ResponseWriter, r *http.Request) {
	if s.denyUnlicensedPlayHTTP(w, r) {
		return
	}
	u := userFrom(r)
	n, err := s.table.RollInitiative(chi.URLParam(r, "roomID"), u.ID)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"initiative": n})
}

func (s *Server) startRoom(w http.ResponseWriter, r *http.Request) {
	if s.denyUnlicensedPlayHTTP(w, r) {
		return
	}
	u := userFrom(r)
	roomID := chi.URLParam(r, "roomID")
	if err := s.table.Start(roomID, u.ID); err != nil {
		s.writeAppError(w, err)
		return
	}
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}
	s.openTale(roomID, u.ID, locale)
	httperr.JSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) completeRoom(w http.ResponseWriter, r *http.Request) {
	if s.denyUnlicensedPlayHTTP(w, r) {
		return
	}
	u := userFrom(r)
	ids, err := s.table.Complete(chi.URLParam(r, "roomID"), u.ID)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	for _, id := range ids {
		s.svc.GrantLantern(id, game.TaleCompleteXP)
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
	if s.denyUnlicensedPlayHTTP(w, r) {
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
	turn = s.narrateTurn(roomID, u.ID, body.Locale, body.Notes, turn)
	s.rememberHero(roomID, u.ID)
	httperr.JSON(w, http.StatusOK, turn)
}

func (s *Server) openTale(roomID, userID, locale string) {
	pub, err := s.table.View(roomID, userID)
	if err != nil || len(pub.Turns) == 0 {
		return
	}
	last := pub.Turns[len(pub.Turns)-1]
	if last.Kind != "story" {
		return
	}
	_ = s.narrateTurn(roomID, userID, locale, last.Notes, last)
}

func (s *Server) narrateTurn(roomID, userID, locale, notes string, turn game.Turn) game.Turn {
	pub, err := s.table.View(roomID, userID)
	if err != nil {
		return turn
	}
	actorName := ""
	names := make([]string, 0, len(pub.Characters))
	for _, ch := range pub.Characters {
		names = append(names, ch.Name)
		if ch.UserID == userID {
			actorName = ch.Name
		}
	}
	opening := ""
	worldBrief := ""
	if pub.UniverseID != "" {
		seedOpening, seedDesc, _, _, ok := s.worlds.SceneSeed(pub.UniverseID)
		if ok {
			opening = strings.TrimSpace(seedOpening)
			if opening == "" {
				opening = strings.TrimSpace(seedDesc)
			}
		}
		if brief, ok := s.worlds.TableBrief(pub.UniverseID); ok {
			worldBrief = brief
		}
	}
	cast := tableCast(pub.Characters)
	if turn.Kind == "story" {
		if locale == "tr" {
			actorName = "Anlatıcı"
		} else {
			actorName = "Storyteller"
		}
		if strings.TrimSpace(notes) == "" {
			notes = opening
		}
	}
	for _, prev := range pub.Turns {
		if prev.Kind != "story" || prev.Narrative == nil {
			continue
		}
		if live := strings.TrimSpace(prev.Narrative.Prose); live != "" {
			opening = live
			break
		}
	}
	prior := make([]string, 0, 3)
	for _, prev := range pub.Turns {
		if prev.Narrative == nil || prev.Kind == "story" {
			continue
		}
		p := strings.TrimSpace(prev.Narrative.Prose)
		if p != "" {
			prior = append(prior, p)
		}
	}
	if len(prior) > 3 {
		prior = prior[len(prior)-3:]
	}
	facts := s.table.Chronicle(roomID)
	if len(facts) == 0 && opening != "" {
		facts = game.SceneMemory(opening)
	}
	beat := packBeat(actorName, turn, notes)
	n := s.llm.Narrate(gateway.NarrateRequest{
		Locale:        locale,
		RoomID:        roomID,
		RoomName:      pub.Name,
		ActorName:     actorName,
		Kind:          turn.Kind,
		Notes:         beat,
		DiceSystem:    turn.DiceSystem,
		Rolls:         turn.Rolls,
		Total:         turn.Total,
		Success:       turn.Success,
		PresenceNames: names,
		ThemeID:       pub.ThemeID,
		Opening:       opening,
		Prior:         prior,
		WorldBrief:    worldBrief,
		Cast:          cast,
		Facts:         facts,
	})
	narr := game.Narrative{Locale: n.Locale, Prose: n.Prose, NPCLines: []game.NPCLine{}}
	for _, line := range n.NPCLines {
		narr.NPCLines = append(narr.NPCLines, game.NPCLine{NPCID: line.NPCID, Text: line.Text})
	}
	_ = s.table.AttachNarrative(roomID, narr)
	if turn.Kind == "story" {
		s.table.Remember(roomID, game.SceneMemory(n.Prose)...)
	} else if turn.Kind == "action" {
		s.table.Remember(roomID, game.BeatMemory(actorName, notes, turn.Success))
	}
	turn.Narrative = &narr
	go s.paintScene(roomID, pub.ThemeID, pub.Name, notes, n.Prose)
	return turn
}

func tableCast(chars []game.Character) []gateway.CastMember {
	out := make([]gateway.CastMember, 0, len(chars))
	for _, ch := range chars {
		name := strings.TrimSpace(ch.Name)
		if name == "" || name == "system_admin" {
			continue
		}
		if len(out) >= 8 {
			break
		}
		back := strings.TrimSpace(ch.Backstory)
		if len(back) > 300 {
			back = strings.TrimSpace(back[:300])
		}
		out = append(out, gateway.CastMember{
			Name:      name,
			Species:   strings.TrimSpace(ch.Species),
			Path:      strings.TrimSpace(ch.Path),
			Backstory: back,
		})
	}
	return out
}

func packBeat(actor string, turn game.Turn, notes string) string {
	deed := strings.TrimSpace(notes)
	switch turn.Kind {
	case "say":
		if deed == "" {
			return actor + " asks a question. No roll. Answer only. Do not change the scene."
		}
		return actor + " asks (no roll, do not change the scene, only answer): " + deed
	case "pass", "wait":
		return actor + " passes. No roll. One short beat. Do not change the scene."
	case "story":
		return deed
	}
	outcome := "MISS"
	if turn.Success != nil && *turn.Success {
		outcome = "HIT"
	}
	line := actor + " attempts a deed. " + outcome + "."
	if deed != "" {
		line += " Deed: " + deed + "."
	}
	line += " Narrate this outcome in third person. Do not paste the deed. Never mention dice, counts, or hidden difficulty."
	return line
}

func (s *Server) paintScene(roomID, themeID, roomName, notes, prose string) {
	card := s.images.Compose(worker.Request{
		ThemeID:  themeID,
		RoomName: roomName,
		Notes:    notes,
		Prose:    prose,
	})
	_ = s.table.AttachScene(roomID, game.Scene{
		ThemeID:      card.ThemeID,
		VisualPrompt: card.VisualPrompt,
		ImageSVG:     card.ImageSVG,
		Inference:    card.Inference,
	})
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

func (s *Server) adminPacks(w http.ResponseWriter, _ *http.Request) {
	httperr.JSON(w, http.StatusOK, map[string]any{"packs": s.llm.Packs()})
}

func (s *Server) adminPutPack(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID string `json:"id"`
		EN string `json:"en"`
		TR string `json:"tr"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	if err := s.llm.PutPack(body.ID, body.EN, body.TR); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"packs": s.llm.Packs()})
}

func (s *Server) adminLobbies(w http.ResponseWriter, _ *http.Request) {
	rows := s.table.Lobbies()
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]any{
			"id": row.ID, "name": row.Name, "universe_id": row.UniverseID, "join_mode": row.JoinMode,
			"started": row.Started, "completed": row.Completed, "seats": row.Seats,
			"started_at": lobbyStartedAt(row.StartedAt),
		})
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"lobbies": out})
}

func lobbyStartedAt(at *time.Time) any {
	if at == nil {
		return nil
	}
	return at.UTC().Format(time.RFC3339)
}

func (s *Server) adminCloseRoom(w http.ResponseWriter, r *http.Request) {
	if err := s.table.AdminClose(chi.URLParam(r, "roomID")); err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, map[string]any{"closed": true})
}

func (s *Server) adminLive(w http.ResponseWriter, r *http.Request) {
	httperr.JSON(w, http.StatusOK, s.llm.Live(strings.TrimSpace(r.URL.Query().Get("room_id"))))
}
