package product

import (
	"context"
	"fmt"

	"gobilling/internal/pkg/clock"
	"gobilling/internal/pkg/id"
	"gobilling/internal/platform/database"
)

type Service struct {
	productRepo ProductRepository
	planRepo    PlanRepository
	db          *database.DB
	clock       clock.Clock
}

func NewService(productRepo ProductRepository, planRepo PlanRepository, db *database.DB, clock clock.Clock) *Service {
	return &Service{
		productRepo: productRepo,
		planRepo:    planRepo,
		db:          db,
		clock:       clock,
	}
}

func (s *Service) CreateProduct(ctx context.Context, name string, description *string, metadata map[string]string) (*Product, error) {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	product := &Product{
		ID:          id.NewWithPrefix("prod_"),
		Name:        name,
		Description: description,
		Active:      true,
		Metadata:    metadata,
		CreatedAt:   s.clock.Now(),
		UpdatedAt:   s.clock.Now(),
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *Service) GetProduct(ctx context.Context, id string) (*Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

type UpdateProductRequest struct {
	Name        *string
	Description *string
	Metadata    map[string]string
}

func (s *Service) UpdateProduct(ctx context.Context, id string, req UpdateProductRequest) (*Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		product.Name = *req.Name
	}

	if req.Description != nil {
		product.Description = req.Description
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			product.Metadata[k] = v
		}
	}

	product.UpdatedAt = s.clock.Now()

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (s *Service) ListProducts(ctx context.Context, opts ProductListOptions) ([]*Product, error) {
	return s.productRepo.List(ctx, opts)
}

func (s *Service) ArchiveProduct(ctx context.Context, id string) (*Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Archive()

	plans, err := s.planRepo.ListByProductID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}

	err = s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.productRepo.Update(txCtx, product); err != nil {
			return err
		}

		for _, plan := range plans {
			plan.Archive()
			if err := s.planRepo.Update(txCtx, plan); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to archive product: %w", err)
	}

	return product, nil
}

func (s *Service) CreatePlan(ctx context.Context, productID, name string, description *string, pricingType PricingType, amount int64, currency string, billingInterval BillingInterval, billingIntervalCount, trialPeriodDays int, tiers []PricingTier, metadata map[string]string) (*Plan, error) {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if !product.IsActive() {
		return nil, ErrProductNotFound
	}

	plan := &Plan{
		ID:                   id.NewWithPrefix("plan_"),
		ProductID:            productID,
		Name:                 name,
		Description:          description,
		PricingType:          pricingType,
		Amount:               amount,
		Currency:             currency,
		BillingInterval:      billingInterval,
		BillingIntervalCount: billingIntervalCount,
		TrialPeriodDays:      trialPeriodDays,
		Tiers:                tiers,
		Active:               true,
		Metadata:             metadata,
		CreatedAt:            s.clock.Now(),
		UpdatedAt:            s.clock.Now(),
	}

	if plan.Metadata == nil {
		plan.Metadata = make(map[string]string)
	}

	if err := plan.ValidatePricing(); err != nil {
		return nil, err
	}

	if err := s.planRepo.Create(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	return plan, nil
}

func (s *Service) GetPlan(ctx context.Context, id string) (*Plan, error) {
	return s.planRepo.GetByID(ctx, id)
}

func (s *Service) ListPlans(ctx context.Context, opts PlanListOptions) ([]*Plan, error) {
	return s.planRepo.List(ctx, opts)
}

func (s *Service) ArchivePlan(ctx context.Context, id string) (*Plan, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	plan.Archive()

	if err := s.planRepo.Update(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to archive plan: %w", err)
	}

	return plan, nil
}
