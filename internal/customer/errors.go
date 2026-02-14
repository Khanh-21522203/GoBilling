package customer

import "gobilling/internal/platform/errors"

var (
	ErrNotFound                = errors.New("CUSTOMER_NOT_FOUND", "customer not found")
	ErrDeleted                 = errors.New("CUSTOMER_DELETED", "customer is deleted")
	ErrAlreadyDeleted          = errors.New("CUSTOMER_ALREADY_DELETED", "customer already deleted")
	ErrHasActiveSubscriptions  = errors.New("CUSTOMER_HAS_ACTIVE_SUBSCRIPTIONS", "customer has active subscriptions")
	ErrInvalidStatusTransition = errors.New("INVALID_STATUS_TRANSITION", "invalid status transition")
	ErrEmailAlreadyExists      = errors.New("EMAIL_ALREADY_EXISTS", "email already exists")
	ErrInvalidEmail            = errors.New("INVALID_EMAIL", "invalid email format")
)
