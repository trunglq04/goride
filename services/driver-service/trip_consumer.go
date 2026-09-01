package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
	pb "github.com/trunglq04/goride/shared/proto/driver"
)

type TripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *Service) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *TripConsumer) Listen() error {
	consumers := []struct {
		queue   string
		handler func(context.Context, amqp091.Delivery) error
	}{
		{
			queue:   messaging.FindAvailableDriversQueue,
			handler: c.handleTripEvents,
		},
		{
			queue:   messaging.NotifyDriverLocationQueue,
			handler: c.handleDriverLocationUpdate,
		},
	}

	for _, consumer := range consumers {
		if err := c.rabbitmq.ConsumeMessages(consumer.queue, consumer.handler); err != nil {
			return err
		}
	}

	return nil
}

func (c *TripConsumer) handleTripEvents(ctx context.Context, msg amqp091.Delivery) error {
	var tripEvent contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal trip event message",
			"queue", messaging.FindAvailableDriversQueue,
			"routing_key", msg.RoutingKey,
			"err", err,
		)
		return err
	}

	var payload messaging.TripEventData
	if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal trip event payload",
			"queue", messaging.FindAvailableDriversQueue,
			"routing_key", msg.RoutingKey,
			"trip_id", payload.Trip.GetId(),
			"err", err,
		)
		return err
	}

	slog.InfoContext(ctx, "Trip event received",
		"routing_key", msg.RoutingKey,
		"trip_id", payload.Trip.GetId(),
	)

	switch msg.RoutingKey {
	case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
		if err := c.service.HandleTripCreated(ctx, payload); err != nil {
			return err
		}
	case contracts.TripEventCanceled:
		if err := c.service.HandleTripCanceled(ctx, payload); err != nil {
			return err
		}
	default:
		slog.WarnContext(ctx, "Unknown trip event routing key", "routing_key", msg.RoutingKey)
	}

	return nil
}

func (c *TripConsumer) handleDriverLocationUpdate(ctx context.Context, msg amqp091.Delivery) error {
	var locationEvent contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &locationEvent); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal location update message",
			"queue", messaging.NotifyDriverLocationQueue,
			"routing_key", msg.RoutingKey,
			"driver_id", locationEvent.OwnerID,
			"err", err,
		)
		return err
	}

	var req pb.UpdateDriverLocationRequest
	if err := json.Unmarshal(locationEvent.Data, &req); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal location update payload",
			"queue", messaging.NotifyDriverLocationQueue,
			"routing_key", msg.RoutingKey,
			"driver_id", locationEvent.OwnerID,
			"err", err,
		)
		return err
	}

	if _, err := c.service.UpdateDriverLocation(locationEvent.OwnerID, req.GetLocation(), req.GetGeohash()); err != nil {
		slog.ErrorContext(ctx, "Failed to update driver location",
			"driver_id", locationEvent.OwnerID,
			"err", err,
		)
		return err
	}

	return nil
}
