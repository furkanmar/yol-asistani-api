package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	// schema_migrations tablosu yoksa oluştur
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Uygulanan migration'ları al
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("query migrations: %w", err)
	}
	applied := map[string]bool{}
	for rows.Next() {
		var v string
		rows.Scan(&v)
		applied[v] = true
	}
	rows.Close()

	// Dosyaları oku ve sırala
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		version := strings.TrimSuffix(f, ".sql")
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}

		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations(version) VALUES($1)`, version,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", f, err)
		}

		fmt.Printf("[migrate] applied: %s\n", f)
	}

	return nil
}
