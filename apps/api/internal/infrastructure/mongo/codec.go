package mongostore

import (
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/application/game"
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
)

type userDoc struct {
	ID           string    `bson:"_id"`
	Email        string    `bson:"email"`
	PasswordHash []byte    `bson:"password_hash"`
	Verified     bool      `bson:"verified"`
	TOTPSecret   string    `bson:"totp_secret,omitempty"`
	TOTPEnabled  bool      `bson:"totp_enabled,omitempty"`
	LanternXP    int       `bson:"lantern_xp,omitempty"`
	LanternLevel int       `bson:"lantern_level,omitempty"`
	PortraitID   string    `bson:"portrait_id,omitempty"`
	CreatedAt    time.Time `bson:"created_at"`
}

func (d userDoc) user() *iam.User {
	return &iam.User{
		ID: d.ID, Email: d.Email, PasswordHash: d.PasswordHash, Verified: d.Verified,
		TOTPSecret: d.TOTPSecret, TOTPEnabled: d.TOTPEnabled,
		LanternXP: d.LanternXP, LanternLevel: d.LanternLevel, PortraitID: d.PortraitID, CreatedAt: d.CreatedAt,
	}
}

type otpDoc struct {
	Email     string    `bson:"_id"`
	Hash      []byte    `bson:"hash"`
	ExpiresAt time.Time `bson:"expires_at"`
	Attempts  int       `bson:"attempts"`
}

type licDoc struct {
	ID        string    `bson:"_id"`
	UserID    string    `bson:"user_id"`
	DeviceID  string    `bson:"device_id"`
	Platform  string    `bson:"platform"`
	CreatedAt time.Time `bson:"created_at"`
}

type roomDoc struct {
	ID         string           `bson:"_id"`
	Name       string           `bson:"name"`
	HostID     string           `bson:"host_id"`
	DiceSystem string           `bson:"dice_system"`
	JoinMode   string           `bson:"join_mode"`
	Password   string           `bson:"password,omitempty"`
	Members    []game.Member    `bson:"members"`
	Characters []game.Character `bson:"characters"`
	TurnOrder  []string         `bson:"turn_order"`
	Turns      []game.Turn      `bson:"turns"`
	Started    bool             `bson:"started"`
	Completed  bool             `bson:"completed,omitempty"`
	StartedAt  time.Time        `bson:"started_at,omitempty"`
	UniverseID string           `bson:"universe_id,omitempty"`
	ThemeID    string           `bson:"theme_id,omitempty"`
	PromptPack string           `bson:"prompt_pack_version,omitempty"`
	Scene      *game.Scene      `bson:"scene,omitempty"`
	Chronicle  []string         `bson:"chronicle,omitempty"`
	CreatedAt  time.Time        `bson:"created_at"`
}

func encodeRoom(r *game.Room) roomDoc {
	members := make([]game.Member, 0, len(r.Members))
	for _, m := range r.Members {
		members = append(members, m)
	}
	chars := make([]game.Character, 0, len(r.Characters))
	for _, ch := range r.Characters {
		if ch != nil {
			chars = append(chars, *ch)
		}
	}
	return roomDoc{
		ID: r.ID, Name: r.Name, HostID: r.HostID, DiceSystem: r.DiceSystem, JoinMode: r.JoinMode,
		Password: r.Password, Members: members, Characters: chars, TurnOrder: r.TurnOrder,
		Turns: r.Turns, Started: r.Started, Completed: r.Completed, StartedAt: r.StartedAt,
		UniverseID: r.UniverseID, ThemeID: r.ThemeID,
		PromptPack: r.PromptPack, Scene: r.Scene, Chronicle: r.Chronicle, CreatedAt: r.CreatedAt,
	}
}

func decodeRoom(d roomDoc) *game.Room {
	r := &game.Room{
		ID: d.ID, Name: d.Name, HostID: d.HostID, DiceSystem: d.DiceSystem, JoinMode: d.JoinMode,
		Password: d.Password, Members: map[string]game.Member{}, Characters: map[string]*game.Character{},
		TurnOrder: d.TurnOrder, Turns: d.Turns, Started: d.Started, Completed: d.Completed, StartedAt: d.StartedAt,
		UniverseID: d.UniverseID,
		ThemeID: d.ThemeID, PromptPack: d.PromptPack, Scene: d.Scene, Chronicle: d.Chronicle, CreatedAt: d.CreatedAt,
	}
	for _, m := range d.Members {
		r.Members[m.UserID] = m
	}
	for i := range d.Characters {
		ch := d.Characters[i]
		cp := ch
		r.Characters[ch.UserID] = &cp
	}
	return r
}
