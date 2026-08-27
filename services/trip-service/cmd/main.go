package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/trunglq04/goride/services/trip-service/internal/infrastructure/events"
	"github.com/trunglq04/goride/services/trip-service/internal/infrastructure/grpc"
	"github.com/trunglq04/goride/services/trip-service/internal/infrastructure/repository"
	"github.com/trunglq04/goride/services/trip-service/internal/service"
	"github.com/trunglq04/goride/shared/db"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/metrics"
	"github.com/trunglq04/goride/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	logger.Setup("trip-service")
	log := logger.L()

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:      "trip-service",
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
	metrics.Init("trip-service")
	metrics.StartMetricsServer("trip", ":9091")

	// Init MongoDB
	mongoClient, err := db.NewMongoClient(ctx, db.NewMongoDefaultConfig())
	if err != nil {
		logger.Fatal("Failed to initialize MongoDB", "err", err)
	}
	defer mongoClient.Disconnect(ctx)

	mongodb := db.GetDatabase(mongoClient, db.NewMongoDefaultConfig())

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	mongoDBRepo := repository.NewMongoRepository(mongodb)
	if err := mongoDBRepo.EnsureIndexes(ctx); err != nil {
		logger.Fatal("Failed to ensure indexes", "err", err)
	}

	svc := service.NewService(mongoDBRepo)

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

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", "err", err)
	}
	defer rabbitmq.Close()
	log.Info("RabbitMQ connected")

	// Start trip publisher
	publisher := events.NewTripEventPublisher(rabbitmq)

	// Start driver consumer
	driverConsumer := events.NewDriverConsumer(rabbitmq, svc)
	go func() {
		if err := driverConsumer.Listen(); err != nil {
			logger.Fatal("Failed to listen for driver messages", "err", err)
		}
	}()

	// Start payment consumer
	paymentConsumer := events.NewPaymentConsumer(rabbitmq, svc)
	go func() {
		if err := paymentConsumer.Listen(); err != nil {
			logger.Fatal("Failed to listen for payment messages", "err", err)
		}
	}()

	// Starting the gRPC server
	grpcServer := grpcserver.NewServer(append(
		tracing.WithTracingInterceptors(),
		grpcserver.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			logger.GrpcUnaryServerInterceptor(),
		),
		grpcserver.ChainStreamInterceptor(logger.GrpcStreamServerInterceptor()),
	)...)
	grpc.NewGRPCHandler(grpcServer, svc, publisher)

	log.Info("Starting gRPC server Trip service", "addr", lis.Addr().String())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server failed to serve", "err", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Info("Shutting down trip-service...")
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
