package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX idx_date ON scheduler(date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	dsn := fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL", dbFile)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	if install {
		if _, err := db.ExecContext(ctx, schema); err != nil {
			return fmt.Errorf("failed to create schema: %w", err)
		}
	}

	const checkQuery = `SELECT name FROM sqlite_master WHERE type='table' AND name='scheduler'`
	var tableName string
	if err := db.QueryRowContext(ctx, checkQuery).Scan(&tableName); err != nil {
		return fmt.Errorf("table check failed: %w", err)
	}

	if !strings.EqualFold(tableName, "scheduler") {
		return fmt.Errorf("required table 'scheduler' not found")
	}

	DB = db
	return nil
}
