package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
)

var dbMutex sync.Mutex

type Task struct {
	ID      int64  `db:"id" json:"id"`
	Date    string `db:"date" json:"date"`
	Title   string `db:"title" json:"title"`
	Comment string `db:"comment" json:"comment,omitempty"`
	Repeat  string `db:"repeat" json:"repeat,omitempty"`
}

func AddTask(t *Task) (int64, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	res, err := DB.Exec(
		"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		t.Date, t.Title, t.Comment, t.Repeat,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func UpdateTask(t *Task) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := DB.Exec(
		"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		t.Date, t.Title, t.Comment, t.Repeat, t.ID,
	)
	return err
}

func Tasks() ([]*Task, error) {
	var tasks []*Task
	err := DB.Select(&tasks, "SELECT id, date, title, comment, repeat FROM scheduler")
	return tasks, err
}

func ClearTasks() error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := DB.Exec("DELETE FROM scheduler")
	return err
}

func GetTask(id string) (*Task, error) {
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, err
	}
	return GetTaskByID(idInt)
}

func GetTaskByID(id int64) (*Task, error) {
	var t Task
	err := DB.Get(&t, "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found")
	}
	return &t, err
}

func DeleteTask(id int64) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
	return err
}
