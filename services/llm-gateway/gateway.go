package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/leventkok/tale-role/services/llm-gateway/internal/packs"
	"github.com/leventkok/tale-role/services/llm-gateway/internal/pii"
)

// Redact strips emails, phones, and long digit runs from table text before it is stored or shown.
func Redact(s string) string {
	return pii.Redact(s)
}

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
	Kind     string `json:"kind"`
	Skill    string `json:"skill,omitempty"`
	DC       int    `json:"dc,omitempty"`
	Pressure int    `json:"pressure,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type NarrateRequest struct {
	Locale        string       `json:"locale"`
	RoomID        string       `json:"room_id"`
	RoomName      string       `json:"room_name"`
	ActorName     string       `json:"actor_name"`
	Kind          string       `json:"kind"`
	Notes         string       `json:"notes"`
	DiceSystem    string       `json:"dice_system"`
	Rolls         []int        `json:"rolls"`
	Total         int          `json:"total"`
	Success       *bool        `json:"success"`
	PresenceNames []string     `json:"presence_names"`
	Prior         []string     `json:"prior,omitempty"`
	Opening       string       `json:"opening,omitempty"`
	ThemeID       string       `json:"theme_id,omitempty"`
	WorldBrief    string       `json:"world_brief,omitempty"`
	Cast          []CastMember `json:"cast,omitempty"`
	Facts         []string     `json:"facts,omitempty"`
	Skill         string       `json:"skill,omitempty"`
}

type CastMember struct {
	Name      string `json:"name"`
	Species   string `json:"species,omitempty"`
	Path      string `json:"path,omitempty"`
	Backstory string `json:"backstory,omitempty"`
	STR       int    `json:"str,omitempty"`
	DEX       int    `json:"dex,omitempty"`
	CON       int    `json:"con,omitempty"`
	INT       int    `json:"int,omitempty"`
	WIS       int    `json:"wis,omitempty"`
	CHA       int    `json:"cha,omitempty"`
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
	LiveStoryteller      string `json:"live_storyteller,omitempty"`
	CandidateStoryteller string `json:"candidate_storyteller,omitempty"`
	CandidateReady       bool   `json:"candidate_ready"`
}

type LiveView struct {
	RoomID    string `json:"room_id"`
	Prose     string `json:"prose"`
	Pack      string `json:"prompt_pack"`
	Adapter   string `json:"adapter_id"`
	Streaming bool   `json:"streaming"`
}

type PackStore interface {
	SavePromptPack(id, en, tr string) error
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
	packStore       PackStore
	proseHook       func(roomID, prose string, done bool)
	live            map[string]LiveView
	liveStoryteller string
	liveMechanics   string
	candStoryteller string
	candMechanics   string
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
		LiveStoryteller:      s.liveStoryteller,
		CandidateStoryteller: s.candStoryteller,
		CandidateReady:       strings.TrimSpace(s.candStoryteller) != "",
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
	if adapter != packs.Stub && adapter != packs.Hub && adapter != packs.Candidate {
		return fmt.Errorf("unknown adapter")
	}
	s.mu.Lock()
	if adapter == packs.Hub && !s.weightsReady {
		s.mu.Unlock()
		return fmt.Errorf("hub models missing")
	}
	if adapter == packs.Candidate && strings.TrimSpace(s.candStoryteller) == "" {
		s.mu.Unlock()
		return fmt.Errorf("candidate adapter unset")
	}
	s.pack = pack
	s.adapter = adapter
	s.mu.Unlock()
	go s.reloadSlot(adapter)
	return nil
}

func (s *Service) reloadSlot(adapter string) {
	if adapter == packs.Stub {
		return
	}
	s.mu.Lock()
	story := s.liveStoryteller
	if adapter == packs.Candidate {
		story = s.candStoryteller
	}
	urls := append([]string{}, s.storytellerURLs...)
	s.mu.Unlock()
	if story == "" {
		return
	}
	for _, u := range urls {
		var ignored map[string]any
		_ = s.callJSON(u+"/v1/load", map[string]string{"model_id": story}, &ignored)
	}
}

func (s *Service) ConfigureHub(storyteller, mechanics string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hubStoryteller = strings.TrimSpace(storyteller)
	s.hubMechanics = strings.TrimSpace(mechanics)
	s.liveStoryteller = s.hubStoryteller
	s.liveMechanics = s.hubMechanics
	s.adapterDirSet = s.hubStoryteller != "" || s.hubMechanics != ""
	s.weightsReady = s.adapterDirSet
	if s.weightsReady {
		s.adapter = packs.Hub
	}
}

func (s *Service) ConfigureCandidate(storyteller, mechanics string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.candStoryteller = strings.TrimSpace(storyteller)
	s.candMechanics = strings.TrimSpace(mechanics)
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
	s.record(req.RoomID, s.voice(s.currentPack(), req.Locale)+" notes="+notes, intent, "")
	return intent
}

// PressureFrom is the hidden world-side bonus. Players never see it. 0–8.
func PressureFrom(intent MechanicIntent) int {
	n := intent.Pressure
	if n == 0 && intent.DC > 0 {
		n = intent.DC - 10
	}
	if n < 0 {
		n = 0
	}
	if n > 8 {
		n = 8
	}
	return n
}

func (s *Service) Narrate(req NarrateRequest) Narrative {
	notes := pii.Redact(req.Notes)
	req.Notes = notes
	req.Opening = pii.Redact(req.Opening)
	req.WorldBrief = pii.Redact(req.WorldBrief)
	req.Cast = redactCast(req.Cast)
	locale := taleLocale(req.Locale, req.Opening, notes)
	pack := s.currentPack()
	actor := req.ActorName
	if actor == "" {
		actor = "Someone"
	}
	outcome := outcomeText(locale, req.Kind, req.Success)
	voice := s.voice(pack, locale)
	req.Locale = locale
	var remote Narrative
	host := strings.TrimSpace(req.Opening + " " + notes)
	s.publishLive(req.RoomID, "", false)
	if s.streamNarrate(req, &remote) && runnerProseOK(remote.Prose, locale, req.Kind, req.RoomName, host, req.Success, notes) {
		remote.Locale = locale
		remote.Prose = pii.Redact(remote.Prose)
		if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
			remote.Prose = strings.ReplaceAll(remote.Prose, "system_admin", "")
		}
		s.publishLive(req.RoomID, remote.Prose, true)
		s.record(req.RoomID, voice+" actor="+actor+" notes="+notes, MechanicIntent{}, excerpt(remote.Prose))
		return remote
	}
	if s.callRole("storyteller", "/v1/narrate", req, &remote) && runnerProseOK(remote.Prose, locale, req.Kind, req.RoomName, host, req.Success, notes) {
		remote.Locale = locale
		remote.Prose = pii.Redact(remote.Prose)
		if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
			remote.Prose = strings.ReplaceAll(remote.Prose, "system_admin", "")
		}
		s.publishLive(req.RoomID, remote.Prose, true)
		s.record(req.RoomID, voice+" actor="+actor+" notes="+notes, MechanicIntent{}, excerpt(remote.Prose))
		return remote
	}
	prose := stubProse(locale, pack, actor, req.RoomName, req.ThemeID, outcome, req.DiceSystem, req.Rolls, req.Total, notes, req.Kind, req.Prior, req.Opening, req.WorldBrief, req.Facts)
	if strings.Contains(strings.ToLower(strings.Join(req.PresenceNames, " ")), "system_admin") {
		prose = strings.ReplaceAll(prose, "system_admin", "")
	}
	s.publishLive(req.RoomID, prose, true)
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
	if s.voiceOverride == nil {
		s.voiceOverride = map[string]string{}
	}
	if strings.TrimSpace(en) != "" {
		s.voiceOverride[id+":en"] = strings.TrimSpace(en)
	}
	if strings.TrimSpace(tr) != "" {
		s.voiceOverride[id+":tr"] = strings.TrimSpace(tr)
	}
	store := s.packStore
	s.mu.Unlock()
	if store != nil {
		return store.SavePromptPack(id, strings.TrimSpace(en), strings.TrimSpace(tr))
	}
	return nil
}

func (s *Service) SetPackStore(store PackStore) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.packStore = store
}

func (s *Service) WatchProse(hook func(roomID, prose string, done bool)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proseHook = hook
}

func (s *Service) Live(roomID string) LiveView {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.live == nil {
		return LiveView{RoomID: roomID, Pack: s.pack, Adapter: s.adapter}
	}
	v, ok := s.live[roomID]
	if !ok {
		return LiveView{RoomID: roomID, Pack: s.pack, Adapter: s.adapter}
	}
	return v
}

func (s *Service) publishLive(roomID, prose string, done bool) {
	if strings.TrimSpace(roomID) == "" {
		return
	}
	s.mu.Lock()
	if s.live == nil {
		s.live = map[string]LiveView{}
	}
	view := LiveView{
		RoomID: roomID, Prose: pii.Redact(prose), Pack: s.pack, Adapter: s.adapter, Streaming: !done,
	}
	s.live[roomID] = view
	hook := s.proseHook
	s.mu.Unlock()
	if hook != nil && strings.TrimSpace(view.Prose) != "" {
		hook(roomID, view.Prose, done)
	}
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

func runnerProseOK(prose, locale, kind, tableTitle, host string, success *bool, notes string) bool {
	p := strings.TrimSpace(prose)
	if len(p) < 12 {
		return false
	}
	if pii.AsksPersonal(p) {
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
	if trainingSalad(lower) {
		return false
	}
	if tableTitleLeak(p, tableTitle, host) {
		return false
	}
	if missLooksLikeHit(lower, success) {
		return false
	}
	if actorMovedAgainstDeed(lower, notes) {
		return false
	}
	if kind != "story" && clauseSalad(p) {
		return false
	}
	if kind == "story" && utf8.RuneCountInString(p) < 90 {
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

func missLooksLikeHit(lower string, success *bool) bool {
	if success == nil || *success {
		return false
	}
	tells := []string{
		"without hesitation",
		"recognizing the",
		"the way opens",
		"follows through",
		"tereddüt etmeden",
		"çekinmeden",
	}
	for _, tell := range tells {
		if strings.Contains(lower, tell) {
			return true
		}
	}
	return false
}

func stayPut(notes string) bool {
	low := strings.ToLower(notes)
	tells := []string{
		"without going", "without taking a step", "stay where", "staying where",
		"without walking", "without stepping", "without moving", "don't go", "do not go",
		"remain here", "yerinde kal", "yerimde kal", "yerde kal", "olduğum yerde", "oldugum yerde",
		"adım atmadan", "adım atmıyor", "koridora inmeden", "koridora inmiyor", "inmeden",
	}
	for _, t := range tells {
		if strings.Contains(low, t) {
			return true
		}
	}
	return false
}

func actorMovedAgainstDeed(lower, notes string) bool {
	if !stayPut(notes) {
		return false
	}
	tells := []string{
		"steps into", "steps toward", "he walks", "she walks", "they walk",
		"walks into", "walked down", "walks down", "enters the corridor",
		"leading him deeper", "leading her deeper", "adım atar", "koridora iner",
	}
	for _, t := range tells {
		if strings.Contains(lower, t) {
			return true
		}
	}
	return false
}

func trainingSalad(lower string) bool {
	frags := []string{
		"nöbet dönmez", "nöbet dönüyor", "rün karanlık", "kilit durur", "pim kopar",
		"kahkaha bitince", "kahkaha kopar", "menteşe", "zar ", " der.",
		"yol açılır", "taş cevap verir", "taş susar",
		"hamleyi tamamlar:", "hamleyi kaçırır:",
		"sayı ", "the count is", "uzaktan bir ses sahneyi",
		"çandan önce", "gelene dek", "bir sonraki çan",
		"alet kayar", "zaman biter", " direnir.",
		"motorun", "hold the line", "the watch is unblinded",
		"the bar splinters", "the die reads", "the engine's", "the rune stays dark",
		"the latch yields", "a pin snaps", "what will you do",
	}
	hits := 0
	for _, f := range frags {
		if strings.Contains(lower, f) {
			hits++
		}
	}
	return hits >= 1
}

func tableTitleLeak(prose, title, host string) bool {
	t := strings.TrimSpace(title)
	if t == "" {
		return false
	}
	if utf8.RuneCountInString(t) < 6 && !strings.ContainsAny(t, " \t-") {
		return false
	}
	tl := strings.ToLower(t)
	if strings.Contains(strings.ToLower(host), tl) {
		return false
	}
	return strings.Contains(strings.ToLower(prose), tl)
}

func clauseSalad(p string) bool {
	clauses := strings.FieldsFunc(p, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	n, words, longest := 0, 0, 0
	for _, c := range clauses {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		n++
		w := len(strings.Fields(c))
		words += w
		if w > longest {
			longest = w
		}
	}
	if n < 6 || words == 0 {
		return false
	}
	if longest >= 10 {
		return false
	}
	return words/n < 4
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

func taleLocale(ui, opening, notes string) string {
	for _, t := range []string{strings.TrimSpace(opening), strings.TrimSpace(notes)} {
		if utf8.RuneCountInString(t) < 12 {
			continue
		}
		if strings.ContainsAny(t, "çğıöşüÇĞİÖŞÜ") {
			return "tr"
		}
		low := " " + strings.ToLower(t) + " "
		if strings.Contains(low, " the ") || strings.Contains(low, " you ") || strings.Contains(low, " and ") {
			return "en"
		}
	}
	if ui == "tr" {
		return "tr"
	}
	return "en"
}

func redactCast(in []CastMember) []CastMember {
	if len(in) == 0 {
		return nil
	}
	out := make([]CastMember, 0, len(in))
	for _, row := range in {
		name := strings.TrimSpace(pii.Redact(row.Name))
		if name == "" || name == "system_admin" {
			continue
		}
		if len(out) >= 8 {
			break
		}
		back := pii.Redact(row.Backstory)
		if len(back) > 300 {
			back = strings.TrimSpace(back[:300])
		}
		out = append(out, CastMember{
			Name:      name,
			Species:   strings.TrimSpace(pii.Redact(row.Species)),
			Path:      strings.TrimSpace(pii.Redact(row.Path)),
			Backstory: strings.TrimSpace(back),
			STR:       row.STR, DEX: row.DEX, CON: row.CON,
			INT: row.INT, WIS: row.WIS, CHA: row.CHA,
		})
	}
	return out
}

func stubProse(locale, pack, actor, room, theme, outcome, dice string, rolls []int, total int, notes, kind string, prior []string, opening, worldBrief string, facts []string) string {
	_ = dice
	_ = rolls
	_ = theme
	_ = room
	if pack == packs.V1Terse {
		return fmt.Sprintf("[v1-terse] %s %s.", actor, outcome)
	}
	deed := strings.TrimSpace(notes)
	if kind == "story" || strings.Contains(strings.ToLower(outcome), "storyteller takes the floor") || strings.Contains(outcome, "anlatıcı sözü alır") {
		if deed != "" {
			return deed
		}
		if locale == "tr" {
			return "Fener yanar. Eşikte bir duraklama var. Uzak bir ses gelir. Anlatıcı sözü alır."
		}
		return "A hush. Lanternlight. The storyteller takes the floor."
	}
	loc := beatLocale(locale, deed)
	place := scenePlace(loc)
	ground := strings.TrimSpace(opening + " " + worldBrief + " " + strings.Join(facts, " "))
	if loc == "tr" {
		return literaryTR(kind, actor, place, deed, outcome, total, prior, ground)
	}
	return literaryEN(kind, actor, place, deed, outcome, total, prior)
}

func beatLocale(locale, notes string) string {
	if locale != "tr" {
		return "en"
	}
	n := strings.TrimSpace(notes)
	if n == "" {
		return "tr"
	}
	if strings.ContainsAny(n, "çğıöşüÇĞİÖŞÜ") {
		return "tr"
	}
	low := " " + strings.ToLower(n) + " "
	if strings.Contains(low, " the ") || strings.Contains(low, " and ") || strings.Contains(low, " examine ") {
		return "en"
	}
	return "tr"
}

func scenePlace(locale string) string {
	if locale == "tr" {
		return "salon"
	}
	return "the hall"
}

func pickUnused(lines []string, prior []string, total int) string {
	blob := strings.ToLower(strings.Join(prior, " "))
	unused := make([]string, 0, len(lines))
	for _, line := range lines {
		key := strings.ToLower(line)
		if len(key) > 28 {
			key = key[:28]
		}
		if !strings.Contains(blob, key) {
			unused = append(unused, line)
		}
	}
	if len(unused) == 0 {
		unused = lines
	}
	if len(unused) == 0 {
		return ""
	}
	i := total % len(unused)
	if i < 0 {
		i = -i
	}
	return unused[i]
}

var (
	firstPersonRe = regexp.MustCompile(`(?i)(yorum|mekteyim|maktayım|arım|erim|ırım|urum|ürüm)(\s|[.!?:]|$)`)
	benPrefix     = regexp.MustCompile(`(?i)^ben\s+`)
	imPrefix      = regexp.MustCompile(`(?i)^i['’]m\s+`)
	iPrefix       = regexp.MustCompile(`(?i)^i\s+`)
	yorumRe       = regexp.MustCompile(`yorum\b`)
	mekteyimRe    = regexp.MustCompile(`mekteyim\b`)
	maktayimRe    = regexp.MustCompile(`maktayım\b`)
	aoristMeRe    = regexp.MustCompile(`(arım|erim|ırım|urum|ürüm)\b`)
)

