//go:build integration

package postgres_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/the127/hivetrack/internal/database"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("HIVETRACK_DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "HIVETRACK_DATABASE_URL not set — skipping integration tests")
		os.Exit(0)
	}

	db, err := database.Open(dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	testDB = db
	os.Exit(m.Run())
}
