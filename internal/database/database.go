package database

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/jonx8/chat-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db  *gorm.DB
	cfg *config.Config
}

func New(cfg *config.Config) (*Database, error) {
	dsn := getDSN(cfg)

	slog.Info("Opening connection to PostgreSQL")
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

	slog.Info("PostgreSQL connection pool configured",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"database", cfg.DBName,
		"max_connections", cfg.DBMaxOpenConns,
		"idle_connections", cfg.DBMaxIdleConns,
	)

	return database, err
}

func (d *Database) Close() error {
	slog.Info("Closing database connection",
		"host", d.cfg.DBHost,
		"database", d.cfg.DBName,
	)
	if d.db != nil {
		sqlDB, err := d.db.DB()
		if err != nil {
			return fmt.Errorf("get underlying db: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
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
