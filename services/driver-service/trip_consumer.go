package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"

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
		if err := c.handleFindAndNotifyDrivers(ctx, payload); err != nil {
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

func (c *TripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	packageSlug := payload.Trip.GetSelectedFare().GetPackageSlug()

	suitableIDs := c.service.FindAvailableDrivers(
		packageSlug,
		payload.ExcludeDriverIDs,
	)

	slog.InfoContext(ctx, "Searching for suitable drivers",
		"package_slug", packageSlug,
		"found", len(suitableIDs),
		"excluded_drivers", len(payload.ExcludeDriverIDs),
	)

	if len(suitableIDs) == 0 {
		slog.WarnContext(ctx, "No suitable drivers found, notifying rider",
			"package_slug", packageSlug,
			"trip_id", payload.Trip.GetId(),
		)
		// Notify the rider that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx,
			contracts.TripEventNoDriversFound,
			contracts.AmqpMessage{
				OwnerID: payload.Trip.UserID,
			}); err != nil {
			slog.ErrorContext(ctx, "Failed to publish no-drivers-found event",
				"routing_key", contracts.TripEventNoDriversFound,
				"trip_id", payload.Trip.GetId(),
				"err", err,
			)
			return err
		}

		return nil
	}

	randIndex := rand.Intn(len(suitableIDs))
	suitableDriverID := suitableIDs[randIndex]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal trip event",
			"trip_id", payload.Trip.GetId(),
			"err", err,
		)
		return err
	}

	// Ask drivers if they confirm the trip request
	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.DriverCmdTripRequest,
		contracts.AmqpMessage{
			OwnerID: suitableDriverID,
			Data:    marshalledEvent,
		}); err != nil {
		slog.ErrorContext(ctx, "Failed to publish trip request to driver",
			"routing_key", contracts.DriverCmdTripRequest,
			"driver_id", suitableDriverID,
			"trip_id", payload.Trip.GetId(),
			"err", err,
		)
		return err
	}

	slog.InfoContext(ctx, "Trip request sent to driver",
		"driver_id", suitableDriverID,
		"trip_id", payload.Trip.GetId(),
		"package_slug", packageSlug,
	)

	return nil
}