func playerDeed(notes string) string {
	text := strings.TrimSpace(notes)
	low := strings.ToLower(text)
	if i := strings.Index(low, "deed:"); i >= 0 {
		text = strings.TrimSpace(text[i+len("deed:"):])
		if j := strings.Index(strings.ToLower(text), "narrate this"); j >= 0 {
			text = strings.TrimSpace(text[:j])
		}
		text = strings.Trim(text, ".")
	}
	return strings.TrimSpace(text)
}

func firstPerson(notes string) bool {
	low := strings.ToLower(strings.TrimSpace(notes))
	if strings.HasPrefix(low, "i ") || strings.HasPrefix(low, "i'm ") || strings.HasPrefix(low, "i’m ") || strings.HasPrefix(low, "ben ") {
		return true
	}
	return firstPersonRe.MatchString(low)
}

func voiceDeed(actor, notes string) string {
	deed := playerDeed(notes)
	if deed == "" {
		return ""
	}
	text := deed
	if firstPerson(deed) {
		text = benPrefix.ReplaceAllString(deed, "")
		text = imPrefix.ReplaceAllString(text, "")
		text = iPrefix.ReplaceAllString(text, "")
		text = yorumRe.ReplaceAllString(text, "yor")
		text = mekteyimRe.ReplaceAllString(text, "mekte")
		text = maktayimRe.ReplaceAllString(text, "makta")
		text = aoristMeRe.ReplaceAllStringFunc(text, func(m string) string {
			r := []rune(m)
			if len(r) <= 2 {
				return m
			}
			return string(r[:len(r)-2])
		})
	}
	repls := [][2]string{
		{"çevremi", "çevresini"},
		{"etrafımı", "etrafını"},
		{"kılıncımı", "kılıcını"},
		{"kılıcımı", "kılıcını"},
		{"torbamı", "torbasını"},
		{"sırtıma", "sırtına"},
		{"sirtima", "sırtına"},
		{"yanımı", "yanını"},
		{"önümü", "önünü"},
		{"elimi", "elini"},
		{"başımı", "başını"},
		{"sesimi", "sesini"},
		{"gözlerimi", "gözlerini"},
	}
	for _, pair := range repls {
		text = strings.ReplaceAll(text, pair[0], pair[1])
	}
	text = strings.Trim(strings.TrimSpace(text), ".")
	if text == "" {
		return ""
	}
	lowActor := strings.ToLower(strings.TrimSpace(actor))
	if lowActor != "" && strings.HasPrefix(strings.ToLower(text), lowActor) {
		return text
	}
	runes := []rune(text)
	if len(runes) > 0 {
		lowered := []rune(strings.ToLower(string(runes[0])))
		if len(lowered) == 1 && (len(runes) == 1 || !unicode.IsUpper(runes[1])) {
			runes[0] = lowered[0]
			text = string(runes)
		}
	}
	return strings.TrimSpace(actor + " " + text)
}

