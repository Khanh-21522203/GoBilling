package event

import "context"

type Repository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	Update(ctx context.Context, event *Event) error
	GetPending(ctx context.Context, limit int) ([]*Event, error)
}
