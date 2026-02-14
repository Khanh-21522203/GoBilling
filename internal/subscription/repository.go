package subscription

import "context"

type Repository interface {
	Create(ctx context.Context, subscription *Subscription) error
	GetByID(ctx context.Context, id string) (*Subscription, error)
	Update(ctx context.Context, subscription *Subscription) error
	List(ctx context.Context, opts ListOptions) ([]*Subscription, error)
	GetDueForRenewal(ctx context.Context, limit int) ([]*Subscription, error)
}

type ListOptions struct {
	Limit         int
	StartingAfter string
	CustomerID    *string
	PlanID        *string
	Status        *Status
}
