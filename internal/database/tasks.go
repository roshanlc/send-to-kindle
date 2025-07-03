package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// GetTask retrieves a task
func (db *DB) GetTask(taskID string) (Task, error) {
	query := `SELECT user_id,url,state,error_message,added_at,updated_at FROM tasks WHERE id = ?;`
	result := db.Database.QueryRow(query, taskID)

	if err := result.Err(); err != nil {
		return Task{}, err
	}

	task := Task{
		ID: taskID,
	}
	var userID sql.NullInt32
	var errMsg sql.NullString
	var stateText string
	err := result.Scan(&userID,
		&task.URL,
		&stateText,
		&errMsg,
		&task.AddedAt,
		&task.UpdatedAt)

	if err != nil {
		return Task{}, err
	}

	task.State = TaskState(stateText)

	if userID.Valid {
		task.UserID = int(userID.Int32)
	}

	if errMsg.Valid {
		task.ErrorMsg = errMsg.String
	}

	return task, nil
}

func validateTask(task *Task) error {
	if task.ID == "" {
		return fmt.Errorf("taskID cannot be empty")
	}

	if task.State == "" {
		return fmt.Errorf("state cannot be empty")
	}

	if task.URL == "" {
		return fmt.Errorf("url cannot be empty")
	}

	return nil
}

// AddTask insert a task
func (db *DB) AddTask(task Task) error {
	err := validateTask(&task)
	if err != nil {
		return err
	}

	tx, err := db.Database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO tasks(id, user_id, url, state, error_message) VALUES(?,?,?,?,?);`
	var userID sql.NullInt32
	if task.UserID != 0 {
		userID.Int32 = int32(task.UserID)
		userID.Valid = true
	}
	var errMsg sql.NullString
	if task.ErrorMsg != "" {
		errMsg.String = task.ErrorMsg
		errMsg.Valid = true
	}

	_, err = tx.Exec(query,
		task.ID,
		userID,
		task.URL,
		string(task.State),
		errMsg,
	)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// DeleteTask deletes a task
func (db *DB) DeleteTask(taskID string) error {
	tx, err := db.Database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM tasks WHERE id = ?;`
	result, err := tx.Exec(query, taskID)

	if err != nil {
		return err
	}

	r, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// check the count of rows affected
	if r == 0 {
		return ErrNoRowDeleted
	}
	return nil
}

// UpdateTask updates a task row. Supports updating the state, URL and ErrorMessage of the task only.
// Only provide value for the property to be updated. Keep them empty if field is not be updated.
func (db *DB) UpdateTask(task Task) error {
	if task.ID == "" {
		return fmt.Errorf("taskID cannot be empty")
	}

	tx, err := db.Database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		queryParts []string
		args       []any
	)

	if task.ErrorMsg != "" {
		queryParts = append(queryParts, "error_message = ?")
		args = append(args, task.ErrorMsg)
	}
	if task.State != "" {
		queryParts = append(queryParts, "state = ?")
		args = append(args, string(task.State))
	}
	if task.URL != "" {
		queryParts = append(queryParts, "url = ?")
		args = append(args, task.URL)
	}

	query := fmt.Sprintf(
		`UPDATE tasks SET %s WHERE id = ?;`,
		strings.Join(queryParts, ","))

	args = append(args, task.ID)
	result, err := tx.Exec(query, args...)

	if err != nil {
		return err
	}

	r, err := result.RowsAffected()
	if err != nil {
		return nil
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// check the count of rows affected
	if r == 0 {
		return ErrNoRowDeleted
	}
	return nil
}
