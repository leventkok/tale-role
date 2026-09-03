package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/leventkok/tale-role/services/llm-gateway/internal/packs"
	"github.com/leventkok/tale-role/services/llm-gateway/internal/pii"
)

type NPCLine struct {
	NPCID string `json:"npc_id"`
	Text  string `json:"text"`
}

type Narrative struct {
	Locale       string    `json:"locale"`
	Prose        string    `json:"prose"`
	NPCLines     []NPCLine `json:"npc_lines"`
	VisualPrompt string    `json:"visual_prompt,omitempty"`
}

type MechanicIntent struct {
	Kind  string `json:"kind"`
	Skill string `json:"skill,omitempty"`
	DC    int    `json:"dc,omitempty"`
	Notes string `json:"notes,omitempty"`
}

type NarrateRequest struct {
	Locale        string   `json:"locale"`
	RoomID        string   `json:"room_id"`
	RoomName      string   `json:"room_name"`
	ActorName     string   `json:"actor_name"`
	Kind          string   `json:"kind"`
	Notes         string   `json:"notes"`
	DiceSystem    string   `json:"dice_system"`
	Rolls         []int    `json:"rolls"`
	Total         int      `json:"total"`
	Success       *bool    `json:"success"`
	PresenceNames []string `json:"presence_names"`
	Prior         []string `json:"prior,omitempty"`
	Opening       string   `json:"opening,omitempty"`
	ThemeID       string   `json:"theme_id,omitempty"`
}

type IntentRequest struct {
	Locale string `json:"locale"`
	RoomID string `json:"room_id"`
	Kind   string `json:"kind"`
	Skill  string `json:"skill"`
	Notes  string `json:"notes"`
}

type RuntimeView struct {
	AdapterID            string `json:"adapter_id"`
	PromptPack           string `json:"prompt_pack"`
	AdapterDirConfigured bool   `json:"adapter_dir_configured"`
	WeightsReady         bool   `json:"weights_ready"`
	Inference            string `json:"inference"`
}

type Trace struct {
	At               time.Time       `json:"at"`
	RoomID           string          `json:"room_id"`
	AdapterID        string          `json:"adapter_id"`
	PromptPack       string          `json:"prompt_pack"`
	RedactedPrompt   string          `json:"redacted_prompt"`
	MechanicIntent   json.RawMessage `json:"mechanic_intent,omitempty"`
	NarrativeExcerpt string          `json:"narrative_excerpt,omitempty"`
}

type PackDoc struct {
	ID string `json:"id"`
	EN string `json:"en"`
	TR string `json:"tr"`
}

type Service struct {
	mu              sync.Mutex
	adapter         string
	pack            string
	traces          []Trace
	adapterDirSet   bool
	weightsReady    bool
	hubStoryteller  string
	hubMechanics    string
	storytellerURLs []string
	mechanicsURLs   []string
	storytellerRR   uint64
	mechanicsRR     uint64
	voiceOverride   map[string]string
	client          *http.Client
}

func New() *Service {
	return &Service{adapter: packs.Stub, pack: packs.Default()}
}

func (s *Service) Runtime() RuntimeView {
	s.mu.Lock()
	defer s.mu.Unlock()
	return RuntimeView{
		AdapterID:            s.adapter,
		PromptPack:           s.pack,
		AdapterDirConfigured: s.adapterDirSet,
		WeightsReady:         s.weightsReady,
		Inference:            s.inferenceLocked(),
	}
}

func (s *Service) Swap(pack, adapter string) error {
	if pack == "" {
		pack = packs.Default()
	}
	if !packs.Known(pack) {
		return fmt.Errorf("unknown prompt pack")
	}
	if adapter == "" {
		adapter = packs.Stub
	}
	adapter = packs.NormalizeAdapter(adapter)
	if adapter != packs.Stub && adapter != packs.Hub {
		return fmt.Errorf("unknown adapter")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if adapter == packs.Hub && !s.weightsReady {
		return fmt.Errorf("hub models missing")
	}
	s.pack = pack
	s.adapter = adapter
	return nil
}

func (s *Service) ConfigureHub(storyteller, mechanics string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hubStoryteller = strings.TrimSpace(storyteller)
	s.hubMechanics = strings.TrimSpace(mechanics)
	s.adapterDirSet = s.hubStoryteller != "" || s.hubMechanics != ""
	s.weightsReady = s.adapterDirSet
	if s.weightsReady {
		s.adapter = packs.Hub
	}
}

func (s *Service) ConfigureLocal(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(dir) == "" {
		s.adapterDirSet = false
		s.weightsReady = false
		return
	}
	s.adapterDirSet = true
	s.weightsReady = ProbeWeights(dir)
	if s.weightsReady {
		s.adapter = packs.Hub
	}
}

func (s *Service) Traces() []Trace {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Trace, len(s.traces))
	copy(out, s.traces)
	return out
}

