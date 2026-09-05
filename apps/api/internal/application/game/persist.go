package game

type RoomSink interface {
	UpsertRoom(*Room) error
	DeleteRoom(id string) error
}

func CloneRoom(r *Room) *Room {
	if r == nil {
		return nil
	}
	cp := *r
	cp.TurnOrder = append([]string{}, r.TurnOrder...)
	cp.Turns = append([]Turn{}, r.Turns...)
	cp.Members = map[string]Member{}
	for k, v := range r.Members {
		cp.Members[k] = v
	}
	cp.Characters = map[string]*Character{}
	for k, v := range r.Characters {
		if v == nil {
			continue
		}
		ch := *v
		cp.Characters[k] = &ch
	}
	if r.Scene != nil {
		sc := *r.Scene
		cp.Scene = &sc
	}
	cp.Chronicle = append([]string{}, r.Chronicle...)
	return &cp
}

func (t *Table) SetSink(s RoomSink) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sink = s
}

func (t *Table) Load(rooms []*Room) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, r := range rooms {
		if r == nil || r.ID == "" {
			continue
		}
		if r.Members == nil {
			r.Members = map[string]Member{}
		}
		if r.Characters == nil {
			r.Characters = map[string]*Character{}
		}
		t.rooms[r.ID] = r
	}
}

func (t *Table) persist(r *Room) {
	if t.sink == nil || r == nil {
		return
	}
	_ = t.sink.UpsertRoom(CloneRoom(r))
}

func (t *Table) remove(id string) {
	if t.sink == nil || id == "" {
		return
	}
	_ = t.sink.DeleteRoom(id)
}
