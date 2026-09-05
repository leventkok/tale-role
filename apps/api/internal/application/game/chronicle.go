package game

import (
	"strings"
	"unicode/utf8"
)

func SceneMemory(prose string) []string {
	low := strings.ToLower(prose)
	out := make([]string, 0, 6)
	if strings.Contains(low, "oym") || strings.Contains(low, "carving") {
		out = append(out, "Duvarlardaki oymalar uğulduyor; taşın altında bir nefes var.")
	}
	if strings.Contains(low, "koridor") || strings.Contains(low, "corridor") {
		out = append(out, "Karanlık bir koridor açık; bir yerde metal yere sürtünüyor.")
	}
	if strings.Contains(low, "torba") || strings.Contains(low, "bag") {
		out = append(out, "Yerinde bir torba duruyor.")
	}
	if strings.Contains(low, "madalyon") || strings.Contains(low, "medallion") {
		out = append(out, "Bir madalyonda Uyan yazıyor.")
	}
	if strings.Contains(low, "çatlak") || strings.Contains(low, "mavi") || strings.Contains(low, "crack") || strings.Contains(low, "blue light") {
		out = append(out, "Tavandan soluk mavi ışık sızıyor; çatlak nefes alıyor.")
	}
	return out
}

func BeatMemory(actor, notes string, success *bool) string {
	deed := strings.TrimSpace(notes)
	low := strings.ToLower(deed)
	if i := strings.Index(low, "deed:"); i >= 0 {
		deed = strings.TrimSpace(deed[i+len("deed:"):])
		if j := strings.Index(strings.ToLower(deed), "narrate this"); j >= 0 {
			deed = strings.TrimSpace(deed[:j])
		}
		deed = strings.Trim(deed, ".")
	}
	if deed == "" {
		return ""
	}
	if utf8.RuneCountInString(deed) > 90 {
		r := []rune(deed)
		deed = strings.TrimSpace(string(r[:90]))
	}
	mark := "sonuç belirsiz"
	if success != nil {
		if *success {
			mark = "isabet"
		} else {
			mark = "kaçırdı"
		}
	}
	line := strings.TrimSpace(actor + ": " + deed + " (" + mark + ").")
	if utf8.RuneCountInString(line) > 140 {
		r := []rune(line)
		line = strings.TrimSpace(string(r[:140]))
	}
	return line
}

func (t *Table) Chronicle(roomID string) []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return nil
	}
	if len(r.Chronicle) > 0 {
		return append([]string{}, r.Chronicle...)
	}
	return deriveChronicle(r)
}

func (t *Table) Remember(roomID string, lines ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r, ok := t.rooms[roomID]
	if !ok {
		return
	}
	for _, line := range lines {
		rememberLocked(r, line)
	}
	t.persist(r)
}

func rememberLocked(r *Room, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	low := strings.ToLower(line)
	for _, old := range r.Chronicle {
		if strings.ToLower(old) == low {
			return
		}
	}
	r.Chronicle = append(r.Chronicle, line)
	r.Chronicle = capChronicle(r.Chronicle)
}

func capChronicle(rows []string) []string {
	pinned := make([]string, 0, 5)
	beats := make([]string, 0, 8)
	for _, row := range rows {
		if isBeatMemory(row) {
			beats = append(beats, row)
		} else {
			pinned = append(pinned, row)
		}
	}
	if len(pinned) > 5 {
		pinned = pinned[len(pinned)-5:]
	}
	room := 8 - len(pinned)
	if room < 0 {
		room = 0
	}
	if len(beats) > room {
		beats = beats[len(beats)-room:]
	}
	return append(pinned, beats...)
}

func isBeatMemory(line string) bool {
	return strings.Contains(line, " (") && strings.HasSuffix(line, ").")
}

func deriveChronicle(r *Room) []string {
	out := []string{}
	for _, turn := range r.Turns {
		if turn.Kind == "story" && turn.Narrative != nil {
			out = append(out, SceneMemory(turn.Narrative.Prose)...)
			break
		}
	}
	for _, turn := range r.Turns {
		if turn.Kind != "action" {
			continue
		}
		name := turn.ActorID
		if ch := r.Characters[turn.ActorID]; ch != nil && strings.TrimSpace(ch.Name) != "" {
			name = ch.Name
		}
		if line := BeatMemory(name, turn.Notes, turn.Success); line != "" {
			out = append(out, line)
		}
	}
	return capChronicle(out)
}
