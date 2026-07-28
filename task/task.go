package task

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyTask = errors.New("empty task, not allowed")
	ErrNotFound  = errors.New("task not found")
)

type Status string

const (
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

type Item struct {
	ID          int64
	Task        string
	Status      Status
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type Task []Item

func NewItem(description string, now time.Time) (Item, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return Item{}, ErrEmptyTask
	}

	return Item{
		Task:        description,
		Status:      StatusTodo,
		CreatedAt:   now,
		CompletedAt: nil,
	}, nil
}

func (i *Item) MarkDoing() {
	i.Status = StatusDoing
	i.CompletedAt = nil
}

func (i *Item) Complete(now time.Time) {
	i.Status = StatusDone
	i.CompletedAt = &now
}

func (t Task) Counter() int {
	total := 0

	for _, item := range t {
		if item.Status != StatusDone {
			total++
		}
	}

	return total
}
