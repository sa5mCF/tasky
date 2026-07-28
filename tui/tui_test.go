package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/samEscom/tasky/store"
	"github.com/samEscom/tasky/task"
)

func TestAddTaskPersistsAndUpdatesList(t *testing.T) {
	filename := t.TempDir() + "/tasks.json"
	model := New(filename, nil)
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

	got, err := store.Load(filename)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(got) != 1 || got[0].Task != "write TUI tests" {
		t.Fatalf("expected saved task, got %#v", got)
	}
}

func TestCompleteTaskPersistsAndReturnsToList(t *testing.T) {
	filename := t.TempDir() + "/tasks.json"
	model := New(filename, task.Task{{Task: "ship feature", CreatedAt: time.Now()}})
	model.operation = opComplete
	model.focus = tasksFocus

	model.executeTaskOperation()

	if !model.tasks[0].Done {
		t.Fatal("expected selected task to be completed")
	}
	if model.operation != opList {
		t.Fatalf("expected operation to reset to list, got %d", model.operation)
	}

	got, err := store.Load(filename)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(got) != 1 || !got[0].Done {
		t.Fatalf("expected completed task on disk, got %#v", got)
	}
}

func TestNavigationUsesBothSections(t *testing.T) {
	model := New("tasks.json", task.Task{{Task: "first"}, {Task: "second"}})

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
	model := New("tasks.json", task.Task{
		{Task: "first"},
		{Task: "second", Doing: true},
		{Task: "third", Done: true},
	})
	view := model.View()

	for _, section := range []string{"Operations", "Tasks", "first", "DOING", "DONE"} {
		if !strings.Contains(view, section) {
			t.Fatalf("expected view to contain %q", section)
		}
	}

	if !strings.Contains(ansi.Strip(view), "╮  ╭") {
		t.Fatal("expected a visible gap between the operation and task panels")
	}
}
