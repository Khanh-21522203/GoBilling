package worker

import (
	"context"
	"log/slog"
	"time"

	"gobilling/internal/invoice"
	"gobilling/internal/payment"
	"gobilling/internal/platform/database"
	"gobilling/internal/product"
	"gobilling/internal/subscription"
)

type RenewalWorker struct {
	subRepo     subscription.Repository
	planRepo    product.PlanRepository
	invoiceRepo invoice.Repository
	paymentRepo payment.Repository
	provider    payment.Provider
	db          *database.DB
	interval    time.Duration
	batchSize   int
}

func NewRenewalWorker(
	subRepo subscription.Repository,
	planRepo product.PlanRepository,
	invoiceRepo invoice.Repository,
	paymentRepo payment.Repository,
	provider payment.Provider,
	db *database.DB,
) *RenewalWorker {
	return &RenewalWorker{
		subRepo:     subRepo,
		planRepo:    planRepo,
		invoiceRepo: invoiceRepo,
		paymentRepo: paymentRepo,
		provider:    provider,
		db:          db,
		interval:    1 * time.Minute,
		batchSize:   10,
	}
}

func (w *RenewalWorker) Start(ctx context.Context) error {
	slog.Info("starting renewal worker", "interval", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("renewal worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				slog.Error("renewal worker error", "error", err)
			}
		}
	}
}

func (w *RenewalWorker) processBatch(ctx context.Context) error {
	subscriptions, err := w.subRepo.GetDueForRenewal(ctx, w.batchSize)
	if err != nil {
		return err
	}

	if len(subscriptions) == 0 {
		return nil
	}

	slog.Info("processing renewals", "count", len(subscriptions))

	for _, sub := range subscriptions {
		if err := w.processRenewal(ctx, sub); err != nil {
			slog.Error("failed to process renewal", "subscription_id", sub.ID, "error", err)
		}
	}

	return nil
}

func (w *RenewalWorker) processRenewal(ctx context.Context, sub *subscription.Subscription) error {
	if sub.CancelAtPeriodEnd {
		sub.Status = subscription.StatusCanceled
		now := time.Now().UTC()
		sub.EndedAt = &now
		return w.subRepo.Update(ctx, sub)
	}

	plan, err := w.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return err
	}

	var newPeriodEnd time.Time
	if plan.BillingInterval == product.BillingIntervalMonthly {
		newPeriodEnd = sub.CurrentPeriodEnd.AddDate(0, plan.BillingIntervalCount, 0)
	} else {
		newPeriodEnd = sub.CurrentPeriodEnd.AddDate(plan.BillingIntervalCount, 0, 0)
	}

	sub.ExtendPeriod(newPeriodEnd)

	if err := sub.Activate(); err != nil {
		return err
	}

	return w.subRepo.Update(ctx, sub)
}
