package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/black-coda/go-boilerplate/internal/config"

	"github.com/jackc/pgx/v5"
	tern "github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, logger *zerolog.Logger, cfg *config.Config) error {

	// URL-encode the password

	dsn := DsnConfig(cfg)

	conn, err := pgx.Connect(ctx, dsn.String())
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	migrator, err := tern.NewMigrator(ctx, conn, "schema_version")
	if err != nil {
		return fmt.Errorf("constructing database migrator: %w", err)
	}
	subtree, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("retrieving database migrations subtree: %w", err)
	}
	if err := migrator.LoadMigrations(subtree); err != nil {
		return fmt.Errorf("loading database migrations: %w", err)
	}
	version, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("retrieving current database migration version")
	}
	if err := migrator.Migrate(ctx); err != nil {
		return err
	}
	if version == int32(len(migrator.Migrations)) {
		logger.Info().Msgf("database schema up to date, version %d", len(migrator.Migrations))
	} else {
		logger.Info().Msgf("migrated database schema, from %d to %d", version, len(migrator.Migrations))
	}
	return nil
}
