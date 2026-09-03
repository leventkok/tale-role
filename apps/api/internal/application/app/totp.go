package app

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func newTOTPSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "="), nil
}

func decodeTOTPSecret(secret string) ([]byte, error) {
	s := strings.ToUpper(strings.TrimSpace(secret))
	s = strings.ReplaceAll(s, " ", "")
	if n := len(s) % 8; n != 0 {
		s += strings.Repeat("=", 8-n)
	}
	return base32.StdEncoding.DecodeString(s)
}

func totpAt(secret string, counter int64) (string, error) {
	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return "", err
	}
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(counter))
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	off := sum[len(sum)-1] & 0x0f
	bin := int(sum[off]&0x7f)<<24 | int(sum[off+1])<<16 | int(sum[off+2])<<8 | int(sum[off+3])
	return fmt.Sprintf("%06d", bin%1_000_000), nil
}

func CodeNow(secret string) (string, error) {
	return totpAt(secret, time.Now().Unix()/30)
}

func totpValid(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}
	slot := now.Unix() / 30
	for _, w := range []int64{-1, 0, 1} {
		got, err := totpAt(secret, slot+w)
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

func otpauthURL(email, secret string) string {
	label := url.PathEscape("Tale Role:" + email)
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", "Tale Role")
	q.Set("algorithm", "SHA1")
	q.Set("digits", "6")
	q.Set("period", "30")
	return "otpauth://totp/" + label + "?" + q.Encode()
}
