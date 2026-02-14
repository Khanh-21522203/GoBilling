package webhook

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gobilling/internal/pkg/id"
)

type Service struct {
	endpointRepo EndpointRepository
	deliveryRepo DeliveryRepository
}

func NewService(endpointRepo EndpointRepository, deliveryRepo DeliveryRepository) *Service {
	return &Service{
		endpointRepo: endpointRepo,
		deliveryRepo: deliveryRepo,
	}
}

func (s *Service) CreateEndpoint(ctx context.Context, url string, events []string) (*WebhookEndpoint, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	now := time.Now().UTC()
	endpoint := &WebhookEndpoint{
		ID:        id.NewWithPrefix("we_"),
		URL:       url,
		Secret:    secret,
		Events:    events,
		Active:    true,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.endpointRepo.Create(ctx, endpoint); err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (s *Service) GetEndpoint(ctx context.Context, id string) (*WebhookEndpoint, error) {
	return s.endpointRepo.GetByID(ctx, id)
}

func (s *Service) UpdateEndpoint(ctx context.Context, id string, url *string, events []string, active *bool) (*WebhookEndpoint, error) {
	endpoint, err := s.endpointRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if url != nil {
		endpoint.URL = *url
	}

	if events != nil {
		endpoint.Events = events
	}

	if active != nil {
		endpoint.Active = *active
	}

	endpoint.UpdatedAt = time.Now().UTC()

	if err := s.endpointRepo.Update(ctx, endpoint); err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (s *Service) DeleteEndpoint(ctx context.Context, id string) error {
	return s.endpointRepo.Delete(ctx, id)
}

func (s *Service) ListEndpoints(ctx context.Context, active *bool, limit int) ([]*WebhookEndpoint, error) {
	return s.endpointRepo.List(ctx, active, limit)
}

func (s *Service) EnqueueDelivery(ctx context.Context, endpointID, eventID string) error {
	delivery := &WebhookDelivery{
		ID:                id.NewWithPrefix("wd_"),
		WebhookEndpointID: endpointID,
		EventID:           eventID,
		Status:            DeliveryStatusPending,
		AttemptCount:      0,
		CreatedAt:         time.Now().UTC(),
	}

	return s.deliveryRepo.Create(ctx, delivery)
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(bytes), nil
}
