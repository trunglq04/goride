package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/trunglq04/goride/services/trip-service/internal/infrastructure/events"
	"github.com/trunglq04/goride/services/trip-service/internal/infrastructure/grpc"
	"github.com/trunglq04/goride/services/trip-service/internal/infrastructure/repository"
	"github.com/trunglq04/goride/services/trip-service/internal/service"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "trip-service",
		Environment:    env.GetString("ENVIRONMENT", "developement"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:4318"),
	}

	traceShutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to initialize the tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer traceShutdown(ctx)
	defer cancel()

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("ERROR: Failed to listen: %v", err)
	}

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	inmemRepo := repository.NewInmemReposity()
	svc := service.NewService(inmemRepo)

	publisher := events.NewTripEventPublisher(rabbitmq)

	// Start driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go func() {
		if err := driverConsumer.Listen(); err != nil {
			log.Fatalf("ERROR: Failed to listen to the message: %v", err)
		}
	}()

	// Start payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go func() {
		if err := paymentConsumer.Listen(); err != nil {
			log.Fatalf("ERROR: Failed to listen to the message: %v", err)
		}
	}()

	// Starting the gRPC server
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpc.NewGRPCHandler(grpcServer, svc, publisher)

	log.Printf("Starting gRPC server Trip service on port %s", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("ERROR: Failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Printf("Shutting down the server...")
	grpcServer.GracefulStop()

}

//Setup main with graceful shutdown (HTTP)
/*
func main() {
	inmemRepo := repository.NewInmemReposity()
	svc := service.NewService(inmemRepo)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	httpHandler := h.HttpHandler{Service: svc}

	router.POST("/preview", httpHandler.HandleTripPreview)

	server := &http.Server{
		Addr:    ":8083",
		Handler: router,
	}

	errorsChan := make(chan error, 1)

	go func() {
		errorsChan <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errorsChan:
		log.Printf("Error starting the server: %v", err)
	case sig := <-shutdown:
		log.Printf("Server is shuting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not shut down the server gracefully: %v", err)
			server.Close()
		}
	}
}
*/
