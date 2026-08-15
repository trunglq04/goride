package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
	pbd "github.com/trunglq04/goride/shared/proto/driver"
)

type driverConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(rabbimq *messaging.RabbitMQ, service domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitmq: rabbimq,
		service:  service,
	}
}

func (c *driverConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(
		messaging.DriverTripResponseQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var message contracts.AmqpMessage
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				return err
			}

			var payload messaging.DriverTripResponseData
			err = json.Unmarshal(message.Data, &payload)
			if err != nil {
				log.Printf("Failed to unmarshal driver trip response: %v", err)
				return err
			}

			log.Printf("driver response received message: %+v", payload)

			switch msg.RoutingKey {
			case contracts.DriverCmdTripAccept:
				if err := c.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
					log.Printf("Failed to handle the trip accept: %v", err)
					return err
				}
			case contracts.DriverCmdTripDecline:
				if err := c.handleTripDeclined(ctx, payload.TripID, payload.RiderID); err != nil {
					log.Printf("Failed to handle the trip decline: %v", err)
					return err
				}
			}

			// Unknown routing key — discard, do not requeue.
			log.Printf("unknown driver response routing key: %s", msg.RoutingKey)
			return nil
		})
}

func (c *driverConsumer) handleTripDeclined(ctx context.Context, tripID, riderID string) error {
	// When a driver declines, we should try to find another driver

	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	newPayload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.TripEventDriverNotInterested,
		contracts.AmqpMessage{
			OwnerID: riderID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}

	return nil
}

func (c *driverConsumer) handleTripAccepted(ctx context.Context, tripID string, driver *pbd.Driver) error {
	// 1. Fetch the first
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	if trip == nil {
		return fmt.Errorf("Trip was not found %s", tripID)
	}

	// 2. Update the trip
	if err := c.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		log.Printf("Failed to update the trip: %v", err)
		return err
	}

	trip, err = c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}

	// 3. Driver has been assigned -> publish this event to RB
	marshalledTrip, err := json.Marshal(trip)
	if err != nil {
		return err
	}

	// Notify the rider that a driver has been assigned
	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.TripEventDriverAssigned,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    marshalledTrip,
		}); err != nil {
		return err
	}

	marshalledPayload, err := json.Marshal(messaging.PaymentTripResponseData{
		TripID:   tripID,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})

	if err := c.rabbitmq.PublishMessage(ctx,
		contracts.PaymentCmdCreateSession,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}

	return nil
}
