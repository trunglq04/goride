package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/trunglq04/goride/services/payment-service/internal/infrastructure/events"
	"github.com/trunglq04/goride/services/payment-service/internal/infrastructure/stripe"
	"github.com/trunglq04/goride/services/payment-service/internal/service"
	"github.com/trunglq04/goride/services/payment-service/pkg/types"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/metrics"
	"github.com/trunglq04/goride/shared/tracing"
)

var GrpcAddr = env.GetString("GRPC_ADDR", ":9004")

func main() {
	logger.Setup("payment-service")
	log := logger.L()

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:      "payment-service",
		Environment:      env.GetString("ENVIRONMENT", "developement"),
		ExporterEndpoint: env.GetString("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
	}

	traceShutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		logger.Fatal("Failed to initialize the tracer", "err", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer traceShutdown(ctx)
	defer cancel()

	// Initialize Prometheus metrics
	metrics.Init("payment-service")
	metrics.StartMetricsServer("payment", ":9091")

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	// Setup graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	appURL := env.GetString("APP_URL", "localhost:3000")

	stripeCfg := &types.PaymentConfig{
		StripeSecretKey: env.GetString("STRIPE_SECRET_KEY", ""),
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", appURL+"?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", appURL+"?payment=cancel"),
	}
	if stripeCfg.StripeSecretKey == "" {
		logger.Fatal("STRIPE_SECRET_KEY is not set")
		return
	}

	// Stripe processor
	paymentProcessor := stripe.NewStripeClient(stripeCfg)

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", "err", err)
	}
	defer rabbitmq.Close()

	log.Info("RabbitMQ connected")

	// Event publisher
	publisher := events.NewPaymentEventPublisher(rabbitmq)

	// Service
	svc := service.NewPaymentService(paymentProcessor, publisher)

	// Trip Consumer
	tripConsumer := events.NewTripConsumer(rabbitmq, svc)
	go tripConsumer.Listen()

	log.Info("Payment service started")

	// wait for the shutdown signal
	<-ctx.Done()
	log.Info("Shutting down payment-service...")
}
