// admin-bootstrap grants the first administrator to an existing verified user.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/XATAB1CH/klev-liga/internal/config"
	"github.com/XATAB1CH/klev-liga/internal/services/access"
	"github.com/XATAB1CH/klev-liga/internal/storages/postgres/repo"
	"github.com/XATAB1CH/klev-liga/migrations"
	"github.com/XATAB1CH/klev-liga/pkg/pg"
)

func run() error {
	configPath := flag.String("c", "./config.yaml", "path to API config")
	userID := flag.String("user-id", "", "UUID of an existing verified active user (or use -username)")
	username := flag.String("username", "", "username of an existing verified active user (or use -user-id)")
	reason := flag.String("reason", "Initial administrator setup", "audit reason, 3–1000 characters")
	flag.Parse()
	if (*userID == "") == (strings.TrimSpace(*username) == "") {
		return fmt.Errorf("provide exactly one of -username or -user-id for an existing verified account")
	}
	var id uuid.UUID
	var err error
	if *userID != "" {
		id, err = uuid.Parse(*userID)
		if err != nil || id == uuid.Nil {
			return fmt.Errorf("invalid -user-id UUID")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := config.New(ctx, *configPath)
	if err != nil {
		return err
	}
	db, err := pg.Connect(ctx, cfg.Connections.Postgres.Main)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := pg.MigrateFromPool(ctx, db, migrations.Migrations); err != nil {
		return err
	}
	store := repo.NewStore(db)
	if *username != "" {
		id, err = store.GetUserIDByUsername(ctx, strings.TrimSpace(*username))
		if err != nil {
			return err
		}
	}
	if err := access.NewService(store).BootstrapAdmin(ctx, id, *reason, "bootstrap:"+uuid.NewString()); err != nil {
		return fmt.Errorf("bootstrap refused (an active administrator must not already exist): %w", err)
	}
	fmt.Printf("Administrator assigned to %s; audit event saved.\n", id)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
