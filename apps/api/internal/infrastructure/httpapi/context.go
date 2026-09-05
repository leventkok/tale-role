package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/leventkok/tale-role/apps/api/internal/application/app"
	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
)

type ctxKey int

const userKey ctxKey = 1
const deviceKey ctxKey = 2

func withUser(ctx context.Context, u *iam.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func withDeviceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, deviceKey, id)
}

func userFrom(r interface{ Context() context.Context }) *iam.User {
	u, _ := r.Context().Value(userKey).(*iam.User)
	return u
}

func deviceIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(deviceKey).(string)
	return id
}

func attachDevice(r *http.Request) *http.Request {
	id := strings.TrimSpace(r.Header.Get("X-TaleRole-Device"))
	if id == "" {
		return r
	}
	return r.WithContext(withDeviceID(r.Context(), id))
}

func (s *Server) denyUnlicensedPlay(user *iam.User, device string) error {
	if strings.TrimSpace(device) == "" {
		return nil
	}
	if user == nil {
		return app.ErrUnauthorized
	}
	if s.adminEmail != "" && user.Email == s.adminEmail {
		return nil
	}
	return s.svc.RequireDesktopLicense(user.ID, device)
}

func (s *Server) denyUnlicensedPlayHTTP(w http.ResponseWriter, r *http.Request) bool {
	err := s.denyUnlicensedPlay(userFrom(r), deviceIDFromCtx(r.Context()))
	if err == nil {
		return false
	}
	s.writeAppError(w, err)
	return true
}
