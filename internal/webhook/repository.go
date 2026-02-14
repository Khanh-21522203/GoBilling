package webhook

import "context"

type EndpointRepository interface {
	Create(ctx context.Context, endpoint *WebhookEndpoint) error
	GetByID(ctx context.Context, id string) (*WebhookEndpoint, error)
	Update(ctx context.Context, endpoint *WebhookEndpoint) error
	List(ctx context.Context, active *bool, limit int) ([]*WebhookEndpoint, error)
	Delete(ctx context.Context, id string) error
}

type DeliveryRepository interface {
	Create(ctx context.Context, delivery *WebhookDelivery) error
	GetByID(ctx context.Context, id string) (*WebhookDelivery, error)
	Update(ctx context.Context, delivery *WebhookDelivery) error
	GetPending(ctx context.Context, limit int) ([]*WebhookDelivery, error)
	ListByEndpoint(ctx context.Context, endpointID string, limit int) ([]*WebhookDelivery, error)
}
