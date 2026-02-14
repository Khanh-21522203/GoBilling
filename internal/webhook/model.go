package webhook

import "time"

type WebhookEndpoint struct {
	ID        string
	URL       string
	Secret    string
	Events    []string
	Active    bool
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusSkipped   DeliveryStatus = "skipped"
)

type WebhookDelivery struct {
	ID                string
	WebhookEndpointID string
	EventID           string
	Status            DeliveryStatus
	ResponseCode      *int
	ResponseBody      *string
	AttemptCount      int
	NextAttemptAt     *time.Time
	DeliveredAt       *time.Time
	CreatedAt         time.Time
}

func (d *WebhookDelivery) MarkDelivered(responseCode int, responseBody string) {
	d.Status = DeliveryStatusDelivered
	d.ResponseCode = &responseCode
	d.ResponseBody = &responseBody
	now := time.Now().UTC()
	d.DeliveredAt = &now
}

func (d *WebhookDelivery) MarkFailed(responseCode int, responseBody string) {
	d.Status = DeliveryStatusFailed
	d.ResponseCode = &responseCode
	d.ResponseBody = &responseBody
}

func (d *WebhookDelivery) ScheduleRetry(delay time.Duration) {
	d.AttemptCount++
	nextAttempt := time.Now().UTC().Add(delay)
	d.NextAttemptAt = &nextAttempt
}

func GetRetryDelay(attemptCount int) time.Duration {
	switch attemptCount {
	case 1:
		return 5 * time.Minute
	case 2:
		return 30 * time.Minute
	case 3:
		return 2 * time.Hour
	case 4:
		return 8 * time.Hour
	case 5:
		return 24 * time.Hour
	default:
		return 0
	}
}
