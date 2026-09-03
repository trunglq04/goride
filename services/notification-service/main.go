package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
)

// UserRegisteredEvent mirrors the event published by the auth service.
type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	OTP    string `json:"otp"`
}

func main() {
	logger.Setup("notification-service")
	log := logger.L()
	log.Info("Starting Notification Service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// RabbitMQ connection
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", "err", err)
	}
	defer rabbitmq.Close()
	log.Info("RabbitMQ connected")

	// Email sender
	emailSender := NewEmailSender()

	// Consume OTP email events
	err = rabbitmq.ConsumeMessages(messaging.SendEmailOTPQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		log.InfoContext(ctx, "Received email event",
			"routing_key", msg.RoutingKey,
		)

		// Unmarshal the AMQP envelope
		var amqpMsg contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &amqpMsg); err != nil {
			slog.ErrorContext(ctx, "Failed to unmarshal AMQP message", "err", err)
			return err
		}

		// Unmarshal the inner event data
		var event UserRegisteredEvent
		if err := json.Unmarshal(amqpMsg.Data, &event); err != nil {
			slog.ErrorContext(ctx, "Failed to unmarshal event data", "err", err)
			return err
		}

		log.InfoContext(ctx, "Processing OTP email",
			"user_id", event.UserID,
			"email", event.Email,
		)

		// Send the OTP email
		if err := emailSender.SendOTPEmail(ctx, event.Email, event.OTP); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.Fatal("Failed to start consuming messages", "err", err)
	}

	log.Info("Notification service listening for OTP events",
		"queue", messaging.SendEmailOTPQueue,
	)

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Info("Received shutdown signal")
	case <-ctx.Done():
	}

	log.Info("Notification service shut down")
}
