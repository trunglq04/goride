package events

import (
	"context"
	"encoding/json"

	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
)

type PaymentEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewPaymentEventPublisher(rabbitmq *messaging.RabbitMQ) *PaymentEventPublisher {
	return &PaymentEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *PaymentEventPublisher) PublishSessionCreated(ctx context.Context, tripID, sessionID, userID string, amount float64, currency string) error {
	paymentPayload := messaging.PaymentEventSessionCreatedData{
		TripID:    tripID,
		SessionID: sessionID,
		Amount:    amount,
		Currency:  currency,
	}

	payloadBytes, err := json.Marshal(paymentPayload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx,
		contracts.PaymentEventSessionCreated,
		contracts.AmqpMessage{
			OwnerID: userID,
			Data:    payloadBytes,
		},
	)
}
