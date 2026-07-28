package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	sqliterepository "github.com/samEscom/tasky/adapter/sqlite"
	"github.com/samEscom/tasky/application"
	"github.com/samEscom/tasky/task"
)

func TestAddTaskPersistsAndUpdatesList(t *testing.T) {
	ctx, service := newTestService(t)
	model := New(ctx, service, nil)
	model.operation = opAdd

	updated, _ := model.activateOperation()
	model = updated.(Model)
	model.input.SetValue("write TUI tests")

	updated, _ = model.updateInput(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.editing {
		t.Fatal("expected add input to close after saving")
	}
	if len(model.tasks) != 1 || model.tasks[0].Task != "write TUI tests" {
		t.Fatalf("expected added task in TUI state, got %#v", model.tasks)
	}
	if model.operation != opList || model.focus != tasksFocus {
		t.Fatalf("expected task list to become active after add, got operation=%d focus=%d", model.operation, model.focus)
	}

	got, err := service.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 1 || got[0].Task != "write TUI tests" {
		t.Fatalf("expected saved task, got %#v", got)
	}
	if got[0].ID == 0 {
		t.Fatal("expected SQLite to assign a stable ID")
	}
}

func TestCompleteTaskPersistsAndReturnsToList(t *testing.T) {
	ctx, service := newTestService(t)
	created, err := service.Add(ctx, "ship feature")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	model := New(ctx, service, task.Task{created})
	model.operation = opComplete
	model.focus = tasksFocus

	model.executeTaskOperation()

	if model.tasks[0].Status != task.StatusDone {
		t.Fatal("expected selected task to be completed")
	}
	if model.operation != opList {
		t.Fatalf("expected operation to reset to list, got %d", model.operation)
	}

	got, err := service.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 1 || got[0].Status != task.StatusDone {
		t.Fatalf("expected completed task on disk, got %#v", got)
	}
}

func TestNavigationUsesBothSections(t *testing.T) {
	model := New(context.Background(), nil, task.Task{
		{ID: 1, Task: "first", Status: task.StatusTodo},
		{ID: 2, Task: "second", Status: task.StatusTodo},
	})

	model.moveSelection(1)
	if model.operation != opAdd {
		t.Fatalf("expected menu to move to add, got %d", model.operation)
	}

	model.focus = tasksFocus
	model.moveSelection(1)
	if model.selectedTask != 1 {
		t.Fatalf("expected task selection to move to second task, got %d", model.selectedTask)
	}

	model.moveSelection(1)
	if model.selectedTask != 0 {
		t.Fatalf("expected task selection to wrap, got %d", model.selectedTask)
	}
}

func TestViewContainsBothSections(t *testing.T) {
	model := New(context.Background(), nil, task.Task{
		{ID: 4, Task: "first", Status: task.StatusTodo},
		{ID: 7, Task: "second", Status: task.StatusDoing},
		{ID: 9, Task: "third", Status: task.StatusDone},
	})
	view := model.View()

	for _, section := range []string{"Operations", "Tasks", "4  [TODO]", "7  [DOING]", "9  [DONE]"} {
		if !strings.Contains(view, section) {
			t.Fatalf("expected view to contain %q", section)
		}
	}

	if !strings.Contains(ansi.Strip(view), "╮  ╭") {
		t.Fatal("expected a visible gap between the operation and task panels")
	}
}

func newTestService(t *testing.T) (context.Context, *application.Service) {
	t.Helper()

	ctx := context.Background()
	dir := t.TempDir()
	repository, err := sqliterepository.Open(
		ctx,
		filepath.Join(dir, "tasks.db"),
		filepath.Join(dir, "missing.json"),
	)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() {
		if err := repository.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})

	return ctx, application.NewService(repository, nil)
}
