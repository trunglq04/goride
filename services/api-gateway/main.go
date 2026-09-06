package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trunglq04/goride/shared/auth"
	"github.com/trunglq04/goride/shared/env"
	"github.com/trunglq04/goride/shared/logger"
	"github.com/trunglq04/goride/shared/messaging"
	"github.com/trunglq04/goride/shared/metrics"
	"github.com/trunglq04/goride/shared/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	logger.Setup("api-gateway")
	log := logger.L()
	log.Info("Starting API Gateway", "http_addr", httpAddr)

	// Initialize Tracing
	tracerCfg := tracing.Config{
		ServiceName:      "api-gateway",
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

	log.Info("Tracing initialized successfully")

	// Initialize Prometheus metrics
	metrics.Init("api-gateway")
	metrics.StartMetricsServer("api-gateway", ":9091")

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", "err", err)
	}
	defer rabbitmq.Close()

	log.Info("RabbitMQ connected")

	// Load RSA public key for JWT validation
	publicKeyPath := env.GetString("JWT_PUBLIC_KEY_PATH", "/etc/secrets/jwt_public.pem")
	publicKey, err := auth.LoadPublicKey(publicKeyPath)
	if err != nil {
		logger.Fatal("Failed to load RSA public key", "path", publicKeyPath, "err", err)
	}
	log.Info("RSA public key loaded", "path", publicKeyPath)

	// JWT authentication middleware
	jwtMiddleware := auth.JWTAuthMiddleware(publicKey)

	// Global middleware stack
	globalMiddleware := []func(http.Handler) http.Handler{
		recoveryMiddleware(),
		requestLogger(),
		corsMiddleware(),
		metrics.MetricMiddleware(),
	}

	mux := http.NewServeMux()

	// ---- Public auth routes (no JWT required) ----
	authPublic := chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/auth/register":
				handleRegister(w, r)
			case "/auth/login":
				handleLogin(w, r)
			case "/auth/verify-otp":
				handleVerifyOTP(w, r)
			case "/auth/resend-otp":
				handleResendOTP(w, r)
			case "/auth/refresh":
				handleRefreshToken(w, r)
			case "/auth/logout":
				handleLogout(w, r)
			default:
				http.NotFound(w, r)
			}
		}),
		authRateLimiter(),
	)
	mux.Handle("/auth/register", authPublic)
	mux.Handle("/auth/login", authPublic)
	mux.Handle("/auth/verify-otp", authPublic)
	mux.Handle("/auth/resend-otp", authPublic)
	mux.Handle("/auth/refresh", authPublic)
	mux.Handle("/auth/logout", authPublic)

	// ---- Protected auth routes (JWT required) ----
	mux.Handle("/auth/me", chain(http.HandlerFunc(handleGetMe), jwtMiddleware))

	// ---- Protected trip routes (JWT required) ----
	mux.Handle("/trip/preview", chain(http.HandlerFunc(handleTripPreview), jwtMiddleware))
	mux.Handle("/trip/start", chain(http.HandlerFunc(handleTripStart), jwtMiddleware))
	mux.Handle("/trip/cancel", chain(http.HandlerFunc(handleTripCancel), jwtMiddleware))

	// ---- WebSocket routes ----
	mux.Handle("/ws/drivers", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDriversWebSocket(w, r, rabbitmq)
	}))
	mux.Handle("/ws/riders", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebSocket(w, r, rabbitmq)
	}))

	// ---- Webhook (Stripe validates via its own signature) ----
	mux.Handle("/webhook/stripe", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	}))

	// Wrap the entire mux with global middleware and OTel HTTP instrumentation
	handler := chain(mux, globalMiddleware...)
	handler = otelhttp.NewHandler(handler, tracerCfg.ServiceName)

	server := &http.Server{
		Addr:    httpAddr,
		Handler: handler,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("HTTP server listening", "addr", httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Server error", "err", err)
		}
		log.Info("Server stopped", "err", err)

	case sig := <-shutdown:
		log.Info("Server is shutting down due to signal", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error("Graceful shutdown failed", "err", err)
			if cerr := server.Close(); cerr != nil {
				log.Error("Forced server close failed", "err", cerr)
			}
		} else {
			log.Info("Server shut down gracefully")
		}
	}
}
