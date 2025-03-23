package main

import (
	"log"

	"github.com/Yandex-Practicum/go_final_project/app"
)

func main() {
	app := app.New()
	log.Fatal(app.Run())
}
