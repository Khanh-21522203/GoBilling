package subscription

import (
	"context"
	"fmt"
	"time"

	"gobilling/internal/customer"
	"gobilling/internal/invoice"
	"gobilling/internal/payment"
	"gobilling/internal/pkg/clock"
	"gobilling/internal/pkg/id"
	"gobilling/internal/platform/database"
	"gobilling/internal/product"
)

type Service struct {
	repo            Repository
	customerRepo    customer.Repository
	planRepo        product.PlanRepository
	invoiceRepo     invoice.Repository
	paymentProvider payment.Provider
	db              *database.DB
	clock           clock.Clock
}

func NewService(
	repo Repository,
	customerRepo customer.Repository,
	planRepo product.PlanRepository,
	invoiceRepo invoice.Repository,
	paymentProvider payment.Provider,
	db *database.DB,
	clock clock.Clock,
) *Service {
	return &Service{
		repo:            repo,
		customerRepo:    customerRepo,
		planRepo:        planRepo,
		invoiceRepo:     invoiceRepo,
		paymentProvider: paymentProvider,
		db:              db,
		clock:           clock,
	}
}

type CreateRequest struct {
	CustomerID string
	PlanID     string
	Quantity   int
	Metadata   map[string]string
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Subscription, error) {
	cust, err := s.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, err
	}

	if !cust.IsActive() {
		return nil, ErrCustomerNotActive
	}

	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	if !plan.IsActive() {
		return nil, ErrPlanNotActive
	}

	now := s.clock.Now()
	var periodStart, periodEnd time.Time
	var trialStart, trialEnd *time.Time

	if plan.TrialPeriodDays > 0 {
		periodStart = now
		periodEnd = now.AddDate(0, 0, plan.TrialPeriodDays)
		trialStart = &now
		trialEnd = &periodEnd
	} else {
		periodStart = now
		if plan.BillingInterval == product.BillingIntervalMonthly {
			periodEnd = now.AddDate(0, plan.BillingIntervalCount, 0)
		} else {
			periodEnd = now.AddDate(plan.BillingIntervalCount, 0, 0)
		}
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	sub := &Subscription{
		ID:                 id.NewWithPrefix("sub_"),
		CustomerID:         req.CustomerID,
		PlanID:             req.PlanID,
		Status:             StatusTrialing,
		Quantity:           quantity,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		TrialStart:         trialStart,
		TrialEnd:           trialEnd,
		CancelAtPeriodEnd:  false,
		Metadata:           req.Metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if sub.Metadata == nil {
		sub.Metadata = make(map[string]string)
	}

	if plan.TrialPeriodDays == 0 {
		sub.Status = StatusActive
	}

	var createdInvoice *invoice.Invoice

	err = s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, sub); err != nil {
			return err
		}

		if plan.TrialPeriodDays == 0 && plan.Amount > 0 {
			inv, err := s.generateInvoice(txCtx, sub, plan)
			if err != nil {
				return err
			}
			createdInvoice = inv
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	if createdInvoice != nil && createdInvoice.Total > 0 {
		_ = createdInvoice
	}

	return sub, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Subscription, error) {
	return s.repo.List(ctx, opts)
}

type CancelRequest struct {
	CancelAtPeriodEnd bool
}

func (s *Service) Cancel(ctx context.Context, id string, req CancelRequest) (*Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if sub.IsCanceled() {
		return nil, ErrAlreadyCanceled
	}

	if req.CancelAtPeriodEnd {
		if err := sub.ScheduleCancellation(); err != nil {
			return nil, err
		}
	} else {
		if err := sub.CancelImmediately(); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	return sub, nil
}

type UpdatePlanRequest struct {
	NewPlanID string
}

func (s *Service) UpdatePlan(ctx context.Context, id string, req UpdatePlanRequest) (*Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !sub.CanBeModified() {
		return nil, ErrCannotModify
	}

	newPlan, err := s.planRepo.GetByID(ctx, req.NewPlanID)
	if err != nil {
		return nil, err
	}

	if !newPlan.IsActive() {
		return nil, ErrPlanNotActive
	}

	oldPlan, err := s.planRepo.GetByID(ctx, sub.PlanID)
	if err != nil {
		return nil, err
	}

	if newPlan.Currency != oldPlan.Currency {
		return nil, fmt.Errorf("currency mismatch")
	}

	if err := sub.ChangePlan(req.NewPlanID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	return sub, nil
}

func (s *Service) generateInvoice(ctx context.Context, sub *Subscription, plan *product.Plan) (*invoice.Invoice, error) {
	now := s.clock.Now()

	inv := &invoice.Invoice{
		ID:             id.NewWithPrefix("inv_"),
		InvoiceNumber:  fmt.Sprintf("INV-%d", now.Unix()),
		CustomerID:     sub.CustomerID,
		SubscriptionID: &sub.ID,
		Status:         invoice.StatusDraft,
		Currency:       plan.Currency,
		Subtotal:       0,
		DiscountAmount: 0,
		TaxAmount:      0,
		Total:          0,
		AmountPaid:     0,
		AmountDue:      0,
		PeriodStart:    &sub.CurrentPeriodStart,
		PeriodEnd:      &sub.CurrentPeriodEnd,
		Metadata:       make(map[string]string),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.invoiceRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	lineItem := &invoice.LineItem{
		ID:          id.NewWithPrefix("li_"),
		InvoiceID:   inv.ID,
		Description: fmt.Sprintf("%s x %d", plan.Name, sub.Quantity),
		Quantity:    int64(sub.Quantity),
		UnitAmount:  plan.Amount,
		Amount:      plan.Amount * int64(sub.Quantity),
		PeriodStart: &sub.CurrentPeriodStart,
		PeriodEnd:   &sub.CurrentPeriodEnd,
		Metadata:    make(map[string]string),
		CreatedAt:   now,
	}

	if err := s.invoiceRepo.CreateLineItem(ctx, lineItem); err != nil {
		return nil, err
	}

	inv.AddLineItem(lineItem)

	if err := inv.Finalize(); err != nil {
		return nil, err
	}

	if err := s.invoiceRepo.Update(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}
