package event

import (
	"context"
	"encoding/json"
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

func (r *PostgresRepository) Create(ctx context.Context, event *Event) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO events (id, type, payload, status, retry_count, next_retry_at, delivered_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	q := database.GetQuerier(ctx, r.db)
	_, err = q.Exec(ctx, query,
		event.ID,
		event.Type,
		payloadJSON,
		event.Status,
		event.RetryCount,
		event.NextRetryAt,
		event.DeliveredAt,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, type, payload, status, retry_count, next_retry_at, delivered_at, created_at
		FROM events
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	row := q.QueryRow(ctx, query, id)

	var event Event
	var payloadJSON []byte

	err := row.Scan(
		&event.ID,
		&event.Type,
		&payloadJSON,
		&event.Status,
		&event.RetryCount,
		&event.NextRetryAt,
		&event.DeliveredAt,
		&event.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &event, nil
}

func (r *PostgresRepository) Update(ctx context.Context, event *Event) error {
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		UPDATE events
		SET status = $2, retry_count = $3, next_retry_at = $4, delivered_at = $5, payload = $6
		WHERE id = $1
	`

	q := database.GetQuerier(ctx, r.db)
	result, err := q.Exec(ctx, query,
		event.ID,
		event.Status,
		event.RetryCount,
		event.NextRetryAt,
		event.DeliveredAt,
		payloadJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

func (r *PostgresRepository) GetPending(ctx context.Context, limit int) ([]*Event, error) {
	query := `
		SELECT id, type, payload, status, retry_count, next_retry_at, delivered_at, created_at
		FROM events
		WHERE status = 'pending' AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	q := database.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	defer rows.Close()

	events := make([]*Event, 0)
	for rows.Next() {
		var event Event
		var payloadJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.Type,
			&payloadJSON,
			&event.Status,
			&event.RetryCount,
			&event.NextRetryAt,
			&event.DeliveredAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		events = append(events, &event)
	}

	return events, nil
}
