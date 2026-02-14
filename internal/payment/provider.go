package payment

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type Provider interface {
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error)
	Refund(ctx context.Context, req ProviderRefundRequest) (*ProviderRefundResponse, error)
}

type ChargeRequest struct {
	Amount          int64
	Currency        string
	PaymentMethodID string
	IdempotencyKey  string
	Metadata        map[string]string
}

type ChargeResponse struct {
	ProviderID string
	Status     string
	Message    string
}

type ProviderRefundRequest struct {
	PaymentProviderID string
	Amount            int64
	Currency          string
	Reason            string
}

type ProviderRefundResponse struct {
	ProviderID string
	Status     string
	Message    string
}

type StubProvider struct {
	successRate float64
	rand        *rand.Rand
}

func NewStubProvider(successRate float64) *StubProvider {
	return &StubProvider{
		successRate: successRate,
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *StubProvider) Charge(ctx context.Context, req ChargeRequest) (*ChargeResponse, error) {
	time.Sleep(100 * time.Millisecond)

	if p.rand.Float64() < p.successRate {
		return &ChargeResponse{
			ProviderID: fmt.Sprintf("ch_%d", time.Now().UnixNano()),
			Status:     "succeeded",
			Message:    "Payment succeeded",
		}, nil
	}

	return &ChargeResponse{
		Status:  "failed",
		Message: "Insufficient funds",
	}, ErrProviderDeclined
}

func (p *StubProvider) Refund(ctx context.Context, req ProviderRefundRequest) (*ProviderRefundResponse, error) {
	time.Sleep(100 * time.Millisecond)

	if p.rand.Float64() < 0.95 {
		return &ProviderRefundResponse{
			ProviderID: fmt.Sprintf("re_%d", time.Now().UnixNano()),
			Status:     "succeeded",
			Message:    "Refund succeeded",
		}, nil
	}

	return &ProviderRefundResponse{
		Status:  "failed",
		Message: "Refund failed",
	}, fmt.Errorf("refund failed")
}
