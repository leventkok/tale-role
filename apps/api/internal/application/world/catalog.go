package world

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
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
	ID        string `json:"id"`
	NameEN    string `json:"name_en"`
	NameTR    string `json:"name_tr,omitempty"`
	Alignment string `json:"alignment"`
	Voice     string `json:"voice,omitempty"`
}

type Universe struct {
	ID                string    `json:"id"`
	OwnerID           string    `json:"owner_id"`
	NameEN            string    `json:"name_en"`
	NameTR            string    `json:"name_tr,omitempty"`
	ThemeID           string    `json:"theme_id"`
	DiceSystem        string    `json:"dice_system"`
	RulesetID         string    `json:"ruleset_id"`
	PromptPackVersion string    `json:"prompt_pack_version"`
	ContentRating     string    `json:"content_rating"`
	Era               string    `json:"era,omitempty"`
	Tone              string    `json:"tone,omitempty"`
	Taboos            string    `json:"taboos,omitempty"`
	NPCs              []NPC     `json:"npcs"`
	CompiledPrompt    string    `json:"compiled_prompt,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
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
	Taboos        string
	NPCs          []NPC
}

type Catalog struct {
	mu    sync.Mutex
	items map[string]*Universe
}

func NewCatalog() *Catalog {
	return &Catalog{items: map[string]*Universe{}}
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
		Taboos:            strings.TrimSpace(d.Taboos),
		NPCs:              npcs,
		CreatedAt:         time.Now().UTC(),
	}
	u.CompiledPrompt = Compile(*u)
	c.mu.Lock()
	c.items[u.ID] = u
	c.mu.Unlock()
	return u, nil
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

func (c *Catalog) GetForHost(id, hostID string) (*Universe, error) {
	return c.Get(id, hostID)
}

func Compile(u Universe) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Tale Core. Theme %s. Dice %s. Rating %s.\n", u.ThemeID, u.DiceSystem, u.ContentRating)
	if u.Era != "" {
		fmt.Fprintf(&b, "Era: %s.\n", u.Era)
	}
	if u.Tone != "" {
		fmt.Fprintf(&b, "Tone: %s.\n", u.Tone)
	}
	if u.Taboos != "" {
		fmt.Fprintf(&b, "Taboos: %s.\n", u.Taboos)
	}
	if len(u.NPCs) > 0 {
		b.WriteString("NPCs:\n")
		for _, n := range u.NPCs {
			fmt.Fprintf(&b, "- %s (%s)", n.NameEN, n.Alignment)
			if n.Voice != "" {
				fmt.Fprintf(&b, ": %s", n.Voice)
			}
			b.WriteByte('\n')
		}
	}
	b.WriteString("Never invent dice or HP. Omit invisible spectators.\n")
	return b.String()
}
