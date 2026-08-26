package events

import (
	"context"
	"encoding/json"
	"log/slog"

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
				slog.ErrorContext(ctx, "Failed to unmarshal payment command message",
					"queue", messaging.PaymentTripResponseQueue,
					"routing_key", msg.RoutingKey,
					"err", err,
				)
				return err
			}

			var payload messaging.PaymentTripResponseData
			if err := json.Unmarshal(message.Data, &payload); err != nil {
				slog.ErrorContext(ctx, "Failed to unmarshal PaymentTripResponseData",
					"queue", messaging.PaymentTripResponseQueue,
					"routing_key", msg.RoutingKey,
					"trip_id", payload.TripID,
					"err", err,
				)
				return err
			}

			switch msg.RoutingKey {
			case contracts.PaymentCmdCreateSession:
				if err := c.handleTripAccepted(ctx, payload); err != nil {
					slog.ErrorContext(ctx, "Failed to handle trip accepted",
						"trip_id", payload.TripID,
						"user_id", payload.UserID,
						"driver_id", payload.DriverID,
						"err", err,
					)
					return err
				}
			default:
				slog.WarnContext(ctx, "Unknown payment command routing key",
					"routing_key", msg.RoutingKey,
				)
			}

			return nil
		})
}

func (c *TripConsumer) handleTripAccepted(ctx context.Context, payload messaging.PaymentTripResponseData) error {
	log := slog.Default()
	log.InfoContext(ctx, "Handling trip accepted by driver, creating payment session",
		"trip_id", payload.TripID,
		"user_id", payload.UserID,
		"driver_id", payload.DriverID,
		"amount", payload.Amount,
		"currency", payload.Currency,
	)

	paymentSession, err := c.service.CreatePaymentSession(
		ctx,
		payload.TripID,
		payload.UserID,
		payload.DriverID,
		int64(payload.Amount),
		payload.Currency,
	)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create payment session",
			"trip_id", payload.TripID,
			"err", err,
		)
		return err
	}

	log.InfoContext(ctx, "Payment session created",
		"trip_id", payload.TripID,
		"session_id", paymentSession.StripeSessionID,
	)

	// Publish payment session created event
	paymentPayload := messaging.PaymentEventSessionCreatedData{
		TripID:    payload.TripID,
		SessionID: paymentSession.StripeSessionID,
		Amount:    float64(paymentSession.Amount) / 100.0, // Convert from cents to dollars
		Currency:  paymentSession.Currency,
	}

	payloadBytes, err := json.Marshal(paymentPayload)
	if err != nil {
		log.ErrorContext(ctx, "Failed to marshal payment session created event",
			"trip_id", payload.TripID,
			"err", err,
		)
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.PaymentEventSessionCreated,
		contracts.AmqpMessage{
			OwnerID: payload.UserID,
			Data:    payloadBytes,
		},
	); err != nil {
		log.ErrorContext(ctx, "Failed to publish payment session created event",
			"routing_key", contracts.PaymentEventSessionCreated,
			"trip_id", payload.TripID,
			"err", err,
		)
		return err
	}

	log.InfoContext(ctx, "Published payment session created event",
		"routing_key", contracts.PaymentEventSessionCreated,
		"trip_id", payload.TripID,
	)
	return nil
}
