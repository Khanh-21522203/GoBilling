package product

import "context"

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, product *Product) error
	List(ctx context.Context, opts ProductListOptions) ([]*Product, error)
}

type ProductListOptions struct {
	Limit         int
	StartingAfter string
	Active        *bool
}

type PlanRepository interface {
	Create(ctx context.Context, plan *Plan) error
	GetByID(ctx context.Context, id string) (*Plan, error)
	Update(ctx context.Context, plan *Plan) error
	List(ctx context.Context, opts PlanListOptions) ([]*Plan, error)
	ListByProductID(ctx context.Context, productID string) ([]*Plan, error)
}

type PlanListOptions struct {
	Limit         int
	StartingAfter string
	ProductID     *string
	Active        *bool
}
