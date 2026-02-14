package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gobilling/internal/event"
	"gobilling/internal/platform/database"
	"gobilling/internal/webhook"
)

type WebhookDeliveryWorker struct {
	deliveryRepo webhook.DeliveryRepository
	endpointRepo webhook.EndpointRepository
	eventRepo    event.Repository
	db           *database.DB
	httpClient   *http.Client
	interval     time.Duration
	batchSize    int
	maxRetries   int
}

func NewWebhookDeliveryWorker(
	deliveryRepo webhook.DeliveryRepository,
	endpointRepo webhook.EndpointRepository,
	eventRepo event.Repository,
	db *database.DB,
) *WebhookDeliveryWorker {
	return &WebhookDeliveryWorker{
		deliveryRepo: deliveryRepo,
		endpointRepo: endpointRepo,
		eventRepo:    eventRepo,
		db:           db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		interval:   1 * time.Second,
		batchSize:  10,
		maxRetries: 5,
	}
}

func (w *WebhookDeliveryWorker) Start(ctx context.Context) error {
	slog.Info("starting webhook delivery worker", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("webhook delivery worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				slog.Error("webhook delivery worker error", "error", err)
			}
		}
	}
}

func (w *WebhookDeliveryWorker) processBatch(ctx context.Context) error {
	deliveries, err := w.deliveryRepo.GetPending(ctx, w.batchSize)
	if err != nil {
		return err
	}

	if len(deliveries) == 0 {
		return nil
	}

	slog.Debug("processing webhook deliveries", "count", len(deliveries))

	for _, delivery := range deliveries {
		if err := w.processDelivery(ctx, delivery); err != nil {
			slog.Error("failed to process delivery", "delivery_id", delivery.ID, "error", err)
		}
	}

	return nil
}

func (w *WebhookDeliveryWorker) processDelivery(ctx context.Context, delivery *webhook.WebhookDelivery) error {
	endpoint, err := w.endpointRepo.GetByID(ctx, delivery.WebhookEndpointID)
	if err != nil {
		return err
	}

	if !endpoint.Active {
		delivery.Status = webhook.DeliveryStatusSkipped
		return w.deliveryRepo.Update(ctx, delivery)
	}

	evt, err := w.eventRepo.GetByID(ctx, delivery.EventID)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return err
	}

	timestamp := time.Now().Unix()
	signature := webhook.GenerateSignature(endpoint.Secret, timestamp, payload)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", fmt.Sprintf("sha256=%s", signature))
	req.Header.Set("X-Webhook-ID", delivery.ID)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", timestamp))

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return w.handleDeliveryError(ctx, delivery, 0, err.Error())
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.MarkDelivered(resp.StatusCode, bodyStr)
		return w.deliveryRepo.Update(ctx, delivery)
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		delivery.MarkFailed(resp.StatusCode, bodyStr)
		return w.deliveryRepo.Update(ctx, delivery)
	}

	return w.handleDeliveryError(ctx, delivery, resp.StatusCode, bodyStr)
}

func (w *WebhookDeliveryWorker) handleDeliveryError(ctx context.Context, delivery *webhook.WebhookDelivery, statusCode int, body string) error {
	if delivery.AttemptCount >= w.maxRetries {
		delivery.MarkFailed(statusCode, body)
		slog.Error("webhook delivery failed after max retries", "delivery_id", delivery.ID)
		return w.deliveryRepo.Update(ctx, delivery)
	}

	delay := webhook.GetRetryDelay(delivery.AttemptCount + 1)
	if delay == 0 {
		delivery.MarkFailed(statusCode, body)
		return w.deliveryRepo.Update(ctx, delivery)
	}

	delivery.ScheduleRetry(delay)
	slog.Info("webhook delivery retry scheduled",
		"delivery_id", delivery.ID,
		"attempt", delivery.AttemptCount,
		"next_attempt", delivery.NextAttemptAt)

	return w.deliveryRepo.Update(ctx, delivery)
}
