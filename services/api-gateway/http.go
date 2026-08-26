package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"

	"github.com/gin-gonic/gin"
)

func handleTripStart(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var reqBody startTripRequest
	if err := c.ShouldBindBodyWithJSON(&reqBody); err != nil {
		log.WarnContext(ctx, "Failed to parse trip start request body", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON data"})
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create trip service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach trip service"})
		return
	}

	defer tripService.Close()

	tripStart, err := tripService.Client.CreateTrip(ctx, reqBody.toProto())
	if err != nil {
		log.ErrorContext(ctx, "Failed to call trip start",
			"user_id", reqBody.UserID,
			"ride_fare_id", reqBody.RideFareID,
			"err", err,
		)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to call trip start"})
		return
	}

	log.InfoContext(ctx, "Trip started",
		"trip_id", tripStart.TripID,
		"user_id", reqBody.UserID,
		"ride_fare_id", reqBody.RideFareID,
	)

	response := contracts.APIResponse{Data: tripStart}

	c.JSON(http.StatusCreated, response)

}

func handleTripPreview(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var reqBody previewTripRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		log.WarnContext(ctx, "Failed to parse trip preview request body", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON data"})
		return
	}

	// validation
	if reqBody.UserID == "" {
		log.WarnContext(ctx, "Trip preview request is missing user ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	// so we create a new client for each connection to avoid server crashing
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create trip service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to reach trip service"})
		return
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(ctx, reqBody.toProto())
	if err != nil {
		log.ErrorContext(ctx, "Failed to call trip preview",
			"user_id", reqBody.UserID,
			"err", err,
		)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to call trip preview"})
		return
	}

	log.InfoContext(ctx, "Trip preview generated",
		"user_id", reqBody.UserID,
		"fares", len(tripPreview.RideFares),
	)

	response := contracts.APIResponse{Data: tripPreview}

	c.JSON(http.StatusCreated, response)
}

func handleStripeWebhook(c *gin.Context, rb *messaging.RabbitMQ) {
	ctx := c.Request.Context()
	log := logger.L()

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.ErrorContext(ctx, "Failed to read webhook body", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read body"})
		return
	}
	defer c.Request.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webhookKey == "" {
		log.Error("Stripe webhook key is not configured", "env", "STRIPE_WEBHOOK_KEY")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook is not configured"})
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		body,
		c.GetHeader("Stripe-Signature"),
		webhookKey,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Invalid Stripe webhook signature", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	log.InfoContext(ctx, "Received Stripe event",
		"event_type", event.Type,
		"event_id", event.ID,
	)

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.ErrorContext(ctx, "Failed to parse checkout session payload",
				"event_id", event.ID,
				"err", err,
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}
		log.InfoContext(ctx, "Checkout session completed",
			"event_id", event.ID,
			"session_id", session.ID,
			"trip_id", payload.TripID,
			"user_id", payload.UserID,
			"driver_id", payload.DriverID,
		)

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.ErrorContext(ctx, "Failed to marshal payment payload",
				"event_id", event.ID,
				"err", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}

		message := contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    payloadBytes,
		}

		if err := rb.PublishMessage(
			ctx,
			contracts.PaymentEventSuccess,
			message,
		); err != nil {
			log.ErrorContext(ctx, "Failed to publish payment success event",
				"event_id", event.ID,
				"routing_key", contracts.PaymentEventSuccess,
				"err", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish payment event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	default:
		// Acknowledge all other event types — returning non-2xx causes Stripe to retry
		log.DebugContext(ctx, "Unhandled Stripe event type ignored",
			"event_type", event.Type,
			"event_id", event.ID,
		)
		c.JSON(http.StatusOK, gin.H{"received": true})
	}
}
