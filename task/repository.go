package task

import "context"

type Repository interface {
	List(context.Context) (Task, error)
	FindByID(context.Context, int64) (Item, error)
	Create(context.Context, Item) (Item, error)
	Update(context.Context, Item) error
	Delete(context.Context, int64) error
}