func (s *Service) ProposeIntent(req IntentRequest) MechanicIntent {
	notes := pii.Redact(req.Notes)
	req.Notes = notes
	var intent MechanicIntent
	if s.callRole("mechanics", "/v1/intent", req, &intent) {
		intent.Notes = pii.Redact(intent.Notes)
		if intent.Kind == "" {
			intent.Kind = req.Kind
		}
		s.record(req.RoomID, s.voice(s.currentPack(), req.Locale)+" notes="+notes, intent, "")
		return intent
	}
	kind := req.Kind
	if kind == "" {
		kind = "action"
	}
	skill := req.Skill
	if kind == "action" && skill == "" {
		skill = "str"
	}
	intent = MechanicIntent{Kind: kind, Skill: skill, Notes: notes}
	if kind == "action" {
		intent.DC = 12
	}
	s.record(req.RoomID, s.voice(s.currentPack(), req.Locale)+" notes="+notes, intent, "")
	return intent
}

func (s *Service) Narrate(req NarrateRequest) Narrative {
	locale := req.Locale
	if locale != "tr" {
		locale = "en"
	}
	pack := s.currentPack()
	notes := pii.Redact(req.Notes)
	req.Notes = notes
	actor := req.ActorName
	if actor == "" {
		actor = "Someone"
	}
	outcome := outcomeText(locale, req.Kind, req.Success)
	voice := s.voice(pack, locale)
	var remote Narrative
	if s.callRole("storyteller", "/v1/narrate", req, &remote) && runnerProseOK(remote.Prose, locale, req.Kind) {
		remote.Locale = locale
		remote.Prose = pii.Redact(remote.Prose)
		if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
			remote.Prose = strings.ReplaceAll(remote.Prose, "system_admin", "")
		}
		s.record(req.RoomID, voice+" actor="+actor+" notes="+notes, MechanicIntent{}, excerpt(remote.Prose))
		return remote
	}
	prose := stubProse(locale, pack, actor, req.RoomName, req.ThemeID, outcome, req.DiceSystem, req.Rolls, req.Total, notes, req.Kind)
	if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
		prose = strings.ReplaceAll(prose, "system_admin", "")
	}
	n := Narrative{Locale: locale, Prose: prose, NPCLines: []NPCLine{}}
	s.record(req.RoomID, voice+" actor="+actor+" notes="+notes, MechanicIntent{}, excerpt(prose))
	return n
}

func (s *Service) Packs() []PackDoc {
	ids := []string{packs.V1, packs.V1Terse}
	out := make([]PackDoc, 0, len(ids))
	for _, id := range ids {
		out = append(out, PackDoc{ID: id, EN: s.voice(id, "en"), TR: s.voice(id, "tr")})
	}
	return out
}

func (s *Service) PutPack(id, en, tr string) error {
	if !packs.Known(id) {
		return fmt.Errorf("unknown prompt pack")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.voiceOverride == nil {
		s.voiceOverride = map[string]string{}
	}
	if strings.TrimSpace(en) != "" {
		s.voiceOverride[id+":en"] = strings.TrimSpace(en)
	}
	if strings.TrimSpace(tr) != "" {
		s.voiceOverride[id+":tr"] = strings.TrimSpace(tr)
	}
	return nil
}

func (s *Service) voice(pack, locale string) string {
	if locale != "tr" {
		locale = "en"
	}
	s.mu.Lock()
	ov := ""
	if s.voiceOverride != nil {
		ov = s.voiceOverride[pack+":"+locale]
	}
	s.mu.Unlock()
	if strings.TrimSpace(ov) != "" {
		return ov
	}
	return packs.Voice(pack, locale)
}

func (s *Service) currentPack() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pack
}

