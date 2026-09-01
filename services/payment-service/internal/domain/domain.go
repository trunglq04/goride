package domain

import (
	"context"

	"github.com/trunglq04/goride/services/payment-service/pkg/types"
)

type Service interface {
	CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) (*types.PaymentIntent, error)
}

type PaymentProcessor interface {
	CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata map[string]string) (string, error)
}

type PaymentEventPublisher interface {
	PublishSessionCreated(ctx context.Context, tripID, sessionID, userID string, amount float64, currency string) error
}
