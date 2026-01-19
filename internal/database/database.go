package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/jonx8/chat-service/internal/config"
	"github.com/jonx8/chat-service/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db  *gorm.DB
	cfg *config.Config
}

func New(ctx context.Context, cfg *config.Config) (*Database, error) {
	dsn := getDSN(cfg)

	slog.Info("Opening connection to PostgreSQL...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:          true,
		DisableAutomaticPing: false,
	})

	if err != nil {
		slog.Error("Failed to connect to PostgreSQL",
			"host", cfg.DBHost,
			"error", err,
		)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("PostgreSQL connection established",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"database", cfg.DBName,
	)

	database := &Database{
		db:  db,
		cfg: cfg,
	}

	if err := database.setupPool(); err != nil {
		return nil, fmt.Errorf("setup pool: %w", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}

	if err := migrations.RunMigrations(ctx, sqlDB); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	slog.Info("PostgreSQL connection pool configured",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"database", cfg.DBName,
		"max_connections", cfg.DBMaxOpenConns,
		"idle_connections", cfg.DBMaxIdleConns,
	)

	return database, err
}

func (d *Database) DB() (*sql.DB, error) {
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return nil, fmt.Errorf("get underlying db: %w", err)
		}
		return sqlDB, nil
	}
	return nil, fmt.Errorf("gorm DB is nil")
}

func (d *Database) Gorm() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.DB()
	if err != nil {
		return fmt.Errorf("can't close connection: %w", err)
	}
	return sqlDB.Close()
}

func (d *Database) setupPool() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("get underlying db: %w", err)
	}

	sqlDB.SetMaxIdleConns(d.cfg.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(d.cfg.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(d.cfg.DBConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(d.cfg.DBConnMaxIdleTime) * time.Second)

	return nil
}

func getDSN(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		url.QueryEscape(cfg.DBPassword),
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)
}
