package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go1f/pkg/db"
)

func init() {
	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			addTaskHandler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

func validateRepeat(rule string) bool {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return true
	}
	parts := strings.Fields(rule)
	if len(parts) != 2 || parts[0] != "d" {
		return false
	}
	_, err := strconv.Atoi(parts[1])
	return err == nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t db.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(t.Title) == "" {
		writeError(w, "title required", http.StatusBadRequest)
		return
	}

	if t.Date == "" || strings.EqualFold(t.Date, "today") {
		t.Date = time.Now().Format(DateLayout)
	} else {
		if _, err := time.Parse(DateLayout, t.Date); err != nil {
			writeError(w, "invalid date format (use YYYYMMDD)", http.StatusBadRequest)
			return
		}
	}

	if !validateRepeat(t.Repeat) {
		writeError(w, "invalid repeat format", http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&t)
	if err != nil {
		writeError(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"id": id})
}
