package events

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
)

type paymentConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewPaymentConsumer(rb *messaging.RabbitMQ, service domain.TripService) *paymentConsumer {
	return &paymentConsumer{
		rabbitmq: rb,
		service:  service,
	}
}

func (c *paymentConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(
		messaging.NotifyPaymentSuccessQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var message contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &message); err != nil {
				slog.ErrorContext(ctx, "Failed to unmarshal payment status message",
					"queue", messaging.NotifyPaymentSuccessQueue,
					"routing_key", msg.RoutingKey,
					"err", err,
				)
				return err
			}

			var payload messaging.PaymentStatusUpdateData
			if err := json.Unmarshal(message.Data, &payload); err != nil {
				slog.ErrorContext(ctx, "Failed to unmarshal payment status payload",
					"queue", messaging.NotifyPaymentSuccessQueue,
					"routing_key", msg.RoutingKey,
					"trip_id", payload.TripID,
					"err", err,
				)
				return err
			}

			slog.InfoContext(ctx, "Trip completed and payed, updating trip status",
				"trip_id", payload.TripID,
				"user_id", payload.UserID,
				"driver_id", payload.DriverID,
			)

			if err := c.service.UpdateTrip(
				ctx,
				payload.TripID,
				"payed",
				nil,
				nil,
			); err != nil {
				slog.ErrorContext(ctx, "Failed to update trip status to payed",
					"trip_id", payload.TripID,
					"err", err,
				)
				return err
			}

			return nil
		})
}
