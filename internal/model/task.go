package model

import (
	"strconv"
	"sync/atomic"
	"time"
)

var taskCounter uint64

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DurationStr string     `json:"duration,omitempty"`
	Result      string     `json:"result,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

func NewTask(title, description string) *Task {
	return &Task{
		ID:          generateTaskID(),
		Title:       title,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
	}
}

// generateTaskID generates a sequential task ID
func generateTaskID() string {
	return strconv.FormatUint(atomic.AddUint64(&taskCounter, 1), 10)
}

// Duration returns the duration of task processing in milliseconds
func (t *Task) Duration() float64 {
	if t.StartedAt == nil {
		return 0
	}
	endTime := time.Now()
	if t.CompletedAt != nil {
		endTime = *t.CompletedAt
	}
	return float64(endTime.Sub(*t.StartedAt).Milliseconds())
}

// UpdateStatus updates the task status and related timestamps
func (t *Task) UpdateStatus(status TaskStatus) {
	t.Status = status
	now := time.Now()

	switch status {
	case StatusProcessing:
		t.StartedAt = &now
	case StatusCompleted, StatusFailed:
		t.CompletedAt = &now
	}

	if t.StartedAt != nil {
		endTime := now
		if t.CompletedAt != nil {
			endTime = *t.CompletedAt
		}
		t.DurationStr = endTime.Sub(*t.StartedAt).String()
	}
}
