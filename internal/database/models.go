package database

import (
	"time"
)

type TaskState string

const (
	Completed TaskState = "complete"
	Pending   TaskState = "pending"
	Ongoing   TaskState = "ongoing"
	Failed    TaskState = "failed"
)

// Task holds details about a task entity
type Task struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id,omitempty"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	State     TaskState `json:"state"`
	ErrorMsg  string    `json:"error_msg,omitempty"`
	AddedAt   time.Time `json:"added_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
	SmtpTo   []string  `json:"smtp_to"`
	AddedAt  time.Time `json:"added_at"`
}
