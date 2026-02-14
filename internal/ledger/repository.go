package ledger

import "context"

type Repository interface {
	Create(ctx context.Context, tx *LedgerTransaction) error
	GetByCustomerID(ctx context.Context, customerID string, limit int) ([]*LedgerTransaction, error)
	GetBalance(ctx context.Context, customerID string) (int64, error)
}
