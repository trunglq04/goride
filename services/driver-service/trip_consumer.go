package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
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
	return c.rabbitmq.ConsumeMessages(
		messaging.FindAvailableDriversQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var tripEvent contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				return err
			}

			var payload messaging.TripEventData
			if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
				log.Printf("Failed to unmarshal payload: %v", err)
				return err
			}

			log.Printf("Driver received message: %+v", payload)

			switch msg.RoutingKey {
			case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
				if err := c.handleFindAndNotifyDrivers(ctx, payload); err != nil {
					return err
				}

				return msg.Ack(false)
			}

			log.Printf("unknown trip event routing key: %s", msg.RoutingKey)

			return msg.Ack(false)
		})
}

func (c *TripConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs := c.service.FindAvailableDrivers(payload.Trip.GetSelectedFare().PackageSlug)

	log.Printf("Found suitable %v drivers", len(suitableIDs))

	if len(suitableIDs) == 0 {
		// Notify the rider that no drivers are available
		if err := c.rabbitmq.PublishMessage(ctx,
			contracts.TripEventNoDriversFound,
			contracts.AmqpMessage{
				OwnerID: payload.Trip.UserID,
			}); err != nil {
			log.Printf("Failed to publish message to exchange: %v", err)
			return err
		}

		return nil
	}

	suitableDriverID := suitableIDs[0]

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
		log.Printf("Failed to publish message to exchange: %v", err)
		return err
	}

	return nil
}
