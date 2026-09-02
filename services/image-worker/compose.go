package worker

import (
	"fmt"
	"regexp"
	"strings"
)

const Marker = "[redacted]"

var (
	emailRe  = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	digitsRe = regexp.MustCompile(`\b\d{13,19}\b`)
)

var themes = map[string][2]string{
	"high-fantasy":     {"#1c3d2e", "#c9a227"},
	"gothic-horror":    {"#1a1018", "#8b6b9e"},
	"space-opera":      {"#071428", "#4ec6e8"},
	"cyber-noir":       {"#0a0612", "#e23e8c"},
	"post-apocalyptic": {"#2a1a0c", "#d9843a"},
	"fairytale":        {"#3a2040", "#f0c4d8"},
}

type Request struct {
	ThemeID  string
	RoomName string
	Notes    string
	Prose    string
}

type Card struct {
	ThemeID      string `json:"theme_id"`
	VisualPrompt string `json:"visual_prompt"`
	ImageSVG     string `json:"image_svg"`
	Inference    string `json:"inference"`
}

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Compose(req Request) Card {
	theme := req.ThemeID
	if _, ok := themes[theme]; !ok {
		theme = "high-fantasy"
	}
	prompt := buildPrompt(theme, req.RoomName, req.Notes, req.Prose)
	return Card{
		ThemeID:      theme,
		VisualPrompt: prompt,
		ImageSVG:     stubSVG(theme),
		Inference:    "stub",
	}
}

func buildPrompt(theme, room, notes, prose string) string {
	room = redact(strings.TrimSpace(room))
	if room == "" {
		room = "the table"
	}
	extra := redact(strings.TrimSpace(notes))
	if extra == "" {
		extra = redact(strings.TrimSpace(prose))
	}
	if len(extra) > 180 {
		extra = extra[:180]
	}
	extra = strings.ReplaceAll(extra, "system_admin", "")
	var b strings.Builder
	fmt.Fprintf(&b, "Tale Role stub scene. Theme %s. Place %s. No dice, HP, or spectators.", theme, room)
	if extra != "" {
		fmt.Fprintf(&b, " Mood: %s", extra)
	}
	return b.String()
}

func stubSVG(theme string) string {
	pair := themes[theme]
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 500" role="img" data-theme="%s">`+
			`<title>%s</title>`+
			`<rect width="400" height="500" fill="%s"/>`+
			`<circle cx="200" cy="150" r="54" fill="%s" opacity="0.75"/>`+
			`<text x="200" y="455" text-anchor="middle" fill="%s" font-size="20" font-family="Georgia,serif">%s</text>`+
			`</svg>`,
		theme, theme, pair[0], pair[1], pair[1], theme,
	)
}

func redact(s string) string {
	s = emailRe.ReplaceAllString(s, Marker)
	s = digitsRe.ReplaceAllString(s, Marker)
	s = strings.ReplaceAll(s, "system_admin", "")
	return s
}
