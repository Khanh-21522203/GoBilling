package invoice

import "context"

type Repository interface {
	Create(ctx context.Context, invoice *Invoice) error
	GetByID(ctx context.Context, id string) (*Invoice, error)
	GetByIDWithLineItems(ctx context.Context, id string) (*Invoice, error)
	Update(ctx context.Context, invoice *Invoice) error
	List(ctx context.Context, opts ListOptions) ([]*Invoice, error)
	CreateLineItem(ctx context.Context, item *LineItem) error
	GetLineItems(ctx context.Context, invoiceID string) ([]*LineItem, error)
}

type ListOptions struct {
	Limit          int
	StartingAfter  string
	CustomerID     *string
	SubscriptionID *string
	Status         *Status
}
