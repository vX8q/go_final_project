package db

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TaskForAPI struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (t *Task) ToTaskForAPI() TaskForAPI {
	return TaskForAPI{
		ID:      strconv.FormatInt(t.ID, 10),
		Date:    t.Date,
		Title:   t.Title,
		Comment: t.Comment,
		Repeat:  t.Repeat,
	}
}

func AddTask(db *sql.DB, task *Task) (int64, error) {
	const query = `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`
	log.Printf("AddTask: Выполнение запроса: %s, task: %+v", query, task)
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		log.Printf("AddTask: Ошибка выполнения запроса: %v", err)
		return 0, fmt.Errorf("AddTask: ошибка выполнения запроса: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("AddTask: Ошибка получения LastInsertId: %v", err)
		return 0, fmt.Errorf("AddTask: ошибка получения LastInsertId: %w", err)
	}
	log.Printf("AddTask: Успешно добавлена задача с ID: %d", id)
	return id, nil
}

func scanTasks(rows *sql.Rows) ([]*Task, error) {
	var tasks []*Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			log.Printf("scanTasks: Ошибка сканирования строки задачи: %v", err)
			continue
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		log.Printf("scanTasks: Ошибка после итерации по строкам: %v", err)
		return nil, fmt.Errorf("ошибка при чтении результатов из БД: %w", err)
	}

	if tasks == nil {
		log.Println("scanTasks: Нет задач для возврата")
		return []*Task{}, nil
	}
	log.Printf("scanTasks: Успешно считано %d задач", len(tasks))
	return tasks, nil
}

func FetchTasks(db *sql.DB, limit int) ([]*Task, error) {
	if db == nil {
		return nil, fmt.Errorf("FetchTasks: экземпляр БД не должен быть nil")
	}

	query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		ORDER BY date ASC
		LIMIT ?
	`
	log.Printf("FetchTasks: Выполнение запроса: %s, limit: %d", query, limit)
	rows, err := db.Query(query, limit)
	if err != nil {
		log.Printf("FetchTasks: Ошибка выполнения запроса: %v", err)
		return nil, fmt.Errorf("ошибка получения задач из БД: %w", err)
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func SearchTasks(db *sql.DB, searchQuery string, limit int) ([]*Task, error) {
	if db == nil {
		return nil, fmt.Errorf("SearchTasks: экземпляр БД не должен быть nil")
	}

	datePattern := `^\d{2}\.\d{2}\.\d{4}$`
	isDate, _ := regexp.MatchString(datePattern, searchQuery)

	var rows *sql.Rows
	var err error
	var query string

	if isDate {
		parsedDate, parseErr := time.Parse("02.01.2006", searchQuery)
		if parseErr != nil {
			log.Printf("SearchTasks: Ошибка парсинга даты '%s': %v", searchQuery, parseErr)
			return nil, fmt.Errorf("неверный формат даты для поиска: %w", parseErr)
		}
		dbDateStr := parsedDate.Format("20060102")

		query = `
			SELECT id, date, title, comment, repeat
			FROM scheduler
			WHERE date = ?
			ORDER BY date ASC
			LIMIT ?`
		log.Printf("SearchTasks: Выполнение запроса по дате: %s, date: %s, limit: %d", query, dbDateStr, limit)
		rows, err = db.Query(query, dbDateStr, limit)
	} else {
		searchPattern := "%" + strings.ToLower(searchQuery) + "%"
		query = `
			SELECT id, date, title, comment, repeat
			FROM scheduler
			WHERE LOWER(title) LIKE ? OR LOWER(comment) LIKE ?
			ORDER BY date ASC
			LIMIT ?`
		log.Printf("SearchTasks: Выполнение запроса по тексту: %s, searchPattern: %s, limit: %d", query, searchPattern, limit)
		rows, err = db.Query(query, searchPattern, searchPattern, limit)
	}

	if err != nil {
		log.Printf("SearchTasks: Ошибка выполнения запроса: %v", err)
		return nil, fmt.Errorf("ошибка поиска задач в БД: %w", err)
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func GetTask(db *sql.DB, idStr string) (*Task, error) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("GetTask: Некорректный формат ID '%s': %v", idStr, err)
		return nil, fmt.Errorf("некорректный идентификатор задачи")
	}

	query := `
		SELECT id, date, title, comment, repeat
		FROM scheduler
		WHERE id = ?
	`
	log.Printf("GetTask: Выполнение запроса: %s, id: %d", query, id)
	row := db.QueryRow(query, id)

	var task Task
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("GetTask: Задача с ID %d не найдена", id)
			return nil, fmt.Errorf("задача не найдена")
		}
		log.Printf("GetTask: Ошибка сканирования строки: %v", err)
		return nil, fmt.Errorf("ошибка получения задачи из БД: %w", err)
	}

	log.Printf("GetTask: Успешно получена задача с ID: %d", id)
	return &task, nil
}

func UpdateTask(db *sql.DB, task *Task) error {
	query := `
		UPDATE scheduler
		SET date = ?, title = ?, comment = ?, repeat = ?
		WHERE id = ?
	`
	log.Printf("UpdateTask: Выполнение запроса: %s, task: %+v", query, task)
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		log.Printf("UpdateTask: Ошибка выполнения запроса: %v", err)
		return fmt.Errorf("ошибка обновления задачи в БД: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Printf("UpdateTask: Ошибка получения RowsAffected: %v", err)
		return fmt.Errorf("ошибка при проверке обновления: %w", err)
	}

	if count == 0 {
		log.Printf("UpdateTask: Не найдена задача с ID %d для обновления", task.ID)
		return fmt.Errorf(`incorrect id for updating task`)
	}

	log.Printf("UpdateTask: Успешно обновлена задача с ID: %d", task.ID)
	return nil
}

func DeleteTask(db *sql.DB, idStr string) error {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("DeleteTask: Некорректный формат ID '%s': %v", idStr, err)
		return fmt.Errorf("некорректный идентификатор задачи")
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	log.Printf("DeleteTask: Выполнение запроса: %s, id: %d", query, id)
	res, err := db.Exec(query, id)
	if err != nil {
		log.Printf("DeleteTask: Ошибка выполнения запроса: %v", err)
		return fmt.Errorf("ошибка удаления задачи из БД: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Printf("DeleteTask: Ошибка получения RowsAffected: %v", err)
		return fmt.Errorf("ошибка при проверке удаления: %w", err)
	}

	if count == 0 {
		log.Printf("DeleteTask: Не найдена задача с ID %d для удаления", id)
		return fmt.Errorf("задача не найдена")
	}

	log.Printf("DeleteTask: Успешно удалена задача с ID: %d", id)
	return nil
}
