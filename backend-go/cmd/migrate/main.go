package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/infisical/api/internal/libs/logutil"
)

func main() {
	logger := slog.New(logutil.NewContextHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		logger.ErrorContext(context.Background(), "usage: migrate <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		if err := runUp(logger); err != nil {
			logger.ErrorContext(context.Background(), "migration failed", slog.Any("error", err))
			os.Exit(1)
		}
	default:
		logger.ErrorContext(context.Background(), "unknown command", slog.String("command", os.Args[1]))
		os.Exit(1)
	}
}

func runUp(logger *slog.Logger) error {
	ctx := context.Background()

	connURI := os.Getenv("DB_CONNECTION_URI")
	if connURI == "" {
		return errors.New("DB_CONNECTION_URI is required")
	}

	db, err := openSQLDB(connURI, os.Getenv("DB_ROOT_CERT"))
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	logger.InfoContext(ctx, "connected to database")

	if err := ensureSchemaMigrationsTable(ctx, db, logger); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	migrations, err := discoverMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("discovering migrations: %w", err)
	}
	logger.InfoContext(ctx, "discovered migration files", slog.Int("count", len(migrations)))

	if len(migrations) == 0 {
		logger.InfoContext(ctx, "no migration files found")
		return nil
	}

	applied, err := getAppliedVersions(ctx, db)
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}
	logger.DebugContext(ctx, "previously applied migrations", slog.Int("count", len(applied)))

	for _, m := range migrations {
		if applied[m.Version] {
			logger.DebugContext(ctx, "migration already applied, skipping",
				slog.String("version", m.Version),
				slog.String("name", m.Name),
			)
			continue
		}

		if err := applyMigration(ctx, db, m, logger); err != nil {
			return fmt.Errorf("applying migration %s (%s): %w", m.Version, m.Name, err)
		}
	}

	logger.InfoContext(ctx, "all migrations applied successfully")
	return nil
}

type Migration struct {
	Version string
	Name    string
	SQL     string
}

func openSQLDB(connURI, rootCert string) (*sql.DB, error) {
	if rootCert == "" {
		db, err := sql.Open("pgx", connURI)
		if err != nil {
			return nil, fmt.Errorf("opening database: %w", err)
		}
		configurePool(db)
		return db, nil
	}

	connConfig, err := pgx.ParseConfig(connURI)
	if err != nil {
		return nil, fmt.Errorf("parsing connection URI: %w", err)
	}

	caPEM, err := base64.StdEncoding.DecodeString(rootCert)
	if err != nil {
		return nil, fmt.Errorf("decoding DB_ROOT_CERT base64: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse DB root certificate PEM")
	}

	tlsCfg := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true,
	}

	parsed, err := url.Parse(connURI)
	if err == nil {
		sslMode := parsed.Query().Get("sslmode")
		if strings.EqualFold(sslMode, "verify-ca") || strings.EqualFold(sslMode, "verify-full") {
			tlsCfg.InsecureSkipVerify = false
		}
	}

	connConfig.TLSConfig = tlsCfg

	db := stdlib.OpenDB(*connConfig)
	configurePool(db)
	return db, nil
}

func configurePool(db *sql.DB) {
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(10 * time.Minute)
}

func ensureSchemaMigrationsTable(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return err
	}

	logger.DebugContext(ctx, "ensured schema_migrations table exists")
	return nil
}

func discoverMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading migrations directory %q: %w", dir, err)
	}

	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		base := strings.TrimSuffix(name, ".up.sql")
		parts := strings.SplitN(base, "_", 2)

		version := parts[0]
		desc := ""
		if len(parts) > 1 {
			desc = parts[1]
		}

		if version == "" {
			continue
		}

		path := filepath.Join(dir, name)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading migration file %q: %w", path, err)
		}

		sqlContent := strings.TrimSpace(string(data))
		if sqlContent == "" {
			continue
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    desc,
			SQL:     sqlContent,
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version")
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applied, nil
}

func applyMigration(ctx context.Context, db *sql.DB, m Migration, logger *slog.Logger) error {
	logger.InfoContext(ctx, "applying migration",
		slog.String("version", m.Version),
		slog.String("name", m.Name),
		slog.Int("sql_size", len(m.SQL)),
	)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("executing migration SQL: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)`,
		m.Version, m.Name, time.Now().UTC(),
	); err != nil {
		return fmt.Errorf("recording migration version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	tx = nil

	logger.InfoContext(ctx, "migration applied",
		slog.String("version", m.Version),
		slog.String("name", m.Name),
	)

	return nil
}