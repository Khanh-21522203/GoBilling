package subscription

import "gobilling/internal/platform/errors"

var (
	ErrNotFound                = errors.New("SUBSCRIPTION_NOT_FOUND", "subscription not found")
	ErrInvalidStatusTransition = errors.New("INVALID_STATUS_TRANSITION", "invalid status transition")
	ErrAlreadyCanceled         = errors.New("SUBSCRIPTION_ALREADY_CANCELED", "subscription already canceled")
	ErrCannotModify            = errors.New("SUBSCRIPTION_CANNOT_MODIFY", "subscription cannot be modified in current status")
	ErrCustomerNotActive       = errors.New("CUSTOMER_NOT_ACTIVE", "customer must be active")
	ErrPlanNotActive           = errors.New("PLAN_NOT_ACTIVE", "plan must be active")
)
