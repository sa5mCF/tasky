package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samEscom/tasky/task"
)

func TestRepositoryCRUDAndPersistence(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "tasks.db")
	legacyPath := filepath.Join(t.TempDir(), "missing.json")
	createdAt := time.Date(2026, time.July, 28, 10, 0, 0, 123, time.FixedZone("CST", -6*60*60))

	repository, err := Open(ctx, databasePath, legacyPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	first, err := repository.Create(ctx, task.Item{
		Task:      "write adapter",
		Status:    task.StatusTodo,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	second, err := repository.Create(ctx, task.Item{
		Task:      "remove adapter",
		Status:    task.StatusTodo,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("expected stable IDs 1 and 2, got %d and %d", first.ID, second.ID)
	}

	first.MarkDoing()
	if err := repository.Update(ctx, first); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	completedAt := createdAt.Add(time.Hour)
	first.Complete(completedAt)
	if err := repository.Update(ctx, first); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if err := repository.Delete(ctx, second.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if err := repository.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	repository, err = Open(ctx, databasePath, legacyPath)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer repository.Close()

	tasks, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one persisted task, got %d", len(tasks))
	}
	got := tasks[0]
	if got.ID != first.ID || got.Status != task.StatusDone {
		t.Fatalf("unexpected persisted task: %#v", got)
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt %v, got %v", createdAt, got.CreatedAt)
	}
	if got.CompletedAt == nil || !got.CompletedAt.Equal(completedAt) {
		t.Fatalf("expected CompletedAt %v, got %v", completedAt, got.CompletedAt)
	}

	third, err := repository.Create(ctx, task.Item{
		Task:      "stable id",
		Status:    task.StatusTodo,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if third.ID != 3 {
		t.Fatalf("expected ID 3 after deletion, got %d", third.ID)
	}

	if _, err := repository.FindByID(ctx, 99); !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("expected ErrNotFound from FindByID, got %v", err)
	}
	if err := repository.Delete(ctx, 99); !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("expected ErrNotFound from Delete, got %v", err)
	}
}

func TestLegacyJSONMigrationRunsOnce(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	databasePath := filepath.Join(dir, "tasks.db")
	legacyPath := filepath.Join(dir, "tasks.json")
	createdAt := time.Date(2026, time.July, 27, 10, 0, 0, 0, time.UTC)
	completedAt := createdAt.Add(24 * time.Hour)

	legacy := []legacyItem{
		{
			Task:      "doing task",
			Doing:     true,
			CreatedAt: createdAt,
		},
		{
			Task:        "done takes priority",
			Doing:       true,
			Done:        true,
			CreatedAt:   createdAt,
			CompletedAt: &completedAt,
		},
	}
	original, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if err := os.WriteFile(legacyPath, original, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	repository, err := Open(ctx, databasePath, legacyPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	tasks, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected two imported tasks, got %d", len(tasks))
	}
	if tasks[0].ID != 1 || tasks[0].Status != task.StatusDoing {
		t.Fatalf("unexpected first imported task: %#v", tasks[0])
	}
	if tasks[1].ID != 2 || tasks[1].Status != task.StatusDone {
		t.Fatalf("unexpected second imported task: %#v", tasks[1])
	}
	if tasks[1].CompletedAt == nil || !tasks[1].CompletedAt.Equal(completedAt) {
		t.Fatalf("expected preserved completion time, got %v", tasks[1].CompletedAt)
	}

	for _, item := range tasks {
		if err := repository.Delete(ctx, item.ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	}
	if err := repository.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	gotJSON, err := os.ReadFile(legacyPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(gotJSON) != string(original) {
		t.Fatal("expected legacy JSON to remain unchanged")
	}

	repository, err = Open(ctx, databasePath, legacyPath)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer repository.Close()

	tasks, err = repository.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected migration not to run again, got %d tasks", len(tasks))
	}
}

func TestInvalidLegacyJSONRollsBackAndCanRetry(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	databasePath := filepath.Join(dir, "tasks.db")
	legacyPath := filepath.Join(dir, "tasks.json")

	if err := os.WriteFile(legacyPath, []byte("{"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if _, err := Open(ctx, databasePath, legacyPath); err == nil {
		t.Fatal("expected invalid legacy JSON to fail")
	}

	if err := os.WriteFile(legacyPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	repository, err := Open(ctx, databasePath, legacyPath)
	if err != nil {
		t.Fatalf("expected migration retry to succeed, got %v", err)
	}
	defer repository.Close()

	tasks, err := repository.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected empty task list, got %d", len(tasks))
	}
}