func containsAny(low string, keys ...string) bool {
	for _, k := range keys {
		if strings.Contains(low, k) {
			return true
		}
	}
	return false
}

func lookingInBag(notes string) bool {
	low := strings.ToLower(notes)
	if containsAny(low, "kapat", "bağla") {
		return false
	}
	return containsAny(low, "içindekilere", "içine bak", "göz at", "torbaya bak", "look in", "inside the bag")
}

func bagDeed(notes string) bool {
	return lookingInBag(notes)
}

func carryingBag(notes string) bool {
	low := strings.ToLower(notes)
	if lookingInBag(notes) {
		return false
	}
	hasBag := containsAny(low, "torba", "çanta", "bag", "pouch")
	return hasBag && containsAny(low, "sırt", "sirt", "omuz", "pick up", "alip", "alıp")
}

func groundTR(success bool, opening string, prior []string, total int, notes string) string {
	blob := strings.ToLower(opening)
	var lines []string
	if success {
		if strings.Contains(blob, "oym") {
			lines = append(lines, "Duvarlardaki oymalar uğuldar; taşın altındaki nefes kesilmez.")
		}
		if strings.Contains(blob, "koridor") {
			lines = append(lines, "Karanlık koridor aynı ağızla açık kalır.")
		}
		if strings.Contains(blob, "metal") {
			lines = append(lines, "Sürtünen metal yerini değiştirmez; henüz görünmez.")
		}
		if strings.Contains(blob, "madalyon") {
			lines = append(lines, "Madalyondaki yazı yerinde kalır.")
		}
		if strings.Contains(blob, "mavi") || strings.Contains(blob, "çatlak") {
			lines = append(lines, "Tavandan sızan mavi ışık çatlağı tutmaya devam eder.")
		}
	} else {
		if strings.Contains(blob, "oym") {
			lines = append(lines, "Oymaların uğultusu düşer, sonra daha alçak bir tondan döner.")
		}
		if strings.Contains(blob, "koridor") {
			lines = append(lines, "Koridorun dibinde metal bir karış daha sürtünür.")
		}
		if strings.Contains(blob, "metal") {
			lines = append(lines, "Koridordaki metal bir karış daha yaklaşır; kaynak hâlâ görünmez.")
		}
		if strings.Contains(blob, "madalyon") {
			lines = append(lines, "Madalyon soğur. Kazınmış yazı yerinde kalır.")
		}
		if strings.Contains(blob, "mavi") || strings.Contains(blob, "çatlak") {
			lines = append(lines, "Tavandaki mavi ışık bir an kesilir, çatlak yine nefes alır.")
		}
	}
	if len(lines) == 0 {
		if success {
			return pickUnused(hitLinesTR(""), prior, total)
		}
		return pickUnused(missLinesTR(""), prior, total)
	}
	return pickUnused(preferGround(lines, notes), prior, total)
}

