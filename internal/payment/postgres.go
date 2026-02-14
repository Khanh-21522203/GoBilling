package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gobilling/internal/platform/database"
)

type PostgresRepository struct {
	db *database.DB
}

func NewPostgresRepository(db *database.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (
			id, invoice_id, amount, currency, status,
			payment_method_id, provider_id, failure_code, failure_message,
			idempotency_key, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		payment.ID,
		payment.InvoiceID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethodID,
		payment.ProviderID,
		payment.FailureCode,
		payment.FailureMessage,
		payment.IdempotencyKey,
		payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Payment, error) {
	query := `
		SELECT id, invoice_id, amount, currency, status,
		       payment_method_id, provider_id, failure_code, failure_message,
		       idempotency_key, metadata, created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var p Payment
	err := row.Scan(
		&p.ID,
		&p.InvoiceID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.PaymentMethodID,
		&p.ProviderID,
		&p.FailureCode,
		&p.FailureMessage,
		&p.IdempotencyKey,
		&p.Metadata,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &p, nil
}

func (r *PostgresRepository) Update(ctx context.Context, payment *Payment) error {
	query := `
		UPDATE payments
		SET status = $2, provider_id = $3, failure_code = $4,
		    failure_message = $5, metadata = $6, updated_at = $7
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		payment.ID,
		payment.Status,
		payment.ProviderID,
		payment.FailureCode,
		payment.FailureMessage,
		payment.Metadata,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context, opts ListOptions) ([]*Payment, error) {
	query := `
		SELECT id, invoice_id, amount, currency, status,
		       payment_method_id, provider_id, failure_code, failure_message,
		       idempotency_key, metadata, created_at, updated_at
		FROM payments
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if opts.InvoiceID != nil {
		query += fmt.Sprintf(" AND invoice_id = $%d", argPos)
		args = append(args, *opts.InvoiceID)
		argPos++
	}

	if opts.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *opts.Status)
		argPos++
	}

	if opts.StartingAfter != "" {
		query += fmt.Sprintf(" AND id > $%d", argPos)
		args = append(args, opts.StartingAfter)
		argPos++
	}

	query += " ORDER BY id ASC"
	query += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, opts.Limit)

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	payments := make([]*Payment, 0)
	for rows.Next() {
		var p Payment
		err := rows.Scan(
			&p.ID,
			&p.InvoiceID,
			&p.Amount,
			&p.Currency,
			&p.Status,
			&p.PaymentMethodID,
			&p.ProviderID,
			&p.FailureCode,
			&p.FailureMessage,
			&p.IdempotencyKey,
			&p.Metadata,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &p)
	}

	return payments, nil
}

func (r *PostgresRepository) GetByInvoiceID(ctx context.Context, invoiceID string) ([]*Payment, error) {
	status := StatusSucceeded
	return r.List(ctx, ListOptions{
		InvoiceID: &invoiceID,
		Status:    &status,
		Limit:     100,
	})
}

type PostgresRefundRepository struct {
	db *database.DB
}

func NewPostgresRefundRepository(db *database.DB) *PostgresRefundRepository {
	return &PostgresRefundRepository{db: db}
}

func (r *PostgresRefundRepository) Create(ctx context.Context, refund *Refund) error {
	query := `
		INSERT INTO refunds (
			id, payment_id, amount, currency, status, reason,
			provider_id, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		refund.ID,
		refund.PaymentID,
		refund.Amount,
		refund.Currency,
		refund.Status,
		refund.Reason,
		refund.ProviderID,
		refund.Metadata,
		refund.CreatedAt,
		refund.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}

	return nil
}

func (r *PostgresRefundRepository) GetByID(ctx context.Context, id string) (*Refund, error) {
	query := `
		SELECT id, payment_id, amount, currency, status, reason,
		       provider_id, metadata, created_at, updated_at
		FROM refunds
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var refund Refund
	err := row.Scan(
		&refund.ID,
		&refund.PaymentID,
		&refund.Amount,
		&refund.Currency,
		&refund.Status,
		&refund.Reason,
		&refund.ProviderID,
		&refund.Metadata,
		&refund.CreatedAt,
		&refund.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefundNotFound
		}
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return &refund, nil
}

func (r *PostgresRefundRepository) GetByPaymentID(ctx context.Context, paymentID string) (*Refund, error) {
	query := `
		SELECT id, payment_id, amount, currency, status, reason,
		       provider_id, metadata, created_at, updated_at
		FROM refunds
		WHERE payment_id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, paymentID)

	var refund Refund
	err := row.Scan(
		&refund.ID,
		&refund.PaymentID,
		&refund.Amount,
		&refund.Currency,
		&refund.Status,
		&refund.Reason,
		&refund.ProviderID,
		&refund.Metadata,
		&refund.CreatedAt,
		&refund.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefundNotFound
		}
		return nil, fmt.Errorf("failed to get refund: %w", err)
	}

	return &refund, nil
}

func (r *PostgresRefundRepository) Update(ctx context.Context, refund *Refund) error {
	query := `
		UPDATE refunds
		SET status = $2, provider_id = $3, updated_at = $4
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		refund.ID,
		refund.Status,
		refund.ProviderID,
		refund.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update refund: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrRefundNotFound
	}

	return nil
}
