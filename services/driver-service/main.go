package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {
	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "driver-service",
		Environment:    env.GetString("ENVIRONMENT", "developement"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "jaeger:14268/api/traces"),
	}

	traceShutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to initialize the tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer traceShutdown(ctx)
	defer cancel()

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	// Handle graceful shutdown
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

	svc := NewService()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("Starting RabbitMQ connection")

	// Starting the gRPC server
	grpcServer := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	NewGrpcHandler(grpcServer, svc)

	tripConsumer := NewTripConsumer(rabbitmq, svc)
	go func() {
		if err := tripConsumer.Listen(); err != nil {
			log.Fatalf("ERROR: Failed to listen to the message: %v", err)
		}
	}()

	log.Printf("Starting gRPC server Driver service on port %s", lis.Addr().String())

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
