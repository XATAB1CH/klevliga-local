package http

import (
	"context"

	authDomain "github.com/XATAB1CH/klev-liga/internal/domains/auth"
)

type ctxKey int

const (
	ctxKeyRequestID ctxKey = iota
	ctxKeyUser
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, id)
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyRequestID).(string)

	return id
}

func withUser(ctx context.Context, u *authDomain.User) context.Context {
	return context.WithValue(ctx, ctxKeyUser, u)
}

func currentUser(ctx context.Context) (*authDomain.User, bool) {
	u, ok := ctx.Value(ctxKeyUser).(*authDomain.User)

	return u, ok
}
