package invoice

import "time"

type Status string

const (
	StatusDraft         Status = "draft"
	StatusOpen          Status = "open"
	StatusPaid          Status = "paid"
	StatusVoid          Status = "void"
	StatusUncollectible Status = "uncollectible"
)

func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusOpen, StatusPaid, StatusVoid, StatusUncollectible:
		return true
	}
	return false
}

type Invoice struct {
	ID             string
	InvoiceNumber  string
	CustomerID     string
	SubscriptionID *string
	Status         Status
	Currency       string
	Subtotal       int64
	DiscountAmount int64
	TaxAmount      int64
	Total          int64
	AmountPaid     int64
	AmountDue      int64
	PeriodStart    *time.Time
	PeriodEnd      *time.Time
	DueDate        *time.Time
	PaidAt         *time.Time
	VoidedAt       *time.Time
	Metadata       map[string]string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LineItems      []*LineItem
}

type LineItem struct {
	ID          string
	InvoiceID   string
	Description string
	Quantity    int64
	UnitAmount  int64
	Amount      int64
	PeriodStart *time.Time
	PeriodEnd   *time.Time
	Metadata    map[string]string
	CreatedAt   time.Time
}

func (i *Invoice) IsDraft() bool {
	return i.Status == StatusDraft
}

func (i *Invoice) IsOpen() bool {
	return i.Status == StatusOpen
}

func (i *Invoice) IsPaid() bool {
	return i.Status == StatusPaid
}

func (i *Invoice) CanBeFinalized() bool {
	return i.Status == StatusDraft
}

func (i *Invoice) CanBeVoided() bool {
	return i.Status == StatusOpen
}

func (i *Invoice) Finalize() error {
	if !i.CanBeFinalized() {
		return ErrCannotFinalize
	}

	if len(i.LineItems) == 0 {
		return ErrNoLineItems
	}

	i.CalculateTotals()

	if i.Total == 0 {
		i.Status = StatusPaid
		now := time.Now().UTC()
		i.PaidAt = &now
	} else {
		i.Status = StatusOpen
	}

	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *Invoice) Void() error {
	if !i.CanBeVoided() {
		return ErrCannotVoid
	}

	i.Status = StatusVoid
	now := time.Now().UTC()
	i.VoidedAt = &now
	i.UpdatedAt = now
	return nil
}

func (i *Invoice) MarkPaid() error {
	if i.Status != StatusOpen {
		return ErrNotOpen
	}

	i.Status = StatusPaid
	now := time.Now().UTC()
	i.PaidAt = &now
	i.AmountPaid = i.Total
	i.AmountDue = 0
	i.UpdatedAt = now
	return nil
}

func (i *Invoice) MarkUncollectible() error {
	if i.Status != StatusOpen {
		return ErrNotOpen
	}

	i.Status = StatusUncollectible
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *Invoice) RecordPayment(amount int64) {
	i.AmountPaid += amount
	i.AmountDue = i.Total - i.AmountPaid

	if i.AmountDue <= 0 {
		i.Status = StatusPaid
		now := time.Now().UTC()
		i.PaidAt = &now
	}

	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) CalculateTotals() {
	i.Subtotal = 0
	for _, item := range i.LineItems {
		i.Subtotal += item.Amount
	}

	i.Total = i.Subtotal - i.DiscountAmount + i.TaxAmount
	i.AmountDue = i.Total - i.AmountPaid
}

func (i *Invoice) AddLineItem(item *LineItem) {
	i.LineItems = append(i.LineItems, item)
	i.CalculateTotals()
}
