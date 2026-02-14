package event

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type Event struct {
	ID           string
	Type         string
	Payload      map[string]interface{}
	Status       Status
	RetryCount   int
	NextRetryAt  *time.Time
	DeliveredAt  *time.Time
	CreatedAt    time.Time
}

func (e *Event) MarkDelivered() {
	e.Status = StatusDelivered
	now := time.Now().UTC()
	e.DeliveredAt = &now
}

func (e *Event) MarkFailed() {
	e.Status = StatusFailed
}

func (e *Event) ScheduleRetry(delay time.Duration) {
	e.RetryCount++
	nextRetry := time.Now().UTC().Add(delay)
	e.NextRetryAt = &nextRetry
}
