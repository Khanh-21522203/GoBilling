package worker

import (
	"context"
	"log/slog"
	"time"

	"gobilling/internal/event"
	"gobilling/internal/platform/database"
)

type OutboxWorker struct {
	eventRepo event.Repository
	db        *database.DB
	interval  time.Duration
	batchSize int
	maxRetries int
}

func NewOutboxWorker(eventRepo event.Repository, db *database.DB) *OutboxWorker {
	return &OutboxWorker{
		eventRepo:  eventRepo,
		db:         db,
		interval:   500 * time.Millisecond,
		batchSize:  10,
		maxRetries: 5,
	}
}

func (w *OutboxWorker) Start(ctx context.Context) error {
	slog.Info("starting outbox worker", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("outbox worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				slog.Error("outbox worker error", "error", err)
			}
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) error {
	events, err := w.eventRepo.GetPending(ctx, w.batchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	slog.Debug("processing events", "count", len(events))

	for _, evt := range events {
		if err := w.processEvent(ctx, evt); err != nil {
			slog.Error("failed to process event", "event_id", evt.ID, "error", err)
		}
	}

	return nil
}

func (w *OutboxWorker) processEvent(ctx context.Context, evt *event.Event) error {
	err := w.db.WithTransaction(ctx, func(txCtx context.Context) error {
		slog.Info("dispatching event", "event_id", evt.ID, "type", evt.Type)

		evt.MarkDelivered()
		return w.eventRepo.Update(txCtx, evt)
	})

	if err != nil {
		if evt.RetryCount >= w.maxRetries {
			evt.MarkFailed()
			_ = w.eventRepo.Update(ctx, evt)
			slog.Error("event failed after max retries", "event_id", evt.ID)
			return err
		}

		delay := time.Duration(1<<uint(evt.RetryCount)) * time.Minute
		evt.ScheduleRetry(delay)
		_ = w.eventRepo.Update(ctx, evt)
		return err
	}

	return nil
}
