
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/infisical/api/internal/config"
	"github.com/infisical/api/internal/database/pg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	migrationsDir = "./migrations"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.GetConfiguredSlogLevel(),
	}))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		var validationErr *config.ValidationError
		if errors.As(err, &validationErr) {
			logger.ErrorContext(context.Background(), "invalid environment variables")
			for _, issue := range validationErr.Issues {
				logger.ErrorContext(context.Background(), "  "+issue)
			}
		} else {
			logger.ErrorContext(context.Background(), "failed to load config", slog.Any("error", err))
		}
		os.Exit(1)
	}

	if len(os.Args) &lt; 2 {
		logger.ErrorContext(context.Background(), "invalid usage: expected 'migrate up'")
		os.Exit(1)
	}

	command := os.Args[1]
	if command != "up" {
		logger.ErrorContext(context.Background(), "invalid command: only 'up' is supported")
		os.Exit(1)
	}

	ctx := context.Background()

	db, err := pg.NewPostgresDB(ctx, cfg.DBConnectionURI, cfg.DBRootCert, cfg.DBReadReplicas)
	if err != nil {
		logger.ErrorContext(ctx, "failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	migrator := NewMigrator(db.Primary(), migrationsDir, logger)

	if err := migrator.MigrateUp(ctx); err != nil {
		logger.ErrorContext(ctx, "migration failed", slog.Any("error", err))
		os.Exit(1)
	}

	logger.InfoContext(ctx, "migrations completed successfully")
}

type Migrator struct {
	db            *pgxpool.Pool
	migrationsDir string
	logger        *slog.Logger
}

func NewMigrator(db *pgxpool.Pool, migrationsDir string, logger *slog.Logger) *Migrator {
	return &amp;Migrator{
		db:            db,
		migrationsDir: migrationsDir,
		logger:        logger,
	}
}

type Migration struct {
	Name    string
	Path    string
	Content string
}

func (m *Migrator) MigrateUp(ctx context.Context) error {
	if err := m.ensureSchemaMigrationsTable(ctx); err != nil {
		return err
	}

	migrations, err := m.loadMigrations()
	if err != nil {
		return err
	}

	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return err
	}

	appliedMap := make(map[string]bool)
	for _, name := range appliedMigrations {
		appliedMap[name] = true
	}

	for _, migration := range migrations {
		if appliedMap[migration.Name] {
			m.logger.InfoContext(ctx, "migration already applied", slog.String("name", migration.Name))
			continue
		}

		m.logger.InfoContext(ctx, "applying migration", slog.String("name", migration.Name))
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
		}
		m.logger.InfoContext(ctx, "migration applied successfully", slog.String("name", migration.Name))
	}

	return nil
}

func (m *Migrator) ensureSchemaMigrationsTable(ctx context.Context) error {
	_, err := m.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) loadMigrations() ([]*Migration, error) {
	entries, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Migration{}, nil
		}
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []*Migration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(m.migrationsDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, &amp;Migration{
			Name:    entry.Name(),
			Path:    path,
			Content: string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name &lt; migrations[j].Name
	})

	return migrations, nil
}

func (m *Migrator) getAppliedMigrations(ctx context.Context) ([]string, error) {
	rows, err := m.db.Query(ctx, "SELECT name FROM schema_migrations ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&amp;name); err != nil {
			return nil, fmt.Errorf("failed to scan migration name: %w", err)
		}
		migrations = append(migrations, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migrations: %w", err)
	}

	return migrations, nil
}

func (m *Migrator) applyMigration(ctx context.Context, migration *Migration) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				m.logger.ErrorContext(ctx, "failed to rollback transaction", slog.Any("error", rollbackErr))
			}
		}
	}()

	statements := splitSQLStatements(migration.Content)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute SQL statement: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", migration.Name); err != nil {
		return fmt.Errorf("failed to record migration in schema_migrations: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func splitSQLStatements(sql string) []string {
	var statements []string
	var currentStatement strings.Builder
	var inSingleQuote, inDoubleQuote bool

	for _, char := range sql {
		switch char {
		case '\'':
			inSingleQuote = !inSingleQuote
			currentStatement.WriteRune(char)
		case '"':
			inDoubleQuote = !inDoubleQuote
			currentStatement.WriteRune(char)
		case ';':
			if !inSingleQuote &amp;&amp; !inDoubleQuote {
				statements = append(statements, currentStatement.String())
				currentStatement.Reset()
			} else {
				currentStatement.WriteRune(char)
			}
		default:
			currentStatement.WriteRune(char)
		}
	}

	if currentStatement.Len() &gt; 0 {
		statements = append(statements, currentStatement.String())
	}

	return statements
}

