package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applies every migration file that has not run yet, in numeric order, and records
// each one in schema_migrations. A MySQL advisory lock keeps two instances from migrating at once.
// Note: MySQL DDL is not transactional, so a migration that fails half-way can leave partial
// changes behind. Keep files small, and fix a failed one by hand before restarting.
func Migrate(dsn string) error {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	db, err := sql.Open("mysql", dsn+sep+"multiStatements=true")
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	conn, err := db.Conn(ctx) // one pinned connection: the advisory lock belongs to it
	if err != nil {
		return fmt.Errorf("cannot connect to the database: %w", err)
	}
	defer conn.Close()

	var locked int
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK('adaptiq_migrations', 60)").Scan(&locked); err != nil || locked != 1 {
		return fmt.Errorf("could not acquire the migration lock")
	}
	defer conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK('adaptiq_migrations')")

	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INT PRIMARY KEY, name VARCHAR(100) NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP)`); err != nil {
		return err
	}

	applied := map[int]bool{}
	rows, err := conn.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return err
	}
	for rows.Next() {
    var v int
    if err := rows.Scan(&v); err != nil {
        rows.Close()
        return err
    }
    applied[v] = true
}

if err := rows.Err(); err != nil {
    rows.Close()
    return err
}

if err := rows.Close(); err != nil {
    return err
}

	files, err := migrationFS.ReadDir("migrations") // returned sorted by file name
	if err != nil {
		return err
	}
	for _, f := range files {
		version, err := strconv.Atoi(strings.SplitN(f.Name(), "_", 2)[0])
		if err != nil {
			return fmt.Errorf("migration %q must start with a number", f.Name())
		}
		if applied[version] {
			continue
		}
		body, err := migrationFS.ReadFile("migrations/" + f.Name())
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s failed: %w", f.Name(), err)
		}
		if _, err := conn.ExecContext(ctx, "INSERT INTO schema_migrations (version, name) VALUES (?, ?)", version, f.Name()); err != nil {
			return err
		}
		log.Printf("migration applied: %s", f.Name())
	}
	return nil
}