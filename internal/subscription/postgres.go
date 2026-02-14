package subscription

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

func (r *PostgresRepository) Create(ctx context.Context, subscription *Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, customer_id, plan_id, status, quantity,
			current_period_start, current_period_end,
			trial_start, trial_end, cancel_at_period_end,
			canceled_at, ended_at, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		subscription.ID,
		subscription.CustomerID,
		subscription.PlanID,
		subscription.Status,
		subscription.Quantity,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.TrialStart,
		subscription.TrialEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CanceledAt,
		subscription.EndedAt,
		subscription.Metadata,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, status, quantity,
		       current_period_start, current_period_end,
		       trial_start, trial_end, cancel_at_period_end,
		       canceled_at, ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var sub Subscription
	err := row.Scan(
		&sub.ID,
		&sub.CustomerID,
		&sub.PlanID,
		&sub.Status,
		&sub.Quantity,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.TrialStart,
		&sub.TrialEnd,
		&sub.CancelAtPeriodEnd,
		&sub.CanceledAt,
		&sub.EndedAt,
		&sub.Metadata,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &sub, nil
}

func (r *PostgresRepository) Update(ctx context.Context, subscription *Subscription) error {
	query := `
		UPDATE subscriptions
		SET customer_id = $2, plan_id = $3, status = $4, quantity = $5,
		    current_period_start = $6, current_period_end = $7,
		    trial_start = $8, trial_end = $9, cancel_at_period_end = $10,
		    canceled_at = $11, ended_at = $12, metadata = $13, updated_at = $14
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		subscription.ID,
		subscription.CustomerID,
		subscription.PlanID,
		subscription.Status,
		subscription.Quantity,
		subscription.CurrentPeriodStart,
		subscription.CurrentPeriodEnd,
		subscription.TrialStart,
		subscription.TrialEnd,
		subscription.CancelAtPeriodEnd,
		subscription.CanceledAt,
		subscription.EndedAt,
		subscription.Metadata,
		subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) List(ctx context.Context, opts ListOptions) ([]*Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, status, quantity,
		       current_period_start, current_period_end,
		       trial_start, trial_end, cancel_at_period_end,
		       canceled_at, ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if opts.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argPos)
		args = append(args, *opts.CustomerID)
		argPos++
	}

	if opts.PlanID != nil {
		query += fmt.Sprintf(" AND plan_id = $%d", argPos)
		args = append(args, *opts.PlanID)
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
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*Subscription, 0)
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.CustomerID,
			&sub.PlanID,
			&sub.Status,
			&sub.Quantity,
			&sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd,
			&sub.TrialStart,
			&sub.TrialEnd,
			&sub.CancelAtPeriodEnd,
			&sub.CanceledAt,
			&sub.EndedAt,
			&sub.Metadata,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, nil
}

func (r *PostgresRepository) GetDueForRenewal(ctx context.Context, limit int) ([]*Subscription, error) {
	query := `
		SELECT id, customer_id, plan_id, status, quantity,
		       current_period_start, current_period_end,
		       trial_start, trial_end, cancel_at_period_end,
		       canceled_at, ended_at, metadata, created_at, updated_at
		FROM subscriptions
		WHERE current_period_end <= NOW()
		  AND status IN ('active', 'past_due')
		  AND cancel_at_period_end = FALSE
		ORDER BY current_period_end ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions due for renewal: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]*Subscription, 0)
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.CustomerID,
			&sub.PlanID,
			&sub.Status,
			&sub.Quantity,
			&sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd,
			&sub.TrialStart,
			&sub.TrialEnd,
			&sub.CancelAtPeriodEnd,
			&sub.CanceledAt,
			&sub.EndedAt,
			&sub.Metadata,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, nil
}
