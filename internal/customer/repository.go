package customer

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, id string) (*Customer, error)
	GetByEmail(ctx context.Context, email string) (*Customer, error)
	Update(ctx context.Context, customer *Customer) error
	List(ctx context.Context, opts ListOptions) ([]*Customer, error)
	Delete(ctx context.Context, id string) error
	HasActiveSubscriptions(ctx context.Context, customerID string) (bool, error)
}

type ListOptions struct {
	Limit         int
	StartingAfter string
	Email         *string
	Status        *Status
	ExternalID    *string
}
