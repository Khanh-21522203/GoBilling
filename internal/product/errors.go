package product

import "gobilling/internal/platform/errors"

var (
	ErrProductNotFound = errors.New("PRODUCT_NOT_FOUND", "product not found")
	ErrPlanNotFound    = errors.New("PLAN_NOT_FOUND", "plan not found")
	ErrPlanInactive    = errors.New("PLAN_INACTIVE", "plan is not active")
	ErrInvalidAmount   = errors.New("INVALID_AMOUNT", "amount must be non-negative")
	ErrMissingTiers    = errors.New("MISSING_TIERS", "tiered pricing requires at least one tier")
	ErrInvalidTiers    = errors.New("INVALID_TIERS", "final tier must have null up_to")
	ErrInvalidCurrency = errors.New("INVALID_CURRENCY", "invalid currency code")
)
