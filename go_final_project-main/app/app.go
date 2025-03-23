package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	port string
	db   *sql.DB
}

func New() *App {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	dbPath := getDBPath()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	if shouldInstallDB(dbPath) {
		if err := createSchema(db); err != nil {
			log.Fatal(err)
		}
	}

	return &App{
		port: port,
		db:   db,
	}
}

func getDBPath() string {
	if dbFile := os.Getenv("TODO_DBFILE"); dbFile != "" {
		return dbFile
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(filepath.Dir(exePath), "scheduler.db")
}

func shouldInstallDB(dbPath string) bool {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return true
	}
	return false
}

func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT NOT NULL,
			title TEXT NOT NULL,
			comment TEXT,
			repeat TEXT
		);
		
		CREATE INDEX idx_date ON scheduler(date);
	`)
	return err
}

func (a *App) Run() error {
	defer a.db.Close()

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Printf("Server running on :%s\n", a.port)
	return http.ListenAndServe(":"+a.port, nil)
}
