package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/metrics"
	"github.com/trunglq04/goride/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {
	logger.Setup("driver-service")
	log := logger.L()

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:    "driver-service",
		Environment:    env.GetString("ENVIRONMENT", "developement"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "jaeger:14268/api/traces"),
	}

	traceShutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		logger.Fatal("Failed to initialize the tracer", "err", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer traceShutdown(ctx)
	defer cancel()

	// Initialize Prometheus metrics
	metrics.Init("driver-service")
	metrics.StartMetricsServer(":9091")

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
		logger.Fatal("Failed to listen", "addr", GrpcAddr, "err", err)
	}

	svc := NewService()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", "err", err)
	}
	defer rabbitmq.Close()

	log.Info("RabbitMQ connected")

	// Starting the gRPC server
	grpcServer := grpcserver.NewServer(append(
		tracing.WithTracingInterceptors(),
		grpcserver.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			logger.GrpcUnaryServerInterceptor(),
		),
		grpcserver.ChainStreamInterceptor(logger.GrpcStreamServerInterceptor()),
	)...)
	NewGrpcHandler(grpcServer, svc)

	tripConsumer := NewTripConsumer(rabbitmq, svc)
	go func() {
		if err := tripConsumer.Listen(); err != nil {
			logger.Fatal("Failed to listen for trip messages", "err", err)
		}
	}()

	log.Info("Starting gRPC server Driver service", "addr", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed to serve", "err", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Info("Shutting down driver-service...")
	grpcServer.GracefulStop()
}
