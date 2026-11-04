package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type ctxKey struct{}

func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(ctxKey{}).(pgx.Tx)

	return tx, ok
}
