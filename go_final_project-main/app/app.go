package app

import (
	"log"
	"net/http"
	"os"
)

type App struct {
	port string
}

func New() *App {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}
	return &App{port: port}
}

func (a *App) Run() error {
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	log.Printf("Сервер запущен на порту :%s\n", a.port)
	return http.ListenAndServe(":"+a.port, nil)
}
