package database

import (
	"database/sql"
	"errors"
)

const (
	// create tables query
	tablesQuery = `CREATE TABLE IF NOT EXISTS tasks(
id TEXT PRIMARY KEY,             -- UUIDv4, e.g., "f47ac10b-58cc-4372-a567-0e02b2c3d479"
user_id INT,                    -- Nullable: for server-authenticated users
url TEXT NOT NULL,               -- URL to download/process
state TEXT NOT NULL CHECK (state IN ('pending', 'ongoing', 'complete', 'failed')),
error_message TEXT DEFAULT NULL, -- Optional: error/log message
added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users(
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
email TEXT NOT NULL UNIQUE,
password TEXT NOT NULL,
smtp_to TEXT,
added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

	// trigger to update the updated_at timestamp
	triggerQuery = `CREATE TRIGGER IF NOT EXISTS trg_update_timestamp
AFTER UPDATE ON tasks
FOR EACH ROW
BEGIN
  UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;`
)

var (
	ErrNilDBConn    = errors.New("nil database connection")
	ErrNoRowDeleted = errors.New("no matching row could be found to delete")
	ErrNoRowUpdated = errors.New("no matching row could be found to update")
)

type DB struct {
	Database *sql.DB
}

// New returns a DB struct instance
func New(db *sql.DB) (*DB, error) {
	if db == nil {
		return nil, ErrNilDBConn
	}
	return &DB{
		Database: db,
	}, nil
}

// Setup intiates database setup operation
func (d *DB) Setup() error {
	if d.Database == nil {
		return ErrNilDBConn
	}

	tx, err := d.Database.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(tablesQuery)

	if err != nil {
		return err
	}

	_, err = tx.Exec(triggerQuery)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
