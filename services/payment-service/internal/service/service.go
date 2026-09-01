package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/trunglq04/goride/services/payment-service/internal/domain"
	"github.com/trunglq04/goride/services/payment-service/pkg/types"
)

type paymentService struct {
	paymentProcessor domain.PaymentProcessor
	publisher        domain.PaymentEventPublisher
}

func NewPaymentService(paymentProcessor domain.PaymentProcessor, publisher domain.PaymentEventPublisher) domain.Service {
	return &paymentService{
		paymentProcessor: paymentProcessor,
		publisher:        publisher,
	}
}

// CreatePaymentSession creates a new payment session for a trip
func (s *paymentService) CreatePaymentSession(ctx context.Context, tripID string, userID string, driverID string, amount int64, currency string) (*types.PaymentIntent, error) {
	metadata := map[string]string{
		"trip_id":   tripID,
		"user_id":   userID,
		"driver_id": driverID,
	}

	sessionID, err := s.paymentProcessor.CreatePaymentSession(ctx, amount, currency, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment session: %w", err)
	}

	paymentIntent := &types.PaymentIntent{
		ID:              uuid.New().String(),
		TripID:          tripID,
		UserID:          userID,
		DriverID:        driverID,
		Amount:          amount,
		Currency:        currency,
		StripeSessionID: sessionID,
		CreatedAt:       time.Now(),
	}

	if err := s.publisher.PublishSessionCreated(ctx, tripID, sessionID, userID, float64(amount)/100.0, currency); err != nil {
		return nil, fmt.Errorf("failed to publish payment session created event: %w", err)
	}

	return paymentIntent, nil
}
