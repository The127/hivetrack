package database

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations
var migrationFiles embed.FS

// Open connects to PostgreSQL with retries.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Retry connection up to 10 times
	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		if i == maxAttempts-1 {
			return nil, fmt.Errorf("connecting to database after %d attempts: %w", maxAttempts, err)
		}
		time.Sleep(2 * time.Second)
	}

	return db, nil
}

// Migrate applies all pending migrations.
func Migrate(db *sql.DB) error {
	src := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationFiles,
		Root:       "migrations",
	}

	_, err := migrate.Exec(db, "postgres", src, migrate.Up)
	if err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}
