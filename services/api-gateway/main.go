package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.Println("Starting API Gateway")

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "api-gateway",
		Environment:    env.GetString("ENVIRONMENT", "developement"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "jaeger:4317"),
	}

	traceShutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to initialize the tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer traceShutdown(ctx)
	defer cancel()

	log.Println("Init tracing successfully!")

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(otelgin.Middleware(tracerCfg.ServiceName))

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	trip := router.Group("/trip")
	trip.POST("/preview", enableCORS(handleTripPreview))
	trip.POST("/start", enableCORS(handleTripStart))

	// WebSocket
	ws := router.Group("/ws")
	ws.GET("/drivers", func(c *gin.Context) { handleDriversWebSocket(c, rabbitmq) })
	ws.GET("/riders", func(c *gin.Context) { handleRidersWebSocket(c, rabbitmq) })

	// Webhook
	wh := router.Group("/webhook")
	wh.POST("/stripe", func(c *gin.Context) { handleStripeWebhook(c, rabbitmq) })

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
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
		log.Printf("Server stopped: %v", err)

	case sig := <-shutdown:
		log.Printf("Server is shutting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown Failed: %v", err)
			if cerr := server.Close(); cerr != nil {
				log.Printf("Forced server close Failed: %v", cerr)
			}
		} else {
			log.Printf("Server shut down gracefully")
		}
	}
}
