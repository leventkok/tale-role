package app

// Mailer delivers a one-time code. Implementations must never log the code.
type Mailer interface {
	SendOTP(email, code string) error
}
