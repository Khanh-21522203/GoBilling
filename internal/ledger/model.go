package ledger

import "time"

type TransactionType string

const (
	TypeCharge     TransactionType = "charge"
	TypePayment    TransactionType = "payment"
	TypeRefund     TransactionType = "refund"
	TypeCredit     TransactionType = "credit"
	TypeAdjustment TransactionType = "adjustment"
)

type LedgerTransaction struct {
	ID          string
	CustomerID  string
	Type        TransactionType
	Amount      int64
	Currency    string
	InvoiceID   *string
	PaymentID   *string
	RefundID    *string
	Description string
	CreatedAt   time.Time
}

func NewChargeEntry(customerID, invoiceID string, amount int64, currency, description string) *LedgerTransaction {
	now := time.Now().UTC()
	return &LedgerTransaction{
		Type:        TypeCharge,
		CustomerID:  customerID,
		InvoiceID:   &invoiceID,
		Amount:      amount,
		Currency:    currency,
		Description: description,
		CreatedAt:   now,
	}
}

func NewPaymentEntry(customerID, paymentID string, amount int64, currency, description string) *LedgerTransaction {
	now := time.Now().UTC()
	return &LedgerTransaction{
		Type:        TypePayment,
		CustomerID:  customerID,
		PaymentID:   &paymentID,
		Amount:      amount,
		Currency:    currency,
		Description: description,
		CreatedAt:   now,
	}
}

func NewRefundEntry(customerID, refundID string, amount int64, currency, description string) *LedgerTransaction {
	now := time.Now().UTC()
	return &LedgerTransaction{
		Type:        TypeRefund,
		CustomerID:  customerID,
		RefundID:    &refundID,
		Amount:      amount,
		Currency:    currency,
		Description: description,
		CreatedAt:   now,
	}
}
