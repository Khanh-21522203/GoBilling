package customer

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

func (r *PostgresRepository) Create(ctx context.Context, customer *Customer) error {
	query := `
		INSERT INTO customers (
			id, email, name, external_id, status, metadata, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		customer.ID,
		customer.Email,
		customer.Name,
		customer.ExternalID,
		customer.Status,
		customer.Metadata,
		customer.Version,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Customer, error) {
	query := `
		SELECT id, email, name, external_id, status, metadata, version, created_at, updated_at, deleted_at
		FROM customers
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var customer Customer
	err := row.Scan(
		&customer.ID,
		&customer.Email,
		&customer.Name,
		&customer.ExternalID,
		&customer.Status,
		&customer.Metadata,
		&customer.Version,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*Customer, error) {
	query := `
		SELECT id, email, name, external_id, status, metadata, version, created_at, updated_at, deleted_at
		FROM customers
		WHERE email = $1 AND deleted_at IS NULL
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, email)

	var customer Customer
	err := row.Scan(
		&customer.ID,
		&customer.Email,
		&customer.Name,
		&customer.ExternalID,
		&customer.Status,
		&customer.Metadata,
		&customer.Version,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&customer.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return &customer, nil
}

func (r *PostgresRepository) Update(ctx context.Context, customer *Customer) error {
	query := `
		UPDATE customers
		SET email = $2, name = $3, external_id = $4, status = $5, 
		    metadata = $6, version = version + 1, updated_at = $7
		WHERE id = $1 AND version = $8
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		customer.ID,
		customer.Email,
		customer.Name,
		customer.ExternalID,
		customer.Status,
		customer.Metadata,
		customer.UpdatedAt,
		customer.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	customer.Version++
	return nil
}

func (r *PostgresRepository) List(ctx context.Context, opts ListOptions) ([]*Customer, error) {
	query := `
		SELECT id, email, name, external_id, status, metadata, version, created_at, updated_at, deleted_at
		FROM customers
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argPos := 1

	if opts.Email != nil {
		query += fmt.Sprintf(" AND email = $%d", argPos)
		args = append(args, *opts.Email)
		argPos++
	}

	if opts.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *opts.Status)
		argPos++
	}

	if opts.ExternalID != nil {
		query += fmt.Sprintf(" AND external_id = $%d", argPos)
		args = append(args, *opts.ExternalID)
		argPos++
	}

	if opts.StartingAfter != "" {
		query += fmt.Sprintf(" AND id > $%d", argPos)
		args = append(args, opts.StartingAfter)
		argPos++
	}

	query += " ORDER BY id ASC"
	query += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, opts.Limit+1)

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	customers := make([]*Customer, 0, opts.Limit)
	for rows.Next() {
		var customer Customer
		err := rows.Scan(
			&customer.ID,
			&customer.Email,
			&customer.Name,
			&customer.ExternalID,
			&customer.Status,
			&customer.Metadata,
			&customer.Version,
			&customer.CreatedAt,
			&customer.UpdatedAt,
			&customer.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, &customer)
	}

	return customers, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE customers
		SET status = $2, deleted_at = $3, updated_at = $3
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query, id, StatusDeleted, ctx.Value("now"))

	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) HasActiveSubscriptions(ctx context.Context, customerID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM subscriptions
			WHERE customer_id = $1 AND status IN ('trialing', 'active', 'past_due')
		)
	`

	q := database.GetQuerier(ctx, r.db)
	var exists bool
	err := q.QueryRow(ctx, query, customerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active subscriptions: %w", err)
	}

	return exists, nil
}
