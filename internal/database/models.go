package database

import "time"

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
	State     TaskState `json:"state"`
	ErrorMsg  string    `json:"error_msg,omitempty"`
	AddedAt   time.Time `json:"added_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