func preferGround(lines []string, notes string) []string {
	deed := strings.ToLower(notes)
	var keys []string
	if containsAny(deed, "çatla", "mavi") {
		keys = append(keys, "çatlak", "mavi")
	}
	if strings.Contains(deed, "oym") {
		keys = append(keys, "oym")
	}
	if containsAny(deed, "koridor", "ilerle", "yürü", "walk") {
		keys = append(keys, "koridor", "metal")
	}
	if containsAny(deed, "madalyon", "medallion") {
		keys = append(keys, "madalyon")
	}
	if len(keys) == 0 {
		return lines
	}
	picked := make([]string, 0, len(lines))
	for _, ln := range lines {
		low := strings.ToLower(ln)
		for _, k := range keys {
			if strings.Contains(low, k) {
				picked = append(picked, ln)
				break
			}
		}
	}
	if len(picked) == 0 {
		return lines
	}
	return picked
}

func followTR(deed string, success bool, opening string, prior []string, total int) string {
	ground := groundTR(success, opening, prior, total, deed)
	low := strings.ToLower(deed)
	if lookingInBag(deed) {
		if success {
			return pickUnused(hitLinesTR(deed), prior, total)
		}
		return pickUnused(missLinesTR(deed), prior, total)
	}
	if strings.Contains(low, "kapat") && containsAny(low, "torba", "çanta", "ağz", "bag") {
		if success {
			return "Ağız kapanır. Dikkat odaya döner. " + ground
		}
		return "Ağız kapanır ama oda cevap vermez. " + ground
	}
	if carryingBag(deed) {
		if success {
			return "Torba sırta oturur. Dikkat odaya döner. " + ground
		}
		return "Torba omza kayar ama oda cevap vermez. " + ground
	}
	if containsAny(low, "bağır", "kim var", "seslen") {
		if success {
			return "Çağrı taşa çarpar. Bir şey kımıldar; henüz yüz göstermez. " + ground
		}
		return "Ses koridorda kırılır. Cevap gelmez. " + ground
	}
	if containsAny(low, "yürü", "koridor", "ilerle", "çatla") {
		if success {
			return "Koridor bir adım alır; oymalar yanından geçer. " + ground
		}
		return "Adım tamamlanmaz. " + ground
	}
	return ground
}

