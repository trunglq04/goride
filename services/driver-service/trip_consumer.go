package main

import (
	"context"
	"encoding/json"
	"log"
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
		log.Printf("ERROR: Failed to unmarshal message: %v", err)
		return err
	}

	var payload messaging.TripEventData
	if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
		log.Printf("ERROR: Failed to unmarshal payload: %v", err)
		return err
	}

	log.Printf("Driver received message: %+v", payload)

	switch msg.RoutingKey {
	case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
		if err := c.handleFindAndNotifyDrivers(ctx, payload); err != nil {
			return err
		}
	default:
		log.Printf("unknown trip event routing key: %s", msg.RoutingKey)
	}

	return nil
}

func (c *TripConsumer) handleDriverLocationUpdate(ctx context.Context, msg amqp091.Delivery) error {
	var locationEvent contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &locationEvent); err != nil {
		log.Printf("ERROR: Failed to unmarshal message: %v", err)
		return err
	}

	var req pb.UpdateDriverLocationRequest
	if err := json.Unmarshal(locationEvent.Data, &req); err != nil {
		log.Printf("ERROR: Failed to unmarshal payload: %v", err)
		return err
	}

	if _, err := c.service.UpdateDriverLocation(locationEvent.OwnerID, req.GetLocation(), req.GetGeohash()); err != nil {
		log.Printf("ERROR: Failed to update driver location: %v", err)
		return err
	}

	return nil
}

func (c *TripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs := c.service.FindAvailableDrivers(
		payload.Trip.GetSelectedFare().PackageSlug,
		payload.ExcludeDriverIDs,
	)

	log.Printf("Found suitable %v drivers", len(suitableIDs))

	if len(suitableIDs) == 0 {
		log.Printf("No suitable drivers found for packageSlug=%q, notifying rider", payload.Trip.GetSelectedFare().GetPackageSlug())
		// Notify the rider that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx,
			contracts.TripEventNoDriversFound,
			contracts.AmqpMessage{
				OwnerID: payload.Trip.UserID,
			}); err != nil {
			log.Printf("ERROR: Failed to publish message to exchange: %v", err)
			return err
		}

		return nil
	}

	randIndex := rand.Intn(len(suitableIDs))
	suitableDriverID := suitableIDs[randIndex]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Ask drivers if they confirm the trip request
	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.DriverCmdTripRequest,
		contracts.AmqpMessage{
			OwnerID: suitableDriverID,
			Data:    marshalledEvent,
		}); err != nil {
		log.Printf("ERROR: Failed to publish message to exchange: %v", err)
		return err
	}

	return nil
}
