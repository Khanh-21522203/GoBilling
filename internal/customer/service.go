package customer

import (
	"context"
	"fmt"

	"gobilling/internal/pkg/clock"
	"gobilling/internal/pkg/id"
	"gobilling/internal/platform/database"
)

type Service struct {
	repo  Repository
	db    *database.DB
	clock clock.Clock
}

func NewService(repo Repository, db *database.DB, clock clock.Clock) *Service {
	return &Service{
		repo:  repo,
		db:    db,
		clock: clock,
	}
}

type CreateRequest struct {
	Email      string
	Name       string
	ExternalID *string
	Metadata   map[string]string
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Customer, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	customer := &Customer{
		ID:         id.NewWithPrefix("cus_"),
		Email:      req.Email,
		Name:       req.Name,
		ExternalID: req.ExternalID,
		Status:     StatusActive,
		Metadata:   req.Metadata,
		Version:    1,
		CreatedAt:  s.clock.Now(),
		UpdatedAt:  s.clock.Now(),
	}

	if customer.Metadata == nil {
		customer.Metadata = make(map[string]string)
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	return customer, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

type UpdateRequest struct {
	Name       *string
	Email      *string
	ExternalID *string
	Metadata   map[string]string
}

func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if customer.IsDeleted() {
		return nil, ErrDeleted
	}

	if !customer.CanBeModified() {
		return nil, ErrInvalidStatusTransition
	}

	if req.Email != nil && *req.Email != customer.Email {
		if err := validateEmail(*req.Email); err != nil {
			return nil, err
		}
		existing, err := s.repo.GetByEmail(ctx, *req.Email)
		if err == nil && existing != nil && existing.ID != customer.ID {
			return nil, ErrEmailAlreadyExists
		}
		customer.Email = *req.Email
	}

	if req.Name != nil {
		customer.Name = *req.Name
	}

	if req.ExternalID != nil {
		customer.ExternalID = req.ExternalID
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			customer.Metadata[k] = v
		}
	}

	customer.UpdatedAt = s.clock.Now()

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

func (s *Service) List(ctx context.Context, opts ListOptions) ([]*Customer, error) {
	return s.repo.List(ctx, opts)
}

func (s *Service) Suspend(ctx context.Context, id string) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := customer.Suspend(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to suspend customer: %w", err)
	}

	return customer, nil
}

func (s *Service) Reactivate(ctx context.Context, id string) (*Customer, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := customer.Reactivate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to reactivate customer: %w", err)
	}

	return customer, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	hasActive, err := s.repo.HasActiveSubscriptions(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check subscriptions: %w", err)
	}

	if hasActive {
		return ErrHasActiveSubscriptions
	}

	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := customer.Delete(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	return nil
}
