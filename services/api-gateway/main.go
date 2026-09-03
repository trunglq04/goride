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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
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
		ServiceName:    "api-gateway",
		Environment:    env.GetString("ENVIRONMENT", "developement"),
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

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(otelgin.Middleware(tracerCfg.ServiceName))
	router.Use(metrics.MetricMiddleware())
	corsConfig(router)

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

	// ---- Public auth routes (no JWT required) ----
	authGroup := router.Group("/auth")
	authGroup.Use(authRateLimiter())
	authGroup.POST("/register", handleRegister)
	authGroup.POST("/login", handleLogin)
	authGroup.POST("/verify-otp", handleVerifyOTP)
	authGroup.POST("/resend-otp", handleResendOTP)
	authGroup.POST("/refresh", handleRefreshToken)
	authGroup.POST("/logout", handleLogout)

	// ---- Protected auth routes (JWT required) ----
	authProtected := router.Group("/auth")
	authProtected.Use(jwtMiddleware)
	authProtected.GET("/me", handleGetMe)

	// ---- Protected trip routes ----
	trip := router.Group("/trip")
	trip.Use(jwtMiddleware)
	trip.POST("/preview", handleTripPreview)
	trip.POST("/start", handleTripStart)
	trip.POST("/cancel", handleTripCancel)

	// WebSocket (JWT validated via query param or connection upgrade)
	ws := router.Group("/ws")
	ws.GET("/drivers", func(c *gin.Context) { handleDriversWebSocket(c, rabbitmq) })
	ws.GET("/riders", func(c *gin.Context) { handleRidersWebSocket(c, rabbitmq) })

	// Webhook (Stripe validates via its own signature)
	wh := router.Group("/webhook")
	wh.POST("/stripe", func(c *gin.Context) { handleStripeWebhook(c, rabbitmq) })

	server := &http.Server{
		Addr:    httpAddr,
		Handler: router,
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
