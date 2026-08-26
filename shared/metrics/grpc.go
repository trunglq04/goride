// Package metrics — gRPC server and client interceptors.
package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that records
// grpc_requests_total and grpc_request_duration_seconds.
// It must be registered after metrics.Init() is called.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if GRPCRequestsTotal == nil {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		code := status.Code(err).String()
		GRPCRequestsTotal.WithLabelValues(info.FullMethod, code).Inc()
		GRPCRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

		return resp, err
	}
}

// DialOptionWithMetrics returns a grpc.DialOption that installs a client-side
// unary interceptor recording grpc_client_requests_total and
// grpc_client_request_duration_seconds for every outbound call.
// Safe to call before Init() — the interceptor becomes a no-op when metrics are nil.
func DialOptionWithMetrics() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if GRPCClientRequestsTotal == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start).Seconds()

		code := status.Code(err).String()
		GRPCClientRequestsTotal.WithLabelValues(method, code).Inc()
		GRPCClientRequestDuration.WithLabelValues(method).Observe(duration)

		return err
	})
}
