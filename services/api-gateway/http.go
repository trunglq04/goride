package main

import (
	"log"
	"net/http"

	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/contracts"

	"github.com/gin-gonic/gin"
)

func handleTripStart(c *gin.Context) {
	var reqBody startTripRequest
	if err := c.ShouldBindBodyWithJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse JSON data"})
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
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to call trip start"})
		return
	}

	response := contracts.APIResponse{Data: tripStart}

	c.JSON(http.StatusCreated, response)

}

func handleTripPreview(c *gin.Context) {
	var reqBody previewTripRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse JSON data"})
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

	// avoid resource leaks
	defer tripService.Close()

	// resp, err := http.Post("http://trip-service:8083/preview", "application/json", reader)
	tripPreview, err := tripService.Client.PreviewTrip(c, reqBody.toProto())
	if err != nil {
		log.Print(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to call trip preview"})
		return
	}

	response := contracts.APIResponse{Data: tripPreview}

	c.JSON(http.StatusCreated, response)
}
