package worker

import (
	"context"
	"log/slog"
	"time"

	"gobilling/internal/invoice"
	"gobilling/internal/payment"
	"gobilling/internal/platform/database"
)

type PaymentRetryWorker struct {
	invoiceRepo invoice.Repository
	paymentRepo payment.Repository
	provider    payment.Provider
	db          *database.DB
	interval    time.Duration
	batchSize   int
}

func NewPaymentRetryWorker(
	invoiceRepo invoice.Repository,
	paymentRepo payment.Repository,
	provider payment.Provider,
	db *database.DB,
) *PaymentRetryWorker {
	return &PaymentRetryWorker{
		invoiceRepo: invoiceRepo,
		paymentRepo: paymentRepo,
		provider:    provider,
		db:          db,
		interval:    1 * time.Minute,
		batchSize:   10,
	}
}

func (w *PaymentRetryWorker) Start(ctx context.Context) error {
	slog.Info("starting payment retry worker", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("payment retry worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				slog.Error("payment retry worker error", "error", err)
			}
		}
	}
}

func (w *PaymentRetryWorker) processBatch(ctx context.Context) error {
	slog.Debug("checking for payment retries")
	return nil
}
