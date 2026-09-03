package game

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultDice     = "d20"
	StatBudget      = 18
	StatMin         = 1
	StatMax         = 6
	DefaultActionDC = 12
)

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrBadStats     = errors.New("invalid stats")
	ErrHasCharacter = errors.New("character exists")
	ErrNoCharacter  = errors.New("character required")
	ErrBadPassword  = errors.New("unauthorized")
	ErrUnknownDice  = errors.New("unknown dice system")
	ErrUnknownSkill = errors.New("unknown skill")
	ErrNotYourTurn  = errors.New("not your turn")
	ErrInitiative   = errors.New("initiative required")
	ErrHasInit      = errors.New("already rolled")
)

type Stats struct {
	STR int `json:"str" bson:"str"`
	DEX int `json:"dex" bson:"dex"`
	CON int `json:"con" bson:"con"`
	INT int `json:"int" bson:"int"`
	WIS int `json:"wis" bson:"wis"`
	CHA int `json:"cha" bson:"cha"`
}

func (s Stats) Total() int {
	return s.STR + s.DEX + s.CON + s.INT + s.WIS + s.CHA
}

func (s Stats) Valid() bool {
	vals := []int{s.STR, s.DEX, s.CON, s.INT, s.WIS, s.CHA}
	for _, v := range vals {
		if v < StatMin || v > StatMax {
			return false
		}
	}
	return s.Total() == StatBudget
}

func (s Stats) Skill(name string) (int, error) {
	switch strings.ToLower(name) {
	case "str":
		return s.STR, nil
	case "dex":
		return s.DEX, nil
	case "con":
		return s.CON, nil
	case "int":
		return s.INT, nil
	case "wis":
		return s.WIS, nil
	case "cha":
		return s.CHA, nil
	default:
		return 0, ErrUnknownSkill
	}
}

type Character struct {
	UserID        string   `json:"user_id" bson:"user_id"`
	Name          string   `json:"name" bson:"name"`
	Species       string   `json:"species,omitempty" bson:"species,omitempty"`
	Path          string   `json:"path,omitempty" bson:"path,omitempty"`
	Backstory     string   `json:"backstory,omitempty" bson:"backstory,omitempty"`
	Skills        []string `json:"skills,omitempty" bson:"skills,omitempty"`
	Stats         Stats    `json:"stats" bson:"stats"`
	HP            int      `json:"hp" bson:"hp"`
	XP            int      `json:"xp" bson:"xp"`
	Level         int      `json:"level" bson:"level"`
	Initiative    int      `json:"initiative" bson:"initiative"`
	HasInitiative bool     `json:"has_initiative" bson:"has_initiative"`
}

type Sheet struct {
	Name      string
	Species   string
	Path      string
	Backstory string
	Stats     Stats
	Skills    []string
}

type Member struct {
	UserID string `json:"user_id" bson:"user_id"`
	Role   string `json:"role" bson:"role"`
}

type Turn struct {
	ActorID    string     `json:"actor_id" bson:"actor_id"`
	Kind       string     `json:"kind" bson:"kind"`
	DiceSystem string     `json:"dice_system" bson:"dice_system"`
	Rolls      []int      `json:"rolls" bson:"rolls"`
	Total      int        `json:"total" bson:"total"`
	Success    *bool      `json:"success" bson:"success"`
	Notes      string     `json:"notes,omitempty" bson:"notes,omitempty"`
	Narrative  *Narrative `json:"narrative,omitempty" bson:"narrative,omitempty"`
}

type Narrative struct {
	Locale   string    `json:"locale" bson:"locale"`
	Prose    string    `json:"prose" bson:"prose"`
	NPCLines []NPCLine `json:"npc_lines" bson:"npc_lines"`
}

type NPCLine struct {
	NPCID string `json:"npc_id" bson:"npc_id"`
	Text  string `json:"text" bson:"text"`
}

type Scene struct {
	ThemeID      string `json:"theme_id" bson:"theme_id"`
	VisualPrompt string `json:"visual_prompt" bson:"visual_prompt"`
	ImageSVG     string `json:"image_svg" bson:"image_svg"`
	Inference    string `json:"inference" bson:"inference"`
}

