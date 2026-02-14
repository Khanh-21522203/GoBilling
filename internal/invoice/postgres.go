package invoice

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

func (r *PostgresRepository) Create(ctx context.Context, invoice *Invoice) error {
	query := `
		INSERT INTO invoices (
			id, invoice_number, customer_id, subscription_id, status, currency,
			subtotal, discount_amount, tax_amount, total, amount_paid, amount_due,
			period_start, period_end, due_date, paid_at, voided_at,
			metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		invoice.ID,
		invoice.InvoiceNumber,
		invoice.CustomerID,
		invoice.SubscriptionID,
		invoice.Status,
		invoice.Currency,
		invoice.Subtotal,
		invoice.DiscountAmount,
		invoice.TaxAmount,
		invoice.Total,
		invoice.AmountPaid,
		invoice.AmountDue,
		invoice.PeriodStart,
		invoice.PeriodEnd,
		invoice.DueDate,
		invoice.PaidAt,
		invoice.VoidedAt,
		invoice.Metadata,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Invoice, error) {
	query := `
		SELECT id, invoice_number, customer_id, subscription_id, status, currency,
		       subtotal, discount_amount, tax_amount, total, amount_paid, amount_due,
		       period_start, period_end, due_date, paid_at, voided_at,
		       metadata, created_at, updated_at
		FROM invoices
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var inv Invoice
	err := row.Scan(
		&inv.ID,
		&inv.InvoiceNumber,
		&inv.CustomerID,
		&inv.SubscriptionID,
		&inv.Status,
		&inv.Currency,
		&inv.Subtotal,
		&inv.DiscountAmount,
		&inv.TaxAmount,
		&inv.Total,
		&inv.AmountPaid,
		&inv.AmountDue,
		&inv.PeriodStart,
		&inv.PeriodEnd,
		&inv.DueDate,
		&inv.PaidAt,
		&inv.VoidedAt,
		&inv.Metadata,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &inv, nil
}

func (r *PostgresRepository) GetByIDWithLineItems(ctx context.Context, id string) (*Invoice, error) {
	invoice, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lineItems, err := r.GetLineItems(ctx, id)
	if err != nil {
		return nil, err
	}

	invoice.LineItems = lineItems
	return invoice, nil
}

func (r *PostgresRepository) Update(ctx context.Context, invoice *Invoice) error {
	query := `
		UPDATE invoices
		SET status = $2, subtotal = $3, discount_amount = $4, tax_amount = $5,
		    total = $6, amount_paid = $7, amount_due = $8,
		    due_date = $9, paid_at = $10, voided_at = $11,
		    metadata = $12, updated_at = $13
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		invoice.ID,
		invoice.Status,
		invoice.Subtotal,
		invoice.DiscountAmount,
		invoice.TaxAmount,
		invoice.Total,
		invoice.AmountPaid,
		invoice.AmountDue,
		invoice.DueDate,
		invoice.PaidAt,
		invoice.VoidedAt,
		invoice.Metadata,
		invoice.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context, opts ListOptions) ([]*Invoice, error) {
	query := `
		SELECT id, invoice_number, customer_id, subscription_id, status, currency,
		       subtotal, discount_amount, tax_amount, total, amount_paid, amount_due,
		       period_start, period_end, due_date, paid_at, voided_at,
		       metadata, created_at, updated_at
		FROM invoices
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if opts.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argPos)
		args = append(args, *opts.CustomerID)
		argPos++
	}

	if opts.SubscriptionID != nil {
		query += fmt.Sprintf(" AND subscription_id = $%d", argPos)
		args = append(args, *opts.SubscriptionID)
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
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	invoices := make([]*Invoice, 0)
	for rows.Next() {
		var inv Invoice
		err := rows.Scan(
			&inv.ID,
			&inv.InvoiceNumber,
			&inv.CustomerID,
			&inv.SubscriptionID,
			&inv.Status,
			&inv.Currency,
			&inv.Subtotal,
			&inv.DiscountAmount,
			&inv.TaxAmount,
			&inv.Total,
			&inv.AmountPaid,
			&inv.AmountDue,
			&inv.PeriodStart,
			&inv.PeriodEnd,
			&inv.DueDate,
			&inv.PaidAt,
			&inv.VoidedAt,
			&inv.Metadata,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &inv)
	}

	return invoices, nil
}

func (r *PostgresRepository) CreateLineItem(ctx context.Context, item *LineItem) error {
	query := `
		INSERT INTO invoice_line_items (
			id, invoice_id, description, quantity, unit_amount, amount,
			period_start, period_end, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		item.ID,
		item.InvoiceID,
		item.Description,
		item.Quantity,
		item.UnitAmount,
		item.Amount,
		item.PeriodStart,
		item.PeriodEnd,
		item.Metadata,
		item.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create line item: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetLineItems(ctx context.Context, invoiceID string) ([]*LineItem, error) {
	query := `
		SELECT id, invoice_id, description, quantity, unit_amount, amount,
		       period_start, period_end, metadata, created_at
		FROM invoice_line_items
		WHERE invoice_id = $1
		ORDER BY created_at ASC
	`

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items: %w", err)
	}
	defer rows.Close()

	items := make([]*LineItem, 0)
	for rows.Next() {
		var item LineItem
		err := rows.Scan(
			&item.ID,
			&item.InvoiceID,
			&item.Description,
			&item.Quantity,
			&item.UnitAmount,
			&item.Amount,
			&item.PeriodStart,
			&item.PeriodEnd,
			&item.Metadata,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}
		items = append(items, &item)
	}

	return items, nil
}
