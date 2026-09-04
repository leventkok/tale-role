package world

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/leventkok/tale-role/apps/api/internal/application/game"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrForbidden = errors.New("forbidden")
	ErrInvalid   = errors.New("invalid request")
)

var themes = map[string]struct{}{
	"high-fantasy": {}, "gothic-horror": {}, "space-opera": {},
	"cyber-noir": {}, "post-apocalyptic": {}, "fairytale": {},
}

type NPC struct {
	ID        string `json:"id" bson:"id"`
	NameEN    string `json:"name_en" bson:"name_en"`
	NameTR    string `json:"name_tr,omitempty" bson:"name_tr,omitempty"`
	Alignment string `json:"alignment" bson:"alignment"`
	Voice     string `json:"voice,omitempty" bson:"voice,omitempty"`
}

type Hero struct {
	UserID    string     `json:"user_id" bson:"user_id"`
	Name      string     `json:"name" bson:"name"`
	Species   string     `json:"species,omitempty" bson:"species,omitempty"`
	Path      string     `json:"path,omitempty" bson:"path,omitempty"`
	Backstory string     `json:"backstory,omitempty" bson:"backstory,omitempty"`
	Skills    []string   `json:"skills,omitempty" bson:"skills,omitempty"`
	Stats     game.Stats `json:"stats" bson:"stats"`
	HP        int        `json:"hp" bson:"hp"`
	XP        int        `json:"xp" bson:"xp"`
	Level     int        `json:"level" bson:"level"`
}

type Universe struct {
	ID                string    `json:"id" bson:"_id"`
	OwnerID           string    `json:"owner_id" bson:"owner_id"`
	NameEN            string    `json:"name_en" bson:"name_en"`
	NameTR            string    `json:"name_tr,omitempty" bson:"name_tr,omitempty"`
	ThemeID           string    `json:"theme_id" bson:"theme_id"`
	DiceSystem        string    `json:"dice_system" bson:"dice_system"`
	RulesetID         string    `json:"ruleset_id" bson:"ruleset_id"`
	PromptPackVersion string    `json:"prompt_pack_version" bson:"prompt_pack_version"`
	ContentRating     string    `json:"content_rating" bson:"content_rating"`
	Era               string    `json:"era,omitempty" bson:"era,omitempty"`
	Tone              string    `json:"tone,omitempty" bson:"tone,omitempty"`
	Description       string    `json:"description,omitempty" bson:"description,omitempty"`
	Opening           string    `json:"opening,omitempty" bson:"opening,omitempty"`
	Taboos            string    `json:"taboos,omitempty" bson:"taboos,omitempty"`
	NPCs              []NPC     `json:"npcs" bson:"npcs"`
	Heroes            []Hero    `json:"heroes,omitempty" bson:"heroes,omitempty"`
	CompiledPrompt    string    `json:"compiled_prompt,omitempty" bson:"compiled_prompt,omitempty"`
	CreatedAt         time.Time `json:"created_at" bson:"created_at"`
}

type Summary struct {
	ID                string `json:"id"`
	NameEN            string `json:"name_en"`
	ThemeID           string `json:"theme_id"`
	DiceSystem        string `json:"dice_system"`
	PromptPackVersion string `json:"prompt_pack_version"`
}

type Draft struct {
	NameEN        string
	NameTR        string
	ThemeID       string
	DiceSystem    string
	ContentRating string
	Era           string
	Tone          string
	Description   string
	Opening       string
	Taboos        string
	NPCs          []NPC
}

type Catalog struct {
	mu    sync.Mutex
	items map[string]*Universe
	sink  UniverseSink
}

type UniverseSink interface {
	UpsertUniverse(*Universe) error
	DeleteUniverse(id string) error
}

func NewCatalog() *Catalog {
	return &Catalog{items: map[string]*Universe{}}
}

func (c *Catalog) SetSink(s UniverseSink) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sink = s
}

func (c *Catalog) Load(rows []*Universe) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, u := range rows {
		if u == nil || u.ID == "" {
			continue
		}
		cp := *u
		c.items[u.ID] = &cp
	}
}

func (c *Catalog) persist(u *Universe) {
	if c.sink == nil || u == nil {
		return
	}
	cp := *u
	_ = c.sink.UpsertUniverse(&cp)
}

