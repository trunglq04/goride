package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/services/payment-service/internal/domain"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
)

type TripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service domain.Service) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *TripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(
		messaging.PaymentTripResponseQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var message contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &message); err != nil {
				log.Printf("ERROR: Failed to unmarshal message: %v", err)
				return err
			}

			var payload messaging.PaymentTripResponseData
			if err := json.Unmarshal(message.Data, &payload); err != nil {
				log.Printf("ERROR: Failed to unmarshal PaymentTripResponseData: %v", err)
				return err
			}

			switch msg.RoutingKey {
			case contracts.PaymentCmdCreateSession:
				if err := c.handleTripAccepted(ctx, payload); err != nil {
					log.Printf("ERROR: Failed to handle trip accepted: %v", err)
					return err
				}
			}

			return nil
		})
}

func (c *TripConsumer) handleTripAccepted(ctx context.Context, payload messaging.PaymentTripResponseData) error {
	log.Printf("Handling trip accepted by driver: %s", payload.TripID)

	paymentSession, err := c.service.CreatePaymentSession(
		ctx,
		payload.TripID,
		payload.UserID,
		payload.DriverID,
		int64(payload.Amount),
		payload.Currency,
	)
	if err != nil {
		log.Printf("ERROR: Failed to create payment session: %v", err)
		return err
	}

	log.Printf("Payment session created: %s", paymentSession.StripeSessionID)

	// Publish payment session created event
	paymentPayload := messaging.PaymentEventSessionCreatedData{
		TripID:    payload.TripID,
		SessionID: paymentSession.StripeSessionID,
		Amount:    float64(paymentSession.Amount) / 100.0, // Convert from cents to dollars
		Currency:  paymentSession.Currency,
	}

	payloadBytes, err := json.Marshal(paymentPayload)
	if err != nil {
		log.Printf("ERROR: Failed to marshal payment session payload: %v", err)
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.PaymentEventSessionCreated,
		contracts.AmqpMessage{
			OwnerID: payload.UserID,
			Data:    payloadBytes,
		},
	); err != nil {
		log.Printf("ERROR: Failed to publish payment session created event: %v", err)
		return err
	}

	log.Printf("Published payment session created event for trip: %s", payload.TripID)
	return nil
}
