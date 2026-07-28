package application

import (
	"context"
	"time"

	"github.com/samEscom/tasky/task"
)

type Service struct {
	repository task.Repository
	now        func() time.Time
}

func NewService(repository task.Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}

	return &Service{
		repository: repository,
		now:        now,
	}
}

func (s *Service) List(ctx context.Context) (task.Task, error) {
	return s.repository.List(ctx)
}

func (s *Service) Add(ctx context.Context, description string) (task.Item, error) {
	item, err := task.NewItem(description, s.now())
	if err != nil {
		return task.Item{}, err
	}

	return s.repository.Create(ctx, item)
}

func (s *Service) MarkDoing(ctx context.Context, id int64) error {
	item, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	item.MarkDoing()
	return s.repository.Update(ctx, item)
}

func (s *Service) Complete(ctx context.Context, id int64) error {
	item, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	item.Complete(s.now())
	return s.repository.Update(ctx, item)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}
