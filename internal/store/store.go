// Package store owns NextGate's SQLite connection lifecycle and schema. It
// configures connections for concurrency, integrity, and secure deletion, and
// exposes the active handle through [DB].
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// DB is the process-wide SQLite handle installed by [ConnectDB] and cleared by
// [CloseDB]. It is nil while no connection is active.
//
// The handle is safe for concurrent database operations, but callers must
// synchronize calls to [ConnectDB] and [CloseDB] with other access to DB.
var DB *sql.DB

// IsUniqueError reports whether err is caused by a UNIQUE constraint
// violation.
func IsUniqueError(err error) bool {
	var sqliteErr *sqlite.Error

	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}

// ConnectDB opens an SQLite database at dbPath, applies Sekura's connection
// pragmas, verifies that the database responds, and installs the handle as
// [DB].
//
// For regular filesystem paths, ConnectDB creates the database file when
// needed and enforces mode 0600. Paths beginning with ":memory:" or "file:" are
// passed directly to the SQLite driver without file preparation.
//
// The connection enables WAL journaling, foreign-key enforcement, a five-second
// busy timeout, and secure deletion. If any step fails, ConnectDB closes the
// candidate handle and leaves [DB] unchanged.
func ConnectDB(dbPath string) error {
	if !strings.HasPrefix(dbPath, ":memory:") && !strings.HasPrefix(dbPath, "file:") {
		file, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)

		if err != nil {
			return fmt.Errorf("create database file: %w", err)
		}

		file.Close()

		if err := os.Chmod(dbPath, 0600); err != nil {
			return fmt.Errorf("set database permissions: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode and foreign keys for better concurrency and integrity.
	// secure_delete overwrites deleted content so stale ciphertext does not
	// linger in free pages.
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA secure_delete = ON;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return fmt.Errorf("exec pragma %q: %w", pragma, err)
		}
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	DB = db

	return nil
}

// CloseDB closes the active SQLite handle and clears [DB]. It returns nil when
// no handle is active and leaves [DB] set if closing fails.
func CloseDB() error {
	if DB == nil {
		return nil
	}

	if err := DB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	DB = nil

	return nil
}

// CreateTables creates the users, refresh_tokens, and providers tables when
// they do not already exist. Existing tables and rows are left unchanged.
//
// It returns an error when [DB] is nil or the schema cannot be applied.
func CreateTables() error {
	if DB == nil {
		return errors.New("database not connected")
	}

	const schema = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	token TEXT NOT NULL,
	expires_at DATETIME NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	base_url TEXT NOT NULL UNIQUE,
	api_key TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS apikeys (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	api_key TEXT NOT NULL UNIQUE,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

	UNIQUE (title, user_id)
);
`

	if _, err := DB.Exec(schema); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	return nil
}
