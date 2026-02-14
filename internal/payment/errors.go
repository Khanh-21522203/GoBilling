package payment

import "gobilling/internal/platform/errors"

var (
	ErrNotFound          = errors.New("PAYMENT_NOT_FOUND", "payment not found")
	ErrFailed            = errors.New("PAYMENT_FAILED", "payment failed")
	ErrCannotRefund      = errors.New("PAYMENT_CANNOT_REFUND", "payment cannot be refunded")
	ErrAlreadyRefunded   = errors.New("PAYMENT_ALREADY_REFUNDED", "payment already refunded")
	ErrRefundNotFound    = errors.New("REFUND_NOT_FOUND", "refund not found")
	ErrProviderTimeout   = errors.New("PAYMENT_PROVIDER_TIMEOUT", "payment provider timeout")
	ErrProviderDeclined  = errors.New("PAYMENT_DECLINED", "payment declined by provider")
)
