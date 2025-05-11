package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go1f/pkg/db"
	date "go1f/pkg/nextdate"
)

const DateLayout = "20060102"

type FlexibleDate string

func (fd *FlexibleDate) UnmarshalJSON(data []byte) error {

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fd = FlexibleDate(s)
		return nil
	}

	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*fd = FlexibleDate(strconv.FormatInt(n, 10))
		return nil
	}

	var f float64
	if err := json.Unmarshal(data, &f); err == nil {
		*fd = FlexibleDate(strconv.FormatInt(int64(f), 10))
		return nil
	}

	return fmt.Errorf("поле 'date' должно быть строкой или числом, представляющим дату, получено: %s", string(data))
}

type addTaskRequest struct {
	Date    FlexibleDate `json:"date"`
	Title   string       `json:"title"`
	Comment string       `json:"comment"`
	Repeat  string       `json:"repeat"`
}

type updateTaskRequest struct {
	ID      string       `json:"id"`
	Date    FlexibleDate `json:"date"`
	Title   string       `json:"title"`
	Comment string       `json:"comment"`
	Repeat  string       `json:"repeat"`
}

type addTaskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func writeError(w http.ResponseWriter, msg string, code int) {
	resp := addTaskResponse{Error: msg}
	loggableData, _ := json.Marshal(resp)
	log.Printf("writeError: Status=%d, Body=%s", code, string(loggableData))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func writeJson(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка формирования JSON ответа: %v, data: %+v", err, data)
		http.Error(w, `{"error":"Ошибка формирования JSON ответа"}`, http.StatusInternalServerError)
	}
}

func TaskHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("TaskHandler: Запрос %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			addTaskHandler(dbConn, w, r)
		case http.MethodGet:
			getTaskHandler(dbConn, w, r)
		case http.MethodPut:
			updateTaskHandler(dbConn, w, r)
		case http.MethodDelete:
			deleteTaskHandler(dbConn, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func addTaskHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req addTaskRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("addTaskHandler: Ошибка декодирования JSON: %v", err)
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("addTaskHandler: Запрос: %+v", req)

	if strings.TrimSpace(req.Title) == "" {
		log.Println("addTaskHandler: Пустой заголовок")
		writeError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	dateStr, err := normalizeDate(string(req.Date), req.Repeat)
	if err != nil {
		log.Printf("addTaskHandler: Ошибка normalizeDate: %v", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := &db.Task{
		Date:    dateStr,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	id, err := db.AddTask(dbConn, task)
	if err != nil {
		log.Printf("addTaskHandler: Ошибка AddTask: %v", err)
		writeError(w, "db insert error", http.StatusInternalServerError)
		return
	}

	resp := addTaskResponse{ID: id}
	log.Printf("addTaskHandler: Успех, ID: %d", id)
	json.NewEncoder(w).Encode(resp)
}

func getTaskHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		log.Println("getTaskHandler: Не указан идентификатор задачи")
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(dbConn, idStr)
	if err != nil {
		if err.Error() == "задача не найдена" {
			log.Printf("getTaskHandler: Задача с ID %s не найдена", idStr)
			writeError(w, "Задача не найдена", http.StatusNotFound)
		} else if err.Error() == "некорректный идентификатор задачи" {
			log.Printf("getTaskHandler: Некорректный идентификатор задачи '%s': %v", idStr, err)
			writeError(w, "Некорректный идентификатор", http.StatusBadRequest)
		} else {
			log.Printf("getTaskHandler: Ошибка получения задачи из БД: %v", err)
			writeError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	taskForAPI := task.ToTaskForAPI()
	log.Printf("getTaskHandler: Успешно получена задача для API с ID: %s", taskForAPI.ID)
	writeJson(w, taskForAPI, http.StatusOK)
}

func updateTaskHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req updateTaskRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("updateTaskHandler: Ошибка декодирования JSON: %v", err)
		writeError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("updateTaskHandler: Запрос: %+v", req)

	if strings.TrimSpace(req.ID) == "" {
		log.Println("updateTaskHandler: Не указан идентификатор задачи для обновления")
		writeError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	idInt64, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		log.Printf("updateTaskHandler: Некорректный формат ID задачи '%s': %v", req.ID, err)
		writeError(w, "Некорректный идентификатор задачи", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Title) == "" {
		log.Println("updateTaskHandler: Пустой заголовок задачи")
		writeError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	dateStr, err := normalizeDate(string(req.Date), req.Repeat)
	if err != nil {
		log.Printf("updateTaskHandler: Ошибка normalizeDate: %v", err)
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := &db.Task{
		ID:      idInt64,
		Date:    dateStr,
		Title:   req.Title,
		Comment: req.Comment,
		Repeat:  req.Repeat,
	}

	err = db.UpdateTask(dbConn, task)
	if err != nil {
		if err.Error() == "incorrect id for updating task" {
			log.Printf("updateTaskHandler: Задача с ID %d не найдена для обновления", task.ID)
			writeError(w, "Задача не найдена", http.StatusNotFound)
		} else {
			log.Printf("updateTaskHandler: Ошибка обновления задачи в БД: %v", err)
			writeError(w, "Ошибка обновления задачи", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("updateTaskHandler: Успешно обновлена задача с ID: %d", task.ID)
	writeJson(w, struct{}{}, http.StatusOK)
}

func deleteTaskHandler(dbConn *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		log.Println("deleteTaskHandler: Не указан идентификатор задачи для удаления")
		writeError(w, "Не указан идентификатор задачи", http.StatusBadRequest)
		return
	}

	err := db.DeleteTask(dbConn, idStr)
	if err != nil {
		if err.Error() == "задача не найдена" {
			log.Printf("deleteTaskHandler: Задача с ID %s не найдена для удаления", idStr)
			writeError(w, "Задача не найдена", http.StatusNotFound)
		} else if err.Error() == "некорректный идентификатор задачи" {
			log.Printf("deleteTaskHandler: Некорректный идентификатор задачи '%s': %v", idStr, err)
			writeError(w, "Некорректный идентификатор", http.StatusBadRequest)
		} else {
			log.Printf("deleteTaskHandler: Ошибка удаления задачи из БД: %v", err)
			writeError(w, "Ошибка удаления задачи", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("deleteTaskHandler: Успешно удалена задача с ID: %s", idStr)
	writeJson(w, struct{}{}, http.StatusOK)
}

func DoneTaskHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			log.Println("DoneTaskHandler: Не указан идентификатор задачи")
			writeError(w, "Не указан идентификатор", http.StatusBadRequest)
			return
		}

		task, err := db.GetTask(dbConn, idStr)
		if err != nil {
			if err.Error() == "задача не найдена" {
				log.Printf("DoneTaskHandler: Задача с ID %s не найдена", idStr)
				writeError(w, "Задача не найдена", http.StatusNotFound)
			} else if err.Error() == "некорректный идентификатор задачи" {
				log.Printf("DoneTaskHandler: Некорректный идентификатор задачи '%s': %v", idStr, err)
				writeError(w, "Некорректный идентификатор", http.StatusBadRequest)
			} else {
				log.Printf("DoneTaskHandler: Ошибка получения задачи из БД: %v", err)
				writeError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if task.Repeat == "" {

			err = db.DeleteTask(dbConn, idStr)
			if err != nil {

				log.Printf("DoneTaskHandler: Ошибка при удалении одноразовой задачи с ID %s: %v", idStr, err)

				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("DoneTaskHandler: Успешно удалена одноразовая задача с ID: %s", idStr)
		} else {

			now := time.Now()
			next, err := date.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				log.Printf("DoneTaskHandler: Ошибка расчета следующей даты для задачи с ID %s (repeat: '%s'): %v", idStr, task.Repeat, err)

				writeError(w, fmt.Sprintf("Ошибка расчета следующей даты: %v", err), http.StatusInternalServerError)
				return
			}

			task.Date = next
			err = db.UpdateTask(dbConn, task)
			if err != nil {

				log.Printf("DoneTaskHandler: Ошибка при обновлении даты задачи с ID %s: %v", idStr, err)
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("DoneTaskHandler: Успешно обновлена дата задачи с ID: %s на %s", idStr, next)
		}

		writeJson(w, struct{}{}, http.StatusOK)
	}
}

func normalizeDate(dateStr, repeat string) (string, error) {
	const layout = "20060102"
	now := time.Now()
	today := truncDay(now)

	log.Printf("normalizeDate: DateStr: '%s', Repeat: '%s'", dateStr, repeat)

	if dateStr == "" {
		result := today.Format(layout)
		log.Printf("normalizeDate: DateStr пустой, Result: '%s'", result)
		return result, nil
	}

	t, err := time.Parse(layout, dateStr)
	if err != nil {
		log.Printf("normalizeDate: Ошибка Parse: %v", err)
		return "", errors.New("invalid date format")
	}

	if t.Before(today) {

		if repeat != "" {

			next, err := date.NextDate(now, dateStr, repeat)
			if err != nil {
				log.Printf("normalizeDate: Ошибка NextDate для прошедшей даты: %v", err)
				return "", errors.New("invalid repeat format or date calculation error")
			}
			log.Printf("normalizeDate: Прошедшая дата с repeat, рассчитана следующая: '%s'", next)
			return next, nil
		} else {

			result := today.Format(layout)
			log.Printf("normalizeDate: Прошедшая дата без repeat, используется сегодняшняя: '%s'", result)
			return result, nil
		}
	}

	log.Printf("normalizeDate: Дата не в прошлом, используется указанная: '%s'", dateStr)
	return dateStr, nil
}

func truncDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
