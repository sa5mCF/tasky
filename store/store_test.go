package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samEscom/tasky/task"
)

func TestLoadMissingFile(t *testing.T) {
	todos, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("expected missing file to load cleanly, got %v", err)
	}

	if len(todos) != 0 {
		t.Fatalf("expected empty task list, got %d items", len(todos))
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "data.json")

	createdAt := time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, time.July, 28, 12, 30, 0, 0, time.UTC)
	todos := task.Task{
		{
			Task:        "write docs",
			Doing:       true,
			Done:        true,
			CreatedAt:   createdAt,
			CompletedAt: &completedAt,
		},
	}

	if err := Save(filename, todos); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := Load(filename)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got))
	}

	if got[0].Task != todos[0].Task {
		t.Fatalf("expected task %q, got %q", todos[0].Task, got[0].Task)
	}

	if !got[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %v, got %v", createdAt, got[0].CreatedAt)
	}

	if got[0].CompletedAt == nil || !got[0].CompletedAt.Equal(completedAt) {
		t.Fatalf("expected CompletedAt %v, got %v", completedAt, got[0].CompletedAt)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "data.json")

	if err := os.WriteFile(filename, []byte("{"), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	if _, err := Load(filename); err == nil {
		t.Fatal("expected invalid JSON to return an error")
	}
}
