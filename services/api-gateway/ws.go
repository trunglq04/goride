package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/proto/driver"

	"github.com/gin-gonic/gin"
)

var connManager = messaging.NewConnectionManager()

func handleRidersWebSocket(c *gin.Context, rb *messaging.RabbitMQ) {
	ctx := c.Request.Context()
	log := logger.L()

	wsConn, err := connManager.Upgrade(c)
	if err != nil {
		log.WarnContext(ctx, "WebSocket upgrade failed", "role", "rider", "err", err)
		return
	}

	defer func(wsConn *websocket.Conn) {
		err := wsConn.Close()
		if err != nil {
			log.DebugContext(ctx, "WebSocket close failed", "role", "rider", "err", err)
		}
	}(wsConn)

	userID := c.Query("userID")
	if userID == "" {
		log.WarnContext(ctx, "Rider WebSocket connection rejected: no user ID provided")
		return
	}

	connManager.Add(userID, wsConn)
	defer connManager.Remove(userID)
	log.DebugContext(ctx, "Rider WebSocket connected", "user_id", userID)

	// Initialize queue consumers
	queues := []string{
		messaging.NotifyDriverNoDriversFoundQueue,  // When no drivers found
		messaging.NotifyDriverAssignQueue,          // When a driver is assigned to the trip
		messaging.NotifyPaymentSessionCreatedQueue, // When a payment session is created
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.ErrorContext(ctx, "Failed to start rider consumer",
				"queue", q,
				"user_id", userID,
				"err", err,
			)
		}
	}

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.InfoContext(ctx, "Rider WebSocket disconnected", "user_id", userID, "err", err)
			break
		}

		log.DebugContext(ctx, "Received message from rider", "user_id", userID, "body_size", len(message))
	}
}

func handleDriversWebSocket(c *gin.Context, rb *messaging.RabbitMQ) {
	ctx := c.Request.Context()
	log := logger.L()

	wsConn, err := connManager.Upgrade(c)
	if err != nil {
		log.WarnContext(ctx, "WebSocket upgrade failed", "role", "driver", "err", err)
		return
	}

	defer func(wsConn *websocket.Conn) {
		err := wsConn.Close()
		if err != nil {
			log.DebugContext(ctx, "WebSocket close failed", "role", "driver", "err", err)
		}
	}(wsConn)

	userID := c.Query("userID")
	if userID == "" {
		log.WarnContext(ctx, "Driver WebSocket connection rejected: no user ID provided")
		return
	}

	packageSlug := c.Query("packageSlug")
	if packageSlug == "" {
		log.WarnContext(ctx, "Driver WebSocket connection rejected: no package slug provided", "user_id", userID)
		return
	}

	connManager.Add(userID, wsConn)

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create driver service client", "user_id", userID, "err", err)
		connManager.Remove(userID)
		return
	}

	// Closing connections
	defer func() {
		connManager.Remove(userID)

		res, err := driverService.Client.UnregisterDriver(c, &driver.UnregisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		if err != nil {
			log.ErrorContext(ctx, "Failed to unregister driver",
				"user_id", userID,
				"package_slug", packageSlug,
				"err", err,
			)
		} else {
			log.InfoContext(ctx, "Driver unregistered",
				"user_id", res.Driver.Id,
				"package_slug", packageSlug,
			)
		}

		driverService.Close()
	}()

	driverData, err := driverService.Client.RegisterDriver(c, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to register driver",
			"user_id", userID,
			"package_slug", packageSlug,
			"err", err,
		)
		return
	}
	log.InfoContext(ctx, "Driver registered via WebSocket",
		"user_id", userID,
		"package_slug", packageSlug,
	)

	if err := connManager.SendMessage(userID,
		contracts.WSMessage{
			Type: contracts.DriverCmdRegister,
			Data: driverData.Driver,
		}); err != nil {
		log.ErrorContext(ctx, "Failed to send registration confirmation to driver",
			"user_id", userID,
			"err", err,
		)
		return
	}

	// Initialize queue consumers
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.ErrorContext(ctx, "Failed to start driver consumer",
				"queue", q,
				"user_id", userID,
				"err", err,
			)
		}
	}

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.InfoContext(ctx, "Driver WebSocket disconnected", "user_id", userID, "err", err)
			break
		}

		type driverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMsg driverMessage
		if err := json.Unmarshal(message, &driverMsg); err != nil {
			log.WarnContext(ctx, "Failed to unmarshal driver message",
				"user_id", userID,
				"body_size", len(message),
				"err", err,
			)
			continue
		}

		log.DebugContext(ctx, "Received message from driver",
			"user_id", userID,
			"type", driverMsg.Type,
			"body_size", len(message),
		)

		// Handle the different message types
		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			// TODO: Handle driver's location update in the future
			if err := rb.PublishMessage(c,
				driverMsg.Type,
				contracts.AmqpMessage{
					OwnerID: userID,
					Data:    driverMsg.Data,
				}); err != nil {
				log.ErrorContext(ctx, "Failed to publish driver location to RabbitMQ",
					"user_id", userID,
					"routing_key", driverMsg.Type,
					"err", err,
				)
			}
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			if err := rb.PublishMessage(c,
				driverMsg.Type,
				contracts.AmqpMessage{
					OwnerID: userID,
					Data:    driverMsg.Data,
				}); err != nil {
				log.ErrorContext(ctx, "Failed to publish trip response to RabbitMQ",
					"user_id", userID,
					"routing_key", driverMsg.Type,
					"err", err,
				)
			}
		default:
			log.WarnContext(ctx, "Unknown driver message type",
				"user_id", userID,
				"type", driverMsg.Type,
			)
		}
	}
}