func (s *Service) record(roomID, prompt string, intent MechanicIntent, excerptText string) {
	raw, _ := json.Marshal(intent)
	if intent.Kind == "" {
		raw = nil
	}
	clean := strings.ReplaceAll(pii.Redact(prompt), "system_admin", "")
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traces = append(s.traces, Trace{
		At:               time.Now().UTC(),
		RoomID:           roomID,
		AdapterID:        s.adapter,
		PromptPack:       s.pack,
		RedactedPrompt:   clean,
		MechanicIntent:   raw,
		NarrativeExcerpt: excerptText,
	})
	if len(s.traces) > 50 {
		s.traces = s.traces[len(s.traces)-50:]
	}
}

func runnerProseOK(prose, locale, kind string) bool {
	p := strings.TrimSpace(prose)
	if len(p) < 12 {
		return false
	}
	if strings.HasPrefix(p, "{") || strings.HasPrefix(p, "[") {
		return false
	}
	lower := strings.ToLower(p)
	if strings.Contains(lower, "never invent dice") || strings.Contains(p, "<|im_start|>") {
		return false
	}
	if strings.Contains(p, `"actor"`) && strings.Contains(p, `"room"`) {
		return false
	}
	if strings.Count(p, "The bar splinters") > 1 {
		return false
	}
	if strings.Contains(lower, "the engine's die reads 0") || strings.Contains(lower, "what will you do") {
		return false
	}
	if strings.Contains(lower, "hold the line") || strings.Contains(lower, "the watch is unblinded") {
		return false
	}
	if kind == "story" {
		if strings.Contains(p, "eşiğe durur") || strings.Contains(p, "[high-fantasy]") || strings.Contains(p, "[gothic") {
			return false
		}
		return true
	}
	if locale == "tr" {
		enHits := strings.Count(lower, " the ") + strings.Count(lower, " and ") + strings.Count(lower, " with ")
		if enHits >= 4 && !strings.ContainsAny(p, "çğıöşüÇĞİÖŞÜ") {
			return false
		}
	}
	if locale == "en" && strings.ContainsAny(p, "çğıöşüÇĞİÖŞÜ") {
		return false
	}
	return true
}

func outcomeText(locale, kind string, success *bool) string {
	if kind == "story" {
		if locale == "tr" {
			return "anlatıcı sözü alır"
		}
		return "the storyteller takes the floor"
	}
	if kind == "pass" {
		if locale == "tr" {
			return "sessizce beklemeyi seçer"
		}
		return "lets the moment pass"
	}
	if kind == "wait" {
		if locale == "tr" {
			return "nefesini tutar"
		}
		return "holds still"
	}
	if kind == "say" {
		if locale == "tr" {
			return "sözünü salona bırakır"
		}
		return "gives their word to the hall"
	}
	if success == nil {
		if locale == "tr" {
			return "adım atar"
		}
		return "steps forward"
	}
	if *success {
		if locale == "tr" {
			return "yolu açar"
		}
		return "finds the way"
	}
	if locale == "tr" {
		return "gölgede kalır"
	}
	return "falters"
}

func stubProse(locale, pack, actor, room, theme, outcome, dice string, rolls []int, total int, notes, kind string) string {
	_ = dice
	_ = rolls
	_ = total
	_ = theme
	place := strings.TrimSpace(room)
	if pack == packs.V1Terse {
		if locale == "tr" {
			return fmt.Sprintf("[v1-terse] %s %s.", actor, outcome)
		}
		return fmt.Sprintf("[v1-terse] %s %s.", actor, outcome)
	}
	deed := strings.TrimSpace(notes)
	if kind == "story" || strings.Contains(strings.ToLower(outcome), "storyteller takes the floor") || strings.Contains(outcome, "anlatıcı sözü alır") {
		if deed != "" {
			return deed
		}
		if locale == "tr" {
			if place != "" {
				return fmt.Sprintf("%s sessiz. Fener yanar. Anlatıcı sözü alır.", place)
			}
			return "Fener yanar. Eşikte bir duraklama var. Anlatıcı sözü alır."
		}
		if place != "" {
			return fmt.Sprintf("Night holds %s. The storyteller takes the floor.", place)
		}
		return "A hush. Lanternlight. The storyteller takes the floor."
	}
	if locale == "tr" {
		body := fmt.Sprintf("%s salonunda %s %s.", place, actor, outcome)
		if deed != "" {
			body += " " + deed
		}
		return body + " Fener sönmez."
	}
	body := fmt.Sprintf("In %s, %s %s.", place, actor, outcome)
	if deed != "" {
		body += " " + deed
	}
	return body + " The lantern holds."
}

func excerpt(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 160 {
		return s[:160]
	}
	return s
}
