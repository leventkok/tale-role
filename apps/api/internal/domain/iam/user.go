package iam

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash []byte
	Verified     bool
	CreatedAt    time.Time
}

type OTP struct {
	Email     string
	Hash      []byte
	ExpiresAt time.Time
	Attempts  int
}
