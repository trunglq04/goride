package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/proto/driver"

	"github.com/gin-gonic/gin"
)

var connManager = messaging.NewConnectionManager()

func handleRidersWebSocket(c *gin.Context, rb *messaging.RabbitMQ) {
	wsConn, err := connManager.Upgrade(c)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	defer func(wsConn *websocket.Conn) {
		err := wsConn.Close()
		if err != nil {
			log.Printf("WebSocket close failed: %v", err)
		}
	}(wsConn)

	userID := c.Query("userID")
	if userID == "" {
		log.Println("No user ID provided")
		return
	}

	connManager.Add(userID, wsConn)
	defer connManager.Remove(userID)

	// Initialize queue consumers
	queues := []string{
		messaging.NotifyDriverNoDriversFoundQueue,  // When no drivers found
		messaging.NotifyDriverAssignQueue,          // When a driver is assigned to the trip
		messaging.NotifyPaymentSessionCreatedQueue, // When a payment session is created
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("ERROR: Failed to start rider consumer for queue: %s: err: %v", q, err)
		}
	}

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("(Rider) Error reading message: %v", err)
			break
		}

		log.Printf("Received message: %s", message)
	}
}

func handleDriversWebSocket(c *gin.Context, rb *messaging.RabbitMQ) {
	wsConn, err := connManager.Upgrade(c)
	if err != nil {
		log.Printf("WebSocket upgrade Failed: %v", err)
		return
	}

	defer func(wsConn *websocket.Conn) {
		err := wsConn.Close()
		if err != nil {
			log.Printf("WebSocket close Failed: %v", err)
		}
	}(wsConn)

	userID := c.Query("userID")
	if userID == "" {
		log.Println("No user Id provided")
		return
	}

	packageSlug := c.Query("packageSlug")
	if packageSlug == "" {
		log.Println("No package slug provided")
		return
	}

	connManager.Add(userID, wsConn)

	driverService, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	// Closing connections
	defer func() {
		connManager.Remove(userID)

		res, err := driverService.Client.UnregisterDriver(c, &driver.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		if err != nil {
			log.Printf("Error unregistering driver: %v", err)
		}

		driverService.Close()

		log.Println("Driver unregistered: ", res.Driver.Id)
	}()

	driverData, err := driverService.Client.RegisterDriver(c, &driver.RegisterDriverRequest{
		DriverID:    userID,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Error registering driver: %v", err)
		return
	}

	if err := connManager.SendMessage(userID,
		contracts.WSMessage{
			Type: contracts.DriverCmdRegister,
			Data: driverData.Driver,
		}); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	// Initialize queue consumers
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)

		if err := consumer.Start(); err != nil {
			log.Printf("ERROR: Failed to start driver consumer for queue: %s: err: %v", q, err)
		}
	}

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("(Driver) Error reading message: %v", err)
			break
		}

		type driverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMsg driverMessage
		if err := json.Unmarshal(message, &driverMsg); err != nil {
			log.Printf("Error unmarshalling driver message: %v", err)
			continue
		}

		// Handle the different message types
		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			// TODO: Handle driver's location update in the future
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			if err := rb.PublishMessage(c,
				driverMsg.Type,
				contracts.AmqpMessage{
					OwnerID: userID,
					Data:    driverMsg.Data,
				}); err != nil {
				log.Printf("Error publishing message to RabbitMQ: %v", err)
			}
		default:
			log.Printf("Unknown message type: %s", driverMsg.Type)
		}
	}
}
