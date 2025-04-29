package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR NOT NULL,
    comment TEXT,
    repeat VARCHAR(128)
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

func Init(dbFile string) error {
	var err error
	DB, err = sqlx.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	DB.MustExec(schema)
	return nil
}
