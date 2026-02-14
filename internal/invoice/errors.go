package invoice

import "gobilling/internal/platform/errors"

var (
	ErrNotFound        = errors.New("INVOICE_NOT_FOUND", "invoice not found")
	ErrCannotFinalize  = errors.New("INVOICE_CANNOT_FINALIZE", "invoice cannot be finalized")
	ErrCannotVoid      = errors.New("INVOICE_CANNOT_VOID", "invoice cannot be voided")
	ErrNotOpen         = errors.New("INVOICE_NOT_OPEN", "invoice is not open")
	ErrAlreadyPaid     = errors.New("INVOICE_ALREADY_PAID", "invoice already paid")
	ErrNoLineItems     = errors.New("INVOICE_NO_LINE_ITEMS", "invoice has no line items")
	ErrHasPayments     = errors.New("INVOICE_HAS_PAYMENTS", "invoice has payments")
)
