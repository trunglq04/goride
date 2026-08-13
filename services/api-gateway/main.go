package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/messaging"

	"github.com/gin-gonic/gin"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Starting API Gateway")

	// gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	trip := router.Group("/trip")
	trip.POST("/preview", enableCORSHandler(handleTripPreview))
	trip.POST("/start", enableCORSHandler(handleTripStart))

	ws := router.Group("/ws")
	ws.GET("/drivers", func(c *gin.Context) { handleDriversWebSocket(c, rabbitmq) })
	ws.GET("/riders", func(c *gin.Context) { handleRidersWebSocket(c, rabbitmq) })

	server := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server listening on %s", httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
		log.Printf("Server stopped: %v", err)

	case sig := <-shutdown:
		log.Printf("Server is shuting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if cerr := server.Close(); cerr != nil {
				log.Printf("Forced server close failed: %v", cerr)
			}
		} else {
			log.Printf("Server shut down gracefully")
		}
	}
}
