package pii

import "regexp"

const Marker = "[redacted]"

var (
	emailRe  = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	digitsRe = regexp.MustCompile(`\b\d{13,19}\b`)
	phoneRe  = regexp.MustCompile(`\+?\d[\d\s().\-]{8,}\d`)
)

// Redact strips emails, long digit runs, and phone-like strings before a prompt leaves this service.
func Redact(s string) string {
	s = emailRe.ReplaceAllString(s, Marker)
	s = digitsRe.ReplaceAllString(s, Marker)
	s = phoneRe.ReplaceAllString(s, Marker)
	return s
}

func ContainsLeak(s string) bool {
	return emailRe.MatchString(s) || digitsRe.MatchString(s)
}

var askRe = regexp.MustCompile(`(?i)(e-?mail|e-?posta|telefon numar|phone number|cep telefon|tc kimlik|kredi kart|credit card|social security)`)

// AsksPersonal is true when the narrator tries to collect real-world identity.
func AsksPersonal(s string) bool {
	return askRe.MatchString(s)
}
