package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
)

const (
	// AuthExchange is the RabbitMQ exchange for auth events.
	AuthExchange = "auth"

	// RoutingKeyUserRegistered is published when a new user registers.
	RoutingKeyUserRegistered = "user.event.registered"

	// RoutingKeyOTPResent is published when a user requests a new OTP.
	RoutingKeyOTPResent = "user.event.otp_resent"
)

// UserRegisteredEvent is the payload published to RabbitMQ when a user registers.
type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	OTP    string `json:"otp"`
}

// AuthPublisher publishes auth events to RabbitMQ.
type AuthPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

// NewAuthPublisher creates a new auth event publisher.
func NewAuthPublisher(rabbitmq *messaging.RabbitMQ) *AuthPublisher {
	return &AuthPublisher{rabbitmq: rabbitmq}
}

// PublishUserRegistered publishes a user.event.registered event so the
// notification-service can send the OTP email.
func (p *AuthPublisher) PublishUserRegistered(ctx context.Context, userID, email, otp string) error {
	event := UserRegisteredEvent{
		UserID: userID,
		Email:  email,
		OTP:    otp,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user registered event: %w", err)
	}

	message := contracts.AmqpMessage{
		OwnerID: userID,
		Data:    eventBytes,
	}

	if err := p.rabbitmq.PublishMessageToExchange(ctx, messaging.AuthExchange, RoutingKeyUserRegistered, message); err != nil {
		slog.ErrorContext(ctx, "Failed to publish user registered event",
			"user_id", userID,
			"err", err,
		)
		return err
	}

	slog.InfoContext(ctx, "Published user registered event",
		"user_id", userID,
		"email", email,
	)
	return nil
}

// PublishOTPResent publishes a user.event.otp_resent event for OTP re-delivery.
func (p *AuthPublisher) PublishOTPResent(ctx context.Context, userID, email, otp string) error {
	event := UserRegisteredEvent{
		UserID: userID,
		Email:  email,
		OTP:    otp,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP resent event: %w", err)
	}

	message := contracts.AmqpMessage{
		OwnerID: userID,
		Data:    eventBytes,
	}

	if err := p.rabbitmq.PublishMessageToExchange(ctx, messaging.AuthExchange, RoutingKeyOTPResent, message); err != nil {
		slog.ErrorContext(ctx, "Failed to publish OTP resent event",
			"user_id", userID,
			"err", err,
		)
		return err
	}

	slog.InfoContext(ctx, "Published OTP resent event",
		"user_id", userID,
		"email", email,
	)
	return nil
}
