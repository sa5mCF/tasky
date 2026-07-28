package task

import (
	"errors"
	"testing"
	"time"
)

func TestNewItem(t *testing.T) {
	now := time.Date(2026, time.July, 28, 10, 0, 0, 0, time.UTC)

	item, err := NewItem("  write tests  ", now)
	if err != nil {
		t.Fatalf("NewItem failed: %v", err)
	}

	if item.Task != "write tests" {
		t.Fatalf("expected trimmed description, got %q", item.Task)
	}
	if item.Status != StatusTodo {
		t.Fatalf("expected todo status, got %q", item.Status)
	}
	if !item.CreatedAt.Equal(now) {
		t.Fatalf("expected CreatedAt %v, got %v", now, item.CreatedAt)
	}
}

func TestNewItemRejectsEmptyDescription(t *testing.T) {
	if _, err := NewItem("   ", time.Now()); !errors.Is(err, ErrEmptyTask) {
		t.Fatalf("expected ErrEmptyTask, got %v", err)
	}
}

func TestMarkDoingReopensCompletedTask(t *testing.T) {
	completedAt := time.Now()
	item := Item{
		Status:      StatusDone,
		CompletedAt: &completedAt,
	}

	item.MarkDoing()

	if item.Status != StatusDoing {
		t.Fatalf("expected doing status, got %q", item.Status)
	}
	if item.CompletedAt != nil {
		t.Fatal("expected CompletedAt to be cleared")
	}
}

func TestComplete(t *testing.T) {
	completedAt := time.Date(2026, time.July, 28, 12, 30, 0, 0, time.UTC)
	item := Item{Status: StatusDoing}

	item.Complete(completedAt)

	if item.Status != StatusDone {
		t.Fatalf("expected done status, got %q", item.Status)
	}
	if item.CompletedAt == nil || !item.CompletedAt.Equal(completedAt) {
		t.Fatalf("expected CompletedAt %v, got %v", completedAt, item.CompletedAt)
	}
}

func TestCounter(t *testing.T) {
	todos := Task{
		{Status: StatusTodo},
		{Status: StatusDone},
		{Status: StatusDoing},
	}

	if got := todos.Counter(); got != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", got)
	}
}
