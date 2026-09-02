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

type Service struct {
	mu             sync.Mutex
	adapter        string
	pack           string
	traces         []Trace
	adapterDirSet  bool
	weightsReady   bool
	hubStoryteller string
	hubMechanics   string
	storytellerURL string
	mechanicsURL   string
	client         *http.Client
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
	if base, ok := s.useLocal("mechanics"); ok {
		var intent MechanicIntent
		if err := s.callJSON(base+"/v1/intent", req, &intent); err == nil {
			intent.Notes = pii.Redact(intent.Notes)
			if intent.Kind == "" {
				intent.Kind = req.Kind
			}
			s.record(req.RoomID, packs.Voice(s.currentPack(), req.Locale)+" notes="+notes, intent, "")
			return intent
		}
	}
	kind := req.Kind
	if kind == "" {
		kind = "action"
	}
	skill := req.Skill
	if kind == "action" && skill == "" {
		skill = "str"
	}
	intent := MechanicIntent{Kind: kind, Skill: skill, Notes: notes}
	if kind == "action" {
		intent.DC = 12
	}
	s.record(req.RoomID, packs.Voice(s.currentPack(), req.Locale)+" notes="+notes, intent, "")
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
	voice := packs.Voice(pack, locale)
	if base, ok := s.useLocal("storyteller"); ok {
		var remote Narrative
		if err := s.callJSON(base+"/v1/narrate", req, &remote); err == nil && strings.TrimSpace(remote.Prose) != "" {
			remote.Locale = locale
			remote.Prose = pii.Redact(remote.Prose)
			if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
				remote.Prose = strings.ReplaceAll(remote.Prose, "system_admin", "")
			}
			s.record(req.RoomID, voice+" actor="+actor+" notes="+notes, MechanicIntent{}, excerpt(remote.Prose))
			return remote
		}
	}
	prose := stubProse(locale, pack, actor, req.RoomName, req.ThemeID, outcome, req.DiceSystem, req.Rolls, req.Total, notes)
	if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
		prose = strings.ReplaceAll(prose, "system_admin", "")
	}
	n := Narrative{Locale: locale, Prose: prose, NPCLines: []NPCLine{}}
	s.record(req.RoomID, voice+" actor="+actor+" notes="+notes, MechanicIntent{}, excerpt(prose))
	return n
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

func outcomeText(locale, kind string, success *bool) string {
	if kind == "pass" {
		if locale == "tr" {
			return "pas geçti"
		}
		return "passes"
	}
	if kind == "wait" {
		if locale == "tr" {
			return "bekliyor"
		}
		return "waits"
	}
	if success == nil {
		if locale == "tr" {
			return "hareket eder"
		}
		return "acts"
	}
	if *success {
		if locale == "tr" {
			return "isabet eder"
		}
		return "hits"
	}
	if locale == "tr" {
		return "kaçırır"
	}
	return "misses"
}

func stubProse(locale, pack, actor, room, theme, outcome, dice string, rolls []int, total int, notes string) string {
	rollBits := ""
	if len(rolls) > 0 {
		parts := make([]string, len(rolls))
		for i, n := range rolls {
			parts[i] = fmt.Sprintf("%d", n)
		}
		rollBits = strings.Join(parts, "+")
	}
	place := room
	if theme != "" {
		place = room + " [" + theme + "]"
	}
	if pack == packs.V1Terse {
		if locale == "tr" {
			return fmt.Sprintf("[v1-terse] %s %s.", actor, outcome)
		}
		return fmt.Sprintf("[v1-terse] %s %s.", actor, outcome)
	}
	if locale == "tr" {
		return fmt.Sprintf(
			"%s salonunda %s %s. Motorun zarı (%s %s, toplam %d) anlatıyı bağlar. %s",
			place, actor, outcome, dice, rollBits, total, notes,
		)
	}
	return fmt.Sprintf(
		"In %s, %s %s. The engine's dice (%s %s, total %d) bind the tale. %s",
		place, actor, outcome, dice, rollBits, total, notes,
	)
}

func excerpt(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 160 {
		return s[:160]
	}
	return s
}
