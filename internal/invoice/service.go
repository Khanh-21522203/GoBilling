package invoice

import (
	"context"
	"fmt"

	"gobilling/internal/pkg/clock"
)

type Service struct {
	repo  Repository
	clock clock.Clock
}

func NewService(repo Repository, clock clock.Clock) *Service {
	return &Service{
		repo:  repo,
		clock: clock,
	}
}

func (s *Service) GetByID(ctx context.Context, invoiceID string) (*Invoice, error) {
	return s.repo.GetByIDWithLineItems(ctx, invoiceID)
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Invoice, error) {
	return s.repo.List(ctx, opts)
}

func (s *Service) Finalize(ctx context.Context, invoiceID string) (*Invoice, error) {
	inv, err := s.repo.GetByIDWithLineItems(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if err := inv.Finalize(); err != nil {
		return nil, err
	}

	if inv.Status == StatusOpen {
		dueDate := s.clock.Now().AddDate(0, 0, 30)
		inv.DueDate = &dueDate
	}

	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to finalize invoice: %w", err)
	}

	return inv, nil
}

func (s *Service) Void(ctx context.Context, invoiceID string) (*Invoice, error) {
	inv, err := s.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	if err := inv.Void(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, inv); err != nil {
		return nil, fmt.Errorf("failed to void invoice: %w", err)
	}

	return inv, nil
}

func (s *Service) MarkPaid(ctx context.Context, invoiceID string) error {
	inv, err := s.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	if err := inv.MarkPaid(); err != nil {
		return err
	}

	return s.repo.Update(ctx, inv)
}

func (s *Service) MarkUncollectible(ctx context.Context, invoiceID string) error {
	inv, err := s.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	if err := inv.MarkUncollectible(); err != nil {
		return err
	}

	return s.repo.Update(ctx, inv)
}

func (s *Service) RecordPayment(ctx context.Context, invoiceID string, amount int64) error {
	inv, err := s.repo.GetByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	inv.RecordPayment(amount)

	return s.repo.Update(ctx, inv)
}
