package task

import (
	"errors"
	"time"
)

type Item struct {
	Task        string
	Doing       bool
	Done        bool
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type Task []Item

func (t *Task) Add(task string) {
	todo := Item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: nil,
	}

	*t = append(*t, todo)
}

func (t *Task) Complete(index int) error {
	item, err := t.item(index)
	if err != nil {
		return err
	}

	now := time.Now()
	item.CompletedAt = &now
	item.Done = true

	return nil
}

func (t *Task) Delete(index int) error {
	if err := t.validateIndex(index); err != nil {
		return err
	}

	ls := *t
	*t = append(ls[:index-1], ls[index:]...)

	return nil
}

func (t *Task) Doing(index int) error {
	item, err := t.item(index)
	if err != nil {
		return err
	}

	item.Doing = true

	return nil
}

func (t *Task) Counter() int {
	total := 0

	for _, item := range *t {
		if !item.Done {
			total++
		}
	}

	return total
}

func (t *Task) item(index int) (*Item, error) {
	if err := t.validateIndex(index); err != nil {
		return nil, err
	}

	return &(*t)[index-1], nil
}

func (t *Task) validateIndex(index int) error {
	if index <= 0 || index > len(*t) {
		return errors.New("invalid index")
	}
	return nil
}
