package queue

import (
	"sync"

	"github.com/google/uuid"
)

type Task struct {
	ID  uuid.UUID
	URL string
}

type TaskQueue struct {
	queue []Task
	lock  sync.Mutex
	cond  *sync.Cond
}

// NewTaskQueue returns a new TaskQueue
func NewTaskQueue() *TaskQueue {
	q := &TaskQueue{
		queue: []Task{},
	}
	q.cond = sync.NewCond(&q.lock)
	return q
}

// Enqueue adds a task to the queue
func (t *TaskQueue) Enqueue(task Task) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.queue = append(t.queue, task)
	t.cond.Signal()
}

// Dequeue blocks until a task is available
func (t *TaskQueue) Dequeue() Task {
	t.lock.Lock()
	defer t.lock.Unlock()

	if len(t.queue) == 0 {
		t.cond.Wait()
	}
	task := t.queue[0]
	t.queue = t.queue[1:]
	return task
}

// NewTask creates a task entity
func NewTask(id uuid.UUID, url string) Task {
	return Task{ID: id, URL: url}
}
