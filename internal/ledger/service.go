package ledger

import (
	"context"
	"fmt"

	"gobilling/internal/pkg/id"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RecordCharge(ctx context.Context, customerID, invoiceID string, amount int64, currency, description string) error {
	entry := NewChargeEntry(customerID, invoiceID, amount, currency, description)
	entry.ID = id.NewWithPrefix("ltr_")
	return s.repo.Create(ctx, entry)
}

func (s *Service) RecordPayment(ctx context.Context, customerID, paymentID string, amount int64, currency, description string) error {
	entry := NewPaymentEntry(customerID, paymentID, amount, currency, description)
	entry.ID = id.NewWithPrefix("ltr_")
	return s.repo.Create(ctx, entry)
}

func (s *Service) RecordRefund(ctx context.Context, customerID, refundID string, amount int64, currency, description string) error {
	entry := NewRefundEntry(customerID, refundID, amount, currency, description)
	entry.ID = id.NewWithPrefix("ltr_")
	return s.repo.Create(ctx, entry)
}

func (s *Service) GetBalance(ctx context.Context, customerID string) (int64, error) {
	balance, err := s.repo.GetBalance(ctx, customerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance, nil
}

func (s *Service) GetTransactions(ctx context.Context, customerID string, limit int) ([]*LedgerTransaction, error) {
	return s.repo.GetByCustomerID(ctx, customerID, limit)
}
