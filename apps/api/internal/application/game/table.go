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
)

type Stats struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
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
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Stats  Stats  `json:"stats"`
	HP     int    `json:"hp"`
}

type Member struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // player | gm | system_admin
}

type Turn struct {
	ActorID    string     `json:"actor_id"`
	Kind       string     `json:"kind"`
	DiceSystem string     `json:"dice_system"`
	Rolls      []int      `json:"rolls"`
	Total      int        `json:"total"`
	Success    *bool      `json:"success"`
	Notes      string     `json:"notes,omitempty"`
	Narrative  *Narrative `json:"narrative,omitempty"`
}

type Narrative struct {
	Locale   string    `json:"locale"`
	Prose    string    `json:"prose"`
	NPCLines []NPCLine `json:"npc_lines"`
}

type NPCLine struct {
	NPCID string `json:"npc_id"`
	Text  string `json:"text"`
}

type Scene struct {
	ThemeID      string `json:"theme_id"`
	VisualPrompt string `json:"visual_prompt"`
	ImageSVG     string `json:"image_svg"`
	Inference    string `json:"inference"`
}

type Room struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	HostID     string                `json:"host_id"`
	DiceSystem string                `json:"dice_system"`
	JoinMode   string                `json:"join_mode"` // link | password
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
	Scene             *Scene      `json:"scene,omitempty"`
}

type Table struct {
	mu    sync.Mutex
	rooms map[string]*Room
	roll  func(sides int) int
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
	if joinMode == "" {
		joinMode = "link"
	}
	if joinMode != "link" && joinMode != "password" {
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
	return nil
}

func (t *Table) SetCharacter(roomID, userID, name string, stats Stats) error {
	if !stats.Valid() || strings.TrimSpace(name) == "" {
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
	r.Characters[userID] = &Character{
		UserID: userID,
		Name:   name,
		Stats:  stats,
		HP:     8 + stats.CON,
	}
	return nil
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
	type scored struct {
		id    string
		total int
	}
	var rows []scored
	for id, ch := range r.Characters {
		roll := t.roll(20)
		rows = append(rows, scored{id: id, total: roll + ch.Stats.DEX})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].total > rows[j].total })
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		order = append(order, row.id)
	}
	r.TurnOrder = order
	r.Started = true
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
	if kind == "" {
		kind = "action"
	}
	turn := Turn{ActorID: userID, Kind: kind, DiceSystem: r.DiceSystem, Notes: notes}
	if kind == "pass" || kind == "wait" {
		r.Turns = append(r.Turns, turn)
		return turn, nil
	}
	if dc <= 0 {
		dc = DefaultActionDC
	}
	bonus, err := ch.Stats.Skill(skill)
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
	r.Turns = append(r.Turns, turn)
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
