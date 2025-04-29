package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go1f/pkg/db"

	"github.com/gorilla/mux"
)

const DateLayout = "20060102"

func InitRoutes(r *mux.Router) {
	r.HandleFunc("/api/nextdate", NextDateHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/tasks", TasksHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/task", TaskByIDHandler).Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	r.HandleFunc("/api/task/done", TaskDoneHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/tasks", TasksHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/tasks/clear", ClearTasksHandler).Methods(http.MethodPost)
}

type TaskRequest struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TaskResponse struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TasksResponse struct {
	Tasks []*TaskResponse `json:"tasks"`
}

func TaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, "ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		t, err := db.GetTask(id)
		if err != nil {
			writeError(w, "Task not found", http.StatusNotFound)
			return
		}
		writeJSON(w, TaskResponse{
			ID:      t.ID,
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		})

	case http.MethodPut:
		var req TaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		idInt, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			writeError(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		task := &db.Task{
			ID:      idInt,
			Date:    req.Date,
			Title:   req.Title,
			Comment: req.Comment,
			Repeat:  req.Repeat,
		}

		if err := db.UpdateTask(task); err != nil {
			writeError(w, "Update failed", http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]interface{}{})

	case http.MethodDelete:
		idInt, _ := strconv.ParseInt(id, 10, 64)
		if err := db.DeleteTask(idInt); err != nil {
			writeError(w, "Delete failed", http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]interface{}{})
	}
}

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, "ID required", http.StatusBadRequest)
		return
	}

	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, "Task not found", http.StatusNotFound)
		return
	}

	if t.Repeat == "" {
		if err := db.DeleteTask(t.ID); err != nil {
			writeError(w, "Delete error", http.StatusInternalServerError)
			return
		}
	} else {
		next, err := NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			writeError(w, "Invalid repeat rule", http.StatusBadRequest)
			return
		}
		t.Date = next
		if err := db.UpdateTask(t); err != nil {
			writeError(w, "Update error", http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, map[string]interface{}{})
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)
	tasks, err := db.Tasks()
	if err != nil {
		writeError(w, "Database error", http.StatusInternalServerError)
		return
	}

	search := strings.ToLower(r.URL.Query().Get("search"))
	resp := TasksResponse{}

	for _, t := range tasks {
		tr := &TaskResponse{
			ID:      t.ID,
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		}

		if search == "" || containsTask(tr, search) {
			resp.Tasks = append(resp.Tasks, tr)
		}
	}
	writeJSON(w, resp)
}

func ClearTasksHandler(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)
	if err := db.ClearTasks(); err != nil {
		writeError(w, "Clear failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{})
}

func recoverPanic(w http.ResponseWriter) {
	if rec := recover(); rec != nil {
		writeError(w, "Internal error", http.StatusInternalServerError)
	}
}

func containsTask(t *TaskResponse, s string) bool {
	search := strings.ToLower(s)
	return strings.Contains(strings.ToLower(t.Title), search) ||
		strings.Contains(strings.ToLower(t.Comment), search) ||
		strings.Contains(strings.ToLower(t.Date), search)
}
