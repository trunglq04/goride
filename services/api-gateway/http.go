package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v86"
	"github.com/stripe/stripe-go/v86/webhook"
	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/messaging"

	"github.com/gin-gonic/gin"
)

func handleTripStart(c *gin.Context) {
	var reqBody startTripRequest
	if err := c.ShouldBindBodyWithJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON data"})
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	defer tripService.Close()

	tripStart, err := tripService.Client.CreateTrip(c, reqBody.toProto())
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to call trip start"})
		return
	}

	response := contracts.APIResponse{Data: tripStart}

	c.JSON(http.StatusCreated, response)

}

func handleTripPreview(c *gin.Context) {
	var reqBody previewTripRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON data"})
		return
	}

	// validation
	if reqBody.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	// Why we need to create a new client for each connection:
	// because if a service is down, we don't want to block the whole application
	// so we create a new client for each connection
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Fatal(err)
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(c, reqBody.toProto())
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to call trip preview"})
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	c.JSON(http.StatusCreated, response)
}

func handleStripeWebhook(c *gin.Context, rb *messaging.RabbitMQ) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read body"})
		return
	}
	defer c.Request.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webhookKey == "" {
		log.Printf("Webhook key is required")
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
		log.Printf("Error verifying webhook signature: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	log.Printf("Received Stripe event: %v", event)

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		payload := messaging.PaymentStatusUpdateData{
			TripID:   session.Metadata["trip_id"],
			UserID:   session.Metadata["user_id"],
			DriverID: session.Metadata["driver_id"],
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling payload: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal payload"})
			return
		}

		message := contracts.AmqpMessage{
			OwnerID: session.Metadata["user_id"],
			Data:    payloadBytes,
		}

		if err := rb.PublishMessage(
			c,
			contracts.PaymentEventSuccess,
			message,
		); err != nil {
			log.Printf("Error publishing payment event: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish payment event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	}
}