type Room struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	HostID     string                `json:"host_id"`
	DiceSystem string                `json:"dice_system"`
	JoinMode   string                `json:"join_mode"` // public | password (legacy: link)
	Password   string                `json:"-"`
	Members    map[string]Member     `json:"-"`
	Characters map[string]*Character `json:"-"`
	TurnOrder  []string              `json:"turn_order"`
	Turns      []Turn                `json:"turns"`
	Started    bool                  `json:"started"`
	UniverseID string                `json:"universe_id,omitempty"`
	ThemeID    string                `json:"theme_id,omitempty"`
	PromptPack string                `json:"prompt_pack_version,omitempty"`
	Scene      *Scene                `json:"-"`
	CreatedAt  time.Time             `json:"created_at"`
}

type Lobby struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	JoinMode string `json:"join_mode"`
	Started  bool   `json:"started"`
	Seats    int    `json:"seats"`
}

type PublicRoom struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	HostID            string      `json:"host_id"`
	DiceSystem        string      `json:"dice_system"`
	JoinMode          string      `json:"join_mode"`
	Started           bool        `json:"started"`
	TurnOrder         []string    `json:"turn_order"`
	Presence          []Member    `json:"presence"`
	Characters        []Character `json:"characters"`
	Turns             []Turn      `json:"turns"`
	UniverseID        string      `json:"universe_id,omitempty"`
	ThemeID           string      `json:"theme_id,omitempty"`
	PromptPackVersion string      `json:"prompt_pack_version,omitempty"`
	CurrentActorID    string      `json:"current_actor_id,omitempty"`
	Scene             *Scene      `json:"scene,omitempty"`
}

type Table struct {
	mu    sync.Mutex
	rooms map[string]*Room
	roll  func(sides int) int
	sink  RoomSink
}

func NewTable() *Table {
	return &Table{
		rooms: map[string]*Room{},
		roll:  cryptoDie,
	}
}

func (t *Table) Create(hostID, name, joinMode, password, dice string) (*Room, error) {
	if name == "" || hostID == "" {
		return nil, ErrForbidden
	}
	if dice == "" {
		dice = DefaultDice
	}
	if dice != "d20" && dice != "2d6" {
		return nil, ErrUnknownDice
	}
	if joinMode == "" || joinMode == "link" {
		joinMode = "public"
	}
	if joinMode != "public" && joinMode != "password" {
		return nil, ErrForbidden
	}
	if joinMode == "password" && password == "" {
		return nil, ErrBadPassword
	}
	r := &Room{
		ID:         uuid.NewString(),
		Name:       name,
		HostID:     hostID,
		DiceSystem: dice,
		JoinMode:   joinMode,
		Password:   password,
		Members:    map[string]Member{hostID: {UserID: hostID, Role: "gm"}},
		Characters: map[string]*Character{},
		CreatedAt:  time.Now().UTC(),
	}
	t.mu.Lock()
	t.rooms[r.ID] = r
	t.mu.Unlock()
	t.persist(r)
	return r, nil
}

func (t *Table) BindUniverse(roomID, universeID, themeID, pack string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	r.UniverseID = universeID
	r.ThemeID = themeID
	r.PromptPack = pack
	t.persist(r)
	return nil
}

func (t *Table) Join(roomID, userID, password, role string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	if role != "system_admin" && r.JoinMode == "password" && r.Password != password {
		return ErrBadPassword
	}
	if _, exists := r.Members[userID]; exists {
		return nil
	}
	if role == "" {
		role = "player"
	}
	r.Members[userID] = Member{UserID: userID, Role: role}
	t.persist(r)
	return nil
}

func (t *Table) SetCharacter(roomID, userID, name string, stats Stats) error {
	return t.SetSheet(roomID, userID, Sheet{Name: name, Stats: stats})
}

func (t *Table) SetSheet(roomID, userID string, sheet Sheet) error {
	if !sheet.Stats.Valid() || strings.TrimSpace(sheet.Name) == "" || !ValidSkills(sheet.Skills) {
		return ErrBadStats
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	m, ok := r.Members[userID]
	if !ok || m.Role == "system_admin" {
		return ErrForbidden
	}
	if _, exists := r.Characters[userID]; exists {
		return ErrHasCharacter
	}
	skills := make([]string, 0, len(sheet.Skills))
	for _, id := range sheet.Skills {
		skills = append(skills, strings.ToLower(strings.TrimSpace(id)))
	}
	level := 1
	r.Characters[userID] = &Character{
		UserID:    userID,
		Name:      strings.TrimSpace(sheet.Name),
		Species:   strings.TrimSpace(sheet.Species),
		Path:      strings.TrimSpace(sheet.Path),
		Backstory: strings.TrimSpace(sheet.Backstory),
		Skills:    skills,
		Stats:     sheet.Stats,
		HP:        MaxHP(sheet.Stats, level),
		XP:        0,
		Level:     level,
	}
	t.persist(r)
	return nil
}

func (t *Table) SeatHero(roomID, userID string, hero Character) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	if _, exists := r.Characters[userID]; exists {
		return nil
	}
	m, ok := r.Members[userID]
	if !ok || m.Role == "system_admin" {
		return ErrForbidden
	}
	cp := hero
	cp.UserID = userID
	cp.HasInitiative = false
	cp.Initiative = 0
	if cp.Level < 1 {
		cp.Level = 1
	}
	if cp.HP <= 0 {
		cp.HP = MaxHP(cp.Stats, cp.Level)
	}
	r.Characters[userID] = &cp
	t.persist(r)
	return nil
}

