package memory

import (
	"strings"
	"sync"
	"time"

	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
	"github.com/leventkok/tale-role/apps/api/internal/domain/license"
)

type Store struct {
	mu       sync.RWMutex
	users    map[string]*iam.User
	otp      map[string]*iam.OTP
	licenses map[string]*license.ProductLicense
}

func NewStore() *Store {
	return &Store{
		users:    map[string]*iam.User{},
		otp:      map[string]*iam.OTP{},
		licenses: map[string]*license.ProductLicense{},
	}
}

func (s *Store) PutUser(u *iam.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[strings.ToLower(u.Email)] = u
}

func (s *Store) GetUser(email string) (*iam.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[strings.ToLower(email)]
	return u, ok
}

func (s *Store) GetUserByID(id string) (*iam.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.ID == id {
			return u, true
		}
	}
	return nil, false
}

func (s *Store) PutOTP(o *iam.OTP) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.otp[strings.ToLower(o.Email)] = o
}

func (s *Store) GetOTP(email string) (*iam.OTP, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.otp[strings.ToLower(email)]
	if !ok || time.Now().After(o.ExpiresAt) {
		return nil, false
	}
	return o, true
}

func (s *Store) DeleteOTP(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.otp, strings.ToLower(email))
}

func (s *Store) PutLicense(l *license.ProductLicense) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.licenses[l.UserID+":"+l.DeviceID] = l
}

func (s *Store) LicensesForUser(userID string) []*license.ProductLicense {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*license.ProductLicense{}
	for _, l := range s.licenses {
		if l.UserID == userID {
			out = append(out, l)
		}
	}
	return out
}
