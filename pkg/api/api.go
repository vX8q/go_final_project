package api

import "net/http"

func Init() {
	http.HandleFunc("/api/tasks", tasksHandler)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
}