func (c *Catalog) Create(ownerID string, d Draft) (*Universe, error) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(d.NameEN) == "" {
		return nil, ErrInvalid
	}
	if _, ok := themes[d.ThemeID]; !ok {
		return nil, ErrInvalid
	}
	dice := d.DiceSystem
	if dice == "" {
		dice = "d20"
	}
	if dice != "d20" && dice != "2d6" {
		return nil, ErrInvalid
	}
	rating := d.ContentRating
	if rating == "" {
		rating = "teen"
	}
	if rating != "everyone" && rating != "teen" && rating != "mature" {
		return nil, ErrInvalid
	}
	npcs := make([]NPC, 0, len(d.NPCs))
	for _, n := range d.NPCs {
		name := strings.TrimSpace(n.NameEN)
		if name == "" {
			continue
		}
		al := n.Alignment
		if al == "" {
			al = "neutral"
		}
		if al != "good" && al != "evil" && al != "neutral" {
			return nil, ErrInvalid
		}
		id := n.ID
		if id == "" {
			id = uuid.NewString()
		}
		npcs = append(npcs, NPC{ID: id, NameEN: name, NameTR: strings.TrimSpace(n.NameTR), Alignment: al, Voice: strings.TrimSpace(n.Voice)})
	}
	u := &Universe{
		ID:                uuid.NewString(),
		OwnerID:           ownerID,
		NameEN:            strings.TrimSpace(d.NameEN),
		NameTR:            strings.TrimSpace(d.NameTR),
		ThemeID:           d.ThemeID,
		DiceSystem:        dice,
		RulesetID:         "tale-core",
		PromptPackVersion: "v1",
		ContentRating:     rating,
		Era:               strings.TrimSpace(d.Era),
		Tone:              strings.TrimSpace(d.Tone),
		Description:       strings.TrimSpace(d.Description),
		Opening:           strings.TrimSpace(d.Opening),
		Taboos:            strings.TrimSpace(d.Taboos),
		NPCs:              npcs,
		CreatedAt:         time.Now().UTC(),
	}
	u.CompiledPrompt = Compile(*u)
	c.mu.Lock()
	c.items[u.ID] = u
	c.mu.Unlock()
	c.persist(u)
	return u, nil
}

func (c *Catalog) UpsertHero(universeID string, h Hero) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, ok := c.items[universeID]
	if !ok {
		return ErrNotFound
	}
	out := make([]Hero, 0, len(u.Heroes)+1)
	replaced := false
	for _, row := range u.Heroes {
		if row.UserID == h.UserID {
			out = append(out, h)
			replaced = true
			continue
		}
		out = append(out, row)
	}
	if !replaced {
		out = append(out, h)
	}
	u.Heroes = out
	c.persist(u)
	return nil
}

func (c *Catalog) Hero(universeID, userID string) (Hero, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, ok := c.items[universeID]
	if !ok {
		return Hero{}, false
	}
	for _, row := range u.Heroes {
		if row.UserID == userID {
			return row, true
		}
	}
	return Hero{}, false
}

func (c *Catalog) ForgetPlayer(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, u := range c.items {
		kept := u.Heroes[:0]
		changed := false
		for _, h := range u.Heroes {
			if h.UserID == userID {
				changed = true
				continue
			}
			kept = append(kept, h)
		}
		if changed {
			u.Heroes = kept
			c.persist(u)
		}
	}
}

func (c *Catalog) List(ownerID string) []Summary {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := []Summary{}
	for _, u := range c.items {
		if u.OwnerID != ownerID {
			continue
		}
		out = append(out, Summary{ID: u.ID, NameEN: u.NameEN, ThemeID: u.ThemeID, DiceSystem: u.DiceSystem, PromptPackVersion: u.PromptPackVersion})
	}
	return out
}

func (c *Catalog) SceneSeed(id string) (opening, description, nameEN, nameTR string, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, found := c.items[id]
	if !found {
		return "", "", "", "", false
	}
	return u.Opening, u.Description, u.NameEN, u.NameTR, true
}

func (c *Catalog) Get(id, userID string) (*Universe, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, ok := c.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	if u.OwnerID != userID {
		return nil, ErrForbidden
	}
	cp := *u
	return &cp, nil
}

