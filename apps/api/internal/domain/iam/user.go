package iam

import (
	"strings"
	"time"
)

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	Verified     bool
	TOTPSecret   string
	TOTPEnabled  bool
	LanternXP    int
	LanternLevel int
	PortraitID   string
	CreatedAt    time.Time
}

func KnownPortrait(id string) bool {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "warden", "blade", "sage", "ranger", "oath":
		return true
	default:
		return false
	}
}

func NormalizePortrait(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if KnownPortrait(id) {
		return id
	}
	return "warden"
}

func (u *User) GrantLantern(n int) {
	if u == nil || n <= 0 {
		return
	}
	if u.LanternLevel < 1 {
		u.LanternLevel = 1
	}
	u.LanternXP += n
	for u.LanternXP >= 100*u.LanternLevel {
		u.LanternXP -= 100 * u.LanternLevel
		u.LanternLevel++
	}
}

type OTP struct {
	Email     string
	Hash      []byte
	ExpiresAt time.Time
	Attempts  int
}
