package errors

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func New(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Wrap(err error, code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

var (
	ErrNotFound            = New("NOT_FOUND", "resource not found")
	ErrValidation          = New("VALIDATION_ERROR", "validation failed")
	ErrConflict            = New("CONFLICT", "resource conflict")
	ErrInvalidState        = New("INVALID_STATE", "invalid state transition")
	ErrUnauthorized        = New("UNAUTHORIZED", "unauthorized")
	ErrForbidden           = New("FORBIDDEN", "forbidden")
	ErrInternal            = New("INTERNAL_ERROR", "internal server error")
	ErrDatabaseConnection  = New("DATABASE_CONNECTION", "database connection failed")
	ErrDatabaseTimeout     = New("DATABASE_TIMEOUT", "database operation timeout")
)

var (
	ErrCustomerNotFound                = New("CUSTOMER_NOT_FOUND", "customer not found")
	ErrCustomerDeleted                 = New("CUSTOMER_DELETED", "customer is deleted")
	ErrCustomerHasActiveSubscriptions  = New("CUSTOMER_HAS_ACTIVE_SUBSCRIPTIONS", "customer has active subscriptions")
	ErrProductNotFound                 = New("PRODUCT_NOT_FOUND", "product not found")
	ErrPlanNotFound                    = New("PLAN_NOT_FOUND", "plan not found")
	ErrPlanInactive                    = New("PLAN_INACTIVE", "plan is not active")
	ErrSubscriptionNotFound            = New("SUBSCRIPTION_NOT_FOUND", "subscription not found")
	ErrSubscriptionNotActive           = New("SUBSCRIPTION_NOT_ACTIVE", "subscription is not active")
	ErrSubscriptionAlreadyCanceled     = New("SUBSCRIPTION_ALREADY_CANCELED", "subscription already canceled")
	ErrInvoiceNotFound                 = New("INVOICE_NOT_FOUND", "invoice not found")
	ErrInvoiceNotOpen                  = New("INVOICE_NOT_OPEN", "invoice is not open")
	ErrInvoiceAlreadyPaid              = New("INVOICE_ALREADY_PAID", "invoice already paid")
	ErrPaymentNotFound                 = New("PAYMENT_NOT_FOUND", "payment not found")
	ErrPaymentFailed                   = New("PAYMENT_FAILED", "payment failed")
	ErrRefundAlreadyExists             = New("REFUND_ALREADY_EXISTS", "payment already refunded")
	ErrRateLimitExceeded               = New("RATE_LIMIT_EXCEEDED", "rate limit exceeded")
	ErrIdempotencyConflict             = New("IDEMPOTENCY_CONFLICT", "idempotency key conflict")
	ErrRequestInProgress               = New("REQUEST_IN_PROGRESS", "request already in progress")
)

func Is(err error, target *Error) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == target.Code
	}
	return false
}