func hitLinesTR(deed string) []string {
	if bagDeed(deed) {
		return []string{
			"Torbanın ağzı açılır. Fener kumaşa, toza ve soğuk bir kenara düşer; adları henüz yok, ama artık görünürler.",
			"Ağız gevşer. İçeride kayış, katlanmış kumaş, parmak ucunda metal. Hiçbiri masayı kapatmaz.",
		}
	}
	return []string{
		"Fener hareketin üstüne düşer; görünen şey sahneyi değiştirir.",
		"Oda boyun eğer. Bir sonraki seçenek açık kalır.",
		"Parmakların altındaki dünya cevap verir; masa kapanmaz.",
	}
}

func missLinesTR(deed string) []string {
	if bagDeed(deed) {
		return []string{
			"Torba kayar; ağız kapanır. İçeride bir şey kumaşa takılır ve karanlıkta kalır.",
			"Parmaklar kumaşı bulur ama ağız bir an önce kapanır. Torba sırrını tutar.",
		}
	}
	return []string{
		"Koridordaki metal bir karış daha yaklaşır; kaynak hâlâ görünmez.",
		"Oymaların uğultusu düşer, sonra daha alçak bir tondan döner.",
		"Karanlıkta bir nefes tutulur. Cevap gelmez; sahne kapanmaz.",
		"Madalyon soğur. Kazınmış yazı yerinde kalır.",
		"Tavandaki mavi ışık bir an kesilir, çatlak yine nefes alır.",
		"Taşın altından kısa bir tık gelir. Sıra yine oyuncuda.",
	}
}

