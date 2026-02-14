package event

import (
	"context"
	"time"

	"gobilling/internal/pkg/id"
)

type Publisher interface {
	Publish(ctx context.Context, eventType string, payload map[string]interface{}) error
}

type OutboxPublisher struct {
	repo Repository
}

func NewOutboxPublisher(repo Repository) *OutboxPublisher {
	return &OutboxPublisher{repo: repo}
}

func (p *OutboxPublisher) Publish(ctx context.Context, eventType string, payload map[string]interface{}) error {
	event := &Event{
		ID:         id.NewWithPrefix("evt_"),
		Type:       eventType,
		Payload:    payload,
		Status:     StatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now().UTC(),
	}

	return p.repo.Create(ctx, event)
}
