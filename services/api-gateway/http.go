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
	"github.com/trunglq04/goride/shared/util"
)

func handleTripStart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var reqBody startTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.WarnContext(ctx, "Failed to parse trip start request body", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Failed to parse JSON data")
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create trip service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Failed to reach trip service")
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
		util.WriteError(w, http.StatusBadGateway, "Failed to call trip start")
		return
	}

	log.InfoContext(ctx, "Trip started",
		"trip_id", tripStart.TripID,
		"user_id", reqBody.UserID,
		"ride_fare_id", reqBody.RideFareID,
	)

	util.WriteJSON(w, http.StatusCreated, contracts.APIResponse{Data: tripStart})
}

func handleTripCancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var reqBody cancelTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.WarnContext(ctx, "Failed to parse trip cancel request body", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Failed to parse JSON data")
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create trip service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Failed to reach trip service")
		return
	}
	defer tripService.Close()

	cancelRes, err := tripService.Client.CancelTrip(ctx, reqBody.toProto())
	if err != nil {
		log.ErrorContext(ctx, "Failed to call trip cancel",
			"user_id", reqBody.UserID,
			"trip_id", reqBody.TripID,
			"err", err,
		)
		util.WriteError(w, http.StatusBadGateway, "Failed to call trip cancel")
		return
	}

	log.InfoContext(ctx, "Trip canceled via HTTP",
		"trip_id", cancelRes.TripID,
		"user_id", reqBody.UserID,
	)

	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: cancelRes})
}

func handleTripPreview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var reqBody previewTripRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.WarnContext(ctx, "Failed to parse trip preview request body", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Failed to parse JSON data")
		return
	}

	// validation
	if reqBody.UserID == "" {
		log.WarnContext(ctx, "Trip preview request is missing user ID")
		util.WriteError(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// so we create a new client for each connection to avoid server crashing
	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create trip service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Failed to reach trip service")
		return
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(ctx, reqBody.toProto())
	if err != nil {
		log.ErrorContext(ctx, "Failed to call trip preview",
			"user_id", reqBody.UserID,
			"err", err,
		)
		util.WriteError(w, http.StatusBadGateway, "Failed to call trip preview")
		return
	}

	log.InfoContext(ctx, "Trip preview generated",
		"user_id", reqBody.UserID,
		"fares", len(tripPreview.RideFares),
	)

	util.WriteJSON(w, http.StatusCreated, contracts.APIResponse{Data: tripPreview})
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	ctx := r.Context()
	log := logger.L()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.ErrorContext(ctx, "Failed to read webhook body", "err", err)
		util.WriteError(w, http.StatusInternalServerError, "Failed to read body")
		return
	}
	defer r.Body.Close()

	webhookKey := env.GetString("STRIPE_WEBHOOK_KEY", "")
	if webhookKey == "" {
		log.Error("Stripe webhook key is not configured", "env", "STRIPE_WEBHOOK_KEY")
		util.WriteError(w, http.StatusInternalServerError, "Webhook is not configured")
		return
	}

	event, err := webhook.ConstructEventWithOptions(
		body,
		r.Header.Get("Stripe-Signature"),
		webhookKey,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		log.WarnContext(ctx, "Invalid Stripe webhook signature", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Invalid signature")
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
			util.WriteError(w, http.StatusBadRequest, "Invalid payload")
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
			util.WriteError(w, http.StatusInternalServerError, "Failed to marshal payload")
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
			util.WriteError(w, http.StatusInternalServerError, "Failed to publish payment event")
			return
		}

		util.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
	default:
		// Acknowledge all other event types — returning non-2xx causes Stripe to retry
		log.DebugContext(ctx, "Unhandled Stripe event type ignored",
			"event_type", event.Type,
			"event_id", event.ID,
		)
		util.WriteJSON(w, http.StatusOK, map[string]bool{"received": true})
	}
}
