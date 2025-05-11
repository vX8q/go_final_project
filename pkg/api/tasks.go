package api

import (
	"database/sql"
	"go1f/pkg/db"
	"log"
	"net/http"
)

type TasksResp struct {
	Tasks []db.TaskForAPI `json:"tasks"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

func TasksHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		searchQuery := r.URL.Query().Get("search")

		var tasks []*db.Task
		var err error

		if searchQuery != "" {
			log.Printf("Запрос /api/tasks с поиском: '%s'", searchQuery)
			tasks, err = db.SearchTasks(dbConn, searchQuery, 50)
		} else {
			log.Println("Запрос /api/tasks без поиска")
			tasks, err = db.FetchTasks(dbConn, 50)
		}

		if err != nil {
			log.Printf("Ошибка при получении задач: %v", err)

			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if tasks == nil {
			tasks = []*db.Task{}
		}

		tasksForAPI := make([]db.TaskForAPI, len(tasks))
		for i, task := range tasks {
			tasksForAPI[i] = task.ToTaskForAPI()
		}

		log.Printf("Успешный ответ /api/tasks, количество задач: %d", len(tasksForAPI))

		writeJson(w, TasksResp{Tasks: tasksForAPI}, http.StatusOK)
	}
}
