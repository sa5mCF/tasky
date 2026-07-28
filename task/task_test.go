package task

import (
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	var todos Task

	todos.Add("write tests")

	if got := len(todos); got != 1 {
		t.Fatalf("expected 1 task, got %d", got)
	}

	if got := todos[0].Task; got != "write tests" {
		t.Fatalf("expected task text %q, got %q", "write tests", got)
	}

	if todos[0].CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestComplete(t *testing.T) {
	todos := Task{
		{
			Task:      "ship it",
			CreatedAt: time.Now(),
		},
	}

	if err := todos.Complete(1); err != nil {
		t.Fatalf("complete failed: %v", err)
	}

	if !todos[0].Done {
		t.Fatal("expected task to be marked done")
	}

	if todos[0].CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set")
	}

	if todos[0].CompletedAt.IsZero() {
		t.Fatal("expected CompletedAt to be non-zero")
	}
}

func TestDoing(t *testing.T) {
	todos := Task{
		{
			Task:      "in progress",
			CreatedAt: time.Now(),
		},
	}

	if err := todos.Doing(1); err != nil {
		t.Fatalf("doing failed: %v", err)
	}

	if !todos[0].Doing {
		t.Fatal("expected task to be marked doing")
	}
}

func TestDelete(t *testing.T) {
	todos := Task{
		{Task: "one"},
		{Task: "two"},
	}

	if err := todos.Delete(1); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if got := len(todos); got != 1 {
		t.Fatalf("expected 1 task after delete, got %d", got)
	}

	if got := todos[0].Task; got != "two" {
		t.Fatalf("expected remaining task %q, got %q", "two", got)
	}
}

func TestInvalidIndex(t *testing.T) {
	var todos Task

	if err := todos.Complete(1); err == nil {
		t.Fatal("expected complete to fail on empty task list")
	}

	if err := todos.Doing(0); err == nil {
		t.Fatal("expected doing to fail on invalid index")
	}

	if err := todos.Delete(-1); err == nil {
		t.Fatal("expected delete to fail on invalid index")
	}
}

func TestCounter(t *testing.T) {
	todos := Task{
		{Done: false},
		{Done: true},
		{Done: false},
	}

	if got := todos.Counter(); got != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", got)
	}
}