func hitLinesEN(deed string) []string {
	if bagDeed(deed) {
		return []string{
			"The bag's mouth opens. Cloth, dust, and a cold edge take the light; unnamed yet, but seen.",
			"The drawstring yields. A strap, folded cloth, metal under a fingertip. None of it closes the table.",
		}
	}
	return []string{
		"Light finds the motion; the room shifts to match what is now seen.",
		"The world yields. Another choice stays open.",
		"The thing under the hand answers; the table does not close.",
	}
}

func missLinesEN(deed string) []string {
	if bagDeed(deed) {
		return []string{
			"The bag slips; the mouth pinches shut. Something snags in the cloth and stays dark.",
			"Fingers find fabric, then the mouth closes first. The bag keeps its secret.",
		}
	}
	return []string{
		"Down the corridor, metal drags one pace closer.",
		"The carvings drop a note, then return lower.",
		"A breath holds in the dark. No answer yet.",
		"The medallion goes cold. The scratched word stays.",
		"Pale ceiling light dies for a beat, then returns.",
		"Stone ticks once underfoot. The next move is still yours.",
	}
}

func literaryTR(kind, actor, place, deed, outcome string, total int, prior []string, opening string) string {
	if place == "" {
		place = "salon"
	}
	switch kind {
	case "say":
		if deed != "" {
			return fmt.Sprintf("%s sözü salona bırakır: %s Fener sönmez.", actor, deed)
		}
		return fmt.Sprintf("%s %s'de sessizliği kırar. Fener sönmez.", actor, place)
	case "pass":
		return fmt.Sprintf("%s bu eli bırakır. %s bekler. Fener sönmez.", actor, place)
	case "wait":
		return fmt.Sprintf("%s %s'de nefesini tutar. Henüz hamle yok. Fener sönmez.", actor, place)
	}
	success := strings.Contains(outcome, "yolu açar")
	shift := followTR(deed, success, opening, prior, total)
	if stayPut(deed) {
		if success {
			return fmt.Sprintf("%s hamleyi yerinde tutar. %s Gitmediği yer açık kalmaz.", actor, shift)
		}
		return fmt.Sprintf("%s hamleyi kaçırır. %s Gitmediği yer açık kalmaz.", actor, shift)
	}
	if v := voiceDeed(actor, deed); v != "" {
		return fmt.Sprintf("%s. %s", v, shift)
	}
	if success {
		return fmt.Sprintf("%s. %s", actor, shift)
	}
	return fmt.Sprintf("%s hamleyi kaçırır. %s", actor, shift)
}

