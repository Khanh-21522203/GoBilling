package ledger

import (
	"context"
	"fmt"

	"gobilling/internal/platform/database"
)

type PostgresRepository struct {
	db *database.DB
}

func NewPostgresRepository(db *database.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, tx *LedgerTransaction) error {
	query := `
		INSERT INTO ledger_transactions (
			id, customer_id, type, amount, currency,
			invoice_id, payment_id, refund_id, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		tx.ID,
		tx.CustomerID,
		tx.Type,
		tx.Amount,
		tx.Currency,
		tx.InvoiceID,
		tx.PaymentID,
		tx.RefundID,
		tx.Description,
		tx.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create ledger transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByCustomerID(ctx context.Context, customerID string, limit int) ([]*LedgerTransaction, error) {
	query := `
		SELECT id, customer_id, type, amount, currency,
		       invoice_id, payment_id, refund_id, description, created_at
		FROM ledger_transactions
		WHERE customer_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, customerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*LedgerTransaction, 0)
	for rows.Next() {
		var tx LedgerTransaction
		err := rows.Scan(
			&tx.ID,
			&tx.CustomerID,
			&tx.Type,
			&tx.Amount,
			&tx.Currency,
			&tx.InvoiceID,
			&tx.PaymentID,
			&tx.RefundID,
			&tx.Description,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ledger transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (r *PostgresRepository) GetBalance(ctx context.Context, customerID string) (int64, error) {
	query := `
		SELECT COALESCE(
			SUM(CASE 
				WHEN type IN ('charge') THEN amount
				WHEN type IN ('payment', 'refund', 'credit') THEN -amount
				ELSE 0
			END), 0
		) as balance
		FROM ledger_transactions
		WHERE customer_id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	var balance int64
	err := q.QueryRow(ctx, query, customerID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}
