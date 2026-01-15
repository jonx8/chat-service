package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
)

type gooseLogger struct{}

func (l *gooseLogger) Printf(format string, args ...interface{}) {
	slog.Info(fmt.Sprintf(format, args...))
}

func (l *gooseLogger) Fatalf(format string, args ...interface{}) {
	slog.Error(fmt.Sprintf(format, args...))
}

//go:embed *.sql
var migrationsDir embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	slog.Info("Checking database migrations")

	goose.SetBaseFS(migrationsDir)
	goose.SetLogger(&gooseLogger{})

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	currentVersion, err := goose.GetDBVersion(db)
	if err != nil {
		slog.Warn("cannot get current migration version", "error", err)
		currentVersion = 0
	}

	migrationFiles, err := migrationsDir.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read migration files: %w", err)
	}

	migrationsCount := len(migrationFiles)
	slog.Info("Migration status",
		"current_version", currentVersion,
		"available_migrations", migrationsCount,
	)

	if int64(migrationsCount) <= currentVersion {
		slog.Info("Database is up to date, no migrations needed")
		return nil
	}

	slog.Info("Applying database migrations")

	if err := goose.UpContext(ctx, db, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	newVersion, _ := goose.GetDBVersion(db)
	slog.Info("Migrations applied successfully",
		"previous_version", currentVersion,
		"new_version", newVersion,
		"applied_migrations", newVersion-currentVersion,
	)

	return nil
}