func (t *Table) RollInitiative(roomID, userID string) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return 0, ErrNotFound
	}
	if r.Started {
		return 0, ErrForbidden
	}
	m, ok := r.Members[userID]
	if !ok || m.Role == "system_admin" {
		return 0, ErrForbidden
	}
	ch, ok := r.Characters[userID]
	if !ok {
		return 0, ErrNoCharacter
	}
	if ch.HasInitiative {
		return 0, ErrHasInit
	}
	sides := 20
	if r.DiceSystem == "2d6" {
		sides = 6
	}
	n := t.roll(sides)
	if r.DiceSystem == "2d6" {
		n += t.roll(6)
	}
	ch.Initiative = n
	ch.HasInitiative = true
	r.Turns = append(r.Turns, Turn{
		ActorID:    userID,
		Kind:       "init",
		DiceSystem: r.DiceSystem,
		Rolls:      []int{n},
		Total:      n,
	})
	t.persist(r)
	return n, nil
}

func (t *Table) Start(roomID, userID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	if r.HostID != userID {
		return ErrForbidden
	}
	if len(r.Characters) == 0 {
		return ErrNoCharacter
	}
	type scored struct {
		id    string
		total int
	}
	var rows []scored
	for id, ch := range r.Characters {
		if !ch.HasInitiative {
			return ErrInitiative
		}
		rows = append(rows, scored{id: id, total: ch.Initiative})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].total == rows[j].total {
			return rows[i].id < rows[j].id
		}
		return rows[i].total > rows[j].total
	})
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		order = append(order, row.id)
	}
	r.TurnOrder = order
	r.Started = true
	r.Turns = append(r.Turns, Turn{ActorID: "storyteller", Kind: "story", DiceSystem: r.DiceSystem})
	t.persist(r)
	return nil
}

func (t *Table) Act(roomID, userID, kind, skill, notes string, dc int) (Turn, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return Turn{}, ErrNotFound
	}
	m, ok := r.Members[userID]
	if !ok || m.Role == "system_admin" {
		return Turn{}, ErrForbidden
	}
	ch, ok := r.Characters[userID]
	if !ok {
		return Turn{}, ErrNoCharacter
	}
	if !r.Started {
		return Turn{}, ErrForbidden
	}
	if cur := currentActor(r); cur != "" && cur != userID {
		return Turn{}, ErrNotYourTurn
	}
	if kind == "" {
		kind = "say"
	}
	turn := Turn{ActorID: userID, Kind: kind, DiceSystem: r.DiceSystem, Notes: notes}
	if kind == "pass" || kind == "wait" {
		r.Turns = append(r.Turns, turn)
		t.persist(r)
		return turn, nil
	}
	if kind == "say" {
		ch.GrantXP(8)
		r.Turns = append(r.Turns, turn)
		t.persist(r)
		return turn, nil
	}
	if dc <= 0 {
		dc = DefaultActionDC
	}
	bonus, err := ch.CheckBonus(skill)
	if err != nil {
		return Turn{}, err
	}
	var rolls []int
	total := bonus
	if r.DiceSystem == "2d6" {
		a, b := t.roll(6), t.roll(6)
		rolls = []int{a, b}
		total += a + b
	} else {
		n := t.roll(20)
		rolls = []int{n}
		total += n
	}
	okHit := total >= dc
	turn.Rolls = rolls
	turn.Total = total
	turn.Success = &okHit
	if okHit {
		ch.GrantXP(15)
	} else {
		ch.GrantXP(5)
	}
	r.Turns = append(r.Turns, turn)
	t.persist(r)
	return turn, nil
}

func (t *Table) AttachScene(roomID string, sc Scene) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return ErrNotFound
	}
	cp := sc
	r.Scene = &cp
	t.persist(r)
	return nil
}

