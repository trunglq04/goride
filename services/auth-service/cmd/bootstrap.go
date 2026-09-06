package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trunglq04/goride/services/auth-service/internal/domain"
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

// App gom toàn bộ dependency đã khởi tạo của service.
type App struct {
	grpcServer *grpcserver.Server
	grpcAddr   string
	database   *sql.DB
	rabbitmq   *messaging.RabbitMQ
	svc        domain.AuthService
}

// Setup khởi tạo toàn bộ dependency theo đúng thứ tự.
// Trả về App, 1 hàm shutdown gộp (LIFO), và error nếu có bước fail.
func Setup(ctx context.Context) (*App, func(), error) {
	app := &App{}
	var cleanups []func()

	// ---- Tracing ----
	traceShutdown, err := setupTracing()
	if err != nil {
		return nil, nil, err
	}
	cleanups = append(cleanups, traceShutdown)

	// ---- Metrics ----
	setupMetrics()

	// ---- JWT private key ----
	privateKey, err := loadJWTKey()
	if err != nil {
		return nil, nil, err
	}

	// ---- PostgreSQL ----
	pgDB, err := setupPostgres()
	if err != nil {
		return nil, nil, err
	}
	app.database = pgDB
	cleanups = append(cleanups, func() { pgDB.Close() })

	// ---- RabbitMQ ----
	rabbitmq, err := setupRabbitMQ()
	if err != nil {
		return nil, nil, err
	}
	app.rabbitmq = rabbitmq
	cleanups = append(cleanups, func() { rabbitmq.Close() })

	// ---- Dependency wiring ----
	repo := repository.NewPostgresRepository(pgDB)
	publisher := events.NewAuthPublisher(rabbitmq)
	app.svc = service.NewAuthService(repo, publisher, privateKey)

	// ---- gRPC server ----
	app.grpcAddr = env.GetString("GRPC_ADDR", ":9094")
	app.grpcServer = grpcserver.NewServer(append(tracing.WithTracingInterceptors(),
		grpcserver.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			logger.GrpcUnaryServerInterceptor(),
		),
		grpcserver.ChainStreamInterceptor(logger.GrpcStreamServerInterceptor()),
	)...)
	authgrpc.NewGRPCHandler(app.grpcServer, app.svc)

	shutdown := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	return app, shutdown, nil
}

// ---- Run: serve, chờ signal, graceful shutdown ----

func (a *App) Run() error {
	log := logger.L()

	lis, err := net.Listen("tcp", a.grpcAddr)
	if err != nil {
		return err
	}

	log.Info("gRPC server listening", "addr", lis.Addr().String())

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- a.grpcServer.Serve(lis)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		log.Info("Received signal, shutting down", "signal", sig)
	case err := <-serverErr:
		log.Error("gRPC server error", "err", err)
	}

	log.Info("Shutting down auth-service...")
	stopped := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(stopped)
	}()

	const shutdownTimeout = 10 * time.Second
	select {
	case <-stopped:
		log.Info("gRPC server stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.Warn("Graceful stop timed out, forcing shutdown", "timeout", shutdownTimeout)
		a.grpcServer.Stop()
	}

	return nil
}

// ---- Các hàm setup/load riêng lẻ ----

func setupTracing() (func(), error) {
	cfg := tracing.Config{
		ServiceName:      "auth-service",
		Environment:      env.GetString("ENVIRONMENT", "development"),
		ExporterEndpoint: env.GetString("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
	}

	traceShutdown, err := tracing.InitTracer(cfg)
	if err != nil {
		return nil, err
	}

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		traceShutdown(ctx)
	}, nil
}

func setupMetrics() {
	metrics.Init("auth-service")
	metrics.StartMetricsServer("auth-service", ":9091")
}

func loadJWTKey() (*rsa.PrivateKey, error) {
	path := env.GetString("JWT_PRIVATE_KEY_PATH", "/etc/secrets/jwt_private.pem")
	return auth.LoadPrivateKey(path)
}

func setupPostgres() (*sql.DB, error) {
	return db.NewPostgresClient(db.NewPostgresDefaultConfig())
}

func setupRabbitMQ() (*messaging.RabbitMQ, error) {
	uri := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
	return messaging.NewRabbitMQ(uri)
}