func (c *Catalog) PublicName(id, locale string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, ok := c.items[id]
	if !ok {
		return "", false
	}
	if locale == "tr" && u.NameTR != "" {
		return u.NameTR, true
	}
	return u.NameEN, true
}

func (c *Catalog) GetForHost(id, hostID string) (*Universe, error) {
	return c.Get(id, hostID)
}

func (c *Catalog) ExportFor(ownerID string) []Universe {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := []Universe{}
	for _, u := range c.items {
		if u.OwnerID != ownerID {
			continue
		}
		cp := *u
		out = append(out, cp)
	}
	return out
}

func (c *Catalog) ForgetOwner(ownerID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, u := range c.items {
		if u.OwnerID == ownerID {
			delete(c.items, id)
			if c.sink != nil {
				_ = c.sink.DeleteUniverse(id)
			}
		}
	}
}

var lookName = map[string]string{
	"high-fantasy":     "high fantasy",
	"gothic-horror":    "gothic horror",
	"space-opera":      "space opera",
	"cyber-noir":       "cyber noir",
	"post-apocalyptic": "a ruined world",
	"fairytale":        "fairytale",
}

func (c *Catalog) TableBrief(id string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u, ok := c.items[id]
	if !ok {
		return "", false
	}
	return NarrationBrief(*u), true
}

func clipBrief(s string, n int) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if n <= 0 || len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n]))
}

func NarrationBrief(u Universe) string {
	var b strings.Builder
	name := strings.TrimSpace(u.NameEN)
	if name != "" {
		fmt.Fprintf(&b, "World: %s\n", name)
	}
	if look := lookName[u.ThemeID]; look != "" {
		fmt.Fprintf(&b, "Look: %s\n", look)
	}
	if u.Era != "" {
		fmt.Fprintf(&b, "Age: %s\n", clipBrief(u.Era, 200))
	}
	if u.Tone != "" {
		fmt.Fprintf(&b, "Mood: %s\n", clipBrief(u.Tone, 200))
	}
	if u.Description != "" {
		fmt.Fprintf(&b, "The tale of this place:\n%s\n", clipBrief(u.Description, 1200))
	}
	if u.Opening != "" {
		fmt.Fprintf(&b, "Opening scene:\n%s\n", clipBrief(u.Opening, 1200))
	}
	if u.Taboos != "" {
		fmt.Fprintf(&b, "Do not depict: %s\n", clipBrief(u.Taboos, 400))
	}
	if len(u.NPCs) > 0 {
		b.WriteString("People of this place:\n")
		for i, n := range u.NPCs {
			if i >= 8 {
				break
			}
			label := n.NameEN
			if n.NameTR != "" && n.NameTR != n.NameEN {
				label = n.NameEN + " / " + n.NameTR
			}
			fmt.Fprintf(&b, "- %s (%s)", label, n.Alignment)
			if n.Voice != "" {
				fmt.Fprintf(&b, ": %s", clipBrief(n.Voice, 240))
			}
			b.WriteByte('\n')
		}
	}
	return strings.TrimSpace(b.String())
}

func Compile(u Universe) string {
	var b strings.Builder
	fmt.Fprintf(&b, "World %s. Mood %s.\n", u.NameEN, u.ThemeID)
	if u.Era != "" {
		fmt.Fprintf(&b, "Age: %s.\n", u.Era)
	}
	if u.Tone != "" {
		fmt.Fprintf(&b, "Feeling: %s.\n", u.Tone)
	}
	if u.Description != "" {
		fmt.Fprintf(&b, "The tale of this place:\n%s\n", u.Description)
	}
	if u.Opening != "" {
		fmt.Fprintf(&b, "Opening scene:\n%s\n", u.Opening)
	}
	if u.Taboos != "" {
		fmt.Fprintf(&b, "Do not depict: %s.\n", u.Taboos)
	}
	if len(u.NPCs) > 0 {
		b.WriteString("People of this place:\n")
		for _, n := range u.NPCs {
			fmt.Fprintf(&b, "- %s (%s)", n.NameEN, n.Alignment)
			if n.Voice != "" {
				fmt.Fprintf(&b, ": %s", n.Voice)
			} else {
				b.WriteString(": no notes — adapt to the tale")
			}
			b.WriteByte('\n')
		}
	}
	b.WriteString("Never invent dice or HP. Omit invisible spectators.\n")
	return b.String()
}