func literaryEN(kind, actor, place, deed, outcome string, total int, prior []string) string {
	if place == "" {
		place = "the hall"
	}
	switch kind {
	case "say":
		if deed != "" {
			return fmt.Sprintf("%s gives their word in %s: %s The lantern holds.", actor, place, deed)
		}
		return fmt.Sprintf("%s breaks the hush in %s. The lantern holds.", actor, place)
	case "pass":
		return fmt.Sprintf("%s lets the beat pass in %s. The lantern holds.", actor, place)
	case "wait":
		return fmt.Sprintf("%s holds still in %s. Breath only. The lantern holds.", actor, place)
	}
	if strings.Contains(outcome, "finds the way") {
		shift := pickUnused(hitLinesEN(deed), prior, total)
		if stayPut(deed) {
			return fmt.Sprintf("%s holds the beat. %s The place they refused stays unentered.", actor, shift)
		}
		if v := voiceDeed(actor, deed); v != "" {
			return fmt.Sprintf("%s. %s", v, shift)
		}
		return fmt.Sprintf("%s follows through. %s", actor, shift)
	}
	shift := pickUnused(missLinesEN(deed), prior, total)
	if stayPut(deed) {
		return fmt.Sprintf("%s misses. %s The place they refused stays unentered.", actor, shift)
	}
	if v := voiceDeed(actor, deed); v != "" {
		return fmt.Sprintf("%s. %s misses. %s", v, actor, shift)
	}
	return fmt.Sprintf("%s misses. %s", actor, shift)
}

func excerpt(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 160 {
		return s[:160]
	}
	return s
}
