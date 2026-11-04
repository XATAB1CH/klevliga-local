package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg map[string]string) (*pgxpool.Pool, error) {
	pgCfg, err := pgxpool.ParseConfig(connString(cfg))
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func connString(cfg map[string]string) string {
	var sb strings.Builder

	for key, value := range cfg {
		if key == "" || value == "" {
			continue
		}

		value = strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(value)
		fmt.Fprintf(&sb, "%s='%s' ", key, value)
	}

	return strings.TrimSpace(sb.String())
}
