package events

import (
	"context"
	"encoding/json"

	"github.com/trunglq04/goride/services/trip-service/internal/domain"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	msg := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	tripEventJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    tripEventJSON,
	})
}
