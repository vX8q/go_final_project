package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go1f/pkg/api"
	customdb "go1f/pkg/db"
	date "go1f/pkg/nextdate"
)

const DateLayout = "20060102"

func Run() error {
	dbPath := os.Getenv("TODO_DB")
	if dbPath == "" {
		dbPath = "scheduler.db"
	}

	log.Printf("Run: Инициализация БД с путем: %s", dbPath)
	if err := customdb.Init(dbPath); err != nil {
		log.Printf("Run: Ошибка инициализации БД: %v", err)
		return fmt.Errorf("ошибка инициализации БД: %w", err)
	}
	log.Println("Run: БД инициализирована успешно")

	port := getPort()

	mux := http.NewServeMux()

	setupRoutes(mux, customdb.DB)

	log.Printf("Run: Запуск сервера на порту :%d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Printf("Run: Ошибка при запуске сервера: %v", err)
		return fmt.Errorf("ошибка запуска сервера: %w", err)
	}
	return nil
}

func getPort() int {
	port := 7540
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err != nil {
			log.Printf("getPort: Ошибка преобразования TODO_PORT '%s': %v, используется %d", envPort, err, port)
		} else {
			port = p
			log.Printf("getPort: Используется порт %d из TODO_PORT", port)
		}
	} else {
		log.Printf("getPort: Используется порт по умолчанию %d", port)
	}
	return port
}

func setupRoutes(mux *http.ServeMux, db *sql.DB) {
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	mux.HandleFunc("/api/nextdate", nextDateHandler)

	mux.HandleFunc("/api/tasks", api.TasksHandler(db))
	mux.HandleFunc("/api/task", api.TaskHandler(db))
	mux.HandleFunc("/api/task/done", api.DoneTaskHandler(db))
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	nowStr := q.Get("now")
	dateStr := q.Get("date")
	repeat := q.Get("repeat")

	log.Printf("nextDateHandler: Запрос: Now='%s', Date='%s', Repeat='%s'", nowStr, dateStr, repeat)

	var now time.Time
	var err error
	if nowStr == "" {
		now = time.Now()
		log.Println("nextDateHandler: Now не указан, используется текущее время")
	} else {
		now, err = time.Parse(DateLayout, nowStr)
		if err != nil {
			log.Printf("nextDateHandler: Ошибка Parse для Now: %v", err)
			sendTextError(w, "invalid now parameter", http.StatusBadRequest)
			return
		}
		log.Printf("nextDateHandler: Now: %s", now.Format(DateLayout))
	}

	nextDate, err := date.NextDate(now, dateStr, repeat)
	if err != nil {
		log.Printf("nextDateHandler: Ошибка NextDate: %v", err)
		sendTextError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("nextDateHandler: Результат: %s", nextDate)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

func sendTextError(w http.ResponseWriter, msg string, code int) {
	log.Printf("sendTextError: Error='%s', Code=%d", msg, code)
	http.Error(w, msg, code)
}
