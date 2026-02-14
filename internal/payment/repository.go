package payment

import "context"

type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	List(ctx context.Context, opts ListOptions) ([]*Payment, error)
	GetByInvoiceID(ctx context.Context, invoiceID string) ([]*Payment, error)
}

type ListOptions struct {
	Limit         int
	StartingAfter string
	InvoiceID     *string
	Status        *Status
}

type RefundRepository interface {
	Create(ctx context.Context, refund *Refund) error
	GetByID(ctx context.Context, id string) (*Refund, error)
	GetByPaymentID(ctx context.Context, paymentID string) (*Refund, error)
	Update(ctx context.Context, refund *Refund) error
}

type RetryRepository interface {
	Create(ctx context.Context, retry *PaymentRetry) error
	GetDueRetries(ctx context.Context, limit int) ([]*PaymentRetry, error)
	Update(ctx context.Context, retry *PaymentRetry) error
}
