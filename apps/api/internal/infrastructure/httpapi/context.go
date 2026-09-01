package httpapi

import (
	"context"

	"github.com/leventkok/tale-role/apps/api/internal/domain/iam"
)

type ctxKey int

const userKey ctxKey = 1

func withUser(ctx context.Context, u *iam.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func userFrom(r interface{ Context() context.Context }) *iam.User {
	u, _ := r.Context().Value(userKey).(*iam.User)
	return u
}
