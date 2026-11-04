package pg

import (
	"context"
	"database/sql"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, db *sql.DB, fs embed.FS, opts ...goose.ProviderOption) error {
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fs, opts...)
	if err != nil {
		return err
	}
	defer provider.Close()

	if _, err := provider.Up(ctx); err != nil {
		return err
	}

	return nil
}

func MigrateFromPool(ctx context.Context, pool *pgxpool.Pool, fs embed.FS, opts ...goose.ProviderOption) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	return Migrate(ctx, db, fs, opts...)
}
