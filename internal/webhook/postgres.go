package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gobilling/internal/platform/database"
)

type PostgresEndpointRepository struct {
	db *database.DB
}

func NewPostgresEndpointRepository(db *database.DB) *PostgresEndpointRepository {
	return &PostgresEndpointRepository{db: db}
}

func (r *PostgresEndpointRepository) Create(ctx context.Context, endpoint *WebhookEndpoint) error {
	eventsJSON, err := json.Marshal(endpoint.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	query := `
		INSERT INTO webhook_endpoints (id, url, secret, events, active, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err = q.Exec(ctx, query,
		endpoint.ID,
		endpoint.URL,
		endpoint.Secret,
		eventsJSON,
		endpoint.Active,
		endpoint.Metadata,
		endpoint.CreatedAt,
		endpoint.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create webhook endpoint: %w", err)
	}

	return nil
}

func (r *PostgresEndpointRepository) GetByID(ctx context.Context, id string) (*WebhookEndpoint, error) {
	query := `
		SELECT id, url, secret, events, active, metadata, created_at, updated_at
		FROM webhook_endpoints
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var endpoint WebhookEndpoint
	var eventsJSON []byte

	err := row.Scan(
		&endpoint.ID,
		&endpoint.URL,
		&endpoint.Secret,
		&eventsJSON,
		&endpoint.Active,
		&endpoint.Metadata,
		&endpoint.CreatedAt,
		&endpoint.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("webhook endpoint not found")
		}
		return nil, fmt.Errorf("failed to get webhook endpoint: %w", err)
	}

	if err := json.Unmarshal(eventsJSON, &endpoint.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return &endpoint, nil
}

func (r *PostgresEndpointRepository) Update(ctx context.Context, endpoint *WebhookEndpoint) error {
	eventsJSON, err := json.Marshal(endpoint.Events)
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	query := `
		UPDATE webhook_endpoints
		SET url = $2, events = $3, active = $4, metadata = $5, updated_at = $6
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		endpoint.ID,
		endpoint.URL,
		eventsJSON,
		endpoint.Active,
		endpoint.Metadata,
		endpoint.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook endpoint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook endpoint not found")
	}

	return nil
}

func (r *PostgresEndpointRepository) List(ctx context.Context, active *bool, limit int) ([]*WebhookEndpoint, error) {
	query := `
		SELECT id, url, secret, events, active, metadata, created_at, updated_at
		FROM webhook_endpoints
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if active != nil {
		query += fmt.Sprintf(" AND active = $%d", argPos)
		args = append(args, *active)
		argPos++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, limit)

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook endpoints: %w", err)
	}
	defer rows.Close()

	endpoints := make([]*WebhookEndpoint, 0)
	for rows.Next() {
		var endpoint WebhookEndpoint
		var eventsJSON []byte

		err := rows.Scan(
			&endpoint.ID,
			&endpoint.URL,
			&endpoint.Secret,
			&eventsJSON,
			&endpoint.Active,
			&endpoint.Metadata,
			&endpoint.CreatedAt,
			&endpoint.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook endpoint: %w", err)
		}

		if err := json.Unmarshal(eventsJSON, &endpoint.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		endpoints = append(endpoints, &endpoint)
	}

	return endpoints, nil
}

func (r *PostgresEndpointRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM webhook_endpoints WHERE id = $1`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook endpoint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook endpoint not found")
	}

	return nil
}

type PostgresDeliveryRepository struct {
	db *database.DB
}

func NewPostgresDeliveryRepository(db *database.DB) *PostgresDeliveryRepository {
	return &PostgresDeliveryRepository{db: db}
}

func (r *PostgresDeliveryRepository) Create(ctx context.Context, delivery *WebhookDelivery) error {
	query := `
		INSERT INTO webhook_deliveries (
			id, webhook_endpoint_id, event_id, status, response_code, response_body,
			attempt_count, next_attempt_at, delivered_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err := q.Exec(ctx, query,
		delivery.ID,
		delivery.WebhookEndpointID,
		delivery.EventID,
		delivery.Status,
		delivery.ResponseCode,
		delivery.ResponseBody,
		delivery.AttemptCount,
		delivery.NextAttemptAt,
		delivery.DeliveredAt,
		delivery.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create webhook delivery: %w", err)
	}

	return nil
}

func (r *PostgresDeliveryRepository) GetByID(ctx context.Context, id string) (*WebhookDelivery, error) {
	query := `
		SELECT id, webhook_endpoint_id, event_id, status, response_code, response_body,
		       attempt_count, next_attempt_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var delivery WebhookDelivery
	err := row.Scan(
		&delivery.ID,
		&delivery.WebhookEndpointID,
		&delivery.EventID,
		&delivery.Status,
		&delivery.ResponseCode,
		&delivery.ResponseBody,
		&delivery.AttemptCount,
		&delivery.NextAttemptAt,
		&delivery.DeliveredAt,
		&delivery.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("webhook delivery not found")
		}
		return nil, fmt.Errorf("failed to get webhook delivery: %w", err)
	}

	return &delivery, nil
}

func (r *PostgresDeliveryRepository) Update(ctx context.Context, delivery *WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries
		SET status = $2, response_code = $3, response_body = $4,
		    attempt_count = $5, next_attempt_at = $6, delivered_at = $7
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		delivery.ID,
		delivery.Status,
		delivery.ResponseCode,
		delivery.ResponseBody,
		delivery.AttemptCount,
		delivery.NextAttemptAt,
		delivery.DeliveredAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook delivery: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook delivery not found")
	}

	return nil
}

func (r *PostgresDeliveryRepository) GetPending(ctx context.Context, limit int) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, webhook_endpoint_id, event_id, status, response_code, response_body,
		       attempt_count, next_attempt_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE status = 'pending' AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deliveries: %w", err)
	}
	defer rows.Close()

	deliveries := make([]*WebhookDelivery, 0)
	for rows.Next() {
		var delivery WebhookDelivery
		err := rows.Scan(
			&delivery.ID,
			&delivery.WebhookEndpointID,
			&delivery.EventID,
			&delivery.Status,
			&delivery.ResponseCode,
			&delivery.ResponseBody,
			&delivery.AttemptCount,
			&delivery.NextAttemptAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}

func (r *PostgresDeliveryRepository) ListByEndpoint(ctx context.Context, endpointID string, limit int) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, webhook_endpoint_id, event_id, status, response_code, response_body,
		       attempt_count, next_attempt_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE webhook_endpoint_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, endpointID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list deliveries: %w", err)
	}
	defer rows.Close()

	deliveries := make([]*WebhookDelivery, 0)
	for rows.Next() {
		var delivery WebhookDelivery
		err := rows.Scan(
			&delivery.ID,
			&delivery.WebhookEndpointID,
			&delivery.EventID,
			&delivery.Status,
			&delivery.ResponseCode,
			&delivery.ResponseBody,
			&delivery.AttemptCount,
			&delivery.NextAttemptAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}
