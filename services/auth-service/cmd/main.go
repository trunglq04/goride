package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/trunglq04/goride/services/auth-service/internal/infrastructure/events"
	authgrpc "github.com/trunglq04/goride/services/auth-service/internal/infrastructure/grpc"
	"github.com/trunglq04/goride/services/auth-service/internal/infrastructure/repository"
	"github.com/trunglq04/goride/services/auth-service/internal/service"
	"github.com/trunglq04/goride/shared/auth"
	"github.com/trunglq04/goride/shared/db"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/metrics"
	"github.com/trunglq04/goride/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var grpcAddr = ":9094"

func main() {
	logger.Setup("auth-service")
	log := logger.L()
	log.Info("Starting Auth Service", "grpc_addr", grpcAddr)

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:      "auth-service",
		Environment:      env.GetString("ENVIRONMENT", "developement"),
		ExporterEndpoint: env.GetString("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
	}

	traceShutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		logger.Fatal("Failed to initialize the tracer", "err", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer traceShutdown(ctx)

	// Initialize Prometheus metrics
	metrics.Init("auth-service")
	metrics.StartMetricsServer("auth", ":9091")

	// Load RSA private key for JWT signing
	privateKeyPath := env.GetString("JWT_PRIVATE_KEY_PATH", "/etc/secrets/jwt_private.pem")
	privateKey, err := auth.LoadPrivateKey(privateKeyPath)
	if err != nil {
		logger.Fatal("Failed to load RSA private key", "path", privateKeyPath, "err", err)
	}
	log.Info("RSA private key loaded", "path", privateKeyPath)

	// Init PostgreSQL
	pgCfg := db.NewPostgresDefaultConfig()
	pgDB, err := db.NewPostgresClient(pgCfg)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", "err", err)
	}
	defer pgDB.Close()

	// RabbitMQ connection
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", "err", err)
	}
	defer rabbitmq.Close()
	log.Info("RabbitMQ connected")

	// Wire up dependencies
	repo := repository.NewPostgresRepository(pgDB)
	publisher := events.NewAuthPublisher(rabbitmq)
	svc := service.NewAuthService(repo, publisher, privateKey)

	// Signal handling
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("Failed to listen", "addr", grpcAddr, "err", err)
	}

	// Start gRPC server
	grpcServer := grpcserver.NewServer(append(
		tracing.WithTracingInterceptors(),
		grpcserver.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			logger.GrpcUnaryServerInterceptor(),
		),
		grpcserver.ChainStreamInterceptor(logger.GrpcStreamServerInterceptor()),
	)...)
	authgrpc.NewGRPCHandler(grpcServer, svc)

	log.Info("Starting gRPC server Auth service", "addr", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed to serve", "err", err)
			cancel()
		}
	}()

	// Wait for shutdown
	<-ctx.Done()
	log.Info("Shutting down auth-service...")
	grpcServer.GracefulStop()
}