func (t *Table) AttachNarrative(roomID string, n Narrative) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok || len(r.Turns) == 0 {
		return ErrNotFound
	}
	cp := n
	r.Turns[len(r.Turns)-1].Narrative = &cp
	t.persist(r)
	return nil
}

func (t *Table) View(roomID, userID string) (*PublicRoom, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return nil, ErrNotFound
	}
	if _, ok := r.Members[userID]; !ok {
		return nil, ErrForbidden
	}
	return snapshot(r), nil
}

func (t *Table) Public(roomID string) (*PublicRoom, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return nil, ErrNotFound
	}
	return snapshot(r), nil
}

func (t *Table) Lobbies() []Lobby {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Lobby, 0, len(t.rooms))
	for _, r := range t.rooms {
		seats := 0
		for _, m := range r.Members {
			if m.Role != "system_admin" {
				seats++
			}
		}
		out = append(out, Lobby{
			ID: r.ID, Name: r.Name, JoinMode: r.JoinMode, Started: r.Started, Seats: seats,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func currentActor(r *Room) string {
	if !r.Started || len(r.TurnOrder) == 0 {
		return ""
	}
	n := 0
	for _, turn := range r.Turns {
		if turn.Kind == "story" || turn.Kind == "init" || turn.ActorID == "storyteller" {
			continue
		}
		n++
	}
	return r.TurnOrder[n%len(r.TurnOrder)]
}

func snapshot(r *Room) *PublicRoom {
	out := &PublicRoom{
		ID:                r.ID,
		Name:              r.Name,
		HostID:            r.HostID,
		DiceSystem:        r.DiceSystem,
		JoinMode:          r.JoinMode,
		Started:           r.Started,
		TurnOrder:         append([]string{}, r.TurnOrder...),
		Presence:          []Member{},
		Characters:        []Character{},
		Turns:             append([]Turn{}, r.Turns...),
		UniverseID:        r.UniverseID,
		ThemeID:           r.ThemeID,
		PromptPackVersion: r.PromptPack,
		CurrentActorID:    currentActor(r),
	}
	if r.Scene != nil {
		cp := *r.Scene
		out.Scene = &cp
	}
	for _, m := range r.Members {
		if m.Role == "system_admin" {
			continue
		}
		out.Presence = append(out.Presence, m)
	}
	for _, ch := range r.Characters {
		out.Characters = append(out.Characters, *ch)
	}
	return out
}

func cryptoDie(sides int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 1
	}
	return int(n.Int64()) + 1
}

func (t *Table) UseDie(fn func(sides int) int) {
	t.roll = fn
}

type ExportedRoom struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type ExportedCharacter struct {
	RoomID    string   `json:"room_id"`
	Name      string   `json:"name"`
	Species   string   `json:"species,omitempty"`
	Path      string   `json:"path,omitempty"`
	Backstory string   `json:"backstory,omitempty"`
	Skills    []string `json:"skills,omitempty"`
	HP        int      `json:"hp"`
	XP        int      `json:"xp"`
	Level     int      `json:"level"`
	Stats     Stats    `json:"stats"`
}

func (t *Table) ExportFor(userID string) (rooms []ExportedRoom, chars []ExportedCharacter) {
	t.mu.Lock()
	defer t.mu.Unlock()
	rooms = []ExportedRoom{}
	chars = []ExportedCharacter{}
	for _, r := range t.rooms {
		m, ok := r.Members[userID]
		if !ok {
			continue
		}
		rooms = append(rooms, ExportedRoom{ID: r.ID, Name: r.Name, Role: m.Role})
		if ch, ok := r.Characters[userID]; ok {
			chars = append(chars, ExportedCharacter{
				RoomID: r.ID, Name: ch.Name, Species: ch.Species, Path: ch.Path, Backstory: ch.Backstory,
				Skills: ch.Skills, HP: ch.HP, XP: ch.XP, Level: ch.Level, Stats: ch.Stats,
			})
		}
	}
	return rooms, chars
}

func (t *Table) ForgetUser(userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, r := range t.rooms {
		if r.HostID == userID {
			delete(t.rooms, id)
			t.remove(id)
			continue
		}
		delete(r.Members, userID)
		delete(r.Characters, userID)
		order := make([]string, 0, len(r.TurnOrder))
		for _, uid := range r.TurnOrder {
			if uid != userID {
				order = append(order, uid)
			}
		}
		r.TurnOrder = order
		for i := range r.Turns {
			if r.Turns[i].ActorID == userID {
				r.Turns[i].ActorID = "erased"
				r.Turns[i].Notes = ""
				r.Turns[i].Narrative = nil
			}
		}
		t.persist(r)
	}
}
