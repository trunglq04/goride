package main

import (
	"context"

	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
)

type EventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewEventPublisher(rabbitmq *messaging.RabbitMQ) *EventPublisher {
	return &EventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *EventPublisher) PublishTripRequest(ctx context.Context, driverID string, payloadBytes []byte) error {
	return p.rabbitmq.PublishMessage(ctx,
		contracts.DriverCmdTripRequest,
		contracts.AmqpMessage{
			OwnerID: driverID,
			Data:    payloadBytes,
		},
	)
}

func (p *EventPublisher) PublishTripCanceled(ctx context.Context, driverID string, payloadBytes []byte) error {
	return p.rabbitmq.PublishMessage(ctx,
		contracts.DriverCmdTripCanceled,
		contracts.AmqpMessage{
			OwnerID: driverID,
			Data:    payloadBytes,
		},
	)
}

func (p *EventPublisher) PublishNoDriversFound(ctx context.Context, userID string, payloadBytes []byte) error {
	return p.rabbitmq.PublishMessage(ctx,
		contracts.TripEventNoDriversFound,
		contracts.AmqpMessage{
			OwnerID: userID,
			Data:    payloadBytes,
		},
	)
}
