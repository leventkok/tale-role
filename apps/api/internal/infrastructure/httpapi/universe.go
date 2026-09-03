package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leventkok/tale-role/apps/api/internal/application/world"
	"github.com/leventkok/tale-role/apps/api/internal/shared/httperr"
)

func (s *Server) listUniverses(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	httperr.JSON(w, http.StatusOK, map[string]any{"universes": s.worlds.List(u.ID)})
}

func (s *Server) getUniverse(w http.ResponseWriter, r *http.Request) {
	u := userFrom(r)
	doc, err := s.worlds.Get(chi.URLParam(r, "universeID"), u.ID)
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	httperr.JSON(w, http.StatusOK, doc)
}

func (s *Server) createUniverse(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NameEN        string      `json:"name_en"`
		NameTR        string      `json:"name_tr"`
		ThemeID       string      `json:"theme_id"`
		DiceSystem    string      `json:"dice_system"`
		ContentRating string      `json:"content_rating"`
		Era           string      `json:"era"`
		Tone          string      `json:"tone"`
		Description   string      `json:"description"`
		Opening       string      `json:"opening"`
		Taboos        string      `json:"taboos"`
		NPCs          []world.NPC `json:"npcs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httperr.Write(w, s.log, http.StatusBadRequest, "invalid request", err)
		return
	}
	u := userFrom(r)
	doc, err := s.worlds.Create(u.ID, world.Draft{
		NameEN:        body.NameEN,
		NameTR:        body.NameTR,
		ThemeID:       body.ThemeID,
		DiceSystem:    body.DiceSystem,
		ContentRating: body.ContentRating,
		Era:           body.Era,
		Tone:          body.Tone,
		Description:   body.Description,
		Opening:       body.Opening,
		Taboos:        body.Taboos,
		NPCs:          body.NPCs,
	})
	if err != nil {
		s.writeAppError(w, err)
		return
	}
	s.svc.GrantLantern(u.ID, 25)
	httperr.JSON(w, http.StatusCreated, doc)
}
