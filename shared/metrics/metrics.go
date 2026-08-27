// Package metrics provides a thin Prometheus instrumentation layer for goride services.
// Call Init(serviceName) once at startup, then StartMetricsServer(addr) to expose /metrics.
package metrics

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPRequestsTotal counts total HTTP requests handled by the service.
	HTTPRequestsTotal *prometheus.CounterVec

	// HTTPRequestDuration records HTTP request latency in seconds.
	HTTPRequestDuration *prometheus.HistogramVec

	// GRPCRequestsTotal counts total gRPC unary calls received (server-side).
	GRPCRequestsTotal *prometheus.CounterVec

	// GRPCRequestDuration records gRPC unary call latency in seconds (server-side).
	GRPCRequestDuration *prometheus.HistogramVec

	// GRPCClientRequestsTotal counts outbound gRPC unary calls made by this service.
	GRPCClientRequestsTotal *prometheus.CounterVec

	// GRPCClientRequestDuration records outbound gRPC unary call latency in seconds.
	GRPCClientRequestDuration *prometheus.HistogramVec

	// AMQPMessagesPublished counts messages published to RabbitMQ.
	AMQPMessagesPublished *prometheus.CounterVec

	// AMQPMessagesConsumed counts messages consumed from RabbitMQ, labelled by outcome.
	AMQPMessagesConsumed *prometheus.CounterVec
)

// Init registers all metrics for the given service name.
// It must be called once before using any metric helpers.
func Init(serviceName string) {
	labels := prometheus.Labels{"service": serviceName}

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of HTTP requests received.",
			ConstLabels: labels,
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request latency in seconds.",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "grpc_requests_total",
			Help:        "Total number of gRPC unary calls received.",
			ConstLabels: labels,
		},
		[]string{"method", "status"},
	)

	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "grpc_request_duration_seconds",
			Help:        "gRPC unary call latency in seconds.",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	AMQPMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "amqp_messages_published_total",
			Help:        "Total number of messages published to RabbitMQ.",
			ConstLabels: labels,
		},
		[]string{"routing_key"},
	)

	AMQPMessagesConsumed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "amqp_messages_consumed_total",
			Help:        "Total number of messages consumed from RabbitMQ.",
			ConstLabels: labels,
		},
		[]string{"queue", "status"},
	)

	GRPCClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "grpc_client_requests_total",
			Help:        "Total number of outbound gRPC unary calls made by this service.",
			ConstLabels: labels,
		},
		[]string{"method", "status"},
	)

	GRPCClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "grpc_client_request_duration_seconds",
			Help:        "Outbound gRPC unary call latency in seconds.",
			ConstLabels: labels,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	slog.Info("Prometheus metrics initialized", "service", serviceName)
}

// StartMetricsServer starts a dedicated HTTP server on addr (e.g. ":9091")
// that exposes /metrics for Prometheus scraping. It runs in the background.
func StartMetricsServer(server, addr string) {
	r := gin.New()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	go func() {
		slog.Info("Prometheus metrics server listening", "server", server, "addr", addr)
		if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("Prometheus metrics server error", "server", server, "addr", addr, "err", err)
		}
	}()
}

// RecordPublish increments the AMQP published counter.
// Safe to call even before Init() — the counter will be nil and the call is a no-op.
func RecordPublish(routingKey string) {
	if AMQPMessagesPublished != nil {
		AMQPMessagesPublished.WithLabelValues(routingKey).Inc()
	}
}

// RecordConsume increments the AMQP consumed counter.
// status should be "ok" or "error".
// Safe to call even before Init().
func RecordConsume(queue, status string) {
	if AMQPMessagesConsumed != nil {
		AMQPMessagesConsumed.WithLabelValues(queue, status).Inc()
	}
}
