package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gobilling/internal/platform/database"
)

type PostgresProductRepository struct {
	db *database.DB
}

func NewPostgresProductRepository(db *database.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product *Product) error {
	query := `
		INSERT INTO products (id, name, description, active, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Active,
		product.Metadata,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id string) (*Product, error) {
	query := `
		SELECT id, name, description, active, metadata, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var product Product
	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Active,
		&product.Metadata,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, product *Product) error {
	query := `
		UPDATE products
		SET name = $2, description = $3, active = $4, metadata = $5, updated_at = $6
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Active,
		product.Metadata,
		product.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *PostgresProductRepository) List(ctx context.Context, opts ProductListOptions) ([]*Product, error) {
	query := `
		SELECT id, name, description, active, metadata, created_at, updated_at
		FROM products
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if opts.Active != nil {
		query += fmt.Sprintf(" AND active = $%d", argPos)
		args = append(args, *opts.Active)
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
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := make([]*Product, 0)
	for rows.Next() {
		var product Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Active,
			&product.Metadata,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	return products, nil
}

type PostgresPlanRepository struct {
	db *database.DB
}

func NewPostgresPlanRepository(db *database.DB) *PostgresPlanRepository {
	return &PostgresPlanRepository{db: db}
}

func (r *PostgresPlanRepository) Create(ctx context.Context, plan *Plan) error {
	tiersJSON, err := json.Marshal(plan.Tiers)
	if err != nil {
		return fmt.Errorf("failed to marshal tiers: %w", err)
	}

	query := `
		INSERT INTO plans (
			id, product_id, name, description, pricing_type, amount, currency,
			billing_interval, billing_interval_count, trial_period_days, tiers,
			active, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err = q.Exec(ctx, query,
		plan.ID,
		plan.ProductID,
		plan.Name,
		plan.Description,
		plan.PricingType,
		plan.Amount,
		plan.Currency,
		plan.BillingInterval,
		plan.BillingIntervalCount,
		plan.TrialPeriodDays,
		tiersJSON,
		plan.Active,
		plan.Metadata,
		plan.CreatedAt,
		plan.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	return nil
}

func (r *PostgresPlanRepository) GetByID(ctx context.Context, id string) (*Plan, error) {
	query := `
		SELECT id, product_id, name, description, pricing_type, amount, currency,
		       billing_interval, billing_interval_count, trial_period_days, tiers,
		       active, metadata, created_at, updated_at
		FROM plans
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var plan Plan
	var tiersJSON []byte
	err := row.Scan(
		&plan.ID,
		&plan.ProductID,
		&plan.Name,
		&plan.Description,
		&plan.PricingType,
		&plan.Amount,
		&plan.Currency,
		&plan.BillingInterval,
		&plan.BillingIntervalCount,
		&plan.TrialPeriodDays,
		&tiersJSON,
		&plan.Active,
		&plan.Metadata,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	if tiersJSON != nil {
		if err := json.Unmarshal(tiersJSON, &plan.Tiers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tiers: %w", err)
		}
	}

	return &plan, nil
}

func (r *PostgresPlanRepository) Update(ctx context.Context, plan *Plan) error {
	tiersJSON, err := json.Marshal(plan.Tiers)
	if err != nil {
		return fmt.Errorf("failed to marshal tiers: %w", err)
	}

	query := `
		UPDATE plans
		SET name = $2, description = $3, active = $4, metadata = $5, 
		    tiers = $6, updated_at = $7
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		plan.ID,
		plan.Name,
		plan.Description,
		plan.Active,
		plan.Metadata,
		tiersJSON,
		plan.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrPlanNotFound
	}

	return nil
}

func (r *PostgresPlanRepository) List(ctx context.Context, opts PlanListOptions) ([]*Plan, error) {
	query := `
		SELECT id, product_id, name, description, pricing_type, amount, currency,
		       billing_interval, billing_interval_count, trial_period_days, tiers,
		       active, metadata, created_at, updated_at
		FROM plans
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if opts.ProductID != nil {
		query += fmt.Sprintf(" AND product_id = $%d", argPos)
		args = append(args, *opts.ProductID)
		argPos++
	}

	if opts.Active != nil {
		query += fmt.Sprintf(" AND active = $%d", argPos)
		args = append(args, *opts.Active)
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
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	plans := make([]*Plan, 0)
	for rows.Next() {
		var plan Plan
		var tiersJSON []byte
		err := rows.Scan(
			&plan.ID,
			&plan.ProductID,
			&plan.Name,
			&plan.Description,
			&plan.PricingType,
			&plan.Amount,
			&plan.Currency,
			&plan.BillingInterval,
			&plan.BillingIntervalCount,
			&plan.TrialPeriodDays,
			&tiersJSON,
			&plan.Active,
			&plan.Metadata,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		if tiersJSON != nil {
			if err := json.Unmarshal(tiersJSON, &plan.Tiers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tiers: %w", err)
			}
		}

		plans = append(plans, &plan)
	}

	return plans, nil
}

func (r *PostgresPlanRepository) ListByProductID(ctx context.Context, productID string) ([]*Plan, error) {
	return r.List(ctx, PlanListOptions{
		ProductID: &productID,
		Limit:     100,
	})
}
