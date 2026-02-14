package payment

import "time"

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusRefunded   Status = "refunded"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusSucceeded, StatusFailed, StatusRefunded:
		return true
	}
	return false
}

type Payment struct {
	ID              string
	InvoiceID       string
	Amount          int64
	Currency        string
	Status          Status
	PaymentMethodID *string
	ProviderID      *string
	FailureCode     *string
	FailureMessage  *string
	IdempotencyKey  string
	Metadata        map[string]string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (p *Payment) IsSucceeded() bool {
	return p.Status == StatusSucceeded
}

func (p *Payment) IsFailed() bool {
	return p.Status == StatusFailed
}

func (p *Payment) IsRefunded() bool {
	return p.Status == StatusRefunded
}

func (p *Payment) CanBeRefunded() bool {
	return p.Status == StatusSucceeded
}

func (p *Payment) MarkSucceeded(providerID string) {
	p.Status = StatusSucceeded
	p.ProviderID = &providerID
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkFailed(code, message string) {
	p.Status = StatusFailed
	p.FailureCode = &code
	p.FailureMessage = &message
	p.UpdatedAt = time.Now().UTC()
}

func (p *Payment) MarkRefunded() {
	p.Status = StatusRefunded
	p.UpdatedAt = time.Now().UTC()
}

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
)

type Refund struct {
	ID         string
	PaymentID  string
	Amount     int64
	Currency   string
	Status     RefundStatus
	Reason     *string
	ProviderID *string
	Metadata   map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *Refund) MarkSucceeded(providerID string) {
	r.Status = RefundStatusSucceeded
	r.ProviderID = &providerID
	r.UpdatedAt = time.Now().UTC()
}

func (r *Refund) MarkFailed() {
	r.Status = RefundStatusFailed
	r.UpdatedAt = time.Now().UTC()
}

type PaymentRetry struct {
	ID            string
	InvoiceID     string
	AttemptNumber int
	ScheduledAt   time.Time
	AttemptedAt   *time.Time
	Status        string
	PaymentID     *string
	CreatedAt     time.Time
}
