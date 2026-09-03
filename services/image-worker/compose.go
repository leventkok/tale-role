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
	fmt.Fprintf(&b, "Tale Role painted tableau. Gold filigree frame. Theme %s. Place %s. No dice, HP, or spectators.", theme, room)
	if extra != "" {
		fmt.Fprintf(&b, " Mood: %s", extra)
	}
	return b.String()
}

func stubSVG(theme string) string {
	pair := themes[theme]
	bg, gold := pair[0], pair[1]
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 500" role="img" data-theme="%s" data-art="tableau">`+
			`<title>%s</title>`+
			`<defs>`+
			`<linearGradient id="sky" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="#0a0806"/></linearGradient>`+
			`<radialGradient id="glow" cx="50%%" cy="28%%" r="42%%"><stop offset="0" stop-color="%s" stop-opacity="0.5"/><stop offset="1" stop-color="%s" stop-opacity="0"/></radialGradient>`+
			`<linearGradient id="gold" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#f3e0a8"/><stop offset="0.5" stop-color="%s"/><stop offset="1" stop-color="#6e4e14"/></linearGradient>`+
			`</defs>`+
			`<rect width="400" height="500" fill="url(#sky)"/>`+
			`<rect width="400" height="500" fill="url(#glow)"/>`+
			`%s`+
			`<rect x="16" y="16" width="368" height="468" fill="none" stroke="url(#gold)" stroke-width="12"/>`+
			`<rect x="26" y="26" width="348" height="448" fill="none" stroke="%s" stroke-width="2" opacity="0.75"/>`+
			`<path d="M28 48 L48 28" stroke="#f3e0a8" stroke-width="3"/>`+
			`<path d="M372 48 L352 28" stroke="#f3e0a8" stroke-width="3"/>`+
			`<path d="M28 452 L48 472" stroke="#c9a227" stroke-width="3"/>`+
			`<path d="M372 452 L352 472" stroke="#c9a227" stroke-width="3"/>`+
			`<circle cx="28" cy="28" r="6" fill="#e2c07a"/>`+
			`<circle cx="372" cy="28" r="6" fill="#e2c07a"/>`+
			`<circle cx="28" cy="472" r="6" fill="#c9a227"/>`+
			`<circle cx="372" cy="472" r="6" fill="#c9a227"/>`+
			`<rect x="70" y="428" width="260" height="28" rx="4" fill="#120e0a" stroke="%s" stroke-width="1.5"/>`+
			`<text x="200" y="447" text-anchor="middle" fill="%s" font-size="14" font-family="Georgia,serif">%s</text>`+
			`</svg>`,
		theme, theme, bg, gold, bg, gold, landscape(theme, gold), gold, gold, gold, theme,
	)
}

func landscape(theme, gold string) string {
	switch theme {
	case "gothic-horror":
		return `<path d="M0 310 L70 220 L90 250 L130 180 L170 240 L220 150 L260 230 L310 170 L400 300 V500 H0z" fill="#120c14"/>` +
			`<path d="M210 150 L218 90 L226 150" fill="` + gold + `" opacity="0.35"/>` +
			`<ellipse cx="300" cy="110" rx="28" ry="28" fill="none" stroke="` + gold + `" stroke-width="2" opacity="0.5"/>`
	case "space-opera":
		return `<circle cx="80" cy="90" r="3" fill="` + gold + `"/><circle cx="320" cy="70" r="2" fill="#e8f4ff"/>` +
			`<circle cx="140" cy="130" r="1.5" fill="` + gold + `"/>` +
			`<ellipse cx="200" cy="210" rx="90" ry="18" fill="` + gold + `" opacity="0.25"/>` +
			`<path d="M40 360 Q200 280 360 360 L400 500 H0z" fill="#0b1824"/>`
	case "cyber-noir":
		return `<path d="M0 260 L60 200 L80 240 L140 160 L180 220 L250 140 L300 210 L360 170 L400 250 V500 H0z" fill="#101018"/>` +
			`<path d="M90 200 V320 M180 160 V340 M270 150 V330" stroke="` + gold + `" stroke-width="2" opacity="0.45"/>`
	case "post-apocalyptic":
		return `<path d="M0 300 L90 230 L140 270 L220 190 L300 260 L400 220 V500 H0z" fill="#1e1912"/>` +
			`<path d="M160 250 L190 180 L210 250" fill="` + gold + `" opacity="0.3"/>` +
			`<ellipse cx="70" cy="120" rx="40" ry="12" fill="` + gold + `" opacity="0.2"/>`
	case "fairytale":
		return `<path d="M0 340 Q80 240 160 320 Q240 220 320 310 Q360 250 400 300 V500 H0z" fill="#24182c"/>` +
			`<circle cx="70" cy="100" r="8" fill="` + gold + `" opacity="0.45"/>` +
			`<circle cx="310" cy="80" r="5" fill="` + gold + `" opacity="0.35"/>` +
			`<path d="M200 220 Q220 160 240 220 Q220 210 200 220" fill="` + gold + `" opacity="0.4"/>`
	default:
		return `<path d="M0 320 L80 210 L120 250 L180 160 L240 230 L300 150 L360 240 L400 210 V500 H0z" fill="#14241c"/>` +
			`<path d="M0 380 Q200 320 400 390 V500 H0z" fill="#1c3d2e"/>` +
			`<ellipse cx="80" cy="90" r="22" fill="` + gold + `" opacity="0.28"/>`
	}
}

func redact(s string) string {
	s = emailRe.ReplaceAllString(s, Marker)
	s = digitsRe.ReplaceAllString(s, Marker)
	s = strings.ReplaceAll(s, "system_admin", "")
	return s
}
