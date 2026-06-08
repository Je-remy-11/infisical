package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/infisical/api/internal/config"
	"github.com/infisical/api/internal/libs/logutil"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	migrationsDir   = "./migrations"
	migrationsTable = "schema_migrations"
	advisoryLockKey = "infisical-migrate-up"
)

var migrationFileRe = regexp.MustCompile(`^(\d{14})_.+\.up\.sql$`)

type migration struct {
	version  string
	filename string
	content  string
}

func main() {
	logger := slog.New(logutil.NewContextHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetConfiguredSlogLevel(),
	})))
	slog.SetDefault(logger)

	if len(os.Args) < 2 || os.Args[1] != "up" {
		logger.ErrorContext(context.Background(), "usage: migrate up")
		os.Exit(1)
	}

	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.ErrorContext(ctx, "failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	if cfg.DBConnectionURI == "" {
		logger.ErrorContext(ctx, "DB_CONNECTION_URI is required")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", cfg.DBConnectionURI)
	if err != nil {
		logger.ErrorContext(ctx, "failed to open database connection", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to ping database", slog.Any("error", err))
		os.Exit(1)
	}
	logger.InfoContext(ctx, "database connection established")

	if err := runMigrations(ctx, db, logger); err != nil {
		logger.ErrorContext(ctx, "migration failed", slog.Any("error", err))
		os.Exit(1)
	}

	logger.InfoContext(ctx, "all migrations applied successfully")
}

func runMigrations(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("ensuring migrations table: %w", err)
	}
	logger.InfoContext(ctx, "ensured schema_migrations table exists")

	migrations, err := readMigrationFiles()
	if err != nil {
		return fmt.Errorf("reading migration files: %w", err)
	}
	if len(migrations) == 0 {
		logger.InfoContext(ctx, "no migration files found", slog.String("dir", migrationsDir))
		return nil
	}
	logger.InfoContext(ctx, "found migration files", slog.Int("count", len(migrations)))

	applied, err := getAppliedVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("getting applied versions: %w", err)
	}

	var pending []migration
	for _, m := range migrations {
		if !applied[m.version] {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		logger.InfoContext(ctx, "no pending migrations")
		return nil
	}
	logger.InfoContext(ctx, "pending migrations", slog.Int("count", len(pending)))

	lockTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning lock transaction: %w", err)
	}
	defer lockTx.Rollback()

	lockID := stringToAdvisoryLockID(advisoryLockKey)
	if _, err := lockTx.ExecContext(ctx, "SELECT pg_advisory_xact_lock($1)", lockID); err != nil {
		return fmt.Errorf("acquiring migration advisory lock: %w", err)
	}
	logger.InfoContext(ctx, "acquired migration advisory lock")

	appliedAfterLock, err := getAppliedVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("re-checking applied versions after lock: %w", err)
	}

	for _, m := range pending {
		if appliedAfterLock[m.version] {
			logger.InfoContext(ctx, "skipping already-applied migration (applied by another process)",
				slog.String("version", m.version), slog.String("file", m.filename))
			continue
		}
		if err := applyMigration(ctx, db, m, logger); err != nil {
			return fmt.Errorf("applying migration %s: %w", m.filename, err)
		}
	}

	if err := lockTx.Commit(); err != nil {
		return fmt.Errorf("committing lock transaction: %w", err)
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version    VARCHAR(14) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`, migrationsTable))
	return err
}

func readMigrationFiles() ([]migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading migrations directory: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFileRe.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, migration{
			version:  matches[1],
			filename: entry.Name(),
			content:  string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func getAppliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT version FROM %s", migrationsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

func applyMigration(ctx context.Context, db *sql.DB, m migration, logger *slog.Logger) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	logger.InfoContext(ctx, "applying migration",
		slog.String("version", m.version), slog.String("file", m.filename))

	if _, err := tx.ExecContext(ctx, m.content); err != nil {
		return fmt.Errorf("executing SQL: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf("INSERT INTO %s (version) VALUES ($1)", migrationsTable),
		m.version,
	); err != nil {
		return fmt.Errorf("recording migration version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	logger.InfoContext(ctx, "migration applied",
		slog.String("version", m.version), slog.String("file", m.filename))
	return nil
}

func stringToAdvisoryLockID(s string) int64 {
	h := sha256.Sum256([]byte(s))
	return int64(binary.BigEndian.Uint64(h[:8]))
}
