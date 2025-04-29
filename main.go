package main

import (
	"log"
	"net/http"
	"os"

	"go1f/pkg/api"
	"go1f/pkg/db"

	"github.com/gorilla/mux"
)

func main() {

	if err := db.Init("scheduler.db"); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.DB.Close()

	router := mux.NewRouter()

	router.PathPrefix("/web/").Handler(
		http.StripPrefix("/web/", http.FileServer(http.Dir("./web"))),
	)

	api.InitRoutes(router)

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
