package main

import (
	"go1f/pkg/db"
	"go1f/pkg/server"
	"log"
)

func main() {
	// Инициализация БД
	if err := db.Init("scheduler.db"); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.DB.Close()

	// Запуск сервера
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
