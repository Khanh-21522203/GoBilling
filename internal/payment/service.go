package payment

import (
	"context"
	"fmt"

	"gobilling/internal/invoice"
	"gobilling/internal/pkg/clock"
	"gobilling/internal/pkg/id"
	"gobilling/internal/platform/database"
)

type LedgerService interface {
	RecordPayment(ctx context.Context, customerID, paymentID string, amount int64, currency, description string) error
	RecordRefund(ctx context.Context, customerID, refundID string, amount int64, currency, description string) error
}

type Service struct {
	repo            Repository
	refundRepo      RefundRepository
	invoiceRepo     invoice.Repository
	provider        Provider
	ledgerService   LedgerService
	db              *database.DB
	clock           clock.Clock
}

func NewService(
	repo Repository,
	refundRepo RefundRepository,
	invoiceRepo invoice.Repository,
	provider Provider,
	ledgerService LedgerService,
	db *database.DB,
	clock clock.Clock,
) *Service {
	return &Service{
		repo:          repo,
		refundRepo:    refundRepo,
		invoiceRepo:   invoiceRepo,
		provider:      provider,
		ledgerService: ledgerService,
		db:            db,
		clock:         clock,
	}
}

type AttemptRequest struct {
	InvoiceID       string
	PaymentMethodID string
	IdempotencyKey  string
}

func (s *Service) Attempt(ctx context.Context, req AttemptRequest) (*Payment, error) {
	inv, err := s.invoiceRepo.GetByID(ctx, req.InvoiceID)
	if err != nil {
		return nil, err
	}

	if !inv.IsOpen() {
		return nil, ErrNotFound
	}

	now := s.clock.Now()
	payment := &Payment{
		ID:              id.NewWithPrefix("pay_"),
		InvoiceID:       req.InvoiceID,
		Amount:          inv.AmountDue,
		Currency:        inv.Currency,
		Status:          StatusPending,
		PaymentMethodID: &req.PaymentMethodID,
		IdempotencyKey:  req.IdempotencyKey,
		Metadata:        make(map[string]string),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	var finalPayment *Payment

	err = s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, payment); err != nil {
			return err
		}

		chargeResp, err := s.provider.Charge(txCtx, ChargeRequest{
			Amount:          payment.Amount,
			Currency:        payment.Currency,
			PaymentMethodID: req.PaymentMethodID,
			IdempotencyKey:  req.IdempotencyKey,
			Metadata:        payment.Metadata,
		})

		if err != nil {
			payment.MarkFailed("provider_error", err.Error())
			if updateErr := s.repo.Update(txCtx, payment); updateErr != nil {
				return updateErr
			}
			return err
		}

		if chargeResp.Status == "succeeded" {
			payment.MarkSucceeded(chargeResp.ProviderID)
			if err := s.repo.Update(txCtx, payment); err != nil {
				return err
			}

			inv.RecordPayment(payment.Amount)
			if err := s.invoiceRepo.Update(txCtx, inv); err != nil {
				return err
			}

			// Create ledger entry for payment
			if s.ledgerService != nil {
				if err := s.ledgerService.RecordPayment(txCtx, inv.CustomerID, payment.ID, payment.Amount, payment.Currency, "Payment received"); err != nil {
					return err
				}
			}
		} else {
			payment.MarkFailed("declined", chargeResp.Message)
			if err := s.repo.Update(txCtx, payment); err != nil {
				return err
			}
		}

		finalPayment = payment
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("payment attempt failed: %w", err)
	}

	return finalPayment, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Payment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Payment, error) {
	return s.repo.List(ctx, opts)
}

func (s *Service) Refund(ctx context.Context, paymentID string, reason string) (*Refund, error) {
	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if !payment.CanBeRefunded() {
		return nil, ErrCannotRefund
	}

	existing, err := s.refundRepo.GetByPaymentID(ctx, paymentID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyRefunded
	}

	now := s.clock.Now()
	refund := &Refund{
		ID:        id.NewWithPrefix("ref_"),
		PaymentID: paymentID,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Status:    RefundStatusPending,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if reason != "" {
		refund.Reason = &reason
	}

	var finalRefund *Refund

	err = s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.refundRepo.Create(txCtx, refund); err != nil {
			return err
		}

		if payment.ProviderID == nil {
			return fmt.Errorf("payment has no provider ID")
		}

		refundResp, err := s.provider.Refund(txCtx, ProviderRefundRequest{
			PaymentProviderID: *payment.ProviderID,
			Amount:            refund.Amount,
			Currency:          refund.Currency,
			Reason:            reason,
		})

		if err != nil {
			refund.MarkFailed()
			if updateErr := s.refundRepo.Update(txCtx, refund); updateErr != nil {
				return updateErr
			}
			return err
		}

		if refundResp.Status == "succeeded" {
			refund.MarkSucceeded(refundResp.ProviderID)
			if err := s.refundRepo.Update(txCtx, refund); err != nil {
				return err
			}

			payment.MarkRefunded()
			if err := s.repo.Update(txCtx, payment); err != nil {
				return err
			}

			// Create ledger entry for refund
			if s.ledgerService != nil {
				inv, err := s.invoiceRepo.GetByID(txCtx, payment.InvoiceID)
				if err == nil {
					if err := s.ledgerService.RecordRefund(txCtx, inv.CustomerID, refund.ID, refund.Amount, refund.Currency, "Refund processed"); err != nil {
						return err
					}
				}
			}
		} else {
			refund.MarkFailed()
			if err := s.refundRepo.Update(txCtx, refund); err != nil {
				return err
			}
		}

		finalRefund = refund
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("refund failed: %w", err)
	}

	return finalRefund, nil
}
