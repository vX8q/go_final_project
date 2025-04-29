package handlers

import (
	"encoding/json"
	"go1f/pkg/db"
	"net/http"
)

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	if task.Date == "" || task.Title == "" {
		respondWithError(w, http.StatusBadRequest, "Date and title are required")
		return
	}

	res, err := db.DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	id, _ := res.LastInsertId()
	respondWithJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
