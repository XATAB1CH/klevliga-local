// Package repo — слой доступа к PostgreSQL поверх sqlc-адаптеров.
package repo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/sqlc"
	"github.com/XATAB1CH/klev-liga/pkg/pg"
)

// WithTx — интерфейс выполнения кода внутри транзакции.
type WithTx interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Store struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
		q:  sqlc.New(db),
	}
}

func (s *Store) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = pg.WithTx(ctx, tx)

	if err := fn(ctx); err != nil {
		_ = tx.Rollback(ctx)

		return err
	}

	return tx.Commit(ctx)
}

// queries возвращает Queries, привязанные к транзакции из ctx, если она есть.
func (s *Store) queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := pg.TxFromContext(ctx); ok {
		return s.q.WithTx(tx)
	}

	return s.q
}
