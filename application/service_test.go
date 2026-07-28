package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/samEscom/tasky/task"
)

type fakeRepository struct {
	tasks  task.Task
	nextID int64
}

func (r *fakeRepository) List(context.Context) (task.Task, error) {
	return append(task.Task(nil), r.tasks...), nil
}

func (r *fakeRepository) FindByID(_ context.Context, id int64) (task.Item, error) {
	for _, item := range r.tasks {
		if item.ID == id {
			return item, nil
		}
	}
	return task.Item{}, task.ErrNotFound
}

func (r *fakeRepository) Create(_ context.Context, item task.Item) (task.Item, error) {
	r.nextID++
	item.ID = r.nextID
	r.tasks = append(r.tasks, item)
	return item, nil
}

func (r *fakeRepository) Update(_ context.Context, item task.Item) error {
	for index := range r.tasks {
		if r.tasks[index].ID == item.ID {
			r.tasks[index] = item
			return nil
		}
	}
	return task.ErrNotFound
}

func (r *fakeRepository) Delete(_ context.Context, id int64) error {
	for index := range r.tasks {
		if r.tasks[index].ID == id {
			r.tasks = append(r.tasks[:index], r.tasks[index+1:]...)
			return nil
		}
	}
	return task.ErrNotFound
}

func TestServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, time.July, 28, 10, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository, func() time.Time { return now })

	created, err := service.Add(ctx, "  write repository  ")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if created.ID != 1 || created.Task != "write repository" || created.Status != task.StatusTodo {
		t.Fatalf("unexpected created task: %#v", created)
	}

	if err := service.MarkDoing(ctx, created.ID); err != nil {
		t.Fatalf("MarkDoing failed: %v", err)
	}
	doing, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if doing.Status != task.StatusDoing {
		t.Fatalf("expected doing status, got %q", doing.Status)
	}

	now = now.Add(2 * time.Hour)
	if err := service.Complete(ctx, created.ID); err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	completed, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if completed.Status != task.StatusDone {
		t.Fatalf("expected done status, got %q", completed.Status)
	}
	if completed.CompletedAt == nil || !completed.CompletedAt.Equal(now) {
		t.Fatalf("expected completion time %v, got %v", now, completed.CompletedAt)
	}

	if err := service.MarkDoing(ctx, created.ID); err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	reopened, err := repository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if reopened.Status != task.StatusDoing || reopened.CompletedAt != nil {
		t.Fatalf("expected reopened task, got %#v", reopened)
	}

	if err := service.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := repository.FindByID(ctx, created.ID); !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestServiceRejectsEmptyTask(t *testing.T) {
	service := NewService(&fakeRepository{}, time.Now)

	if _, err := service.Add(context.Background(), "   "); !errors.Is(err, task.ErrEmptyTask) {
		t.Fatalf("expected ErrEmptyTask, got %v", err)
	}
}

func TestServiceReturnsNotFound(t *testing.T) {
	service := NewService(&fakeRepository{}, time.Now)

	if err := service.Complete(context.Background(), 42); !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
